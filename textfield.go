package ui

import (
	"image"
	"strings"
	"unicode/utf8"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/exp/textinput"
)

// TextField is a single-line box the player can type into. Typing goes
// through the platform's input method (Ebitengine's exp/textinput), so
// Japanese/Chinese input methods and on-screen keyboards work, not just plain
// keys.
//
// Click a field to focus it, or call UI.Focus. While it has focus it takes the
// keyboard: Backspace, Delete, Left, Right, Home and End edit, Enter calls
// OnSubmit, and Escape gives up focus.
//
//	name := ui.NewTextField("Your name").OnSubmit(func(s string) { ... })
type TextField struct {
	node
	text        string
	caret       int // byte offset into text
	placeholder string
	width       int
	maxLen      int // in runes; 0 means no limit
	onChange    func(string)
	onSubmit    func(string)

	focused  bool
	arranged bool   // laid out this frame; a field that is not loses focus
	th       *Theme // the theme it was last laid out with, for the input method
	scroll   int    // pixels of text hidden on the left to keep the caret visible

	composer    textinput.Composer
	composition string // text the input method is still composing
	compCaret   int    // caret inside composition, in bytes
	imeStart    int    // the part of text handed to the input method:
	imeEnd      int    // text[imeStart:imeEnd]
}

// NewTextField creates an empty text field. placeholder is shown greyed out
// while it is empty; pass "" for none.
func NewTextField(placeholder string) *TextField {
	f := &TextField{placeholder: placeholder}
	f.composer.OnNewSession = f.onNewSession
	f.composer.OnComposition = func(c *textinput.Composition) {
		f.composition = c.Text()
		f.compCaret, _ = c.SelectionRangeInBytes()
	}
	f.composer.OnCommit = f.onCommit
	f.composer.OnEndByUser = func() { f.focused = false }
	return f
}

// SetText replaces the text and puts the caret at the end. It does not call
// OnChange.
func (f *TextField) SetText(s string) *TextField {
	f.composer.Confirm()
	f.text = singleLine(s)
	f.caret = len(f.text)
	return f
}

// Text returns the typed text, without any half-composed input.
func (f *TextField) Text() string { return f.text }

// SetPlaceholder changes the greyed-out hint shown while the field is empty.
func (f *TextField) SetPlaceholder(s string) *TextField { f.placeholder = s; return f }

// Width sets the field's width in pixels, overriding the theme's FieldWidth.
func (f *TextField) Width(n int) *TextField { f.width = n; return f }

// MaxLength limits how many characters can be typed. 0 means no limit.
func (f *TextField) MaxLength(n int) *TextField { f.maxLen = n; return f }

// OnChange sets a callback that runs with the new text after every edit.
func (f *TextField) OnChange(fn func(string)) *TextField { f.onChange = fn; return f }

// OnSubmit sets a callback that runs with the text when Enter is pressed.
// Without one, Enter in a modal presses the modal's first button instead.
func (f *TextField) OnSubmit(fn func(string)) *TextField { f.onSubmit = fn; return f }

// Focused reports whether the field has the keyboard.
func (f *TextField) Focused() bool { return f.focused }

func (f *TextField) measure(th *Theme, maxW int) image.Point {
	w := f.width
	if w <= 0 {
		w = th.FieldWidth
	}
	if maxW > 0 && w > maxW {
		w = maxW
	}
	return image.Pt(w, th.lineHeight()+2*th.ButtonPaddingY)
}

func (f *TextField) arrange(th *Theme, r image.Rectangle) {
	f.rect = r
	f.th = th
	f.arranged = true
}

// inner is the area text is drawn in.
func (f *TextField) inner(th *Theme) image.Rectangle {
	return image.Rect(
		f.rect.Min.X+th.ButtonPaddingX, f.rect.Min.Y+th.ButtonPaddingY,
		f.rect.Max.X-th.ButtonPaddingX, f.rect.Max.Y-th.ButtonPaddingY,
	)
}

// display is the text as shown: the typed text with any half-composed input
// at the caret, and where the caret is inside that string.
func (f *TextField) display() (s string, caret int) {
	if f.composition == "" {
		return f.text, f.caret
	}
	return f.text[:f.caret] + f.composition + f.text[f.caret:], f.caret + f.compCaret
}

func (f *TextField) draw(th *Theme, dst *ebiten.Image, _ *drawState) {
	border := th.FieldBorder
	if f.focused {
		border = th.FieldFocus
	}
	fillRect(dst, f.rect, th.FieldBackground)
	strokeRect(dst, f.rect, border)

	in := f.inner(th)
	if in.Empty() {
		return
	}
	s, caret := f.display()
	caretX := advance(th.Face, s[:caret])

	// Scroll so the caret stays inside the box.
	switch {
	case advance(th.Face, s) < in.Dx():
		f.scroll = 0
	case caretX-f.scroll > in.Dx()-1:
		f.scroll = caretX - in.Dx() + 1
	case caretX < f.scroll:
		f.scroll = caretX
	}

	clip := dst.SubImage(in).(*ebiten.Image)
	x, y := in.Min.X-f.scroll, in.Min.Y+Center.offset(in.Dy(), th.lineHeight())
	if s == "" {
		drawString(clip, th.Face, f.placeholder, in.Min.X, y, th.Placeholder)
	} else {
		drawString(clip, th.Face, s, x, y, th.Text)
	}
	if f.composition != "" { // underline what is still being composed
		ux := x + advance(th.Face, f.text[:f.caret])
		uy := y + th.lineHeight() - 1
		fillRect(clip, image.Rect(ux, uy, ux+advance(th.Face, f.composition), uy+1), th.Text)
	}
	if f.focused {
		cx := x + caretX
		fillRect(clip, image.Rect(cx, y, cx+1, y+th.lineHeight()), th.Text)
	}
}

