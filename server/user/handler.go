package user

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/just-jane-inc/just-poker/server/just"
)

type ApiKey struct {
	UserID string `json:"user_id"`
	KeyID  string `json:"key_id"`
	Token  string `json:"token"`
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

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		panic(err)
	}
	defer conn.Release()

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

// OnCreateUser Create a user
// Internal only, unauthenticated
func OnCreateUser(w http.ResponseWriter, r *http.Request) {
	just.Logger.Debugf("create user request received...")
	var dto UserDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		just.BadRequest(err.Error(), 0).WriteJSONResponse(w)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		just.Logger.Errorf("error acquiring db conn: %v", err)
		just.InternalError("internal server error").WriteJSONResponse(w)
		return
	}
	defer conn.Release()

	var userID string
	stmt := `insert into poker_users (username, twitch_user) values ($1, $2) returning id`
	if err = conn.QueryRow(ctx, stmt, dto.DisplayName, dto.TwitchID).Scan(&userID); err != nil {
		just.Logger.Errorf("error getting user id after insert: %v", err)
		just.InternalError("internal server error").WriteJSONResponse(w)
		return
	}

	just.Logger.Debugf("user [%s] created", userID)

	keyID, token, err := just.CreateAPIKey(conn, ctx, userID, dto.DisplayName)
	if err != nil {
		just.Logger.Errorf("error acquiring api key: %v", err)
		just.InternalError("internal server error").WriteJSONResponse(w)
		return
	}

	just.OK("new_user", ApiKey{KeyID: keyID, Token: token, UserID: userID}).WriteJSONResponse(w)
}

// OnGetUsers Get all users
// Internal only, unauthenticated

func OnGetUsers(w http.ResponseWriter, r *http.Request) {
	twitchID := r.PathValue("twitch_id")

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		just.Logger.Errorf("encountered error acquiring db conn: %v", err)
		just.InternalError("error with database").WriteJSONResponse(w)
		return
	}
	defer conn.Release()

	stmt := `select username, id from poker_users where twitch_user=$1`
	rows, err := conn.Query(ctx, stmt, twitchID)
	just.Logger.Debugf("querying rows")
	if err != nil {
		just.Logger.Errorf("encountered error acquiring db conn: %v", err)
		just.InternalError("error with database").WriteJSONResponse(w)
		return
	}

	defer rows.Close()
	just.Logger.Debugf("row query success")

	users := make([]UserDTO, 0)
	for rows.Next() {
		var user UserDTO
		err := rows.Scan(&user.DisplayName, &user.UserID)
		if err != nil {
			just.Logger.Errorf("error scanning row when getting users: %v", err)
			just.InternalError("internal server error").WriteJSONResponse(w)
			return
		}

		users = append(users, user)
	}

	just.OK("get_users", users).WriteJSONResponse(w)
}

// OnRenewKey Renew user key
// Internal only, unauthenticated

func OnRenewKey(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		just.Logger.Errorf("error acquiring db con: %v", err)
		just.InternalError("internal server errror").WriteJSONResponse(w)
		return
	}
	defer conn.Release()

	var username string
	stmt := `select username from poker_users where id=$1`
	if err = conn.QueryRow(ctx, stmt, userID).Scan(&username); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			just.NotFound("user not found", int(just.UserNotFound)).WriteJSONResponse(w)
		} else {
			just.Logger.Errorf("error getting row: %v", err)
			just.InternalError("internal server errror").WriteJSONResponse(w)
		}

		return
	}

	if err = just.DeleteAPIKey(conn, ctx, userID); err != nil {
		just.Logger.Errorf("error acquiring db con: %v", err)
		just.InternalError("internal server errror").WriteJSONResponse(w)
		return
	}

	keyID, token, err := just.CreateAPIKey(conn, ctx, userID, username)
	if err != nil {
		just.Logger.Errorf("error acquiring db con: %v", err)
		just.InternalError("internal server errror").WriteJSONResponse(w)
		return
	}

	just.OK("new_key", ApiKey{KeyID: keyID, Token: token, UserID: userID}).WriteJSONResponse(w)
}

// OnDeleteUser Delete user
// Internal only, unauthenticated

func OnDeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := just.DBConnPool.Acquire(ctx)
	if err != nil {
		just.Logger.Errorf("error acquiring db con: %v", err)
		just.InternalError("internal server errror").WriteJSONResponse(w)
		return
	}
	defer conn.Release()

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
