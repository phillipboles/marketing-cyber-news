package service

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"

	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/domain/entities"
	"github.com/phillipboles/aci-backend/internal/repository"
)

// AdminService handles admin-only business logic
type AdminService struct {
	articleRepo  repository.ArticleRepository
	sourceRepo   repository.SourceRepository
	userRepo     repository.UserRepository
	auditLogRepo repository.AuditLogRepository
}

// NewAdminService creates a new admin service instance
func NewAdminService(
	articleRepo repository.ArticleRepository,
	sourceRepo repository.SourceRepository,
	userRepo repository.UserRepository,
	auditLogRepo repository.AuditLogRepository,
) *AdminService {
	if articleRepo == nil {
		panic("articleRepo cannot be nil")
	}
	if sourceRepo == nil {
		panic("sourceRepo cannot be nil")
	}
	if userRepo == nil {
		panic("userRepo cannot be nil")
	}
	if auditLogRepo == nil {
		panic("auditLogRepo cannot be nil")
	}

	return &AdminService{
		articleRepo:  articleRepo,
		sourceRepo:   sourceRepo,
		userRepo:     userRepo,
		auditLogRepo: auditLogRepo,
	}
}

// UpdateArticle updates an article (admin-only)
func (s *AdminService) UpdateArticle(
	ctx context.Context,
	articleID uuid.UUID,
	updates map[string]interface{},
	adminUserID uuid.UUID,
	ipAddress, userAgent string,
) (*domain.Article, error) {
	if articleID == uuid.Nil {
		return nil, fmt.Errorf("article ID is required")
	}

	if adminUserID == uuid.Nil {
		return nil, fmt.Errorf("admin user ID is required")
	}

	// Get existing article
	article, err := s.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get article: %w", err)
	}

	// Store old state for audit log
	oldState, err := articleToMap(article)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize old state: %w", err)
	}

	// Apply updates to article
	if err := applyArticleUpdates(article, updates); err != nil {
		return nil, fmt.Errorf("failed to apply updates: %w", err)
	}

	// Validate updated article
	if err := article.Validate(); err != nil {
		return nil, fmt.Errorf("invalid article: %w", err)
	}

	// Update article in database
	if err := s.articleRepo.Update(ctx, article); err != nil {
		return nil, fmt.Errorf("failed to update article: %w", err)
	}

	// Store new state for audit log
	newState, err := articleToMap(article)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize new state: %w", err)
	}

	// Log audit event
	if err := s.LogAuditEvent(
		ctx,
		&adminUserID,
		"update_article",
		"article",
		&articleID,
		oldState,
		newState,
		&ipAddress,
		&userAgent,
	); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("failed to log audit event: %v\n", err)
	}

	return article, nil
}

// DeleteArticle permanently deletes an article (admin-only)
func (s *AdminService) DeleteArticle(
	ctx context.Context,
	articleID uuid.UUID,
	adminUserID uuid.UUID,
	ipAddress, userAgent string,
) error {
	if articleID == uuid.Nil {
		return fmt.Errorf("article ID is required")
	}

	if adminUserID == uuid.Nil {
		return fmt.Errorf("admin user ID is required")
	}

	// Get article for audit log
	article, err := s.articleRepo.GetByID(ctx, articleID)
	if err != nil {
		return fmt.Errorf("failed to get article: %w", err)
	}

	oldState, err := articleToMap(article)
	if err != nil {
		return fmt.Errorf("failed to serialize article state: %w", err)
	}

	// Delete article
	if err := s.articleRepo.Delete(ctx, articleID); err != nil {
		return fmt.Errorf("failed to delete article: %w", err)
	}

	// Log audit event
	if err := s.LogAuditEvent(
		ctx,
		&adminUserID,
		"delete_article",
		"article",
		&articleID,
		oldState,
		nil,
		&ipAddress,
		&userAgent,
	); err != nil {
		fmt.Printf("failed to log audit event: %v\n", err)
	}

	return nil
}

// ListSources lists all sources including inactive ones (admin-only)
func (s *AdminService) ListSources(ctx context.Context) ([]*domain.Source, error) {
	sources, err := s.sourceRepo.List(ctx, false)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}

	return sources, nil
}

// CreateSource creates a new source (admin-only)
func (s *AdminService) CreateSource(
	ctx context.Context,
	source *domain.Source,
	adminUserID uuid.UUID,
	ipAddress, userAgent string,
) (*domain.Source, error) {
	if source == nil {
		return nil, fmt.Errorf("source cannot be nil")
	}

	if adminUserID == uuid.Nil {
		return nil, fmt.Errorf("admin user ID is required")
	}

	if err := source.Validate(); err != nil {
		return nil, fmt.Errorf("invalid source: %w", err)
	}

	// Check for duplicate URL
	existing, _ := s.sourceRepo.GetByURL(ctx, source.URL)
	if existing != nil {
		return nil, fmt.Errorf("source with URL %s already exists", source.URL)
	}

	// Create source
	if err := s.sourceRepo.Create(ctx, source); err != nil {
		return nil, fmt.Errorf("failed to create source: %w", err)
	}

	// Log audit event
	newState, err := sourceToMap(source)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize source state: %w", err)
	}

	if err := s.LogAuditEvent(
		ctx,
		&adminUserID,
		"create_source",
		"source",
		&source.ID,
		nil,
		newState,
		&ipAddress,
		&userAgent,
	); err != nil {
		fmt.Printf("failed to log audit event: %v\n", err)
	}

	return source, nil
}

