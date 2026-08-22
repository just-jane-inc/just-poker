package game

type HandStartEventDTO struct {
	ID                 int `json:"id"`
	BigBlindCost       int `json:"big_blind_cost"`
	BigBlindPosition   int `json:"big_blind_position"`
	SmallBlindCost     int `json:"small_blind_cost"`
	SmallBlindPosition int `json:"small_blind_position"`
	ButtonPosition     int `json:"button_position"`
}

type RoundStartEventDTO struct {
	Type RoundType `json:"round_type"`
}

type PayoutEventDTO struct {
	PlayerID string       `json:"player_id"`
	Chips    ChipStackDTO `json:"chips"`
}

type TurnStartEventDTO struct {
	PlayerID  string    `json:"player_id"`
	RoundType RoundType `json:"round_type"`
	BetAmount int       `json:"bet_amount"`
}
