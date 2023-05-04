// Copyright 2021 Siôn le Roux.  All rights reserved.
// Use of this source code is subject to an MIT-style
// licence which can be found in the LICENSE file.

package main

import (
	"github.com/Frabjous-Studios/bankwave/internal"
	"github.com/Frabjous-Studios/bankwave/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"log"
	"os"
)

func main() {
	gameWidth, gameHeight := 640, 480
	err := os.Setenv("EBITENGINE_GRAPHICS_LIBRARY", "opengl")
	if err != nil {
		debug.Println("could not set gfx lib")
	}

	ebiten.SetWindowSize(gameWidth, gameHeight)
	ebiten.SetWindowTitle("BankWave: Neon Networth")

	game := &internal.Game{
		Width:  gameWidth,
		Height: gameHeight,
		ACtx:   audio.NewContext(internal.SampleRate),
	}

	game.CurrScene = internal.NewLogoScene(game)
	if err := ebiten.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
