package internal

import (
	"fmt"
	"github.com/DrJosh9000/yarn"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/math/fixed"
	"image"
	"log"
	"strings"
	"sync"
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

	Day      Day // Customers is a list of Yarnspinner nodes happening on the current day
	Customer Sprite

	Sprites []Sprite

	State  *GameState
	Runner *DialogueRunner

	till    *Till
	counter *BaseSprite

	bubbles *Bubbles
	options []*Line

	holding     Sprite
	clickStart  image.Point
	clickOffset image.Point

	vars yarn.MapVariableStorage
	mut  sync.Mutex

	speaking  *sync.Cond
	selectOpt *sync.Cond
	selection int
}

func NewMainScene(g *Game) *MainScene {
	runner, err := NewDialogueRunner()
	if err != nil {
		log.Fatal(err)
	}
	result := &MainScene{
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
		Day: Days[0],
		State: &GameState{
			CurrentNode: "Start",
			Vars:        make(yarn.MapVariableStorage),
		},

		till: NewTill(),
		counter: &BaseSprite{
			X: 112, Y: 152,
			Img: Resources.images["counter"],
		},
		vars: make(yarn.MapVariableStorage),
	}
	result.bubbles = NewBubbles(result)
	result.speaking = sync.NewCond(&result.mut)
	result.selectOpt = sync.NewCond(&result.mut)
	return result
}

func (m *MainScene) Update() error {
	if !m.Runner.running {
		go m.startRunner()
		m.Runner.running = true
	}
	m.bubbles.Update()

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
			} else {
				// check for dialogue option
				for idx, opt := range m.options {
					fmt.Println("checking!")
					if cPos.Mul(ScaleFactor).In(opt.Rect) {
						fmt.Println("selected", idx)
						m.selection = idx
						m.selectOpt.Broadcast()
						m.options = nil
						break
					}
				}
			}
		}
	}
	for _, opt := range m.options {
		if opt == nil {
			continue
		}
		if cPos.Mul(ScaleFactor).In(opt.Rect) {
			opt.highlighted = true
		} else {
			opt.highlighted = false
		}
	}

	if m.holding != nil {
		mPos := cursorPos()
		m.holding.SetPos(mPos.Add(m.clickOffset))
	}

	return nil
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

	// draw dialogue bubbles.
	m.bubbles.DrawTo(screen)

	// draw dialogue options
	m.drawOptions(screen)
}

// 290, 151, 140, 40

var OptionsBounds = rect(300, 240, 280, 80)

func (m *MainScene) drawOptions(screen *ebiten.Image) {
	m.bubbles.txt.SetTarget(screen)
	feed := m.bubbles.txt.NewFeed(fixed.P(OptionsBounds.Min.X, OptionsBounds.Min.Y))
	for _, opt := range m.options {
		if opt.highlighted {
			m.bubbles.txt.SetColor(fontColorHighlight)
		} else {
			m.bubbles.txt.SetColor(fontColor)
		}
		opt.Rect = m.bubbles.print(feed, opt, OptionsBounds)
		feed.LineBreak()
	}
}

func (m *MainScene) startRunner() {
	if err := m.Runner.Start(m.vars, m); err != nil {
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

func (m *MainScene) NodeStart(name string) error {
	fmt.Println("start node", name)
	return nil
}

func (m *MainScene) PrepareForLines(lineIDs []string) error {
	return nil
}

func (m *MainScene) Line(line yarn.Line) error {
	m.mut.Lock()
	defer m.mut.Unlock()
	m.bubbles.SetLine(m.Runner.Render(line))

	for !m.bubbles.IsDone() && !m.Runner.IsLastLine(line) {
		fmt.Println("locked!")
		m.speaking.Wait()
		fmt.Println("unlocked; checking again")
	}
	return nil
}

func (m *MainScene) Options(options []yarn.Option) (int, error) {
	m.mut.Lock()
	defer m.mut.Unlock()
	m.options = make([]*Line, 0, len(options))
	for _, opt := range options {
		m.options = append(m.options, NewLine(m.Runner.Render(opt.Line)))
	}
	m.selectOpt.Wait() // wait for the player to make a selection
	return m.selection, nil
}

func (m *MainScene) Command(command string) error {
	fmt.Println("run command:", command)
	command = strings.TrimSpace(command)
	tokens := strings.Split(command, " ")
	if len(tokens) == 0 {
		return fmt.Errorf("bad command: %s", command)
	}
	switch tokens[0] {
	default:
		return fmt.Errorf("unknown command %s", tokens[0])
	}
}

func (m *MainScene) NodeComplete(nodeName string) error {
	fmt.Println("node done", nodeName)
	return nil
}

func (m *MainScene) DialogueComplete() error {
	m.bubbles.SetLine("")
	fmt.Println("dialogue complete")
	return nil
}

type Portrait struct {
	*BaseSprite
}

func newPortrait(body, head string) Sprite {
	b, h := Resources.bodies[body], Resources.heads[head]
	img := ebiten.NewImage(100, 100)
	img.DrawImage(b, nil)
	img.DrawImage(h, nil)
	return &Portrait{
		BaseSprite: &BaseSprite{
			Img: img,
			X:   170,
			Y:   52,
		},
	}
}

func newRandPortrait() Sprite {
	return newPortrait(randMapKey(Resources.bodies), randMapKey(Resources.heads))
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
