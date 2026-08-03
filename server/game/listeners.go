package game

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
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

// OnListenerRequest godoc
// @Summary      Get Next Event
// @Description  gets next event from listener queue
// @Tags         Game
// @Produce      json
// @Param listener_id path string true "the id of the game"
// @Param game_id path string true "the id of the listener"
// @Success      200 {object} just.ResponseMessage[any]
// @Router       /game/{game_id}/listener/{listener_id} [get]
func OnListenerRequest(w http.ResponseWriter, r *http.Request) {
	listenerID := r.PathValue("listener_id")
	gameID := r.PathValue("game_id")

	g, ok := CurrentGames[gameID]
	if !ok {
		just.NotFound(
			"game_id not found",
			int(just.GameNotFound),
		).WriteJSONResponse(w)
		return
	}

	l, ok := g.listeners[listenerID]
	if !ok {
		just.NotFound(
			"listener_id not found",
			int(just.GameNotFound),
		).WriteJSONResponse(w)
		return
	}

	resp := <-l.queue
	just.OK(resp.Type, resp.Data).WriteJSONResponse(w)
}

// OnCreateListener godoc
// @Summary      Creates Listener
// @Description  creates a new listener
// @Tags         Game
// @Accept       json
// @Produce      json
// @Param game_id path string true "the id of the game to create a listener in"
// @Success      200 {object} just.ResponseMessage[string] "the id of the listener"
// @Router       /game/{game_id}/listener [post]
func OnCreateListener(w http.ResponseWriter, r *http.Request) {
	gameID := r.PathValue("game_id")

	g, ok := CurrentGames[gameID]
	if !ok {
		just.NotFound(
			"game_id not found",
			int(just.GameNotFound),
		).WriteJSONResponse(w)
		return
	}

	id := uuid.NewString()
	g.listeners[id] = CreateListener()

	just.OK("listener.created", id).WriteJSONResponse(w)
}
