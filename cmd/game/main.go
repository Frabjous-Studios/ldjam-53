// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/Frabjous-Studios/ebitengine-game-template/internal"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"log"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
)

func main() {
	var err error
	err = os.Setenv("EBITENGINE_GRAPHICS_LIBRARY", "opengl")
	if err != nil {
		log.Println("unable to set graphics lib to opengl:", err)
	}
	gameWidth, gameHeight := 640, 480

	ebiten.SetWindowSize(gameWidth, gameHeight)
	ebiten.SetWindowTitle("ebitengine-game-template")

	game := &internal.Game{
		Width:  gameWidth,
		Height: gameHeight,
		ACtx:   audio.NewContext(internal.SampleRate),
	}
	game.CurrScene, err = internal.NewMainMenuScene(game)
	if err != nil {
		log.Fatal(err)
	}
	game.CurrScene = internal.NewMainScene(game) // TODO: remove

	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
