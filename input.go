package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Input is one frame of pointer and keyboard input, in screen coordinates.
type Input struct {
	Cursor   image.Point
	Pressed  bool // the primary button or a touch went down this frame
	Released bool // the primary button or a touch went up this frame
	Escape   bool // closes the top modal
	Enter    bool // presses the top modal's first button
}

// ReadInput reads this frame's input from Ebitengine: the left mouse button,
// the first touch, and the Escape and Enter keys.
func ReadInput() Input {
	x, y := ebiten.CursorPosition()
	in := Input{
		Cursor:   image.Pt(x, y),
		Pressed:  inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		Released: inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),
		Escape:   inpututil.IsKeyJustPressed(ebiten.KeyEscape),
		Enter:    inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter),
	}

	if ids := ebiten.AppendTouchIDs(nil); len(ids) > 0 {
		x, y := ebiten.TouchPosition(ids[0])
		in.Cursor = image.Pt(x, y)
	}
	if ids := inpututil.AppendJustPressedTouchIDs(nil); len(ids) > 0 {
		x, y := ebiten.TouchPosition(ids[0])
		in.Cursor, in.Pressed = image.Pt(x, y), true
	}
	if ids := inpututil.AppendJustReleasedTouchIDs(nil); len(ids) > 0 {
		x, y := inpututil.TouchPositionInPreviousTick(ids[0])
		in.Cursor, in.Released = image.Pt(x, y), true
	}
	return in
}
