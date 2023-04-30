package internal

import (
	"math/rand"
	"strings"
	"time"
)

type Account struct {
	Owner    string
	Number   string
	Checking int
}

type Day struct {
	// Sequence is a sequence of YarnSpinner nodes; the node 'random' is replaced by one of the random nodes in
	// random. There is an implicit infinite string of random nodes at the end of the day.
	Sequence []string

	// Random is the list of random YarnSpinner nodes that can occur this day.
	Random []string

	// SlipsAccepted
	Slips []*DepositSlip

	EndNode string

	Accounts map[string]*Account

	curr int
}

func Days() []*Day {
	result := []*Day{
		0: { // more deposits than withdrawals
			EndNode: "Manager_Day1_End",
			// TODO: manager intro day
			Sequence: []string{"random", "random", "random", "drone", "random", "random", "OldMan"},
			Random:   []string{"RandomDeposit_Polite", "RandomDeposit_Rude", "RandomDeposit_Polite", "RandomDeposit_Rude", "RandomWithdrawal_Polite", "RandomWithdrawal_Rude"},
		},
		1: {
			EndNode:  "Manager_Day2_End",
			Sequence: []string{},
			Random:   []string{"RandomDeposit_Polite", "RandomDeposit_Rude", "RandomDeposit_Polite", "RandomDeposit_Rude", "RandomWithdrawal_Polite", "RandomWithdrawal_Rude"},
		},
	}
	for _, day := range result {
		rand.Shuffle(len(day.Random), func(i, j int) {
			day.Random[i], day.Random[j] = day.Random[j], day.Random[i]
		})
		day.Accounts = make(map[string]*Account)
		// TODO: create some initial random accounts.
	}
	return result
}

// Next retrieves the next node on the given day. Pass in the amount of time spent on this day to determine when to
// trigger the manager for reconciliation.
func (d *Day) Next(t time.Duration) string {
	if t >= DayLength {
		return d.EndNode // day over! Manager time!!!
	}
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
