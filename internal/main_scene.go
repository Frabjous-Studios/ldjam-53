package internal

import (
	"fmt"
	"github.com/DrJosh9000/yarn"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"golang.org/x/image/math/fixed"
	"image"
	"strconv"
	"strings"
	"sync"
	"time"
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

	debouceTime time.Time
}

func NewMainScene(g *Game) *MainScene {
	var err error
	result := &MainScene{
		Game:    g,
		Sprites: []Sprite{},
		Day:     Days[0],

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

	result.Runner, err = NewDialogueRunner(result.vars, result)
	if err != nil {
		panic(err)
	}
	return result
}

func (m *MainScene) Update() error {
	if err := m.updateInput(); err != nil {
		debug.Printf("error from updateInput: %v", err)
	}
	m.bubbles.Update()

	runnerP := m.Runner.Portrait()
	if runnerP == nil {
		m.Customer = nil
	} else {
		// TODO: animate the customer into position
		m.Customer = runnerP
		m.Customer.SetPos(image.Pt(170, 53))
	}

	cPos := cursorPos()
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

const debounceDuration = 300 * time.Millisecond

// updateInput is debounced.
func (m *MainScene) updateInput() error {
	if time.Now().Before(m.debouceTime) {
		return nil
	}

	if !m.Runner.running {
		go m.startRunner()
		m.Runner.running = true
	}

	keys = inpututil.AppendJustPressedKeys(keys)

	if len(keys) > 0 {
		m.AdvanceDialogue()
		m.debouceTime = time.Now().Add(debounceDuration)
	}

	cPos := cursorPos()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if m.holding != nil {
			if cPos.In(m.till.Bounds()) { // if over Till; drop on Till
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
				m.BringToFront(m.holding)
				m.clickStart = cPos
				m.clickOffset = m.holding.Pos().Sub(m.clickStart)
				m.till.Remove(m.holding) // remove it from the Till (maybe)
			} else {
				selected := false
				// check for dialogue option
				for idx, opt := range m.options {
					if cPos.Mul(ScaleFactor).In(opt.Rect) {
						m.selection = idx
						m.selectOpt.Broadcast()
						m.options = nil
						selected = true
						break
					}
				}
				if !selected { // advance the dialogue if nothing was selected.
					m.AdvanceDialogue()
				}
			}
		}
		m.debouceTime = time.Now().Add(debounceDuration)
	}
	return nil
}

func (m *MainScene) BringToFront(s Sprite) {
	for idx, o := range m.Sprites {
		if s == o {
			m.Sprites = append(m.Sprites[:idx], append(m.Sprites[idx+1:], s)...)
			return
		}
	}
}

func (m *MainScene) AdvanceDialogue() {
	m.bubbles.BeDone()
}

func (m *MainScene) Draw(screen *ebiten.Image) {
	m.drawBg(screen)
	m.till.DrawTo(screen)

	if m.Customer != nil {
		m.Customer.DrawTo(screen)
	}

	m.counter.DrawTo(screen)

	// draw all the sprites in their draw order.
	for _, sprite := range m.Sprites {
		sprite.DrawTo(screen)
	}

	// draw dialogue bubbles.
	if m.bubbles.DrawTo(screen) { // draw options only if the bubbles are already totally drawn
		m.drawOptions(screen)
	}
}

func (m *MainScene) drawBg(screen *ebiten.Image) {
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Scale(ScaleFactor, ScaleFactor)
	screen.DrawImage(Resources.GetImage("bg_bg.png"), opts)

	// TODO: draw commuters
	opts.GeoM.Reset()
	opts.GeoM.Scale(ScaleFactor, ScaleFactor)
	screen.DrawImage(Resources.GetImage("bg_fg.png"), opts)
}

var OptionsBounds = rect(300, 240, 280, 80)

func (m *MainScene) drawOptions(screen *ebiten.Image) {
	m.bubbles.txt.SetTarget(screen)
	feed := m.bubbles.txt.NewFeed(fixed.P(OptionsBounds.Min.X, OptionsBounds.Min.Y))
	for _, opt := range m.options {
		if opt.crawlStart.IsZero() {
			opt.crawlStart = time.Now()
		}
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
	if err := m.Runner.DoNode(m.Day.Next()); err != nil {
		debug.Printf("error starting runner: %v", err)
		return
	}
	m.Runner.running = false
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
	debug.Println("start node", name)
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
		m.speaking.Wait()
	}
	return nil
}

func (m *MainScene) Options(options []yarn.Option) (int, error) {
	m.mut.Lock()
	defer m.mut.Unlock()
	m.options = make([]*Line, 0, len(options))
	for _, opt := range options {
		m.options = append(m.options, NewOption(m.Runner.Render(opt.Line)))
	}
	m.selectOpt.Wait() // wait for the player to make a selection
	return m.selection, nil
}

func (m *MainScene) Command(command string) error {
	debug.Println("run command:", command)
	command = strings.TrimSpace(command)
	tokens := strings.Split(command, " ")
	if len(tokens) == 0 {
		return fmt.Errorf("bad command: %s", command)
	}
	switch tokens[0] {
	case "put_counter":
		return m.putCounter(tokens[1:])
	case "put_cash":
		return m.putCash(tokens[1:])
	case "put_coins":
		return m.putCoinsCmd(tokens[1:])
	case "put_cash_and_coins":
		return m.putCashAndCoins(tokens[1:])
	default:
		return fmt.Errorf("unknown command %s", tokens[0])
	}
}

func (m *MainScene) putCoinsCmd(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("call to put_coins bad arguments: %v", args)
	}
	amt, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("call to put_coins wasn't integer: %v", err)
	}
	if amt < 0 {
		return fmt.Errorf("amount passed to put_coins was negative: %v", err)
	}
	m.putCoins(amt)
	return nil
}

