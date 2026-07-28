package game

type PlayerIntent string
type PlayerState string

const (
	player_state_unset    PlayerState = "unset"
	player_state_inactive PlayerState = "inactive"
	player_state_active   PlayerState = "active"
	player_state_folded   PlayerState = "folded"
	player_state_all_in   PlayerState = "all_in"
	player_state_won      PlayerState = "won"
	player_state_out      PlayerState = "out"
)

const (
	player_intent_unset  = "unset"
	player_intent_ante   = "ante"
	player_intent_check  = "check"
	player_intent_call   = "call"
	player_intent_raise  = "raise"
	player_intent_all_in = "all_in"
	player_intent_fold   = "fold"
)

type stack map[int]int

type player struct {
	UserID          string
	DisplayName     string
	UserType        string
	state           PlayerState
	position        int
	pocket          []*card
	chips           stack
	currentBet      stack
	potContribution int
}

func (p player) ToString() string {
	return ""
	// cards := fmt.Sprintf("[%c%c] [%c%c]", p.pocket[0].rank, p.pocket[0].suite, p.pocket[1].rank, p.pocket[1].suite)
	// return fmt.Sprintf("%s: %s %s\n", p.DisplayName, cards, ChipsToString(p.chips))
}

func (s stack) Sum() int {
	total := 0
	for d, c := range s {
		total += (d * c)
	}

	return total
}
