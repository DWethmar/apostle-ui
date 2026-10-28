// Package ui is a small retained-mode UI for Ebitengine games: panels with
// text and buttons placed on the screen, and modals that sit on top of
// everything.
//
// Build elements once, keep pointers to the ones you want to change, and call
// Update and Draw from your game:
//
//	u := ui.New()
//	gold := ui.NewText("Gold: 0")
//	u.Place(ui.NewPanel(gold), ui.TopLeft, 8, 8)
//	u.Alert("Welcome", "You wake up in a dark cave.")
//
//	func (g *Game) Update() error {
//		if g.ui.Update() {
//			return nil // the UI used this frame's input
//		}
//		g.gold.SetText(fmt.Sprintf("Gold: %d", g.coins))
//		...
//	}
//
//	func (g *Game) Draw(screen *ebiten.Image) {
//		... // draw the world first
//		g.ui.Draw(screen)
//	}
package ui

import (
	"image"
	"slices"

	"github.com/hajimehoshi/ebiten/v2"
)

// screenMargin keeps modals this far from the screen edge.
const screenMargin = 8

// UI holds everything on screen: placed elements and a stack of open modals.
type UI struct {
	// Theme is the look of every element. Change it at any time.
	Theme Theme

	placed  []*placement
	modals  []*Modal
	screen  image.Point
	cursor  image.Point
	pressed clickable  // what the current press started on
	focused *TextField // the field that has the keyboard, or nil
}

type placement struct {
	el      Element
	anchor  Anchor
	marginX int
	marginY int
	at      *image.Point // set by PlaceAt: a fixed top-left corner instead of an anchor
}

func (u *UI) placementOf(e Element) *placement {
	for _, p := range u.placed {
		if p.el == e {
			return p
		}
	}
	p := &placement{el: e}
	u.placed = append(u.placed, p)
	return p
}

// New creates a UI with the default theme.
func New() *UI {
	return &UI{Theme: DefaultTheme(), cursor: image.Pt(-1, -1)}
}

// Place puts e on the screen at an anchor. The margins push it inward from
// the anchored edges; for a centered axis they shift it right or down.
// Placing an element that is already placed moves it.
func (u *UI) Place(e Element, a Anchor, marginX, marginY int) {
	p := u.placementOf(e)
	p.anchor, p.marginX, p.marginY, p.at = a, marginX, marginY, nil
}

// PlaceAt puts e on the screen with its top-left corner at (x, y), moved
// inward if it would stick out past an edge. Use it for things that follow a
// point, like a popup next to a character; call it every frame as the point
// moves. Elements are drawn in the order they were first placed.
func (u *UI) PlaceAt(e Element, x, y int) {
	p := u.placementOf(e)
	at := image.Pt(x, y)
	p.at = &at
}

// Remove takes a placed element off the screen.
func (u *UI) Remove(e Element) {
	u.placed = slices.DeleteFunc(u.placed, func(p *placement) bool { return p.el == e })
}

// Open shows m on top of everything else. Opening a modal that is already
// open does nothing.
func (u *UI) Open(m *Modal) *Modal {
	if m.ui != nil {
		return m
	}
	m.ui = u
	u.modals = append(u.modals, m)
	u.pressed = nil
	u.Blur() // a field under the modal must not keep typing
	return m
}

// Focus gives f the keyboard, taking it from any other field. Call it after
// Open to start typing in a modal's field straight away.
func (u *UI) Focus(f *TextField) {
	if u.focused == f {
		return
	}
	u.Blur()
	if f != nil {
		u.focused = f
		f.focused = true
		f.caret = min(f.caret, len(f.text))
	}
}

// Blur takes the keyboard away from the focused field, if any. Anything the
// input method was still composing is kept.
func (u *UI) Blur() {
	if f := u.focused; f != nil {
		f.composer.Confirm()
		f.focused = false
		u.focused = nil
	}
}

// Focused returns the field that has the keyboard, or nil. While it is not
// nil, the game should leave the keyboard alone (no camera keys and so on).
func (u *UI) Focused() *TextField { return u.focused }