func (f *TextField) hit(_ *Theme, p image.Point) clickable {
	if p.In(f.rect) {
		return f
	}
	return nil
}

// click does nothing: a field takes focus when it is pressed (see UI.pointer).
func (f *TextField) click() {}

// placeCaret moves the caret to the character boundary nearest p.
func (f *TextField) placeCaret(th *Theme, p image.Point) {
	f.composer.Confirm()
	x := p.X - f.inner(th).Min.X + f.scroll
	best, bestDist := 0, abs(x)
	for i := range f.text {
		if i == 0 {
			continue
		}
		if d := abs(x - advance(th.Face, f.text[:i])); d < bestDist {
			best, bestDist = i, d
		}
	}
	if d := abs(x - advance(th.Face, f.text)); d < bestDist {
		best = len(f.text)
	}
	f.caret = best
}

// edit applies this frame's typing and editing keys.
func (f *TextField) edit(in Input) {
	if in.Text != "" {
		f.composer.Confirm()
		f.insert(in.Text)
	}
	if !(in.Backspace || in.Delete || in.Left || in.Right || in.Home || in.End) {
		return
	}
	// Edits outside the input method end its session first (see textinput.Composer).
	f.composer.Confirm()
	switch {
	case in.Backspace && f.caret > 0:
		_, n := utf8.DecodeLastRuneInString(f.text[:f.caret])
		f.text = f.text[:f.caret-n] + f.text[f.caret:]
		f.caret -= n
		f.changed()
	case in.Delete && f.caret < len(f.text):
		_, n := utf8.DecodeRuneInString(f.text[f.caret:])
		f.text = f.text[:f.caret] + f.text[f.caret+n:]
		f.changed()
	case in.Left && f.caret > 0:
		_, n := utf8.DecodeLastRuneInString(f.text[:f.caret])
		f.caret -= n
	case in.Right && f.caret < len(f.text):
		_, n := utf8.DecodeRuneInString(f.text[f.caret:])
		f.caret += n
	case in.Home:
		f.caret = 0
	case in.End:
		f.caret = len(f.text)
	}
}

// insert puts s at the caret, cut to fit MaxLength. It must not call
// composer.Confirm: it runs inside OnCommit.
func (f *TextField) insert(s string) {
	s = singleLine(s)
	if f.maxLen > 0 {
		room := f.maxLen - utf8.RuneCountInString(f.text)
		if room <= 0 {
			return
		}
		if utf8.RuneCountInString(s) > room {
			s = string([]rune(s)[:room])
		}
	}
	if s == "" {
		return
	}
	f.text = f.text[:f.caret] + s + f.text[f.caret:]
	f.caret += len(s)
	f.changed()
}

func (f *TextField) changed() {
	if f.onChange != nil {
		f.onChange(f.text)
	}
}

// onNewSession tells the input method where the caret is on screen and what
// text surrounds it.
func (f *TextField) onNewSession() *textinput.SessionOptions {
	if f.th == nil {
		return nil // not laid out yet; the composer asks again next frame
	}
	in := f.inner(f.th)
	x := in.Min.X - f.scroll + advance(f.th.Face, f.text[:f.caret])
	y := in.Min.Y + Center.offset(in.Dy(), f.th.lineHeight())
	f.imeStart, f.imeEnd = 0, len(f.text)
	return &textinput.SessionOptions{
		CaretBounds:     image.Rect(x, y, x+1, y+f.th.lineHeight()),
		TextBeforeCaret: f.text[:f.caret],
		TextAfterCaret:  f.text[f.caret:],
	}
}

// onCommit applies text the input method finished. Usually that is an insert
// at the caret; sometimes the input method rewrites the text around it.
func (f *TextField) onCommit(c *textinput.Commit) {
	if before, after := c.IsSurroundingTextReplaced(); before || after {
		b, a := c.SurroundingText()
		f.text = f.text[:f.imeStart] + b + singleLine(c.Text()) + a + f.text[f.imeEnd:]
		f.caret = f.imeStart + len(b) + len(singleLine(c.Text()))
		f.changed()
		return
	}
	f.insert(c.Text())
}

// singleLine drops line breaks.
func singleLine(s string) string {
	if !strings.ContainsAny(s, "\r\n") {
		return s
	}
	return strings.NewReplacer("\r", "", "\n", "").Replace(s)
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
