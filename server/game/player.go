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
	PlayerIntentUnset PlayerIntent = "unset"
	PlayerIntentAnte  PlayerIntent = "ante"
	PlayerIntentCheck PlayerIntent = "check"
	PlayerIntentCall  PlayerIntent = "call"
	PlayerIntentRaise PlayerIntent = "raise"
	PlayerIntentAllIn PlayerIntent = "all_in"
	PlayerIntentFold  PlayerIntent = "fold"
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

// GetPlayerWithID gets a player from a game by id
func (g *game) GetPlayerWithID(playerID string) *player {
	for _, p := range g.table.players {
		if p.UserID == playerID {
			return p
		}
	}

	return nil
}

// NextPlayer gets the next at the table from a given positional offset
// this _only_ excludes playes whose [player.state] is [PlayerStateOut]
func (g *game) NextPlayer(offset int) *player {
	t := g.table
	for i := range len(t.players) {
		idx := (offset + i + 1) % len(t.players)
		p := t.players[idx]
		if p.state != PlayerStateOut {
			return p
		}
	}

	return nil
}

// returns the player after offset in turn order
// whos turn it would be if offset just ended
func (g *game) NextInactivePlayer(offset int) *player {
	t := g.table
	for i := range len(t.players) {
		idx := (offset + i + 1) % len(t.players)
		p := t.players[idx]
		if p.state == PlayerStateInactive {
			return p
		}
	}

	just.Logger.Debug("no inactive players")
	return nil
}
