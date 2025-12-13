package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/phillipboles/aci-backend/internal/api/middleware"
	"github.com/phillipboles/aci-backend/internal/api/response"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/service"
)

// AdminHandler handles admin-only HTTP requests
type AdminHandler struct {
	adminService *service.AdminService
}

// NewAdminHandler creates a new admin handler instance
func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	if adminService == nil {
		panic("adminService cannot be nil")
	}

	return &AdminHandler{
		adminService: adminService,
	}
}

// UpdateArticleRequest represents the request body for updating an article
type UpdateArticleRequest struct {
	Severity    *string `json:"severity,omitempty"`
	IsPublished *bool   `json:"is_published,omitempty"`
	Title       *string `json:"title,omitempty"`
	Summary     *string `json:"summary,omitempty"`
	Content     *string `json:"content,omitempty"`
}

// UpdateArticle handles PUT /v1/admin/articles/{id}
func (h *AdminHandler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get article ID from URL
	articleIDStr := chi.URLParam(r, "id")
	if articleIDStr == "" {
		response.BadRequest(w, "Article ID is required")
		return
	}

	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		response.BadRequestWithDetails(w, "Invalid article ID format", err.Error(), requestID)
		return
	}

	// Get admin user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse request body
	var req UpdateArticleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to decode update article request")
		response.BadRequestWithDetails(w, "Invalid request body", err.Error(), requestID)
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Severity != nil {
		updates["severity"] = *req.Severity
	}
	if req.IsPublished != nil {
		updates["is_published"] = *req.IsPublished
	}
	if req.Title != nil {
		updates["title"] = *req.Title
	}
	if req.Summary != nil {
		updates["summary"] = *req.Summary
	}
	if req.Content != nil {
		updates["content"] = *req.Content
	}

	if len(updates) == 0 {
		response.BadRequest(w, "No updates provided")
		return
	}

	// Get IP and User-Agent for audit log
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	// Update article
	article, err := h.adminService.UpdateArticle(
		ctx,
		articleID,
		updates,
		claims.UserID,
		ipAddress,
		userAgent,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("article_id", articleID.String()).
			Msg("Failed to update article")
		response.InternalError(w, "Failed to update article", requestID)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("article_id", articleID.String()).
		Str("admin_user_id", claims.UserID.String()).
		Msg("Article updated successfully")

	response.Success(w, article)
}

// DeleteArticle handles DELETE /v1/admin/articles/{id}
func (h *AdminHandler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get article ID from URL
	articleIDStr := chi.URLParam(r, "id")
	if articleIDStr == "" {
		response.BadRequest(w, "Article ID is required")
		return
	}

	articleID, err := uuid.Parse(articleIDStr)
	if err != nil {
		response.BadRequestWithDetails(w, "Invalid article ID format", err.Error(), requestID)
		return
	}

	// Get admin user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Get IP and User-Agent for audit log
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	// Delete article
	if err := h.adminService.DeleteArticle(
		ctx,
		articleID,
		claims.UserID,
		ipAddress,
		userAgent,
	); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("article_id", articleID.String()).
			Msg("Failed to delete article")
		response.InternalError(w, "Failed to delete article", requestID)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("article_id", articleID.String()).
		Str("admin_user_id", claims.UserID.String()).
		Msg("Article deleted successfully")

	response.NoContent(w)
}

// ListSources handles GET /v1/admin/sources
func (h *AdminHandler) ListSources(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// List all sources (including inactive)
	sources, err := h.adminService.ListSources(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to list sources")
		response.InternalError(w, "Failed to list sources", requestID)
		return
	}

	response.Success(w, sources)
}

// CreateSourceRequest represents the request body for creating a source
type CreateSourceRequest struct {
	Name        string   `json:"name"`
	URL         string   `json:"url"`
	Description *string  `json:"description,omitempty"`
	TrustScore  *float64 `json:"trust_score,omitempty"`
}

