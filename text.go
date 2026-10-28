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
// highlight, or a click handler to make it stand out from the text around it:
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

// OnClick makes the span clickable. Unless it has its own Color it is drawn
// in the theme's Link color, and it is underlined while the pointer is over it.
func (s *Span) OnClick(fn func()) *Span { s.onClick = fn; return s }

func (s *Span) click() {
	if s.onClick != nil {
		s.onClick()
	}
}

// Text is a block of text made of one or more spans. It wraps at spaces when
// its parent limits its width (inside a modal, or a panel with a Width), and
// always breaks at "\n". Runs of spaces collapse to one.
type Text struct {
	node
	spans []*Span
	color color.Color // default for spans without a color; nil means the theme's
	align Align
	title bool // drawn in the theme's Title color

	lines []textLine // set by arrange
}

// textLine is one laid-out line: runs of text placed left to right.
type textLine struct {
	runs  []textRun
	width int
}

// textRun is a piece of one span on one line.
type textRun struct {
	span *Span
	text string
	x, w int // offset from the line's start, and width
}

// NewText creates plain text.
func NewText(s string) *Text { return &Text{spans: []*Span{NewSpan(s)}} }

// NewRichText creates text from styled spans.
func NewRichText(spans ...*Span) *Text { return &Text{spans: spans} }

// SetText replaces the content with plain text. Safe to call every frame.
func (t *Text) SetText(s string) *Text {
	if len(t.spans) == 1 {
		if sp := t.spans[0]; sp.color == nil && sp.highlight == nil && sp.onClick == nil {
			sp.text = s // already plain: reuse the span
			return t
		}
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

// Align sets how lines are aligned horizontally: Start, Center or End.
func (t *Text) Align(a Align) *Text { t.align = a; return t }

func (t *Text) measure(th *Theme, maxW int) image.Point {
	lines := layoutText(th, t.spans, maxW)
	w := 0
	for _, l := range lines {
		w = max(w, l.width)
	}
	return image.Pt(w, len(lines)*th.lineHeight())
}

func (t *Text) arrange(th *Theme, r image.Rectangle) {
	t.rect = r
	t.lines = layoutText(th, t.spans, r.Dx())
}

// runRect is where a run sits on screen.
func (t *Text) runRect(th *Theme, line int, r textRun) image.Rectangle {
	lh := th.lineHeight()
	x := t.rect.Min.X + t.align.offset(t.rect.Dx(), t.lines[line].width) + r.x
	y := t.rect.Min.Y + line*lh
	return image.Rect(x, y, x+r.w, y+lh)
}

func (t *Text) draw(th *Theme, dst *ebiten.Image, st *drawState) {
	var hovered *Span
	if c, ok := t.hit(th, st.cursor).(*Span); ok {
		hovered = c
	}
	underlineY := int(math.Ceil(th.Face.Metrics().HAscent)) + 1

	for i, line := range t.lines {
		for _, run := range line.runs {
			r := t.runRect(th, i, run)
			if run.span.highlight != nil {
				fillRect(dst, r, run.span.highlight)
			}
			c := t.colorOf(th, run.span, run.span == hovered)
			drawString(dst, th, run.text, r.Min.X, r.Min.Y, c)
			if run.span == hovered {
				y := min(r.Min.Y+underlineY, r.Max.Y-1)
				fillRect(dst, image.Rect(r.Min.X, y, r.Max.X, y+1), c)
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

func (t *Text) hit(th *Theme, p image.Point) clickable {
	if !p.In(t.rect) {
		return nil
	}
	for i, line := range t.lines {
		for _, run := range line.runs {
			if run.span.onClick != nil && p.In(t.runRect(th, i, run)) {
				return run.span
			}
		}
	}
	return nil
}

// layoutText places spans into lines no wider than maxW pixels, breaking at
// spaces and at "\n". maxW <= 0 means only break at "\n". A word wider than
// maxW gets a line of its own and overflows. A word may cross spans, e.g.
// "key" and "." stay together.
func layoutText(th *Theme, spans []*Span, maxW int) []textLine {
	spaceW := th.textWidth(" ")
	var (
		lines []textLine
		cur   textLine
		x     int
		space *Span     // the span of the space before the next word
		word  []textRun // pieces of the word being collected
		wordW int
	)
	endLine := func() {
		cur.width = x
		lines = append(lines, mergeRuns(cur))
		cur, x, space = textLine{}, 0, nil
	}
	placeWord := func() {
		if len(word) == 0 {
			return
		}
		if len(cur.runs) > 0 && maxW > 0 && x+spaceW+wordW > maxW {
			endLine()
		}
		if len(cur.runs) > 0 {
			if space == nil {
				space = word[0].span
			}
			cur.runs = append(cur.runs, textRun{span: space, text: " ", x: x, w: spaceW})
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
				w := th.textWidth(s[:i])
				word = append(word, textRun{span: sp, text: s[:i], w: w})
				wordW += w
				s = s[i:]
			}
		}
	}
	placeWord()
	endLine()
	return lines
}

// mergeRuns joins neighbouring runs of the same span, so a highlight covers
// the spaces between its words and each span is drawn in one call per line.
func mergeRuns(l textLine) textLine {
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
	return l
}

func drawString(dst *ebiten.Image, th *Theme, s string, x, y int, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, th.Face, op)
}