// Alert opens a modal with a title, a message and an OK button.
func (u *UI) Alert(title, message string) *Modal {
	return u.Open(NewModal(title, NewText(message)).AddButton("OK", nil))
}

// ModalOpen reports whether any modal is open.
func (u *UI) ModalOpen() bool { return len(u.modals) > 0 }

// CloseAll closes every open modal, top first.
func (u *UI) CloseAll() {
	for len(u.modals) > 0 {
		u.modals[len(u.modals)-1].Close()
	}
}

// SetScreenSize sets the screen size used for layout. Draw also sets it from
// the screen image, but Ebitengine can run Update before the first Draw and
// several times between draws, so call this from your game's Layout to keep
// clicks matched to what is on screen:
//
//	func (g *Game) Layout(int, int) (int, int) {
//		g.ui.SetScreenSize(screenW, screenH)
//		return screenW, screenH
//	}
func (u *UI) SetScreenSize(w, h int) { u.screen = image.Pt(w, h) }

// Update reads mouse, touch and keyboard input from Ebitengine and handles
// it, including typing into the focused text field. Call it at the start of
// your game's Update. It returns true when the UI used the input this frame
// (a modal is open, the pointer was pressed or released over a placed
// element, or a key went to the UI), meaning the game should ignore it.
func (u *UI) Update() bool { return u.HandleInput(u.readInput()) }

// UpdateKeys is the keyboard half of Update, for games that feed pointer
// events in themselves with HandleInput: it handles typing into the focused
// text field and the Escape and Enter keys. Call it once per Update. It
// returns true when a key went to the UI.
func (u *UI) UpdateKeys() bool {
	in := u.readInput()
	in.Pressed, in.Released = false, false
	u.layout()
	used := u.handleKeys(in)
	u.layout()
	return used
}

// readInput is ReadInput plus typing through the focused field's input
// method. Keys the input method used are cleared so they are not handled
// twice.
func (u *UI) readInput() Input {
	in := ReadInput()
	if f := u.focused; f != nil {
		if handled, err := f.composer.Update(); err == nil && handled {
			in.clearKeys()
			in.imeUsed = true
		}
	}
	return in
}

// HandleInput is Update with input you supply, for tests or custom controls.
// Input.Text is typed into the focused field.
func (u *UI) HandleInput(in Input) bool {
	u.cursor = in.Cursor
	u.layout()
	pointer := u.handlePointer(in) // first: a click can move focus
	keys := u.handleKeys(in)
	u.layout() // callbacks may have closed a modal or hidden the focused field
	return pointer || keys
}

func (u *UI) handlePointer(in Input) bool {
	if top := u.top(); top != nil {
		u.pointer(in, func(p image.Point) clickable { return top.panel.hit(&u.Theme, p) })
		return true
	}
	over := false
	for _, p := range u.placed {
		if !p.el.base().hidden && in.Cursor.In(p.el.base().rect) {
			over = true
		}
	}
	u.pointer(in, u.hitPlaced)
	return over && (in.Pressed || in.Released)
}

// handleKeys sends keys to the focused field, then to the top modal.
func (u *UI) handleKeys(in Input) bool {
	used := in.imeUsed
	if f := u.focused; f != nil {
		used = used || in.anyKey()
		f.edit(in)
		switch {
		case in.Escape:
			u.Blur() // and let a modal close too
		case in.Enter && f.onSubmit != nil:
			f.composer.Confirm()
			f.onSubmit(f.text)
			return true
		case !in.Enter:
			return used
		}
	}
	if top := u.top(); top != nil {
		switch {
		case in.Escape && top.dismissible:
			top.Close()
			return true
		case in.Enter:
			if b := top.firstButton(); b != nil {
				b.click()
			}
			return true
		}
	}
	return used
}

// pointer turns a press followed by a release over the same element into a
// click. A press on a text field focuses it; a press anywhere else takes the
// focus away.
func (u *UI) pointer(in Input, hit func(image.Point) clickable) {
	if in.Pressed {
		u.pressed = hit(in.Cursor)
		if f, ok := u.pressed.(*TextField); ok {
			u.Focus(f)
			f.placeCaret(&u.Theme, in.Cursor)
		} else {
			u.Blur()
		}
	}
	if in.Released {
		target := u.pressed
		u.pressed = nil
		if target != nil && hit(in.Cursor) == target {
			target.click()
		}
	}
}

