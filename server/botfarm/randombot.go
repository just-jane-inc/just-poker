// Package botfarm
// Collection of Farms - bwopbwop1
package botfarm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"slices"

	"github.com/just-jane-inc/just-poker/server/game"
	"github.com/just-jane-inc/just-poker/server/just"
)

// PokerBot godoc
//
//		An interface for implementing bots in golang
//
//	 What we know is a drop, what we don't know is an ocean. -hhei21
type PokerBot interface {
	OnGameStateUpdate(game.GameDTO)
}

type RandomBot struct {
	id        string
	position  int
	playerDTO *game.PlayerDTO
	conn      *just.PlayerUpdateConnection
	token     string
	gameID    string
}

// CreateRandomBot godoc
//
// It is at this moment jane regretted out sourcing comments to chat - Goblinz181
func CreateRandomBot(gameID string, user just.AuthorizedUser, token string) RandomBot {
	bot := RandomBot{
		id:        user.ID,
		position:  -1,
		playerDTO: nil,
		token:     token,
		gameID:    gameID,
	}

	bot.conn = just.UpdateHub.AddPlayerToHub(gameID, &user)
	go func() {
		for e := range bot.conn.MessageChannel {
			just.Logger.Debugf("received event [%s] in bot: %s", e.EventType, user.Name)
			if e.EventType != string(game.GameStatusChanged) {
				continue
			}

			// Just refactor using MONADS - mausHAX
			gameDTO := e.Data.(game.GameDTO)
			if bot.position == -1 {
				for _, player := range gameDTO.Table.Players {
					if player.UserID == bot.id {
						just.Logger.Debugf("setting player position %d for bot %s", player.Position, user.Name)
						bot.position = player.Position
						break
					}
				}
			}

			bot.OnGameStateUpdate(gameDTO)
		}
	}()

	return bot
}

func (b *RandomBot) OnGameStateUpdate(dto game.GameDTO) {
	just.Logger.Debugf("handling game state update for random bot [%s] in game [%s]", b.id, dto.ID)
	if dto.Table.CurrentRound.CurrentPlayerPosition != b.position {
		just.Logger.Debugf("it is current %d turn skipping for %d", dto.Table.CurrentRound.CurrentPlayerPosition, b.position)
		return
	}

	just.Logger.Debugf("it is currently random bots turn...")
	if dto.Table.CurrentRound.CurrentRoundType == game.RoundTypeCompleted {
		return
	}

	if dto.Table.CurrentRound.CurrentRoundType == game.RoundTypeUnset {
		return
	}

	just.Logger.Debugf("producing move for bot")

	b.playerDTO = &dto.Table.Players[b.position]
	validActions := b.getValidActions(dto)

	just.Logger.Debugf("received %d of valid actions", len(validActions))
	intent := validActions[rand.IntN(len(validActions))]
	just.Logger.Debugf("selected %s action", intent)

	var err error
	var chips game.ChipStackDTO
	switch intent {
	case game.PlayerIntentAnte:
		chips, err = b.getChipsFromStack(dto.Table.CurrentRound.Bet, dto.Config.ChipDenominations)
		if err != nil {
			just.Logger.Errorf("encountered error while getting chips from stack: %v", err)
			break
		}

		err = b.sendAction(game.PlayerActionDTO{Intent: game.PlayerIntentAnte, Bet: chips})

	case game.PlayerIntentCheck:
		err = b.sendAction(game.PlayerActionDTO{Intent: game.PlayerIntentCheck})
	case game.PlayerIntentFold:
		err = b.sendAction(game.PlayerActionDTO{Intent: game.PlayerIntentFold})
	case game.PlayerIntentCall:
		chips, err = b.getChipsFromStack(dto.Table.CurrentRound.Bet-b.playerDTO.CurrentBet.Sum(), dto.Config.ChipDenominations)
		if err != nil {
			just.Logger.Errorf("encountered error while getting chips from stack: %v", err)
			break
		}

		err = b.sendAction(game.PlayerActionDTO{Intent: game.PlayerIntentCall, Bet: chips})
	case game.PlayerIntentRaise:
		maxRaise := b.playerDTO.Stack.Sum() / dto.Table.CurrentHand.BigBlind
		amountToCall := dto.Table.CurrentRound.Bet - b.playerDTO.CurrentBet.Sum()
		minRaise := (amountToCall / dto.Table.CurrentHand.BigBlind) + 1

		just.Logger.Infof("getting raise from range of min: %d max: %d", minRaise, maxRaise)

		var raiseAmount int
		if maxRaise < minRaise {
			raiseAmount = minRaise
		} else if maxRaise == minRaise {
			raiseAmount = minRaise
		} else {
			raiseAmount = rand.IntN(maxRaise-minRaise) + minRaise
		}

		raiseAmount = raiseAmount * dto.Table.CurrentHand.BigBlind
		chips, err = b.getChipsFromStack(raiseAmount, dto.Config.ChipDenominations)
		if err != nil {
			just.Logger.Errorf("encountered error while getting chips from stack: %v", err)
			break
		}

		err = b.sendAction(game.PlayerActionDTO{Intent: game.PlayerIntentRaise, Bet: chips})
	case game.PlayerIntentAllIn:
		err = b.sendAction(game.PlayerActionDTO{Intent: game.PlayerIntentAllIn})
	}

	if err != nil {
		just.Logger.Errorf("error encountered making move for random bot: %v", err)
	}
}

