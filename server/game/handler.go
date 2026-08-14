package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/just-jane-inc/just-poker/server/just"
)

// OnEvalHand godoc
// @Summary      Evaluate a Hand
// @Description  Evaluator? I hardly...
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param hand body []CardDTO true "hand to evaluate, either 5 or 7 cards"
// @Success      200 {object} just.HandEvaluationDTO
// @Router       /hand-evaluator/evaluate/ [post]
func OnEvalHand(w http.ResponseWriter, r *http.Request) {
	resp, err := http.Post(just.Env.PokerEvalURL, "application/json", r.Body)
	if err != nil {
		just.Logger.Errorf("encountered error getting hand evaluation: %v", err)
		just.InternalError("failed to execute request").WriteJSONResponse(w)
	}

	var dto just.HandEvaluationDTO
	if err := json.NewDecoder(resp.Body).Decode(&dto); err != nil {
		just.Logger.Error("OH NO")
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	just.WriteJSONResponse(w, 200, dto)
}

// OnDeleteGame godoc
// @Summary      Delete a Game
// @Description  Delete a game
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path string true "ID of the Game to delete"
// @Success      200 {object} just.ResponseMessage[any]
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
// @Security BearerAuth
// @Router       /game/{game_id} [delete]
func OnDeleteGame(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("game_id")
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	userType, err := just.GetUserType(userID)
	if err != nil {
		just.Logger.Errorf("usertype not found for user with ID %s", userID)
		just.NotFound("user type not found", just.Unknown).WriteJSONResponse(w)
		return
	}

	if userType != just.UserTypeAdmin {
		just.Unauthorized().WriteJSONResponse(w)
		return
	}

	g, err := CurrentGames.RemoveGame(gameID)
	if err != nil {
		just.NotFound("game not found", just.GameNotFound).WriteJSONResponse(w)
		return
	}

	t := time.Now()
	g.endedAt = &t

	go handleUpdates(g, g.AsDTO(), "game_delete")
	just.OK("success", struct{}{})
}

// OnJoinGameRequest godoc
// @Summary      Join a Game
// @Description  Join an open game
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path string true "ID of the Game to join"
// @Success      200 {object} just.ResponseMessage[any]
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
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
	g, ok := CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound("game not found", just.GameNotFound).WriteJSONResponse(w)
		return
	}

	g.joinGameLock.Lock()
	err = g.TryJoinGame(username, userid)
	g.joinGameLock.Unlock()

	if err != nil {
		var pokerErr *just.PokerError
		if errors.As(err, &pokerErr) {
			just.BadRequest(pokerErr.Message, pokerErr.Code).WriteJSONResponse(w)
		} else {
			just.Logger.Errorf("encountered error when processing join game request: %v", err)
			just.BadRequest("unknown error occurred", just.Unknown).WriteJSONResponse(w)
		}
	} else {
		just.Logger.Debugf("[%s] joined game [%s]", userid, gameID)
		just.OK("table_joined", struct{}{}).WriteJSONResponse(w)
	}
}

// OnStartNextHand godoc
// @Summary      Start Next Hand
// @Description  Starts the next poker hand
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path string true "ID of the Game to join"
// @Param new_hand body NewHandDTO true "a dto containing new hand information"
// @Success      200
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
// @Security BearerAuth
// @Router       /game/{game_id}/hand [post]
func OnStartNextHand(w http.ResponseWriter, r *http.Request) {
	_, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	g, ok := CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound("game id does not exist", 0).WriteJSONResponse(w)
		return
	}

	var dto NewHandDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	if pokerError := g.table.nextHandWithDeck(dto.Deck, g.config.BigBlind, g.config.SmallBlind); pokerError != nil {
		just.Logger.Errorf("encountered error setting next hand: %v", err)
		just.BadRequest(pokerError.Message, pokerError.Code).WriteJSONResponse(w)
		return
	}

	just.Logger.Debug("sending game state updates serially in next hand handler")
	handleUpdates(g, g.AsDTO(), "game_state_update")
	just.OK("hand_started", struct{}{}).WriteJSONResponse(w)
}

