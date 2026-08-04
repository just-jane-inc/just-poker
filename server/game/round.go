package game

import "strconv"

type RoundType string

// reason: we use snake case for enumerated constants
//
//goland:noinspection GoSnakeCaseUsage
const (
	round_type_unset     RoundType = "unset"
	round_type_setup     RoundType = "setup"
	round_type_pre_flop  RoundType = "pre_flop"
	round_type_flop      RoundType = "flop"
	round_type_turn      RoundType = "turn"
	round_type_river     RoundType = "river"
	round_type_completed RoundType = "completed"
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