func (b *RandomBot) getChipsFromStack(amount int, denominations []int) (game.ChipStackDTO, error) {
	// check if we just have, in a single denomination, the chips that are required
	slices.Sort(denominations)
	for denomination, count := range b.playerDTO.Stack {
		if count <= 0 {
			continue
		}

		if amount%denomination == 0 && amount/denomination <= count {
			return game.ChipStackDTO{denomination: amount / denomination}, nil
		}
	}

	sumThing := 0
	for denomination := range denominations {
		count, ok := b.playerDTO.Stack[denomination]
		if !ok {
			continue
		}

		if denomination < amount {
			sumThing += count * denomination
		}
	}

	slices.Reverse(denominations)
	// we can for sure make the thing with chips that we have without doing exchanges
	if sumThing > amount {
		stack := game.ChipStackDTO{}
		current := amount
		for denomination := range denominations {
			if current == 0 {
				break
			}

			count, ok := b.playerDTO.Stack[denomination]
			if !ok {
				continue
			}

			if denomination > amount {
				continue
			}

			for range count {
				current -= denomination
				stack[denomination] += 1
				if current == 0 {
					break
				}
			}
		}

		return stack, nil
	}

	// we require an exchange
	for i, d := range denominations {
		count, ok := b.playerDTO.Stack[d]
		if !ok {
			continue
		}

		if count <= 0 {
			continue
		}

		nextGuy := denominations[i+1]
		count = d / nextGuy
		exchangeDTO := game.ChipExchangeDTO{Give: game.ChipStackDTO{d: 1}, Receive: game.ChipStackDTO{nextGuy: count}}
		if err := b.chipExchange(exchangeDTO); err != nil {
			return nil, err
		}

		return b.getChipsFromStack(amount, denominations)
	}

	return nil, fmt.Errorf("erm")
}

func (b *RandomBot) chipExchange(dto game.ChipExchangeDTO) error {
	just.Logger.Infof("preforming chip exchange: %-v", dto)
	jsonData, err := json.Marshal(dto)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://localhost:%s/game/%s/chip/exchange", just.Env.Port, b.gameID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+b.token)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	someStuff, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	just.Logger.Infof("%s", string(someStuff))

	for denomination, count := range dto.Give {
		b.playerDTO.Stack[denomination] -= count
	}

	for denomination, count := range dto.Receive {
		b.playerDTO.Stack[denomination] += count
	}

	return nil
}

func (b *RandomBot) getValidActions(gameDTO game.GameDTO) []game.PlayerIntent {
	if gameDTO.Table.CurrentRound.CurrentRoundType == game.RoundTypeAnte {
		return []game.PlayerIntent{game.PlayerIntentAnte}
	}

	actions := []game.PlayerIntent{game.PlayerIntentFold, game.PlayerIntentAllIn}
	me := gameDTO.Table.Players[b.position]

	playerBet := me.CurrentBet.Sum()
	if gameDTO.Table.CurrentRound.Bet == playerBet {
		actions = append(actions, game.PlayerIntentCheck)
	}

	if gameDTO.Table.CurrentRound.Bet > playerBet {
		if me.Stack.Sum() >= gameDTO.Table.CurrentRound.Bet-playerBet {
			actions = append(actions, game.PlayerIntentCall)
		}

		if me.Stack.Sum() > gameDTO.Table.CurrentRound.Bet-playerBet-gameDTO.Config.BigBlind {
			actions = append(actions, game.PlayerIntentRaise)
		}
	}

	return actions
}

// sendAction godoc
//
// Gambling is an art and if this bot was an artist it would probably die broke \o/ - bwopbwop1
func (b RandomBot) sendAction(action game.PlayerActionDTO) error {
	just.Logger.Infof("sending action for bot %s with intent %s", b.playerDTO.DisplayName, action.Intent)
	jsonData, err := json.Marshal(action)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(
		"POST",
		fmt.Sprintf("http://localhost:%s/game/%s/action", just.Env.Port, b.gameID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+b.token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	return nil
}
