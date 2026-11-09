package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Button is a clickable label. Its callback runs when the pointer is pressed
// and released over it.
type Button struct {
	node
	label    string
	onClick  func()
	disabled bool
	selected bool
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

// SetSelected marks the button as the chosen one, e.g. the active tool in a
// toolbar. It is drawn in the theme's ButtonSelected color.
func (b *Button) SetSelected(s bool) *Button { b.selected = s; return b }

// Selected reports whether the button is marked as selected.
func (b *Button) Selected() bool { return b.selected }

// Disabled reports whether the button is disabled.
func (b *Button) Disabled() bool { return b.disabled }

func (b *Button) measure(th *Theme, _ int) image.Point {
	return image.Pt(
		th.textWidth(b.label)+2*th.ButtonPaddingX,
		th.lineHeight()+2*th.ButtonPaddingY+th.ButtonDepth,
	)
}

func (b *Button) arrange(_ *Theme, r image.Rectangle) { b.rect = r }

func (b *Button) draw(th *Theme, dst *ebiten.Image, st *drawState) {
	hover := st.cursor.In(b.rect)
	pressed := hover && st.pressed == b && !b.disabled
	bg, fg := th.ButtonBackground, th.ButtonText
	switch {
	case b.disabled:
		fg = th.ButtonDisabled
	case pressed:
		bg = th.ButtonPressed
	case b.selected:
		bg = th.ButtonSelected
	case hover:
		bg = th.ButtonHover
	}
	// A raised button's face sits ButtonDepth above its edge; pressed, it sinks onto the
	// edge, leaving a gap above it.
	face := b.rect
	if d := th.ButtonDepth; d > 0 {
		edge := image.Rect(face.Min.X, face.Min.Y+d, face.Max.X, face.Max.Y)
		th.buttonShape(dst, edge, th.ButtonEdge)
		face.Max.Y -= d
		if pressed {
			face = face.Add(image.Pt(0, d))
		}
	}
	th.buttonShape(dst, face, bg)

	x := face.Min.X + Center.offset(face.Dx(), th.textWidth(b.label))
	y := face.Min.Y + Center.offset(face.Dy(), th.lineHeight())
	drawString(dst, th.Face, b.label, x, y, fg)
}

// buttonShape fills r in c, with the theme's button corners and border.
func (th *Theme) buttonShape(dst *ebiten.Image, r image.Rectangle, c color.Color) {
	if th.ButtonRadius > 0 {
		fillRoundRect(dst, r, th.ButtonRadius, c)
		if th.ButtonBorder != nil {
			strokeRoundRect(dst, r, th.ButtonRadius, 1, th.ButtonBorder)
		}
		return
	}
	fillRect(dst, r, c)
	if th.ButtonBorder != nil {
		strokeRect(dst, r, th.ButtonBorder)
	}
}

func (b *Button) hit(_ *Theme, p image.Point) clickable {
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
