package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
	"github.com/phillipboles/aci-backend/internal/util/sanitizer"
	"github.com/phillipboles/aci-backend/internal/util/slug"
)

// ArticleService handles article business logic
type ArticleService struct {
	articleRepo      repository.ArticleRepository
	categoryRepo     repository.CategoryRepository
	sourceRepo       repository.SourceRepository
	webhookLogRepo   repository.WebhookLogRepository
	competitorFilter *CompetitorFilter
	relevanceScorer  *RelevanceScorer
	slugGenerator    *slug.Generator
	sanitizer        *sanitizer.Sanitizer
}

// ArticleCreatedData represents article creation data from webhook
type ArticleCreatedData struct {
	Title          string
	Content        string
	Summary        string
	CategorySlug   string
	Severity       string
	Tags           []string
	SourceURL      string
	SourceName     string
	PublishedAt    string
	CVEs           []string
	Vendors        []string
	SkipEnrichment bool
}

// ArticleUpdatedData represents article update data from webhook
type ArticleUpdatedData struct {
	Title       *string
	Content     *string
	Summary     *string
	Severity    *string
	Tags        []string
	CVEs        []string
	Vendors     []string
	IsPublished *bool
}

// NewArticleService creates a new article service
func NewArticleService(
	articleRepo repository.ArticleRepository,
	categoryRepo repository.CategoryRepository,
	sourceRepo repository.SourceRepository,
	webhookLogRepo repository.WebhookLogRepository,
) *ArticleService {
	return &ArticleService{
		articleRepo:      articleRepo,
		categoryRepo:     categoryRepo,
		sourceRepo:       sourceRepo,
		webhookLogRepo:   webhookLogRepo,
		competitorFilter: NewCompetitorFilter(),
		relevanceScorer:  NewRelevanceScorer(),
		slugGenerator:    slug.NewGenerator(),
		sanitizer:        sanitizer.NewSanitizer(),
	}
}

// CreateArticle creates a new article from webhook data
func (s *ArticleService) CreateArticle(ctx context.Context, data ArticleCreatedData) (*domain.Article, error) {
	// Validate input
	if err := s.validateArticleData(data); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check for duplicate source_url
	existing, err := s.articleRepo.GetBySourceURL(ctx, data.SourceURL)
	if err != nil && !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check for duplicate: %w", err)
	}

	if existing != nil {
		return nil, fmt.Errorf("article with source URL already exists: %s", data.SourceURL)
	}

	// Get category by slug
	category, err := s.categoryRepo.GetBySlug(ctx, data.CategorySlug)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	// Get or create source
	source, err := s.getOrCreateSource(ctx, data.SourceURL, data.SourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to get or create source: %w", err)
	}

	// Generate unique slug
	articleSlug := s.slugGenerator.GenerateUnique(data.Title)

	// Sanitize HTML content
	sanitizedContent := s.sanitizer.SanitizeHTML(data.Content)

	// Parse severity
	severity := domain.Severity(strings.ToLower(data.Severity))
	if !severity.IsValid() {
		severity = domain.SeverityInformational
	}

	// Parse published_at
	var publishedAt time.Time
	if data.PublishedAt != "" {
		publishedAt, err = time.Parse(time.RFC3339, data.PublishedAt)
		if err != nil {
			publishedAt = time.Now()
		}
	} else {
		publishedAt = time.Now()
	}

	// Create article
	now := time.Now()

	// Initialize slices to empty if nil (required for NOT NULL database constraints)
	tags := data.Tags
	if tags == nil {
		tags = []string{}
	}
	cves := data.CVEs
	if cves == nil {
		cves = []string{}
	}
	vendors := data.Vendors
	if vendors == nil {
		vendors = []string{}
	}

	article := &domain.Article{
		ID:                 uuid.New(),
		Title:              data.Title,
		Slug:               articleSlug,
		Content:            sanitizedContent,
		CategoryID:         category.ID,
		SourceID:           source.ID,
		SourceURL:          data.SourceURL,
		Severity:           severity,
		Tags:               tags,
		CVEs:               cves,
		Vendors:            vendors,
		RecommendedActions: []string{},
		IOCs:               []domain.IOC{},
		ReadingTimeMinutes: s.sanitizer.CalculateReadingTime(sanitizedContent),
		ViewCount:          0,
		IsPublished:        true,
		PublishedAt:        publishedAt,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	// Set summary if provided
	if data.Summary != "" {
		article.Summary = &data.Summary
	}

	// Run competitor filter
	article.CompetitorScore, article.IsCompetitorFavorable = s.competitorFilter.Score(
		article.Title,
		article.Content,
	)

	// Calculate Armor relevance score
	article.ArmorRelevance = s.relevanceScorer.Score(article)

	// Generate CTA if relevant
	article.ArmorCTA = s.relevanceScorer.GenerateCTA(article)

	// Validate article
	if err := article.Validate(); err != nil {
		return nil, fmt.Errorf("article validation failed: %w", err)
	}

	// Save to database
	if err := s.articleRepo.Create(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to create article: %w", err)
	}

	return article, nil
}

