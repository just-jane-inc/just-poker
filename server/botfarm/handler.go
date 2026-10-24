package botfarm

import (
	"fmt"
	"net/http"

	"github.com/just-jane-inc/just-poker/server/game"
	"github.com/just-jane-inc/just-poker/server/just"
)

type BotUser struct {
	User  *just.AuthorizedUser
	Token string
}

var UsersThatWeHave = []BotUser{
	{User: &just.AuthorizedUser{ID: "45", Name: "jane", Type: "bot"}, Token: "bahms.8Guly1SIQICcp_aM.o9SIBWHyJ2J9Yd5jrSZbQaq9H-ZKBtdI1QcJwRCRxHs"},
	{User: &just.AuthorizedUser{ID: "48", Name: "red", Type: "bot"}, Token: "bahms.LqEwOpyhXZ7tZRUf.q_gZf2_hpcfdV3H6gvkEt7cRBb3Lcb-LNoYd-nGYE4o"},
	{User: &just.AuthorizedUser{ID: "49", Name: "wolf", Type: "bot"}, Token: "bahms.hZ60mP01P6B43n3b.kAXQ2HjEyuH9kJWZPFwuj2_Aci2vCumwN23COzaNaAo"},
	{User: &just.AuthorizedUser{ID: "50", Name: "rae", Type: "bot"}, Token: "bahms.E6js0u42WHAi9gJ-.lhs4O4FBs1DCzJMI4DaOfAgZhimRQGSC8YOG35Dt1pU"},
}

// OnCreateTestGame godoc
// @Summary	     Creates a test game
// @Description  Creates a test game with random bot oponents to play against, joins the game with the requested user token and starts it
// @Tags         Testing
// @Produce      json
// @Success      200 {object} just.ResponseMessage[game.GameDTO]
// @Failure      400 {object} just.ResponseMessage[just.ErrorDTO]
// @Security BearerAuth
// @Router       /test/game [get]
func OnCreateTestGame(w http.ResponseWriter, r *http.Request) {
	user, _ := just.GetAuthorizedUser(r)
	if user == nil {
		just.Unauthorized().WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("request received to create test game for user [%s]", user.Name)
	config := game.NewGameConfigDTO{
		PlayerCount:       8,
		StartingChips:     game.ChipStackDTO{25: 20, 50: 10, 100: 5, 500: 1},
		BigBlind:          50,
		SmallBlind:        25,
		AutoStartHands:    true,
		ChipDenominations: []int{25, 50, 100, 500},
	}

	g, pokerError := game.CreateGameFromConfig(config)
	if pokerError != nil {
		just.Logger.Debugf("encountered error creating game from config: %v", pokerError)
		just.BadRequest(pokerError.Message, pokerError.Code).WriteJSONResponse(w)
		return
	}

	if err := game.CurrentGames.InsertGame(g); err != nil {
		just.InternalError(fmt.Sprintf("encountered error inserting game into cache: %v game could not be stored", err)).WriteJSONResponse(w)
		return
	}

	for _, user := range UsersThatWeHave {
		g.TryJoinGame(user.User)
		CreateRandomBot(g.AsDTO().ID, *user.User, user.Token)
	}

	just.OK("game_lobby_created", g.AsDTO().ID).WriteJSONResponse(w)
	just.Logger.Debugf("test game lobby [%s] created for [%s]", g.AsDTO().ID, user.ID)
}
