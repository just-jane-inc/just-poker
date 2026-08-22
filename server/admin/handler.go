// Package admin for admin only actions
package admin

import (
	"encoding/json"
	"net/http"

	"github.com/just-jane-inc/just-poker/server/game"
	"github.com/just-jane-inc/just-poker/server/just"
)

const (
	GameStatusPaused GameStatus = "paused"
	GameStatusNormal GameStatus = "normal"
)

type (
	GameStatus    string
	GameStatusDTO struct {
		Status GameStatus `json:"status"`
	}
)

// OnUpdateGameStatus godoc
// @Summary      Update Game Status
// @Description  Changes the status of an active game, for use by admins
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param game_id path string true "ID of the Game to update status of"
// @Param status body GameStatusDTO true "the status of the game to set"
// @Success      200
// @Success      400
// @Security BearerAuth
// @Router       /admin/game/{game_id}/status [post]
func OnUpdateGameStatus(w http.ResponseWriter, r *http.Request) {
	user, err := just.GetAuthorizedUser(r)
	if user == nil {
		just.Unauthorized().WriteJSONResponse(w)
		return
	}

	if user.NotType(just.UserTypeAdmin, just.UserTypeGameMaster) {
		// TODO: admonish people for doing this
		just.Forbidden().WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	g, ok := game.CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound("game not found", just.GameNotFound).WriteJSONResponse(w)
		return
	}

	var dto GameStatusDTO
	if err = json.NewDecoder(r.Body).Decode(&dto); err != nil {
		just.BadRequest("invalid input structure", 0).WriteJSONResponse(w)
		return
	}

	switch dto.Status {
	case GameStatusPaused:
		err = g.PauseGameExecution()
		if err != nil {
			just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
			return
		}

	case GameStatusNormal:
		err = g.ResumeGameExecution()
		if err != nil {
			just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
			return
		}
	}

	just.OK("status_updated", "success").WriteJSONResponse(w)
}

// OnUpdateGameTable updates game table
// @Summary      Change Game Table
// @Description  Changes the state of an active game
// @Tags         Admin
// @Accept       json
// @Produce      json
// @Param game_id path string true "ID of the Game to update status of"
// @Param status body game.TableDTO true "the table struct to apply"
// @Success      200
// @Success      400
// @Security BearerAuth
// @Router       /admin/game/{game_id}/table [post]
func OnUpdateGameTable(w http.ResponseWriter, r *http.Request) {
	user, err := just.GetAuthorizedUser(r)
	if user == nil {
		just.Unauthorized().WriteJSONResponse(w)
		return
	}

	if user.NotType(just.UserTypeAdmin, just.UserTypeGameMaster) {
		// TODO: admonish people for doing this
		just.Forbidden().WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	g, ok := game.CurrentGames.GetGame(gameID)
	if !ok {
		just.NotFound("game not found", just.GameNotFound).WriteJSONResponse(w)
		return
	}

	var dto game.TableDTO
	if err = json.NewDecoder(r.Body).Decode(&dto); err != nil {
		just.BadRequest("invalid input structure", 0).WriteJSONResponse(w)
		return
	}

	if err = g.OverWriteTable(dto); err != nil {
		just.BadRequest(err.Error(), just.Unknown).WriteJSONResponse(w)
		return
	}

	just.OK("table_overwrite", "success").WriteJSONResponse(w)
}
