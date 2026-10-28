package ui

import "image"

// Anchor is the screen point a placed element is attached to.
type Anchor int

const (
	TopLeft Anchor = iota
	Top
	TopRight
	Left
	Middle
	Right
	BottomLeft
	Bottom
	BottomRight
)

// position is the top-left corner of an element of the given size anchored
// on a screen, kept on screen where it fits.
func (a Anchor) position(size, screen image.Point, marginX, marginY int) image.Point {
	col, row := int(a)%3, int(a)/3
	x := place(col, size.X, screen.X, marginX)
	y := place(row, size.Y, screen.Y, marginY)
	return image.Pt(clamp(x, 0, screen.X-size.X), clamp(y, 0, screen.Y-size.Y))
}

// place positions one axis: 0 = start, 1 = center, 2 = end.
func place(where, size, screen, margin int) int {
	switch where {
	case 0:
		return margin
	case 1:
		return (screen-size)/2 + margin
	default:
		return screen - size - margin
	}
}

func clamp(v, lo, hi int) int {
	if hi < lo {
		hi = lo
	}
	return min(max(v, lo), hi)
}
