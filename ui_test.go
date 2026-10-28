package ui

import (
	"image"
	"strings"
	"testing"
)

func newTestUI() *UI {
	u := New()
	u.SetScreenSize(320, 240)
	return u
}

// click presses and releases the pointer at the center of r.
func click(u *UI, r image.Rectangle) bool {
	p := image.Pt((r.Min.X+r.Max.X)/2, (r.Min.Y+r.Max.Y)/2)
	a := u.HandleInput(Input{Cursor: p, Pressed: true})
	b := u.HandleInput(Input{Cursor: p, Released: true})
	return a || b
}

// lineTexts lays out spans and returns each line's text.
func lineTexts(th *Theme, maxW int, spans ...*Span) []string {
	var out []string
	for _, l := range layoutText(th, spans, maxW) {
		var b strings.Builder
		for _, r := range l.runs {
			b.WriteString(r.text)
		}
		out = append(out, b.String())
	}
	return out
}

func TestLayoutWraps(t *testing.T) {
	th := DefaultTheme()
	w := th.textWidth("aaa bbb")
	got := lineTexts(&th, w, NewSpan("aaa   bbb ccc\nddd"))
	want := []string{"aaa bbb", "ccc", "ddd"}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("lines = %q, want %q", got, want)
	}
	if got := lineTexts(&th, 0, NewSpan("a b\nc")); len(got) != 2 {
		t.Fatalf("no limit should only break at newlines, got %q", got)
	}
	if got := lineTexts(&th, 0, NewSpan("")); len(got) != 1 {
		t.Fatalf("empty text should be one empty line, got %q", got)
	}
}

func TestLayoutAcrossSpans(t *testing.T) {
	th := DefaultTheme()
	// "key" and "." are separate spans but one word: they must not split.
	w := th.textWidth("a rusty key")
	got := lineTexts(&th, w, NewSpan("a rusty "), NewSpan("key"), NewSpan("."))
	want := []string{"a rusty", "key."}
	if strings.Join(got, "|") != strings.Join(want, "|") {
		t.Fatalf("lines = %q, want %q", got, want)
	}

	// Neighbouring pieces of one span merge into one run, spaces included.
	key := NewSpan("rusty old key")
	lines := layoutText(&th, []*Span{NewSpan("a "), key, NewSpan(" here")}, 0)
	if n := len(lines[0].runs); n != 3 {
		t.Fatalf("runs = %d, want 3 (plain, key, plain)", n)
	}
	if r := lines[0].runs[1]; r.span != key || r.text != "rusty old key" {
		t.Fatalf("middle run = %q, want the whole key span", r.text)
	}
}

func TestClickableSpan(t *testing.T) {
	u := newTestUI()
	n := 0
	key := NewSpan("key").OnClick(func() { n++ })
	txt := NewRichText(NewSpan("You found a "), key, NewSpan("."))
	u.Place(NewPanel(txt), TopLeft, 0, 0)
	u.HandleInput(Input{})

	var keyRect image.Rectangle
	for _, r := range txt.lines[0].runs {
		if r.span == key {
			keyRect = txt.runRect(&u.Theme, 0, r)
		}
	}
	if keyRect.Empty() {
		t.Fatal("key span was not laid out")
	}

	u.HandleInput(Input{Cursor: keyRect.Min.Add(image.Pt(1, 1))})
	if !u.OverClickable() {
		t.Error("pointer over the key should be over something clickable")
	}
	if !click(u, keyRect) || n != 1 {
		t.Fatalf("clicking the span: clicks = %d, want 1", n)
	}

	// Plain words around it do nothing.
	plain := image.Rect(txt.Bounds().Min.X, txt.Bounds().Min.Y, txt.Bounds().Min.X+4, txt.Bounds().Min.Y+4)
	click(u, plain)
	if n != 1 {
		t.Fatal("clicking plain text fired the span")
	}
	u.HandleInput(Input{Cursor: plain.Min})
	if u.OverClickable() {
		t.Error("plain text should not count as clickable")
	}
}

