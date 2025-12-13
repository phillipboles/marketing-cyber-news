package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/domain/entities"
)

// Repository interfaces define contracts for data persistence layer
// Implementations will be in postgres/ and redis/ subdirectories

// UserRepository defines operations for user persistence
// NOTE: Uses entities.User which is the concrete type used by services
type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// ArticleRepository defines operations for article persistence
type ArticleRepository interface {
	Create(ctx context.Context, article *domain.Article) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Article, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Article, error)
	GetBySourceURL(ctx context.Context, sourceURL string) (*domain.Article, error)
	List(ctx context.Context, filter *domain.ArticleFilter) ([]*domain.Article, int, error)
	Update(ctx context.Context, article *domain.Article) error
	Delete(ctx context.Context, id uuid.UUID) error
	IncrementViewCount(ctx context.Context, id uuid.UUID) error
}

// AlertRepository defines operations for alert persistence
type AlertRepository interface {
	Create(ctx context.Context, alert *domain.Alert) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Alert, error)
	GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Alert, error)
	Update(ctx context.Context, alert *domain.Alert) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetActiveAlerts(ctx context.Context) ([]*domain.Alert, error)
}

// AlertMatchRepository defines operations for alert matches
type AlertMatchRepository interface {
	Create(ctx context.Context, match *domain.AlertMatch) error
	GetByAlertID(ctx context.Context, alertID uuid.UUID) ([]*domain.AlertMatch, error)
	MarkNotified(ctx context.Context, id uuid.UUID) error
}

// RefreshTokenRepository defines operations for refresh token management
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *domain.RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	RevokeAllForUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// SessionRepository defines operations for session management (Redis)
type SessionRepository interface {
	Set(ctx context.Context, key string, data interface{}, expiry time.Duration) error
	Get(ctx context.Context, key string) (interface{}, error)
	Delete(ctx context.Context, key string) error
}

// CategoryRepository defines operations for category persistence
type CategoryRepository interface {
	Create(ctx context.Context, category *domain.Category) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	GetBySlug(ctx context.Context, slug string) (*domain.Category, error)
	List(ctx context.Context) ([]*domain.Category, error)
	Update(ctx context.Context, category *domain.Category) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// SourceRepository defines operations for source persistence
type SourceRepository interface {
	Create(ctx context.Context, source *domain.Source) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Source, error)
	GetByURL(ctx context.Context, url string) (*domain.Source, error)
	GetByName(ctx context.Context, name string) (*domain.Source, error)
	List(ctx context.Context, activeOnly bool) ([]*domain.Source, error)
	Update(ctx context.Context, source *domain.Source) error
	Delete(ctx context.Context, id uuid.UUID) error
}

// WebhookLogRepository defines operations for webhook log persistence
type WebhookLogRepository interface {
	Create(ctx context.Context, log *domain.WebhookLog) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.WebhookLog, error)
	Update(ctx context.Context, log *domain.WebhookLog) error
	List(ctx context.Context, limit, offset int) ([]*domain.WebhookLog, error)
}

// AuditLogRepository defines operations for audit log persistence
type AuditLogRepository interface {
	Create(ctx context.Context, log *domain.AuditLog) error
	List(ctx context.Context, filter *domain.AuditLogFilter) ([]*domain.AuditLog, int, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error)
}

// BookmarkRepository defines operations for bookmark persistence
type BookmarkRepository interface {
	Create(ctx context.Context, userID, articleID uuid.UUID) error
	Delete(ctx context.Context, userID, articleID uuid.UUID) error
	IsBookmarked(ctx context.Context, userID, articleID uuid.UUID) (bool, error)
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Article, int, error)
	CountByUserID(ctx context.Context, userID uuid.UUID) (int, error)
}

// ArticleReadRepository defines operations for article read tracking
type ArticleReadRepository interface {
	Create(ctx context.Context, userID, articleID uuid.UUID, readingTimeSeconds int) error
	GetByUserID(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*ArticleRead, int, error)
	GetUserStats(ctx context.Context, userID uuid.UUID) (*UserReadStats, error)
}

// ArticleRead represents an article read record with article details
type ArticleRead struct {
	ID                 uuid.UUID
	UserID             uuid.UUID
	ArticleID          uuid.UUID
	ReadAt             time.Time
	ReadingTimeSeconds int
	Article            *domain.Article
}

// UserReadStats represents user reading statistics
type UserReadStats struct {
	TotalArticlesRead      int
	TotalReadingTime       int
	TotalBookmarks         int
	TotalAlerts            int
	TotalAlertMatches      int
	FavoriteCategory       string
	ArticlesThisWeek       int
	ArticlesThisMonth      int
	AverageReadingTime     float64
}
