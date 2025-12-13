package websocket

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

const (
	// writeWait is the time allowed to write a message to the peer
	writeWait = 10 * time.Second

	// pongWait is the time allowed to read the next pong message from the peer
	pongWait = 60 * time.Second

	// pingPeriod is the period for sending ping messages (must be less than pongWait)
	pingPeriod = 30 * time.Second

	// maxMessageSize is the maximum message size allowed from peer
	maxMessageSize = 4096

	// sendChannelSize is the buffer size for the send channel
	sendChannelSize = 256

	// tokenExpiryWarningThreshold is how many seconds before expiry to warn
	tokenExpiryWarningThreshold = 60
)

// Client represents a WebSocket client connection
type Client struct {
	hub  *Hub
	conn *websocket.Conn
	send chan []byte

	// User info from JWT
	userID uuid.UUID
	email  string
	role   string

	// JWT expiration for token_expiring warnings
	tokenExp time.Time

	// Subscribed channels
	channels map[string]bool
}

// NewClient creates a new WebSocket client
func NewClient(hub *Hub, conn *websocket.Conn, userID uuid.UUID, email, role string, tokenExp time.Time) *Client {
	return &Client{
		hub:      hub,
		conn:     conn,
		send:     make(chan []byte, sendChannelSize),
		userID:   userID,
		email:    email,
		role:     role,
		tokenExp: tokenExp,
		channels: make(map[string]bool),
	}
}

// ReadPump reads messages from the WebSocket connection
func (c *Client) ReadPump() {
	defer func() {
		c.hub.UnregisterClient(c)
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		var msg Message
		err := c.conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error().
					Err(err).
					Str("user_id", c.userID.String()).
					Msg("WebSocket read error")
			}
			break
		}

		c.handleMessage(&msg)
	}
}

// WritePump writes messages to the WebSocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	// Start token expiry monitor
	go c.monitorTokenExpiry()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages from the client
func (c *Client) handleMessage(msg *Message) {
	if msg == nil {
		return
	}

	log.Debug().
		Str("user_id", c.userID.String()).
		Str("message_type", string(msg.Type)).
		Msg("Received message")

	switch msg.Type {
	case MessageTypeSubscribe:
		c.handleSubscribe(msg)

	case MessageTypeUnsubscribe:
		c.handleUnsubscribe(msg)

	case MessageTypePing:
		c.handlePing()

	default:
		c.sendError("invalid_message_type", fmt.Sprintf("Invalid message type: %s", msg.Type))
	}
}

// handleSubscribe processes a subscribe request
func (c *Client) handleSubscribe(msg *Message) {
	var payload SubscribePayload
	if err := msg.UnmarshalPayload(&payload); err != nil {
		c.sendError("invalid_payload", "Invalid subscribe payload")
		return
	}

	if payload.Channel == "" {
		c.sendError("invalid_channel", "Channel is required")
		return
	}

	if err := c.hub.Subscribe(c, payload.Channel); err != nil {
		c.sendError("subscribe_failed", err.Error())
		return
	}

	// Send confirmation
	response, err := NewMessage(MessageTypeSubscribed, &SubscribedPayload{
		Channel: payload.Channel,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create subscribed message")
		return
	}

	_ = c.SendMessage(response)
}

// handleUnsubscribe processes an unsubscribe request
func (c *Client) handleUnsubscribe(msg *Message) {
	var payload UnsubscribePayload
	if err := msg.UnmarshalPayload(&payload); err != nil {
		c.sendError("invalid_payload", "Invalid unsubscribe payload")
		return
	}

	if payload.Channel == "" {
		c.sendError("invalid_channel", "Channel is required")
		return
	}

	c.hub.Unsubscribe(c, payload.Channel)

	// Send confirmation
	response, err := NewMessage(MessageTypeUnsubscribed, &UnsubscribedPayload{
		Channel: payload.Channel,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create unsubscribed message")
		return
	}

	_ = c.SendMessage(response)
}

// handlePing processes a ping request
func (c *Client) handlePing() {
	msg, err := NewMessage(MessageTypePong, nil)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create pong message")
		return
	}

	_ = c.SendMessage(msg)
}

// monitorTokenExpiry monitors JWT expiration and sends warnings
func (c *Client) monitorTokenExpiry() {
	if c.tokenExp.IsZero() {
		return
	}

	// Calculate warning time (60 seconds before expiry)
	warningTime := c.tokenExp.Add(-tokenExpiryWarningThreshold * time.Second)
	now := time.Now()

	// If already past warning time, send immediately
	if now.After(warningTime) {
		c.sendTokenExpiringWarning()
		return
	}

	// Wait until warning time
	timer := time.NewTimer(warningTime.Sub(now))
	defer timer.Stop()

	<-timer.C
	c.sendTokenExpiringWarning()
}

// sendTokenExpiringWarning sends a token expiring warning
func (c *Client) sendTokenExpiringWarning() {
	expiresIn := int(time.Until(c.tokenExp).Seconds())
	if expiresIn < 0 {
		expiresIn = 0
	}

	msg, err := NewMessage(MessageTypeTokenExpiring, &TokenExpiringPayload{
		ExpiresAt: c.tokenExp,
		ExpiresIn: expiresIn,
	})
	if err != nil {
		log.Error().Err(err).Msg("Failed to create token expiring message")
		return
	}

	_ = c.SendMessage(msg)

	log.Info().
		Str("user_id", c.userID.String()).
		Int("expires_in", expiresIn).
		Msg("Token expiring warning sent")
}

// sendError sends an error message to the client
func (c *Client) sendError(code, message string) {
	msg, err := NewErrorMessage(code, message)
	if err != nil {
		log.Error().Err(err).Msg("Failed to create error message")
		return
	}

	_ = c.SendMessage(msg)
}

// SendMessage sends a message to this client
func (c *Client) SendMessage(msg *Message) error {
	if msg == nil {
		return fmt.Errorf("message is required")
	}

	msgBytes, err := msg.Marshal()
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	select {
	case c.send <- msgBytes:
		return nil
	default:
		return fmt.Errorf("send channel full")
	}
}
