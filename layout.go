package ui

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Stack lines up its children in a column (top to bottom) or a row (left to
// right). Create one with Column or Row.
type Stack struct {
	node
	vertical bool
	children []Element
	gap      int // < 0 means the theme's Gap
	width    int // fixed width; 0 means fit the content
	align    Align
	justify  Align
}

// Column stacks children top to bottom. By default each child is stretched to
// the column's width.
func Column(children ...Element) *Stack {
	return &Stack{vertical: true, children: children, gap: -1, align: Stretch}
}

// Row places children left to right, centered vertically.
func Row(children ...Element) *Stack {
	return &Stack{children: children, gap: -1, align: Center}
}

// Add appends children.
func (s *Stack) Add(children ...Element) *Stack {
	s.children = append(s.children, children...)
	return s
}

// Clear removes all children.
func (s *Stack) Clear() *Stack { s.children = nil; return s }

// Children returns the stack's children.
func (s *Stack) Children() []Element { return s.children }

// Gap sets the space between children, overriding the theme.
func (s *Stack) Gap(n int) *Stack { s.gap = n; return s }

// Width fixes the stack's width; text inside wraps to fit.
func (s *Stack) Width(n int) *Stack { s.width = n; return s }

// Align sets how children are placed across the stack: horizontally in a
// column, vertically in a row.
func (s *Stack) Align(a Align) *Stack { s.align = a; return s }

// Justify sets where children are placed along the stack when there is spare
// room, e.g. Justify(End) pushes a row's children to the right.
func (s *Stack) Justify(a Align) *Stack { s.justify = a; return s }

func (s *Stack) gapFor(th *Theme) int {
	if s.gap < 0 {
		return th.Gap
	}
	return s.gap
}

// innerWidth is the width the stack may use, given what its parent offers.
func (s *Stack) innerWidth(maxW int) int {
	if s.width > 0 && (maxW <= 0 || s.width < maxW) {
		return s.width
	}
	return maxW
}

// sizes measures the visible children. For a row, each child may use what is
// left of the width after the children before it.
func (s *Stack) sizes(th *Theme, maxW int) []image.Point {
	gap := s.gapFor(th)
	sizes := make([]image.Point, len(s.children))
	used := 0
	for i, c := range s.children {
		if c.base().hidden {
			continue
		}
		avail := maxW
		if !s.vertical && maxW > 0 {
			avail = max(maxW-used, 1)
		}
		sizes[i] = c.measure(th, avail)
		used += sizes[i].X + gap
	}
	return sizes
}

func (s *Stack) measure(th *Theme, maxW int) image.Point {
	inner := s.innerWidth(maxW)
	gap := s.gapFor(th)
	var main, cross, n int
	for i, sz := range s.sizes(th, inner) {
		if s.children[i].base().hidden {
			continue
		}
		if n > 0 {
			main += gap
		}
		n++
		if s.vertical {
			main += sz.Y
			cross = max(cross, sz.X)
		} else {
			main += sz.X
			cross = max(cross, sz.Y)
		}
	}
	if s.vertical {
		if s.width > 0 {
			cross = inner
		}
		return image.Pt(cross, main)
	}
	if s.width > 0 {
		main = max(main, inner)
	}
	return image.Pt(main, cross)
}

func (s *Stack) arrange(th *Theme, r image.Rectangle) {
	s.rect = r
	gap := s.gapFor(th)
	sizes := s.sizes(th, r.Dx())

	total, n := 0, 0
	for i, sz := range sizes {
		if s.children[i].base().hidden {
			continue
		}
		if n > 0 {
			total += gap
		}
		n++
		if s.vertical {
			total += sz.Y
		} else {
			total += sz.X
		}
	}

	if s.vertical {
		y := r.Min.Y + s.justify.offset(r.Dy(), total)
		for i, c := range s.children {
			if c.base().hidden {
				continue
			}
			sz := sizes[i]
			x, w := r.Min.X, r.Dx()
			if s.align != Stretch {
				x, w = r.Min.X+s.align.offset(r.Dx(), sz.X), sz.X
			}
			c.arrange(th, image.Rect(x, y, x+w, y+sz.Y))
			y += sz.Y + gap
		}
		return
	}

	x := r.Min.X + s.justify.offset(r.Dx(), total)
	for i, c := range s.children {
		if c.base().hidden {
			continue
		}
		sz := sizes[i]
		y, h := r.Min.Y, r.Dy()
		if s.align != Stretch {
			y, h = r.Min.Y+s.align.offset(r.Dy(), sz.Y), sz.Y
		}
		c.arrange(th, image.Rect(x, y, x+sz.X, y+h))
		x += sz.X + gap
	}
}

func (s *Stack) draw(th *Theme, dst *ebiten.Image, st *drawState) {
	for _, c := range s.children {
		if !c.base().hidden {
			c.draw(th, dst, st)
		}
	}
}

func (s *Stack) hit(th *Theme, p image.Point) clickable {
	for i := len(s.children) - 1; i >= 0; i-- {
		c := s.children[i]
		if c.base().hidden {
			continue
		}
		if h := c.hit(th, p); h != nil {
			return h
		}
	}
	return nil
}

// Panel is a box with a background and border that stacks its children in a
// column.
type Panel struct {
	node
	body       *Stack
	width      int
	pad        int // < 0 means the theme's Padding
	radius     int // < 0 means the theme's PanelRadius
	borderW    int // < 0 means the theme's PanelBorderWidth
	background color.Color
	border     color.Color
}

// NewPanel creates a panel holding children top to bottom.
func NewPanel(children ...Element) *Panel {
	return &Panel{body: Column(children...), pad: -1, radius: -1, borderW: -1}
}

// Add appends children.
func (p *Panel) Add(children ...Element) *Panel { p.body.Add(children...); return p }

// Clear removes all children.
func (p *Panel) Clear() *Panel { p.body.Clear(); return p }

// Children returns the panel's children.
func (p *Panel) Children() []Element { return p.body.Children() }

// Width fixes the panel's outer width; text inside wraps to fit.
func (p *Panel) Width(n int) *Panel { p.width = n; return p }

// Gap sets the space between children, overriding the theme.
func (p *Panel) Gap(n int) *Panel { p.body.Gap(n); return p }

// Align sets how children are placed horizontally. The default, Stretch, makes
// every child as wide as the panel.
func (p *Panel) Align(a Align) *Panel { p.body.Align(a); return p }

// Padding sets the space between the panel's edge and its children.
func (p *Panel) Padding(n int) *Panel { p.pad = n; return p }

// Background sets the fill color, overriding the theme.
func (p *Panel) Background(c color.Color) *Panel { p.background = c; return p }

// Border sets the border color, overriding the theme.
func (p *Panel) Border(c color.Color) *Panel { p.border = c; return p }

// Radius rounds the panel's corners by n pixels, overriding the theme. 0 is square.
func (p *Panel) Radius(n int) *Panel { p.radius = n; return p }

// BorderWidth sets how thick the border is in pixels, overriding the theme. 0 draws
// no border. The border is drawn inside the panel, over its padding.
func (p *Panel) BorderWidth(n int) *Panel { p.borderW = n; return p }

func (p *Panel) padFor(th *Theme) int { return orTheme(p.pad, th.Padding) }

// orTheme is v, or the theme's value when v is unset (< 0).
func orTheme(v, theme int) int {
	if v < 0 {
		return theme
	}
	return v
}

func (p *Panel) measure(th *Theme, maxW int) image.Point {
	pad := p.padFor(th)
	outer := maxW
	if p.width > 0 && (maxW <= 0 || p.width < maxW) {
		outer = p.width
	}
	inner := 0
	if outer > 0 {
		inner = max(outer-2*pad, 1)
	}
	sz := p.body.measure(th, inner)
	w := sz.X + 2*pad
	if p.width > 0 {
		w = max(w, outer)
	}
	return image.Pt(w, sz.Y+2*pad)
}

func (p *Panel) arrange(th *Theme, r image.Rectangle) {
	p.rect = r
	pad := p.padFor(th)
	p.body.arrange(th, r.Inset(pad))
}

func (p *Panel) draw(th *Theme, dst *ebiten.Image, st *drawState) {
	bg, border := p.background, p.border
	if bg == nil {
		bg = th.PanelBackground
	}
	if border == nil {
		border = th.PanelBorder
	}
	radius, bw := orTheme(p.radius, th.PanelRadius), orTheme(p.borderW, th.PanelBorderWidth)
	if radius == 0 && bw <= 1 { // the plain box: pixel-exact, as before
		fillRect(dst, p.rect, bg)
		if bw == 1 {
			strokeRect(dst, p.rect, border)
		}
	} else {
		fillRoundRect(dst, p.rect, radius, bg)
		if bw > 0 {
			shade, shine := th.PanelBorderShade, th.PanelBorderShine
			if border != th.PanelBorder { // the depth is made for the theme's border colour
				shade, shine = nil, nil
			}
			strokeBevel(dst, p.rect, radius, bw, border, shade, shine)
		}
	}
	p.body.draw(th, dst, st)
}

func (p *Panel) hit(th *Theme, pt image.Point) clickable { return p.body.hit(th, pt) }

// roundRect is a path around r inset by inset pixels, with corners of the given
// radius, shrunk so it never exceeds half the shorter side.
func roundRect(r image.Rectangle, radius, inset float32) *vector.Path {
	x0, y0 := float32(r.Min.X)+inset, float32(r.Min.Y)+inset
	x1, y1 := float32(r.Max.X)-inset, float32(r.Max.Y)-inset
	rad := max(min(radius-inset, (x1-x0)/2, (y1-y0)/2), 0)
	var p vector.Path
	p.MoveTo(x0+rad, y0)
	p.LineTo(x1-rad, y0)
	p.ArcTo(x1, y0, x1, y0+rad, rad)
	p.LineTo(x1, y1-rad)
	p.ArcTo(x1, y1, x1-rad, y1, rad)
	p.LineTo(x0+rad, y1)
	p.ArcTo(x0, y1, x0, y1-rad, rad)
	p.LineTo(x0, y0+rad)
	p.ArcTo(x0, y0, x0+rad, y0, rad)
	p.Close()
	return &p
}

func fillRoundRect(dst *ebiten.Image, r image.Rectangle, radius int, c color.Color) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.FillPath(dst, roundRect(r, float32(radius), 0), nil, op)
}

