package game

import (
	"github.com/just-jane-inc/just-poker/server/just"
)

type (
	PlayerIntent string
	PlayerState  string
)

// reason: leave me alone, snake case is nice here.
//
//goland:noinspection GoSnakeCaseUsage
const (
	PlayerStateUnset    PlayerState = "unset"
	PlayerStateInactive PlayerState = "inactive"
	PlayerStateActive   PlayerState = "active"
	PlayerStateFolded   PlayerState = "folded"
	PlayerStateAllIn    PlayerState = "all_in"
	PlayerStateWon      PlayerState = "won"
	PlayerStateOut      PlayerState = "out"
)

// reason: leave me alone, snake case is nice here.
//
//goland:noinspection GoSnakeCaseUsage
const (
	PlayerIntentUnset = "unset"
	PlayerIntentAnte  = "ante"
	PlayerIntentCheck = "check"
	PlayerIntentCall  = "call"
	PlayerIntentRaise = "raise"
	PlayerIntentAllIn = "all_in"
	PlayerIntentFold  = "fold"
)

type stack map[int]int

type player struct {
	UserID          string
	DisplayName     string
	UserType        just.UserType
	state           PlayerState
	position        int
	pocket          []*card
	chips           stack
	currentBet      stack
	potContribution int
}

func (s stack) Sum() int {
	total := 0
	for d, c := range s {
		total += (d * c)
	}

	return total
}
