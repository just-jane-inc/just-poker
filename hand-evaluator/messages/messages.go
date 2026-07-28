package messages

type Card struct {
	Rank rune `json:"rank"`
	Suit rune `json:"suit"`
}

type Response struct {
	Error      string `json:"error"`
	Evaluation int    `json:"evaluation"`
}
