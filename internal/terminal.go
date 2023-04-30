package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type Terminal struct {
	*BaseSprite
	till *Till

	operational bool

	bg *ebiten.Image
}

func NewTerminal(till *Till) *Terminal {
	result := &Terminal{
		till: till,
		BaseSprite: &BaseSprite{
			X: 0,
			Y: 72,
		},
		operational: true, // TODO: turn this OFF for the first day!
	}
	result.bg = Resources.images["terminal"]
	result.Img = ebiten.NewImage(result.bg.Bounds().Dx(), result.bg.Bounds().Dy())

	return result
}

func (t *Terminal) DrawTo(screen *ebiten.Image) {
	t.Img.Clear()
	// draw background
	opts := &ebiten.DrawImageOptions{}

	t.Img.DrawImage(t.bg, opts)

}

func (t *Terminal) Update() {

}
