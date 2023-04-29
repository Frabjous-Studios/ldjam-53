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

func (s *BaseSprite) ClampToRect(r image.Rectangle) {
	s.X = clamp(s.X, r.Min.X, r.Max.X-s.Bounds().Dx())
	s.Y = clamp(s.Y, r.Min.Y, r.Max.Y-s.Bounds().Dy())
}

type MainScene struct {
	Game *Game

	Sprites []Sprite

	State  *GameState
	Runner *DialogueRunner

	till    *till
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
			newCoin(5, 25, 20),
			newCoin(25, 45, 20),
		},
		State: &GameState{
			CurrentNode: "Start",
			Vars:        make(yarn.MapVariableStorage),
		},
		till: &till{
			BaseSprite: &BaseSprite{
				X: 0, Y: 172,
				Img: Resources.images["till"],
			},
			DropTargets: [2][5]image.Rectangle{
				CoinTargets: { //
					rect(4, 50, 20, 15),
					rect(25, 50, 20, 15),
					rect(46, 50, 20, 15),
					rect(67, 50, 20, 15),
					rect(88, 50, 20, 15),
				},
				BillTargets: { // bill targets
					rect(4, 3, 20, 45),
					rect(25, 3, 20, 45),
					rect(46, 3, 20, 45),
					rect(67, 3, 20, 45),
					rect(88, 3, 20, 45),
				},
			},
		},
		counter: &BaseSprite{
			X: 112, Y: 152,
			Img: Resources.images["counter"],
		},
	}
}

func (m *MainScene) Update() error {
	if !m.Runner.running {
		go m.startRunner()
		m.Runner.running = true
	}

	cPos := cursorPos()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if m.holding != nil {
			fmt.Println("tillBounds", m.till.Bounds(), "cPos", cPos)
			if cPos.In(m.till.Bounds()) { // if over till; drop on till
				fmt.Println("inside!")
				if !m.till.Drop(m.holding) {
					m.counterDrop()
				} else {
					m.holding = nil
				}
			} else {
				m.counterDrop()
			}
		} else { // pick up the thing under the cursor
			m.holding = m.spriteUnderCursor()
			if m.holding != nil {
				m.clickStart = cPos
				m.clickOffset = m.holding.Pos().Sub(m.clickStart)
				m.till.Remove(m.holding) // remove it from the till (maybe)
			}
		}
	}

	if m.holding != nil {
		mPos := cursorPos()
		m.holding.SetPos(mPos.Add(m.clickOffset))
	}

	return nil
}

func (m *MainScene) pickUp() {
	m.holding = m.spriteUnderCursor()
	if m.holding != nil {
		m.clickStart = cursorPos()
		m.clickOffset = m.holding.Pos().Sub(m.clickStart)
	}
}

func (m *MainScene) counterDrop() {
	m.holding.SetPos(clampToCounter(m.holding.Pos()))
	m.holding = nil
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
	Value  int // Value is in cents.
	IsCoin bool
}

// newBill creates in local coordinates on the counter.
func newBill(denom int, x, y int) Sprite {
	x = clamp(x+112, 112, 320-43)
	y = clamp(y+152, 152, 240-43)

	img := Resources.images[fmt.Sprintf("bill_%d", denom)]
	return &Money{
		Value:  denom * 100,
		IsCoin: false,
		BaseSprite: &BaseSprite{
			X:   x,
			Y:   y,
			Img: img,
		},
	}
}

// newCoin creates in local coordinates on the counter.
func newCoin(denom int, x, y int) Sprite {
	x = clamp(x+112, 112, 320-15)
	y = clamp(y+152, 152, 240-15)
	img := Resources.images[fmt.Sprintf("coin_%d", denom)]
	return &Money{
		Value:  denom,
		IsCoin: true,
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

const CoinTargets = 0
const BillTargets = 1

type till struct {
	*BaseSprite

	DropTargets [2][5]image.Rectangle
	BillSlots   [5][]*Money
	CoinSlots   [5][]*Money
}

// Drop drops the provided sprite on the till; landing it in the location needed.
func (t *till) Drop(s Sprite) bool {
	m, ok := s.(*Money)
	if !ok {
		fmt.Println("not money!")
		return false
	}
	var targets [5]image.Rectangle
	if m.IsCoin {
		targets = t.DropTargets[CoinTargets]
	} else {
		targets = t.DropTargets[BillTargets]
	}
	fmt.Println("targets", targets)
	// find drop target with max area intersection
	bestIdx, maxA := -1, 0
	for idx, rect := range targets {
		sz := m.Bounds().Intersect(rect.Add(t.Pos())).Size()
		fmt.Println("rect:", rect, "sz:", sz, "pos:", t.Pos())
		a := sz.X * sz.Y
		if a > 0 && a > maxA {
			bestIdx = idx
			maxA = a
		}
	}
	if bestIdx == -1 {
		return false
	}
	r := targets[bestIdx].Add(t.Pos())
	m.ClampToRect(r)
	if m.IsCoin {
		t.CoinSlots[bestIdx] = append(t.CoinSlots[bestIdx], m)
	} else {
		t.BillSlots[bestIdx] = append(t.BillSlots[bestIdx], m)
	}
	return true
}

// Remove removes the provided money from the till; checking the top of each stack of bills and coins for it.
func (t *till) Remove(s Sprite) {
	m, ok := s.(*Money)
	if !ok {
		return
	}
	for i := 0; i < 5; i++ {
		if len(t.BillSlots[i]) > 0 && t.BillSlots[i][len(t.BillSlots[i])-1] == m {
			fmt.Println("removed", m.Value, "from", i)
			t.BillSlots[i] = t.BillSlots[i][:len(t.BillSlots[i])-1]
			return
		}
		if len(t.CoinSlots[i]) > 0 && t.CoinSlots[i][len(t.CoinSlots[i])-1] == m {
			t.CoinSlots[i] = t.CoinSlots[i][:len(t.CoinSlots[i])-1]
			return
		}
	}
}

func rect(x, y, w, h int) image.Rectangle {
	return image.Rect(x, y, x+w, y+h)
}
