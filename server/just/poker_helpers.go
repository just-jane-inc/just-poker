package just

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
)

type Card struct {
	Rank rune `json:"rank"`
	Suit rune `json:"suit"`
}

type response struct {
	Error      string `json:"error"`
	Evaluation int    `json:"evaluation"`
}

func GetHandScore(cards []Card) (int, error) {
	jsonData, err := json.Marshal(cards)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(Env.PokerEvalURL, "application/json", bytes.NewReader(jsonData))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var r response
	if err = json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return 0, err
	}

	if r.Error != "" {
		return 0, errors.New(r.Error)
	}

	return r.Evaluation, nil
}
