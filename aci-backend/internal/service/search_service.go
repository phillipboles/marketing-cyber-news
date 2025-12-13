package service

import (
	"context"
	"fmt"

	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
)

// SearchService handles article search operations
type SearchService struct {
	articleRepo repository.ArticleRepository
}

// NewSearchService creates a new search service instance
func NewSearchService(articleRepo repository.ArticleRepository) *SearchService {
	if articleRepo == nil {
		panic("articleRepo cannot be nil")
	}

	return &SearchService{
		articleRepo: articleRepo,
	}
}

// SearchResult represents a search result with relevance score
type SearchResult struct {
	Article   *domain.Article `json:"article"`
	Score     float64         `json:"score"`
	Highlight string          `json:"highlight,omitempty"`
}

// Search performs full-text search on articles
// Uses PostgreSQL full-text search with ranking
func (s *SearchService) Search(ctx context.Context, query string, filter *domain.ArticleFilter) ([]*SearchResult, int, error) {
	if query == "" {
		return nil, 0, fmt.Errorf("search query cannot be empty")
	}

	if filter == nil {
		filter = domain.NewArticleFilter()
	}

	filter.SearchQuery = &query

	if err := filter.Validate(); err != nil {
		return nil, 0, fmt.Errorf("invalid filter: %w", err)
	}

	articles, total, err := s.articleRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to search articles: %w", err)
	}

	results := make([]*SearchResult, len(articles))
	for i, article := range articles {
		results[i] = &SearchResult{
			Article:   article,
			Score:     1.0, // Repository should provide relevance score
			Highlight: extractHighlight(article, query),
		}
	}

	return results, total, nil
}

// SemanticSearch performs vector similarity search using embeddings
// Falls back to full-text search if embeddings are not available
func (s *SearchService) SemanticSearch(ctx context.Context, query string, limit int) ([]*SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("search query cannot be empty")
	}

	if limit <= 0 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	// TODO: Implement vector similarity search using pgvector
	// For now, fall back to full-text search
	filter := &domain.ArticleFilter{
		SearchQuery: &query,
		Page:        1,
		PageSize:    limit,
	}

	articles, _, err := s.articleRepo.List(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to perform semantic search: %w", err)
	}

	results := make([]*SearchResult, len(articles))
	for i, article := range articles {
		results[i] = &SearchResult{
			Article:   article,
			Score:     1.0,
			Highlight: extractHighlight(article, query),
		}
	}

	return results, nil
}

// extractHighlight creates a text snippet highlighting the search query
func extractHighlight(article *domain.Article, query string) string {
	if article.Summary != nil && *article.Summary != "" {
		return *article.Summary
	}

	// Return first 200 characters of content as fallback
	if len(article.Content) > 200 {
		return article.Content[:200] + "..."
	}

	return article.Content
}
