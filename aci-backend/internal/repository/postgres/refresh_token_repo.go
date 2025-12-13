package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/phillipboles/aci-backend/internal/domain"
	domainerrors "github.com/phillipboles/aci-backend/internal/domain/errors"
)

// RefreshTokenRepository implements repository.RefreshTokenRepository for PostgreSQL
type RefreshTokenRepository struct {
	db *DB
}

// NewRefreshTokenRepository creates a new PostgreSQL refresh token repository
func NewRefreshTokenRepository(db *DB) *RefreshTokenRepository {
	if db == nil {
		panic("database cannot be nil")
	}
	return &RefreshTokenRepository{db: db}
}

// Create inserts a new refresh token into the database
func (r *RefreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	if token == nil {
		return fmt.Errorf("token cannot be nil")
	}

	if err := token.Validate(); err != nil {
		return fmt.Errorf("invalid token: %w", err)
	}

	query := `
		INSERT INTO refresh_tokens (
			id, user_id, token_hash, expires_at, created_at,
			revoked_at, last_used_at, ip_address, user_agent
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		token.ID,
		token.UserID,
		token.Token, // Should be pre-hashed by caller
		token.ExpiresAt,
		token.CreatedAt,
		token.RevokedAt,
		token.LastUsedAt,
		token.IPAddress,
		token.UserAgent,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

// GetByTokenHash retrieves a non-revoked, non-expired refresh token by its hash
func (r *RefreshTokenRepository) GetByTokenHash(ctx context.Context, tokenHash string) (*domain.RefreshToken, error) {
	if tokenHash == "" {
		return nil, fmt.Errorf("token hash cannot be empty")
	}

	query := `
		SELECT
			id, user_id, token_hash, expires_at, created_at,
			revoked_at, last_used_at, ip_address, user_agent
		FROM refresh_tokens
		WHERE token_hash = $1
			AND revoked_at IS NULL
			AND expires_at > NOW()
	`

	var token domain.RefreshToken
	err := r.db.Pool.QueryRow(ctx, query, tokenHash).Scan(
		&token.ID,
		&token.UserID,
		&token.Token, // Actually token_hash from DB
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.RevokedAt,
		&token.LastUsedAt,
		&token.IPAddress,
		&token.UserAgent,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domainerrors.NotFoundError{
				Resource: "refresh_token",
				ID:       "token",
			}
		}
		return nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return &token, nil
}

// Revoke marks a refresh token as revoked by setting revoked_at timestamp
func (r *RefreshTokenRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("token ID cannot be nil")
	}

	now := time.Now()
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE id = $1 AND revoked_at IS NULL
	`

	result, err := r.db.Pool.Exec(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &domainerrors.NotFoundError{
			Resource: "refresh_token",
			ID:       id.String(),
		}
	}

	return nil
}

// RevokeAllForUser marks all refresh tokens for a user as revoked
func (r *RefreshTokenRepository) RevokeAllForUser(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	now := time.Now()
	query := `
		UPDATE refresh_tokens
		SET revoked_at = $2
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := r.db.Pool.Exec(ctx, query, userID, now)
	if err != nil {
		return fmt.Errorf("failed to revoke all tokens for user: %w", err)
	}

	// Note: Not returning error if no rows affected - user may have no active tokens
	return nil
}

// DeleteExpired removes expired refresh tokens from the database
// This should be called periodically by a cleanup job
func (r *RefreshTokenRepository) DeleteExpired(ctx context.Context) error {
	query := `
		DELETE FROM refresh_tokens
		WHERE expires_at < NOW()
	`

	result, err := r.db.Pool.Exec(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to delete expired tokens: %w", err)
	}

	// Log how many tokens were deleted (for monitoring)
	_ = result.RowsAffected()

	return nil
}
