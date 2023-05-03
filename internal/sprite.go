package internal

import (
	"github.com/Frabjous-Studios/bankwave/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"image"
	"math"
)

type Sprite interface {
	DrawTo(screen *ebiten.Image)
	Bounds() image.Rectangle
	Pos() image.Point
	SetPos(image.Point)
	MoveX(float64)
	MoveY(float64)
}

type BaseSprite struct {
	X, Y   int
	fX, fY float64
	Img    *ebiten.Image
}

func (s *BaseSprite) DrawTo(screen *ebiten.Image) {
	if s.Img == nil {
		debug.Println("image for sprite was nil at point:", s.X, s.Y)
		return
	}
	opt := &ebiten.DrawImageOptions{}
	opt.GeoM.Translate(float64(s.X), float64(s.Y))
	opt.GeoM.Scale(ScaleFactor, ScaleFactor)
	screen.DrawImage(s.Img, opt)
}

func (s *BaseSprite) Bounds() image.Rectangle {
	return s.Img.Bounds().Add(image.Pt(s.X, s.Y))
}

func (s *BaseSprite) Pos() image.Point {
	return image.Pt(s.X, s.Y)
}

func (s *BaseSprite) SetPos(pt image.Point) {
	s.X = pt.X
	s.Y = pt.Y
}

func (s *BaseSprite) ClampToRect(r image.Rectangle) {
	s.X = clamp(s.X, r.Min.X, r.Max.X-s.Bounds().Dx())
	s.Y = clamp(s.Y, r.Min.Y, r.Max.Y-s.Bounds().Dy())
}

func (s *BaseSprite) MoveX(amt float64) {
	s.fX += amt
	px, fx := math.Modf(s.fX)
	s.fX = fx
	s.X = s.X + int(math.Round(px))
}

func (s *BaseSprite) MoveY(amt float64) {
	s.fY += amt
	px, fx := math.Modf(s.fY)
	s.fY = fx
	s.Y = s.Y + int(px)
}
