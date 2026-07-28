package game

import (
	"context"
	"fmt"
	"maps"
	"net/http"
	"poker_server/just"
	"strconv"
	"sync"
	"time"
)

func (this stack) MergeWith(that stack) stack {
	result := make(stack)
	for d, c := range that {
		result[d] += c
	}

	for d, c := range this {
		result[d] += c
	}

	return result
}

type PlayerActionThing struct {
	action    PlayerActionDTO
	onError   chan error
	onSuccess chan any
}

type game struct {
	started_at          *time.Time
	id                  string
	joinGameLock        sync.Mutex
	config              NewGameConfigDTO
	table               *table
	playerActionChannel chan PlayerActionThing
}

func (g *game) AsDTO() GameDTO {
	return GameDTO{
		StartedAt: g.started_at,
		ID:        g.id,
		Config:    g.config,
		Table:     g.table.AsDTO(),
	}
}

func CreateGameFromConfig(config NewGameConfigDTO) (*game, error) {
	g := &game{
		joinGameLock:        sync.Mutex{},
		config:              config,
		table:               &table{},
		playerActionChannel: make(chan PlayerActionThing),
	}

	g.table.players = make([]*player, 0)
	g.table.deck = &deck{}

	conn, err := just.DBConnPool.Acquire(context.Background())
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var id int
	stmt := `insert into just_poker_game (starting_chips, player_count) values ($1, $2) RETURNING game_id`
	err = conn.QueryRow(context.Background(), stmt, config.StartingChips, config.PlayerCount).Scan(&id)
	if err != nil {
		return nil, err
	}

	g.id = strconv.Itoa(id)
	return g, nil
}

// TODO: should this just be called and return error? do I need
// a response channel really?
func (g *game) TryJoinGame(username, userID string, ch chan just.HttpResponse) {
	g.joinGameLock.Lock()
	defer g.joinGameLock.Unlock()

	if g.started_at != nil {
		ch <- just.BadRequest("game has already started", 0)
		return
	}

	if g.table == nil {
		ch <- just.HttpResponse{
			Code: http.StatusInternalServerError,
			Object: map[string]string{
				"msg": "table does not yet exist, game has not been initialized",
			},
		}

		return
	}

	if len(g.table.players) >= g.config.PlayerCount {
		ch <- just.BadRequest("the table is full", 0)
		return
	}

	// check for duplicate player
	for _, p := range g.table.players {
		if p.UserID == userID {
			ch <- just.BadRequest("player already exists at table dayo", 0)
			return
		}
	}

	p := &player{
		UserID:      userID,
		UserType:    "test",
		DisplayName: username,
	}

	p.chips = make(stack)
	maps.Copy(p.chips, g.config.StartingChips.AsStack())

	g.table.players = append(g.table.players, p)
	ch <- just.OK("join table success", 0)
}

func (g *game) TryStartGame() error {
	g.joinGameLock.Lock()
	defer g.joinGameLock.Unlock()

	if g.started_at != nil {
		return fmt.Errorf("game already started")
	}

	ctx := context.Background()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		return err
	}

	stmt := `
	UPDATE public.just_poker_game 
	SET player_ids = array_cat(player_ids, $1::bigint[])
	WHERE game_id = $2`

	players := make([]int, len(g.table.players))
	for i, p := range g.table.players {
		id, err := strconv.Atoi(p.UserID)
		if err != nil {
			return err
		}

		players[i] = id
		p.position = i
	}

	_, err = conn.Exec(ctx, stmt, players, g.id)
	if err != nil {
		return err
	}

	err = g.TryStartNewHand()
	if err != nil {
		return err
	}

	t := time.Now()
	g.started_at = &t
	return nil
}

func (g *game) TryStartNewHand() error {
	just.Logger.Debugf("attempting to start hand [%d]", g.table.currentHand.Count+1)

	t := g.table

	pot := t.pot.Sum()
	if pot > 0 {
		return fmt.Errorf("pot still has chips, cannot start new hand until the previous one is resolved")
	}

	// TODO: growing blinds based on configurations
	t.currentHand.SmallBlind = g.config.SmallBlind
	t.currentHand.BigBlind = g.config.BigBlind
	t.currentHand.Button = (t.currentHand.Button + 1) % len(t.players)

	t.deck.Reset()
	t.street = make([]*card, 0)
	t.pot = make(map[int]int)

	t.currentRound.currentRoundType = round_type_unset
	t.NextRound()
	t.currentRound.currentAggressor = t.NextPosition(t.currentHand.Button + 2)

	return nil
}