// UpdateSource updates a source (admin-only)
func (s *AdminService) UpdateSource(
	ctx context.Context,
	sourceID uuid.UUID,
	updates map[string]interface{},
	adminUserID uuid.UUID,
	ipAddress, userAgent string,
) (*domain.Source, error) {
	if sourceID == uuid.Nil {
		return nil, fmt.Errorf("source ID is required")
	}

	if adminUserID == uuid.Nil {
		return nil, fmt.Errorf("admin user ID is required")
	}

	// Get existing source
	source, err := s.sourceRepo.GetByID(ctx, sourceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	oldState, err := sourceToMap(source)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize old state: %w", err)
	}

	// Apply updates
	if err := applySourceUpdates(source, updates); err != nil {
		return nil, fmt.Errorf("failed to apply updates: %w", err)
	}

	// Validate updated source
	if err := source.Validate(); err != nil {
		return nil, fmt.Errorf("invalid source: %w", err)
	}

	// Update source
	if err := s.sourceRepo.Update(ctx, source); err != nil {
		return nil, fmt.Errorf("failed to update source: %w", err)
	}

	newState, err := sourceToMap(source)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize new state: %w", err)
	}

	// Log audit event
	if err := s.LogAuditEvent(
		ctx,
		&adminUserID,
		"update_source",
		"source",
		&sourceID,
		oldState,
		newState,
		&ipAddress,
		&userAgent,
	); err != nil {
		fmt.Printf("failed to log audit event: %v\n", err)
	}

	return source, nil
}

// DeleteSource deactivates a source (soft delete, admin-only)
func (s *AdminService) DeleteSource(
	ctx context.Context,
	sourceID uuid.UUID,
	adminUserID uuid.UUID,
	ipAddress, userAgent string,
) error {
	if sourceID == uuid.Nil {
		return fmt.Errorf("source ID is required")
	}

	if adminUserID == uuid.Nil {
		return fmt.Errorf("admin user ID is required")
	}

	// Get source
	source, err := s.sourceRepo.GetByID(ctx, sourceID)
	if err != nil {
		return fmt.Errorf("failed to get source: %w", err)
	}

	oldState, err := sourceToMap(source)
	if err != nil {
		return fmt.Errorf("failed to serialize source state: %w", err)
	}

	// Soft delete by deactivating
	source.Deactivate()
	if err := s.sourceRepo.Update(ctx, source); err != nil {
		return fmt.Errorf("failed to deactivate source: %w", err)
	}

	newState, err := sourceToMap(source)
	if err != nil {
		return fmt.Errorf("failed to serialize source state: %w", err)
	}

	// Log audit event
	if err := s.LogAuditEvent(
		ctx,
		&adminUserID,
		"delete_source",
		"source",
		&sourceID,
		oldState,
		newState,
		&ipAddress,
		&userAgent,
	); err != nil {
		fmt.Printf("failed to log audit event: %v\n", err)
	}

	return nil
}

// ListUsers lists all users with pagination (admin-only)
func (s *AdminService) ListUsers(ctx context.Context, limit, offset int) ([]*entities.User, int, error) {
	if limit < 0 {
		return nil, 0, fmt.Errorf("limit must be non-negative")
	}

	if offset < 0 {
		return nil, 0, fmt.Errorf("offset must be non-negative")
	}

	// Since UserRepository doesn't have a List method, we'll need to implement one
	// For now, return an error indicating this needs implementation
	return nil, 0, fmt.Errorf("ListUsers not yet implemented - UserRepository needs List method")
}

// UpdateUser updates a user (admin-only)
func (s *AdminService) UpdateUser(
	ctx context.Context,
	userID uuid.UUID,
	updates map[string]interface{},
	adminUserID uuid.UUID,
	ipAddress, userAgent string,
) (*entities.User, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	if adminUserID == uuid.Nil {
		return nil, fmt.Errorf("admin user ID is required")
	}

	// Get existing user
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	oldState, err := userToMap(user)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize old state: %w", err)
	}

	// Apply updates
	if err := applyUserUpdates(user, updates); err != nil {
		return nil, fmt.Errorf("failed to apply updates: %w", err)
	}

	// Update user
	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	newState, err := userToMap(user)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize new state: %w", err)
	}

	// Log audit event
	if err := s.LogAuditEvent(
		ctx,
		&adminUserID,
		"update_user",
		"user",
		&userID,
		oldState,
		newState,
		&ipAddress,
		&userAgent,
	); err != nil {
		fmt.Printf("failed to log audit event: %v\n", err)
	}

	return user, nil
}

