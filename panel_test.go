package ui

import "testing"

func TestPanelRadiusAndBorderWidth(t *testing.T) {
	th := DefaultTheme()
	th.PanelRadius, th.PanelBorderWidth = 6, 2

	p := NewPanel(NewText("x"))
	if r, w := orTheme(p.radius, th.PanelRadius), orTheme(p.borderW, th.PanelBorderWidth); r != 6 || w != 2 {
		t.Errorf("unset panel uses radius %d, border %d; want the theme's 6 and 2", r, w)
	}
	p.Radius(0).BorderWidth(0)
	if r, w := orTheme(p.radius, th.PanelRadius), orTheme(p.borderW, th.PanelBorderWidth); r != 0 || w != 0 {
		t.Errorf("panel set to 0 uses radius %d, border %d; 0 must override the theme", r, w)
	}
}

func TestPanelStyleKeepsLayout(t *testing.T) {
	u := newTestUI()
	plain := NewPanel(NewText("hello"))
	round := NewPanel(NewText("hello")).Radius(12).BorderWidth(3)
	u.Place(plain, TopLeft, 0, 0)
	u.Place(round, TopLeft, 0, 100)
	u.HandleInput(Input{})
	if plain.Bounds().Size() != round.Bounds().Size() {
		t.Errorf("radius and border width changed the size: %v vs %v", plain.Bounds().Size(), round.Bounds().Size())
	}
}
