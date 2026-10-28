package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Element is anything that can be put on screen with UI.Place or inside a
// panel, column, row or modal. Create elements with NewText, NewButton,
// NewPanel, Column and Row.
type Element interface {
	base() *node
	// measure reports the size the element wants. maxW is the width available
	// to it; 0 or less means no limit.
	measure(th *Theme, maxW int) image.Point
	// arrange gives the element its final screen rectangle.
	arrange(th *Theme, r image.Rectangle)
	draw(th *Theme, dst *ebiten.Image, st *drawState)
	// hit returns what is clickable under p (a Button or a Span), or nil.
	hit(th *Theme, p image.Point) clickable
}

// clickable is something that reacts to clicks: a Button or a Span.
type clickable interface {
	click()
}

// drawState is what drawing needs to know about the pointer.
type drawState struct {
	cursor  image.Point
	pressed clickable // what the current press started on
}

// node holds what every element shares. Its exported methods are available on
// every element type.
type node struct {
	rect   image.Rectangle
	hidden bool
}

func (n *node) base() *node { return n }

// Bounds is where the element was last laid out, in screen pixels.
func (n *node) Bounds() image.Rectangle { return n.rect }

// SetVisible shows or hides the element. A hidden element takes up no space.
func (n *node) SetVisible(v bool) { n.hidden = !v }

// Visible reports whether the element is shown.
func (n *node) Visible() bool { return !n.hidden }

// Align positions content inside the space it is given.
type Align int

const (
	// Stretch fills the available space. It is the default for columns.
	// On text it behaves like Start.
	Stretch Align = iota
	Start
	Center
	End
)

// offset is where something of size child starts inside space.
func (a Align) offset(space, child int) int {
	switch a {
	case Center:
		return (space - child) / 2
	case End:
		return space - child
	default:
		return 0
	}
}
