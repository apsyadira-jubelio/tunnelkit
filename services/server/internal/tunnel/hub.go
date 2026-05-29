package tunnel

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// TunnelHub manages active tunnel connections
type TunnelHub struct {
	mu       sync.RWMutex
	agents   map[uuid.UUID]*AgentConn // tunnel_id -> agent connection
}

type AgentConn struct {
	Conn     *websocket.Conn
	TunnelID uuid.UUID
	UserID   uuid.UUID
	LastPing time.Time
}

func NewTunnelHub() *TunnelHub {
	return &TunnelHub{
		agents: make(map[uuid.UUID]*AgentConn),
	}
}

// WSMessage represents a WebSocket control message
type WSMessage struct {
	Type      string         `json:"type"`
	Version   string         `json:"version,omitempty"`
	TunnelID  string         `json:"tunnel_id,omitempty"`
	StreamID  int            `json:"stream_id,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Code      string         `json:"code,omitempty"`
	Message   string         `json:"message,omitempty"`
}

func (h *TunnelHub) HandleWS(c echo.Context) error {
	userID := c.Get("user_id").(string)
	uid, _ := uuid.Parse(userID)

	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	// Wait for hello message
	_, msg, err := ws.ReadMessage()
	if err != nil {
		return err
	}

	var hello WSMessage
	if err := json.Unmarshal(msg, &hello); err != nil {
		ws.WriteJSON(WSMessage{Type: "error", Code: "INVALID_FORMAT", Message: "invalid message format"})
		return err
	}

	if hello.Type != "hello" {
		ws.WriteJSON(WSMessage{Type: "error", Code: "EXPECTED_HELLO", Message: "expected hello message"})
		return nil
	}

	tunnelID, err := uuid.Parse(hello.TunnelID)
	if err != nil {
		ws.WriteJSON(WSMessage{Type: "error", Code: "INVALID_TUNNEL", Message: "invalid tunnel id"})
		return nil
	}

	// Register agent
	agent := &AgentConn{
		Conn:     ws,
		TunnelID: tunnelID,
		UserID:   uid,
		LastPing: time.Now(),
	}

	h.mu.Lock()
	h.agents[tunnelID] = agent
	h.mu.Unlock()

	// Send hello response
	ws.WriteJSON(WSMessage{Type: "hello", Version: "1.0"})

	// Cleanup on disconnect
	defer func() {
		h.mu.Lock()
		delete(h.agents, tunnelID)
		h.mu.Unlock()
	}()

	// Keep connection alive
	for {
		_, msg, err := ws.ReadMessage()
		if err != nil {
			break
		}

		var cmd WSMessage
		if err := json.Unmarshal(msg, &cmd); err != nil {
			continue
		}

		switch cmd.Type {
		case "ping":
			agent.LastPing = time.Now()
			ws.WriteJSON(WSMessage{Type: "pong"})
		case "response":
			// Handle response from local service
			// Forward to waiting client
		}
	}

	return nil
}

// ForwardToAgent forwards an HTTP request to the tunnel agent
func (h *TunnelHub) ForwardToAgent(tunnelID uuid.UUID, req *http.Request) (*http.Response, error) {
	h.mu.RLock()
	agent, ok := h.agents[tunnelID]
	h.mu.RUnlock()

	if !ok {
		return nil, echo.NewHTTPError(http.StatusBadGateway, "tunnel agent not connected")
	}

	// TODO: Implement request forwarding via yamux multiplexing
	_ = agent
	return nil, nil
}
