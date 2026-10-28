package ui

import (
	"image"
	"image/color"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Text is a block of text. It wraps at word boundaries when its parent limits
// its width (inside a modal, or a panel with a Width), and always breaks at
// "\n".
type Text struct {
	node
	text  string
	color color.Color // nil means the theme's color
	align Align
	title bool // drawn in the theme's Title color

	lines []string // the wrapped lines, set by arrange
}

// NewText creates a text element.
func NewText(s string) *Text { return &Text{text: s} }

// SetText changes the text. Safe to call every frame.
func (t *Text) SetText(s string) *Text { t.text = s; return t }

// Text returns the current text.
func (t *Text) Text() string { return t.text }

// Color sets the text color, overriding the theme.
func (t *Text) Color(c color.Color) *Text { t.color = c; return t }

// Align sets how lines are aligned horizontally: Start, Center or End.
func (t *Text) Align(a Align) *Text { t.align = a; return t }

func (t *Text) measure(th *Theme, maxW int) image.Point {
	lines := wrapLines(th, t.text, maxW)
	w := 0
	for _, l := range lines {
		w = max(w, th.textWidth(l))
	}
	return image.Pt(w, len(lines)*th.lineHeight())
}

func (t *Text) arrange(th *Theme, r image.Rectangle) {
	t.rect = r
	t.lines = wrapLines(th, t.text, r.Dx())
}

func (t *Text) draw(th *Theme, dst *ebiten.Image, _ *drawState) {
	c := t.color
	if c == nil {
		c = th.Text
		if t.title {
			c = th.Title
		}
	}
	lh := th.lineHeight()
	for i, l := range t.lines {
		x := t.rect.Min.X + t.align.offset(t.rect.Dx(), th.textWidth(l))
		drawString(dst, th, l, x, t.rect.Min.Y+i*lh, c)
	}
}

func (t *Text) hit(image.Point) clickable { return nil }

// wrapLines splits s into lines no wider than maxW pixels, breaking at spaces
// and at "\n". maxW <= 0 means only break at "\n". A word wider than maxW is
// kept whole on its own line.
func wrapLines(th *Theme, s string, maxW int) []string {
	paragraphs := strings.Split(s, "\n")
	if maxW <= 0 {
		return paragraphs
	}
	var lines []string
	for _, p := range paragraphs {
		words := strings.Fields(p)
		if len(words) == 0 {
			lines = append(lines, "")
			continue
		}
		line := words[0]
		for _, w := range words[1:] {
			if next := line + " " + w; th.textWidth(next) <= maxW {
				line = next
				continue
			}
			lines = append(lines, line)
			line = w
		}
		lines = append(lines, line)
	}
	return lines
}

func drawString(dst *ebiten.Image, th *Theme, s string, x, y int, c color.Color) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(x), float64(y))
	op.ColorScale.ScaleWithColor(c)
	text.Draw(dst, s, th.Face, op)
}
