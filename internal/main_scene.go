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

	Customers []string // Customers is a list of Yarnspinner nodes happening on the current day.

	Sprites []Sprite

	State  *GameState
	Runner *DialogueRunner

	till    *Till
	counter *BaseSprite

	holding     Sprite
	clickStart  image.Point
	clickOffset image.Point

	lines []*DialogueLine
	opts  []*DialogueLine
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
		Customers: Days[0],
		State: &GameState{
			CurrentNode: "Start",
			Vars:        make(yarn.MapVariableStorage),
		},
		till: NewTill(),
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
			if cPos.In(m.till.Bounds()) { // if over Till; drop on Till
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
				m.till.Remove(m.holding) // remove it from the Till (maybe)
			}
		}
	}

	if m.holding != nil {
		mPos := cursorPos()
		m.holding.SetPos(mPos.Add(m.clickOffset))
	}

	// update dialogue
	m.updateDialogue()
	return nil
}

func (m *MainScene) updateDialogue() {
	lines, err := m.Runner.Lines()
	if err != nil {
		debug.Println("error getting lines: ", err)
	}
	opts, err := m.Runner.Options()
	if err != nil {
		debug.Println("error getting options:", err)
		return
	}

	m.lines = renderAttributedStr(lines)
	m.opts = m.renderOpts(opts)

	// TODO: have the bubbles push themselves upwards until they're off the screen and can be safely culled.
}

func (m *MainScene) renderOpts(astrs []*yarn.AttributedString) []*DialogueLine {
	lines := renderAttributedStr(astrs)
	result := make([]*DialogueLine, len(lines))
	for i, line := range lines {
		result[i] = line
		fmt.Println(line.Line)
		if i < len(m.opts) {
			result[i].Highlighted = m.opts[i].Highlighted // transfer highlighted and clicked info from last frame
		}
	}
	return result
}

func renderAttributedStr(strs []*yarn.AttributedString) []*DialogueLine {
	var result []*DialogueLine
	for _, str := range strs {
		result = append(result, &DialogueLine{
			Line:       str.String(),
			IsCustomer: true,
			// TODO: set BaseSprite
		})
	}
	return result
}

func (m *MainScene) Draw(screen *ebiten.Image) {
	// draw Till
	m.till.DrawTo(screen)

	// draw counter
	m.counter.DrawTo(screen)

	// draw all the sprites in their draw order.
	for _, sprite := range m.Sprites {
		sprite.DrawTo(screen)
	}
}

type DialogueLine struct {
	*BaseSprite // 9-patch

	Line        string
	IsCustomer  bool
	Highlighted bool
	Clickbox    image.Rectangle
}

func (m *MainScene) startRunner() {
	if err := m.Runner.Start(m.State); err != nil {
		debug.Printf("error starting runner: %v", err)
		return
	}
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

type Money struct {
	*BaseSprite
	Value  int // Value is in cents.
	IsCoin bool
}

// clampToCounter clamps the provided point to the counter range (hardcoded)
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

func rect(x, y, w, h int) image.Rectangle {
	return image.Rect(x, y, x+w, y+h)
}
