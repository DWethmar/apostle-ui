// Demo shows the ui package: a HUD panel, a menu, alerts and a yes/no modal.
//
//	go run ./examples/demo
package main

import (
	"fmt"
	"image/color"
	"log"

	"github.com/hajimehoshi/ebiten/v2"

	ui "github.com/dwethmar/apostle-ui"
)

const screenW, screenH = 480, 270

type game struct {
	ui   *ui.UI
	gold *ui.Text
	coin int
}

func newGame() *game {
	g := &game{ui: ui.New()}
	u := g.ui

	// A HUD panel whose text changes every frame.
	g.gold = ui.NewText("")
	u.Place(ui.NewPanel(
		g.gold,
		ui.NewText("Day 1").Color(color.RGBA{0x9a, 0xc8, 0xff, 0xff}),
	), ui.TopLeft, 8, 8)

	// A menu of buttons.
	u.Place(ui.NewPanel(
		ui.NewText("Menu").Align(ui.Center),
		ui.NewButton("Show message", func() {
			u.Alert("A note", "Someone scratched words into the wall: "+
				"\"The eclipse comes for those who seek it.\" "+
				"The letters are old, but the blood is not.")
		}),
		ui.NewButton("Leave the cave", g.askToLeave),
		ui.NewButton("Find gold", func() { g.coin += 10 }),
	).Width(140), ui.TopRight, 8, 8)

	u.Place(ui.NewText("Esc closes a modal, Enter presses its first button."), ui.Bottom, 0, 8)
	return g
}

func (g *game) askToLeave() {
	g.ui.Open(ui.NewModal("Leave the cave?",
		ui.NewText("It is dark outside and you hear wolves. You can come back later."),
	).
		AddButton("Leave", func() { g.ui.Alert("Outside", "The night is cold.") }).
		AddButton("Stay", nil))
}

func (g *game) Update() error {
	if g.ui.Update() {
		return nil // the UI used the input this frame
	}
	g.gold.SetText(fmt.Sprintf("Gold: %d", g.coin))
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
