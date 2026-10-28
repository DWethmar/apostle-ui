package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Button is a clickable label. Its callback runs when the pointer is pressed
// and released over it.
type Button struct {
	node
	label    string
	onClick  func()
	disabled bool
}

// NewButton creates a button. onClick may be nil.
func NewButton(label string, onClick func()) *Button {
	return &Button{label: label, onClick: onClick}
}

// SetLabel changes the button's text.
func (b *Button) SetLabel(s string) *Button { b.label = s; return b }

// Label returns the button's text.
func (b *Button) Label() string { return b.label }

// SetDisabled greys the button out and stops it reacting to clicks.
func (b *Button) SetDisabled(d bool) *Button { b.disabled = d; return b }

// Disabled reports whether the button is disabled.
func (b *Button) Disabled() bool { return b.disabled }

func (b *Button) measure(th *Theme, _ int) image.Point {
	return image.Pt(
		th.textWidth(b.label)+2*th.ButtonPaddingX,
		th.lineHeight()+2*th.ButtonPaddingY,
	)
}

func (b *Button) arrange(_ *Theme, r image.Rectangle) { b.rect = r }

func (b *Button) draw(th *Theme, dst *ebiten.Image, st *drawState) {
	hover := st.cursor.In(b.rect)
	bg, fg := th.ButtonBackground, th.ButtonText
	switch {
	case b.disabled:
		fg = th.ButtonDisabled
	case hover && st.pressed == b:
		bg = th.ButtonPressed
	case hover:
		bg = th.ButtonHover
	}
	fillRect(dst, b.rect, bg)
	strokeRect(dst, b.rect, th.PanelBorder)

	x := b.rect.Min.X + Center.offset(b.rect.Dx(), th.textWidth(b.label))
	y := b.rect.Min.Y + Center.offset(b.rect.Dy(), th.lineHeight())
	drawString(dst, th, b.label, x, y, fg)
}

func (b *Button) hit(p image.Point) clickable {
	if b.disabled || !p.In(b.rect) {
		return nil
	}
	return b
}

func (b *Button) click() {
	if !b.disabled && b.onClick != nil {
		b.onClick()
	}
}
