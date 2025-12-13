package handlers

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/phillipboles/aci-backend/internal/api/response"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
)

// CategoryHandler handles category-related HTTP requests
type CategoryHandler struct {
	categoryRepo repository.CategoryRepository
	articleRepo  repository.ArticleRepository
}

// NewCategoryHandler creates a new category handler instance
func NewCategoryHandler(categoryRepo repository.CategoryRepository, articleRepo repository.ArticleRepository) *CategoryHandler {
	if categoryRepo == nil {
		panic("categoryRepo cannot be nil")
	}
	if articleRepo == nil {
		panic("articleRepo cannot be nil")
	}

	return &CategoryHandler{
		categoryRepo: categoryRepo,
		articleRepo:  articleRepo,
	}
}

// CategoryResponse represents a category in API responses
type CategoryResponse struct {
	ID           uuid.UUID `json:"id"`
	Name         string    `json:"name"`
	Slug         string    `json:"slug"`
	Description  *string   `json:"description,omitempty"`
	Color        string    `json:"color"`
	Icon         *string   `json:"icon,omitempty"`
	ArticleCount *int      `json:"article_count,omitempty"`
}

// List handles GET /v1/categories - returns all categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getCategoryRequestID(ctx)

	// Check if article counts should be included
	includeCounts := r.URL.Query().Get("include_counts") == "true"

	categories, err := h.categoryRepo.List(ctx)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Msg("Failed to list categories")
		response.InternalError(w, "Failed to retrieve categories", requestID)
		return
	}

	categoryResponses := make([]CategoryResponse, len(categories))
	for i, category := range categories {
		categoryResp := toCategoryResponse(category)

		// Optionally include article counts
		if includeCounts {
			count, err := h.getCategoryArticleCount(ctx, category.ID)
			if err != nil {
				log.Warn().
					Err(err).
					Str("category_id", category.ID.String()).
					Msg("Failed to get article count for category")
			} else {
				categoryResp.ArticleCount = &count
			}
		}

		categoryResponses[i] = categoryResp
	}

	response.Success(w, categoryResponses)
}

// GetBySlug handles GET /v1/categories/{slug} - returns a single category by slug
func (h *CategoryHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	requestID := getCategoryRequestID(ctx)

	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.BadRequest(w, "Category slug is required")
		return
	}

	category, err := h.categoryRepo.GetBySlug(ctx, slug)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", requestID).
			Str("slug", slug).
			Msg("Failed to get category by slug")
		response.NotFound(w, "Category not found")
		return
	}

	categoryResp := toCategoryResponse(category)

	// Include article count for detail view
	count, err := h.getCategoryArticleCount(ctx, category.ID)
	if err != nil {
		log.Warn().
			Err(err).
			Str("category_id", category.ID.String()).
			Msg("Failed to get article count for category")
	} else {
		categoryResp.ArticleCount = &count
	}

	response.Success(w, categoryResp)
}

// getCategoryArticleCount retrieves the count of articles for a category
func (h *CategoryHandler) getCategoryArticleCount(ctx context.Context, categoryID uuid.UUID) (int, error) {
	filter := &domain.ArticleFilter{
		CategoryID: &categoryID,
		Page:       1,
		PageSize:   1,
	}

	_, total, err := h.articleRepo.List(ctx, filter)
	if err != nil {
		return 0, err
	}

	return total, nil
}

// toCategoryResponse converts domain category to API response
func toCategoryResponse(category *domain.Category) CategoryResponse {
	if category == nil {
		return CategoryResponse{}
	}

	return CategoryResponse{
		ID:          category.ID,
		Name:        category.Name,
		Slug:        category.Slug,
		Description: category.Description,
		Color:       category.Color,
		Icon:        category.Icon,
	}
}

// getCategoryRequestID extracts request ID from context
func getCategoryRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value("request_id").(string); ok {
		return reqID
	}
	return ""
}
