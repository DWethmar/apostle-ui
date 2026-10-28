package ui

// Modal is a dialog shown in the middle of the screen above everything else.
// While a modal is open it takes all input: the game and the elements behind
// it receive no clicks.
//
// Escape closes the top modal (unless Dismissible(false)), and Enter presses
// its first button.
//
//	u.Open(ui.NewModal("Leave the cave?", ui.NewText("It is dark outside.")).
//		AddButton("Yes", leave).
//		AddButton("No", nil))
type Modal struct {
	ui          *UI
	panel       *Panel
	title       *Text
	body        *Stack
	buttons     *Stack
	dismissible bool
	onClose     func()
}

// NewModal creates a modal with a title and body elements. Pass "" for no
// title. Show it with UI.Open.
func NewModal(title string, children ...Element) *Modal {
	m := &Modal{dismissible: true}
	m.title = NewText(title)
	m.title.title = true
	m.title.SetVisible(title != "")
	m.body = Column(children...)
	m.buttons = Row().Justify(End)
	m.buttons.SetVisible(false)
	m.panel = NewPanel(m.title, m.body, m.buttons).Gap(10)
	return m
}

// Add appends elements to the modal's body.
func (m *Modal) Add(children ...Element) *Modal { m.body.Add(children...); return m }

// AddButton adds a button to the bottom row. Clicking it runs onClick (which
// may be nil) and then closes the modal.
func (m *Modal) AddButton(label string, onClick func()) *Modal {
	m.buttons.Add(NewButton(label, func() {
		if onClick != nil {
			onClick()
		}
		m.Close()
	}))
	m.buttons.SetVisible(true)
	return m
}

// SetTitle changes the title.
func (m *Modal) SetTitle(s string) *Modal {
	m.title.SetText(s)
	m.title.SetVisible(s != "")
	return m
}

// Width fixes the modal's width instead of fitting it to its content.
func (m *Modal) Width(n int) *Modal { m.panel.Width(n); return m }

// Dismissible sets whether Escape closes the modal. Defaults to true.
func (m *Modal) Dismissible(d bool) *Modal { m.dismissible = d; return m }

// OnClose sets a callback that runs whenever the modal closes, however it was
// closed.
func (m *Modal) OnClose(fn func()) *Modal { m.onClose = fn; return m }

// IsOpen reports whether the modal is currently shown.
func (m *Modal) IsOpen() bool { return m.ui != nil }

// Close hides the modal. Closing a modal that is not open does nothing.
func (m *Modal) Close() {
	u := m.ui
	if u == nil {
		return
	}
	for i, o := range u.modals {
		if o == m {
			u.modals = append(u.modals[:i], u.modals[i+1:]...)
			break
		}
	}
	m.ui = nil
	u.pressed = nil
	if m.onClose != nil {
		m.onClose()
	}
}

// firstButton returns the first enabled bottom-row button, or nil.
func (m *Modal) firstButton() *Button {
	for _, c := range m.buttons.children {
		if b, ok := c.(*Button); ok && !b.disabled && !b.hidden {
			return b
		}
	}
	return nil
}