// CreateSource handles POST /v1/admin/sources
func (h *AdminHandler) CreateSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get admin user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse request body
	var req CreateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to decode create source request")
		response.BadRequestWithDetails(w, "Invalid request body", err.Error(), requestID)
		return
	}

	// Validate required fields
	if req.Name == "" {
		response.BadRequest(w, "Source name is required")
		return
	}

	if req.URL == "" {
		response.BadRequest(w, "Source URL is required")
		return
	}

	// Create source domain object
	source, err := domain.NewSource(req.Name, req.URL, req.Description)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to create source object")
		response.BadRequestWithDetails(w, "Invalid source data", err.Error(), requestID)
		return
	}

	// Set trust score if provided
	if req.TrustScore != nil {
		if err := source.UpdateTrustScore(*req.TrustScore); err != nil {
			response.BadRequestWithDetails(w, "Invalid trust score", err.Error(), requestID)
			return
		}
	}

	// Get IP and User-Agent for audit log
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	// Create source
	createdSource, err := h.adminService.CreateSource(
		ctx,
		source,
		claims.UserID,
		ipAddress,
		userAgent,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to create source")
		response.InternalError(w, "Failed to create source", requestID)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("source_id", createdSource.ID.String()).
		Str("admin_user_id", claims.UserID.String()).
		Msg("Source created successfully")

	response.Created(w, createdSource)
}

// UpdateSourceRequest represents the request body for updating a source
type UpdateSourceRequest struct {
	Name        *string  `json:"name,omitempty"`
	URL         *string  `json:"url,omitempty"`
	Description *string  `json:"description,omitempty"`
	IsActive    *bool    `json:"is_active,omitempty"`
	TrustScore  *float64 `json:"trust_score,omitempty"`
}

// UpdateSource handles PUT /v1/admin/sources/{id}
func (h *AdminHandler) UpdateSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get source ID from URL
	sourceIDStr := chi.URLParam(r, "id")
	if sourceIDStr == "" {
		response.BadRequest(w, "Source ID is required")
		return
	}

	sourceID, err := uuid.Parse(sourceIDStr)
	if err != nil {
		response.BadRequestWithDetails(w, "Invalid source ID format", err.Error(), requestID)
		return
	}

	// Get admin user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse request body
	var req UpdateSourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to decode update source request")
		response.BadRequestWithDetails(w, "Invalid request body", err.Error(), requestID)
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.URL != nil {
		updates["url"] = *req.URL
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.TrustScore != nil {
		updates["trust_score"] = *req.TrustScore
	}

	if len(updates) == 0 {
		response.BadRequest(w, "No updates provided")
		return
	}

	// Get IP and User-Agent for audit log
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	// Update source
	source, err := h.adminService.UpdateSource(
		ctx,
		sourceID,
		updates,
		claims.UserID,
		ipAddress,
		userAgent,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("source_id", sourceID.String()).
			Msg("Failed to update source")
		response.InternalError(w, "Failed to update source", requestID)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("source_id", sourceID.String()).
		Str("admin_user_id", claims.UserID.String()).
		Msg("Source updated successfully")

	response.Success(w, source)
}

// DeleteSource handles DELETE /v1/admin/sources/{id}
func (h *AdminHandler) DeleteSource(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get source ID from URL
	sourceIDStr := chi.URLParam(r, "id")
	if sourceIDStr == "" {
		response.BadRequest(w, "Source ID is required")
		return
	}

	sourceID, err := uuid.Parse(sourceIDStr)
	if err != nil {
		response.BadRequestWithDetails(w, "Invalid source ID format", err.Error(), requestID)
		return
	}

	// Get admin user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Get IP and User-Agent for audit log
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	// Delete (deactivate) source
	if err := h.adminService.DeleteSource(
		ctx,
		sourceID,
		claims.UserID,
		ipAddress,
		userAgent,
	); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("source_id", sourceID.String()).
			Msg("Failed to delete source")
		response.InternalError(w, "Failed to delete source", requestID)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("source_id", sourceID.String()).
		Str("admin_user_id", claims.UserID.String()).
		Msg("Source deleted successfully")

	response.NoContent(w)
}

// ListUsers handles GET /v1/admin/users
func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Parse pagination parameters
	limit, offset := ParseLimitOffset(r)

	// List users
	users, totalCount, err := h.adminService.ListUsers(ctx, limit, offset)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to list users")
		response.InternalError(w, "Failed to list users", requestID)
		return
	}

	// Calculate pagination metadata
	meta := &response.Meta{
		Page:       (offset / limit) + 1,
		PageSize:   limit,
		TotalCount: totalCount,
		TotalPages: (totalCount + limit - 1) / limit,
	}

	response.SuccessWithMeta(w, users, meta)
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Role          *string `json:"role,omitempty"`
	EmailVerified *bool   `json:"email_verified,omitempty"`
	Email         *string `json:"email,omitempty"`
	Name          *string `json:"name,omitempty"`
}

