package internal

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kalexmills/asebiten"
)

type AlarmMode uint8

const (
	AlarmModeUnpressed = iota
	AlarmModeRight
	AlarmModeLeft
)

var AlarmButtonLeft = rect(275, 171, 14, 8)
var AlarmButtonRight = rect(291, 171, 14, 8)

type AlarmButtons struct {
	*BaseSprite
	anim     *asebiten.Animation
	Contents []Sprite
	Mode     AlarmMode
}

func NewAlarmButtons() *AlarmButtons {
	result := &AlarmButtons{
		anim:       Resources.GetAnim("alarm_buttons"),
		BaseSprite: &BaseSprite{X: 236, Y: 170},
		Mode:       AlarmModeUnpressed,
	}
	return result
}

func (s *AlarmButtons) DrawTo(screen *ebiten.Image) {
	s.anim.SetFrame(int(s.Mode))
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(s.X), float64(s.Y))
	opts.GeoM.Scale(ScaleFactor, ScaleFactor)

	s.anim.DrawTo(screen, opts)
}
