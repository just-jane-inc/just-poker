package game

import (
	"github.com/just-jane-inc/just-poker/server/just"
)

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

// sendMessageToConnections sends provided message to all subscribers on a games listener
func (g *game) sendMessageToConnections(eventType string, data any) {
	msg := just.WebsocketMessage[any]{
		Data:      data,
		EventType: eventType,
	}

	for _, conn := range just.UpdateHub.GetChannelsForGame(g.id) {
		select {
		case conn.MessageChannel <- msg:
		default:
			just.Logger.Warnf("buffer for connection [%s]::[%s] full, exiting", conn.GameID, conn.PlayerID)
			conn.Exit <- struct{}{}
		}
	}
}
