package game

type RoundType string

// reason: we use snake case for enumerated constants
//
//goland:noinspection GoSnakeCaseUsage
const (
	RoundTypeUnset     RoundType = "unset"
	RoundTypeAnte      RoundType = "setup"
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

func (dto RoundDTO) AsRound() round {
	// TODO: why is this not a DTO?
	return round{
		bet:                   dto.Bet,
		currentRoundType:      dto.CurrentRoundType,
		currentPlayerPosition: dto.CurrentPlayerPosition,
		currentAggressor:      dto.CurrentAggressor,
	}
}

func (r round) AsDto() RoundDTO {
	return RoundDTO{
		Bet:                   r.bet,
		CurrentPlayerPosition: r.currentPlayerPosition,
		CurrentAggressor:      r.currentAggressor,
		CurrentRoundType:      r.currentRoundType,
	}
}
