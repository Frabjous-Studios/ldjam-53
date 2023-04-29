package internal

import (
	"fmt"
	"github.com/DrJosh9000/yarn"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"image"
	"log"
)

// global ScaleFactor for pixel art.
const ScaleFactor = 2.0

type BaseScene struct {
}

type Sprite interface {
	DrawTo(screen *ebiten.Image)
	Bounds() image.Rectangle
	Pos() image.Point
	SetPos(image.Point)
}

type BaseSprite struct {
	X, Y int
	Img  *ebiten.Image
}

func (s *BaseSprite) DrawTo(screen *ebiten.Image) {
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

type MainScene struct {
	Game *Game

	Sprites []Sprite

	State  *GameState
	Runner *DialogueRunner

	till    *BaseSprite
	counter *BaseSprite

	holding     Sprite
	clickStart  image.Point
	clickOffset image.Point
}

func NewMainScene(g *Game) *MainScene {
	runner, err := NewDialogueRunner()
	if err != nil {
		log.Fatal(err)
	}
	return &MainScene{
		Game:   g,
		Runner: runner,
		Sprites: []Sprite{
			newBill(1, 5, 60),
			newBill(5, 60, 60),
			newBill(10, 45, 60),
			newCoin(1, 5, 20),
			newCoin(1, 25, 20),
			newCoin(1, 45, 20),
		},
		State: &GameState{
			CurrentNode: "Start",
			Vars:        make(yarn.MapVariableStorage),
		},
		till: &BaseSprite{
			X: 112, Y: 152,
			Img: Resources.images["counter"],
		},
		counter: &BaseSprite{
			X: 0, Y: 172,
			Img: Resources.images["till"],
		},
	}
}

func (m *MainScene) Update() error {
	if !m.Runner.running {
		go m.startRunner()
		m.Runner.running = true
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if m.holding != nil {
			m.holding.SetPos(clampToCounter(m.holding.Pos()))
			m.holding = nil
		} else {
			m.holding = m.spriteUnderCursor()
			if m.holding != nil {
				m.clickStart = cursorPos()
				m.clickOffset = m.holding.Pos().Sub(m.clickStart)
			}
		}
	}

	if m.holding != nil {
		mPos := cursorPos()
		m.holding.SetPos(mPos.Add(m.clickOffset))
	}

	return nil
}

func (m *MainScene) spriteUnderCursor() Sprite {
	for _, sprite := range m.Sprites {
		if cursorPos().In(sprite.Bounds()) {
			return sprite
		}
	}
	return nil
}

func (m *MainScene) Draw(screen *ebiten.Image) {
	// draw till
	m.till.DrawTo(screen)

	// draw counter
	m.counter.DrawTo(screen)

	for _, sprite := range m.Sprites {
		sprite.DrawTo(screen)
	}
}

func (m *MainScene) startRunner() {
	if err := m.Runner.Start(m.State); err != nil {
		debug.Printf("error starting runner: %v", err)
		return
	}
}

type Money struct {
	*BaseSprite
	Value int // Value is in cents.
}

// newBill creates in local coordinates on the counter.
func newBill(denom int, x, y int) *Money {
	x = clamp(x+112, 112, 320-43)
	y = clamp(y+152, 152, 240-43)

	img := Resources.images[fmt.Sprintf("bill_%d", denom)]
	return &Money{
		Value: denom * 100,
		BaseSprite: &BaseSprite{
			X:   x,
			Y:   y,
			Img: img,
		},
	}
}

// newCoin creates in local coordinates on the counter.
func newCoin(denom int, x, y int) *Money {
	x = clamp(x+112, 112, 320-15)
	y = clamp(y+152, 152, 240-15)
	img := Resources.images[fmt.Sprintf("coin_%d", denom)]
	return &Money{
		Value: denom,
		BaseSprite: &BaseSprite{
			X:   x,
			Y:   y,
			Img: img,
		},
	}
}

func clampToCounter(pt image.Point) image.Point {
	pt.X = clamp(pt.X, 112, 320-15)
	pt.Y = clamp(pt.Y, 152, 240-15)
	return pt
}

func clamp(x int, min, max int) int {
	if x < min {
		return min
	}
	if x > max {
		return max
	}
	return x
}

func cursorPos() image.Point {
	mx, my := ebiten.CursorPosition()
	return image.Pt(mx/2, my/2)
}
