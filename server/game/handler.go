package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/just-jane-inc/just-poker/server/just"
)

var CurrentGames = make(map[string]*game)

// OnEvalHand godoc
// @Summary      Evaluate a Hand
// @Description  Evaluator? I hardly...
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param hand body []CardDTO true "hand to evaluate, either 5 or 7 cards"
// @Success      200
// @Router       /hand-evaluator/evaluate [post]
func OnEvalHand() {}

// OnJoinGameRequest godoc
// @Summary      Join a Game
// @Description  Join an open game
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path string true "ID of the Game to join"
// @Success      200 {object} just.ResponseMessage[any]
// @Security BearerAuth
// @Router       /game/{game_id}/player [post]
func OnJoinGameRequest(w http.ResponseWriter, r *http.Request) {
	var userid string
	var username string
	var err error
	if userid, username, err = just.GetAuthorizedUser(r); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	just.Logger.Debugf("received request to join game [%s]", gameID)
	g, ok := CurrentGames[gameID]
	if !ok {
		just.NotFound("game not found", int(just.GameNotFound)).WriteJSONResponse(w)
		return
	}

	g.joinGameLock.Lock()
	err = g.TryJoinGame(username, userid)
	g.joinGameLock.Unlock()

	if err != nil {
		var pokerErr *just.PokerError
		if errors.As(err, &pokerErr) {
			just.BadRequest(pokerErr.Message, int(pokerErr.Code)).WriteJSONResponse(w)
		} else {
			just.Logger.Errorf("encountered error when processing join game request: %v", err)
			just.BadRequest("unknown error occurred", int(just.Unknown)).WriteJSONResponse(w)
		}
	} else {
		just.Logger.Debugf("[%s] joined game [%s]", userid, gameID)
		just.OK("table_joined", struct{}{}).WriteJSONResponse(w)
	}
}

// OnGetCurrentGameState godoc
// @Summary      Game State
// @Description  gets the current state of the game from the perspective of the requesting user
// @Tags         Game
// @Produce      json
// @Param game_id path string true "ID of the Game to get the state of"
// @Success      200 {object} just.ResponseMessage[GameDTO]
// @Security BearerAuth
// @Router       /game/{game_id}/state [get]
func OnGetCurrentGameState(w http.ResponseWriter, r *http.Request) {
	userID, _, _ := just.GetAuthorizedUser(r)
	if userID != "" {
		just.Logger.Debugf("getting game state for user with ID %s", userID)
	} else {
		just.Logger.Debug("getting game state")
	}

	gameIDString := r.PathValue("game_id")
	g, ok := CurrentGames[gameIDString]
	if !ok {
		just.NotFound("game id does not exist", 0).WriteJSONResponse(w)
		return
	}

	userType, err := just.GetUserType(userID)
	if err != nil {
		just.InternalError("invalid user").WriteJSONResponse(w)
		return
	}

	dto := g.AsDTO()
	usertype, _ := just.GetUserType(userID)
	switch usertype {
	case just.UserTypeGameMaster, just.UserTypeAdmin:
	default:
		dto.MaskCards(userID)
	}

	if userType != "admin" && userType != "game_master" {
		dto.MaskCards(userID)
	}

	just.OK("game_state", dto).WriteJSONResponse(w)
}

// OnCreateGame godoc
// @Summary      Create Game
// @Description  creates a new game from a configuration file
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param request body NewGameConfigDTO true "an object defining configuration information for the new game"
// @Success      200 {object} just.ResponseMessage[string] "game created - game id as string"
// @Security BearerAuth
// @Router       /game [post]
func OnCreateGame(w http.ResponseWriter, r *http.Request) {
	userID, username, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.Logger.Errorf("encountered error: %v", err)
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("request received to create new game for [%s]", username)

	var config NewGameConfigDTO
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	g, err := createGameFromConfig(config)
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	CurrentGames[g.id] = g
	just.OK("game_lobby_created", g.id).WriteJSONResponse(w)
	just.Logger.Debugf("game lobby [%s] created for [%s]", g.id, userID)
}

