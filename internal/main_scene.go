package internal

import (
	"fmt"
	"github.com/DrJosh9000/yarn"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/tinne26/etxt"
	"golang.org/x/image/colornames"
	"golang.org/x/image/math/fixed"
	"image"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ScaleFactor global scaling factor for pixel art.
const ScaleFactor = 2.0

type BaseScene struct {
}

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

type Hologram struct {
	*BaseSprite
	StartTime time.Time
}

const DayLength = time.Minute // TODO: 10 * time.Minute

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

type SceneState uint8

const (
	StateApproaching SceneState = iota
	StateConversing
	StateDismissing
	StateReporting
)

var startTime time.Time

type MainScene struct {
	Game *Game

	Day          Day // Customers is a list of Yarnspinner nodes happening on the current day
	Customer     *Customer
	CustomerName string

	Sprites       []Sprite
	ReturnedSlips []*DepositSlip

	State  SceneState
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

	speaking         *sync.Cond
	selectOpt        *sync.Cond
	reconcileDismiss *sync.Cond
	selection        int

	debouceTime  time.Time
	dayStartTime time.Time

	txt *etxt.Renderer

	report          *ReconciliationReport
	reportDismissed bool
}

const GameMusic = "FunkyJazz.ogg"

func NewMainScene(g *Game) *MainScene {
	var err error
	startTime = time.Now()
	result := &MainScene{
		Game:       g,
		Sprites:    []Sprite{},
		Day:        Days[0],
		State:      StateApproaching,
		till:       NewTill(),
		counter:    &BaseSprite{X: 112, Y: 152, Img: Resources.images["counter"]},
		terminal:   &BaseSprite{X: 0, Y: 72, Img: Resources.images["terminal"]},
		buttonBase: &BaseSprite{X: 259, Y: 147, Img: Resources.images["call_button"]},
		buttonHolo: &Hologram{
			BaseSprite: &BaseSprite{X: 263, Y: 124, Img: Resources.images["call_button_holo"]},
			StartTime:  time.Now(),
		},
		dayStartTime: time.Now(),
		vars:         make(yarn.MapVariableStorage),
	}
	// generate random bills; [5-20] each.
	for idx, denom := range []int{1, 5, 10, 20, 100} {
		count := rand.Intn(15) + 5
		for i := 0; i < count; i++ {
			bill := newBill(denom, result.till.DropTargets[BillTargets][idx].Min.Add(result.till.Pos().Add(randPoint(2, 2))))
			result.till.BillSlots[idx] = append(result.till.BillSlots[idx], bill)
			result.Sprites = append(result.Sprites, bill)
		}
	}
	// generate random coins; [10-50] each.
	for idx, denom := range []int{1, 5, 10, 25, 50} {
		count := rand.Intn(40) + 10
		for i := 0; i < count; i++ {
			coin := newCoin(denom, result.till.DropTargets[CoinTargets][idx].Min.Add(result.till.Pos().Add(randPoint(3, 3))))
			result.till.CoinSlots[idx] = append(result.till.CoinSlots[idx], coin)
			result.Sprites = append(result.Sprites, coin)
		}
	}
	result.till.StartValue = result.till.Value()

	result.bubbles = NewBubbles(result)
	result.speaking = sync.NewCond(&result.mut)
	result.selectOpt = sync.NewCond(&result.mut)
	result.reconcileDismiss = sync.NewCond(&result.mut)

	result.Runner, err = NewDialogueRunner(result.vars, result)

	result.txt = etxt.NewStdRenderer()
	result.txt.SetFont(Resources.GetFont(IndicatorFont))
	result.txt.SetAlign(etxt.Top, etxt.Left)
	result.txt.SetSizePx(6)
	g.PlayMusic(GameMusic)
	if err != nil {
		panic(err)
	}

	return result
}

const DismissalPxPerSecond = 100

func (m *MainScene) Update() error {
	if err := m.updateInput(); err != nil {
		debug.Printf("error from updateInput: %v", err)
	}
	m.bubbles.Update()

	switch m.State {
	case StateApproaching:
		if !m.Runner.running {
			go m.startRunner()
			m.Runner.running = true
		}
		// TODO: animate the customer approaching
		runnerP := m.Runner.Portrait()
		if runnerP == nil {
			m.Customer = nil
		} else if m.Customer == nil {
			// TODO: animate the customer into position
			m.Customer = runnerP
			m.Customer.SetPos(image.Pt(170, 53))
		}
		if m.Customer != nil {
			debug.Println("transition to conversing")
			m.State = StateConversing
		}
	case StateDismissing:
		m.resetDialogue()
		m.Customer.MoveX(DismissalPxPerSecond / TPS)
		if m.Customer.Pos().X > m.Game.Width/2 {
			debug.Println("transition to approaching")
			m.clearCustomer()
			m.State = StateApproaching
		}
	}
	m.maybeHoverDrone()

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

func (m *MainScene) clearCustomer() {
	m.Customer = nil
	m.Runner.portrait = nil
	m.Runner.portraitImg.Clear()
}

const HoverHeight = 0.2
const HoverSpeedPerSecond = 2 * math.Pi

func (m *MainScene) maybeHoverDrone() {
	if m.Runner == nil || m.Customer == nil {
		return
	}
	if strings.HasPrefix(m.Runner.CurrNodeName, "drone") {
		m.Customer.MoveY(HoverHeight * math.Sin(HoverSpeedPerSecond*time.Now().Sub(startTime).Seconds()))
	}
}

var NextButtonHotspot = rect(275, 151, 14, 8)

const debounceDuration = 300 * time.Millisecond

func (m *MainScene) resetDialogue() {
	m.bubbles.SetLine("")
	m.selectOpt.Broadcast()
	m.speaking.Broadcast()
	m.options = nil
}

var CustomerDropZone = rect(170, 52, 100, 100)

// updateInput is debounced.
func (m *MainScene) updateInput() error {

	newKeys = inpututil.AppendJustPressedKeys(newKeys[:0])
	heldKeys = inpututil.AppendPressedKeys(heldKeys[:0])

	if m.State == StateReporting {
		m.bubbles.BeDone()
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			fmt.Println("report dismissed!")
			m.reportDismissed = true
			if m.reportDismissed {
				fmt.Println("state transition back to conversing")
				m.State = StateConversing
			}
			m.bubbles.TextBounds = DialogueBounds
			m.bubbles.SetLine("")
			m.reconcileDismiss.Broadcast()
		}
		return nil
	}

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
			if cPos.In(CustomerDropZone) {
				m.customerDrop()
			} else if contains(heldKeys, ebiten.KeyShift) {
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
					m.tillDrop()
				} else {
					debug.Println("counter drop")
					m.counterDrop()
				}
			}
		} else { // pick up the thing under the cursor
			if cPos.In(NextButtonHotspot) {
				m.nextButton()
			} else {
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
		}
		m.debouceTime = time.Now().Add(debounceDuration)
	}
	return nil
}

