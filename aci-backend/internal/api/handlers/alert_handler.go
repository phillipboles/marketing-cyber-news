package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/phillipboles/aci-backend/internal/api/middleware"
	"github.com/phillipboles/aci-backend/internal/api/response"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/service"
)

// AlertHandler handles alert-related HTTP requests
type AlertHandler struct {
	alertService *service.AlertService
}

// NewAlertHandler creates a new alert handler instance
func NewAlertHandler(alertService *service.AlertService) *AlertHandler {
	if alertService == nil {
		panic("alertService cannot be nil")
	}

	return &AlertHandler{
		alertService: alertService,
	}
}

// CreateAlertRequest represents the request body for creating an alert
type CreateAlertRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=255"`
	Type  string `json:"type" validate:"required,oneof=keyword category severity vendor cve"`
	Value string `json:"value" validate:"required,min=1,max=500"`
}

// UpdateAlertRequest represents the request body for updating an alert
type UpdateAlertRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=255"`
	Value    *string `json:"value,omitempty" validate:"omitempty,min=1,max=500"`
	IsActive *bool   `json:"is_active,omitempty"`
}

// AlertResponse represents an alert in API responses
type AlertResponse struct {
	ID         uuid.UUID `json:"id"`
	Name       string    `json:"name"`
	Type       string    `json:"type"`
	Value      string    `json:"value"`
	IsActive   bool      `json:"is_active"`
	MatchCount int       `json:"match_count"`
	CreatedAt  string    `json:"created_at"`
	UpdatedAt  string    `json:"updated_at"`
}

// AlertMatchResponse represents an alert match in API responses
type AlertMatchResponse struct {
	ID         uuid.UUID                `json:"id"`
	AlertID    uuid.UUID                `json:"alert_id"`
	ArticleID  uuid.UUID                `json:"article_id"`
	Priority   string                   `json:"priority"`
	MatchedAt  string                   `json:"matched_at"`
	NotifiedAt *string                  `json:"notified_at,omitempty"`
	Article    *ArticleResponse         `json:"article,omitempty"`
}

// Validate validates the CreateAlertRequest
func (r *CreateAlertRequest) Validate() error {
	if r.Name == "" {
		return fmt.Errorf("name is required")
	}

	if len(r.Name) > 255 {
		return fmt.Errorf("name cannot exceed 255 characters")
	}

	if r.Type == "" {
		return fmt.Errorf("type is required")
	}

	alertType := domain.AlertType(r.Type)
	if !alertType.IsValid() {
		return fmt.Errorf("invalid alert type: must be keyword, category, severity, vendor, or cve")
	}

	if r.Value == "" {
		return fmt.Errorf("value is required")
	}

	if len(r.Value) > 500 {
		return fmt.Errorf("value cannot exceed 500 characters")
	}

	// Type-specific validation
	switch alertType {
	case domain.AlertTypeSeverity:
		severity := domain.Severity(r.Value)
		if !severity.IsValid() {
			return fmt.Errorf("invalid severity value: must be critical, high, medium, low, or informational")
		}
	case domain.AlertTypeCategory:
		if _, err := uuid.Parse(r.Value); err != nil {
			return fmt.Errorf("category value must be a valid UUID")
		}
	}

	return nil
}

// Validate validates the UpdateAlertRequest
func (r *UpdateAlertRequest) Validate() error {
	if r.Name != nil {
		if *r.Name == "" {
			return fmt.Errorf("name cannot be empty")
		}
		if len(*r.Name) > 255 {
			return fmt.Errorf("name cannot exceed 255 characters")
		}
	}

	if r.Value != nil {
		if *r.Value == "" {
			return fmt.Errorf("value cannot be empty")
		}
		if len(*r.Value) > 500 {
			return fmt.Errorf("value cannot exceed 500 characters")
		}
	}

	return nil
}

// Create handles POST /v1/alerts - creates a new alert for the authenticated user
func (h *AlertHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get authenticated user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse request body
	var req CreateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to decode alert creation request")
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Alert creation validation failed")
		response.BadRequestWithDetails(w, "Validation failed", err.Error(), requestID)
		return
	}

	// Create alert
	alert, err := h.alertService.Create(ctx, claims.UserID, req.Name, domain.AlertType(req.Type), req.Value)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to create alert")
		response.InternalError(w, "Failed to create alert", requestID)
		return
	}

	alertResp := toAlertResponse(alert)
	response.Created(w, alertResp)
}

// List handles GET /v1/alerts - returns all alerts for the authenticated user
func (h *AlertHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get authenticated user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// List alerts for user
	alerts, err := h.alertService.List(ctx, claims.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to list alerts")
		response.InternalError(w, "Failed to retrieve alerts", requestID)
		return
	}

	alertResponses := make([]AlertResponse, len(alerts))
	for i, alert := range alerts {
		alertResponses[i] = toAlertResponse(alert)
	}

	response.Success(w, alertResponses)
}