// OnExchangeChips godoc
// @Summary      Exchange Chips
// @Description  exchange chips in the players stack with the tables rack
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param request body ChipExchangeDTO true "a specification for the chips to exchange"
// @Param game_id path string true "ID of the Game exchange chips in"
// @Success      200 {object} just.ResponseMessage[any]
// @Security BearerAuth
// @Router       /game/{game_id}/chip/exchange [post]
func OnExchangeChips(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debugf("receieved request to exchange chips")
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.BadRequest(
			err.Error(),
			int(just.UserNotFound),
		).WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	if gameID == "" {
		just.BadRequest(
			"game_id is required in route",
			int(just.GameNotFound),
		).WriteJSONResponse(w)
		return
	}

	g, ok := CurrentGames[gameID]
	if !ok {
		just.BadRequest(
			"game_id not found",
			int(just.GameNotFound),
		).WriteJSONResponse(w)
		return
	}

	var exchange ChipExchangeDTO
	if err := json.NewDecoder(r.Body).Decode(&exchange); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	err = g.ExchangeChips(userID, exchange)
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	just.OK("chips_exchanged", struct{}{}).WriteJSONResponse(w)
}

// OnStartGame godoc
// @Summary      Start Game
// @Description  starts a game from a created game lobby, closing it to joins and starting play
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path string true "the id of the game to start"
// @Success      200 {object} just.ResponseMessage[any]
// @Security BearerAuth
// @Router       /game/{game_id}/started [post]
func OnStartGame(w http.ResponseWriter, r *http.Request) {
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("start game request received from [%s]", userID)
	gameIDString := r.PathValue("game_id")
	g, ok := CurrentGames[gameIDString]
	if !ok {
		just.NotFound("game id does not exist", 0).WriteJSONResponse(w)
		return
	}

	g.joinGameLock.Lock()
	err = g.TryStartGame()
	g.joinGameLock.Unlock()
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	go g.ProccessPlayerActions(make(chan any))

	just.OK("game_started", struct{}{}).WriteJSONResponse(w)
	just.Logger.Debugf("started game with id [%s]", g.id)
}

