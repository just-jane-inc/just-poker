package game

import (
	"fmt"
	"time"

	"github.com/just-jane-inc/just-poker/server/just"
)

type playerAction struct {
	action    PlayerActionDTO
	onError   chan error
	onSuccess chan any
	onMessage chan just.WebsocketMessage[any]
}

func (a PlayerActionDTO) ToString() string {
	return fmt.Sprintf("chips: %#v intent: %s", a.Bet, a.Intent)
}

func (g *game) TryPlayerAction(action PlayerActionDTO) error {
	errorChannel := make(chan error)
	successChannel := make(chan any)
	playerAction := playerAction{
		action:    action,
		onError:   errorChannel,
		onSuccess: successChannel,
	}

	g.playerActionChannel <- &playerAction

	select {
	case <-successChannel:
		return nil
	case err := <-errorChannel:
		return err
	}
}

func (g *game) ProccessPlayerActions(exit chan any) {
	for {
		select {
		case <-exit:
			just.Logger.Debug("exit signal received while processing player actions")
			return
		case action := <-g.playerActionChannel:
			err := g.handlePlayerAction(action.action)
			if err != nil {
				action.onError <- err
			} else {
				action.onSuccess <- struct{}{}
			}
		}
	}
}

func (g *game) coverBet(p *player, chips ChipStackDTO) *just.PokerError {
	// first we go through and ensure that the player can cover
	// every chip they want to bet, we do this before changing any chips
	// in case there is an error (we dont want to unwind)
	stack := chips.asStack()
	for d, c := range stack {
		if c < 0 {
			return &just.PokerError{
				Message: "received negative chip amount",
				Code:    just.InvalidBetAmount,
			}
		}

		if c == 0 {
			continue
		}

		chipCount, ok := p.chips[d]
		if !ok {
			return &just.PokerError{
				Message: fmt.Sprintf("player has no chips with %d denomination", d),
				Code:    just.NotEnoughChips,
			}
		}

		if chipCount < c {
			return &just.PokerError{
				Message: fmt.Sprintf(
					"player cannot cover %d of %d chips with their current count of %d",
					c,
					d,
					chipCount,
				),
				Code: just.NotEnoughChips,
			}
		}
	}

	// now that we have determined the chips can be covered we actually move them
	// from the player to the pot
	for denomination, count := range stack {
		p.chips[denomination] -= count

		if _, exists := g.table.pot[denomination]; !exists {
			g.table.pot[denomination] = 0
		}

		g.table.pot[denomination] += count
	}

	return nil
}