// GetByID handles GET /v1/alerts/{id} - returns a single alert by ID
func (h *AlertHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get authenticated user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse alert ID
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.BadRequest(w, "Alert ID is required")
		return
	}

	alertID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("id", idStr).
			Msg("Invalid alert ID format")
		response.BadRequest(w, "Invalid alert ID format")
		return
	}

	// Get alert with ownership check
	alert, err := h.alertService.GetByID(ctx, alertID, claims.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("alert_id", alertID.String()).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to get alert")
		response.NotFound(w, "Alert not found")
		return
	}

	alertResp := toAlertResponse(alert)
	response.Success(w, alertResp)
}

// Update handles PATCH /v1/alerts/{id} - updates an alert
func (h *AlertHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get authenticated user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse alert ID
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.BadRequest(w, "Alert ID is required")
		return
	}

	alertID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("id", idStr).
			Msg("Invalid alert ID format")
		response.BadRequest(w, "Invalid alert ID format")
		return
	}

	// Parse request body
	var req UpdateAlertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to decode alert update request")
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate request
	if err := req.Validate(); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Alert update validation failed")
		response.BadRequestWithDetails(w, "Validation failed", err.Error(), requestID)
		return
	}

	// Update alert with ownership check
	alert, err := h.alertService.Update(ctx, alertID, claims.UserID, req.Name, req.Value, req.IsActive)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("alert_id", alertID.String()).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to update alert")
		response.NotFound(w, "Alert not found")
		return
	}

	alertResp := toAlertResponse(alert)
	response.Success(w, alertResp)
}

// Delete handles DELETE /v1/alerts/{id} - deletes an alert
func (h *AlertHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get authenticated user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse alert ID
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.BadRequest(w, "Alert ID is required")
		return
	}

	alertID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("id", idStr).
			Msg("Invalid alert ID format")
		response.BadRequest(w, "Invalid alert ID format")
		return
	}

	// Delete alert with ownership check
	if err := h.alertService.Delete(ctx, alertID, claims.UserID); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("alert_id", alertID.String()).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to delete alert")
		response.NotFound(w, "Alert not found")
		return
	}

	response.NoContent(w)
}

// ListMatches handles GET /v1/alerts/{id}/matches - returns all matches for an alert
func (h *AlertHandler) ListMatches(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get authenticated user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse alert ID
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.BadRequest(w, "Alert ID is required")
		return
	}

	alertID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("id", idStr).
			Msg("Invalid alert ID format")
		response.BadRequest(w, "Invalid alert ID format")
		return
	}

	// Parse pagination parameters
	page, pageSize, err := ParsePagination(r)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Invalid pagination parameters")
		response.BadRequestWithDetails(w, "Invalid pagination parameters", err.Error(), requestID)
		return
	}

	// List matches with ownership check
	matches, total, err := h.alertService.ListMatches(ctx, alertID, claims.UserID, page, pageSize)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("alert_id", alertID.String()).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to list alert matches")
		response.NotFound(w, "Alert not found")
		return
	}

	matchResponses := make([]AlertMatchResponse, len(matches))
	for i, match := range matches {
		matchResponses[i] = toAlertMatchResponse(match)
	}

	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
		TotalPages: CalculateTotalPages(total, pageSize),
	}

	response.SuccessWithMeta(w, matchResponses, meta)
}

// toAlertResponse converts domain alert to API response
func toAlertResponse(alert *domain.Alert) AlertResponse {
	if alert == nil {
		return AlertResponse{}
	}

	return AlertResponse{
		ID:         alert.ID,
		Name:       alert.Name,
		Type:       string(alert.Type),
		Value:      alert.Value,
		IsActive:   alert.IsActive,
		MatchCount: alert.MatchCount,
		CreatedAt:  alert.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:  alert.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// toAlertMatchResponse converts domain alert match to API response
func toAlertMatchResponse(match *domain.AlertMatch) AlertMatchResponse {
	if match == nil {
		return AlertMatchResponse{}
	}

	response := AlertMatchResponse{
		ID:        match.ID,
		AlertID:   match.AlertID,
		ArticleID: match.ArticleID,
		Priority:  match.Priority,
		MatchedAt: match.MatchedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if match.NotifiedAt != nil {
		notifiedStr := match.NotifiedAt.Format("2006-01-02T15:04:05Z07:00")
		response.NotifiedAt = &notifiedStr
	}

	if match.Article != nil {
		articleResp := toArticleResponse(match.Article)
		response.Article = &articleResp
	}

	return response
}
