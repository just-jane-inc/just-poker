package game

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"poker_server/just"
)

var (
	CurrentGames = make(map[string]*game)
)

// JoinGameRequest: godoc
// @Summary      Join a Game
// @Description  Join an open game
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path int true "ID of the Game to join"
// @Success      200
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
		just.NotFound("game not found", 0).WriteJSONResponse(w)
		return
	}

	// TODO: does this really need to be async?
	ch := make(chan just.HttpResponse)
	go g.TryJoinGame(username, userid, ch)
	response := <-ch
	response.WriteJSONResponse(w)
	just.Logger.Debugf("[%s] joined game [%s]", userid, gameID)
}

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

// @Summary      Create Game
// @Description  creates a new game from a configuration file
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param request body NewGameConfigDTO true "an object defining configuration information for the new game"
// @Success      200 {object} just.ResponseMessage[string] "game created - game id as string"
// @Router       /game [post]
func OnNewGame(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debug("request received to create new game")

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
	just.Logger.Debugf("game lobby created for game [%s]", g.id)
}

// @Summary      Evaluate Hand
// @Description  evaluates a poker hand with 5 or 7 cards
// @Tags         Utility
// @Accept       json
// @Produce      json
// @Param request body []CardDTO true "an array of cards to evaluate"
// @Success      200 {object} just.ResponseMessage[int] "an integer ranking of the provided hand with 1 being the best possible hand"
// @Router       /utility/eval-hand [post]
func OnEvaluateHand(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debug("request received to evaaluate hand")

	var hand []CardDTO
	if err := json.NewDecoder(r.Body).Decode(&hand); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	handstr := make([]string, len(hand))
	for i, card := range hand {
		handstr[i] = fmt.Sprintf("%c%c", card.Rank, card.Suit)
	}

	score, err := just.GetHandScore(handstr...)
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	just.OK("hand_evaluation", score).WriteJSONResponse(w)
}

// @Summary      Exchange Chips
// @Description  exchange chips in the players stack with the tables rack
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param request body ChipExchangeDTO true "a specification for the chips to exchange"
// @Param game_id path int true "ID of the Game exchange chips in"
// @Success      200
// @Router       /game/{game_id}/rack/exchange [post]
func OnExchangeChips(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debugf("receieved request to exchange chips")

	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.BadRequest(
			err.Error(),
			int(just.UserNotFound)).WriteJSONResponse(w)
	}

	gameID := r.PathValue("game_id")
	if gameID == "" {
		just.BadRequest(
			"game_id is required in route",
			int(just.GameNotFound)).WriteJSONResponse(w)
		return
	}

	g, ok := CurrentGames[gameID]
	if !ok {
		just.BadRequest(
			"game_id not found",
			int(just.GameNotFound)).WriteJSONResponse(w)
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
}

// @Summary      Start Game
// @Description  start a game
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path int true "the id of the game to start"
// @Success      200 {object} just.ResponseMessage[any]
// @Router       /game [post]
func OnStartGame(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debug("start game request received")
	gameIDString := r.PathValue("game_id")
	game, ok := CurrentGames[gameIDString]
	if !ok {
		just.NotFound("game id does not exist", 0).WriteJSONResponse(w)
		return
	}

	err := game.TryStartGame()
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	go game.ProccessPlayerActions(make(chan any))

	just.OK("game_started", struct{}{}).WriteJSONResponse(w)
	just.Logger.Debugf("started game with id [%s]", game.id)
}

// @Summary      Player Action
// @Description  post the action preformed by a player
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path int true "the id of the game"
// @Param request body PlayerActionDTO true "the action the player is preforming"
// @Success      200 {object} just.ResponseMessage[any]
// @Router       /game [post]
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
		just.NotFound(fmt.Sprintf("could not find game with id %s", gameID), 0)
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
