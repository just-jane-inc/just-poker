package game

import (
	"errors"

	"github.com/just-jane-inc/just-poker/server/just"
)

type Listener struct {
	queue   chan just.ResponseMessage[any]
	maxSize int
}

type NewHandState struct {
	Deck       []CardDTO `json:"deck"`
	Button     int       `json:"button_position"`
	BigBlind   int       `json:"big_blind_position"`
	SmallBlind int       `json:"small_blind_position"`
}

type PlayerUpdateMessage struct {
	Position int          `json:"position"`
	Chips    ChipStackDTO `json:"chips"`
	Intent   string       `json:"intent"`
}

func CreateListener() *Listener {
	return &Listener{
		queue:   make(chan just.ResponseMessage[any], 50),
		maxSize: 50,
	}
}

func (l *Listener) Send(msg just.ResponseMessage[any]) error {
	if len(l.queue) >= l.maxSize {
		return errors.New("listener is very behind")
	}

	l.queue <- msg
	return nil
}
