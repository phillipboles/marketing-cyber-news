package service

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/websocket"
	"github.com/rs/zerolog/log"
)

// NotificationService handles broadcasting notifications via WebSocket
type NotificationService struct {
	hub *websocket.Hub
}

// NewNotificationService creates a new notification service
func NewNotificationService(hub *websocket.Hub) (*NotificationService, error) {
	if hub == nil {
		return nil, fmt.Errorf("hub is required")
	}

	return &NotificationService{
		hub: hub,
	}, nil
}

// NotifyNewArticle broadcasts new article to appropriate channels
// Broadcasts to:
// - articles:all
// - articles:{severity} if critical or high
// - articles:category:{slug}
// - articles:vendor:{name} for each vendor
func (s *NotificationService) NotifyNewArticle(article *domain.Article) error {
	if article == nil {
		return fmt.Errorf("article is required")
	}

	// Create message
	msg, err := websocket.NewMessage(websocket.MessageTypeArticleNew, article)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Broadcast to articles:all
	s.hub.Broadcast(websocket.ChannelArticlesAll, msg)

	// Broadcast to severity-specific channels
	switch article.Severity {
	case domain.SeverityCritical:
		s.hub.Broadcast(websocket.ChannelArticlesCritical, msg)
	case domain.SeverityHigh:
		s.hub.Broadcast(websocket.ChannelArticlesHigh, msg)
	}

	// Broadcast to category channel
	if article.Category != nil {
		categoryChannel := websocket.BuildCategoryChannel(article.Category.Slug)
		s.hub.Broadcast(categoryChannel, msg)
	}

	// Broadcast to vendor channels
	for _, vendor := range article.Vendors {
		vendorChannel := websocket.BuildVendorChannel(strings.ToLower(vendor))
		s.hub.Broadcast(vendorChannel, msg)
	}

	log.Info().
		Str("article_id", article.ID.String()).
		Str("title", article.Title).
		Str("severity", string(article.Severity)).
		Int("vendor_count", len(article.Vendors)).
		Msg("New article notification broadcasted")

	return nil
}

// NotifyArticleUpdated broadcasts article update
// Broadcasts to:
// - articles:all
// - articles:{severity} if critical or high
// - articles:category:{slug}
// - articles:vendor:{name} for each vendor
func (s *NotificationService) NotifyArticleUpdated(article *domain.Article) error {
	if article == nil {
		return fmt.Errorf("article is required")
	}

	// Create message
	msg, err := websocket.NewMessage(websocket.MessageTypeArticleUpdated, article)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Broadcast to articles:all
	s.hub.Broadcast(websocket.ChannelArticlesAll, msg)

	// Broadcast to severity-specific channels
	switch article.Severity {
	case domain.SeverityCritical:
		s.hub.Broadcast(websocket.ChannelArticlesCritical, msg)
	case domain.SeverityHigh:
		s.hub.Broadcast(websocket.ChannelArticlesHigh, msg)
	}

	// Broadcast to category channel
	if article.Category != nil {
		categoryChannel := websocket.BuildCategoryChannel(article.Category.Slug)
		s.hub.Broadcast(categoryChannel, msg)
	}

	// Broadcast to vendor channels
	for _, vendor := range article.Vendors {
		vendorChannel := websocket.BuildVendorChannel(strings.ToLower(vendor))
		s.hub.Broadcast(vendorChannel, msg)
	}

	log.Info().
		Str("article_id", article.ID.String()).
		Str("title", article.Title).
		Msg("Article update notification broadcasted")

	return nil
}

// NotifyAlertMatch sends alert match to specific user
// Sends to alerts:user channel for the specific user
func (s *NotificationService) NotifyAlertMatch(userID uuid.UUID, match *domain.AlertMatch) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if match == nil {
		return fmt.Errorf("alert match is required")
	}

	// Create message
	msg, err := websocket.NewMessage(websocket.MessageTypeAlertMatch, match)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Send to all user's connections
	s.hub.BroadcastToUser(userID, msg)

	log.Info().
		Str("user_id", userID.String()).
		Str("alert_id", match.AlertID.String()).
		Str("article_id", match.ArticleID.String()).
		Str("priority", match.Priority).
		Msg("Alert match notification sent to user")

	return nil
}

// BroadcastSystemMessage broadcasts a system message to all connected clients
func (s *NotificationService) BroadcastSystemMessage(message string) error {
	if message == "" {
		return fmt.Errorf("message is required")
	}

	// Create message payload
	payload := map[string]interface{}{
		"message": message,
	}

	msg, err := websocket.NewMessage(websocket.MessageTypeError, payload)
	if err != nil {
		return fmt.Errorf("failed to create message: %w", err)
	}

	// Broadcast to system channel
	s.hub.Broadcast(websocket.ChannelSystem, msg)

	log.Info().
		Str("message", message).
		Msg("System message broadcasted")

	return nil
}

// GetHubStats returns current hub statistics
func (s *NotificationService) GetHubStats() map[string]interface{} {
	return s.hub.GetStats()
}
