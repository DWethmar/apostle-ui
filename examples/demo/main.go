// Demo shows the ui package: a HUD panel, a menu, alerts and a yes/no modal.
//
//	go run ./examples/demo
package main

import (
	"image/color"
	"log"
	"strconv"

	"github.com/hajimehoshi/ebiten/v2"

	ui "github.com/dwethmar/apostle-ui"
)

const screenW, screenH = 480, 270

var (
	goldColor = color.RGBA{0xf2, 0xc9, 0x4c, 0xff}
	bloodRed  = color.RGBA{0xff, 0x6b, 0x6b, 0xff}
	darkRed   = color.RGBA{0x5a, 0x16, 0x16, 0xff}
)

type game struct {
	ui   *ui.UI
	gold *ui.Span
	coin int
}

func newGame() *game {
	g := &game{ui: ui.New()}
	u := g.ui

	// A HUD panel; the gold amount is its own span so it can have a color.
	g.gold = ui.NewSpan("0").Color(goldColor)
	u.Place(ui.NewPanel(
		ui.NewRichText(ui.NewSpan("Gold: "), g.gold),
		ui.NewText("Day 1").Color(color.RGBA{0x9a, 0xc8, 0xff, 0xff}),
	), ui.TopLeft, 8, 8)

	// A menu of buttons.
	u.Place(ui.NewPanel(
		ui.NewText("Menu").Align(ui.Center),
		ui.NewButton("Show message", g.showNote),
		ui.NewButton("Leave the cave", g.askToLeave),
		ui.NewButton("Find gold", func() { g.coin += 10 }),
	).Width(140), ui.TopRight, 8, 8)

	u.Place(ui.NewText("Esc closes a modal, Enter presses its first button."), ui.Bottom, 0, 8)
	return g
}

// showNote shows rich text: a highlighted word and a clickable one.
func (g *game) showNote() {
	g.ui.Open(ui.NewModal("A note", ui.NewRichText(
		ui.NewSpan("Someone scratched words into the wall: \"The "),
		ui.NewSpan("eclipse").Color(bloodRed).Highlight(darkRed),
		ui.NewSpan(" comes for those who seek it.\" The letters are old, but the "),
		ui.NewSpan("blood").OnClick(func() {
			g.ui.Alert("Blood", "Still wet. Whoever wrote this is close.")
		}),
		ui.NewSpan(" is not."),
	)).AddButton("OK", nil))
}

func (g *game) askToLeave() {
	g.ui.Open(ui.NewModal("Leave the cave?",
		ui.NewText("It is dark outside and you hear wolves. You can come back later."),
	).
		AddButton("Leave", func() { g.ui.Alert("Outside", "The night is cold.") }).
		AddButton("Stay", nil))
}

func (g *game) Update() error {
	used := g.ui.Update()
	if g.ui.OverClickable() {
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	} else {
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
	if used {
		return nil // the UI used the input this frame
	}
	g.gold.SetText(strconv.Itoa(g.coin))
	return nil
}

func (g *game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x2b, 0x33, 0x2b, 0xff})
	g.ui.Draw(screen)
}

func (g *game) Layout(int, int) (int, int) {
	g.ui.SetScreenSize(screenW, screenH)
	return screenW, screenH
}

func main() {
	ebiten.SetWindowSize(screenW*2, screenH*2)
	ebiten.SetWindowTitle("apostle-ui demo")
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
