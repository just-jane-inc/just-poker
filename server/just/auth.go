package just

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserType string

func (ut UserType) IsValid() bool {
	switch ut {
	case UserTypeBot, UserTypeHuman, UserTypeAdmin, UserTypeDealer, UserTypeGameMaster:
		return true
	}

	return false
}

const (
	UserTypeBot        UserType = "bot"
	UserTypeHuman      UserType = "human"
	UserTypeAdmin      UserType = "admin"
	UserTypeDealer     UserType = "dealer"
	UserTypeGameMaster UserType = "game_master"
)

type AuthorizedUser struct {
	ID   string
	Name string
	Type UserType
}

func (ut AuthorizedUser) IsType(userType ...UserType) bool {
	return slices.Contains(userType, ut.Type)
}

func (ut AuthorizedUser) NotType(userType ...UserType) bool {
	return !ut.IsType(userType...)
}

func getMAC(secret string, pepper []byte) []byte {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(secret))
	return mac.Sum(nil)
}

func randomString(count int) string {
	val := make([]byte, count)
	_, _ = rand.Read(val)
	return base64.RawURLEncoding.EncodeToString(val)
}

func GetAuthorizedUser(r *http.Request) (*AuthorizedUser, error) {
	token := r.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")

	if len(splitToken) != 2 {
		return nil, errors.New("invalid format")
	}

	token = splitToken[1]
	parts := strings.Split(token, ".")

	if len(parts) != 3 || parts[0] != "bahms" {
		return nil, errors.New("invalid format")
	}

	if parts[1] == "" || parts[2] == "" {
		return nil, errors.New("invalid format")
	}

	keyID, secret := parts[1], parts[2]

	ctx := context.Background()
	conn, err := DBConnPool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	stmt := `select user_id, pu.username, pu.user_type, mac from (poker_api_keys right join poker_users pu on pu.id = poker_api_keys.user_id) where key_id=$1 and revoked_at is null`

	var userID string
	var username string
	var userType UserType
	var mac []byte
	err = conn.QueryRow(ctx, stmt, keyID).Scan(&userID, &username, &userType, &mac)
	if err != nil {
		Logger.Infof("error getting user %v", err)
		return nil, err
	}

	m := getMAC(secret, Env.Pepper)
	if !hmac.Equal(m, mac) {
		Logger.Debugf("mac error - not authenticated")
		return nil, err
	}

	return &AuthorizedUser{
		ID:   userID,
		Name: username,
		Type: userType,
	}, nil
}

func DeleteAPIKey(conn *pgxpool.Conn, ctx context.Context, userID string) error {
	stmt := `delete from poker_api_keys where user_id=$1`
	_, err := conn.Exec(ctx, stmt, userID)
	return err
}

func CreateAPIKey(conn *pgxpool.Conn, ctx context.Context, userID, username string) (keyID string, token string, err error) {
	keyID = randomString(12)
	secret := randomString(32)
	mac := getMAC(secret, Env.Pepper)

	stmt := `insert into poker_api_keys (key_id, user_id, username, mac) values ($1, $2, $3, $4)`
	_, err = conn.Exec(ctx, stmt, keyID, userID, username, mac)
	if err != nil {
		return "", "", err
	}

	return keyID, fmt.Sprintf("bahms.%s.%s", keyID, secret), nil
}

func DeleteUser(userIDStr string, w http.ResponseWriter) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	conn, err := DBConnPool.Acquire(ctx)
	if err != nil {
		Logger.Errorf("error acquiring db con: %v", err)
		InternalError("internal server errror").WriteJSONResponse(w)
		return
	}
	defer conn.Release()

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		Logger.Errorf("invalid user id returned from GetAuthorizedUser: [%s]", userIDStr)
		InternalError(fmt.Sprintf("user id [%s] does not parse to an int, invalid id", userIDStr)).WriteJSONResponse(w)
		return
	}

	stmt := `delete from poker_users where id=$1`
	_, err = conn.Exec(ctx, stmt, userID)
	if err != nil {
		NotFound("user not found", UserNotFound).WriteJSONResponse(w)
		return
	}

	OK("user_deleted", struct{}{}).WriteJSONResponse(w)
}
