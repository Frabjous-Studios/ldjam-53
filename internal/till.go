package internal

import (
	"fmt"
	"image"
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
	// TODO: generate random bills in the Till
	return &Till{
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
}

// Drop drops the provided sprite on the Till; landing it in the location needed.
func (t *Till) Drop(s Sprite) bool {
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

// Remove removes the provided money from the Till; checking the top of each stack of bills and coins for it.
func (t *Till) Remove(s Sprite) {
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

type Money struct {
	*BaseSprite
	Value  int // Value is in cents.
	IsCoin bool
}

// newBill creates a bill of the provided denomination in local coordinates on the counter.
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

// newCoin creates a coin of the provided denomination in local coordinates on the counter.
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