func (m *MainScene) depart() error {
	m.State = StateDismissing // without playing a sound
	return nil
}

func (m *MainScene) nextButton() {
	s := Resources.GetSound(m.Game.ACtx, "Bell-1.ogg")
	s.Rewind()
	s.Play()
	if m.Customer != nil && m.Customer.ImageKey == "manager.png" {
		m.bubbles.SetLine("What?! You think you can dismiss me?!!? I wasn't through talking!") // TODO: randomize
	} else {
		m.State = StateDismissing
	}
}

// cheatValue is some random value added to required withdrawal thresholds for the customer to walk away on their own.
// This keeps the player from letting the customer do their own counting.
func cheatValue() int {
	if rand.Float64() < 0.7 {
		return 0.0
	} else if rand.Float64() < 0.7 {
		return rand.Intn(50)
	} else {
		return rand.Intn(1000) // greedy little bastard
	}
}

func (m *MainScene) customerDrop() {
	if len(m.holding) == 0 {
		return
	}
	debug.Println("dropping on customer!")
	if m.Customer != nil {
		if _, ok := m.holding[0].(*Money); ok {
			totalValue := 0 // figure out the value of this fist full o' cash.
			for _, m := range m.holding {
				totalValue += m.(*Money).Value
			}
			// giving the customer money
			if m.Customer.CustomerIntent == IntentDeposit {
				m.Customer.CashOnCounter -= totalValue
				if m.Customer.DepositSlip != nil && m.Customer.DepositSlip.Value > m.Customer.CashOnCounter { // put cash back to even out deposit
					m.bubbles.SetLine("I'd like to deposit that, actually.") // TODO: randomize
					diff := m.Customer.DepositSlip.Value - m.Customer.CashOnCounter
					m.putCashAndCoinsf(float32(diff) / 100) // make other money out of thin air; I'm trying to deposit; dammit. I won't leave until I do!
				}
			} else if m.Customer.CustomerIntent == IntentWithdraw {
				m.Customer.CashInHand += totalValue
				if m.Customer.DepositSlip != nil && m.Customer.CashInHand+cheatValue() >= m.Customer.DepositSlip.Value {
					m.bubbles.SetLine("Thank you!") // TODO: randomize
					m.depart()
				}
			} // TODO: other intents

			for _, held := range m.holding {
				m.removeSprite(held)
			}
			m.holding = nil
		} else if slip, ok := m.holding[0].(*DepositSlip); ok {
			m.bubbles.SetLine("What? You mean I have to get back in line?!")
			m.removeSprite(m.holding[0])
			m.ReturnedSlips = append(m.ReturnedSlips, slip) // we'll check these at the end of the day.
			m.holding = nil
			// giving back their deposit slip.
		} else if _, ok := m.holding[0].(*Stack); ok {
			if m.Customer.ImageKey == "drone.png" {
				// put it back; TODO: play an error sound?
				m.holding[0].SetPos(randCounterPos())
			} else {
				// you're giving away a stack of money?!!?! Yes please!
				m.bubbles.SetLine("Oh! Thank you!") // TODO: randomize
				m.removeSprite(m.holding[0])
				m.holding = nil
				m.depart() // TODO: this probably breaks the line
			}
		}
	}
}

