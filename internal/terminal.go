package internal

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/tinne26/etxt"
	"golang.org/x/exp/maps"
	"image/color"
	"math"
	"time"
)

type Terminal struct {
	*BaseSprite
	scene *MainScene // yay coupling!!
	txt   *etxt.Renderer

	operational bool

	bg *ebiten.Image

	accountNumber []rune
	keyDebounce   time.Time

	lines []string
}

func NewTerminal(txt *etxt.Renderer, scene *MainScene) *Terminal {
	result := &Terminal{
		scene: scene,
		txt:   txt,
		BaseSprite: &BaseSprite{
			X: 0,
			Y: 72,
		},
		operational: true, // TODO: turn this OFF for the first day!
	}
	result.bg = Resources.images["terminal"]
	result.Img = ebiten.NewImage(result.bg.Bounds().Dx(), result.bg.Bounds().Dy())

	return result
}

func (t *Terminal) DrawTo(screen *ebiten.Image) {
	t.Img.Clear()
	// draw background
	opts := &ebiten.DrawImageOptions{}
	t.Img.DrawImage(t.bg, opts)

	if !t.operational {
		t.BaseSprite.DrawTo(screen)
		return
	}

	const size = 10
	const lineHeight = 2
	t.txt.SetTarget(t.Img)
	t.txt.SetSizePx(size)
	t.txt.SetFont(Resources.GetFont(DialogFont))
	t.txt.SetColor(color.White)
	t.txt.SetAlign(etxt.Top, etxt.Left)
	t.txt.Draw("Account:", 5, 5)

	t.txt.Draw(t.inputField(), 46, 5)

	y := 5 + size + lineHeight
	for _, line := range t.lines {
		t.txt.Draw(line, 5, y)
		y += size + lineHeight
	}

	t.BaseSprite.DrawTo(screen)
}

func (t *Terminal) inputField() string {
	result := string(t.accountNumber)
	if math.Sin(time.Now().Sub(t.keyDebounce).Seconds()*2*math.Pi) > 0 {
		return result + "_"
	}
	return result
}

func (t *Terminal) Update() {
	t.handleKeys()
}

func (t *Terminal) handleKeys() {
	if time.Now().Before(t.keyDebounce) {
		return
	}
	for _, key := range heldKeys {
		switch key {
		case ebiten.KeyDigit0:
			t.appendNumber('0')
		case ebiten.KeyDigit1:
			t.appendNumber('1')
		case ebiten.KeyDigit2:
			t.appendNumber('2')
		case ebiten.KeyDigit3:
			t.appendNumber('3')
		case ebiten.KeyDigit4:
			t.appendNumber('4')
		case ebiten.KeyDigit5:
			t.appendNumber('5')
		case ebiten.KeyDigit6:
			t.appendNumber('6')
		case ebiten.KeyDigit7:
			t.appendNumber('7')
		case ebiten.KeyDigit8:
			t.appendNumber('8')
		case ebiten.KeyDigit9:
			t.appendNumber('9')
		case ebiten.KeyBackspace:
			t.backspace()
		}
	}
}

func union[T comparable](A, B []T) []T {
	keys := make(map[T]struct{})
	for _, a := range A {
		keys[a] = struct{}{}
	}
	for _, b := range B {
		keys[b] = struct{}{}
	}
	return maps.Keys(keys)
}

func (t *Terminal) backspace() {
	if len(t.accountNumber) == 0 {
		return
	}
	t.lines = nil
	t.accountNumber = t.accountNumber[:len(t.accountNumber)-1]
	t.keyDebounce = time.Now().Add(150 * time.Millisecond)
}

func (t *Terminal) appendNumber(n rune) {
	if len(t.accountNumber) == 5 {
		return
	}
	t.accountNumber = append(t.accountNumber, n)
	t.keyDebounce = time.Now().Add(150 * time.Millisecond)
	t.lookup()
}

func (t *Terminal) lookup() {
	if len(t.accountNumber) < 5 {
		return
	}
	acct, ok := t.scene.Day.Accounts[t.GetAccountNumber()]
	if acct == nil || !ok {
		t.lines = []string{"--ACCOUNT NOT FOUND--"}
		return
	}
	t.lines = []string{
		fmt.Sprintf("Owner: %s", acct.Owner),
		fmt.Sprintf("Checking Balance: %.02f", float32(acct.Checking)/100.0),
	}
}

func (t *Terminal) GetAccountNumber() string {
	if len(t.accountNumber) != 5 {
		return ""
	}
	return string(t.accountNumber)
}
