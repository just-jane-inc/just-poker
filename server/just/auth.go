package just

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

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

func GetUserType(userID string) (string, error) {
	ctx := context.Background()
	conn, err := DBConnPool.Acquire(ctx)
	if err != nil {
		return "", err
	}
	defer conn.Release()

	stmt := `select user_type from poker_users where id=$1`
	var userType string
	err = conn.QueryRow(ctx, stmt, userID).Scan(&userType)
	if err != nil {
		Logger.Debugf("userID not found when querying poker_users for user_type: %s", userID)
		return "", err
	}

	return userType, nil
}

func GetAuthorizedUser(r *http.Request) (string, string, error) {
	token := r.Header.Get("Authorization")
	splitToken := strings.Split(token, "Bearer ")

	if len(splitToken) != 2 {
		return "", "", NewPokerError("invalid token format", Unknown)
	}

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
	defer conn.Release()

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
