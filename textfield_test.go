package ui

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// fieldUI places a field in a panel and lays it out.
func fieldUI(t *testing.T, f *TextField) *UI {
	t.Helper()
	u := newTestUI()
	u.Place(NewPanel(f), TopLeft, 0, 0)
	u.HandleInput(Input{})
	return u
}

// focus clicks the field at its left edge.
func focus(u *UI, f *TextField) {
	p := image.Pt(f.Bounds().Min.X+1, (f.Bounds().Min.Y+f.Bounds().Max.Y)/2)
	u.HandleInput(Input{Cursor: p, Pressed: true})
	u.HandleInput(Input{Cursor: p, Released: true})
}

func TestFieldClickFocusesAndTypes(t *testing.T) {
	var changes []string
	f := NewTextField("name").OnChange(func(s string) { changes = append(changes, s) })
	u := fieldUI(t, f)

	if u.HandleInput(Input{Text: "x"}) || f.Text() != "" {
		t.Fatal("typing into an unfocused field should do nothing and be left to the game")
	}

	focus(u, f)
	if !f.Focused() || u.Focused() != f {
		t.Fatal("clicking the field should focus it")
	}
	if !u.HandleInput(Input{Text: "Guts"}) {
		t.Error("typing into a focused field should be reported as used")
	}
	if f.Text() != "Guts" {
		t.Fatalf("text = %q, want Guts", f.Text())
	}
	if len(changes) != 1 || changes[0] != "Guts" {
		t.Fatalf("OnChange calls = %q", changes)
	}

	// Moving the mouse while typing is not "used": the game keeps running.
	if u.HandleInput(Input{Cursor: image.Pt(300, 200)}) {
		t.Error("a plain mouse move should not be reported as used")
	}
}

func TestFieldEditingKeys(t *testing.T) {
	f := NewTextField("")
	u := fieldUI(t, f)
	u.Focus(f)

	steps := []struct {
		in   Input
		want string
		at   int // caret, in bytes
	}{
		{Input{Text: "héllo"}, "héllo", 6},
		{Input{Backspace: true}, "héll", 5},
		{Input{Home: true}, "héll", 0},
		{Input{Right: true}, "héll", 1},
		{Input{Right: true}, "héll", 3}, // over the two-byte é
		{Input{Backspace: true}, "hll", 1},
		{Input{Delete: true}, "hl", 1},
		{Input{Text: "e"}, "hel", 2},
		{Input{End: true}, "hel", 3},
		{Input{Left: true}, "hel", 2},
		{Input{Delete: true}, "he", 2},
		{Input{Delete: true}, "he", 2}, // nothing after the caret
	}
	for i, s := range steps {
		u.HandleInput(s.in)
		if f.Text() != s.want || f.caret != s.at {
			t.Fatalf("step %d: text %q caret %d, want %q caret %d", i, f.Text(), f.caret, s.want, s.at)
		}
	}
}

func TestFieldMaxLengthAndSingleLine(t *testing.T) {
	f := NewTextField("").MaxLength(4)
	u := fieldUI(t, f)
	u.Focus(f)
	u.HandleInput(Input{Text: "ab\ncdef"})
	if f.Text() != "abcd" {
		t.Fatalf("text = %q, want abcd (newline dropped, cut to 4)", f.Text())
	}
	u.HandleInput(Input{Text: "x"})
	if f.Text() != "abcd" {
		t.Fatalf("typed past MaxLength: %q", f.Text())
	}
	f.SetText("one\ntwo")
	if f.Text() != "onetwo" || f.caret != len("onetwo") {
		t.Fatalf("SetText = %q caret %d", f.Text(), f.caret)
	}
}

