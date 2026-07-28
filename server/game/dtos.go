package game

import (
	"strconv"
	"time"
)

// a collection of chips with the same denomination
type ChipCountDTO struct {
	// the denomination for the chips
	Denomination int `json:"denomination"`
	// the number of chips in the collection
	Count int `json:"count"`
}

// a unit of mixed denomination chips that are associated to a shared entity
type ChipStackDTO map[string]int

// data related to a betting round
// @Description Detailed information about a registered user.
// @name UserAccount
type RoundDTO struct {
	// the amount of chips required to call (how does this work with split pots?)
	Bet int `json:"bet"`

	// the index into the play array for the player whose turn it currently is
	CurrentPlayerPosition int `json:"current_player_position"`

	// the player who currently ends the round? the last raise?
	CurrentAggressor int `json:"current_agressor"`

	// the type of round
	CurrentRoundType RoundType `json:"current_round_type"`
}

// a player (a user that is in a game)
// @Description Detailed information about a registered user.
// @name UserAccount
type PlayerDTO struct {
	// the is of the user
	UserID string `json:"user_id"`

	// the users display name
	DisplayName string `json:"display_name"`

	// the type of user TODO: make this an enum
	UserType string `json:"user_type"`

	// the players position at the table, starting with 0 being the
	// first player sitting clockwise from the dealer
	Position int `json:"position"`

	// the cards current held by this player. the cards are only visible
	// to authorized users. TODO: make the prev statement true, also make an anon card
	Hole []CardDTO `json:"hole"`

	// a map of the chips held by the player where the keys represent
	// denominations and the values are a count
	Stack ChipStackDTO `json:"stack"`

	// the stack of chips player currently has put forward in this round
	CurrentBet ChipStackDTO `json:"current_bet"`

	// the sum total the player has contributed to the pot, note that this
	// does not include chips currently in CurrentBet
	PotContribution int `json:"pot_contribution"`

	// the players state
	State string `json:"state"`
}

// the action that a player has taken
// @Description The action that a player has taken...
// @name PlayerAction
type PlayerActionDTO struct {
	// the id of the player preforming the action
	PlayerID string `json:"player_id"`

	// the type of action a player intended to preform
	Intent string `json:"intent"`

	// an optional mapping of chips that is required by some action types.
	Bet ChipStackDTO `json:"chips"`

	// a timestamp capturing when a succesful action was accepted by the game
	AcceptedAt *time.Time `json:"accepted_at"`
}

// the game
// @Description yeah, the game
// @name Game
type GameDTO struct {
	// the id of the game
	ID string `json:"id"`

	// the time that the game started originally
	StartedAt *time.Time `json:"started_at"`

	// the configuration used to setup the game
	Config NewGameConfigDTO `json:"game_config"`

	// the table
	Table TableDTO `json:"table"`
}

// the games configuration
// @Description yeah, the game
// @name Game
type NewGameConfigDTO struct {
	// the number of players (max) the game supports
	PlayerCount int `json:"player_count"`

	// the starting chips to give each player
	StartingChips ChipStackDTO `json:"starting_chips"`

	// the big blind
	BigBlind int `json:"big_blind"`

	// the small blind
	SmallBlind int `json:"small_blind"`
}

// the table
// @Description yeah, the table
// @name Table
type TableDTO struct {
	// An array of players at the table
	Players []PlayerDTO `json:"players"`

	// the chips in the pot
	Pot ChipStackDTO `json:"pot"`

	// the cards that are on the street (community cards)
	Street []CardDTO `json:"street"`

	// the current round
	CurrentRound RoundDTO `json:"current_round"`

	// the current hand
	CurrentHand HandDTO `json:"current_hand"`

	// the position of the button
	ButtonPosition int `json:"button_position"`

	// the position of the small blind
	SmallBlindPosition int `json:"small_blind_position"`

	// the position of the big blind
	BigBlindPosition int `json:"big_blind_position"`
}

// a hand
// @Description yeah, the game
// @name Hand
type HandDTO struct {
	// the id of a hand
	ID string `json:"id"`

	// the (non decreasing) hand counter
	Count int `json:"count"`

	// the index of the position of the player who has the button
	Button int `json:"button"`

	// the amount of chips for the big blind in this hand
	BigBlind int `json:"big_blind"`

	// the amount of chips for the small blind in this hand
	SmallBlind int `json:"small_blind"`

	// the time that this hand started
	StartedAt time.Time `json:"started_at"`
}

// an individual turn
// @Description a turn
// @name Turn
type TurnDTO struct {
	// the id of the turn
	ID string `json:"id"`

	// the (non decreasing) turn counter
	Count int `json:"count"`

	// the time that the turn started
	StartedAt time.Time `json:"started_at"`

	// the id of the player whose turn it is
	PlayerID string `json:"player_id"`

	// the index of the position of the player whose turn it is
	PlayerPosition int `json:"player_position"`
}

// an individual card
// @Description a card
// @name Card
type CardDTO struct {
	// the rank of a card as a rune - int32 ASCII encoding
	Rank rune `json:"rank"`

	// the suit of a card as a rune - int32 ASCII encoding
	Suit rune `json:"suit"`
}

func (c ChipStackDTO) Sum() int {
	total := 0
	for d, count := range c {
		denomination, err := strconv.Atoi(d)
		if err != nil {
			continue
		}

		total += denomination * count
	}

	return total
}

func (p player) AsDTO() PlayerDTO {
	dto := PlayerDTO{
		UserID:          p.UserID,
		DisplayName:     p.DisplayName,
		UserType:        p.UserType,
		Position:        p.position,
		Hole:            make([]CardDTO, 0),
		Stack:           p.chips.AsDto(),
		CurrentBet:      p.currentBet.AsDto(),
		PotContribution: p.potContribution,
		State:           string(p.state),
	}

	for _, card := range p.pocket {
		dto.Hole = append(dto.Hole, card.AsDTO())
	}

	return dto
}
