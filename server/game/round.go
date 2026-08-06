package game

import "strconv"

type RoundType string

// reason: we use snake case for enumerated constants
//
//goland:noinspection GoSnakeCaseUsage
const (
	RoundTypeUnset     RoundType = "unset"
	RoundTypeSetup     RoundType = "setup"
	RoundTypePreFlop   RoundType = "pre_flop"
	RoundTypeFlop      RoundType = "flop"
	RoundTypeTurn      RoundType = "turn"
	RoundTypeRiver     RoundType = "river"
	RoundTypeCompleted RoundType = "completed"
)

type round struct {
	bet                   int
	currentPlayerPosition int
	currentAggressor      int
	currentRoundType      RoundType
}

func (s stack) AsDto() ChipStackDTO {
	dto := make(ChipStackDTO)
	for d, c := range s {
		dto[strconv.Itoa(d)] = c
	}

	return dto
}

func (r round) AsDto() RoundDTO {
	return RoundDTO{
		Bet:                   r.bet,
		CurrentPlayerPosition: r.currentPlayerPosition,
		CurrentAggressor:      r.currentAggressor,
		CurrentRoundType:      r.currentRoundType,
	}
}