/*
Texas Hold’em Rules: Flow of a Hand
At the beginning of the first hand of play, one player will be assigned the dealer button (in home games,
this player will also traditionally act as the dealer for that hand). The player immediately to the left
of the button must post the small blind, while the player two seats to the left of the button must
post the big blind. The size of these blinds is typically determined by the rules of the game.
If any ante is required – common in a tournament situation – players should also contribute it at this point.

Once all blinds have been posted and antes have been paid, the dealer will deal two cards to each player.
Each player may examine their own cards. The play begins with the player to the left of the big blind.
That player may choose to fold, in which case they forfeit their cards and are done with play for that
hand. The player may also choose to call the bet, placing an amount of money into the pot equal to
the size of the big blind. Finally, the player can also choose to raise, increasing the size of the
bet required for other players to stay in the hand.

Moving around the table clockwise, each player may then choose to take any of those options: folding,
calling the current bet, or raising the bet. A round of betting ends when all players but one have
folded (in which case the one remaining player wins the pot), or when all remaining players have
called the current bet. On the first round of betting, if no players raise, the big blind will also
have the option to check, essentially passing his turn; this is because the big blind has already placed
the current bet amount into the pot, but hasn’t yet had a chance to act.

Assuming there are two or more players remaining in the hand after the first round of betting, the
dealer will then deal out three community cards in the middle of the table. These cards are known as
the flop. Play now begins, starting with the first player to the left of the dealer button (if every
player is still in the hand, this will be the small blind). Players have the same options as before;
in addition, if no bet has yet been made in the betting round, players have the option to check. A
round of betting can also end if all players check and no bets are made, along with the other ways discussed above.

If two or more players remain in the hand after the second round of betting, the dealer will place
a fourth community card – known as the turn – on the table. Once again, a round of betting ensues,
using the same rules outlined above. Finally, if two or more players are still around after the third
round of betting, the dealer will place the final community card – the river – on the table. One last
round of betting will commence.

After this final round of betting, all remaining players must reveal their hands. The player with the
best hand according to the hand rankings above will win the pot. If two or more players share the exact
same hand, the pot is split evenly between them. After each hand, the button moves one seat to the
left, as do the responsibilities of posting the small and big blinds.
*/
func (g *game) handlePlayerAction(action PlayerActionDTO) *just.PokerError {
	just.Logger.Debugf(
		"process action [%s] from [%s] in game [%s] - current round type [%s]",
		action.ToString(),
		action.PlayerID,
		g.id,
		g.table.currentRound.currentRoundType,
	)

	if g.table.currentRound.currentRoundType == RoundTypeCompleted {
		return just.NewPokerError(
			"round type is completed - no actions can be received",
			just.TurnOrderViolation,
		)
	}

	p := g.table.players[g.table.currentRound.currentPlayerPosition]
	var actingPlayer *player
	if actingPlayer = g.table.GetPlayerWithID(action.PlayerID); actingPlayer == nil {
		return &just.PokerError{
			Message: fmt.Sprintf("player [%s] not found in game [%s]", action.PlayerID, g.id),
			Code:    just.UserNotFound,
		}
	}

	// we need to first ensure that it is this players turn
	if actingPlayer.UserID != p.UserID {
		return &just.PokerError{
			Message: fmt.Sprintf(
				"turn order violation - [%d] attempted to act during [%d] turn",
				actingPlayer.position,
				p.position,
			),
			Code: just.TurnOrderViolation,
		}
	}

	// the ante 'round' has unique logic and should not mix
	// with the other round types
	if g.table.currentRound.currentRoundType == RoundTypeAnte {
		if err := g.handleAnte(action, p); err != nil {
			return err
		}
	}

	// the comment below is the truth.
	// the comment above is a lie. This totally works ;)
	// - Goblinz181
	switch action.Intent {
	case PlayerIntentAnte:
		// the happy path of the ante intent will have already been handled by this point
		// we do not need to do anything else here, simply ensure that this intent
		// was applied during the setup round.
		if g.table.currentRound.currentRoundType != RoundTypeAnte {
			return &just.PokerError{
				Message: "ante is only accepted during the setup round",
				Code:    just.InvalidActionType,
			}
		}
	case PlayerIntentCheck:
		if p.currentBet.Sum() != g.table.currentRound.bet {
			return &just.PokerError{
				Message: fmt.Sprintf(
					"you must call the current bet of %d with %d chips",
					g.table.currentRound.bet,
					g.table.currentRound.bet-p.currentBet.Sum(),
				),
				Code: just.InvalidBetAmount,
			}
		}

	case PlayerIntentCall:
		totalBet := p.currentBet.mergeWith(action.Bet.asStack())
		if totalBet.Sum() != g.table.currentRound.bet {
			return &just.PokerError{
				Message: fmt.Sprintf(
					"%d is not valid to call the current amount of %d",
					totalBet.Sum(),
					g.table.currentRound.bet,
				),
				Code: just.InvalidBetAmount,
			}
		}

		if action.Bet.Sum() == 0 {
			return just.NewPokerError("did you mean to check?", just.InvalidActionType)
		}

		if err := g.coverBet(p, action.Bet); err != nil {
			return err
		}

		p.currentBet = totalBet

	case PlayerIntentRaise:
		totalBet := p.currentBet.mergeWith(action.Bet.asStack())
		if totalBet.Sum() <= g.table.currentRound.bet {
			return &just.PokerError{
				Message: fmt.Sprintf(
					"%d is not valid to raise the current amount of %d",
					g.table.currentRound.bet,
					g.table.currentRound.bet,
				),
				Code: just.InvalidBetAmount,
			}
		}

		if err := g.coverBet(p, action.Bet); err != nil {
			return err
		}

		p.currentBet = totalBet
		g.table.currentRound.bet = p.currentBet.Sum()
		g.table.currentRound.currentAggressor = p.position

	case PlayerIntentAllIn:
		p.currentBet = p.currentBet.mergeWith(p.chips)
		g.table.currentRound.bet += p.chips.Sum()

		if err := g.coverBet(p, p.chips.AsDto()); err != nil {
			return err
		}

		// check if this all in would actually make the player thhe aggressor,
		// it could be that they just don't have the chips to cover the current
		// round bet
		if p.currentBet.Sum() > g.table.currentRound.bet {
			g.table.currentRound.currentAggressor = p.position
		}

		p.state = PlayerStateAllIn

	case PlayerIntentFold:
		p.state = PlayerStateFolded

	default:
		return just.NewPokerError("unknown intent encountered - please read documentation", just.SkillIssue)
	}

	just.Logger.Debugf("accepted action [%s] for player with id [%s]", action.Intent, action.PlayerID)

	// the signal to websocket listeners for a player action
	// if any errors are encountered after this happens we are
	// in big touble
	g.table.sendMessageToConnections("player_action", action)

	// if everyone but one has folded we are like done now, we dont even need to keep showing cards just like end it

	// after we get to this state we know that the player action
	// has been accepted and we need to compute what to do next
	nextPlayer := g.table.NextPlayer(g.table.currentRound.currentPlayerPosition)
	for nextPlayer.state != PlayerStateInactive && nextPlayer.position != g.table.currentRound.currentAggressor {
		nextPlayer = g.table.NextInactivePlayer(nextPlayer.position)
		if nextPlayer == nil {
			nextPlayer = g.table.players[g.table.currentRound.currentAggressor]
		}
	}

	remainingPlayers := 0
	for _, p := range g.table.players {
		if p.state == PlayerStateFolded {
			continue
		}

		if p.state == PlayerStateOut {
			continue
		}

		remainingPlayers += 1
	}

	just.Logger.Debugf("next inactive player received: [%d] -> [%d]", p.position, nextPlayer.position)

	// detects when the round should end - when the next player to act
	// would be the current aggressor. The current aggressor is the last
	// one to raise, or start the round.
	if remainingPlayers == 1 {
		g.table.currentRound.currentRoundType = RoundTypeRiver
		g.table.nextRound()
	} else if g.table.currentRound.currentAggressor == nextPlayer.position {
		// before shifting to the next round we need to figure out whose turn it is
		// in general we need to ensure that:
		// - the player is not all-in or folded
		// - if it is pre-flop the initial player to check is clockwise the big blind
		// - for all other rounds the initial player to check is clockwise the button
		//
		// note that it is possible to terminate this loop without matching to a player.
		// this will occurr when all players are all in. we still want to go through
		// all rounds in order to produce cards for the river/turn
		var nextRoundFirstPlayer *player
		for nextRoundFirstPlayer == nil && g.table.currentRound.currentRoundType != RoundTypeCompleted {
			// recording the previous round here for logging
			prevRoundType := g.table.currentRound.currentRoundType

			g.table.nextRound()
			if g.table.currentRound.currentRoundType == RoundTypePreFlop {
				nextRoundFirstPlayer = g.table.NextInactivePlayer(g.table.bigBlindPosition)
			} else {
				nextRoundFirstPlayer = g.table.NextInactivePlayer(g.table.buttonPosition)
			}

			just.Logger.Debugf(
				"starting next round for game [%s] [%s]->[%s]",
				g.id,
				prevRoundType,
				g.table.currentRound.currentRoundType,
			)
		}

		if nextRoundFirstPlayer != nil {
			g.table.currentRound.currentAggressor = nextRoundFirstPlayer.position
		}

		nextPlayer = nextRoundFirstPlayer
	}

	// the player state should only change back to inactive if it
	// is still active. the player may have gone all in or folded
	// and that state should not be overwritten
	if p.state == PlayerStateActive {
		p.state = PlayerStateInactive
	}

	// it is possible for p == nextPlayer, we need to ensure that
	// this happens after setting p state to inactive.
	if nextPlayer != nil {
		g.table.currentRound.currentPlayerPosition = nextPlayer.position
		nextPlayer.state = PlayerStateActive
		just.Logger.Debugf("[%s] [%d] -> [%d]", g.table.currentRound.currentRoundType, p.position, nextPlayer.position)
	}

	g.table.currentTurn.StartedAt = time.Now()
	g.table.currentTurn.ID += 1

	if g.table.currentRound.currentRoundType == RoundTypeCompleted {
		for _, p := range g.table.players {
			if p.chips.Sum() == 0 {
				p.state = PlayerStateOut
			}
		}
	}

	return nil
}

