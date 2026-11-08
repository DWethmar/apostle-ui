package ui

import (
	"image"

	"github.com/hajimehoshi/ebiten/v2"
)

// Image shows an *ebiten.Image, such as an icon next to some text:
//
//	ui.Row(ui.NewImage(coinIcon).Size(16, 16), ui.NewText("12 gold"))
type Image struct {
	node
	img  *ebiten.Image
	size image.Point // 0 means the image's own size
}

// NewImage creates an element showing img at its own size. img may be nil,
// which shows nothing and takes no space.
func NewImage(img *ebiten.Image) *Image { return &Image{img: img} }

// SetImage changes the image. Safe to call every frame.
func (i *Image) SetImage(img *ebiten.Image) *Image { i.img = img; return i }

// Image returns the image shown, or nil.
func (i *Image) Image() *ebiten.Image { return i.img }

// Size scales the image to w x h pixels. 0, 0 goes back to its own size.
func (i *Image) Size(w, h int) *Image { i.size = image.Pt(w, h); return i }

func (i *Image) measure(_ *Theme, _ int) image.Point {
	switch {
	case i.img == nil:
		return image.Point{}
	case i.size != image.Point{}:
		return i.size
	default:
		return i.img.Bounds().Size()
	}
}

func (i *Image) arrange(th *Theme, r image.Rectangle) {
	// Keep its own size even when a column stretches it; draw it at the start.
	i.rect = image.Rectangle{Min: r.Min, Max: r.Min.Add(i.measure(th, 0))}
}

func (i *Image) draw(_ *Theme, dst *ebiten.Image, _ *drawState) {
	if i.img == nil || i.rect.Empty() {
		return
	}
	src := i.img.Bounds().Size()
	op := &ebiten.DrawImageOptions{}
	if sz := i.rect.Size(); sz != src {
		op.GeoM.Scale(float64(sz.X)/float64(src.X), float64(sz.Y)/float64(src.Y))
		op.Filter = ebiten.FilterLinear
	}
	op.GeoM.Translate(float64(i.rect.Min.X), float64(i.rect.Min.Y))
	dst.DrawImage(i.img, op)
}

func (i *Image) hit(*Theme, image.Point) clickable { return nil }