func TestSetTextOnRichText(t *testing.T) {
	txt := NewRichText(NewSpan("a"), NewSpan("b").OnClick(func() {}))
	txt.SetText("plain")
	if len(txt.Spans()) != 1 || txt.Text() != "plain" || txt.Spans()[0].onClick != nil {
		t.Fatalf("SetText should replace rich spans with one plain span, got %d spans", len(txt.Spans()))
	}
}

func TestTextWrapsToWidth(t *testing.T) {
	th := DefaultTheme()
	txt := NewText(strings.Repeat("word ", 40))
	one := txt.measure(&th, 0)
	wrapped := txt.measure(&th, 100)
	if wrapped.X > 100 {
		t.Fatalf("wrapped width %d > 100", wrapped.X)
	}
	if wrapped.Y <= one.Y {
		t.Fatalf("wrapped text should be taller: %d vs %d", wrapped.Y, one.Y)
	}
}

func TestPlaceAnchors(t *testing.T) {
	u := newTestUI()
	tl := NewPanel(NewText("top left"))
	br := NewPanel(NewText("bottom right"))
	mid := NewPanel(NewText("middle"))
	u.Place(tl, TopLeft, 4, 6)
	u.Place(br, BottomRight, 4, 6)
	u.Place(mid, Middle, 0, 0)
	u.HandleInput(Input{})

	if got := tl.Bounds().Min; got != image.Pt(4, 6) {
		t.Errorf("TopLeft at %v, want (4,6)", got)
	}
	if got := br.Bounds().Max; got != image.Pt(316, 234) {
		t.Errorf("BottomRight ends at %v, want (316,234)", got)
	}
	r := mid.Bounds()
	if c := (r.Min.X + r.Max.X) / 2; c < 159 || c > 161 {
		t.Errorf("Middle centered at x=%d, want ~160", c)
	}
}

func TestButtonClick(t *testing.T) {
	u := newTestUI()
	n := 0
	b := NewButton("Go", func() { n++ })
	u.Place(NewPanel(b), TopLeft, 0, 0)
	u.HandleInput(Input{})

	if !click(u, b.Bounds()) {
		t.Error("click on a placed button should be reported as used")
	}
	if n != 1 {
		t.Fatalf("clicks = %d, want 1", n)
	}

	// Pressing on the button and releasing elsewhere is not a click.
	in := b.Bounds().Min.Add(image.Pt(2, 2))
	u.HandleInput(Input{Cursor: in, Pressed: true})
	u.HandleInput(Input{Cursor: image.Pt(300, 200), Released: true})
	if n != 1 {
		t.Fatalf("drag-off fired the button")
	}

	b.SetDisabled(true)
	click(u, b.Bounds())
	if n != 1 {
		t.Fatalf("disabled button fired")
	}
}

func TestClickOutsideIsNotUsed(t *testing.T) {
	u := newTestUI()
	u.Place(NewPanel(NewText("hud")), TopLeft, 0, 0)
	u.HandleInput(Input{})
	if click(u, image.Rect(200, 200, 202, 202)) {
		t.Fatal("click on empty screen should be left to the game")
	}
}

func TestModalBlocksInputAndCloses(t *testing.T) {
	u := newTestUI()
	n := 0
	behind := NewButton("Behind", func() { n++ })
	u.Place(behind, TopLeft, 0, 0)

	closed := false
	m := u.Open(NewModal("Title", NewText("Body")).OnClose(func() { closed = true }))
	u.HandleInput(Input{})

	if !click(u, behind.Bounds()) {
		t.Error("input should be used while a modal is open")
	}
	if n != 0 {
		t.Fatal("button behind a modal was clicked")
	}

	u.HandleInput(Input{Escape: true})
	if m.IsOpen() || u.ModalOpen() || !closed {
		t.Fatal("Escape should close the modal and call OnClose")
	}

	click(u, behind.Bounds())
	if n != 1 {
		t.Fatal("button should work again once the modal is closed")
	}
}

