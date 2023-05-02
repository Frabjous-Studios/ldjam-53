package internal

import (
	"errors"
	"github.com/Frabjous-Studios/bankwave/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/solarlune/resound"
	"sync"
	"time"
)

// TPS is the number of ticks per second, read once when the game starts.
var TPS float64
var TPSOnce sync.Once

// Game represents the main game state
type Game struct {
	Width     int
	Height    int
	CurrScene Scene

	ACtx *audio.Context

	playingFilename string
	playingVolume   *resound.Volume
	playingPlayer   *resound.DSPPlayer
	incomingPlayer  *resound.DSPPlayer
	incomingVolume  *resound.Volume

	fadeStart time.Time
}

type Scene interface {
	Update() error
	Draw(*ebiten.Image)
}

// PlayMusic fades out the last track that was playing and fades in a new track
func (g *Game) PlayMusic(file string) {
	if g.playingFilename == file {
		return
	}
	g.playingFilename = file
	g.fadeStart = time.Now()
	loop := Resources.GetMusic(g.ACtx, file)
	if loop == nil {
		debug.Printf("no file found with name: %s", file)
		return
	}
	ch := resound.NewDSPChannel()
	g.incomingVolume = resound.NewVolume(nil).SetStrength(0.0)
	ch.Add("volume", g.incomingVolume)
	g.incomingPlayer = ch.CreatePlayer(loop)
	g.incomingPlayer.Rewind()
	g.incomingPlayer.Play()
}

// Layout is hardcoded for now, may be made dynamic in future
func (g *Game) Layout(outsideWidth int, outsideHeight int) (screenWidth int, screenHeight int) {
	return g.Width, g.Height
}

const crossFadeTime = 5 * time.Second

const maxVolume = 0.75

// Update calculates game logic
func (g *Game) Update() error {
	TPSOnce.Do(func() {
		ebiten.SetTPS(240)
		TPS = float64(ebiten.TPS())
	})

	// Pressing Q any time quits immediately
	if ebiten.IsKeyPressed(ebiten.KeyQ) {
		return errors.New("game quit by player")
	}

	// Pressing F toggles full-screen
	if inpututil.IsKeyJustPressed(ebiten.KeyF) {
		if ebiten.IsFullscreen() {
			ebiten.SetFullscreen(false)
		} else {
			ebiten.SetFullscreen(true)
		}
	}

	if g.incomingPlayer != nil && g.playingPlayer != nil {
		dt := float64(time.Now().Sub(g.fadeStart)) / float64(crossFadeTime)
		if dt >= 1.0 {
			g.incomingVolume.SetStrength(maxVolume)
			g.playingVolume.SetStrength(0.0)
			g.playingPlayer.Pause()
			g.playingPlayer.Rewind()
			g.playingVolume = g.incomingVolume
			g.playingPlayer = g.incomingPlayer
			g.incomingVolume = nil
			g.incomingPlayer = nil
		} else {
			g.incomingVolume.SetStrength(dt * maxVolume)
			g.playingVolume.SetStrength((1.0 - dt) * maxVolume)
		}
	} else if g.incomingPlayer != nil {
		debug.Println("starting new song!")
		g.incomingVolume.SetStrength(maxVolume)
		g.playingPlayer = g.incomingPlayer
		g.playingVolume = g.incomingVolume
		g.incomingVolume = nil
		g.incomingPlayer = nil
	}

	return g.CurrScene.Update()

}

// Draw draws the game screen by one frame
func (g *Game) Draw(screen *ebiten.Image) {
	g.CurrScene.Draw(screen)
}

// ChangeScene sets the current scene to the provided Scene.
func (g *Game) ChangeScene(s Scene) {
	g.CurrScene = s
}
