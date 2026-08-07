// Package game provides you know, the game stuff
// behold, game, packaged for your pleasure
package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"strconv"
	"sync"
	"time"

	"github.com/just-jane-inc/just-poker/server/just"
)

type JoinGameError error

func (g *game) LogGameState(msg string) {
	s, err := json.Marshal(g.AsDTO())
	if err != nil {
		just.Logger.Errorf("encountered error marshaling game: %v", err)
		return
	}

	just.Logger.Debugf("%s \n %s", msg, s)
}

type game struct {
	startedAt           *time.Time
	id                  string
	joinGameLock        sync.Mutex
	isPaused            bool
	config              NewGameConfigDTO
	table               *table
	pauseGameSemaphor   chan any
	playerActionChannel chan *playerAction
	gameEnded           bool
}

func (g *GameDTO) MaskCards(userID string) {
	for _, p := range g.Table.Players {
		if len(p.Hole) == 2 && p.UserID != userID {
			p.Hole[0] = CardDTO{'x', 'x'}
			p.Hole[1] = CardDTO{'x', 'x'}
		}
	}
}

func (g *GameDTO) DeepCopy() GameDTO {
	serialized, _ := json.Marshal(g)
	var dto GameDTO
	_ = json.Unmarshal(serialized, &dto)
	return dto
}

func (g *game) AsDTO() GameDTO {
	return GameDTO{
		StartedAt: g.startedAt,
		ID:        g.id,
		Config:    g.config,
		Table:     g.table.AsDTO(),
	}
}

func (g *game) OverWriteTable(dto TableDTO) error {
	select {
	case g.pauseGameSemaphor <- struct{}{}:
		g.isPaused = true
	default:
	}

	g.table = dto.AsTable()
	return nil
}

func createGameFromConfig(config NewGameConfigDTO) (*game, *just.PokerError) {
	just.Logger.Debugf("%v", config)
	if config.PlayerCount < 2 {
		return nil, just.NewPokerError("player count must be greater then 2 for a poker game", just.InvalidGameConfiguration)
	}

	if config.PlayerCount > 8 {
		return nil, just.NewPokerError("no more then 8 players can poker at once", just.InvalidGameConfiguration)
	}

	if config.SmallBlind >= config.BigBlind {
		return nil, just.NewPokerError("small blind must be greater then the big blind", just.InvalidGameConfiguration)
	}

	if config.SmallBlind <= 0 || config.BigBlind <= 0 {
		return nil, just.NewPokerError("blinds must be greater then 0", just.InvalidGameConfiguration)
	}

	if config.StartingChips.Sum() <= config.BigBlind {
		return nil, just.NewPokerError("insufficient starting chips to poker", just.InvalidGameConfiguration)
	}

	g := &game{
		joinGameLock:        sync.Mutex{},
		config:              config,
		table:               &table{},
		pauseGameSemaphor:   make(chan any),
		playerActionChannel: make(chan *playerAction, 5),
	}

	g.table.currentTurnChannel = nil
	g.table.players = make([]*player, 0)
	g.table.deck = &deck{}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		return nil, just.NewPokerError(err.Error(), just.Unknown)
	}
	defer conn.Release()

	var id int
	stmt := `insert into just_poker_game (starting_chips, player_count) values ($1, $2) RETURNING game_id`
	err = conn.QueryRow(ctx, stmt, config.StartingChips, config.PlayerCount).Scan(&id)
	if err != nil {
		return nil, just.NewPokerError(err.Error(), just.Unknown)
	}

	g.id = strconv.Itoa(id)
	g.table.gameID = g.id
	return g, nil
}

// TODO: should this just be called and return error? do I need
// a response channel really?
func (g *game) TryJoinGame(username, userID string) error {
	just.Logger.Debugf("%s trying to join game %s", username, g.id)

	if g.startedAt != nil {
		return just.NewPokerError("game already started", just.GameAlreadyStarted)
	}

	if g.table == nil {
		return just.NewPokerError("table does not yet exist", just.Unknown)
	}

	if len(g.table.players) >= g.config.PlayerCount {
		return just.NewPokerError("the table has reached its maximum number of players", just.GameIsFull)
	}

	// check for duplicate player
	for _, p := range g.table.players {
		if p.UserID == userID {
			return just.NewPokerError("player already joined table", just.PlayerAlreadyJoined)
		}
	}

	usertype, err := just.GetUserType(userID)
	if err != nil {
		return just.NewPokerError("could not fetch user type", just.Unknown)
	}

	p := &player{
		UserID:      userID,
		UserType:    usertype,
		DisplayName: username,
	}

	p.chips = make(stack)
	maps.Copy(p.chips, g.config.StartingChips.asStack())

	g.table.players = append(g.table.players, p)
	return nil
}

func (g *game) TryStartGame() error {
	if g.startedAt != nil {
		return fmt.Errorf("game already started")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

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

	err = g.table.nextHand(g.config.BigBlind, g.config.SmallBlind)
	if err != nil {
		return err
	}

	t := time.Now()
	g.startedAt = &t
	return nil
}

func (g *game) PauseGameExecution() error {
	select {
	case g.pauseGameSemaphor <- struct{}{}:
		g.isPaused = true
	default:
		return errors.New("game is already paused")
	}

	return nil
}

func (g *game) ResumeGameExecution() error {
	select {
	case <-g.pauseGameSemaphor:
		g.isPaused = false
	default:
		return errors.New("game is already paused")
	}

	return nil
}
