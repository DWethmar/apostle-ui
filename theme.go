package ui

import (
	"image/color"

	"github.com/hajimehoshi/bitmapfont/v4"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

// Theme is the look shared by every element: font, spacing and colors.
// Change fields on UI.Theme to restyle everything at once.
type Theme struct {
	// Face is the font for all text and buttons. Text and Span can override it
	// with Font.
	Face text.Face
	// TitleFace is the font for modal titles. Nil means Face.
	TitleFace text.Face

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

	Link      color.Color // clickable spans
	LinkHover color.Color // clickable spans under the pointer

	PanelBackground color.Color
	PanelBorder     color.Color

	ButtonBackground color.Color
	ButtonHover      color.Color
	ButtonPressed    color.Color
	ButtonSelected   color.Color // a button marked with SetSelected
	ButtonText       color.Color
	ButtonDisabled   color.Color // label color of a disabled button

	FieldWidth      int         // default width of a text field
	FieldBackground color.Color // inside a text field
	FieldBorder     color.Color // a text field's border
	FieldFocus      color.Color // a focused text field's border
	Placeholder     color.Color // a text field's hint while it is empty

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

		Link:      color.RGBA{0x7e, 0xc8, 0xff, 0xff},
		LinkHover: color.RGBA{0xc4, 0xe6, 0xff, 0xff},

		PanelBackground: color.RGBA{0x1e, 0x1e, 0x24, 0xf0},
		PanelBorder:     color.RGBA{0x8a, 0x8a, 0x99, 0xff},

		ButtonBackground: color.RGBA{0x3a, 0x3a, 0x46, 0xff},
		ButtonHover:      color.RGBA{0x50, 0x50, 0x62, 0xff},
		ButtonPressed:    color.RGBA{0x2a, 0x2a, 0x33, 0xff},
		ButtonSelected:   color.RGBA{0x4a, 0x78, 0x46, 0xff},
		ButtonText:       color.RGBA{0xf0, 0xf0, 0xf0, 0xff},
		ButtonDisabled:   color.RGBA{0x80, 0x80, 0x88, 0xff},

		FieldWidth:      160,
		FieldBackground: color.RGBA{0x12, 0x12, 0x16, 0xff},
		FieldBorder:     color.RGBA{0x5a, 0x5a, 0x66, 0xff},
		FieldFocus:      color.RGBA{0x7e, 0xc8, 0xff, 0xff},
		Placeholder:     color.RGBA{0x70, 0x70, 0x7a, 0xff},

		Backdrop: color.RGBA{0x00, 0x00, 0x00, 0x99},
	}
}

// lineHeight is the height of one line in the theme's font.
func (th *Theme) lineHeight() int {
	m := th.Face.Metrics()
	return ceil(m.HAscent + m.HDescent + m.HLineGap)
}

// textWidth is the width of a single line in the theme's font.
func (th *Theme) textWidth(s string) int { return advance(th.Face, s) }
