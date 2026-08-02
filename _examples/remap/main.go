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
	ActionUnknown input.Action = iota
	ActionPing
	ActionRemap
)

func main() {
	ebiten.SetWindowSize(640, 480)

	if err := ebiten.RunGame(newExampleGame()); err != nil {
		log.Fatal(err)
	}
}

type exampleGame struct {
	started bool

	k        input.Key
	prevK    input.Key
	scanning bool

	keyScanner *input.KeyScanner

	inputHandler *input.Handler
	inputSystem  input.System
}

func newExampleGame() *exampleGame {
	g := &exampleGame{}

	g.inputSystem.Init(input.SystemConfig{
		DevicesEnabled: input.AnyDevice,
	})

	return g
}

func (g *exampleGame) Layout(_, _ int) (int, int) {
	return 640, 480
}

func (g *exampleGame) Draw(screen *ebiten.Image) {
	if g.scanning {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("keybind: %s\n<scanning the new keybing>", g.k))
	} else {
		ebitenutil.DebugPrint(screen, fmt.Sprintf("keybind: %s\npress ctrl+enter to remap", g.k))
	}
}

func (g *exampleGame) Update() error {
	g.inputSystem.Update()

	if !g.started {
		g.Init()
		g.started = true
	}

	if !g.scanning && g.inputHandler.ActionIsJustPressed(ActionPing) {
		fmt.Printf("ping! (activated with %s keybind)\n", g.k)
	}

	g.handleRemap()

	return nil
}

func (g *exampleGame) handleRemap() {
	if !g.scanning {
		if g.inputHandler.ActionIsJustPressed(ActionRemap) {
			g.scanning = true
		}
		return
	}

	k, status := g.keyScanner.Scan()
	if k == input.KeyWithModifier(input.KeyEnter, input.ModControl) || k == input.KeyControl || k == input.KeyEnter {
		// reject the keys used to start key scanning for the purposes of the example
	} else if status == input.KeyScanCompleted {
		// Check for the new key to be available.
		// Resolve the conflicts here.
		g.scanning = false
		g.k = k
		g.inputHandler.Remap(g.makeKeymap())
	}
}

func (g *exampleGame) makeKeymap() input.Keymap {
	return input.Keymap{
		ActionPing:  {g.k},
		ActionRemap: {input.KeyWithModifier(input.KeyEnter, input.ModControl)},
	}
}

func (g *exampleGame) Init() {
	g.k = input.KeyQ
	g.inputHandler = g.inputSystem.NewHandler(0, g.makeKeymap())
	g.keyScanner = input.NewKeyScanner(g.inputHandler)
}
