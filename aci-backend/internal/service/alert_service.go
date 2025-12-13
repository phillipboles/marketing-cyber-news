package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
)

// AlertService handles alert business logic
type AlertService struct {
	alertRepo      repository.AlertRepository
	alertMatchRepo repository.AlertMatchRepository
	articleRepo    repository.ArticleRepository
}

// NewAlertService creates a new alert service
func NewAlertService(
	alertRepo repository.AlertRepository,
	alertMatchRepo repository.AlertMatchRepository,
	articleRepo repository.ArticleRepository,
) *AlertService {
	if alertRepo == nil {
		panic("alertRepo cannot be nil")
	}
	if alertMatchRepo == nil {
		panic("alertMatchRepo cannot be nil")
	}
	if articleRepo == nil {
		panic("articleRepo cannot be nil")
	}

	return &AlertService{
		alertRepo:      alertRepo,
		alertMatchRepo: alertMatchRepo,
		articleRepo:    articleRepo,
	}
}

// Create creates a new alert for a user
func (s *AlertService) Create(ctx context.Context, userID uuid.UUID, name string, alertType domain.AlertType, value string) (*domain.Alert, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	if name == "" {
		return nil, fmt.Errorf("alert name is required")
	}

	if !alertType.IsValid() {
		return nil, fmt.Errorf("invalid alert type")
	}

	if value == "" {
		return nil, fmt.Errorf("alert value is required")
	}

	now := time.Now()
	alert := &domain.Alert{
		ID:        uuid.New(),
		UserID:    userID,
		Name:      name,
		Type:      alertType,
		Value:     value,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := alert.Validate(); err != nil {
		return nil, fmt.Errorf("alert validation failed: %w", err)
	}

	if err := s.alertRepo.Create(ctx, alert); err != nil {
		return nil, fmt.Errorf("failed to create alert: %w", err)
	}

	return alert, nil
}

// List returns all alerts for a user with match counts
func (s *AlertService) List(ctx context.Context, userID uuid.UUID) ([]*domain.Alert, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	alerts, err := s.alertRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list alerts: %w", err)
	}

	return alerts, nil
}

// GetByID returns a specific alert with ownership check
func (s *AlertService) GetByID(ctx context.Context, id, userID uuid.UUID) (*domain.Alert, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("alert ID is required")
	}

	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	// Check ownership
	if alert.UserID != userID {
		return nil, fmt.Errorf("alert not found")
	}

	return alert, nil
}

// Update modifies an alert with ownership check
func (s *AlertService) Update(ctx context.Context, id, userID uuid.UUID, name, value *string, isActive *bool) (*domain.Alert, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("alert ID is required")
	}

	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	// Get existing alert
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert: %w", err)
	}

	// Check ownership
	if alert.UserID != userID {
		return nil, fmt.Errorf("alert not found")
	}

	// Update fields if provided
	if name != nil {
		if *name == "" {
			return nil, fmt.Errorf("alert name cannot be empty")
		}
		alert.Name = *name
	}

	if value != nil {
		if *value == "" {
			return nil, fmt.Errorf("alert value cannot be empty")
		}
		alert.Value = *value
	}

	if isActive != nil {
		alert.IsActive = *isActive
	}

	alert.UpdatedAt = time.Now()

	// Validate updated alert
	if err := alert.Validate(); err != nil {
		return nil, fmt.Errorf("alert validation failed: %w", err)
	}

	// Update in database
	if err := s.alertRepo.Update(ctx, alert); err != nil {
		return nil, fmt.Errorf("failed to update alert: %w", err)
	}

	return alert, nil
}

// Delete removes an alert with ownership check
func (s *AlertService) Delete(ctx context.Context, id, userID uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("alert ID is required")
	}

	if userID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	// Get existing alert to check ownership
	alert, err := s.alertRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get alert: %w", err)
	}

	// Check ownership
	if alert.UserID != userID {
		return fmt.Errorf("alert not found")
	}

	// Delete alert
	if err := s.alertRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	return nil
}

// ListMatches returns matches for an alert with ownership check and pagination
func (s *AlertService) ListMatches(ctx context.Context, alertID, userID uuid.UUID, page, pageSize int) ([]*domain.AlertMatch, int, error) {
	if alertID == uuid.Nil {
		return nil, 0, fmt.Errorf("alert ID is required")
	}

	if userID == uuid.Nil {
		return nil, 0, fmt.Errorf("user ID is required")
	}

	if page < 1 {
		return nil, 0, fmt.Errorf("page must be at least 1")
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, 0, fmt.Errorf("page_size must be between 1 and 100")
	}

	// Get alert to check ownership
	alert, err := s.alertRepo.GetByID(ctx, alertID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get alert: %w", err)
	}

	// Check ownership
	if alert.UserID != userID {
		return nil, 0, fmt.Errorf("alert not found")
	}

	// Get matches for alert
	matches, err := s.alertMatchRepo.GetByAlertID(ctx, alertID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get alert matches: %w", err)
	}

	total := len(matches)

	// Apply pagination
	offset := (page - 1) * pageSize
	end := offset + pageSize

	if offset >= total {
		return []*domain.AlertMatch{}, total, nil
	}

	if end > total {
		end = total
	}

	paginatedMatches := matches[offset:end]

	// Populate article details for each match
	for i, match := range paginatedMatches {
		article, err := s.articleRepo.GetByID(ctx, match.ArticleID)
		if err != nil {
			log.Error().
				Err(err).
				Str("article_id", match.ArticleID.String()).
				Msg("Failed to load article for alert match")
			continue
		}
		paginatedMatches[i].Article = article
	}

	return paginatedMatches, total, nil
}

// MatchArticle checks article against all active alerts and creates matches
// This is called when a new article is created
func (s *AlertService) MatchArticle(ctx context.Context, article *domain.Article) ([]*domain.AlertMatch, error) {
	if article == nil {
		return nil, fmt.Errorf("article cannot be nil")
	}

	if article.ID == uuid.Nil {
		return nil, fmt.Errorf("article ID is required")
	}

	// Get all active alerts
	activeAlerts, err := s.alertRepo.GetActiveAlerts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active alerts: %w", err)
	}

	if len(activeAlerts) == 0 {
		return []*domain.AlertMatch{}, nil
	}

	// Check article against each alert
	matches := make([]*domain.AlertMatch, 0)

	for _, alert := range activeAlerts {
		// Check if alert matches article
		if !alert.Matches(article) {
			continue
		}

		// Determine priority based on article severity
		priority := domain.DeterminePriority(article)

		// Create alert match
		now := time.Now()
		match := &domain.AlertMatch{
			ID:        uuid.New(),
			AlertID:   alert.ID,
			ArticleID: article.ID,
			Priority:  priority,
			MatchedAt: now,
		}

		if err := match.Validate(); err != nil {
			log.Error().
				Err(err).
				Str("alert_id", alert.ID.String()).
				Str("article_id", article.ID.String()).
				Msg("Alert match validation failed")
			continue
		}

		// Save match to database
		if err := s.alertMatchRepo.Create(ctx, match); err != nil {
			log.Error().
				Err(err).
				Str("alert_id", alert.ID.String()).
				Str("article_id", article.ID.String()).
				Msg("Failed to create alert match")
			continue
		}

		// Populate alert and article for notification
		match.Alert = alert
		match.Article = article

		matches = append(matches, match)

		log.Info().
			Str("alert_id", alert.ID.String()).
			Str("alert_name", alert.Name).
			Str("article_id", article.ID.String()).
			Str("article_title", article.Title).
			Str("priority", priority).
			Msg("Alert matched article")
	}

	return matches, nil
}
