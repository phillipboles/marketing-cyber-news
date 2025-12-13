package websocket

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

const (
	// DefaultMaxConnectionsPerUser is the maximum connections allowed per user
	DefaultMaxConnectionsPerUser = 5

	// DefaultMaxChannelsPerClient is the maximum channels a client can subscribe to
	DefaultMaxChannelsPerClient = 50
)

// Hub maintains active clients and handles broadcasting
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Client lookup by user ID
	userClients map[uuid.UUID]map[*Client]bool

	// Channel subscriptions (channel -> clients)
	channels map[string]map[*Client]bool

	// Register/unregister channels
	register   chan *Client
	unregister chan *Client

	// Broadcast to specific channel
	broadcast chan *BroadcastMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Connection limits
	maxConnectionsPerUser int
	maxChannelsPerClient  int
}

// BroadcastMessage represents a message to broadcast to a channel
type BroadcastMessage struct {
	Channel string
	Message *Message
}

// HubConfig holds configuration for the hub
type HubConfig struct {
	MaxConnectionsPerUser int
	MaxChannelsPerClient  int
}

// NewHub creates a new WebSocket hub
func NewHub(cfg *HubConfig) *Hub {
	if cfg == nil {
		cfg = &HubConfig{
			MaxConnectionsPerUser: DefaultMaxConnectionsPerUser,
			MaxChannelsPerClient:  DefaultMaxChannelsPerClient,
		}
	}

	if cfg.MaxConnectionsPerUser <= 0 {
		cfg.MaxConnectionsPerUser = DefaultMaxConnectionsPerUser
	}

	if cfg.MaxChannelsPerClient <= 0 {
		cfg.MaxChannelsPerClient = DefaultMaxChannelsPerClient
	}

	return &Hub{
		clients:               make(map[*Client]bool),
		userClients:           make(map[uuid.UUID]map[*Client]bool),
		channels:              make(map[string]map[*Client]bool),
		register:              make(chan *Client),
		unregister:            make(chan *Client),
		broadcast:             make(chan *BroadcastMessage, 256),
		maxConnectionsPerUser: cfg.MaxConnectionsPerUser,
		maxChannelsPerClient:  cfg.MaxChannelsPerClient,
	}
}

// Run starts the hub's main loop
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.handleRegister(client)

		case client := <-h.unregister:
			h.handleUnregister(client)

		case broadcast := <-h.broadcast:
			h.handleBroadcast(broadcast)
		}
	}
}

// handleRegister adds a client to the hub
func (h *Hub) handleRegister(client *Client) {
	if client == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check connection limit for user
	userConns := h.userClients[client.userID]
	if len(userConns) >= h.maxConnectionsPerUser {
		log.Warn().
			Str("user_id", client.userID.String()).
			Int("count", len(userConns)).
			Msg("Max connections per user reached")

		msg, err := NewErrorMessage("max_connections", "Maximum connections per user reached")
		if err == nil {
			_ = client.SendMessage(msg)
		}

		go client.conn.Close()
		return
	}

	// Register client
	h.clients[client] = true

	// Add to user's clients
	if h.userClients[client.userID] == nil {
		h.userClients[client.userID] = make(map[*Client]bool)
	}
	h.userClients[client.userID][client] = true

	log.Info().
		Str("user_id", client.userID.String()).
		Str("email", client.email).
		Int("total_connections", len(h.clients)).
		Msg("Client registered")

	// Send connected message
	msg, err := NewMessage(MessageTypeConnected, &ConnectedPayload{
		Message: "Connected to ACI WebSocket",
	})
	if err == nil {
		_ = client.SendMessage(msg)
	}
}

