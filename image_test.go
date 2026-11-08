package ui

import (
	"image"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

func TestImageSize(t *testing.T) {
	u := newTestUI()
	img := NewImage(ebiten.NewImage(10, 6))
	scaled := NewImage(img.Image()).Size(20, 12)
	none := NewImage(nil)
	label := NewText("x")
	u.Place(Row(img, scaled, none, label).Gap(0), TopLeft, 0, 0)
	u.HandleInput(Input{})

	if got := img.Bounds().Size(); got != image.Pt(10, 6) {
		t.Errorf("own size = %v, want 10x6", got)
	}
	if got := scaled.Bounds().Size(); got != image.Pt(20, 12) {
		t.Errorf("scaled size = %v, want 20x12", got)
	}
	if got := none.Bounds().Size(); got != (image.Point{}) {
		t.Errorf("nil image takes %v, want no space", got)
	}
	if label.Bounds().Min.X != 30 {
		t.Errorf("text starts at x=%d, want 30 (after both images)", label.Bounds().Min.X)
	}
}

func TestImageKeepsSizeInColumn(t *testing.T) {
	u := newTestUI()
	img := NewImage(ebiten.NewImage(8, 8))
	u.Place(NewPanel(img, NewText("a wider line of text")), TopLeft, 0, 0)
	u.HandleInput(Input{})
	if got := img.Bounds().Size(); got != image.Pt(8, 8) {
		t.Errorf("a stretching column resized the image to %v", got)
	}
}
