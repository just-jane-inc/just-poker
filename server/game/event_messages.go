package game

import (
	"time"

	"github.com/just-jane-inc/just-poker/server/just"
)

const (
	GameStarting      EventType = "game_starting"
	HandStarted       EventType = "hand_started"
	RoundStarted      EventType = "round_started"
	HandPayouts       EventType = "hand_payouts"
	TurnStarted       EventType = "turn_started"
	PlayerAction      EventType = "player_action"
	GameStatusChanged EventType = "game_status_changed"
	GameEnding        EventType = "game_ending"
	ChipExchange      EventType = "chip_exchange"
)

type EventType string

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

func (g *game) OnChipExchange(exchange ChipExchangeDTO) {
	sendMessageToConnections(g.id, ChipExchange, exchange)
}

func (g *game) OnPayout(payouts []PayoutEventDTO) {
	sendMessageToConnections(g.id, HandPayouts, payouts)
}

func (g *game) OnHandStarted() {
	t := g.table
	msg := HandStartEventDTO{
		ID:                 t.currentHand.ID,
		BigBlindCost:       t.currentHand.BigBlind,
		BigBlindPosition:   t.bigBlindPosition,
		SmallBlindCost:     t.currentHand.SmallBlind,
		SmallBlindPosition: t.smallBlindPosition,
		ButtonPosition:     t.buttonPosition,
	}

	sendMessageToConnections(g.id, HandStarted, msg)
}

func (g *game) OnRoundStarted() {
	msg := RoundStartEventDTO{
		Type: g.table.currentRound.currentRoundType,
	}

	sendMessageToConnections(g.id, RoundStarted, msg)
}

func (g *game) OnTurnStarted() {
	data := TurnStartEventDTO{
		PlayerID:  g.table.currentTurn.PlayerID,
		RoundType: g.table.currentRound.currentRoundType,
		BetAmount: g.table.currentRound.bet,
	}

	sendMessageToConnections(g.id, TurnStarted, data)
}

func (g *game) OnPlayerAction(action PlayerActionDTO) {
	sendMessageToConnections(g.id, PlayerAction, action)
}

func (g *game) OnGameStarting() {
	sendMessageToConnections(g.id, GameStarting, g.AsDTO())
	g.OnGameStatusChanged()
}

func (g *game) OnGameStatusChanged() {
	dto := g.AsDTO()
	for _, conn := range just.UpdateHub.GetChannelsForGame(dto.ID) {
		just.Logger.Debugf("sending update to player %s", conn.PlayerID)
		dtoNew := dto.DeepCopy()

		if conn.UserType != just.UserTypeAdmin && conn.UserType != just.UserTypeGameMaster {
			dtoNew.MaskCards(conn.PlayerID)
		}

		msg := just.WebsocketMessage[any]{
			Data:      dtoNew,
			EventType: string(GameStatusChanged),
		}

		select {
		case conn.MessageChannel <- msg:
		default:
		}
	}
}

func (g *game) OnGameEnded() {
	just.Logger.Infof("game has ended sending game_over and closing connections game=[%s]", g.id)
	t := time.Now()
	g.endedAt = &t
	dto := g.AsDTO()
	for _, conn := range just.UpdateHub.GetChannelsForGame(g.id) {
		conn.MessageChannel <- just.WebsocketMessage[any]{
			EventType: string(GameEnding),
			Data:      dto.Table.Players,
		}

		conn.SignalExit("game ended")
	}

	delete(CurrentGames.games, g.id)

	if err := just.RecordingHub.OnGameUpdate(dto); err != nil {
		just.Logger.Errorf("error updating game state in elastic: %v", err)
	}
}

// sendMessageToConnections sends provided message to all subscribers on a games listener
func sendMessageToConnections(gameID string, eventType EventType, data any) {
	just.Logger.Debugf("sending [%s] update for game with ID [%s]", eventType, gameID)
	msg := just.WebsocketMessage[any]{
		Data:      data,
		EventType: string(eventType),
	}

	for _, conn := range just.UpdateHub.GetChannelsForGame(gameID) {
		select {
		case conn.MessageChannel <- msg:
		default:
			just.Logger.Warnf("buffer for connection [%s]::[%s] full, exiting", conn.GameID, conn.PlayerID)
			conn.Exit <- struct{}{}
		}
	}
}