func (u *UI) hitPlaced(pt image.Point) clickable {
	for i := len(u.placed) - 1; i >= 0; i-- {
		el := u.placed[i].el
		if el.base().hidden {
			continue
		}
		if c := el.hit(&u.Theme, pt); c != nil {
			return c
		}
	}
	return nil
}

// Draw draws the UI onto screen. Call it at the end of your game's Draw so
// the UI is on top.
func (u *UI) Draw(screen *ebiten.Image) {
	u.screen = screen.Bounds().Size()
	u.layout()
	th := &u.Theme

	// Only the top layer reacts to the pointer.
	off := &drawState{cursor: image.Pt(-1, -1)}
	on := &drawState{cursor: u.cursor, pressed: u.pressed}

	under := on
	if u.ModalOpen() {
		under = off
	}
	for _, p := range u.placed {
		if !p.el.base().hidden {
			p.el.draw(th, screen, under)
		}
	}
	for i, m := range u.modals {
		st := off
		if i == len(u.modals)-1 {
			fillRect(screen, screen.Bounds(), th.Backdrop)
			st = on
		}
		m.panel.draw(th, screen, st)
	}
}

// OverClickable reports whether the pointer is over something clickable that
// is not covered by a modal, e.g. to show a hand cursor:
//
//	if g.ui.OverClickable() {
//		ebiten.SetCursorShape(ebiten.CursorShapePointer)
//	} else {
//		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
//	}
func (u *UI) OverClickable() bool {
	if top := u.top(); top != nil {
		return top.panel.hit(&u.Theme, u.cursor) != nil
	}
	return u.hitPlaced(u.cursor) != nil
}

// CursorShape is the mouse cursor that fits what the pointer is over: a text
// cursor over a text field, a hand over buttons and clickable spans, the
// default elsewhere.
//
//	ebiten.SetCursorShape(g.ui.CursorShape())
func (u *UI) CursorShape() ebiten.CursorShapeType {
	var c clickable
	if top := u.top(); top != nil {
		c = top.panel.hit(&u.Theme, u.cursor)
	} else {
		c = u.hitPlaced(u.cursor)
	}
	switch c.(type) {
	case nil:
		return ebiten.CursorShapeDefault
	case *TextField:
		return ebiten.CursorShapeText
	default:
		return ebiten.CursorShapePointer
	}
}

func (u *UI) top() *Modal {
	if len(u.modals) == 0 {
		return nil
	}
	return u.modals[len(u.modals)-1]
}

// layout measures and positions every placed element and open modal. A
// focused field that is no longer on screen loses focus.
func (u *UI) layout() {
	th := &u.Theme
	f := u.focused
	if f != nil {
		f.arranged = false
	}
	defer func() {
		if f != nil && (!f.arranged || !f.focused) {
			u.Blur()
		}
	}()
	for _, p := range u.placed {
		if p.el.base().hidden {
			continue
		}
		var size, pos image.Point
		if p.at != nil {
			size = p.el.measure(th, u.screen.X)
			pos = image.Pt(clamp(p.at.X, 0, u.screen.X-size.X), clamp(p.at.Y, 0, u.screen.Y-size.Y))
		} else {
			size = p.el.measure(th, u.screen.X-2*p.marginX)
			pos = p.anchor.position(size, u.screen, p.marginX, p.marginY)
		}
		p.el.arrange(th, image.Rectangle{Min: pos, Max: pos.Add(size)})
	}
	for _, m := range u.modals {
		maxW := u.screen.X - 2*screenMargin
		if m.panel.width == 0 && th.ModalWidth > 0 && (maxW <= 0 || th.ModalWidth < maxW) {
			maxW = th.ModalWidth
		}
		size := m.panel.measure(th, maxW)
		pos := Middle.position(size, u.screen, 0, 0)
		m.panel.arrange(th, image.Rectangle{Min: pos, Max: pos.Add(size)})
	}
}
