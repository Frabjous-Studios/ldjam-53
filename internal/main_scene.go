package internal

import (
	"fmt"
	"github.com/DrJosh9000/yarn"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
	"golang.org/x/image/colornames"
	"golang.org/x/image/math/fixed"
	"image"
	"math/rand"
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

type Hologram struct {
	*BaseSprite
	StartTime time.Time
}

func (s *Hologram) DrawTo(screen *ebiten.Image) {
	if s.Img == nil {
		debug.Println("image for sprite was nil at point:", s.X, s.Y)
		return
	}
	opt := &ebiten.DrawImageOptions{}
	opt.Blend = ebiten.BlendSourceOver
	opt.GeoM.Translate(float64(s.X), float64(s.Y))
	opt.GeoM.Scale(ScaleFactor, ScaleFactor)
	opt.ColorScale.Scale(1.0, 1.0, 1.0, 0.7) // TODO: animate this with dT using a curve
	screen.DrawImage(s.Img, opt)
}

type MainScene struct {
	Game *Game

	Day          Day // Customers is a list of Yarnspinner nodes happening on the current day
	Customer     Sprite
	CustomerName string

	Sprites []Sprite

	State  *GameState
	Runner *DialogueRunner

	till       *Till
	counter    *BaseSprite
	terminal   *BaseSprite
	buttonBase *BaseSprite
	buttonHolo *Hologram

	bubbles *Bubbles
	options []*Line

	holding     []Sprite
	clickStart  image.Point
	clickOffset image.Point

	vars yarn.MapVariableStorage
	mut  sync.Mutex

	speaking  *sync.Cond
	selectOpt *sync.Cond
	selection int

	debouceTime time.Time

	txt *etxt.Renderer
}

func NewMainScene(g *Game) *MainScene {
	var err error
	result := &MainScene{
		Game:    g,
		Sprites: []Sprite{},
		Day:     Days[0],

		till:       NewTill(),
		counter:    &BaseSprite{X: 112, Y: 152, Img: Resources.images["counter"]},
		terminal:   &BaseSprite{X: 0, Y: 72, Img: Resources.images["terminal"]},
		buttonBase: &BaseSprite{X: 259, Y: 147, Img: Resources.images["call_button"]},
		buttonHolo: &Hologram{
			BaseSprite: &BaseSprite{X: 263, Y: 124, Img: Resources.images["call_button_holo"]},
			StartTime:  time.Now(),
		},
		vars: make(yarn.MapVariableStorage),
	}
	// generate random bills; [5-20] each.
	for idx, denom := range []int{1, 5, 10, 20, 100} {
		count := rand.Intn(15) + 5
		for i := 0; i < count; i++ {
			bill := newBill(denom, result.till.DropTargets[BillTargets][idx].Min.Add(result.till.Pos().Add(randPoint(2, 2))))
			result.till.BillSlots[idx] = append(result.till.BillSlots[idx], bill.(*Money))
			result.Sprites = append(result.Sprites, bill)
		}
	}
	// generate random coins; [10-50] each.
	for idx, denom := range []int{1, 5, 10, 25, 50} {
		count := rand.Intn(40) + 10
		for i := 0; i < count; i++ {
			coin := newCoin(denom, result.till.DropTargets[CoinTargets][idx].Min.Add(result.till.Pos().Add(randPoint(3, 3))))
			result.till.CoinSlots[idx] = append(result.till.CoinSlots[idx], coin.(*Money))
			result.Sprites = append(result.Sprites, coin)
		}
	}
	result.bubbles = NewBubbles(result)
	result.speaking = sync.NewCond(&result.mut)
	result.selectOpt = sync.NewCond(&result.mut)

	result.Runner, err = NewDialogueRunner(result.vars, result)

	result.txt = etxt.NewStdRenderer()
	result.txt.SetAlign(etxt.Top, etxt.Left)
	result.txt.SetSizePx(6)
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
		m.CustomerName = m.Runner.RandomName()
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
		for _, h := range m.holding { // TODO: does this work?
			h.SetPos(mPos.Add(m.clickOffset))
		}
	}

	return nil
}

const debounceDuration = 100 * time.Millisecond

// updateInput is debounced.
func (m *MainScene) updateInput() error {

	if !m.Runner.running {
		go m.startRunner()
		m.Runner.running = true
	}

	newKeys = inpututil.AppendJustPressedKeys(newKeys[:0])
	heldKeys = inpututil.AppendPressedKeys(heldKeys[:0])

	if time.Now().Before(m.debouceTime) {
		return nil
	}

	if len(newKeys) > 0 {
		m.AdvanceDialogue()
		m.debouceTime = time.Now().Add(debounceDuration)
	}

	cPos := cursorPos()
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		debug.Println("left mouse press", m.holding)
		if len(m.holding) > 0 {
			debug.Println("holding something")
			if contains(heldKeys, ebiten.KeyShift) {
				// TODO: grab all the sprites under cursor?? if they match??
				grabbed := m.spriteUnderCursor()
				if grabbed != nil {
					m.handleGrabbed(grabbed)
				} else {
					debug.Println("counter drop")
					m.counterDrop()
				}
			} else {
				if cPos.In(m.till.Bounds()) { // if over Till; drop on Till
					debug.Println("over till")
					if m.till.DropAll(m.holding) {
						m.holding = nil
					} else {
						debug.Println("unable to drop on till")
					}
				} else {
					debug.Println("counter drop")
					m.counterDrop()
				}
			}
		} else { // pick up the thing under the cursor
			grabbed := m.spriteUnderCursor()
			if grabbed != nil {
				m.handleGrabbed(grabbed)
			} else {
				debug.Println("grabbed nothing; dialog select?")
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
					debug.Println("no dialogue selected; player is impatient.")
					m.AdvanceDialogue()
				}
			}
		}
		m.debouceTime = time.Now().Add(debounceDuration)
	}
	return nil
}

