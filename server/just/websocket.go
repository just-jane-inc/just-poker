package just

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WebsocketMessage[T any] struct {
	ID        int       `json:"id"`
	TimeSent  time.Time `json:"time_sent"`
	EventType string    `json:"event_type"`
	Data      T         `json:"data"`
}

var UpdateHub = &ServerUpdateHub{
	Games: make(map[string]*GameUpdateHub),
}

type ServerUpdateHub struct {
	Games      map[string]*GameUpdateHub
	addHubSync sync.Mutex
}

func (h *ServerUpdateHub) GetChannelsForGame(gameID string) []*PlayerUpdateConnection {
	g, ok := h.Games[gameID]
	if !ok {
		return nil
	}

	connections := make([]*PlayerUpdateConnection, 0)
	for _, c := range g.PlayerConnections {
		connections = append(connections, c)
	}

	return connections
}

func (h *ServerUpdateHub) AddPlayerToHub(gameID string, playerID string) *PlayerUpdateConnection {
	h.addHubSync.Lock()
	defer h.addHubSync.Unlock()

	hub, ok := h.Games[gameID]
	if !ok {
		hub = &GameUpdateHub{
			GameID:            gameID,
			PlayerConnections: make(map[string]*PlayerUpdateConnection),
		}

		h.Games[gameID] = hub
	}

	playerConnection, ok := hub.PlayerConnections[playerID]
	if ok {
		playerConnection.Exit <- true
	}

	playerConnection = &PlayerUpdateConnection{
		GameID:         gameID,
		PlayerID:       playerID,
		MessageChannel: make(chan WebsocketMessage[any], 10),
		Exit:           make(chan any),
	}

	hub.PlayerConnections[playerID] = playerConnection
	return playerConnection
}

type GameUpdateHub struct {
	GameID            string
	PlayerConnections map[string]*PlayerUpdateConnection
}

type PlayerUpdateConnection struct {
	GameID         string
	PlayerID       string
	MessageChannel chan WebsocketMessage[any]
	Exit           chan any
	conn           *websocket.Conn
	msgIDCounter   int
}

// upgrader configures the WebSocket upgrade parameters
// ReadBufferSize and WriteBufferSize control memory allocation per connection
// CheckOrigin determines which origins are allowed to connect
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// In production, implement proper origin checking
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func HandleWebSocket(w http.ResponseWriter, r *http.Request, gameID string, playerID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Errorf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	Logger.Infof("Client connected: %s", conn.RemoteAddr())

	playerConn := UpdateHub.AddPlayerToHub(gameID, playerID)
	playerConn.conn = conn
	go playerConn.handleMessages()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err,
				websocket.CloseGoingAway,
				websocket.CloseAbnormalClosure) {
				Logger.Infof("Unexpected close error: %v", err)
			}
			break
		}

		Logger.Debugf("player %s sent message to ws on game %s", playerID, gameID)
	}

	Logger.Infof("client disconnected: %s", conn.RemoteAddr())
}

func (p *PlayerUpdateConnection) handleMessages() {
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-p.Exit:
			Logger.Debugf("received exit signal for player [%s] connection, closing", p.PlayerID)
			p.conn.Close()
		case msg := <-p.MessageChannel:
			Logger.Debugf("sending message [%d] to player [%s] ws connection", p.msgIDCounter, p.PlayerID)
			msg.ID = p.msgIDCounter
			msg.TimeSent = time.Now()
			bytes, err := json.Marshal(msg)
			if err != nil {
				Logger.Errorf("error marshalling message [%d] in websocket handler: %v", p.msgIDCounter, err)
				continue
			}

			if err = p.conn.WriteMessage(1, bytes); err != nil {
				Logger.Errorf("error writing message [%d] to connection for player [%s]", p.msgIDCounter, p.PlayerID)
			}

			p.msgIDCounter += 1
		case <-ticker.C:
			_ = p.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
