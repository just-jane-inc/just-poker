package game

import (
	"strconv"
	"time"

	"github.com/just-jane-inc/just-poker/server/just"
)

// ActiveGameDTO encodes an active game
//
// TODO: what is this for?
type ActiveGameDTO struct {
	ID      string   `json:"id"`
	Players []string `json:"player_ids"`
}

// HandEvaluationDTO holds the result of a hand evalution
type HandEvaluationDTO struct {
	Error      string `json:"error"`
	Evaluation int    `json:"evaluation"`
}

// NewHandDTO a dto describing data required to start a new hand
type NewHandDTO struct {
	// the deck, optionally provided to determine trhe order of cards
	// if not provided the cards will be ordered randomly by the server
	Deck []CardDTO `json:"deck"`
}

// ChipExchangeDTO a dto describing an exchange request between
// two ChipStackDTO.
type ChipExchangeDTO struct {
	// the stack to give during the exchange
	UserID string `json:"user_id,omitempty"`

	// the stack to give during the exchange
	Give ChipStackDTO `json:"give"`

	// the stack to receive as a result of the exchange
	Receive ChipStackDTO `json:"receive"`
}

// ChipCountDTO a dto describing the denomation and count
// of a single collection of chips
type ChipCountDTO struct {
	// the denomination for the chips
	Denomination int `json:"denomination"`
	// the number of chips in the collection
	Count int `json:"count"`
}

// ChipStackDTO a dto that aliases a map[string]int mapping string
// denominations onto integer counts
type ChipStackDTO map[string]int

// RoundDTO a dto describing information about the current round of play
type RoundDTO struct {
	// the amount of chips required to call (how does this work with split pots?)
	Bet int `json:"bet"`

	// the index into the play array for the player whose turn it currently is
	CurrentPlayerPosition int `json:"current_player_position"`

	// the player who currently ends the round? the last raise?
	CurrentAggressor int `json:"current_aggressor"`

	// the type of round
	CurrentRoundType RoundType `json:"current_round_type,omitempty"`
}

// PlayerDTO a dto representing a single player
type PlayerDTO struct {
	// the id of the user
	UserID string `json:"user_id"`

	// the users display name
	DisplayName string `json:"display_name"`

	// the type of user
	UserType just.UserType `json:"user_type"`

	// the players position at the table, starting with 0 being the
	// first player sitting clockwise from the dealer
	Position int `json:"position"`

	// the cards current held by this player - only visible for authorized users
	// during a game.
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

// PlayerActionDTO a dto describing the action taken by a player
type PlayerActionDTO struct {
	// the id of the player preforming the action
	PlayerID string `json:"player_id"`

	// the type of action a player intended to preform
	Intent PlayerIntent `json:"intent"`

	// an optional mapping of chips that is required by some action types.
	Bet ChipStackDTO `json:"chips"`

	// a timestamp capturing when a succesful action was accepted by the game
	AcceptedAt *time.Time `json:"accepted_at" extensions:"x-nullable"`
}

// GameDTO a dto describing the current state of an entire game
type GameDTO struct {
	// the id of the game
	ID string `json:"id"`

	// the time that the game started originally
	StartedAt *time.Time `json:"started_at" extensions:"x-nullable"`

	// the time that the game ended
	EndedAt *time.Time `json:"ended_at" extensions:"x-nullable"`

	// the configuration used to setup the game
	Config NewGameConfigDTO `json:"game_config"`

	// the table
	Table TableDTO `json:"table"`
}

// NewGameConfigDTO a dto supplying configuring information
// for creating a new game.
type NewGameConfigDTO struct {
	// the number of players (max) the game supports
	PlayerCount int `json:"player_count"`

	// the starting chips to give each player
	StartingChips ChipStackDTO `json:"starting_chips"`

	// the big blind
	BigBlind int `json:"big_blind"`

	// the small blind
	SmallBlind int `json:"small_blind"`

	// a flag which indicates true if the game server should wait
	// for a signal to start hands or if it should do so automatically
	AutoStartHands bool `json:"auto_starts_hands"`

	// a collection of denominations that are available for chips at the table
	ChipDenominations []int `json:"chip_denominations"`

	// the number of milliseconds that a bot has to take a turn
	BotTurnTimeout int `json:"bot_turn_timeout"`
}

// TableDTO a dto describing the full state of a table
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

// HandDTO a dto containing meta-data about a hand
type HandDTO struct {
	// the (non decreasing) hand counter
	ID int `json:"id"`

	// the amount of chips for the big blind in this hand
	BigBlind int `json:"big_blind"`

	// the amount of chips for the small blind in this hand
	SmallBlind int `json:"small_blind"`

	// the time that this hand started
	StartedAt time.Time `json:"started_at"`
}

// TurnDTO a dto encoding meta data related to a specific turn of a hand
type TurnDTO struct {
	// the id of a turn
	ID int `json:"id"`

	// the time that the turn started
	StartedAt time.Time `json:"started_at"`

	// the id of the player whose turn it is
	PlayerID string `json:"player_id"`
}

// CardDTO a dto encoding an individual card
type CardDTO struct {
	// the rank of a card as a rune - int32 ASCII encoding
	Rank rune `json:"rank"`

	// the suit of a card as a rune - int32 ASCII encoding
	Suit rune `json:"suit"`
}

// Sum gets the integer sum of all chips in a stack
func (s ChipStackDTO) Sum() int {
	total := 0
	for d, count := range s {
		denomination, err := strconv.Atoi(d)
		if err != nil {
			continue
		}

		total += denomination * count
	}

	return total
}

// AsDTO converts a player model to its [PlayerDTO] representation
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
