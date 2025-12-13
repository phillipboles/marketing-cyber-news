package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
)

// bookmarkRepo implements repository.BookmarkRepository
type bookmarkRepo struct {
	db *sql.DB
}

// NewBookmarkRepository creates a new bookmark repository instance
func NewBookmarkRepository(db *sql.DB) repository.BookmarkRepository {
	if db == nil {
		panic("db cannot be nil")
	}

	return &bookmarkRepo{db: db}
}

// Create adds a bookmark for a user (idempotent using ON CONFLICT)
func (r *bookmarkRepo) Create(ctx context.Context, userID, articleID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID cannot be empty")
	}

	if articleID == uuid.Nil {
		return fmt.Errorf("articleID cannot be empty")
	}

	query := `
		INSERT INTO bookmarks (user_id, article_id)
		VALUES ($1, $2)
		ON CONFLICT (user_id, article_id) DO NOTHING
	`

	_, err := r.db.ExecContext(ctx, query, userID, articleID)
	if err != nil {
		return fmt.Errorf("failed to create bookmark: %w", err)
	}

	return nil
}

// Delete removes a bookmark for a user
func (r *bookmarkRepo) Delete(ctx context.Context, userID, articleID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID cannot be empty")
	}

	if articleID == uuid.Nil {
		return fmt.Errorf("articleID cannot be empty")
	}

	query := `
		DELETE FROM bookmarks
		WHERE user_id = $1 AND article_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, userID, articleID)
	if err != nil {
		return fmt.Errorf("failed to delete bookmark: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bookmark not found")
	}

	return nil
}

// IsBookmarked checks if an article is bookmarked by a user
func (r *bookmarkRepo) IsBookmarked(ctx context.Context, userID, articleID uuid.UUID) (bool, error) {
	if userID == uuid.Nil {
		return false, fmt.Errorf("userID cannot be empty")
	}

	if articleID == uuid.Nil {
		return false, fmt.Errorf("articleID cannot be empty")
	}

	query := `
		SELECT EXISTS(
			SELECT 1 FROM bookmarks
			WHERE user_id = $1 AND article_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userID, articleID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check bookmark: %w", err)
	}

	return exists, nil
}

// GetByUserID returns paginated bookmarked articles for a user
func (r *bookmarkRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Article, int, error) {
	if userID == uuid.Nil {
		return nil, 0, fmt.Errorf("userID cannot be empty")
	}

	if limit <= 0 {
		return nil, 0, fmt.Errorf("limit must be positive")
	}

	if offset < 0 {
		return nil, 0, fmt.Errorf("offset cannot be negative")
	}

	// First, get total count
	countQuery := `
		SELECT COUNT(*)
		FROM bookmarks b
		WHERE b.user_id = $1
	`

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count bookmarks: %w", err)
	}

	if total == 0 {
		return []*domain.Article{}, 0, nil
	}

	// Get paginated articles with joins
	query := `
		SELECT
			a.id, a.title, a.slug, a.content, a.summary,
			a.category_id, a.source_id, a.source_url,
			a.severity, a.tags, a.cves, a.vendors,
			a.threat_type, a.attack_vector, a.impact_assessment,
			a.recommended_actions, a.iocs,
			a.armor_relevance, a.armor_cta,
			a.reading_time_minutes, a.view_count,
			a.is_published, a.published_at, a.enriched_at,
			a.created_at, a.updated_at,
			c.id, c.name, c.slug, c.color, c.icon, c.description,
			c.created_at,
			s.id, s.name, s.url, s.description, s.is_active,
			s.trust_score, s.last_scraped_at, s.created_at
		FROM bookmarks b
		JOIN articles a ON b.article_id = a.id
		LEFT JOIN categories c ON a.category_id = c.id
		LEFT JOIN sources s ON a.source_id = s.id
		WHERE b.user_id = $1
		ORDER BY b.created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query bookmarks: %w", err)
	}
	defer rows.Close()

	articles := make([]*domain.Article, 0)
	for rows.Next() {
		article, err := scanArticleWithRelations(rows)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan article: %w", err)
		}
		articles = append(articles, article)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return articles, total, nil
}

// CountByUserID returns the total number of bookmarks for a user
func (r *bookmarkRepo) CountByUserID(ctx context.Context, userID uuid.UUID) (int, error) {
	if userID == uuid.Nil {
		return 0, fmt.Errorf("userID cannot be empty")
	}

	query := `
		SELECT COUNT(*)
		FROM bookmarks
		WHERE user_id = $1
	`

	var count int
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count bookmarks: %w", err)
	}

	return count, nil
}

// scanArticleWithRelations scans an article row with joined category and source
func scanArticleWithRelations(rows *sql.Rows) (*domain.Article, error) {
	article := &domain.Article{}
	category := &domain.Category{}
	source := &domain.Source{}

	var iocsJSON []byte
	var ctaJSON []byte

	err := rows.Scan(
		&article.ID,
		&article.Title,
		&article.Slug,
		&article.Content,
		&article.Summary,
		&article.CategoryID,
		&article.SourceID,
		&article.SourceURL,
		&article.Severity,
		&article.Tags,
		&article.CVEs,
		&article.Vendors,
		&article.ThreatType,
		&article.AttackVector,
		&article.ImpactAssessment,
		&article.RecommendedActions,
		&iocsJSON,
		&article.ArmorRelevance,
		&ctaJSON,
		&article.ReadingTimeMinutes,
		&article.ViewCount,
		&article.IsPublished,
		&article.PublishedAt,
		&article.EnrichedAt,
		&article.CreatedAt,
		&article.UpdatedAt,
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Color,
		&category.Icon,
		&category.Description,
		&category.CreatedAt,
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Description,
		&source.IsActive,
		&source.TrustScore,
		&source.LastScrapedAt,
		&source.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to scan article: %w", err)
	}

	// Unmarshal IOCs
	if len(iocsJSON) > 0 {
		if err := json.Unmarshal(iocsJSON, &article.IOCs); err != nil {
			return nil, fmt.Errorf("failed to unmarshal IOCs: %w", err)
		}
	}

	// Unmarshal ArmorCTA
	if len(ctaJSON) > 0 {
		if err := json.Unmarshal(ctaJSON, &article.ArmorCTA); err != nil {
			return nil, fmt.Errorf("failed to unmarshal ArmorCTA: %w", err)
		}
	}

	article.Category = category
	article.Source = source

	return article, nil
}
