package ui

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Span is a run of text inside a Text with its own style. Give it a color, a
// highlight, a font, or a click handler to make it stand out from the text
// around it:
//
//	ui.NewRichText(
//		ui.NewSpan("You found "),
//		ui.NewSpan("a rusty key").OnClick(inspectKey),
//		ui.NewSpan("."),
//	)
type Span struct {
	text      string
	color     color.Color
	highlight color.Color
	face      text.Face
	onClick   func()
}

// NewSpan creates a span of text.
func NewSpan(s string) *Span { return &Span{text: s} }

// SetText changes the span's text. Safe to call every frame.
func (s *Span) SetText(t string) *Span { s.text = t; return s }

// Text returns the span's text.
func (s *Span) Text() string { return s.text }

// Color sets the span's text color.
func (s *Span) Color(c color.Color) *Span { s.color = c; return s }

// Highlight draws a background color behind the span.
func (s *Span) Highlight(c color.Color) *Span { s.highlight = c; return s }

// Font sets the span's font, overriding its Text's font and the theme. Spans
// of different sizes on one line share a baseline.
func (s *Span) Font(f text.Face) *Span { s.face = f; return s }

// OnClick makes the span clickable. Unless it has its own Color it is drawn
// in the theme's Link color, and it is underlined while the pointer is over it.
func (s *Span) OnClick(fn func()) *Span { s.onClick = fn; return s }

func (s *Span) click() {
	if s.onClick != nil {
		s.onClick()
	}
}

func (s *Span) plain() bool {
	return s.color == nil && s.highlight == nil && s.face == nil && s.onClick == nil
}

// Text is a block of text made of one or more spans. It wraps at spaces when
// its parent limits its width (inside a modal, or a panel with a Width), and
// always breaks at "\n". Runs of spaces collapse to one.
type Text struct {
	node
	spans []*Span
	color color.Color // default for spans without a color; nil means the theme's
	face  text.Face   // default for spans without a font; nil means the theme's
	align Align
	title bool // a modal title: uses the theme's Title color and TitleFace

	lines []textLine // set by arrange
}

// textLine is one laid-out line: runs of text placed left to right.
type textLine struct {
	runs     []textRun
	width    int
	y        int // offset from the top of the Text
	height   int
	baseline int // offset from the top of the line
}

// textRun is a piece of one span on one line.
type textRun struct {
	span *Span
	face text.Face
	text string
	x, w int // offset from the line's start, and width
}

// NewText creates plain text.
func NewText(s string) *Text { return &Text{spans: []*Span{NewSpan(s)}} }

// NewRichText creates text from styled spans.
func NewRichText(spans ...*Span) *Text { return &Text{spans: spans} }

// SetText replaces the content with plain text. Safe to call every frame.
func (t *Text) SetText(s string) *Text {
	if len(t.spans) == 1 && t.spans[0].plain() {
		t.spans[0].text = s // already plain: reuse the span
		return t
	}
	t.spans = []*Span{NewSpan(s)}
	return t
}

// Text returns the content without styling.
func (t *Text) Text() string {
	var b strings.Builder
	for _, s := range t.spans {
		b.WriteString(s.text)
	}
	return b.String()
}

// SetSpans replaces the content with styled spans.
func (t *Text) SetSpans(spans ...*Span) *Text { t.spans = spans; return t }

// Add appends spans.
func (t *Text) Add(spans ...*Span) *Text { t.spans = append(t.spans, spans...); return t }

// Spans returns the text's spans.
func (t *Text) Spans() []*Span { return t.spans }

// Color sets the color of every span that has no color of its own.
func (t *Text) Color(c color.Color) *Text { t.color = c; return t }

// Font sets the font of every span that has no font of its own, overriding
// the theme.
func (t *Text) Font(f text.Face) *Text { t.face = f; return t }

// Align sets how lines are aligned horizontally: Start, Center or End.
func (t *Text) Align(a Align) *Text { t.align = a; return t }

// defaultFace is the font for spans without their own.
func (t *Text) defaultFace(th *Theme) text.Face {
	switch {
	case t.face != nil:
		return t.face
	case t.title && th.TitleFace != nil:
		return th.TitleFace
	default:
		return th.Face
	}
}

func (t *Text) measure(th *Theme, maxW int) image.Point {
	lines := layoutText(t.spans, t.defaultFace(th), maxW)
	w, h := 0, 0
	for _, l := range lines {
		w = max(w, l.width)
		h += l.height
	}
	return image.Pt(w, h)
}

func (t *Text) arrange(th *Theme, r image.Rectangle) {
	t.rect = r
	t.lines = layoutText(t.spans, t.defaultFace(th), r.Dx())
}

// runRect is where a run sits on screen: its width, and the full line height.
func (t *Text) runRect(line int, r textRun) image.Rectangle {
	l := t.lines[line]
	x := t.rect.Min.X + t.align.offset(t.rect.Dx(), l.width) + r.x
	y := t.rect.Min.Y + l.y
	return image.Rect(x, y, x+r.w, y+l.height)
}