func (m *MainScene) handleGrabbed(grabbed Sprite) {
	debug.Println("grabbed sprite", grabbed)
	cPos := cursorPos()
	m.BringToFront(grabbed)
	if len(m.holding) == 0 {
		m.addHolding(grabbed)
	} else { // only pick up money in stacks of the same denomination.
		c1, grabbedMoney := grabbed.(*Money)
		c2, haveMoney := m.holding[0].(*Money)
		if grabbedMoney && haveMoney && c1.Value == c2.Value {
			m.addHolding(grabbed)
		}
	}
	m.clickStart = cPos
	m.clickOffset = grabbed.Pos().Sub(m.clickStart)
	m.till.Remove(grabbed) // remove it from the Till (maybe)
}

func (m *MainScene) addHolding(grabbed Sprite) {
	for _, h := range m.holding {
		if grabbed == h {
			return
		}
	}
	m.holding = append(m.holding, grabbed)
}

func contains[T comparable](arr []T, val T) bool {
	for _, t := range arr {
		if t == val {
			return true
		}
	}
	return false
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

	// draw terminal
	m.terminal.DrawTo(screen)

	// draw next button
	m.buttonBase.DrawTo(screen)
	m.buttonHolo.DrawTo(screen)

	// draw all the sprites in their draw order.
	for _, sprite := range m.Sprites {
		sprite.DrawTo(screen)
	}

	// draw dialogue bubbles.
	if m.bubbles.DrawTo(screen) { // draw options only if the bubbles are already totally drawn
		m.drawOptions(screen)
	}

	// draw cash indicator
	m.drawCashIndicator(screen)
}

var IndicatorColor = h2c("00ff00")

const IndicatorFontSize = 36
const IndicatorOffset = -40

func (m *MainScene) drawCashIndicator(screen *ebiten.Image) {
	if len(m.holding) < 2 {
		return
	}
	money, ok := m.holding[0].(*Money)
	if !ok {
		return
	}
	value := len(m.holding) * money.Value / 100

	cPos := cursorPos()
	m.txt.SetColor(IndicatorColor)
	m.txt.SetSizePx(IndicatorFontSize)
	m.txt.SetTarget(screen)
	v, h := m.txt.GetAlign()
	m.txt.SetAlign(etxt.YCenter, etxt.XCenter)
	m.txt.Draw(fmt.Sprintf("$%d", value), ScaleFactor*cPos.X, ScaleFactor*cPos.Y+IndicatorOffset)
	m.txt.SetAlign(v, h)
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
	grabbed := m.spriteUnderCursor()
	if grabbed != nil {
		m.holding = append(m.holding, grabbed)
		m.clickStart = cursorPos()
		m.clickOffset = grabbed.Pos().Sub(m.clickStart)
	}
}

func (m *MainScene) counterDrop() {
	for _, s := range m.holding { // drop ALL
		s.SetPos(clampToCounter(s.Pos()))
	}
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
		case "empty_slip":
			slip := m.randEmptySlip()
			m.Runner.SetDepositSlip(slip)
			m.put(slip)
		case "deposit_slip":
			slip := m.randDepositSlip()
			m.Runner.SetDepositSlip(slip)
			m.put(slip)
			m.putBills(slip.Value / 100)
		case "withdrawal_slip":
			slip := m.randWithdrawalSlip()
			m.Runner.SetDepositSlip(slip)
			m.put(slip)
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

type DepositSlip struct {
	*BaseSprite
	Value         int
	ForDeposit    bool
	ForWithdrawal bool
	AcctNum       int
}

var depositSlipColor = colornames.Black

func (m *MainScene) put(sprite Sprite) {
	m.Sprites = append(m.Sprites, sprite)
}

func (m *MainScene) randEmptySlip() *DepositSlip {
	return m.randSlip("deposit_slip.png")
}

func (m *MainScene) randDepositSlip() *DepositSlip {
	result := m.randSlip("deposit_slip_deposit.png")
	result.ForDeposit = true
	return result
}

func (m *MainScene) randWithdrawalSlip() *DepositSlip {
	result := m.randSlip("deposit_slip_withdrawal.png")
	result.ForWithdrawal = true
	return result
}

var MaxTransactionValue = 1000 // TODO: make this go _DOWN_ as the days go on.

func (m *MainScene) randSlip(path string) *DepositSlip {
	img := ebiten.NewImage(43, 32)
	img.DrawImage(Resources.GetImage(path), nil)

	pos := randCounterPos()
	slip := &DepositSlip{
		AcctNum:    randomAcctNumber(),
		Value:      randomTransactionValue(),
		BaseSprite: &BaseSprite{Img: img, X: pos.X, Y: pos.Y},
	}

	m.txt.SetColor(depositSlipColor)
	m.txt.SetSizePx(10)
	m.txt.SetFont(Resources.GetFont(DialogFont))
	m.txt.SetTarget(img)
	m.txt.Draw(fmt.Sprintf("#%d", slip.AcctNum), 14, 2)

	m.txt.SetFont(Resources.GetFont(DialogFont)) // TODO: make look like handwriting
	m.txt.SetSizePx(10)
	m.txt.SetTarget(img)
	m.txt.Draw(fmt.Sprintf("%d.00", slip.Value/100), 16, 17)

	// TODO: signature?
	return slip
}

func randomAccountValue() int {
	return rand.Intn(10000) // TODO: make this more realistic
}

func randomTransactionValue() int {
	return rand.Intn(MaxTransactionValue) * 100 // TODO: make this more realistic
}

func randomAcctNumber() int {
	return rand.Intn(89999) + 10000
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