// strokeRoundRect draws a border w pixels wide just inside r, following its rounded corners.
func strokeRoundRect(dst *ebiten.Image, r image.Rectangle, radius, w int, c color.Color) {
	op := &vector.DrawPathOptions{AntiAlias: true}
	op.ColorScale.ScaleWithColor(c)
	vector.StrokePath(dst, roundRect(r, float32(radius), float32(w)/2), &vector.StrokeOptions{Width: float32(w)}, op)
}

// strokeBevel is strokeRoundRect with depth: a shade under it and a shine along its top,
// each left out when nil.
func strokeBevel(dst *ebiten.Image, r image.Rectangle, radius, w int, c, shade, shine color.Color) {
	path := roundRect(r, float32(radius), float32(w)/2)
	stroke := func(dy, width float32, c color.Color) {
		if c == nil {
			return
		}
		var p vector.Path
		o := &vector.AddPathOptions{}
		o.GeoM.Translate(0, float64(dy))
		p.AddPath(path, o)
		op := &vector.DrawPathOptions{AntiAlias: true}
		op.ColorScale.ScaleWithColor(c)
		vector.StrokePath(dst, &p, &vector.StrokeOptions{Width: width}, op)
	}
	fw := float32(w)
	stroke(1.5, fw+1.5, shade)
	stroke(0, fw, c)
	stroke(-fw/4, max(1, fw*0.4), shine)
}

func fillRect(dst *ebiten.Image, r image.Rectangle, c color.Color) {
	vector.FillRect(dst, float32(r.Min.X), float32(r.Min.Y), float32(r.Dx()), float32(r.Dy()), c, false)
}

// strokeRect draws a 1px border just inside r.
func strokeRect(dst *ebiten.Image, r image.Rectangle, c color.Color) {
	vector.StrokeRect(dst, float32(r.Min.X)+0.5, float32(r.Min.Y)+0.5, float32(r.Dx()-1), float32(r.Dy()-1), 1, c, false)
}
