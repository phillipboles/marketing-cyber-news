package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/phillipboles/aci-backend/internal/api/middleware"
	"github.com/phillipboles/aci-backend/internal/api/response"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
	"github.com/phillipboles/aci-backend/internal/service"
)

// ArticleHandler handles article-related HTTP requests
type ArticleHandler struct {
	articleRepo       repository.ArticleRepository
	searchService     *service.SearchService
	engagementService *service.EngagementService
}

// NewArticleHandler creates a new article handler instance
func NewArticleHandler(
	articleRepo repository.ArticleRepository,
	searchService *service.SearchService,
	engagementService *service.EngagementService,
) *ArticleHandler {
	if articleRepo == nil {
		panic("articleRepo cannot be nil")
	}
	if searchService == nil {
		panic("searchService cannot be nil")
	}
	if engagementService == nil {
		panic("engagementService cannot be nil")
	}

	return &ArticleHandler{
		articleRepo:       articleRepo,
		searchService:     searchService,
		engagementService: engagementService,
	}
}

// CategorySummary represents a minimal category response
type CategorySummary struct {
	ID    uuid.UUID `json:"id"`
	Name  string    `json:"name"`
	Slug  string    `json:"slug"`
	Color string    `json:"color"`
	Icon  *string   `json:"icon,omitempty"`
}

// SourceSummary represents a minimal source response
type SourceSummary struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	URL  string    `json:"url"`
}

// ArticleResponse represents a single article in list view
type ArticleResponse struct {
	ID                 uuid.UUID        `json:"id"`
	Title              string           `json:"title"`
	Slug               string           `json:"slug"`
	Summary            *string          `json:"summary,omitempty"`
	Category           *CategorySummary `json:"category,omitempty"`
	Source             *SourceSummary   `json:"source,omitempty"`
	Severity           string           `json:"severity"`
	Tags               []string         `json:"tags"`
	CVEs               []string         `json:"cves"`
	Vendors            []string         `json:"vendors"`
	ReadingTimeMinutes int              `json:"reading_time_minutes"`
	ViewCount          int              `json:"view_count"`
	PublishedAt        string           `json:"published_at"`
}

// ArticleDetailResponse represents a full article with all details
type ArticleDetailResponse struct {
	ArticleResponse
	Content            string           `json:"content"`
	ThreatType         *string          `json:"threat_type,omitempty"`
	AttackVector       *string          `json:"attack_vector,omitempty"`
	ImpactAssessment   *string          `json:"impact_assessment,omitempty"`
	RecommendedActions []string         `json:"recommended_actions,omitempty"`
	IOCs               []domain.IOC     `json:"iocs,omitempty"`
	ArmorCTA           *domain.ArmorCTA `json:"armor_cta,omitempty"`
}

// List handles GET /v1/articles - returns paginated list of articles
func (h *ArticleHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	filter, err := parseArticleFilter(r)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to parse article filter")
		response.BadRequestWithDetails(w, "Invalid query parameters", err.Error(), requestID)
		return
	}

	if err := filter.Validate(); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Invalid filter parameters")
		response.BadRequestWithDetails(w, "Invalid filter parameters", err.Error(), requestID)
		return
	}

	articles, total, err := h.articleRepo.List(ctx, filter)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to list articles")
		response.InternalError(w, "Failed to retrieve articles", requestID)
		return
	}

	articleResponses := make([]ArticleResponse, len(articles))
	for i, article := range articles {
		articleResponses[i] = toArticleResponse(article)
	}

	meta := &response.Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalCount: total,
		TotalPages: CalculateTotalPages(total, filter.PageSize),
	}

	response.SuccessWithMeta(w, articleResponses, meta)
}

// GetByID handles GET /v1/articles/{id} - returns a single article by ID
func (h *ArticleHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.BadRequest(w, "Article ID is required")
		return
	}

	articleID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("id", idStr).
			Msg("Invalid article ID format")
		response.BadRequest(w, "Invalid article ID format")
		return
	}

	article, err := h.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("article_id", articleID.String()).
			Msg("Failed to get article")
		response.NotFound(w, "Article not found")
		return
	}

	// Increment view count asynchronously
	go func() {
		bgCtx := context.Background()
		if err := h.articleRepo.IncrementViewCount(bgCtx, articleID); err != nil {
			log.Error().
				Err(err).
				Str("article_id", articleID.String()).
				Msg("Failed to increment view count")
		}
	}()

	articleDetail := toArticleDetailResponse(article)
	response.Success(w, articleDetail)
}

