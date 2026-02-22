package ws

import (
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog"
)

// Message represents a WebSocket message
type Message struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload"`
	Timestamp string      `json:"timestamp"`
}

// Client represents a WebSocket client connection
type Client struct {
	ID     string
	OrgID  string
	UserID string
	Role   string
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *Hub
}

// Hub maintains active client connections and broadcasts messages
type Hub struct {
	// Registered clients by org_id
	clients map[string]map[*Client]bool

	// Channel for broadcasting messages to an org
	broadcast chan *OrgMessage

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Mutex for clients map
	mu sync.RWMutex

	// Logger
	logger zerolog.Logger
}

// OrgMessage is a message targeted at an organization
type OrgMessage struct {
	OrgID   string
	Message []byte
}

// NewHub creates a new Hub
func NewHub(logger zerolog.Logger) *Hub {
	return &Hub{
		clients:    make(map[string]map[*Client]bool),
		broadcast:  make(chan *OrgMessage, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		logger:     logger,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if h.clients[client.OrgID] == nil {
				h.clients[client.OrgID] = make(map[*Client]bool)
			}
			h.clients[client.OrgID][client] = true
			h.mu.Unlock()
			h.logger.Info().Str("client_id", client.ID).Str("org_id", client.OrgID).Msg("client connected")

		case client := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.clients[client.OrgID]; ok {
				if _, ok := clients[client]; ok {
					delete(clients, client)
					close(client.Send)
					if len(clients) == 0 {
						delete(h.clients, client.OrgID)
					}
				}
			}
			h.mu.Unlock()
			h.logger.Info().Str("client_id", client.ID).Str("org_id", client.OrgID).Msg("client disconnected")

		case msg := <-h.broadcast:
			h.mu.RLock()
			clients := h.clients[msg.OrgID]
			h.mu.RUnlock()

			for client := range clients {
				select {
				case client.Send <- msg.Message:
				default:
					h.mu.Lock()
					close(client.Send)
					delete(h.clients[msg.OrgID], client)
					h.mu.Unlock()
				}
			}
		}
	}
}

// Broadcast sends a message to all clients in an organization
func (h *Hub) Broadcast(orgID string, msgType string, payload interface{}) {
	msg := Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to marshal message")
		return
	}

	h.broadcast <- &OrgMessage{
		OrgID:   orgID,
		Message: data,
	}
}

// ClientCount returns the number of connected clients for an org
func (h *Hub) ClientCount(orgID string) int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients[orgID])
}

// BroadcastToRoles sends a message only to clients with one of the specified roles.
func (h *Hub) BroadcastToRoles(orgID string, roles []string, msgType string, payload interface{}) {
	msg := Message{
		Type:      msgType,
		Payload:   payload,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	data, err := json.Marshal(msg)
	if err != nil {
		h.logger.Error().Err(err).Msg("failed to marshal message")
		return
	}

	roleSet := make(map[string]bool, len(roles))
	for _, r := range roles {
		roleSet[r] = true
	}

	h.mu.RLock()
	clients := h.clients[orgID]
	h.mu.RUnlock()

	for client := range clients {
		if !roleSet[client.Role] {
			continue
		}
		select {
		case client.Send <- data:
		default:
			h.mu.Lock()
			close(client.Send)
			delete(h.clients[orgID], client)
			h.mu.Unlock()
		}
	}
}

// ConnectedOrgIDs returns a list of org IDs that have at least one connected client.
func (h *Hub) ConnectedOrgIDs() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	ids := make([]string, 0, len(h.clients))
	for orgID, clients := range h.clients {
		if len(clients) > 0 {
			ids = append(ids, orgID)
		}
	}
	return ids
}

// Upgrader for WebSocket connections
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins - CORS is handled at the HTTP level
		return true
	},
}

// Handler handles WebSocket upgrade requests
type Handler struct {
	Hub *Hub
}

// ServeWS upgrades HTTP connection to WebSocket
func (h *Handler) ServeWS(w http.ResponseWriter, r *http.Request) {
	// Get org_id, user_id, and role from query params (since WS doesn't use auth middleware easily)
	orgID := r.URL.Query().Get("org_id")
	userID := r.URL.Query().Get("user_id")
	role := r.URL.Query().Get("role")

	if orgID == "" {
		http.Error(w, "org_id required", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.Hub.logger.Error().Err(err).Msg("websocket upgrade failed")
		return
	}

	client := &Client{
		ID:     uuid.New().String(),
		OrgID:  orgID,
		UserID: userID,
		Role:   role,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    h.Hub,
	}

	h.Hub.register <- client

	// Start goroutines for reading and writing
	go client.writePump()
	go client.readPump()
}

// writePump pumps messages from the hub to the WebSocket connection
func (c *Client) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// readPump pumps messages from the WebSocket connection to the hub
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512 * 1024)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.Hub.logger.Error().Err(err).Msg("websocket read error")
			}
			break
		}
	}
}
