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
	Escape   bool // unfocuses a text field, closes the top modal
	Enter    bool // submits a text field, presses the top modal's first button

	// Editing keys for the focused text field. ReadInput repeats them while
	// the key is held.
	Backspace, Delete, Left, Right, Home, End bool

	// Text is typed into the focused text field. ReadInput leaves it empty:
	// UI.Update gets typing from the platform's input method instead. Set it
	// to type in tests.
	Text string

	imeUsed bool // the input method used keys this frame
}

func (in *Input) anyKey() bool {
	return in.Escape || in.Enter || in.Backspace || in.Delete ||
		in.Left || in.Right || in.Home || in.End || in.Text != ""
}

func (in *Input) clearKeys() {
	in.Escape, in.Enter = false, false
	in.Backspace, in.Delete, in.Left, in.Right, in.Home, in.End = false, false, false, false, false, false
	in.Text = ""
}

// repeated reports whether k was just pressed, or is held long enough to
// repeat (after half a second, then 20 times a second at 60 ticks/second).
func repeated(k ebiten.Key) bool {
	d := inpututil.KeyPressDuration(k)
	return d == 1 || (d >= 30 && (d-30)%3 == 0)
}

// ReadInput reads this frame's input from Ebitengine: the left mouse button,
// the first touch, and the keys the UI uses.
func ReadInput() Input {
	x, y := ebiten.CursorPosition()
	in := Input{
		Cursor:   image.Pt(x, y),
		Pressed:  inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft),
		Released: inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),
		Escape:   inpututil.IsKeyJustPressed(ebiten.KeyEscape),
		Enter:    inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyNumpadEnter),

		Backspace: repeated(ebiten.KeyBackspace),
		Delete:    repeated(ebiten.KeyDelete),
		Left:      repeated(ebiten.KeyArrowLeft),
		Right:     repeated(ebiten.KeyArrowRight),
		Home:      inpututil.IsKeyJustPressed(ebiten.KeyHome),
		End:       inpututil.IsKeyJustPressed(ebiten.KeyEnd),
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
