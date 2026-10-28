package ui

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Theme is the look shared by every element: font, spacing and colors.
// Change fields on UI.Theme to restyle everything at once.
type Theme struct {
	// Face is the font used for all text.
	Face text.Face

	// Padding is the space between a panel's or modal's edge and its content.
	Padding int
	// Gap is the space between neighbouring elements in a column or row.
	Gap int
	// ButtonPaddingX and ButtonPaddingY are the space around a button's label.
	ButtonPaddingX, ButtonPaddingY int
	// ModalWidth is the widest a modal grows before its text wraps.
	ModalWidth int

	Text  color.Color // body text
	Title color.Color // modal titles

	PanelBackground color.Color
	PanelBorder     color.Color

	ButtonBackground color.Color
	ButtonHover      color.Color
	ButtonPressed    color.Color
	ButtonText       color.Color
	ButtonDisabled   color.Color // label color of a disabled button

	// Backdrop is drawn over everything behind the top modal.
	Backdrop color.Color
}

// DefaultTheme is a dark theme using a 12px bitmap font.
func DefaultTheme() Theme {
	return Theme{
		Face:           text.NewGoXFace(bitmapfont.Face),
		Padding:        8,
		Gap:            6,
		ButtonPaddingX: 8,
		ButtonPaddingY: 4,
		ModalWidth:     280,

		Text:  color.RGBA{0xe6, 0xe6, 0xe6, 0xff},
		Title: color.RGBA{0xf2, 0xc9, 0x4c, 0xff},

		PanelBackground: color.RGBA{0x1e, 0x1e, 0x24, 0xf0},
		PanelBorder:     color.RGBA{0x8a, 0x8a, 0x99, 0xff},

		ButtonBackground: color.RGBA{0x3a, 0x3a, 0x46, 0xff},
		ButtonHover:      color.RGBA{0x50, 0x50, 0x62, 0xff},
		ButtonPressed:    color.RGBA{0x2a, 0x2a, 0x33, 0xff},
		ButtonText:       color.RGBA{0xf0, 0xf0, 0xf0, 0xff},
		ButtonDisabled:   color.RGBA{0x80, 0x80, 0x88, 0xff},

		Backdrop: color.RGBA{0x00, 0x00, 0x00, 0x99},
	}
}

// lineHeight is the vertical advance of one line of text.
func (th *Theme) lineHeight() int {
	m := th.Face.Metrics()
	return int(math.Ceil(m.HAscent + m.HDescent + m.HLineGap))
}

// textWidth is the width of a single line of text in pixels.
func (th *Theme) textWidth(s string) int {
	return int(math.Ceil(text.Advance(s, th.Face)))
}
