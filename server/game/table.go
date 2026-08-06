package game

import (
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
		if p.state != player_state_out {
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
		if p.state == player_state_inactive {
			return p
		}
	}

	return nil
}

func (t *table) NextRound() {
	// at the begining of a round all active players
	// need to be set to inactive?
	for _, p := range t.players {
		if p.state == player_state_active {
			p.state = player_state_inactive
		}

		if t.currentRound.currentRoundType != round_type_setup {
			p.currentBet = make(stack)
		}
	}

	switch t.currentRound.currentRoundType {
	case round_type_unset: // GOTO setup
		t.currentRound.currentRoundType = round_type_setup
		// this is the first time we enter the next round thing, this is fine?
		// we might handle this in new haand?
		p := t.NextInactivePlayer(t.buttonPosition)
		p.state = player_state_active
		t.currentRound.currentPlayerPosition = p.position
		t.currentRound.bet = t.currentHand.SmallBlind

	case round_type_setup: // GOTO pre-flop
		t.currentRound.currentRoundType = round_type_pre_flop

		// TODO: we really need to stop depending on inactive state
		// at all times, need a different method
		offset := t.NextInactivePlayer(t.buttonPosition).position
		for idx := range len(t.players) {
			p := t.players[(idx+offset)%len(t.players)]

			if p.state == player_state_out {
				continue
			}

			p.pocket = make([]*card, 2)
			p.pocket[0] = t.deck.Draw()
		}
		for idx := range len(t.players) {
			p := t.players[(idx+offset)%len(t.players)]

			if p.state == player_state_out {
				continue
			}

			p.pocket[1] = t.deck.Draw()
		}

	case round_type_pre_flop: // GOTO flop
		for _, p := range t.players {
			p.currentBet = make(stack)
		}

		t.currentRound.currentRoundType = round_type_flop
		t.currentRound.bet = 0

		t.deck.Burn()
		t.street = append(t.street, t.deck.Draw())
		t.street = append(t.street, t.deck.Draw())
		t.street = append(t.street, t.deck.Draw())

	case round_type_flop: // GOTO turn
		t.currentRound.currentRoundType = round_type_turn
		t.currentRound.bet = 0

		t.deck.Burn()
		t.street = append(t.street, t.deck.Draw())

	case round_type_turn: // GOTO river
		t.currentRound.currentRoundType = round_type_river
		t.currentRound.bet = 0

		t.deck.Burn()
		t.street = append(t.street, t.deck.Draw())

	case round_type_river: // GOTO end
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
			t.currentRound.currentRoundType = round_type_completed // blow up?
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

		sendMessageToConnections(t.gameID, "payout", msg)
		t.pot = make(map[int]int)
		just.Logger.Debugf("the winner maybe is: [%v]", winners)
		t.currentRound.currentRoundType = round_type_completed // blow up?
	}

	if t.currentRound.currentRoundType == round_type_unset {
		return
	}

	msg := RoundStartEventDTO{
		Type: t.currentRound.currentRoundType,
	}

	sendMessageToConnections(t.gameID, "round_start", msg)
}

func (t *table) Showdown() map[int]int {
	remainingPlayers := make([]*player, 0)
	for _, p := range t.players {
		if p.state == player_state_folded {
			continue
		}

		if p.state == player_state_out {
			continue
		}

		if p.state == player_state_unset {
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
