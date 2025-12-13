package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/phillipboles/aci-backend/internal/api/middleware"
	"github.com/phillipboles/aci-backend/internal/api/response"
	"github.com/phillipboles/aci-backend/internal/repository"
	"github.com/phillipboles/aci-backend/internal/service"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	engagementService *service.EngagementService
	userRepo          repository.UserRepository
}

// NewUserHandler creates a new user handler instance
func NewUserHandler(
	engagementService *service.EngagementService,
	userRepo repository.UserRepository,
) *UserHandler {
	if engagementService == nil {
		panic("engagementService cannot be nil")
	}
	if userRepo == nil {
		panic("userRepo cannot be nil")
	}

	return &UserHandler{
		engagementService: engagementService,
		userRepo:          userRepo,
	}
}

// UserResponse represents a user profile response
type UserResponse struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
	Role          string  `json:"role"`
	EmailVerified bool    `json:"email_verified"`
	CreatedAt     string  `json:"created_at"`
	LastLoginAt   *string `json:"last_login_at,omitempty"`
}

// UpdateProfileRequest represents a user profile update request
type UpdateProfileRequest struct {
	Name string `json:"name"`
}

// UserStats represents user engagement statistics
type UserStats struct {
	TotalArticlesRead    int     `json:"total_articles_read"`
	TotalReadingTime     int     `json:"total_reading_time_seconds"`
	BookmarkCount        int     `json:"bookmark_count"`
	AlertCount           int     `json:"alert_count"`
	AlertMatchCount      int     `json:"alert_match_count"`
	FavoriteCategory     string  `json:"favorite_category,omitempty"`
	ArticlesThisWeek     int     `json:"articles_this_week"`
	ArticlesThisMonth    int     `json:"articles_this_month"`
	AverageReadingTime   float64 `json:"average_reading_time_seconds"`
}

// GetCurrentUser handles GET /v1/users/me - returns current user profile
func (h *UserHandler) GetCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get user from context (set by auth middleware)
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		log.Error().
			Str("request_id", requestID).
			Msg("User claims not found in context")
		response.Unauthorized(w, "Authentication required")
		return
	}

	user, err := h.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to get user")
		response.InternalError(w, "Failed to retrieve user profile", requestID)
		return
	}

	userResponse := UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		Role:          string(user.Role),
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if user.LastLoginAt != nil {
		lastLogin := user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		userResponse.LastLoginAt = &lastLogin
	}

	response.Success(w, userResponse)
}

// UpdateCurrentUser handles PATCH /v1/users/me - updates current user profile
func (h *UserHandler) UpdateCurrentUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		log.Error().
			Str("request_id", requestID).
			Msg("User claims not found in context")
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse request body
	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to decode request body")
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate input
	if req.Name == "" {
		response.BadRequest(w, "Name is required")
		return
	}

	// Get current user
	user, err := h.userRepo.GetByID(ctx, claims.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to get user")
		response.InternalError(w, "Failed to retrieve user", requestID)
		return
	}

	// Update user
	user.Name = req.Name

	if err := h.userRepo.Update(ctx, user); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to update user")
		response.InternalError(w, "Failed to update user profile", requestID)
		return
	}

	userResponse := UserResponse{
		ID:            user.ID.String(),
		Email:         user.Email,
		Name:          user.Name,
		Role:          string(user.Role),
		EmailVerified: user.EmailVerified,
		CreatedAt:     user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	if user.LastLoginAt != nil {
		lastLogin := user.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		userResponse.LastLoginAt = &lastLogin
	}

	response.Success(w, userResponse)
}

// GetBookmarks handles GET /v1/users/me/bookmarks - returns paginated bookmarks
func (h *UserHandler) GetBookmarks(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		log.Error().
			Str("request_id", requestID).
			Msg("User claims not found in context")
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse pagination parameters
	page, pageSize, err := ParsePagination(r)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Invalid pagination parameters")
		response.BadRequest(w, "Invalid pagination parameters")
		return
	}

	articles, total, err := h.engagementService.GetBookmarks(ctx, claims.UserID, page, pageSize)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to get bookmarks")
		response.InternalError(w, "Failed to retrieve bookmarks", requestID)
		return
	}

	articleResponses := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		articleResponses[i] = toArticleResponse(article)
	}

	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
		TotalPages: CalculateTotalPages(total, pageSize),
	}

	response.SuccessWithMeta(w, articleResponses, meta)
}

// GetReadingHistory handles GET /v1/users/me/history - returns reading history
func (h *UserHandler) GetReadingHistory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		log.Error().
			Str("request_id", requestID).
			Msg("User claims not found in context")
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse pagination parameters
	page, pageSize, err := ParsePagination(r)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Invalid pagination parameters")
		response.BadRequest(w, "Invalid pagination parameters")
		return
	}

	reads, total, err := h.engagementService.GetReadingHistory(ctx, claims.UserID, page, pageSize)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to get reading history")
		response.InternalError(w, "Failed to retrieve reading history", requestID)
		return
	}

	// Convert to response format
	historyResponses := make([]map[string]interface{}, len(reads))
	for i, read := range reads {
		historyResponses[i] = map[string]interface{}{
			"id":                   read.ID.String(),
			"read_at":              read.ReadAt.Format("2006-01-02T15:04:05Z07:00"),
			"reading_time_seconds": read.ReadingTimeSeconds,
			"article":              toArticleResponse(read.Article),
		}
	}

	meta := &response.Meta{
		Page:       page,
		PageSize:   pageSize,
		TotalCount: total,
		TotalPages: CalculateTotalPages(total, pageSize),
	}

	response.SuccessWithMeta(w, historyResponses, meta)
}

// GetStats handles GET /v1/users/me/stats - returns user engagement statistics
func (h *UserHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	// Get user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		log.Error().
			Str("request_id", requestID).
			Msg("User claims not found in context")
		response.Unauthorized(w, "Authentication required")
		return
	}

	stats, err := h.engagementService.GetUserStats(ctx, claims.UserID)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Msg("Failed to get user stats")
		response.InternalError(w, "Failed to retrieve user statistics", requestID)
		return
	}

	userStats := UserStats{
		TotalArticlesRead:  stats.TotalArticlesRead,
		TotalReadingTime:   stats.TotalReadingTime,
		BookmarkCount:      stats.TotalBookmarks,
		AlertCount:         stats.TotalAlerts,
		AlertMatchCount:    stats.TotalAlertMatches,
		FavoriteCategory:   stats.FavoriteCategory,
		ArticlesThisWeek:   stats.ArticlesThisWeek,
		ArticlesThisMonth:  stats.ArticlesThisMonth,
		AverageReadingTime: stats.AverageReadingTime,
	}

	response.Success(w, userStats)
}

