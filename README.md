# apostle-ui

A small UI package for [Ebitengine](https://ebitengine.org/) games, made for
[Apostle](https://github.com/DWethmar/apostle). It does two things:

- **Panels** with text and buttons, placed on the screen (HUDs, menus).
- **Modals**: dialogs in the middle of the screen that take all input until
  they close.

![A modal in the demo](docs/demo.png)

```sh
go get github.com/dwethmar/apostle-ui
go run ./examples/demo   # try it
```

## Usage

Build elements once. Keep a pointer to anything you want to change later.

```go
import ui "github.com/dwethmar/apostle-ui"

u := ui.New()

// A HUD panel in the top-left corner, 8px from the edges.
gold := ui.NewText("Gold: 0")
u.Place(ui.NewPanel(gold), ui.TopLeft, 8, 8)

// A message with an OK button.
u.Alert("Welcome", "You wake up in a dark cave.")

// A question.
u.Open(ui.NewModal("Leave the cave?", ui.NewText("It is dark outside.")).
	AddButton("Leave", leave). // runs leave, then closes
	AddButton("Stay", nil))    // just closes
```

Hook it into your game:

```go
func (g *Game) Update() error {
	if g.ui.Update() {
		return nil // the UI used this frame's input
	}
	g.gold.SetText(fmt.Sprintf("Gold: %d", g.coins))
	// ... the rest of the game
}

func (g *Game) Draw(screen *ebiten.Image) {
	// ... draw the world
	g.ui.Draw(screen) // last, so the UI is on top
}

func (g *Game) Layout(int, int) (int, int) {
	g.ui.SetScreenSize(screenW, screenH)
	return screenW, screenH
}
```

`Update` returns true while a modal is open, or when a click landed on a panel,
so the game knows to ignore it.

## Elements

| Element | Made with | Notes |
|---|---|---|
| Text | `NewText(s)` | Wraps to the width it is given. `SetText`, `Color`, `Align`. |
| Button | `NewButton(label, fn)` | Fires on press + release over it. `SetLabel`, `SetDisabled`. |
| Panel | `NewPanel(children...)` | Background and border; children top to bottom. `Width`, `Padding`, `Gap`, `Align`. |
| Column / Row | `Column(...)`, `Row(...)` | Layout without a box. `Gap`, `Width`, `Align`, `Justify`. |
| Modal | `NewModal(title, children...)` | Open with `u.Open`. `AddButton`, `OnClose`, `Dismissible`, `Width`, `Close`. |

Every element has `SetVisible`, `Visible` and `Bounds`. Hidden elements take no
space.

Anchors for `Place`: `TopLeft`, `Top`, `TopRight`, `Left`, `Middle`, `Right`,
`BottomLeft`, `Bottom`, `BottomRight`.

Keys: **Escape** closes the top modal, **Enter** presses its first button.

## Styling

Everything uses `u.Theme`: font, padding, gap, the widest a modal may grow, and
all colors. The default font is a 12px bitmap font. To use your own:

```go
u.Theme.Face = myFace // any text/v2 Face
u.Theme.Title = color.RGBA{0xff, 0x80, 0x80, 0xff}
```

## Testing

`HandleInput(ui.Input{...})` is `Update` with input you supply, so clicks and
key presses can be tested without a window — see [ui_test.go](ui_test.go).
