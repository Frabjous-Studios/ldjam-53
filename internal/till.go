package internal

import (
	"bytes"
	"fmt"
	"github.com/Frabjous-Studios/ebitengine-game-template/internal/debug"
	"image"
	"math/rand"
	"text/template"
)

const CoinTargets = 0
const BillTargets = 1

type Till struct {
	*BaseSprite

	DropTargets [2][5]image.Rectangle
	BillSlots   [5][]*Money
	CoinSlots   [5][]*Money

	StartValue int // StartValue is the starting value of the till at the beginning of the day.

	DepositSlips []*DepositSlip
}

func NewTill() *Till {
	result := &Till{
		BaseSprite: &BaseSprite{
			X: 0, Y: 172,
			Img: Resources.images["Till"],
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
	}
	return result
}

type ReconciliationReport struct {
	ValidSlips int
	WTFSlips   int

	BillCount     map[string]int
	CoinCount     map[string]int
	ExpectedValue string
	ActualValue   string
	Imbalance     string
}

func (t *Till) Reconcile() *ReconciliationReport {
	report := ReconciliationReport{
		BillCount: make(map[string]int),
		CoinCount: make(map[string]int),
	}

	expectedValue := t.StartValue
	for _, slip := range t.DepositSlips {
		if slip.ForDeposit {
			expectedValue += slip.Value
			report.ValidSlips++
		} else if slip.ForWithdrawal {
			expectedValue -= slip.Value
			report.ValidSlips++
		} else {
			report.WTFSlips++ // wtf? what is this?!
		}
	}
	for _, slots := range t.CoinSlots {
		for _, money := range slots {
			report.CoinCount[fmt.Sprintf("c%d", money.Value)]++
		}
	}
	for _, slots := range t.BillSlots {
		for _, money := range slots {
			report.BillCount[fmt.Sprintf("b%d", money.Value/100)]++
		}
	}
	report.ExpectedValue = fmt.Sprintf("%.02f", float32(expectedValue)/100)
	report.ActualValue = fmt.Sprintf("%.02f", float32(t.Value())/100)
	report.Imbalance = fmt.Sprintf("%.02f", float32(t.Value()-expectedValue)/100)

	return &report
}

var reportTemplate *template.Template

func init() {
	var err error
	const T = `     CURRENCY
--Scrip--     --Tokens--
  1: {{.BillCount.b1 | printf "%3d"}}       1: {{.CoinCount.c1 | printf "%3d"}}
  5: {{.BillCount.b5 | printf "%3d"}}       5: {{.CoinCount.c5 | printf "%3d"}}
 10: {{.BillCount.b10 | printf "%3d"}}      10: {{.CoinCount.c10 | printf "%3d"}}               
 20: {{.BillCount.b20 | printf "%3d"}}       25: {{.CoinCount.c25 | printf "%3d"}}
100: {{.BillCount.b100 | printf "%3d"}}      50: {{.CoinCount.c50 | printf "%3d"}}

--Deposit Slips--
  Valid:  {{.ValidSlips}}
Invalid:  {{.WTFSlips}}

-- RECONCILIATION --
  EXPECTED = {{.ExpectedValue}}
      TILL = {{.ExpectedValue}}
 IMBALANCE = {{.Imbalance}}
`
	reportTemplate, err = template.New("").Parse(T)
	if err != nil {
		panic(fmt.Errorf("unable to parse reconciliation template: %v", err))
	}

}

func (t *ReconciliationReport) String() string {
	var w bytes.Buffer
	err := reportTemplate.Execute(&w, t)
	if err != nil {
		debug.Printf("error executing template: %v", err)
	}
	return w.String()
}

func randPoint(dx, dy int) image.Point {
	return image.Pt(rand.Intn(dx), rand.Intn(dy))
}

func (t *Till) DropAll(sprites []Sprite) bool {
	for _, s := range sprites {
		if !t.Drop(s) {
			return false
		}
	}
	return true
}

// Drop drops the provided sprite on the Till; landing it in the location needed.
func (t *Till) Drop(s Sprite) bool {
	switch s := s.(type) {
	case *Money:
		return t.dropMoney(s)
	case *DepositSlip:
		return t.dropSlip(s)
	case *Stack:
		return t.dropStack(s)
	default:
		return false
	}
}

func (t *Till) dropStack(s *Stack) bool {
	idx := idxForDenom(s.Value)
	for i := 0; i < s.Count; i++ {
		t.dropMoney(newBill(s.Value, t.DropTargets[BillTargets][idx].Min))
	}
	// TODO: play sound
	return true
}

func idxForDenom(denom int) int {
	switch denom {
	case 1:
		return 0
	case 5:
		return 1
	case 10:
		return 2
	case 20:
		return 3
	case 100:
		return 4
	default:
		return -1
	}
}

func (t *Till) dropSlip(s *DepositSlip) bool {
	t.DepositSlips = append(t.DepositSlips, s)
	// TODO: play sound
	return true
}

func (t *Till) dropMoney(m *Money) bool {
	// TODO: play sound
	var targets [5]image.Rectangle
	if m.IsCoin {
		targets = t.DropTargets[CoinTargets]
	} else {
		targets = t.DropTargets[BillTargets]
	}
	// find drop target with max area intersection
	bestIdx, maxA := -1, 0
	for idx, rect := range targets {
		sz := m.Bounds().Intersect(rect.Add(t.Pos())).Size()
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

func (t *Till) Value() int {
	var result int
	for _, stack := range t.BillSlots {
		for _, bill := range stack {
			result += bill.Value
		}
	}
	for _, stack := range t.CoinSlots {
		for _, coin := range stack {
			result += coin.Value
		}
	}
	return result
}

// Remove removes the provided money from the Till; checking the top of each stack of bills and coins for it.
func (t *Till) Remove(s Sprite) {
	m, ok := s.(*Money)
	if !ok {
		return
	}
	for i := 0; i < 5; i++ {
		if len(t.BillSlots[i]) > 0 && t.BillSlots[i][len(t.BillSlots[i])-1] == m {
			t.BillSlots[i] = t.BillSlots[i][:len(t.BillSlots[i])-1]
			return
		}
		if len(t.CoinSlots[i]) > 0 && t.CoinSlots[i][len(t.CoinSlots[i])-1] == m {
			t.CoinSlots[i] = t.CoinSlots[i][:len(t.CoinSlots[i])-1]
			return
		}
	}
}

type Money struct {
	*BaseSprite
	Value  int // Value is in cents.
	IsCoin bool
}

// newBill creates a bill of the provided denomination in local coordinates on the counter.
func newBill(denom int, pt image.Point) *Money {
	img := Resources.images[fmt.Sprintf("bill_%d", denom)]
	return &Money{
		Value:  denom * 100,
		IsCoin: false,
		BaseSprite: &BaseSprite{
			X:   pt.X,
			Y:   pt.Y,
			Img: img,
		},
	}
}

// newCoin creates a coin of the provided denomination in local coordinates on the counter.
func newCoin(denom int, pt image.Point) *Money {
	img := Resources.images[fmt.Sprintf("coin_%d", denom)]
	return &Money{
		Value:  denom,
		IsCoin: true,
		BaseSprite: &BaseSprite{
			X:   pt.X,
			Y:   pt.Y,
			Img: img,
		},
	}
}

func newStack(denom int, pt image.Point) Sprite {
	img := Resources.images[fmt.Sprintf("stack_%d", denom)]
	return &Stack{
		Value:      denom,
		Count:      50,
		BaseSprite: &BaseSprite{X: pt.X, Y: pt.Y, Img: img},
	}
}

func randCounterPos() image.Point {
	pt := image.Pt(rand.Intn(208), rand.Intn(88))
	pt.X = clamp(pt.X+112, 112, 320-15)
	pt.Y = clamp(pt.Y+152, 152, 240-15)
	return pt
}