// UpdateUser handles PUT /v1/admin/users/{id}
func (h *AdminHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get user ID from URL
	userIDStr := chi.URLParam(r, "id")
	if userIDStr == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequestWithDetails(w, "Invalid user ID format", err.Error(), requestID)
		return
	}

	// Get admin user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse request body
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to decode update user request")
		response.BadRequestWithDetails(w, "Invalid request body", err.Error(), requestID)
		return
	}

	// Build updates map
	updates := make(map[string]interface{})
	if req.Role != nil {
		updates["role"] = *req.Role
	}
	if req.EmailVerified != nil {
		updates["email_verified"] = *req.EmailVerified
	}
	if req.Email != nil {
		updates["email"] = *req.Email
	}
	if req.Name != nil {
		updates["name"] = *req.Name
	}

	if len(updates) == 0 {
		response.BadRequest(w, "No updates provided")
		return
	}

	// Get IP and User-Agent for audit log
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	// Update user
	user, err := h.adminService.UpdateUser(
		ctx,
		userID,
		updates,
		claims.UserID,
		ipAddress,
		userAgent,
	)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", userID.String()).
			Msg("Failed to update user")
		response.InternalError(w, "Failed to update user", requestID)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("user_id", userID.String()).
		Str("admin_user_id", claims.UserID.String()).
		Msg("User updated successfully")

	response.Success(w, user)
}

// DeleteUser handles DELETE /v1/admin/users/{id}
func (h *AdminHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get user ID from URL
	userIDStr := chi.URLParam(r, "id")
	if userIDStr == "" {
		response.BadRequest(w, "User ID is required")
		return
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequestWithDetails(w, "Invalid user ID format", err.Error(), requestID)
		return
	}

	// Get admin user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Get IP and User-Agent for audit log
	ipAddress := GetClientIP(r)
	userAgent := r.UserAgent()

	// Delete user
	if err := h.adminService.DeleteUser(
		ctx,
		userID,
		claims.UserID,
		ipAddress,
		userAgent,
	); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", userID.String()).
			Msg("Failed to delete user")

		if err.Error() == "cannot delete your own account" {
			response.BadRequest(w, err.Error())
			return
		}

		response.InternalError(w, "Failed to delete user", requestID)
		return
	}

	log.Info().
		Str("request_id", requestID).
		Str("user_id", userID.String()).
		Str("admin_user_id", claims.UserID.String()).
		Msg("User deleted successfully")

	response.NoContent(w)
}

// ListAuditLogs handles GET /v1/admin/audit-logs
func (h *AdminHandler) ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Parse query parameters
	filter, err := parseAuditLogFilter(r)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to parse audit log filter")
		response.BadRequestWithDetails(w, "Invalid query parameters", err.Error(), requestID)
		return
	}

	// List audit logs
	logs, totalCount, err := h.adminService.ListAuditLogs(ctx, filter)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to list audit logs")
		response.InternalError(w, "Failed to list audit logs", requestID)
		return
	}

	// Calculate pagination metadata
	meta := &response.Meta{
		Page:       (filter.Offset / filter.Limit) + 1,
		PageSize:   filter.Limit,
		TotalCount: totalCount,
		TotalPages: (totalCount + filter.Limit - 1) / filter.Limit,
	}

	response.SuccessWithMeta(w, logs, meta)
}

// Helper functions (shared helpers are in helpers.go)

func parseAuditLogFilter(r *http.Request) (*domain.AuditLogFilter, error) {
	query := r.URL.Query()
	filter := &domain.AuditLogFilter{}

	// Parse pagination
	filter.Limit, filter.Offset = ParseLimitOffset(r)

	// Parse user_id
	if userIDStr := query.Get("user_id"); userIDStr != "" {
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, err
		}
		filter.UserID = &userID
	}

	// Parse action
	if action := query.Get("action"); action != "" {
		filter.Action = &action
	}

	// Parse resource_type
	if resourceType := query.Get("resource_type"); resourceType != "" {
		filter.ResourceType = &resourceType
	}

	// Parse resource_id
	if resourceIDStr := query.Get("resource_id"); resourceIDStr != "" {
		resourceID, err := uuid.Parse(resourceIDStr)
		if err != nil {
			return nil, err
		}
		filter.ResourceID = &resourceID
	}

	// Parse start_date
	if startDateStr := query.Get("start_date"); startDateStr != "" {
		startDate, err := time.Parse(time.RFC3339, startDateStr)
		if err != nil {
			return nil, err
		}
		filter.StartDate = &startDate
	}

	// Parse end_date
	if endDateStr := query.Get("end_date"); endDateStr != "" {
		endDate, err := time.Parse(time.RFC3339, endDateStr)
		if err != nil {
			return nil, err
		}
		filter.EndDate = &endDate
	}

	return filter, nil
}
