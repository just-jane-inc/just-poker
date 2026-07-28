package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

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
func OnEvalHand(w http.ResponseWriter, r *http.Request) {}

// OnJoinGameRequest godoc
// @Summary      Join a Game
// @Description  Join an open game
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path int true "ID of the Game to join"
// @Success      200 {object} just.ResponseMessage[any]
// @Router       /game/{game_id}/player [post]
func OnJoinGameRequest(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debug("join game request received")
	var userid string
	var username string
	var err error
	if userid, username, err = just.GetAuthorizedUser(r); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	g, ok := CurrentGames[gameID]
	if !ok {
		just.NotFound("game not found", int(just.GameNotFound)).WriteJSONResponse(w)
		return
	}

	if err = g.TryJoinGame(username, userid); err != nil {
		var pokerErr *just.PokerError
		if errors.As(err, &pokerErr) {
			just.BadRequest(pokerErr.Message, int(pokerErr.Code)).WriteJSONResponse(w)
		} else {
			just.Logger.Errorf("encountered error when processing join game request: %v", err)
			just.BadRequest("unknown error occurred", int(just.Unknown)).WriteJSONResponse(w)
		}
	} else {
		just.Logger.Debugf("[%s] joined game [%s]", userid, gameID)
		just.OK("table_joined", struct{}{})
	}
}

// OnGetCurrentGameState godoc
// @Summary      Game State
// @Description  gets the current state of the game from the perspective of the requesting user
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path int true "ID of the Game to get the state of"
// @Success      200 {object} just.ResponseMessage[GameDTO]
// @Router       /game/{game_id}/state [get]
func OnGetCurrentGameState(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debug("getting game state")
	gameIDString := r.PathValue("game_id")
	g, ok := CurrentGames[gameIDString]
	if !ok {
		just.NotFound("game id does not exist", 0).WriteJSONResponse(w)
		return
	}

	just.OK("game_state", g.AsDTO()).WriteJSONResponse(w)
}

// OnCreateGame godoc
// @Summary      Create Game
// @Description  creates a new game from a configuration file
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param request body NewGameConfigDTO true "an object defining configuration information for the new game"
// @Success      200 {object} just.ResponseMessage[string] "game created - game id as string"
// @Router       /game [post]
func OnCreateGame(w http.ResponseWriter, r *http.Request) {
	userID, username, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("request received to create new game for [%s]", username)

	var config NewGameConfigDTO
	if err := json.NewDecoder(r.Body).Decode(&config); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	g, err := CreateGameFromConfig(config)
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
// @Param game_id path int true "ID of the Game exchange chips in"
// @Success      200 {object} just.ResponseMessage[any]
// @Router       /game/{game_id}/chip/exchange [post]
func OnExchangeChips(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debugf("receieved request to exchange chips")

	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.BadRequest(
			err.Error(),
			int(just.UserNotFound),
		).WriteJSONResponse(w)
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
	}

	just.OK("chips_exchanged", struct{}{}).WriteJSONResponse(w)
}

// OnStartGame godoc
// @Summary      Start Game
// @Description  starts a game from a created game lobby, closing it to joins and starting play
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path int true "the id of the game to start"
// @Success      200 {object} just.ResponseMessage[any]
// @Router       /game/{game_id}/started [post]
func OnStartGame(w http.ResponseWriter, r *http.Request) {
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("start game request received from [%s]", userID)
	gameIDString := r.PathValue("game_id")
	game, ok := CurrentGames[gameIDString]
	if !ok {
		just.NotFound("game id does not exist", 0).WriteJSONResponse(w)
		return
	}

	err = game.TryStartGame()
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	go game.ProccessPlayerActions(make(chan any))

	just.OK("game_started", struct{}{}).WriteJSONResponse(w)
	just.Logger.Debugf("started game with id [%s]", game.id)
}

// OnPlayerAction godoc
// @Summary      Player Action
// @Description  post the action preformed by a player
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path int true "the id of the game"
// @Param request body PlayerActionDTO true "the action the player is preforming"
// @Success      200 {object} just.ResponseMessage[any]
// @Router       /game/{game_id}/action [post]
func OnPlayerAction(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debug("player action received")
	var playerAction PlayerActionDTO
	if err := json.NewDecoder(r.Body).Decode(&playerAction); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	g, ok := CurrentGames[gameID]
	if !ok {
		just.NotFound(fmt.Sprintf("could not find game with id %s", gameID), 0).WriteJSONResponse(w)
		return
	}

	err := g.TryPlayerAction(playerAction)
	if err != nil {
		log.Println(err.Error())
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	just.OK("action_accepted", struct{}{}).WriteJSONResponse(w)
}
