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
	for _, c := range g.playerConnections {
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
			playerConnections: make(map[string]*PlayerUpdateConnection),
		}

		h.Games[gameID] = hub
	}

	playerConnection, ok := hub.playerConnections[playerID]
	if ok {
		playerConnection.SignalExit("connection already exists")
		delete(hub.playerConnections, playerID)
	}

	usertype, _ := GetUserType(playerID)

	p := &PlayerUpdateConnection{
		GameID:         gameID,
		PlayerID:       playerID,
		MessageChannel: make(chan WebsocketMessage[any], 10),
		Exit:           make(chan any),
		UserType:       usertype,
	}

	hub.playerConnections[playerID] = p
	return p
}

type GameUpdateHub struct {
	GameID            string
	playerConnections map[string]*PlayerUpdateConnection
	connectionsLock   sync.RWMutex
}

func (hub *GameUpdateHub) GetConnectionForPlayer(id string) (*PlayerUpdateConnection, bool) {
	hub.connectionsLock.RLock()
	defer hub.connectionsLock.RUnlock()
	conn, ok := hub.playerConnections[id]
	return conn, ok
}

func (hub *GameUpdateHub) CloseConnection(id string) bool {
	hub.connectionsLock.Lock()
	defer hub.connectionsLock.Unlock()
	_, ok := hub.playerConnections[id]
	if !ok {
		return false
	}

	delete(hub.playerConnections, id)
	return true
}

type PlayerUpdateConnection struct {
	GameID         string
	PlayerID       string
	MessageChannel chan WebsocketMessage[any]
	Exit           chan any
	conn           *websocket.Conn
	MsgIDCounter   int
	UserType       UserType
	exitLock       sync.Mutex
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

func HandleWebSocket(w http.ResponseWriter, r *http.Request, gameID string, playerID string, data any) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		Logger.Errorf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	Logger.Infof("Client connected: %s", conn.RemoteAddr())

	playerConn := UpdateHub.AddPlayerToHub(gameID, playerID)
	defer playerConn.SignalExit("defered from HandleWebSocket closure")
	defer func() {
		gameHub, ok := UpdateHub.Games[gameID]
		if !ok {
			return
		}

		gameHub.connectionsLock.Lock()
		defer gameHub.connectionsLock.Unlock()
		delete(gameHub.playerConnections, playerID)

		if len(gameHub.playerConnections) == 0 {
			// we should be able to delete this hub now righht?
		}
	}()

	playerConn.conn = conn
	playerConn.MessageChannel <- WebsocketMessage[any]{
		EventType: "welcome",
		Data:      data,
	}

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

func (p *PlayerUpdateConnection) SignalExit(reason string) {
	// this lock is taken but never released - only one exit signal
	// should ever reach the channel
	if !p.exitLock.TryLock() {
		return
	}

	select {
	case p.Exit <- true:
		Logger.Debugf("exit signal sent for player [%s] in game [%s] with reason [%s]", p.PlayerID, p.GameID, reason)
		close(p.Exit)
	default:
	}
}

func (p *PlayerUpdateConnection) handleMessages() {
	ticker := time.NewTicker(time.Duration(5) * time.Second)
	defer Logger.Debugf("exiting handle message for [%s]", p.PlayerID)
	defer ticker.Stop()
	defer p.conn.Close()
	keepRunning := true
	for keepRunning {
		select {
		case <-p.Exit:
			Logger.Infof("received exit signal for player [%s] connection, closing", p.PlayerID)
			keepRunning = false
		case msg := <-p.MessageChannel:
			Logger.Debugf("sending message [%d] to player [%s] ws connection", p.MsgIDCounter, p.PlayerID)
			msg.ID = p.MsgIDCounter
			msg.TimeSent = time.Now()
			bytes, err := json.Marshal(msg)
			if err != nil {
				Logger.Errorf("error marshalling message [%d] in websocket handler: %v", p.MsgIDCounter, err)
				continue
			}

			if err = p.conn.WriteMessage(1, bytes); err != nil {
				Logger.Errorf("error writing message [%d] to connection for player [%s]", p.MsgIDCounter, p.PlayerID)
			}

			p.MsgIDCounter += 1
		case <-ticker.C:
			_ = p.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := p.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				p.SignalExit("write deadline exceeded")
				return
			}
		}
	}

	Logger.Infof("closing connection for [%s] in game [%s]", p.PlayerID, p.GameID)
	close(p.MessageChannel)
	for msg := range p.MessageChannel {
		Logger.Debugf("sending message [%d] to player [%s] ws connection", p.MsgIDCounter, p.PlayerID)
		msg.ID = p.MsgIDCounter
		msg.TimeSent = time.Now()
		bytes, err := json.Marshal(msg)
		if err != nil {
			Logger.Errorf("error marshalling message [%d] in websocket handler: %v", p.MsgIDCounter, err)
			continue
		}

		if err = p.conn.WriteMessage(1, bytes); err != nil {
			Logger.Errorf("error writing message [%d] to connection for player [%s]", p.MsgIDCounter, p.PlayerID)
		}

		p.MsgIDCounter += 1
	}
}
