package internal

import (
	"math/rand"
	"strings"
)

// Days is the list of Yarnspinner nodes that happen each day... in this order.
var Days = []Day{
	0: {
		Sequence: []string{"random"},
		Random:   []string{"Test1"},
	},
}

type Account struct {
	Owner    string
	Number   int
	Checking int
	// Savings  int
}

type Day struct {
	// Sequence is a sequence of YarnSpinner nodes; the node 'random' is replaced by one of the random nodes in
	// random. There is an implicit infinite string of random nodes at the end of the day.
	Sequence []string

	// Random is the list of random YarnSpinner nodes that can occur this day.
	Random []string

	// SlipsAccepted
	Slips []*DepositSlip

	Accounts map[int]*Account

	curr int
}

// Next retrieves the next node on the given day.
func (d *Day) Next() string {
	defer func() { // increment d.curr no matter what
		d.curr = d.curr + 1
	}()
	curr := d.curr
	if curr >= len(d.Sequence) {
		curr -= len(d.Sequence)
		if curr >= len(d.Random) { // we're repeating
			return d.Random[rand.Intn(len(d.Random))] // randomly sample from the list
		}
		return d.Random[curr]
	}
	if strings.ToLower(d.Sequence[curr]) == "random" {
		return d.Random[rand.Intn(len(d.Random))]
	}
	return d.Sequence[curr]
}

func (d *Day) AcceptSlip(slip *DepositSlip) {
	d.Slips = append(d.Slips, slip)
}

func init() {
	// shuffle the decks
	for _, day := range Days {
		rand.Shuffle(len(day.Random), func(i, j int) {
			day.Random[i], day.Random[j] = day.Random[j], day.Random[i]
		})
	}
}