// DeleteUser disables a user account (admin-only)
func (s *AdminService) DeleteUser(
	ctx context.Context,
	userID uuid.UUID,
	adminUserID uuid.UUID,
	ipAddress, userAgent string,
) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if adminUserID == uuid.Nil {
		return fmt.Errorf("admin user ID is required")
	}

	if userID == adminUserID {
		return fmt.Errorf("cannot delete your own account")
	}

	// Get user for audit log
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	oldState, err := userToMap(user)
	if err != nil {
		return fmt.Errorf("failed to serialize user state: %w", err)
	}

	// Delete user
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	// Log audit event
	if err := s.LogAuditEvent(
		ctx,
		&adminUserID,
		"delete_user",
		"user",
		&userID,
		oldState,
		nil,
		&ipAddress,
		&userAgent,
	); err != nil {
		fmt.Printf("failed to log audit event: %v\n", err)
	}

	return nil
}

// ListAuditLogs lists audit logs with filtering (admin-only)
func (s *AdminService) ListAuditLogs(ctx context.Context, filter *domain.AuditLogFilter) ([]*domain.AuditLog, int, error) {
	if filter == nil {
		return nil, 0, fmt.Errorf("filter cannot be nil")
	}

	logs, totalCount, err := s.auditLogRepo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list audit logs: %w", err)
	}

	return logs, totalCount, nil
}

// LogAuditEvent logs an admin action to the audit trail
func (s *AdminService) LogAuditEvent(
	ctx context.Context,
	userID *uuid.UUID,
	action string,
	resourceType string,
	resourceID *uuid.UUID,
	oldValue interface{},
	newValue interface{},
	ipAddress *string,
	userAgent *string,
) error {
	auditLog := domain.NewAuditLog(
		userID,
		action,
		resourceType,
		resourceID,
		oldValue,
		newValue,
		ipAddress,
		userAgent,
	)

	if err := s.auditLogRepo.Create(ctx, auditLog); err != nil {
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// Helper functions to convert domain objects to maps for audit logging

func articleToMap(article *domain.Article) (map[string]interface{}, error) {
	data, err := json.Marshal(article)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func sourceToMap(source *domain.Source) (map[string]interface{}, error) {
	data, err := json.Marshal(source)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	return result, nil
}

func userToMap(user *entities.User) (map[string]interface{}, error) {
	data, err := json.Marshal(user)
	if err != nil {
		return nil, err
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	// Remove sensitive fields
	delete(result, "password_hash")

	return result, nil
}

func applyArticleUpdates(article *domain.Article, updates map[string]interface{}) error {
	for key, value := range updates {
		switch key {
		case "severity":
			if severityStr, ok := value.(string); ok {
				article.Severity = domain.Severity(severityStr)
			}
		case "is_published":
			if isPublished, ok := value.(bool); ok {
				article.IsPublished = isPublished
			}
		case "title":
			if title, ok := value.(string); ok {
				article.Title = title
			}
		case "summary":
			if summary, ok := value.(string); ok {
				article.Summary = &summary
			}
		case "content":
			if content, ok := value.(string); ok {
				article.Content = content
			}
		default:
			return fmt.Errorf("unsupported field: %s", key)
		}
	}
	return nil
}

func applySourceUpdates(source *domain.Source, updates map[string]interface{}) error {
	for key, value := range updates {
		switch key {
		case "name":
			if name, ok := value.(string); ok {
				source.Name = name
			}
		case "url":
			if url, ok := value.(string); ok {
				source.URL = url
			}
		case "description":
			if desc, ok := value.(string); ok {
				source.Description = &desc
			}
		case "is_active":
			if isActive, ok := value.(bool); ok {
				source.IsActive = isActive
			}
		case "trust_score":
			if score, ok := value.(float64); ok {
				if err := source.UpdateTrustScore(score); err != nil {
					return err
				}
			}
		default:
			return fmt.Errorf("unsupported field: %s", key)
		}
	}
	return nil
}

func applyUserUpdates(user *entities.User, updates map[string]interface{}) error {
	for key, value := range updates {
		switch key {
		case "role":
			if roleStr, ok := value.(string); ok {
				user.Role = entities.UserRole(roleStr)
			}
		case "email_verified":
			if verified, ok := value.(bool); ok {
				user.EmailVerified = verified
			}
		case "email":
			if email, ok := value.(string); ok {
				user.Email = email
			}
		case "name":
			if name, ok := value.(string); ok {
				user.Name = name
			}
		default:
			return fmt.Errorf("unsupported field: %s", key)
		}
	}
	return nil
}