func TestFieldClickPlacesCaret(t *testing.T) {
	f := NewTextField("").SetText("abcdef")
	u := fieldUI(t, f)
	r := f.Bounds()
	mid := (r.Min.Y + r.Max.Y) / 2

	u.HandleInput(Input{Cursor: image.Pt(r.Max.X-2, mid), Pressed: true})
	if f.caret != len("abcdef") {
		t.Errorf("click at the right edge: caret %d, want end", f.caret)
	}
	u.HandleInput(Input{Cursor: image.Pt(r.Min.X+1, mid), Pressed: true})
	if f.caret != 0 {
		t.Errorf("click at the left edge: caret %d, want 0", f.caret)
	}
	x := f.inner(&u.Theme).Min.X + u.Theme.textWidth("abc")
	u.HandleInput(Input{Cursor: image.Pt(x, mid), Pressed: true})
	if f.caret != 3 {
		t.Errorf("click after \"abc\": caret %d, want 3", f.caret)
	}
}

func TestFieldEnterAndEscape(t *testing.T) {
	var submitted string
	f := NewTextField("").OnSubmit(func(s string) { submitted = s })
	u := fieldUI(t, f)
	u.Focus(f)
	u.HandleInput(Input{Text: "Casca"})
	u.HandleInput(Input{Enter: true})
	if submitted != "Casca" {
		t.Fatalf("OnSubmit got %q", submitted)
	}

	u.HandleInput(Input{Escape: true})
	if f.Focused() || u.Focused() != nil {
		t.Fatal("Escape should unfocus the field")
	}
}

func TestFieldClickElsewhereBlurs(t *testing.T) {
	f := NewTextField("")
	u := fieldUI(t, f)
	u.Focus(f)
	u.HandleInput(Input{Cursor: image.Pt(300, 200), Pressed: true})
	if f.Focused() {
		t.Fatal("pressing elsewhere should unfocus the field")
	}
}

func TestFieldInModal(t *testing.T) {
	u := newTestUI()
	f := NewTextField("name")
	ok := false
	m := u.Open(NewModal("Name", f).AddButton("OK", func() { ok = true }))
	u.Focus(f)
	u.HandleInput(Input{Text: "Griffith"})

	// Without OnSubmit, Enter presses the modal's first button.
	u.HandleInput(Input{Enter: true})
	if !ok || m.IsOpen() {
		t.Fatal("Enter in a field without OnSubmit should press OK")
	}
	if f.Focused() || u.Focused() != nil {
		t.Fatal("a field in a closed modal should lose focus")
	}

	// Escape unfocuses and closes the modal in one press.
	u.Open(m)
	u.Focus(f)
	u.HandleInput(Input{Escape: true})
	if m.IsOpen() || f.Focused() {
		t.Fatal("Escape should unfocus the field and close the modal")
	}
}

func TestOpenModalBlursHUDField(t *testing.T) {
	f := NewTextField("")
	u := fieldUI(t, f)
	u.Focus(f)
	u.Alert("Hi", "there")
	if f.Focused() {
		t.Fatal("opening a modal should unfocus a field behind it")
	}
}

func TestHiddenFieldLosesFocus(t *testing.T) {
	f := NewTextField("")
	u := fieldUI(t, f)
	u.Focus(f)
	f.SetVisible(false)
	u.HandleInput(Input{})
	if f.Focused() || u.Focused() != nil {
		t.Fatal("a hidden field should lose focus")
	}
}

func TestCursorShape(t *testing.T) {
	f := NewTextField("")
	b := NewButton("Go", nil)
	u := newTestUI()
	u.Place(NewPanel(f, b), TopLeft, 0, 0)
	u.HandleInput(Input{})

	cases := []struct {
		at   image.Point
		want ebiten.CursorShapeType
	}{
		{f.Bounds().Min.Add(image.Pt(2, 2)), ebiten.CursorShapeText},
		{b.Bounds().Min.Add(image.Pt(2, 2)), ebiten.CursorShapePointer},
		{image.Pt(300, 200), ebiten.CursorShapeDefault},
	}
	for _, c := range cases {
		u.HandleInput(Input{Cursor: c.at})
		if got := u.CursorShape(); got != c.want {
			t.Errorf("at %v: cursor %v, want %v", c.at, got, c.want)
		}
	}
}
