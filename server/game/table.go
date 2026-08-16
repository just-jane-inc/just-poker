package game

import (
	"errors"
	"fmt"
	"math"
	"slices"
	"sort"
	"time"

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
	denominations      []int
}

// AsTable converts a DTO table into the model for a table
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

func (t table) GetPlayerWithID(playerID string) *player {
	for _, p := range t.players {
		if p.UserID == playerID {
			return p
		}
	}

	return nil
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

	just.Logger.Debug("no inactive players")
	return nil
}

// nextRound handles progressing the round in a game
// unset    -> ante is the initial transition the ensures the round DTO is initialized correctly
// ante     -> pre-flop will deal cards to each player's hole
// pre-flop -> flop will deal the first street (flop)
// flop     -> turn will deal the second street (turn)
// turn     -> river will deal the final street (river)
// river    -> completed will determine winners and give chips to each player, from here a new hand is required
func (t *table) nextRound() {
	// ensure that all players are in a nominal state when we start the next round
	for _, p := range t.players {
		if p.state == PlayerStateActive {
			p.state = PlayerStateInactive
		}

		if t.currentRound.currentRoundType != RoundTypeAnte {
			p.potContribution += p.currentBet.Sum()
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
		handEvaluations := t.GetHandEvaluations()
		payouts := make(map[int]int)
		pots := t.getSplitPots()
		for _, p := range pots {
			just.Logger.Debugf("processing pot for payout: %v", p)
			payout, err := t.handlePayout(handEvaluations, p)
			if err != nil {
				just.Logger.Errorf("encountered error computing payout: %v", err)
				continue
			}

			for position, amount := range payout {
				payouts[position] += amount
			}
		}

		just.Logger.Debugf("payouts calculated: %v", payouts)

		payoutEvents := make([]PayoutEventDTO, 0)
		// we now need to construct from our available denominations a chipstack
		// equal to the required payout
		for position, payout := range payouts {
			currentPlayer := t.players[position]
			e := PayoutEventDTO{PlayerID: currentPlayer.UserID}
			chips := make(stack)
			for _, denomination := range t.denominations {
				if payout >= denomination {
					available, ok := t.pot[denomination]
					if !ok {
						continue
					}

					take := min(available, payout/denomination)
					t.pot[denomination] -= take
					chips[denomination] = take
					payout -= (take * denomination)
				}

				if payout == 0 {
					break
				}
			}

			e.Chips = chips.AsDto()
			payoutEvents = append(payoutEvents, e)
			currentPlayer.chips = currentPlayer.chips.mergeWith(chips)
		}

		if t.pot.Sum() > 0 {
			just.Logger.Errorf("pot was not exhausted during payout %v", t.pot)
		}

		t.sendMessageToConnections("payout", payoutEvents)
		t.pot = make(map[int]int)
		t.currentRound.currentRoundType = RoundTypeCompleted

		// oh also if the game is over we do stuff about it here??
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

// Showdown evaluates all hands at the table
// and returns a mapping of [position]=evaluation
// where the lower an evaluation is numerically the
// better it is as a poker hand.
func (t *table) GetHandEvaluations() map[int]int {
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

	if len(remainingPlayers) == 1 {
		handEvaluations[remainingPlayers[0].position] = 0
	}

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
	g, err := CurrentGames.RemoveGame(t.gameID)
	if err != nil {
		just.Logger.Errorf("encountered error removing game in OnGameOver: %v", err)
	} else {
		now := time.Now()
		g.endedAt = &now
	}

	for _, p := range t.players {
		if p.state != PlayerStateOut {
			p.state = PlayerStateWon
		}
	}
}

func (t *table) nextHandWithDeck(deck []CardDTO, bb int, sb int) *just.PokerError {
	if t.currentRound.currentRoundType != RoundTypeCompleted && t.currentRound.currentRoundType != RoundTypeUnset {
		return just.NewPokerError("a hand is already in progress - can not start a new hand", just.HandAlreadyInProgress)
	}

	err := t.nextHand(bb, sb)
	if err != nil {
		return just.NewPokerError(fmt.Sprintf("encountered error: %v", err), 2025)
	}

	if len(deck) != 52 {
		return nil
	}

	t.deck.cards = make([]card, 52)
	for i, c := range deck {
		t.deck.cards[i] = card{rank: c.Rank, suit: c.Suit}
	}

	return nil
}

func (t *table) nextHand(bb int, sb int) error {
	just.Logger.Debugf("attempting to start hand [%d]", t.currentHand.ID+1)

	pot := t.pot.Sum()
	if pot > 0 {
		return errors.New("pot still has chips, cannot start new hand until the previous one is resolved")
	}

	for _, p := range t.players {
		p.pocket = make([]*card, 0)
		p.potContribution = 0
		p.currentBet = make(stack)

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

// a structure that tracks a pot - particular in split pot
// situations. contains all players that have contributed to the pot
// and their contribution amount.
type pot struct {
	// the set of all players who contributed to this pot
	players map[int]any

	// the amount of chips from each player that has been
	// contributed to this pot
	contribution int
}

// Sum gets the sum of all chips in this pot
func (p pot) Sum() int {
	return len(p.players) * p.contribution
}

// IsEntitled indicates true if the provided player is a contributor to this pot
func (p pot) IsEntitled(position int) bool {
	_, ok := p.players[position]
	return ok
}

// getSplitPots gathers all pots that are created from short stacks going all in.
// this depends on the potContribution field of players and returns a data structure
// which maps total pot contribution to the pot struct.
//
// note that the contribution field of the pot struct is not _total_ contribution,
// that value is encoded in the mapping key, it is strictly the amount contributed
// to that specific pot
func (t *table) getSplitPots() map[int]pot {
	contributions := make([]*player, len(t.players))
	copy(contributions, t.players)

	sort.Slice(contributions, func(i int, j int) bool {
		return contributions[i].potContribution < contributions[j].potContribution
	})

	pots := make(map[int]pot)
	current := 0
	for _, p := range contributions {
		// if this players total pot contribution is not in the mapping yet
		// we need to add it
		_, ok := pots[p.potContribution]
		if !ok {
			pots[p.potContribution] = pot{
				players: make(map[int]any),

				// we subtract current because this value tracks how much
				// all players have contributed to already created pots
				contribution: p.potContribution - current,
			}

			// we update current, this reflects the current amount of total
			// chips that we have put into pots. this relies
			// on the ascending order of contributions being iterated.
			current = p.potContribution
		}

		// add this player to all pots already computed, they have contributed
		// to all previous pots
		for _, pot := range pots {
			pot.players[p.position] = struct{}{}
		}
	}

	return pots
}

// handPayout computes winnings that sould be paid from the current table state.
// it requires a map of position -> eval and a map of position -> payout as arguments.
// when the function terminates the values in winnings will hold the amount from the pot
// that each position should be awarded.
func (t *table) handlePayout(
	handEvaluations map[int]int,
	pot pot,
) (winnings map[int]int, err error) {
	winners := make([]int, 0)
	bestHandEval := math.MaxInt
	for position, eval := range handEvaluations {
		// if this player is not entitled to the pot we skip them
		if !pot.IsEntitled(position) {
			continue
		}

		// if this eval is strictly less then our current best
		// we need to reset the winners array for this hand
		if eval < bestHandEval {
			winners = make([]int, 0)
			bestHandEval = eval
		}

		if eval == bestHandEval {
			winners = append(winners, position)
		}
	}

	if len(winners) == 0 {
		return nil, just.NewPokerError("critical error - no winner for a given pot", just.Unknown)
	}

	winnings = make(map[int]int)
	for _, position := range winners {
		winnings[position] = pot.Sum() / len(winners)
	}

	remainder := pot.Sum() % len(winners)

	// need to update this to get the player from winners who is nearest the button
	winnings[winners[0]] += remainder

	return winnings, nil
}
