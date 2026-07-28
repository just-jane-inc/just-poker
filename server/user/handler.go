package user

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/just-jane-inc/just-poker/server/just"
)

type ApiKey struct {
	UserID string `json:"user_id"`
	KeyID  string `json:"key_id"`
	Token  string `json:"token"`
}

// OnCreateUser godoc
// @Summary      Creates a user
// @Description  How does this differ from summary...
// @Tags         User
// @Accept       json
// @Produce      json
// @Param        request    body     UserDTO  true  "user"
// @Success      200  {object}  just.ResponseMessage[ApiKey]
// @Failure      400  {object}  just.ResponseMessage[just.ErrorDTO]
// @Failure      401  {object}  just.ResponseMessage[just.ErrorDTO]
// @Failure      403  {object}  just.ResponseMessage[just.ErrorDTO]
// @Failure      404  {object}  just.ResponseMessage[just.ErrorDTO]
// @Failure      500  {object}  just.ResponseMessage[just.ErrorDTO]
// @Router       /user [post]
func OnCreateUser(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debugf("create user request received...")
	var dto UserDTO
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
	if err = conn.QueryRow(ctx, stmt, dto.DisplayName, dto.TwitchID).Scan(&userID); err != nil {
		panic(err)
	}

	just.Logger.Debugf("user [%s] created", userID)

	keyID, token, err := just.CreateApiKey(conn, userID, dto.DisplayName)
	if err != nil {
		panic(err)
	}

	just.OK("new_user", ApiKey{KeyID: keyID, Token: token, UserID: userID}).WriteJSONResponse(w)
}

// OnDeleteMe godoc
// @Summary delete requesting user
// @Description  deletes the user associated to the token used to authenticate this request
// @Tags         User
// @Accept       json
// @Produce      json
// @Success      200  {object}  just.ResponseMessage[any]
// @Failure      400  {object}  just.ResponseMessage[just.ErrorDTO]
// @Failure      401  {object}  just.ResponseMessage[just.ErrorDTO]
// @Failure      403  {object}  just.ResponseMessage[just.ErrorDTO]
// @Failure      404  {object}  just.ResponseMessage[just.ErrorDTO]
// @Failure      500  {object}  just.ResponseMessage[just.ErrorDTO]
// @Security BearerAuth
// @Router       /user/me [delete]
func OnDeleteMe(w http.ResponseWriter, r *http.Request) {
	userIDStr, _, err := just.GetAuthorizedUser(r)
	if err != nil {
		just.MissingToken().WriteJSONResponse(w)
		return
	}

	ctx := context.Background()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		panic(err)
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		just.Logger.Errorf("invalid user id returned from GetAuthorizedUser: [%s]", userIDStr)
		just.InternalError(fmt.Sprintf("user id [%s] does not parse to an int, invalid id", userIDStr)).WriteJSONResponse(w)
		return
	}

	stmt := `delete from poker_users where id=$1`
	_, err = conn.Exec(ctx, stmt, userID)
	if err != nil {
		just.NotFound("user not found", int(just.UserNotFound)).WriteJSONResponse(w)
		return
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
	dto.UserID = userID
	stmt := `select display_name, user_type, twitch_id from just_poker_user where user_id = $1`
	err = conn.QueryRow(context.Background(), stmt, userID).Scan(&dto.DisplayName, &dto.UserType, &dto.TwitchID)
	if err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("user [%s] fetched succesfully", dto.DisplayName)
	just.OK("user", dto).WriteJSONResponse(w)
}

func OnGetUsers(w http.ResponseWriter, r *http.Request) {
	twitchID := r.PathValue("twitch_id")

	ctx := context.Background()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		panic(err)
	}

	stmt := `select username, id from poker_users where twitch_user=$1`
	rows, err := conn.Query(ctx, stmt, twitchID)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	users := make([]UserDTO, 0)
	for rows.Next() {
		var user UserDTO
		err := rows.Scan(&user.DisplayName, &user.UserID)
		if err != nil {
			panic(err)
		}

		users = append(users, user)
	}

	just.OK("get_users", users).WriteJSONResponse(w)
}

func OnRenewKey(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	ctx := context.Background()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		// TODO: dont panic
		panic(err)
	}

	var username string
	stmt := `select username from poker_users where id=$1`
	if err = conn.QueryRow(ctx, stmt, userID).Scan(&username); err != nil {
		// TODO: dont panic
		panic(err)
	}

	if err = just.DeleteApiKey(conn, userID); err != nil {
		panic(err)
	}

	keyID, token, err := just.CreateApiKey(conn, userID, username)
	just.OK("new_key", ApiKey{KeyID: keyID, Token: token, UserID: userID}).WriteJSONResponse(w)
}

/*
*
* Internal Use Endpoints
*
 */

func OnDeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	ctx := context.Background()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		panic(err)
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		just.Logger.Errorf("invalid user id returned from GetAuthorizedUser: [%s]", userIDStr)
		just.InternalError(fmt.Sprintf("user id [%s] does not parse to an int, invalid id", userIDStr)).WriteJSONResponse(w)
		return
	}

	stmt := `delete from poker_users where id=$1`
	_, err = conn.Exec(ctx, stmt, userID)
	if err != nil {
		just.NotFound("user not found", int(just.UserNotFound)).WriteJSONResponse(w)
		return
	}

	just.OK("user_deleted", struct{}{}).WriteJSONResponse(w)
}
