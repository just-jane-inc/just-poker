package just

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"net/http"

	"encoding/base64"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type ApiKey struct {
	UserID string `json:"user_id"`
	KeyID  string `json:"key_id"`
	Text   string `json:"token"`
}

type PokerUserDTO struct {
	UserName string `json:"username"`
	UserID   string `json:"user_id"`
}

func getMAC(secret string, pepper []byte) []byte {
	mac := hmac.New(sha256.New, pepper)
	mac.Write([]byte(secret))
	return mac.Sum(nil)
}

func randomString(count int) string {
	val := make([]byte, count)
	rand.Read(val)
	return base64.RawURLEncoding.EncodeToString(val)
}

type CreateUserDTO struct {
	UserName string `json:"username"`
	TwitchID string `json:"twitch_id"`
	UserType string `json:"user_type"`
}

func OnRenewKey(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	ctx := context.Background()
	conn, err := DBConnPool.Acquire(ctx)
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

	if err = DeleteApiKey(conn, userID); err != nil {
		panic(err)
	}

	key, err := CreateApiKey(conn, userID, username)
	OK("new_key", key).WriteJSONResponse(w)
}

/*
func OnDeleteUser(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	ctx := context.Background()
	conn, err := DBConnPool.Acquire(ctx)
	if err != nil {
		panic(err)
	}

	stmt := `delete from poker_users where id=$1`
	_, err = conn.Exec(ctx, stmt, userID)
	if err != nil {
		panic(err)
	}
}
*/

func OnGetUsers(w http.ResponseWriter, r *http.Request) {
	twitchID := r.PathValue("twitch_id")

	ctx := context.Background()
	conn, err := DBConnPool.Acquire(ctx)
	if err != nil {
		panic(err)
	}

	stmt := `select username, id from poker_users where twitch_user=$1`
	rows, err := conn.Query(ctx, stmt, twitchID)
	if err != nil {
		panic(err)
	}

	defer rows.Close()

	users := make([]PokerUserDTO, 0)
	for rows.Next() {
		var user PokerUserDTO
		err := rows.Scan(&user.UserName, &user.UserID)
		if err != nil {
			panic(err)
		}

		users = append(users, user)
	}

	OK("get_users", users).WriteJSONResponse(w)
}

// Gets the userID and username out of a request
func GetAuthorizedUser(r *http.Request) (string, string, error) {
	token := r.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")
	token = splitToken[1]

	parts := strings.Split(token, ".")

	if len(parts) != 3 || parts[0] != "bahms" {
		return "", "", fmt.Errorf("invalid format")
	}

	if parts[1] == "" || parts[2] == "" {
		return "", "", fmt.Errorf("invalid format")
	}

	keyID, secret := parts[1], parts[2]

	ctx := context.Background()
	conn, err := DBConnPool.Acquire(ctx)
	if err != nil {
		return "", "", err
	}

	stmt := `select user_id, username, mac from poker_api_keys where key_id=$1 and revoked_at is null`

	var userID string
	var username string
	var mac []byte
	err = conn.QueryRow(ctx, stmt, keyID).Scan(&userID, &username, &mac)
	if err != nil {
		Logger.Debugf("key error - does not exist")
		return "", "", err
	}

	m := getMAC(secret, Env.Pepper)
	if !hmac.Equal(m, mac) {
		Logger.Debugf("mac error - not authenticated")
		return "", "", err
	}

	return userID, username, nil
}

func DeleteApiKey(conn *pgxpool.Conn, userID string) error {
	stmt := `delete from poker_api_keys where user_id=$1`
	_, err := conn.Exec(context.Background(), stmt, userID)
	return err
}

func CreateApiKey(conn *pgxpool.Conn, userID, username string) (*ApiKey, error) {
	keyID := randomString(12)
	secret := randomString(32)
	mac := getMAC(secret, Env.Pepper)

	stmt := `insert into poker_api_keys (key_id, user_id, username, mac) values ($1, $2, $3, $4)`
	_, err := conn.Exec(context.Background(), stmt, keyID, userID, username, mac)
	if err != nil {
		return nil, err
	}

	return &ApiKey{
		KeyID:  keyID,
		Text:   fmt.Sprintf("bahms.%s.%s", keyID, secret),
		UserID: userID}, nil
}