func (m *MainScene) putCashAndCoins(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("call to put_cash_and_coins bad arguments: %v", args)
	}
	val, err := strconv.ParseFloat(args[0], 32)
	if err != nil {
		return fmt.Errorf("call to put_cash_and_coins wasn't integer: %v", err)
	}
	val *= 100
	valInt := int(val)
	coin := valInt % 100
	bills := valInt / 100
	m.putBills(bills)
	m.putCoins(coin)
	return nil
}

func (m *MainScene) putCash(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("call to put_cash bad arguments: %v", args)
	}
	amt, err := strconv.Atoi(args[0])
	if err != nil {
		return fmt.Errorf("call to put_cash wasn't integer: %v", err)
	}
	if amt < 0 {
		return fmt.Errorf("amount passed to put_cash was negative: %v", err)
	}
	m.putBills(amt)
	return nil
}

func (m *MainScene) putCounter(args []string) error {
	/* TODO:
	id - the character's randomly generated photo id.
	check - a randomly generated check which
	cash_check - a randomly generated check made out to "cash" for cash withdrawal.
	deposit_slip - a randomly generated deposit slip and cash to match.
	withdrawal_slip - a randomly generated withdrawal slip
	*/
	for _, arg := range args {
		arg = strings.TrimSpace(arg)
		if arg == "" {
			continue
		}
		switch arg {
		case "bill_1":
			m.putBill(1)
		case "bill_5":
			m.putBill(5)
		case "bill_10":
			m.putBill(10)
		case "bill_20":
			m.putBill(20)
		case "bill_100":
			m.putBill(100)
		case "coin_1":
			m.Sprites = append(m.Sprites, newCoin(1, randCounterPos()))
		case "coin_5":
			m.Sprites = append(m.Sprites, newCoin(5, randCounterPos()))
		case "coin_10":
			m.Sprites = append(m.Sprites, newCoin(10, randCounterPos()))
		case "coin_25":
			m.Sprites = append(m.Sprites, newCoin(25, randCounterPos()))
		case "coin_50":
			m.Sprites = append(m.Sprites, newCoin(50, randCounterPos()))
		default:
			debug.Printf("unrecognized argument to put_counter: %v", arg)
		}
	}
	return nil
}

func (m *MainScene) putCoins(amt int) {
	var amts []int
	for amt > 0 {
		switch {
		case amt >= 50:
			amt -= 50
			amts = append(amts, 50)
		case amt >= 25:
			amt -= 25
			amts = append(amts, 25)
		case amt >= 10:
			amt -= 10
			amts = append(amts, 10)
		case amt >= 5:
			amt -= 5
			amts = append(amts, 5)
		case amt >= 1:
			amt -= 1
			amts = append(amts, 1)
		}
	}
	for _, amt := range amts {
		m.putCoin(amt)
	}
}

func (m *MainScene) putBills(amt int) {
	var amts []int
	for amt > 0 {
		switch {
		case amt >= 100:
			amt -= 100
			amts = append(amts, 100)
		case amt >= 20:
			amt -= 20
			amts = append(amts, 20)
		case amt >= 10:
			amt -= 10
			amts = append(amts, 10)
		case amt >= 5:
			amt -= 5
			amts = append(amts, 5)
		case amt >= 1:
			amt -= 1
			amts = append(amts, 1)
		}
	}
	for _, amt := range amts {
		m.putBill(amt)
	}
}

func (m *MainScene) putBill(denom int) {
	m.Sprites = append(m.Sprites, newBill(denom, randCounterPos()))
}

func (m *MainScene) putCoin(denom int) {
	m.Sprites = append(m.Sprites, newCoin(denom, randCounterPos()))
}

func (m *MainScene) NodeComplete(nodeName string) error {
	debug.Println("node done", nodeName)
	return nil
}

func (m *MainScene) DialogueComplete() error {
	m.bubbles.SetLine("")
	debug.Println("dialogue complete")
	return nil
}

type Portrait struct {
	*BaseSprite
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