// handleUnregister removes a client from the hub
func (h *Hub) handleUnregister(client *Client) {
	if client == nil {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	if !h.clients[client] {
		return
	}

	// Remove from all channels
	for channel := range client.channels {
		h.unsubscribeNoLock(client, channel)
	}

	// Remove from user's clients
	if userConns := h.userClients[client.userID]; userConns != nil {
		delete(userConns, client)
		if len(userConns) == 0 {
			delete(h.userClients, client.userID)
		}
	}

	// Remove from clients
	delete(h.clients, client)

	// Close send channel
	close(client.send)

	log.Info().
		Str("user_id", client.userID.String()).
		Str("email", client.email).
		Int("total_connections", len(h.clients)).
		Msg("Client unregistered")
}

// handleBroadcast sends a message to all clients in a channel
func (h *Hub) handleBroadcast(bm *BroadcastMessage) {
	if bm == nil || bm.Message == nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.channels[bm.Channel]
	if len(clients) == 0 {
		return
	}

	msgBytes, err := bm.Message.Marshal()
	if err != nil {
		log.Error().
			Err(err).
			Str("channel", bm.Channel).
			Msg("Failed to marshal broadcast message")
		return
	}

	count := 0
	for client := range clients {
		select {
		case client.send <- msgBytes:
			count++
		default:
			// Client send channel is full, skip
			log.Warn().
				Str("user_id", client.userID.String()).
				Str("channel", bm.Channel).
				Msg("Client send channel full, skipping message")
		}
	}

	log.Debug().
		Str("channel", bm.Channel).
		Str("message_type", string(bm.Message.Type)).
		Int("recipients", count).
		Msg("Message broadcast")
}

// RegisterClient adds a client to the hub
func (h *Hub) RegisterClient(client *Client) error {
	if client == nil {
		return fmt.Errorf("client is required")
	}

	h.register <- client
	return nil
}

// UnregisterClient removes a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	if client == nil {
		return
	}

	h.unregister <- client
}

// Subscribe adds a client to a channel
func (h *Hub) Subscribe(client *Client, channel string) error {
	if client == nil {
		return fmt.Errorf("client is required")
	}

	if channel == "" {
		return fmt.Errorf("channel is required")
	}

	if !IsValidChannel(channel) {
		return fmt.Errorf("invalid channel: %s", channel)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// Check if client is registered
	if !h.clients[client] {
		return fmt.Errorf("client not registered")
	}

	// Check channel limit
	if len(client.channels) >= h.maxChannelsPerClient {
		return fmt.Errorf("max channels per client reached")
	}

	// Check if already subscribed
	if client.channels[channel] {
		return nil
	}

	// Add to channel
	if h.channels[channel] == nil {
		h.channels[channel] = make(map[*Client]bool)
	}
	h.channels[channel][client] = true

	// Add to client's channels
	client.channels[channel] = true

	log.Debug().
		Str("user_id", client.userID.String()).
		Str("channel", channel).
		Int("total_channels", len(client.channels)).
		Msg("Client subscribed to channel")

	return nil
}

// Unsubscribe removes a client from a channel
func (h *Hub) Unsubscribe(client *Client, channel string) {
	if client == nil || channel == "" {
		return
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.unsubscribeNoLock(client, channel)
}

// unsubscribeNoLock removes a client from a channel without locking
func (h *Hub) unsubscribeNoLock(client *Client, channel string) {
	// Remove from channel
	if clients := h.channels[channel]; clients != nil {
		delete(clients, client)
		if len(clients) == 0 {
			delete(h.channels, channel)
		}
	}

	// Remove from client's channels
	delete(client.channels, channel)

	log.Debug().
		Str("user_id", client.userID.String()).
		Str("channel", channel).
		Msg("Client unsubscribed from channel")
}

// Broadcast sends a message to all clients in a channel
func (h *Hub) Broadcast(channel string, msg *Message) {
	if channel == "" || msg == nil {
		return
	}

	h.broadcast <- &BroadcastMessage{
		Channel: channel,
		Message: msg,
	}
}

// BroadcastToUser sends a message to all clients of a specific user
func (h *Hub) BroadcastToUser(userID uuid.UUID, msg *Message) {
	if userID == uuid.Nil || msg == nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := h.userClients[userID]
	if len(clients) == 0 {
		return
	}

	msgBytes, err := msg.Marshal()
	if err != nil {
		log.Error().
			Err(err).
			Str("user_id", userID.String()).
			Msg("Failed to marshal user message")
		return
	}

	count := 0
	for client := range clients {
		select {
		case client.send <- msgBytes:
			count++
		default:
			log.Warn().
				Str("user_id", userID.String()).
				Msg("Client send channel full, skipping message")
		}
	}

	log.Debug().
		Str("user_id", userID.String()).
		Str("message_type", string(msg.Type)).
		Int("recipients", count).
		Msg("Message sent to user")
}

// GetConnectionCount returns the number of connections for a user
func (h *Hub) GetConnectionCount(userID uuid.UUID) int {
	if userID == uuid.Nil {
		return 0
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	return len(h.userClients[userID])
}

// GetStats returns hub statistics
func (h *Hub) GetStats() map[string]interface{} {
	h.mu.RLock()
	defer h.mu.RUnlock()

	return map[string]interface{}{
		"total_clients":  len(h.clients),
		"total_users":    len(h.userClients),
		"total_channels": len(h.channels),
	}
}