// GetBySlug handles GET /v1/articles/slug/{slug} - returns a single article by slug
func (h *ArticleHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.BadRequest(w, "Article slug is required")
		return
	}

	article, err := h.articleRepo.GetBySlug(ctx, slug)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("slug", slug).
			Msg("Failed to get article by slug")
		response.NotFound(w, "Article not found")
		return
	}

	// Increment view count asynchronously
	go func() {
		bgCtx := context.Background()
		if err := h.articleRepo.IncrementViewCount(bgCtx, article.ID); err != nil {
			log.Error().
				Err(err).
				Str("article_id", article.ID.String()).
				Msg("Failed to increment view count")
		}
	}()

	articleDetail := toArticleDetailResponse(article)
	response.Success(w, articleDetail)
}

// Search handles GET /v1/articles/search - performs full-text search
func (h *ArticleHandler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getRequestID(ctx)

	query := r.URL.Query().Get("q")
	if query == "" {
		response.BadRequest(w, "Search query parameter 'q' is required")
		return
	}

	filter, err := parseArticleFilter(r)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to parse search filter")
		response.BadRequestWithDetails(w, "Invalid query parameters", err.Error(), requestID)
		return
	}

	results, total, err := h.searchService.Search(ctx, query, filter)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("query", query).
			Msg("Failed to search articles")
		response.InternalError(w, "Failed to search articles", requestID)
		return
	}

	searchResponses := make([]map[string]interface{}, len(results))
	for i, result := range results {
		searchResponses[i] = map[string]interface{}{
			"article":   toArticleResponse(result.Article),
			"score":     result.Score,
			"highlight": result.Highlight,
		}
	}

	meta := &response.Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalCount: total,
		TotalPages: CalculateTotalPages(total, filter.PageSize),
	}

	response.SuccessWithMeta(w, searchResponses, meta)
}

// parseArticleFilter extracts and validates filter parameters from request
func parseArticleFilter(r *http.Request) (*domain.ArticleFilter, error) {
	filter := domain.NewArticleFilter()

	query := r.URL.Query()

	// Parse pagination
	if pageStr := query.Get("page"); pageStr != "" {
		page, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, fmt.Errorf("invalid page parameter: %w", err)
		}
		filter.Page = page
	}

	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		pageSize, err := strconv.Atoi(pageSizeStr)
		if err != nil {
			return nil, fmt.Errorf("invalid page_size parameter: %w", err)
		}
		filter.PageSize = pageSize
	}

	// Parse category_id
	if categoryIDStr := query.Get("category_id"); categoryIDStr != "" {
		categoryID, err := uuid.Parse(categoryIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid category_id parameter: %w", err)
		}
		filter.CategoryID = &categoryID
	}

	// Parse source_id
	if sourceIDStr := query.Get("source_id"); sourceIDStr != "" {
		sourceID, err := uuid.Parse(sourceIDStr)
		if err != nil {
			return nil, fmt.Errorf("invalid source_id parameter: %w", err)
		}
		filter.SourceID = &sourceID
	}

	// Parse severity
	if severityStr := query.Get("severity"); severityStr != "" {
		severity := domain.Severity(severityStr)
		filter.Severity = &severity
	}

	// Parse tags (comma-separated)
	if tagsStr := query.Get("tags"); tagsStr != "" {
		tags := strings.Split(tagsStr, ",")
		trimmedTags := make([]string, 0, len(tags))
		for _, tag := range tags {
			trimmed := strings.TrimSpace(tag)
			if trimmed != "" {
				trimmedTags = append(trimmedTags, trimmed)
			}
		}
		filter.Tags = trimmedTags
	}

	// Parse CVE
	if cveStr := query.Get("cve"); cveStr != "" {
		filter.CVE = &cveStr
	}

	// Parse vendor
	if vendorStr := query.Get("vendor"); vendorStr != "" {
		filter.Vendor = &vendorStr
	}

	// Parse date range
	if dateFromStr := query.Get("date_from"); dateFromStr != "" {
		dateFrom, err := time.Parse(time.RFC3339, dateFromStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date_from parameter (use RFC3339 format): %w", err)
		}
		filter.DateFrom = &dateFrom
	}

	if dateToStr := query.Get("date_to"); dateToStr != "" {
		dateTo, err := time.Parse(time.RFC3339, dateToStr)
		if err != nil {
			return nil, fmt.Errorf("invalid date_to parameter (use RFC3339 format): %w", err)
		}
		filter.DateTo = &dateTo
	}

	return filter, nil
}

// toArticleResponse converts domain article to API response
func toArticleResponse(article *domain.Article) ArticleResponse {
	if article == nil {
		return ArticleResponse{}
	}

	response := ArticleResponse{
		ID:                 article.ID,
		Title:              article.Title,
		Slug:               article.Slug,
		Summary:            article.Summary,
		Severity:           string(article.Severity),
		Tags:               article.Tags,
		CVEs:               article.CVEs,
		Vendors:            article.Vendors,
		ReadingTimeMinutes: article.ReadingTimeMinutes,
		ViewCount:          article.ViewCount,
		PublishedAt:        article.PublishedAt.Format(time.RFC3339),
	}

	if article.Category != nil {
		response.Category = &CategorySummary{
			ID:    article.Category.ID,
			Name:  article.Category.Name,
			Slug:  article.Category.Slug,
			Color: article.Category.Color,
			Icon:  article.Category.Icon,
		}
	}

	if article.Source != nil {
		response.Source = &SourceSummary{
			ID:   article.Source.ID,
			Name: article.Source.Name,
			URL:  article.Source.URL,
		}
	}

	return response
}

