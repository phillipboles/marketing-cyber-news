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

// articleReadRepo implements repository.ArticleReadRepository
type articleReadRepo struct {
	db *sql.DB
}

// NewArticleReadRepository creates a new article read repository instance
func NewArticleReadRepository(db *sql.DB) repository.ArticleReadRepository {
	if db == nil {
		panic("db cannot be nil")
	}

	return &articleReadRepo{db: db}
}

// Create records an article read event and increments view count
func (r *articleReadRepo) Create(ctx context.Context, userID, articleID uuid.UUID, readingTimeSeconds int) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID cannot be empty")
	}

	if articleID == uuid.Nil {
		return fmt.Errorf("articleID cannot be empty")
	}

	if readingTimeSeconds < 0 {
		return fmt.Errorf("reading time cannot be negative")
	}

	// Use the database function to record read and increment view count atomically
	query := `SELECT record_article_read($1, $2, $3)`

	var readID uuid.UUID
	err := r.db.QueryRowContext(ctx, query, userID, articleID, readingTimeSeconds).Scan(&readID)
	if err != nil {
		return fmt.Errorf("failed to record article read: %w", err)
	}

	return nil
}

// GetByUserID returns paginated reading history for a user
func (r *articleReadRepo) GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*repository.ArticleRead, int, error) {
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
		FROM article_reads
		WHERE user_id = $1
	`

	var total int
	err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count article reads: %w", err)
	}

	if total == 0 {
		return []*repository.ArticleRead{}, 0, nil
	}

	// Get paginated reads with article details
	query := `
		SELECT
			ar.id, ar.user_id, ar.article_id, ar.read_at, ar.reading_time_seconds,
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
		FROM article_reads ar
		JOIN articles a ON ar.article_id = a.id
		LEFT JOIN categories c ON a.category_id = c.id
		LEFT JOIN sources s ON a.source_id = s.id
		WHERE ar.user_id = $1
		ORDER BY ar.read_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query article reads: %w", err)
	}
	defer rows.Close()

	reads := make([]*repository.ArticleRead, 0)
	for rows.Next() {
		read := &repository.ArticleRead{}

		var article domain.Article
		var category domain.Category
		var source domain.Source
		var iocsJSON []byte
		var ctaJSON []byte

		err := rows.Scan(
			&read.ID,
			&read.UserID,
			&read.ArticleID,
			&read.ReadAt,
			&read.ReadingTimeSeconds,
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
			return nil, 0, fmt.Errorf("failed to scan article read: %w", err)
		}

		// Unmarshal IOCs
		if len(iocsJSON) > 0 {
			if err := json.Unmarshal(iocsJSON, &article.IOCs); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal IOCs: %w", err)
			}
		}

		// Unmarshal ArmorCTA
		if len(ctaJSON) > 0 {
			if err := json.Unmarshal(ctaJSON, &article.ArmorCTA); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal ArmorCTA: %w", err)
			}
		}

		article.Category = &category
		article.Source = &source
		read.Article = &article

		reads = append(reads, read)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("rows iteration error: %w", err)
	}

	return reads, total, nil
}

// GetUserStats returns comprehensive reading statistics for a user
func (r *articleReadRepo) GetUserStats(ctx context.Context, userID uuid.UUID) (*repository.UserReadStats, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("userID cannot be empty")
	}

	// Use the database function for comprehensive stats
	query := `
		SELECT
			total_reads,
			total_bookmarks,
			total_reading_time_seconds,
			avg_reading_time_seconds,
			COALESCE(favorite_category, ''),
			articles_this_week,
			articles_this_month
		FROM get_user_reading_stats($1)
	`

	stats := &repository.UserReadStats{}
	var favoriteCategory sql.NullString

	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&stats.TotalArticlesRead,
		&stats.TotalBookmarks,
		&stats.TotalReadingTime,
		&stats.AverageReadingTime,
		&favoriteCategory,
		&stats.ArticlesThisWeek,
		&stats.ArticlesThisMonth,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	if favoriteCategory.Valid {
		stats.FavoriteCategory = favoriteCategory.String
	}

	// Get alert counts separately (not in the DB function)
	alertQuery := `
		SELECT
			COUNT(*) as total_alerts,
			COALESCE(SUM(
				CASE WHEN EXISTS(
					SELECT 1 FROM alert_matches am WHERE am.alert_id = a.id
				) THEN 1 ELSE 0 END
			), 0) as total_alert_matches
		FROM alerts a
		WHERE a.user_id = $1
	`

	err = r.db.QueryRowContext(ctx, alertQuery, userID).Scan(
		&stats.TotalAlerts,
		&stats.TotalAlertMatches,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get alert stats: %w", err)
	}

	return stats, nil
}
