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
	pressed clickable // what the current press started on
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
	return m
}

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
// it. Call it at the start of your game's Update. It returns true when the UI
// used the input (a modal is open, or the pointer was pressed or released
// over a placed element), meaning the game should ignore it this frame.
func (u *UI) Update() bool { return u.HandleInput(ReadInput()) }

// HandleInput is Update with input you supply, for tests or custom controls.
func (u *UI) HandleInput(in Input) bool {
	u.cursor = in.Cursor
	u.layout()

	if top := u.top(); top != nil {
		switch {
		case in.Escape && top.dismissible:
			top.Close()
		case in.Enter:
			if b := top.firstButton(); b != nil {
				b.click()
			}
		default:
			u.pointer(in, func(p image.Point) clickable { return top.panel.hit(&u.Theme, p) })
		}
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

// pointer turns a press followed by a release over the same element into a
// click.
func (u *UI) pointer(in Input, hit func(image.Point) clickable) {
	if in.Pressed {
		u.pressed = hit(in.Cursor)
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

func (u *UI) top() *Modal {
	if len(u.modals) == 0 {
		return nil
	}
	return u.modals[len(u.modals)-1]
}

// layout measures and positions every placed element and open modal.
func (u *UI) layout() {
	th := &u.Theme
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
