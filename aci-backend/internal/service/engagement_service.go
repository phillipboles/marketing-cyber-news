package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
)

// EngagementService handles user engagement operations (bookmarks, reads, stats)
type EngagementService struct {
	bookmarkRepo    repository.BookmarkRepository
	articleReadRepo repository.ArticleReadRepository
	articleRepo     repository.ArticleRepository
}

// NewEngagementService creates a new engagement service instance
func NewEngagementService(
	bookmarkRepo repository.BookmarkRepository,
	articleReadRepo repository.ArticleReadRepository,
	articleRepo repository.ArticleRepository,
) *EngagementService {
	if bookmarkRepo == nil {
		panic("bookmarkRepo cannot be nil")
	}
	if articleReadRepo == nil {
		panic("articleReadRepo cannot be nil")
	}
	if articleRepo == nil {
		panic("articleRepo cannot be nil")
	}

	return &EngagementService{
		bookmarkRepo:    bookmarkRepo,
		articleReadRepo: articleReadRepo,
		articleRepo:     articleRepo,
	}
}

// AddBookmark bookmarks an article for a user (idempotent)
func (s *EngagementService) AddBookmark(ctx context.Context, userID, articleID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID is required")
	}

	if articleID == uuid.Nil {
		return fmt.Errorf("articleID is required")
	}

	// Verify article exists
	_, err := s.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return fmt.Errorf("article not found: %w", err)
	}

	if err := s.bookmarkRepo.Create(ctx, userID, articleID); err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}

	return nil
}

// RemoveBookmark removes a bookmark
func (s *EngagementService) RemoveBookmark(ctx context.Context, userID, articleID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID is required")
	}

	if articleID == uuid.Nil {
		return fmt.Errorf("articleID is required")
	}

	if err := s.bookmarkRepo.Delete(ctx, userID, articleID); err != nil {
		return fmt.Errorf("failed to remove bookmark: %w", err)
	}

	return nil
}

// GetBookmarks returns paginated bookmarks for a user
func (s *EngagementService) GetBookmarks(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*domain.Article, int, error) {
	if userID == uuid.Nil {
		return nil, 0, fmt.Errorf("userID is required")
	}

	if page < 1 {
		return nil, 0, fmt.Errorf("page must be at least 1")
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, 0, fmt.Errorf("pageSize must be between 1 and 100")
	}

	offset := (page - 1) * pageSize

	articles, total, err := s.bookmarkRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get bookmarks: %w", err)
	}

	return articles, total, nil
}

// IsBookmarked checks if article is bookmarked by user
func (s *EngagementService) IsBookmarked(ctx context.Context, userID, articleID uuid.UUID) (bool, error) {
	if userID == uuid.Nil {
		return false, fmt.Errorf("userID is required")
	}

	if articleID == uuid.Nil {
		return false, fmt.Errorf("articleID is required")
	}

	isBookmarked, err := s.bookmarkRepo.IsBookmarked(ctx, userID, articleID)
	if err != nil {
		return false, fmt.Errorf("failed to check bookmark status: %w", err)
	}

	return isBookmarked, nil
}

// MarkRead records article read and increments view count
func (s *EngagementService) MarkRead(ctx context.Context, userID, articleID uuid.UUID, readingTimeSeconds *int) error {
	if userID == uuid.Nil {
		return fmt.Errorf("userID is required")
	}

	if articleID == uuid.Nil {
		return fmt.Errorf("articleID is required")
	}

	// Verify article exists
	_, err := s.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return fmt.Errorf("article not found: %w", err)
	}

	// Default reading time to 0 if not provided
	readingTime := 0
	if readingTimeSeconds != nil {
		if *readingTimeSeconds < 0 {
			return fmt.Errorf("reading time cannot be negative")
		}
		readingTime = *readingTimeSeconds
	}

	if err := s.articleReadRepo.Create(ctx, userID, articleID, readingTime); err != nil {
		return fmt.Errorf("failed to record article read: %w", err)
	}

	return nil
}

// GetReadingHistory returns paginated reading history
func (s *EngagementService) GetReadingHistory(ctx context.Context, userID uuid.UUID, page, pageSize int) ([]*repository.ArticleRead, int, error) {
	if userID == uuid.Nil {
		return nil, 0, fmt.Errorf("userID is required")
	}

	if page < 1 {
		return nil, 0, fmt.Errorf("page must be at least 1")
	}

	if pageSize < 1 || pageSize > 100 {
		return nil, 0, fmt.Errorf("pageSize must be between 1 and 100")
	}

	offset := (page - 1) * pageSize

	reads, total, err := s.articleReadRepo.GetByUserID(ctx, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get reading history: %w", err)
	}

	return reads, total, nil
}

// GetUserStats returns engagement statistics
func (s *EngagementService) GetUserStats(ctx context.Context, userID uuid.UUID) (*repository.UserReadStats, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("userID is required")
	}

	stats, err := s.articleReadRepo.GetUserStats(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	return stats, nil
}