func (m *MainScene) tillDrop() {
	if m.till.DropAll(m.holding) {
		if _, ok := m.holding[0].(*DepositSlip); ok {
			m.removeSprite(m.holding[0])
		}
		if _, ok := m.holding[0].(*Stack); ok {
			m.removeSprite(m.holding[0])
		}
		m.holding = nil
	} else {
		debug.Println("unable to drop on till")
	}
}

func (m *MainScene) counterDrop() {
	for _, s := range m.holding { // drop ALL
		s.SetPos(clampToCounter(s.Pos()))
	}
	m.soundDrop(m.holding[0], "counter")
	m.holding = nil
}

// removeSprite removes the provided sprite from the draw queue.
func (m *MainScene) removeSprite(s Sprite) {
	for idx, o := range m.Sprites {
		if s == o {
			m.Sprites = append(m.Sprites[:idx], m.Sprites[idx+1:]...)
			return
		}
	}
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
		if grabbedMoney && haveMoney && c1.IsCoin == c2.IsCoin && c1.Value == c2.Value {
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

	// draw reconciliation report
	if m.State == StateReporting && !m.reportDismissed {
		m.drawReconciliationReport(screen)
	} else {
		// draw dialogue bubbles.
		if m.bubbles.DrawTo(screen) { // draw options only if the bubbles are already totally drawn
			m.drawOptions(screen)
		}
	}

	// draw cash indicator
	m.drawCashIndicator(screen)

	ebitenutil.DebugPrintAt(screen, fmt.Sprintf("%.0f", ebiten.ActualFPS()), 620, 0) // TODO: remove!
}

var IndicatorColor = h2c("00ff00")

const IndicatorFont = "Munro"
const IndicatorFontSize = 36
const IndicatorOffset = -40

func (m *MainScene) drawCashIndicator(screen *ebiten.Image) {
	if len(m.holding) == 0 {
		return
	}
	var (
		value, fracVal int
		isCoin         bool
	)

	if money, ok := m.holding[0].(*Money); ok {
		value = len(m.holding) * money.Value / 100
		fracVal = len(m.holding) * money.Value
		isCoin = money.IsCoin
	} else if stack, ok := m.holding[0].(*Stack); ok {
		value = stack.Count * stack.Value
		isCoin = false
	} else {
		return
	}

	cPos := cursorPos()
	m.txt.SetColor(IndicatorColor)
	m.txt.SetSizePx(IndicatorFontSize)
	m.txt.SetTarget(screen)
	v, h := m.txt.GetAlign()
	m.txt.SetAlign(etxt.YCenter, etxt.XCenter)
	if isCoin {
		m.txt.Draw(fmt.Sprintf("$%.02f", float32(fracVal)/100), ScaleFactor*cPos.X, ScaleFactor*cPos.Y+IndicatorOffset)
	} else {
		m.txt.Draw(fmt.Sprintf("$%d", value), ScaleFactor*cPos.X, ScaleFactor*cPos.Y+IndicatorOffset)
	}

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

var OptionsBounds = rect(300, 240, 240, 80)

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

var ReportBounds = rect(96*2, 20*2, 114*2, 178*2)

func (m *MainScene) drawReconciliationReport(screen *ebiten.Image) {
	m.bubbles.DrawTo(screen)
}

func (m *MainScene) dayLength() time.Duration {
	return time.Now().Sub(m.dayStartTime)
}

func (m *MainScene) startRunner() {
	if err := m.Runner.DoNode(m.Day.Next(m.dayLength())); err != nil {
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

// play the sound of dropping something on the counter
func (m *MainScene) soundDrop(s Sprite, surface string) {
	switch s := s.(type) {
	case *Money:
		if !s.IsCoin {
			return
		}
		switch surface {
		case "counter":
			p := Resources.GetRandSound(m.Game.ACtx, "Coin_Drop-1.ogg", "Coin_Drop-2.ogg")
			p.Rewind()
			p.Play()
		}
	}
}

func (m *MainScene) spriteUnderCursor() Sprite {
	for i := len(m.Sprites) - 1; i >= 0; i-- {
		if cursorPos().In(m.Sprites[i].Bounds()) && !m.isHeld(m.Sprites[i]) {
			return m.Sprites[i]
		}
	}
	return nil
}
func (m *MainScene) isHeld(sprite Sprite) bool {
	for _, t := range m.holding {
		if t == sprite {
			return true
		}
	}
	return false
}

func (m *MainScene) NodeStart(name string) error {
	debug.Println("start node", name)
	if m.State == StateDismissing {
		return yarn.Stop
	}
	return nil
}

func (m *MainScene) PrepareForLines(lineIDs []string) error {
	if m.State == StateDismissing {
		return yarn.Stop
	}
	return nil
}

func (m *MainScene) Line(line yarn.Line) error {
	m.mut.Lock()
	defer m.mut.Unlock()
	m.bubbles.SetLine(m.Runner.Render(line))

	for !m.bubbles.IsDone() && !m.Runner.IsLastLine(line) {
		m.speaking.Wait()
	}
	if m.State == StateDismissing {
		return yarn.Stop
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
	if m.State == StateDismissing {
		return 0, yarn.Stop
	}
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
	case "play_sound":
		return m.playSound(tokens[1:])
	case "depart":
		return m.depart()
	case "set_wrong":
		return m.setWrong()
	case "show_reconciliation_report":
		return m.showReconciliationReport()
	default:
		return fmt.Errorf("unknown command %s", tokens[0])
	}
}

func (m *MainScene) showReconciliationReport() error {
	m.mut.Lock()
	defer m.mut.Unlock()

	m.report = m.till.Reconcile()
	m.bubbles.TextBounds = ReportBounds
	m.bubbles.SetLine(m.report.String())
	m.State = StateReporting

	m.reconcileDismiss.Wait()

	return nil
}

func (m *MainScene) setWrong() error {
	if m.Customer == nil {
		debug.Println("set_wrong called with nil customer")
		return nil
	}
	if m.Customer.DepositSlip != nil {
		m.Customer.DepositSlip.IsWrong = true
	}
	// TODO: set for checks and such as well.
	return nil
}

func (m *MainScene) NodeComplete(nodeName string) error {
	debug.Println("node done", nodeName)
	if m.State == StateDismissing {
		return yarn.Stop
	}
	return nil
}

func (m *MainScene) DialogueComplete() error {
	m.resetDialogue()
	debug.Println("dialogue complete")
	if m.State == StateDismissing {
		return yarn.Stop
	}
	return nil
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
	if m.State == StateDismissing {
		return yarn.Stop
	}
	return nil
}

func (m *MainScene) playSound(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("call to play_sound had bad number of arguments: %v", args)
	}
	player := Resources.GetSound(m.Game.ACtx, args[0])
	if player == nil {
		return fmt.Errorf("call to play_sound with missing sound file: %v", args[0])
	}
	player.Rewind()
	player.Play()
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
	m.putCashAndCoinsf(float32(val))
	if m.State == StateDismissing {
		return yarn.Stop
	}
	return nil
}

func (m *MainScene) putCashAndCoinsf(val float32) {
	val *= 100
	valInt := int(val)
	coin := valInt % 100
	bills := valInt / 100
	m.putBills(bills)
	m.putCoins(coin)
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
	if m.State == StateDismissing {
		return yarn.Stop
	}
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
		switch {
		case arg == "empty_slip":
			slip := m.randEmptySlip()
			m.Runner.SetDepositSlip(slip)
			m.put(slip)
		case strings.HasPrefix(arg, "deposit_slip"):
			slip := m.randDepositSlip()
			if len(arg) > 13 {
				v, err := strconv.Atoi(strings.TrimPrefix(arg, "deposit_slip_"))
				if err != nil {
					debug.Println("bad call to put_counter with deposit_slip value with value:", arg)
				} else {
					slip.Value = v
				}
			}
			m.Runner.SetDepositSlip(slip)
			m.put(slip)
			m.putBills(slip.Value / 100)
		case strings.HasPrefix(arg, "withdrawal_slip"):
			slip := m.randWithdrawalSlip()
			if len(arg) > 17 {
				v, err := strconv.Atoi(strings.TrimPrefix(arg, "withdrawal_slip_"))
				if err != nil {
					debug.Println("bad call to put_counter with deposit_slip value with value:", arg)
				} else {
					slip.Value = v
				}
			}
			m.Runner.SetDepositSlip(slip)
			m.put(slip)
		case arg == "bill_1":
			m.putBill(1)
		case arg == "bill_5":
			m.putBill(5)
		case arg == "bill_10":
			m.putBill(10)
		case arg == "bill_20":
			m.putBill(20)
		case arg == "bill_100":
			m.putBill(100)
		case arg == "stack_1":
			m.putStack(1)
		case arg == "stack_5":
			m.putStack(5)
		case arg == "stack_10":
			m.putStack(10)
		case arg == "stack_20":
			m.putStack(20)
		case arg == "stack_100":
			m.putStack(100)
		case arg == "coin_1":
			m.Sprites = append(m.Sprites, newCoin(1, randCounterPos()))
		case arg == "coin_5":
			m.Sprites = append(m.Sprites, newCoin(5, randCounterPos()))
		case arg == "coin_10":
			m.Sprites = append(m.Sprites, newCoin(10, randCounterPos()))
		case arg == "coin_25":
			m.Sprites = append(m.Sprites, newCoin(25, randCounterPos()))
		case arg == "coin_50":
			m.Sprites = append(m.Sprites, newCoin(50, randCounterPos()))
		default:
			debug.Printf("unrecognized argument to put_counter: %v", arg)
		}
	}
	if m.State == StateDismissing {
		return yarn.Stop
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

func (m *MainScene) putStack(denom int) {
	m.Sprites = append(m.Sprites, newStack(denom, randCounterPos()))
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

type Stack struct {
	*BaseSprite
	Value int
	Count int
}

type DepositSlip struct {
	*BaseSprite
	Value         int
	ForDeposit    bool
	ForWithdrawal bool
	AcctNum       int
	IsWrong       bool // IsWrong means the customer did not fill out this paperwork correctly.
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
	if m.Customer != nil {
		m.Customer.CashOnCounter += denom * 100
	}
	m.Sprites = append(m.Sprites, newBill(denom, randCounterPos()))
}

func (m *MainScene) putCoin(denom int) {
	if m.Customer != nil {
		m.Customer.CashOnCounter += denom
	}
	m.Sprites = append(m.Sprites, newCoin(denom, randCounterPos()))
}

func (m *MainScene) Reconcile() {
	m.till.Reconcile()
}

type Intent string

const (
	IntentWithdraw     = "withdraw"
	IntentDeposit      = "deposit"
	IntentCashCheck    = "cash_check"
	IntentDepositCheck = "deposit_check"
)

type Customer struct {
	*BaseSprite
	MoneyInHand    []*Money
	CashInHand     int
	CashOnCounter  int // CashOnCounter is the total value of the cash the customer has put on the counter.
	ImageKey       string
	CustomerIntent Intent
	CustomerName   string
	DepositSlip    *DepositSlip // DepositSlip may be nil for some customers.
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

func contains[T comparable](arr []T, val T) bool {
	for _, t := range arr {
		if t == val {
			return true
		}
	}
	return false
}
