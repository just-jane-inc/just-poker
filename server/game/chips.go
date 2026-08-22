package game

import (
	"fmt"

	"github.com/just-jane-inc/just-poker/server/just"
)

// AsDto converts the stack model to the DTO representation
func (s stack) AsDto() ChipStackDTO {
	dto := make(ChipStackDTO)
	for d, c := range s {
		dto[d] = c
	}

	return dto
}

// validate ensures that a stack of chips is valid
//
// - all denominations parse as an integer
// - all denominations are positive
// - all counts greater then or equal to zero
func (s ChipStackDTO) validate() *just.PokerError {
	for d, c := range s {

		if d <= 0 {
			return just.NewPokerError("provided a denomination less then 0", just.InvalidChipDenomination)
		}

		if c < 0 {
			return just.NewPokerError("provided a chip count less then 0", just.InvalidChipCount)
		}
	}

	return nil
}

// asStack converts a ChipStackDTO to the model representation of a stack of chips
func (s ChipStackDTO) asStack() stack {
	createStack := make(map[int]int)
	for d, c := range s {
		createStack[d] += c
	}

	return createStack
}

// Contains checks if a stack (that) is a subset of another stack (s)
func (s stack) Contains(that stack) bool {
	for d, c := range that {
		// get the number of chips with the provided denomination
		// available to this stack
		count, ok := s[d]
		if !ok {
			return false
		}

		// check if this stack has enough of the required denomination
		// to cover the check
		if count < c {
			return false
		}
	}

	return true
}

// mergeWith returns a stack that is the union of two stacks without modifying either
func (s stack) mergeWith(other stack) stack {
	result := make(stack)
	for d, c := range other {
		result[d] += c
	}

	for d, c := range s {
		result[d] += c
	}

	return result
}

// ExchangeChips preforms an exchange of chips for a specific player
//
// if the exchange is valid this will alter the player stack. in order for an exchange
// to be valid it must:
//
// - have a valid Give and Receive stack - see [ChipStackDTO.validate]
// - have all denominations in the Give and Receive stack exist in [game.config.denominations]
// - the Give and Receive stacks must have an equal sum
// - the player doing the exchange must have all chips in the Give stack
//
// TODO: this function creates a race condition with coverBet and should be managed.
// TODO: chip exchanges must be sent over the wire as an event
func (g *game) ExchangeChips(exchange ChipExchangeDTO) *just.PokerError {
	g.gameStateLock.Lock()
	defer g.gameStateLock.Unlock()

	if err := exchange.Give.validate(); err != nil {
		return err
	}

	if err := exchange.Receive.validate(); err != nil {
		return err
	}

	for d := range exchange.Give {
		_, ok := g.denominations[d]
		if !ok {
			return just.NewPokerError(fmt.Sprintf("%d is not available for exchange", d), just.InvalidChipDenomination)
		}
	}

	for d := range exchange.Receive {
		_, ok := g.denominations[d]
		if !ok {
			return just.NewPokerError(fmt.Sprintf("%d is not available for exchange", d), just.InvalidChipDenomination)
		}
	}

	if exchange.Give.Sum() != exchange.Receive.Sum() {
		return &just.PokerError{
			Message: "Give and Recieve sum to different values, exchange is invalid",
			Code:    just.InvalidChipExchange,
		}
	}

	var p *player
	if p = g.table.GetPlayerWithID(exchange.UserID); p == nil {
		return &just.PokerError{
			Message: fmt.Sprintf("player [%s] not found at table", exchange.UserID),
			Code:    just.UserNotFound,
		}
	}

	if !p.chips.Contains(exchange.Give.asStack()) {
		return &just.PokerError{
			Message: "player stack does not have enough chips to cover requested exchange",
			Code:    just.InvalidChipExchange,
		}
	}

	for d, c := range exchange.Give {
		p.chips[d] -= c
	}

	for d, c := range exchange.Receive {
		p.chips[d] += c
	}

	g.table.sendMessageToConnections("player_chip_exchange", exchange)
	return nil
}
