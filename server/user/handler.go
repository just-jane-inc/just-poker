package user

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"poker_server/just"
)

// CreateUserRequestHandler godoc
// @Summary      Creates a user
// @Description  How does this differ from summary...
// @Tags         user
// @Accept       json
// @Produce      json
// @Param        User    body     UserDTO  true  "user"
// @Success      200  {object}  UserDTO
// @Router       /user/ [post]
func OnCreateUser(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debugf("create user request received...")
	var dto just.CreateUserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		just.BadRequest(err.Error(), 0)
	}

	ctx := context.Background()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		panic(err)
	}

	var userID string
	stmt := `insert into poker_users (username, twitch_user) values ($1, $2) returning id`
	if err = conn.QueryRow(ctx, stmt, dto.UserName, dto.TwitchID).Scan(&userID); err != nil {
		panic(err)
	}

	just.Logger.Debugf("user [%s] created", userID)

	key, err := just.CreateApiKey(conn, userID, dto.UserName)
	if err != nil {
		panic(err)
	}

	just.OK("new_user", key).WriteJSONResponse(w)
}

func OnDeleteMe(w http.ResponseWriter, r *http.Request) {
	userIDStr, _, err := just.GetAuthorizedUser(r)
	ctx := context.Background()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		panic(err)
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		panic(err)
	}

	stmt := `delete from poker_users where id=$1`
	_, err = conn.Exec(ctx, stmt, userID)
	if err != nil {
		panic(err)
	}

	just.OK("user_deleted", struct{}{}).WriteJSONResponse(w)
}

func OnGetUserRequest(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debug("user creation request received")
	userID := r.PathValue("userID")
	if userID == "" {
		just.BadRequest("userID is empty", 0).WriteJSONResponse(w)
	}

	conn, err := just.DBConnPool.Acquire(context.Background())
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}
	defer conn.Release()

	var dto UserDTO
	dto.ID = userID
	stmt := `select display_name, user_type from just_poker_user where user_id = $1`
	err = conn.QueryRow(context.Background(), stmt, userID).Scan(&dto.DisplayName, &dto.UserType)
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("user [%s] fetched succesfully", dto.DisplayName)
	just.OK("user", dto).WriteJSONResponse(w)
}
