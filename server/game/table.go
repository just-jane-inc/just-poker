package game

import (
	"errors"
	"math"
	"slices"

	"github.com/just-jane-inc/just-poker/server/just"
)

var (
	cardRanks = [13]rune{'A', '2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K'}
	cardSuits = [4]rune{'s', 'd', 'c', 'h'}
)

type table struct {
	currentHand        HandDTO
	currentTurn        TurnDTO
	currentRound       round
	players            []*player
	pot                stack
	street             []*card
	deck               *deck
	buttonPosition     int
	bigBlindPosition   int
	smallBlindPosition int
	gameID             string
	currentTurnChannel chan just.WebsocketMessage[any]
	isHeadsUp          bool
}

func (t table) GetPlayerWithID(playerID string) *player {
	for _, p := range t.players {
		if p.UserID == playerID {
			return p
		}
	}

	return nil
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

func (dto TableDTO) AsTable() *table {
	t := table{
		players:            make([]*player, len(dto.Players)),
		pot:                dto.Pot.asStack(),
		street:             make([]*card, len(dto.Street)),
		currentRound:       dto.CurrentRound.AsRound(),
		currentHand:        dto.CurrentHand,
		buttonPosition:     dto.ButtonPosition,
		smallBlindPosition: dto.SmallBlindPosition,
		bigBlindPosition:   dto.BigBlindPosition,
	}

	for i, p := range dto.Players {
		t.players[i] = &player{
			UserID:          p.UserID,
			DisplayName:     p.DisplayName,
			UserType:        p.UserType,
			state:           PlayerState(p.State),
			position:        p.Position,
			pocket:          make([]*card, len(p.Hole)),
			chips:           p.Stack.asStack(),
			currentBet:      p.CurrentBet.asStack(),
			potContribution: p.PotContribution,
		}

		for j, c := range p.Hole {
			t.players[i].pocket[j] = &card{c.Rank, c.Suit}
		}
	}

	for k, c := range dto.Street {
		t.street[k] = &card{c.Rank, c.Suit}
	}

	return &t
}

func (t table) AsDTO() TableDTO {
	dto := TableDTO{
		Players:            make([]PlayerDTO, len(t.players)),
		Pot:                t.pot.AsDto(),
		Street:             make([]CardDTO, len(t.street)),
		CurrentRound:       t.currentRound.AsDto(),
		CurrentHand:        t.currentHand,
		ButtonPosition:     t.buttonPosition,
		SmallBlindPosition: t.smallBlindPosition,
		BigBlindPosition:   t.bigBlindPosition,
	}

	for i, player := range t.players {
		dto.Players[i] = player.AsDTO()
	}

	for i, card := range t.street {
		if card != nil {
			dto.Street[i] = CardDTO{Rank: card.rank, Suit: card.suit}
		}
	}

	return dto
}

func (t *table) NextPlayer(offset int) *player {
	for i := range len(t.players) {
		idx := (offset + i + 1) % len(t.players)
		p := t.players[idx]
		if p.state != PlayerStateOut {
			return p
		}
	}

	return nil
}

// returns the player after offset in turn order
// whos turn it would be if offset just ended
func (t *table) NextInactivePlayer(offset int) *player {
	for i := range len(t.players) {
		idx := (offset + i + 1) % len(t.players)
		p := t.players[idx]
		if p.state == PlayerStateInactive {
			return p
		}
	}

	return nil
}

func (t *table) nextRound() {
	// at the begining of a round all active players
	// need to be set to inactive?
	for _, p := range t.players {
		if p.state == PlayerStateActive {
			p.state = PlayerStateInactive
		}

		if t.currentRound.currentRoundType != RoundTypeAnte {
			p.currentBet = make(stack)
		}
	}

	switch t.currentRound.currentRoundType {
	case RoundTypeUnset: // GOTO setup
		t.currentRound.currentRoundType = RoundTypeAnte
		t.currentRound.currentAggressor = t.NextInactivePlayer(t.bigBlindPosition).position
		p := t.players[t.smallBlindPosition]
		t.currentRound.currentPlayerPosition = p.position
		t.currentRound.bet = t.currentHand.SmallBlind

		// this must come after setting the current aggressor to deal with heads up play
		p.state = PlayerStateActive

	case RoundTypeAnte: // GOTO pre-flop
		t.currentRound.currentRoundType = RoundTypePreFlop

		// TODO: we really need to stop depending on inactive state
		// at all times, need a different method
		offset := t.NextInactivePlayer(t.buttonPosition).position
		for idx := range len(t.players) {
			p := t.players[(idx+offset)%len(t.players)]

			if p.state == PlayerStateOut {
				continue
			}

			p.pocket = make([]*card, 2)
			p.pocket[0] = t.deck.Draw()
		}
		for idx := range len(t.players) {
			p := t.players[(idx+offset)%len(t.players)]

			if p.state == PlayerStateOut {
				continue
			}

			p.pocket[1] = t.deck.Draw()
		}

	case RoundTypePreFlop: // GOTO flop
		for _, p := range t.players {
			p.currentBet = make(stack)
		}

		t.currentRound.currentRoundType = RoundTypeFlop
		t.currentRound.bet = 0

		t.deck.Burn()
		t.street = append(t.street, t.deck.Draw())
		t.street = append(t.street, t.deck.Draw())
		t.street = append(t.street, t.deck.Draw())

	case RoundTypeFlop: // GOTO turn
		t.currentRound.currentRoundType = RoundTypeTurn
		t.currentRound.bet = 0

		t.deck.Burn()
		t.street = append(t.street, t.deck.Draw())

	case RoundTypeTurn: // GOTO river
		t.currentRound.currentRoundType = RoundTypeRiver
		t.currentRound.bet = 0

		t.deck.Burn()
		t.street = append(t.street, t.deck.Draw())

	case RoundTypeRiver: // GOTO end
		// He beat me... Straight up... Pay him... Pay that man his money
		// Captain_Onosa
		handEvaluations := t.Showdown()

		bestHand := math.MaxInt
		winners := make([]int, 0)
		for position, eval := range handEvaluations {
			if bestHand > eval {
				winners = make([]int, 0)
				bestHand = eval
			}

			if bestHand == eval {
				winners = append(winners, position)
			}
		}

		if len(winners) == 0 {
			just.Logger.Errorf("critical error, hand terminated with no winners")
			t.currentRound.currentRoundType = RoundTypeCompleted // blow up?
			return
		}

		p := t.players[winners[0]]
		for denomination, count := range t.pot {
			p.chips[denomination] += count
		}

		msg := PayoutEventDTO{
			PlayerID: p.UserID,
			Chips:    t.pot.AsDto(),
		}

		t.sendMessageToConnections("payout", msg)
		t.pot = make(map[int]int)
		just.Logger.Debugf("the winner maybe is: [%v]", winners)
		t.currentRound.currentRoundType = RoundTypeCompleted // blow up?
	}

	if t.currentRound.currentRoundType == RoundTypeUnset {
		return
	}

	msg := RoundStartEventDTO{
		Type: t.currentRound.currentRoundType,
	}

	t.sendMessageToConnections("round_start", msg)
}

func (t *table) sendMessageToConnections(eventType string, data any) {
	msg := just.WebsocketMessage[any]{
		Data:      data,
		EventType: eventType,
	}

	for _, conn := range just.UpdateHub.GetChannelsForGame(t.gameID) {
		select {
		case conn.MessageChannel <- msg:
		default:
			just.Logger.Infof("buffer for connection [%s]::[%s] full, exiting", conn.GameID, conn.PlayerID)
			conn.Exit <- struct{}{}
		}
	}
}

func (t *table) Showdown() map[int]int {
	remainingPlayers := make([]*player, 0)
	for _, p := range t.players {
		if p.state == PlayerStateFolded {
			continue
		}

		if p.state == PlayerStateOut {
			continue
		}

		if p.state == PlayerStateUnset {
			continue
		}

		remainingPlayers = append(remainingPlayers, p)
	}

	if len(remainingPlayers) == 0 {
		just.Logger.Errorf("no one wins? how? everyone folded? like everyong? what?")
		return make(map[int]int)
	}

	handEvaluations := make(map[int]int)
	for _, p := range remainingPlayers {
		model := t.GetHand(p.position).Cards
		cards := make([]just.Card, len(model))
		for i, c := range model {
			cards[i] = just.Card{Rank: c.rank, Suit: c.suit}
		}

		eval, err := just.GetHandScore(cards)
		if err != nil {
			just.Logger.Errorf("encountered error evaluating hand %v", err)
			continue
		}

		handEvaluations[p.position] = eval
	}

	return handEvaluations
}

func (t *table) GetHand(position int) Hand {
	hand := slices.Concat(t.street, t.players[position].pocket)
	return Hand{Cards: hand}
}

func (t *table) OnGameOver() {
	_, err := CurrentGames.RemoveGame(t.gameID)
	if err != nil {
		just.Logger.Errorf("encountered error removing game in OnGameOver: %v", err)
	}

	for _, p := range t.players {
		if p.state != PlayerStateOut {
			p.state = PlayerStateWon
		}
	}

	for _, conn := range just.UpdateHub.GetChannelsForGame(t.gameID) {
		conn.MessageChannel <- just.WebsocketMessage[any]{
			EventType: "game_over",
			Data:      t.AsDTO().Players,
		}

		close(conn.MessageChannel)
	}
}

func (t *table) nextHand(bb int, sb int) error {
	just.Logger.Debugf("attempting to start hand [%d]", t.currentHand.ID+1)

	pot := t.pot.Sum()
	if pot > 0 {
		return errors.New("pot still has chips, cannot start new hand until the previous one is resolved")
	}

	for _, p := range t.players {
		p.pocket = make([]*card, 0)

		if p.chips.Sum() == 0 {
			p.state = PlayerStateOut
		}

		// if a player is out we do not want to include them in the hand,
		// this state should be locked in place by everything which updates
		// player state.
		if p.state == PlayerStateOut {
			continue
		}

		p.state = PlayerStateInactive
	}

	var playersRemaining int
	for _, p := range t.players {
		if p.state != PlayerStateOut {
			playersRemaining += 1
		}
	}

	if playersRemaining == 1 {
		t.OnGameOver()
		just.Logger.Infof("game [%s] completed", t.gameID)
		return nil
	}

	t.isHeadsUp = playersRemaining == 2

	t.currentHand.SmallBlind = sb
	t.currentHand.BigBlind = bb
	t.buttonPosition = t.NextInactivePlayer(t.buttonPosition).position

	if t.isHeadsUp {
		// in heads up play the button posts the small blind
		t.smallBlindPosition = t.buttonPosition
	} else {
		t.smallBlindPosition = t.NextInactivePlayer(t.buttonPosition).position
	}

	t.bigBlindPosition = t.NextInactivePlayer(t.smallBlindPosition).position

	t.deck.Reset()
	t.street = make([]*card, 0)
	t.pot = make(map[int]int)
	t.currentRound.currentRoundType = RoundTypeUnset

	t.currentHand.ID += 1
	t.currentTurn.ID = 0

	msg := HandStartEventDTO{
		ID:                 t.currentHand.ID,
		BigBlindCost:       t.currentHand.BigBlind,
		BigBlindPosition:   t.bigBlindPosition,
		SmallBlindCost:     t.currentHand.SmallBlind,
		SmallBlindPosition: t.smallBlindPosition,
		ButtonPosition:     t.buttonPosition,
	}

	t.sendMessageToConnections("hand_started", msg)
	t.nextRound()
	return nil
}
