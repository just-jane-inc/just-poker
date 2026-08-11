package game

import (
	"fmt"
	"strconv"

	"github.com/just-jane-inc/just-poker/server/just"
)

func (s stack) AsDto() ChipStackDTO {
	dto := make(ChipStackDTO)
	for d, c := range s {
		dto[strconv.Itoa(d)] = c
	}

	return dto
}

func (s ChipStackDTO) validate() *just.PokerError {
	for d, c := range s {
		denomination, err := strconv.Atoi(d)
		if err != nil {
			return just.NewPokerError(d+" in stack does not parse as an integer", just.InvalidChipDenomination)
		}

		if denomination <= 0 {
			return just.NewPokerError(d+" provided denomination is invalid", just.InvalidChipDenomination)
		}

		if c < 0 {
			return just.NewPokerError("provided a chip count less then 0", just.InvalidChipCount)
		}
	}

	return nil
}

func (s ChipStackDTO) asStack() stack {
	createStack := make(map[int]int)
	for d, c := range s {
		denomination, err := strconv.Atoi(d)
		if err != nil {
			continue
		}

		createStack[denomination] += c
	}

	return createStack
}

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

// mergeWith returns a stack that is the combination of
// two stacks, it does not modify either stack
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

func (g *game) ExchangeChips(playerID string, exchange ChipExchangeDTO) error {
	if err := exchange.Give.validate(); err != nil {
		return err
	}

	if err := exchange.Receive.validate(); err != nil {
		return err
	}

	for d := range exchange.Give {
		denomination, _ := strconv.Atoi(d)
		_, ok := g.denominations[denomination]
		if !ok {
			return just.NewPokerError(fmt.Sprintf("%s is not available for exchange", d), just.InvalidChipDenomination)
		}
	}

	for d := range exchange.Receive {
		denomination, _ := strconv.Atoi(d)
		_, ok := g.denominations[denomination]
		if !ok {
			return just.NewPokerError(fmt.Sprintf("%s is not available for exchange", d), just.InvalidChipDenomination)
		}
	}

	if exchange.Give.Sum() != exchange.Receive.Sum() {
		return &just.PokerError{
			Message: "Give and Recieve sum to different values, exchange is invalid",
			Code:    just.InvalidChipExchange,
		}
	}

	var p *player
	if p = g.table.GetPlayerWithID(playerID); p == nil {
		return &just.PokerError{
			Message: fmt.Sprintf("player [%s] not found at table", playerID),
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
		denomination, err := strconv.Atoi(d)
		if err != nil {
			continue
		}

		p.chips[denomination] -= c
	}

	for d, c := range exchange.Receive {
		denomination, err := strconv.Atoi(d)
		if err != nil {
			continue
		}

		p.chips[denomination] += c
	}

	return nil
}