// UpdateArticle updates an existing article
func (s *ArticleService) UpdateArticle(ctx context.Context, id uuid.UUID, data ArticleUpdatedData) (*domain.Article, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("article ID is required")
	}

	// Get existing article
	article, err := s.articleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	// Update fields if provided
	if data.Title != nil {
		article.Title = *data.Title
		article.Slug = s.slugGenerator.GenerateUnique(*data.Title)
	}

	if data.Content != nil {
		article.Content = s.sanitizer.SanitizeHTML(*data.Content)
		article.ReadingTimeMinutes = s.sanitizer.CalculateReadingTime(article.Content)
	}

	if data.Summary != nil {
		article.Summary = data.Summary
	}

	if data.Severity != nil {
		severity := domain.Severity(strings.ToLower(*data.Severity))
		if severity.IsValid() {
			article.Severity = severity
		}
	}

	if len(data.Tags) > 0 {
		article.Tags = data.Tags
	}

	if len(data.CVEs) > 0 {
		article.CVEs = data.CVEs
	}

	if len(data.Vendors) > 0 {
		article.Vendors = data.Vendors
	}

	if data.IsPublished != nil {
		article.IsPublished = *data.IsPublished
	}

	// Recalculate scores
	article.CompetitorScore, article.IsCompetitorFavorable = s.competitorFilter.Score(
		article.Title,
		article.Content,
	)
	article.ArmorRelevance = s.relevanceScorer.Score(article)
	article.ArmorCTA = s.relevanceScorer.GenerateCTA(article)

	// Update timestamp
	article.UpdatedAt = time.Now()

	// Validate
	if err := article.Validate(); err != nil {
		return nil, fmt.Errorf("article validation failed: %w", err)
	}

	// Update in database
	if err := s.articleRepo.Update(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to update article: %w", err)
	}

	return article, nil
}

// DeleteArticle soft deletes an article
func (s *ArticleService) DeleteArticle(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("article ID is required")
	}

	if err := s.articleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	return nil
}

// BulkImport processes multiple articles
func (s *ArticleService) BulkImport(ctx context.Context, articles []ArticleCreatedData) (int, []error) {
	successCount := 0
	errors := make([]error, 0)

	for i, data := range articles {
		_, err := s.CreateArticle(ctx, data)
		if err != nil {
			errors = append(errors, fmt.Errorf("article %d: %w", i, err))
		} else {
			successCount++
		}
	}

	return successCount, errors
}

// validateArticleData validates article creation data
func (s *ArticleService) validateArticleData(data ArticleCreatedData) error {
	if data.Title == "" {
		return fmt.Errorf("title is required")
	}

	if data.Content == "" {
		return fmt.Errorf("content is required")
	}

	if data.CategorySlug == "" {
		return fmt.Errorf("category_slug is required")
	}

	if data.SourceURL == "" {
		return fmt.Errorf("source_url is required")
	}

	return nil
}

// getOrCreateSource gets an existing source or creates a new one
func (s *ArticleService) getOrCreateSource(ctx context.Context, sourceURL, sourceName string) (*domain.Source, error) {
	// Try to get existing source by URL first
	source, err := s.sourceRepo.GetByURL(ctx, sourceURL)
	if err == nil {
		return source, nil
	}

	// If error is not "not found", return error
	if !strings.Contains(err.Error(), "not found") {
		return nil, fmt.Errorf("failed to check for existing source by URL: %w", err)
	}

	// Try to get existing source by name if name is provided
	if sourceName != "" {
		source, err = s.sourceRepo.GetByName(ctx, sourceName)
		if err == nil {
			return source, nil
		}
		if !strings.Contains(err.Error(), "not found") {
			return nil, fmt.Errorf("failed to check for existing source by name: %w", err)
		}
	}

	// Create new source
	if sourceName == "" {
		sourceName = sourceURL
	}

	newSource := &domain.Source{
		ID:         uuid.New(),
		Name:       sourceName,
		URL:        sourceURL,
		IsActive:   true,
		TrustScore: 0.5,
		CreatedAt:  time.Now(),
	}

	if err := s.sourceRepo.Create(ctx, newSource); err != nil {
		// Check if it was created by another goroutine (race condition)
		existing, getErr := s.sourceRepo.GetByURL(ctx, sourceURL)
		if getErr == nil {
			return existing, nil
		}
		// Also try by name
		existing, getErr = s.sourceRepo.GetByName(ctx, sourceName)
		if getErr == nil {
			return existing, nil
		}
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	return newSource, nil
}
