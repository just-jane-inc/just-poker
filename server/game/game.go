// Package game provides you know, the game stuff
// behold, game, packaged for your pleasure
package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/just-jane-inc/just-poker/server/just"
)

var CurrentGames = CreateCurrentGamesCache()

type game struct {
	startedAt     *time.Time
	endedAt       *time.Time
	id            string
	isPaused      bool
	config        NewGameConfigDTO
	table         *table
	gameStateLock sync.Mutex
	// TODO: is this dead code?
	pauseGameSemaphor   chan any
	playerActionChannel chan *playerAction
	denominations       map[int]any

	// id of the user who created this game
	createdBy string
}

// GetCurrentGames provides an array of the [ActiveGameDTO] for currently active games
func (c *CurrentGamesCache) GetCurrentGames() []ActiveGameDTO {
	c.currentGameMutex.RLock()
	defer c.currentGameMutex.RUnlock()
	activeGameDTOs := make([]ActiveGameDTO, 0)
	for id, g := range c.games {
		if g.table == nil {
			continue
		}

		playerIDs := make([]string, len(g.table.players))
		for i, p := range g.table.players {
			playerIDs[i] = p.UserID
		}

		obj := ActiveGameDTO{
			ID:      id,
			Players: playerIDs,
		}

		activeGameDTOs = append(activeGameDTOs, obj)
	}

	return activeGameDTOs
}

// InsertGame adds a game into the map of games
func (c *CurrentGamesCache) InsertGame(g *game) error {
	// TODO: is Lock/Unlock taking the write lock?
	c.currentGameMutex.Lock()
	defer c.currentGameMutex.Unlock()

	_, ok := c.games[g.id]
	if ok {
		return errors.New("game with provided ID already exists")
	}

	c.games[g.id] = g
	return nil
}

// RemoveGame removes a game by id from the cache of current games
func (c *CurrentGamesCache) RemoveGame(id string) (*game, error) {
	c.currentGameMutex.Lock()
	defer c.currentGameMutex.Unlock()

	g, ok := c.games[id]
	if !ok {
		return nil, fmt.Errorf("game with id %s does not exists", id)
	}

	delete(c.games, id)
	return g, nil
}

// GetGame gets a game from the cache and returns it
func (c *CurrentGamesCache) GetGame(id string) (*game, bool) {
	c.currentGameMutex.RLock()
	defer c.currentGameMutex.RUnlock()

	g, ok := c.games[id]
	return g, ok
}

// CreateCurrentGamesCache creates the cache
func CreateCurrentGamesCache() *CurrentGamesCache {
	return &CurrentGamesCache{
		games: make(map[string]*game),
	}
}

// CurrentGamesCache contains a map from string to game as well as a mutex used
// to synchronize access to the cache
type CurrentGamesCache struct {
	games            map[string]*game
	currentGameMutex sync.RWMutex
}

// MaskCards alters a dto so that cards are masked for users that should not see them
//
// TODO: should this return the modified copy rather then modify in place?
func (g *GameDTO) MaskCards(userID string) {
	for _, p := range g.Table.Players {
		if len(p.Hole) == 2 && p.UserID != userID {
			p.Hole[0] = CardDTO{'x', 'x'}
			p.Hole[1] = CardDTO{'x', 'x'}
		}
	}
}

// DeepCopy creates a copy of a game DTO
func (g *GameDTO) DeepCopy() GameDTO {
	serialized, _ := json.Marshal(g)
	var dto GameDTO
	_ = json.Unmarshal(serialized, &dto)
	return dto
}

// AsDTO gets the [GameDTO] representation of the game model
func (g *game) AsDTO() GameDTO {
	return GameDTO{
		StartedAt: g.startedAt,
		EndedAt:   g.endedAt,
		ID:        g.id,
		Config:    g.config,
		Table:     g.table.AsDTO(),
	}
}

// OverWriteTable overrides the table in a game
func (g *game) OverWriteTable(dto TableDTO) error {
	select {
	case g.pauseGameSemaphor <- struct{}{}:
		g.isPaused = true
	default:
	}

	g.table = dto.AsTable()
	return nil
}

