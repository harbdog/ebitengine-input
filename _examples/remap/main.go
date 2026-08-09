//go:build example

package main

import (
	"fmt"
	"log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	input "github.com/quasilyte/ebitengine-input"
)

const (
	width  int = 640
	height int = 480
)

const (
	ActionUnknown input.Action = iota
	ActionPing
	ActionMove
	ActionRemapKey
	ActionRemapAxes
)

func main() {
	ebiten.SetWindowSize(width, height)

	if err := ebiten.RunGame(newExampleGame()); err != nil {
		log.Fatal(err)
	}
}

type exampleGame struct {
	started bool

	k            input.Key
	axes         input.Key
	scanningKey  bool
	scanningAxes bool

	keyScanner *input.KeyScanner

	pos       input.Vec
	character string

	inputHandler *input.Handler
	inputSystem  input.System
}

func newExampleGame() *exampleGame {
	g := &exampleGame{
		pos:       input.Vec{X: 200, Y: 200},
		character: "@",
	}

	g.inputSystem.Init(input.SystemConfig{
		DevicesEnabled: input.AnyDevice,
	})

	return g
}

func (g *exampleGame) Layout(_, _ int) (int, int) {
	return width, height
}

func screenClamp(x, y float64) (float64, float64) {
	if x < 0 {
		x = 0
	}
	if x >= float64(width) {
		x = float64(width) - 1
	}
	if y < 0 {
		y = 0
	}
	if y >= float64(height) {
		y = float64(height) - 1
	}
	return x, y
}

func (g *exampleGame) Draw(screen *ebiten.Image) {
	if g.scanningKey {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("keybind: %s\naxes: %s\n<scanning the new keybind>", g.k, g.axes))
	} else if g.scanningAxes {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("keybind: %s\naxes: %s\n<scanning the new axes>", g.k, g.axes))
	} else {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("keybind: %s\naxes: %s\npress ctrl+enter to remap keybind\nor shift+enter to remap axes", g.k, g.axes))
	}
	ebitenutil.DebugPrintAt(screen, g.character, int(g.pos.X), int(g.pos.Y))
}

func (g *exampleGame) Update() error {
	g.inputSystem.Update()

	if !g.started {
		g.Init()
		g.started = true
	}

	if !g.scanningKey && !g.scanningAxes {
		// react to ping action by "highlighting" the "character"
		if g.inputHandler.ActionIsJustPressed(ActionPing) {
			fmt.Printf("ping! (activated with %s keybind)\n", g.k)
		}
		if g.inputHandler.ActionIsPressed(ActionPing) {
			g.character = "***\n*@*\n***"
		} else {
			g.character = "@"
		}

		// react to move action by moving the "character"
		if info, ok := g.inputHandler.PressedActionInfo(ActionMove); ok {
			if info.IsMouseWheelEvent() {
				// mouse wheel moves in reverse direction of draw position
				g.pos.Y -= info.Pos.Y
			} else if info.IsMouseMotionEvent() {
				// mouse move position (info.Pos) is absolute, the delta position (info.DeltaPos)
				// can be useful, typically when using ebiten.CursorModeCaptured
				g.pos.X += info.DeltaPos.X
				g.pos.Y += info.DeltaPos.Y
			} else {
				g.pos.X += info.Pos.X
				g.pos.Y += info.Pos.Y
			}
		}

		// clamp position to screen window size
		g.pos.X, g.pos.Y = screenClamp(g.pos.X, g.pos.Y)
	}

	// keep scanning of keys separate from axes to ensure events are isolated to just what is needed
	if !g.scanningAxes {
		g.handleRemapKey()
	}
	if !g.scanningKey {
		g.handleRemapAxes()
	}

	return nil
}

func (g *exampleGame) handleRemapKey() {
	if !g.scanningKey {
		if g.inputHandler.ActionIsJustPressed(ActionRemapKey) {
			g.scanningKey = true
		}
		return
	}

	k, status := g.keyScanner.Scan()
	if k == input.KeyWithModifier(input.KeyEnter, input.ModControl) ||
		k == input.KeyWithModifier(input.KeyEnter, input.ModShift) ||
		k == input.KeyControl || k == input.KeyShift || k == input.KeyEnter {
		// reject the keys used to start key scanning for the purposes of the example
	} else if status == input.KeyScanCompleted {
		// Check for the new key to be available.
		// Resolve the conflicts here.
		g.scanningKey = false
		g.k = k
		g.inputHandler.Remap(g.makeKeymap())
	}
}

func (g *exampleGame) handleRemapAxes() {
	if !g.scanningAxes {
		if g.inputHandler.ActionIsJustPressed(ActionRemapAxes) {
			g.scanningAxes = true
		}
		return
	}

	axes, status := g.keyScanner.ScanAxes()
	if status == input.KeyScanCompleted {
		// Check for the new key to be available.
		// Resolve the conflicts here.
		g.scanningAxes = false
		g.axes = axes
		g.inputHandler.Remap(g.makeKeymap())
	}
}

func (g *exampleGame) makeKeymap() input.Keymap {
	return input.Keymap{
		ActionPing:      {g.k},
		ActionMove:      {g.axes},
		ActionRemapKey:  {input.KeyWithModifier(input.KeyEnter, input.ModControl)},
		ActionRemapAxes: {input.KeyWithModifier(input.KeyEnter, input.ModShift)},
	}
}

func (g *exampleGame) Init() {
	g.k = input.KeyQ
	g.axes = input.KeyMouseMotion
	g.inputHandler = g.inputSystem.NewHandler(0, g.makeKeymap())
	g.keyScanner = input.NewKeyScanner(g.inputHandler)
}
