package websocket

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// MessageType represents WebSocket message types
type MessageType string

const (
	// Client -> Server
	MessageTypeSubscribe   MessageType = "subscribe"
	MessageTypeUnsubscribe MessageType = "unsubscribe"
	MessageTypePing        MessageType = "ping"

	// Server -> Client
	MessageTypeConnected      MessageType = "connected"
	MessageTypeSubscribed     MessageType = "subscribed"
	MessageTypeUnsubscribed   MessageType = "unsubscribed"
	MessageTypePong           MessageType = "pong"
	MessageTypeTokenExpiring  MessageType = "token_expiring"
	MessageTypeError          MessageType = "error"
	MessageTypeArticleNew     MessageType = "article.new"
	MessageTypeArticleUpdated MessageType = "article.updated"
	MessageTypeAlertMatch     MessageType = "alert.match"
)

// Message is the envelope for all WebSocket messages
type Message struct {
	Type      MessageType     `json:"type"`
	ID        string          `json:"id,omitempty"`
	Timestamp time.Time       `json:"timestamp"`
	Payload   json.RawMessage `json:"payload,omitempty"`
}

// SubscribePayload represents a subscribe message payload
type SubscribePayload struct {
	Channel string `json:"channel"`
}

// UnsubscribePayload represents an unsubscribe message payload
type UnsubscribePayload struct {
	Channel string `json:"channel"`
}

// ConnectedPayload represents a connected message payload
type ConnectedPayload struct {
	Message string `json:"message"`
}

// SubscribedPayload represents a subscribed message payload
type SubscribedPayload struct {
	Channel string `json:"channel"`
}

// UnsubscribedPayload represents an unsubscribed message payload
type UnsubscribedPayload struct {
	Channel string `json:"channel"`
}

// ErrorPayload represents an error message payload
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// TokenExpiringPayload represents a token expiring warning payload
type TokenExpiringPayload struct {
	ExpiresAt time.Time `json:"expires_at"`
	ExpiresIn int       `json:"expires_in"` // Seconds until expiration
}

// NewMessage creates a new message with timestamp and ID
func NewMessage(msgType MessageType, payload interface{}) (*Message, error) {
	var payloadBytes json.RawMessage
	var err error

	if payload != nil {
		payloadBytes, err = json.Marshal(payload)
		if err != nil {
			return nil, err
		}
	}

	return &Message{
		Type:      msgType,
		ID:        uuid.New().String(),
		Timestamp: time.Now(),
		Payload:   payloadBytes,
	}, nil
}

// NewErrorMessage creates an error message
func NewErrorMessage(code, message string) (*Message, error) {
	return NewMessage(MessageTypeError, &ErrorPayload{
		Code:    code,
		Message: message,
	})
}

// Marshal serializes the message to JSON
func (m *Message) Marshal() ([]byte, error) {
	return json.Marshal(m)
}

// UnmarshalPayload unmarshals the payload into the target type
func (m *Message) UnmarshalPayload(target interface{}) error {
	if m.Payload == nil {
		return nil
	}
	return json.Unmarshal(m.Payload, target)
}

// Channel names and helpers

const (
	// Channel prefixes
	ChannelPrefixArticles  = "articles:"
	ChannelPrefixAlerts    = "alerts:"
	ChannelPrefixSystem    = "system"

	// Predefined channels
	ChannelArticlesAll      = "articles:all"
	ChannelArticlesCritical = "articles:critical"
	ChannelArticlesHigh     = "articles:high"
	ChannelAlertsUser       = "alerts:user"
	ChannelSystem           = "system"
)

// BuildCategoryChannel builds a channel name for a specific category
func BuildCategoryChannel(categorySlug string) string {
	return ChannelPrefixArticles + "category:" + categorySlug
}

// BuildVendorChannel builds a channel name for a specific vendor
func BuildVendorChannel(vendorName string) string {
	return ChannelPrefixArticles + "vendor:" + vendorName
}

// IsValidChannel validates a channel name
func IsValidChannel(channel string) bool {
	if channel == "" {
		return false
	}

	validChannels := map[string]bool{
		ChannelArticlesAll:      true,
		ChannelArticlesCritical: true,
		ChannelArticlesHigh:     true,
		ChannelAlertsUser:       true,
		ChannelSystem:           true,
	}

	// Check predefined channels
	if validChannels[channel] {
		return true
	}

	// Check prefix-based channels
	if len(channel) > len(ChannelPrefixArticles) {
		prefix := channel[:len(ChannelPrefixArticles)]
		if prefix == ChannelPrefixArticles {
			// articles:category:{slug} or articles:vendor:{name}
			return true
		}
	}

	if len(channel) > len(ChannelPrefixAlerts) {
		prefix := channel[:len(ChannelPrefixAlerts)]
		if prefix == ChannelPrefixAlerts {
			return true
		}
	}

	return false
}