func TestModalNotDismissible(t *testing.T) {
	u := newTestUI()
	m := u.Open(NewModal("Stay", NewText("Body")).Dismissible(false))
	u.HandleInput(Input{Escape: true})
	if !m.IsOpen() {
		t.Fatal("Escape closed a non-dismissible modal")
	}
}

func TestModalButtons(t *testing.T) {
	u := newTestUI()
	yes := false
	m := u.Open(NewModal("Q", NewText("Sure?")).
		AddButton("Yes", func() { yes = true }).
		AddButton("No", nil))
	u.HandleInput(Input{})

	no := m.buttons.children[1].(*Button)
	click(u, no.Bounds())
	if yes || m.IsOpen() {
		t.Fatal("No should close without running Yes")
	}

	u.Open(m)
	u.HandleInput(Input{Enter: true})
	if !yes || m.IsOpen() {
		t.Fatal("Enter should press the first button and close the modal")
	}
}

func TestModalStack(t *testing.T) {
	u := newTestUI()
	var second *Modal
	first := u.Open(NewModal("First", NewText("1")).
		AddButton("Next", func() { second = u.Alert("Second", "2") }))
	u.HandleInput(Input{Enter: true})

	if first.IsOpen() || second == nil || !second.IsOpen() {
		t.Fatal("a button that opens another modal should leave only the new one open")
	}
	u.CloseAll()
	if u.ModalOpen() {
		t.Fatal("CloseAll left a modal open")
	}
}

func TestModalLayout(t *testing.T) {
	u := newTestUI()
	m := u.Alert("Long", strings.Repeat("lots of words ", 60))
	u.HandleInput(Input{})
	r := m.panel.Bounds()
	if r.Dx() > u.Theme.ModalWidth {
		t.Errorf("modal width %d > ModalWidth %d", r.Dx(), u.Theme.ModalWidth)
	}
	if c := (r.Min.X + r.Max.X) / 2; c < 159 || c > 161 {
		t.Errorf("modal centered at x=%d, want ~160", c)
	}

	// A narrow screen limits the modal too.
	u.SetScreenSize(120, 240)
	u.HandleInput(Input{})
	if r := m.panel.Bounds(); r.Min.X < 0 || r.Max.X > 120 {
		t.Errorf("modal %v does not fit a 120px screen", r)
	}
}

func TestHiddenTakesNoSpace(t *testing.T) {
	u := newTestUI()
	a, b := NewText("a"), NewText("b")
	p := NewPanel(a, b)
	u.Place(p, TopLeft, 0, 0)
	u.HandleInput(Input{})
	full := p.Bounds().Dy()

	b.SetVisible(false)
	u.HandleInput(Input{})
	if got := p.Bounds().Dy(); got >= full {
		t.Fatalf("hiding a child should shrink the panel: %d vs %d", got, full)
	}
}

func TestPanelWidth(t *testing.T) {
	u := newTestUI()
	txt := NewText(strings.Repeat("wrap me ", 30))
	p := NewPanel(txt).Width(150)
	u.Place(p, TopLeft, 0, 0)
	u.HandleInput(Input{})
	if got := p.Bounds().Dx(); got != 150 {
		t.Fatalf("panel width %d, want 150", got)
	}
	if txt.Bounds().Max.X > p.Bounds().Max.X {
		t.Fatal("text overflows its panel")
	}
}

func TestRowJustifyEnd(t *testing.T) {
	u := newTestUI()
	b := NewButton("OK", nil)
	row := Row(b).Justify(End)
	p := NewPanel(NewText(strings.Repeat("x", 30)), row)
	u.Place(p, TopLeft, 0, 0)
	u.HandleInput(Input{})
	if b.Bounds().Max.X != row.Bounds().Max.X {
		t.Fatalf("button ends at %d, row at %d", b.Bounds().Max.X, row.Bounds().Max.X)
	}
}