// OnPlayerAction godoc
// @Summary      Player Action
// @Description  post the action preformed by a player
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path string true "the id of the game"
// @Param request body PlayerActionDTO true "the action the player is preforming"
// @Success      200 {object} just.ResponseMessage[any]
// @Security BearerAuth
// @Router       /game/{game_id}/action [post]
func OnPlayerAction(w http.ResponseWriter, r *http.Request) {
	// TODO: assert that the player producing the action dto is the one in the auth token

	var err error
	var userID string

	userID, _, err = just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	just.Logger.Debug("player action received")
	var playerAction PlayerActionDTO
	if err := json.NewDecoder(r.Body).Decode(&playerAction); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	playerAction.PlayerID = userID
	gameID := r.PathValue("game_id")
	g, ok := CurrentGames[gameID]
	if !ok {
		just.NotFound(fmt.Sprintf("could not find game with id %s", gameID), 0).WriteJSONResponse(w)
		return
	}

	if g.isPaused {
		just.InvalidPlayerActionGameIsPaused().WriteJSONResponse(w)
		return
	}

	err = g.TryPlayerAction(playerAction)
	if err != nil {
		log.Println(err.Error())
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	go handleUpdates(g.AsDTO())
	just.OK("action_accepted", struct{}{}).WriteJSONResponse(w)
}

type ActiveGameDTO struct {
	ID      string   `json:"id"`
	Players []string `json:"player_ids"`
}

// OnGetCurrentActiveGames godoc
// @Summary      Gets Active Games
// @Description  gets all games that are currently being played
// @Tags         Game
// @Produce      json
// @Success      200 {object} []ActiveGameDTO
// @Router       /game [get]
func OnGetCurrentActiveGames(w http.ResponseWriter, _ *http.Request) {
	games := make([]ActiveGameDTO, 0)
	for id, g := range CurrentGames {
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

		games = append(games, obj)
	}

	just.WriteJSONResponse(w, 200, games)
}

// OnRegisterListener _liiiisten_
// @Summary     Register Listener
// @Description  creates a listener that will begin buffering game events that can be queried from an endpoint
// @Tags         Game
// @Param game_id path string true "ID of the Game to listen to"
// @Success      200
// @Router       /game/{game_id}/state/listen [post]
func OnRegisterListener(w http.ResponseWriter, r *http.Request) {
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	just.Logger.Debugf("received request to connect to game state updates from game [%s] from player [%s]", gameID, userID)

	g, ok := CurrentGames[gameID]
	if !ok {
		just.NotFound("game not found", int(just.GameNotFound)).WriteJSONResponse(w)
		return
	}

	conn := just.UpdateHub.AddPlayerToHub(gameID, userID)
	conn.MessageChannel <- just.WebsocketMessage[any]{
		EventType: "welcome",
		Data:      g.AsDTO(),
	}

	just.OK("listener_created", g.AsDTO()).WriteJSONResponse(w)
}

// OnGetNextListenerEvent _liiiisten_
// @Summary     Get Listener
// @Description  creates a listener that will begin buffering game events that can be queried from an endpoint
// @Tags         Game
// @Param game_id path string true "ID of the Game to get events from"
// @Produce      json
// @Success      200 {object} GameDTO
// @Router       /game/{game_id}/state/listen [get]
func OnGetNextListenerEvent(w http.ResponseWriter, r *http.Request) {
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	just.Logger.Debugf("received request to connect to game state updates from game [%s] from player [%s]", gameID, userID)

	_, ok := CurrentGames[gameID]
	if !ok {
		just.NotFound("game not found", int(just.GameNotFound)).WriteJSONResponse(w)
		return
	}

	h, ok := just.UpdateHub.Games[gameID]
	if !ok {
		just.NotFound("no listening hub found for game id", 0).WriteJSONResponse(w)
		return
	}

	conn, ok := h.PlayerConnections[userID]
	if !ok {
		just.NotFound("no listening hub found for user", 0).WriteJSONResponse(w)
		return
	}

	msg := <-conn.MessageChannel
	msg.TimeSent = time.Now()
	msg.ID = conn.MsgIDCounter
	conn.MsgIDCounter += 1
	just.WriteJSONResponse(w, http.StatusOK, msg)
}

// OnCreateGameConnection godoc
// @Summary      Connect Updates
// @Description  gets all game updates
// @Tags         Game
// @Produce      json
// @Param game_id path string true "ID of the Game to get events from"
// @Success      200
// @Router       /game/{game_id}/state/ws [get]
func OnCreateGameConnection(w http.ResponseWriter, r *http.Request) {
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	just.Logger.Debugf("received request to connect to game state updates from game [%s] from player [%s]", gameID, userID)

	g, ok := CurrentGames[gameID]
	if !ok {
		just.NotFound("game not found", int(just.GameNotFound)).WriteJSONResponse(w)
		return
	}

	dto := g.AsDTO()
	usertype, _ := just.GetUserType(userID)
	switch usertype {
	case just.UserTypeGameMaster, just.UserTypeAdmin:
	default:
		dto.MaskCards(userID)
	}

	dto.MaskCards(userID)
	just.HandleWebSocket(w, r, gameID, userID, dto)
}

func handleUpdates(dto GameDTO) {
	err := just.RecordingHub.OnGameUpdate(dto)
	if err != nil {
		just.Logger.Errorf("error updating game state in elastic: %v", err)
	}

	just.Logger.Debugf("sending updates for game with ID [%s]", dto.ID)
	for _, conn := range just.UpdateHub.GetChannelsForGame(dto.ID) {
		just.Logger.Debugf("sending update to player %s", conn.PlayerID)
		dtoNew := dto.DeepCopy()
		dtoNew.MaskCards(conn.PlayerID)
		msg := just.WebsocketMessage[any]{
			Data:      dtoNew,
			EventType: "game_state_update",
		}

		select {
		case conn.MessageChannel <- msg:
		default:
		}

	}

	if dto.EndedAt == nil {
		return
	}

	for _, conn := range just.UpdateHub.GetChannelsForGame(dto.ID) {
		just.Logger.Debugf("sending update to player %s", conn.PlayerID)
		dtoNew := dto.DeepCopy()
		dtoNew.MaskCards(conn.PlayerID)
		msg := just.WebsocketMessage[any]{
			Data:      dtoNew,
			EventType: "game_state_update",
		}

		select {
		case conn.MessageChannel <- msg:
		default:
		}

		just.Logger.Debugf("send message to player %s success", conn.PlayerID)
	}

	delete(just.UpdateHub.Games, dto.ID)
}
