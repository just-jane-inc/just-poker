package game

import (
	"encoding/json"
	"net/http"

	"github.com/just-jane-inc/just-poker/server/game"
	"github.com/just-jane-inc/just-poker/server/just"
)

const (
	game_status_paused GameStatus = "paused"
	game_status_normal GameStatus = "normal"
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
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	userType, err := just.GetUserType(userID)
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	if userType != "admin" && userType != "game_master" {
		// TODO: forbidden and admonish people for doing this
		just.BadRequest("invalid", 0).WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	g, ok := game.CurrentGames[gameID]
	if !ok {
		just.NotFound("game not found", int(just.GameNotFound)).WriteJSONResponse(w)
		return
	}

	var dto GameStatusDTO
	if err = json.NewDecoder(r.Body).Decode(&dto); err != nil {
		just.BadRequest("invalid input structure", 0).WriteJSONResponse(w)
		return
	}

	switch dto.Status {
	case game_status_paused:
		err = g.PauseGameExecution()
		if err != nil {
			just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
			return
		}

	case game_status_normal:
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
	userID, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	userType, err := just.GetUserType(userID)
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	if userType != "admin" && userType != "game_master" {
		// TODO: forbidden and admonish people for doing this
		just.BadRequest("invalid", 0).WriteJSONResponse(w)
		return
	}

	gameID := r.PathValue("game_id")
	g, ok := game.CurrentGames[gameID]
	if !ok {
		just.NotFound("game not found", int(just.GameNotFound)).WriteJSONResponse(w)
		return
	}

	var dto game.TableDTO
	if err = json.NewDecoder(r.Body).Decode(&dto); err != nil {
		just.BadRequest("invalid input structure", 0).WriteJSONResponse(w)
		return
	}

	if err = g.OverWriteTable(dto); err != nil {
		just.BadRequest(err.Error(), int(just.Unknown)).WriteJSONResponse(w)
		return
	}

	just.OK("table_overwrite", "success").WriteJSONResponse(w)
}
