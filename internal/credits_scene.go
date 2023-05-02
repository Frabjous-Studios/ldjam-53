package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type CreditsScene struct {
	Game *Game
	bg   *ebiten.Image
}

func NewCreditsScene(game *Game) (*CreditsScene, error) {
	game.PlayMusic("No_Surprises_Parody.ogg")
	return &CreditsScene{
		Game: game,
		bg:   Resources.GetImage("credits.png"),
	}, nil
}

func (s *CreditsScene) Update() error {
	keys := inpututil.AppendPressedKeys(nil)

	if contains(keys, ebiten.KeyEscape) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		s.Game.ChangeScene(NewMainMenuScene(s.Game))
	}
	return nil
}

func (s *CreditsScene) Draw(screen *ebiten.Image) {
	if s.bg != nil {
		screen.DrawImage(s.bg, nil)
	}
}