// OnGetCurrentGameState godoc
// @Summary      Game State
// @Description  gets the current state of the game from the perspective of the requesting user
// @Tags         Game
// @Produce      json
// @Param game_id path string true "ID of the Game to get the state of"
// @Success      200 {object} just.ResponseMessage[GameDTO]
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
// @Security BearerAuth
// @Router       /game/{game_id}/state [get]
func OnGetCurrentGameState(w http.ResponseWriter, r *http.Request) {
	userID, _, _ := just.GetAuthorizedUser(r)
	if userID != "" {
		just.Logger.Debugf("getting game state for user with ID %s", userID)
	} else {
		just.Logger.Debug("getting game state")
	}

	gameID := r.PathValue("game_id")
	g, ok := CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound("game id does not exist", 0).WriteJSONResponse(w)
		return
	}

	dto := g.AsDTO()
	usertype, _ := just.GetUserType(userID)
	switch usertype {
	case just.UserTypeGameMaster, just.UserTypeAdmin:
	default:
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
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
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

	g, pokerError := createGameFromConfig(config)
	if pokerError != nil {
		just.BadRequest(pokerError.Message, pokerError.Code).WriteJSONResponse(w)
		return
	}

	if err = CurrentGames.InsertGame(g); err != nil {
		just.InternalError(fmt.Sprintf("encountered error inserting game into cache: %v game could not be stored", err)).WriteJSONResponse(w)
		return
	}

	just.OK("game_lobby_created", g.id).WriteJSONResponse(w)
	just.Logger.Debugf("game lobby [%s] created for [%s]", g.id, userID)
}

