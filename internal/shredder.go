package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kalexmills/asebiten"
	"image"
)

type ShredderMode uint8

const (
	ModeShred ShredderMode = iota
	ModeScan
	ModeDisabled
)

type Shredder struct {
	*BaseSprite

	anim *asebiten.Animation

	Operational bool
	Mode        ShredderMode
}

func NewShredder() *Shredder {
	result := &Shredder{
		anim:       Resources.GetAnim("shredder"),
		Mode:       ModeDisabled,
		BaseSprite: &BaseSprite{X: 112, Y: 171},
	}
	result.anim.Pause()
	return result
}

func (s *Shredder) Bounds() image.Rectangle {
	return s.anim.Bounds().Add(image.Pt(s.X, s.Y))
}

func (s *Shredder) DrawTo(screen *ebiten.Image) {
	s.anim.SetFrame(int(s.Mode))
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(s.X), float64(s.Y))
	opts.GeoM.Scale(ScaleFactor, ScaleFactor)

	s.anim.DrawTo(screen, opts)
}

func (s *Shredder) enable() {
	s.Operational = true
	s.Mode = ModeScan
}

func (s *Shredder) toggle() {
	if !s.Operational {
		return
	}
	switch s.Mode {
	case ModeScan:
		s.Mode = ModeShred
	case ModeShred:
		s.Mode = ModeScan
	}
}