// createGameFromConfig creates a game object from a configuration file
func createGameFromConfig(config NewGameConfigDTO) (*game, *just.PokerError) {
	just.Logger.Debugf("%v", config)
	if config.PlayerCount < 2 {
		return nil, just.NewPokerError("player count must be greater then 2 for a poker game", just.InvalidGameConfiguration)
	}

	if config.PlayerCount > 8 {
		return nil, just.NewPokerError("no more then 8 players can poker at once", just.InvalidGameConfiguration)
	}

	if err := config.StartingChips.validate(); err != nil {
		return nil, err
	}

	if config.StartingChips.Sum() <= config.BigBlind {
		return nil, just.NewPokerError("insufficient starting chips to poker", just.InvalidGameConfiguration)
	}

	if config.SmallBlind >= config.BigBlind {
		return nil, just.NewPokerError("small blind must be greater then the big blind", just.InvalidGameConfiguration)
	}

	if config.SmallBlind <= 0 || config.BigBlind <= 0 {
		return nil, just.NewPokerError("blinds must be greater then 0", just.InvalidGameConfiguration)
	}

	if len(config.ChipDenominations) == 0 {
		return nil, just.NewPokerError("no denominations made available for chips", just.InvalidGameConfiguration)
	}

	// this is the canosa validation algorithm
	var values []int
	copy(values, config.ChipDenominations)
	slices.Sort(values)

	candidateHighestCommonDenominator := config.ChipDenominations[0]
	if candidateHighestCommonDenominator <= 0 {
		return nil, just.NewPokerError("the lowest value denomination in collection must be greater then zero", just.InvalidGameConfiguration)
	}

	for _, denomination := range config.ChipDenominations {
		if denomination%candidateHighestCommonDenominator != 0 {
			return nil, just.NewPokerError("The lowest value chip in the set of denominations is not the highest common denominator of the set. i don't know what that value is, i will leave it as an excercise for canosa", just.InvalidGameConfiguration)
		}
	}

	if config.SmallBlind%candidateHighestCommonDenominator != 0 {
		return nil, just.NewPokerError("the small blind must be evenly divisible by the lowest chip in the denomination set", just.InvalidGameConfiguration)
	}

	if config.BigBlind%candidateHighestCommonDenominator != 0 {
		return nil, just.NewPokerError("the big blind must be evenly divisible by the lowest chip in the denomination set", just.InvalidGameConfiguration)
	}

	denominations := make(map[int]any)
	for _, d := range config.ChipDenominations {
		denominations[d] = struct{}{}
	}

	for d := range config.StartingChips {
		denomination, err := strconv.Atoi(d)
		if err != nil {
			return nil, just.NewPokerError("provided denomination [%s] from StartingChips does not parse to an integer", just.InvalidGameConfiguration)
		}

		_, ok := denominations[denomination]
		if !ok {
			return nil, just.NewPokerError("provided starting chip denomination [%s] from StartingChips is not available in ChipDenominations", just.InvalidGameConfiguration)
		}
	}

	g := &game{
		gameStateLock:       sync.Mutex{},
		config:              config,
		table:               &table{},
		pauseGameSemaphor:   make(chan any),
		playerActionChannel: make(chan *playerAction, 5),
		denominations:       denominations,
	}

	g.table.currentTurnChannel = nil
	g.table.players = make([]*player, 0)
	g.table.deck = &deck{}
	g.table.denominations = make([]int, 0)
	for denomination := range denominations {
		g.table.denominations = append(g.table.denominations, denomination)
	}

	sort.Slice(g.table.denominations, func(i, j int) bool {
		return g.table.denominations[i] > g.table.denominations[j]
	})

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
func (g *game) TryJoinGame(user *just.AuthorizedUser) error {
	just.Logger.Debugf("%s trying to join game %s", user.Name, g.id)

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
		if p.UserID == user.Id {
			return just.NewPokerError("player already joined table", just.PlayerAlreadyJoined)
		}
	}

	p := &player{
		UserID:      user.Id,
		UserType:    user.Type,
		DisplayName: user.Name,
	}

	p.chips = make(stack)
	maps.Copy(p.chips, g.config.StartingChips.asStack())

	g.table.players = append(g.table.players, p)
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

func (g *game) TryStartGame() error {
	if g.startedAt != nil {
		return fmt.Errorf("game already started")
	}

	now := time.Now()
	g.startedAt = &now

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

	if g.config.AutoStartHands {
		err = g.table.nextHand(g.config.BigBlind, g.config.SmallBlind)
		if err != nil {
			return err
		}
	} else {
		g.table.currentRound.currentRoundType = RoundTypeUnset
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