// OnSetTable godoc
// @Summary      Sets a table
// @Description  creates a game for testing
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path string true "the id of the game to set the table for"
// @Param request body TableDTO true "the game state object to load as a test game"
// @Success      200 {object} just.ResponseMessage[string] "game created - game id as string"
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
// @Security BearerAuth
// @Router       /game/{game_id}/table [post]
func OnSetTable(w http.ResponseWriter, r *http.Request) {
	_, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.Logger.Errorf("encountered error: %v", err)
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")

	var table TableDTO
	if err := json.NewDecoder(r.Body).Decode(&table); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	g, ok := CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound("game not found", just.GameNotFound).WriteJSONResponse(w)
	}

	g.table = table.AsTable()
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
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
// @Security BearerAuth
// @Router       /game/{game_id}/chip/exchange [post]
func OnExchangeChips(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debugf("receieved request to exchange chips")
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.BadRequest(
			err.Error(),
			just.UserNotFound,
		).WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	if gameID == "" {
		just.BadRequest(
			"game_id is required in route",
			just.GameNotFound,
		).WriteJSONResponse(w)
		return
	}

	g, ok := CurrentGames.GetGame(gameID)
	if !ok {
		just.BadRequest(
			"game_id not found",
			just.GameNotFound,
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
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
// @Security BearerAuth
// @Router       /game/{game_id}/started [post]
func OnStartGame(w http.ResponseWriter, r *http.Request) {
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("start game request received from [%s]", userID)
	gameID := r.PathValue("game_id")
	g, ok := CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound("game id does not exist", 0).WriteJSONResponse(w)
		return
	}

	g.joinGameLock.Lock()
	handleUpdates(g, g.AsDTO(), "starting_game")
	err = g.TryStartGame()
	g.joinGameLock.Unlock()
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	handleUpdates(g, g.AsDTO(), "game_state_update")
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
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
// @Security BearerAuth
// @Router       /game/{game_id}/action [post]
func OnPlayerAction(w http.ResponseWriter, r *http.Request) {
	var err error
	var userID string

	userID, _, err = just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	g, ok := CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound(fmt.Sprintf("could not find game with id %s", gameID), 0).WriteJSONResponse(w)
		return
	}

	var playerAction PlayerActionDTO
	if err := json.NewDecoder(r.Body).Decode(&playerAction); err != nil {
		just.BadRequest(err.Error(), just.MalformedRequestBody).WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("player action request received from [%s] for [%s]", userID, gameID)

	playerAction.PlayerID = userID

	if g.isPaused {
		just.InvalidPlayerActionGameIsPaused().WriteJSONResponse(w)
		return
	}

	err = g.TryPlayerAction(playerAction)
	if err != nil {
		var pokerError *just.PokerError
		if errors.As(err, &pokerError) {
			just.BadRequest(pokerError.Message, pokerError.Code).WriteJSONResponse(w)
			return
		}

		just.BadRequest(err.Error(), just.Unknown).WriteJSONResponse(w)
		return
	}

	// we send the update here to snapshot the state how it is in this moment, further checks will alter the tate
	// and would require a new update afterwards.
	handleUpdates(g, g.AsDTO(), "game_state_update")

	if g.IsGameOver() {
		// update the game state and send a new update
		g.table.OnGameOver()
		handleUpdates(g, g.AsDTO(), "game_state_update")
	} else if g.table.currentRound.currentRoundType == RoundTypeCompleted {
		// if the game is configured for auto-starting the next hand we do so now
		if g.config.AutoStartHands {
			// we need to block here until this completes, the handleUpdates snapshot should include the completed state here.
			g.table.nextHand(g.config.BigBlind, g.config.SmallBlind)
			handleUpdates(g, g.AsDTO(), "game_state_update")
		}
	}

	just.OK("action_accepted", struct{}{}).WriteJSONResponse(w)
}

// OnGetCurrentActiveGames godoc
// @Summary      Gets Active Games
// @Description  gets all games that are currently being played
// @Tags         Game
// @Produce      json
// @Success      200 {object} []ActiveGameDTO
// @Router       /game [get]
func OnGetCurrentActiveGames(w http.ResponseWriter, _ *http.Request) {
	just.WriteJSONResponse(w, 200, CurrentGames.GetCurrentGames())
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

	g, ok := CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound("game not found", just.GameNotFound).WriteJSONResponse(w)
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
	h, ok := just.UpdateHub.Games[gameID]
	if !ok {
		just.NotFound("no listening hub found for game id", 0).WriteJSONResponse(w)
		return
	}

	conn, ok := h.GetConnectionForPlayer(userID)
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

	g, ok := CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound("game not found", just.GameNotFound).WriteJSONResponse(w)
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

func handleUpdates(g *game, dto GameDTO, eventType string) {
	just.Logger.Debugf("sending [%s] update for game with ID [%s]", eventType, dto.ID)

	for _, conn := range just.UpdateHub.GetChannelsForGame(dto.ID) {
		just.Logger.Debugf("sending update to player %s", conn.PlayerID)
		dtoNew := dto.DeepCopy()

		if conn.UserType != just.UserTypeAdmin && conn.UserType != just.UserTypeGameMaster {
			dtoNew.MaskCards(conn.PlayerID)
		}

		msg := just.WebsocketMessage[any]{
			Data:      dtoNew,
			EventType: eventType,
		}

		select {
		case conn.MessageChannel <- msg:
		default:
		}
	}

	if g.endedAt != nil {
		just.Logger.Infof("game has ended sending game_over and closing connections game=[%s]", g.id)
		for _, conn := range just.UpdateHub.GetChannelsForGame(g.id) {
			conn.MessageChannel <- just.WebsocketMessage[any]{
				EventType: "game_over",
				Data:      dto.Table.Players,
			}

			conn.SignalExit("game ended")
		}

		delete(CurrentGames.games, g.id)
	}

	if err := just.RecordingHub.OnGameUpdate(dto); err != nil {
		just.Logger.Errorf("error updating game state in elastic: %v", err)
	}
}