func (g *game) IsGameOver() bool {
	playersRemaining := 0
	for _, p := range g.table.players {
		if p.state == PlayerStateOut {
			continue
		}

		playersRemaining += 1
	}

	return playersRemaining <= 1
}

// handleAnte handles provided player action for the setup round type
//
// returns an error if the game is not actually in this round type
// or if the provided action is invalid in some way.
//
// it is assumed that the provided player (p) preforms the provided action.
func (g *game) handleAnte(action PlayerActionDTO, p *player) *just.PokerError {
	if g.table.currentRound.currentRoundType != RoundTypeAnte {
		err, id := just.NewCriticalInternalError()
		just.Logger.Errorf(
			"game [%s] invoked handleAnte during the [%s] round on turn [%d] ERROR-ID={%s}",
			g.id,
			g.table.currentRound.currentRoundType,
			g.table.currentTurn.ID,
			id,
		)

		return err
	}

	if action.Intent != PlayerIntentAnte {
		return &just.PokerError{
			Message: "during this phase only ante actions can be accepted",
			Code:    just.InvalidActionType,
		}
	}

	betAmount := action.Bet.Sum()
	switch p.position {
	case g.table.smallBlindPosition:
		if betAmount != g.table.currentHand.SmallBlind {
			return &just.PokerError{
				Message: fmt.Sprintf("small blind requires exactly %d chips", g.table.currentHand.SmallBlind),
				Code:    just.InvalidBetAmount,
			}
		}

	case g.table.bigBlindPosition:
		if betAmount != g.table.currentHand.BigBlind {
			return &just.PokerError{
				Message: fmt.Sprintf("big blind requires exactly %d chips", g.table.currentHand.BigBlind),
				Code:    just.InvalidBetAmount,
			}
		}

	default:
		err, id := just.NewCriticalInternalError()
		just.Logger.Errorf(
			"error in game [%s] during turn [%d] in handleAnte player [%s] is neither big nore small blind, game cannot continue ERROR-ID={%s}",
			g.id,
			g.table.currentTurn.ID,
			p.UserID,
			id,
		)

		return err
	}

	if err := g.coverBet(p, action.Bet); err != nil {
		return err
	}

	g.table.currentRound.bet = g.config.BigBlind
	p.currentBet = p.currentBet.mergeWith(action.Bet.asStack())
	return nil
}