func (t *Text) draw(th *Theme, dst *ebiten.Image, st *drawState) {
	var hovered *Span
	if c, ok := t.hit(th, st.cursor).(*Span); ok {
		hovered = c
	}

	for i, line := range t.lines {
		for _, run := range line.runs {
			r := t.runRect(i, run)
			if run.span.highlight != nil {
				fillRect(dst, r, run.span.highlight)
			}
			c := t.colorOf(th, run.span, run.span == hovered)
			// Line every run up on the line's baseline.
			y := r.Min.Y + line.baseline - ceil(run.face.Metrics().HAscent)
			drawString(dst, run.face, run.text, r.Min.X, y, c)
			if run.span == hovered {
				uy := min(r.Min.Y+line.baseline+1, r.Max.Y-1)
				fillRect(dst, image.Rect(r.Min.X, uy, r.Max.X, uy+1), c)
			}
		}
	}
}

func (t *Text) colorOf(th *Theme, s *Span, hovered bool) color.Color {
	switch {
	case s.color != nil:
		return s.color
	case s.onClick != nil && hovered:
		return th.LinkHover
	case s.onClick != nil:
		return th.Link
	case t.color != nil:
		return t.color
	case t.title:
		return th.Title
	default:
		return th.Text
	}
}

func (t *Text) hit(_ *Theme, p image.Point) clickable {
	if !p.In(t.rect) {
		return nil
	}
	for i, line := range t.lines {
		for _, run := range line.runs {
			if run.span.onClick != nil && p.In(t.runRect(i, run)) {
				return run.span
			}
		}
	}
	return nil
}

// layoutText places spans into lines no wider than maxW pixels, breaking at
// spaces and at "\n". maxW <= 0 means only break at "\n". A word wider than
// maxW gets a line of its own and overflows. A word may cross spans, e.g.
// "key" and "." stay together. def is the font for spans without one.
func layoutText(spans []*Span, def text.Face, maxW int) []textLine {
	var (
		lines []textLine
		cur   textLine
		x     int
		space *Span     // the span of the space before the next word
		word  []textRun // pieces of the word being collected
		wordW int
	)
	faceOf := func(s *Span) text.Face {
		if s.face != nil {
			return s.face
		}
		return def
	}
	endLine := func() {
		cur.width = x
		lines = append(lines, finishLine(cur, def))
		cur, x, space = textLine{}, 0, nil
	}
	placeWord := func() {
		if len(word) == 0 {
			return
		}
		if space == nil {
			space = word[0].span
		}
		spaceW := advance(faceOf(space), " ")
		if len(cur.runs) > 0 && maxW > 0 && x+spaceW+wordW > maxW {
			endLine()
		}
		if len(cur.runs) > 0 {
			cur.runs = append(cur.runs, textRun{span: space, face: faceOf(space), text: " ", x: x, w: spaceW})
			x += spaceW
		}
		for _, r := range word {
			r.x = x
			cur.runs = append(cur.runs, r)
			x += r.w
		}
		word, wordW, space = word[:0], 0, nil
	}

	for _, sp := range spans {
		f := faceOf(sp)
		for s := sp.text; len(s) > 0; {
			switch s[0] {
			case '\n':
				placeWord()
				endLine()
				s = s[1:]
			case ' ', '\t':
				placeWord()
				if len(cur.runs) > 0 && space == nil {
					space = sp
				}
				s = s[1:]
			default:
				i := strings.IndexAny(s, " \t\n")
				if i < 0 {
					i = len(s)
				}
				w := advance(f, s[:i])
				word = append(word, textRun{span: sp, face: f, text: s[:i], w: w})
				wordW += w
				s = s[i:]
			}
		}
	}
	placeWord()
	endLine()

	y := 0
	for i := range lines {
		lines[i].y = y
		y += lines[i].height
	}
	return lines
}

// finishLine joins neighbouring runs of the same span, so a highlight covers
// the spaces between its words, and sizes the line to its tallest font. An
// empty line takes the height of def.
func finishLine(l textLine, def text.Face) textLine {
	var out []textRun
	for _, r := range l.runs {
		if n := len(out); n > 0 && out[n-1].span == r.span {
			out[n-1].text += r.text
			out[n-1].w += r.w
			continue
		}
		out = append(out, r)
	}
	l.runs = out

	ascent, below := 0, 0 // above and below the baseline
	measure := func(f text.Face) {
		m := f.Metrics()
		ascent = max(ascent, ceil(m.HAscent))
		below = max(below, ceil(m.HDescent+m.HLineGap))
	}
	if len(l.runs) == 0 {
		measure(def)
	}
	for _, r := range l.runs {
		measure(r.face)
	}
	l.baseline, l.height = ascent, ascent+below
	return l
}

func advance(f text.Face, s string) int { return ceil(text.Advance(s, f)) }

func ceil(v float64) int { return int(math.Ceil(v)) }

func drawString(dst *ebiten.Image, f text.Face, s string, x, y int, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, f, op)
}