// toArticleDetailResponse converts domain article to detailed API response
func toArticleDetailResponse(article *domain.Article) ArticleDetailResponse {
	if article == nil {
		return ArticleDetailResponse{}
	}

	return ArticleDetailResponse{
		ArticleResponse:    toArticleResponse(article),
		Content:            article.Content,
		ThreatType:         article.ThreatType,
		AttackVector:       article.AttackVector,
		ImpactAssessment:   article.ImpactAssessment,
		RecommendedActions: article.RecommendedActions,
		IOCs:               article.IOCs,
		ArmorCTA:           article.ArmorCTA,
	}
}


// AddBookmark handles POST /v1/articles/{id}/bookmark - bookmark an article
func (h *ArticleHandler) AddBookmark(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestID := ""
	if reqID, ok := ctx.Value("request_id").(string); ok {
		requestID = reqID
	}

	// Get user from context (set by auth middleware)
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		log.Error().
			Str("request_id", requestID).
			Msg("User claims not found in context")
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse article ID from URL
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.BadRequest(w, "Article ID is required")
		return
	}

	articleID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("id", idStr).
			Msg("Invalid article ID format")
		response.BadRequest(w, "Invalid article ID format")
		return
	}

	// Add bookmark
	if err := h.engagementService.AddBookmark(ctx, claims.UserID, articleID); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Str("article_id", articleID.String()).
			Msg("Failed to add bookmark")
		response.InternalError(w, "Failed to bookmark article", requestID)
		return
	}

	response.SuccessWithMessage(w, map[string]bool{"bookmarked": true}, "Article bookmarked successfully")
}

// RemoveBookmark handles DELETE /v1/articles/{id}/bookmark - remove bookmark
func (h *ArticleHandler) RemoveBookmark(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestID := ""
	if reqID, ok := ctx.Value("request_id").(string); ok {
		requestID = reqID
	}

	// Get user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		log.Error().
			Str("request_id", requestID).
			Msg("User claims not found in context")
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse article ID from URL
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.BadRequest(w, "Article ID is required")
		return
	}

	articleID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("id", idStr).
			Msg("Invalid article ID format")
		response.BadRequest(w, "Invalid article ID format")
		return
	}

	// Remove bookmark
	if err := h.engagementService.RemoveBookmark(ctx, claims.UserID, articleID); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Str("article_id", articleID.String()).
			Msg("Failed to remove bookmark")
		response.InternalError(w, "Failed to remove bookmark", requestID)
		return
	}

	response.SuccessWithMessage(w, map[string]bool{"bookmarked": false}, "Bookmark removed successfully")
}

// MarkReadRequest represents the request body for marking an article as read
type MarkReadRequest struct {
	ReadingTimeSeconds *int `json:"reading_time_seconds,omitempty"`
}

// MarkRead handles POST /v1/articles/{id}/read - mark article as read
func (h *ArticleHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	requestID := ""
	if reqID, ok := ctx.Value("request_id").(string); ok {
		requestID = reqID
	}

	// Get user from context
	claims, ok := middleware.GetUserFromContext(ctx)
	if !ok {
		log.Error().
			Str("request_id", requestID).
			Msg("User claims not found in context")
		response.Unauthorized(w, "Authentication required")
		return
	}

	// Parse article ID from URL
	idStr := chi.URLParam(r, "id")
	if idStr == "" {
		response.BadRequest(w, "Article ID is required")
		return
	}

	articleID, err := uuid.Parse(idStr)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("id", idStr).
			Msg("Invalid article ID format")
		response.BadRequest(w, "Invalid article ID format")
		return
	}

	// Parse optional reading time from request body
	var req MarkReadRequest
	if r.Body != nil && r.ContentLength > 0 {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			log.Warn().
				Err(err).
				Str("request_id", requestID).
				Msg("Failed to decode request body, using default reading time")
		}
	}

	// Mark article as read
	if err := h.engagementService.MarkRead(ctx, claims.UserID, articleID, req.ReadingTimeSeconds); err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("user_id", claims.UserID.String()).
			Str("article_id", articleID.String()).
			Msg("Failed to mark article as read")
		response.InternalError(w, "Failed to mark article as read", requestID)
		return
	}

	response.SuccessWithMessage(w, map[string]bool{"read": true}, "Article marked as read")
}