func (g *game) TryPlayerAction(action PlayerActionDTO) error {
	errorChannel := make(chan error)
	successChannel := make(chan any)
	playerAction := PlayerActionThing{
		action:    action,
		onError:   errorChannel,
		onSuccess: successChannel,
	}

	g.playerActionChannel <- playerAction
	select {
	case _ = <-successChannel:
		return nil
	case err := <-errorChannel:
		return err
	}
}

func (g *game) ProccessPlayerActions(exit chan any) {
	for {
		select {
		case _ = <-exit:
			just.Logger.Debug("exit signal received while processing player actions")
			return
		case action := <-g.playerActionChannel:
			just.Logger.Debug("processing player action")
			err := g.HandlePlayerAction(action.action)
			if err != nil {
				action.onError <- err
			} else {
				action.onSuccess <- struct{}{}
			}
		}
	}
}

func (g *game) TryCoverBet(p *player, chips ChipStackDTO) error {
	// first we go through and ensure that the player can cover
	// every chip they want to bet, we do this before changing any chips
	// in case there is an error (we dont want to unwind)
	stack := chips.AsStack()
	for d, c := range stack {
		player_count, ok := p.chips[d]
		if !ok {
			just.Logger.Errorf("invalid denomination %d %s", d, p.chips.ToString())
			return &just.PokerError{
				Message: fmt.Sprintf("player has no chips with %d denomination", d),
				Code:    just.NotEnoughChips,
			}
		}

		if c < 0 {
			return &just.PokerError{
				Message: "received negative chip amount",
				Code:    just.InvalidBetAmount,
			}
		}

		if player_count < c {
			return &just.PokerError{
				Message: fmt.Sprintf(
					"player cannot cover %d of %d chips with their current count of %d",
					c,
					d,
					player_count),
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

func (a PlayerActionDTO) ToString() string {
	return fmt.Sprintf("player: %s chips: %#v intent: %s", a.PlayerID, a.Bet, a.Intent)
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
func (g *game) HandlePlayerAction(action PlayerActionDTO) error {
	just.Logger.Debugf("handling action attempt for player [%s]", action.ToString())
	p := g.table.players[g.table.currentRound.currentPlayerPosition]

	// we need to first ensure that it is this players turn
	if action.PlayerID != p.UserID {
		return &just.PokerError{
			Message: fmt.Sprintf("turn order violation - play is currently at player in position %d", p.position),
			Code:    just.TurnOrderViolation,
		}
	}

	// the comment below is the truth.
	// the comment above is a lie. This totally works ;)
	// - Goblinz181
	switch action.Intent {
	case player_intent_ante:
		if g.table.currentRound.currentRoundType != round_type_setup {
			return &just.PokerError{
				Message: "during this phase only ante actions can be accepted",
				Code:    just.InvalidActionType,
			}
		}

		betAmount := action.Bet.Sum()
		if p.position == (g.table.NextPosition(g.table.currentHand.Button)) {
			if betAmount != g.table.currentHand.SmallBlind {
				return &just.PokerError{
					Message: fmt.Sprintf("small blind requires exactly %d chips", g.table.currentHand.SmallBlind),
					Code:    just.InvalidBetAmount,
				}
			}

			if err := g.TryCoverBet(p, action.Bet); err != nil {
				return err
			}

		} else if p.position == (g.table.NextPosition(g.table.currentHand.Button + 1)) {
			if betAmount != g.table.currentHand.BigBlind {
				return &just.PokerError{
					Message: fmt.Sprintf("big blind requires exactly %d chips", g.table.currentHand.BigBlind),
					Code:    just.InvalidBetAmount,
				}
			}

			if err := g.TryCoverBet(p, action.Bet); err != nil {
				return err
			}

		} else {
			just.Logger.Errorf(
				"play is currently at %d turn during setup round however they are neither the big nor small blind. game is deadlocked. game state: %v",
				p.position,
				g.AsDTO())

			return &just.PokerError{
				Message: "critical error - game state cannot progress - alert game master",
				Code:    just.Unknown,
			}
		}

		p.currentBet = p.currentBet.MergeWith(action.Bet.AsStack())
		g.table.currentRound.bet = g.config.BigBlind

	case player_intent_check:
		if p.currentBet.Sum() != g.table.currentRound.bet {
			return &just.PokerError{
				Message: fmt.Sprintf(
					"you must call the current bet of %d with %d chips",
					g.table.currentRound.bet,
					g.table.currentRound.bet-p.currentBet.Sum()),
				Code: just.InvalidBetAmount,
			}
		}

	case player_intent_call:
		totalBet := p.currentBet.MergeWith(action.Bet.AsStack())
		if totalBet.Sum() != g.table.currentRound.bet {
			return &just.PokerError{
				Message: fmt.Sprintf(
					"%d is not valid to call the current amount of %d",
					totalBet.Sum(),
					g.table.currentRound.bet),
				Code: just.InvalidBetAmount,
			}
		}

		p.currentBet = totalBet

	case player_intent_raise:
		totalBet := p.currentBet.MergeWith(action.Bet.AsStack())
		if totalBet.Sum() <= g.table.currentRound.bet {
			return &just.PokerError{
				Message: fmt.Sprintf(
					"%d is not valid to raise the current amount of %d",
					g.table.currentRound.bet,
					g.table.currentRound.bet),
				Code: just.InvalidBetAmount,
			}
		}

		if err := g.TryCoverBet(p, action.Bet); err != nil {
			return err
		}

		p.currentBet = totalBet
		g.table.currentRound.bet = p.currentBet.Sum()
		g.table.currentRound.currentAggressor = p.position

	case player_intent_all_in:
		p.currentBet = p.currentBet.MergeWith(p.chips)
		g.table.currentRound.bet += p.chips.Sum()

		// check if this all in would actually make the player thhe aggressor,
		// it could be that they just don't have the chips to cover the current
		// round bet
		if p.currentBet.Sum() > g.table.currentRound.bet {
			g.table.currentRound.currentAggressor = p.position
		}

		p.state = player_state_all_in

	case player_intent_fold:
		p.state = player_state_folded

	default:
		return fmt.Errorf("erm dunno what %s means, sorry", action.Intent)
	}

	just.Logger.Debugf("accepted action [%s] for player with id [%s]", action.Intent, action.PlayerID)

	// after we get to this state we know that the player action
	// has been accepted and we need to compute what to do next
	nextPlayer := g.table.NextInactivePlayer(g.table.currentRound.currentPlayerPosition)
	if nextPlayer == nil {
		nextPlayer = g.table.players[g.table.currentRound.currentAggressor]
	}

	just.Logger.Debugf("next inactive player received: [%d] -> [%d]", p.position, nextPlayer.position)

	// detects when the round should end - when the next player to act
	// would be the current aggressor. The current aggressor is the last
	// one to raise, or start the round.
	if g.table.currentRound.currentAggressor == nextPlayer.position {
		// before shifting to the next round we need to figure out whose turn it is
		// in general we need to ensure that:
		// - the player is not all-in or folded
		// - if it is pre-flop the initial player to check is clockwise the big blind
		// - for all other rounds the initial player to check is clockwise the button
		//
		// note that it is possible to terminate this loop without matching to a player.
		// this will occurr when all players are all in. we still want to go through
		// all rounds in order to produce cards for the river/turn
		var p *player
		for p == nil && g.table.currentRound.currentRoundType != round_type_completed {
			// recording the previous round here for logging
			prevRoundType := g.table.currentRound.currentRoundType

			g.table.NextRound()
			if g.table.currentRound.currentRoundType == round_type_pre_flop {
				p = g.table.NextInactivePlayer(g.table.currentHand.Button + 2)
			} else {
				p = g.table.NextInactivePlayer(g.table.currentHand.Button)
			}

			just.Logger.Debugf(
				"starting next round for game [%s] [%s]->[%s]",
				g.id,
				prevRoundType,
				g.table.currentRound.currentRoundType)
		}

		if p != nil {
			g.table.currentRound.currentPlayerPosition = p.position
			if g.table.currentRound.currentRoundType == round_type_pre_flop {
				g.table.currentRound.currentAggressor = g.table.NextPosition(g.table.currentHand.Button + 2)
			} else {
				g.table.currentRound.currentAggressor = p.position
			}
		}

	} else {
		g.table.currentRound.currentPlayerPosition = nextPlayer.position
	}

	g.table.currentTurn.StartedAt = time.Now()

	if g.table.currentRound.currentRoundType == round_type_completed {
		err := g.TryStartNewHand()
		if err != nil {
			return err
		}
	}

	return nil
}
