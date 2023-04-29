package internal

import (
	"fmt"
	"image"
	"math/rand"
)

const CoinTargets = 0
const BillTargets = 1

type Till struct {
	*BaseSprite

	DropTargets [2][5]image.Rectangle
	BillSlots   [5][]*Money
	CoinSlots   [5][]*Money
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
	m, ok := s.(*Money)
	if !ok {
		return false
	}
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
func newBill(denom int, pt image.Point) Sprite {
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
func newCoin(denom int, pt image.Point) Sprite {
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

func randCounterPos() image.Point {
	pt := image.Pt(rand.Intn(208), rand.Intn(88))
	pt.X = clamp(pt.X+112, 112, 320-15)
	pt.Y = clamp(pt.Y+152, 152, 240-15)
	return pt
}
