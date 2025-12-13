package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/phillipboles/aci-backend/internal/domain"
	domainerrors "github.com/phillipboles/aci-backend/internal/domain/errors"
)

// AlertMatchRepository implements repository.AlertMatchRepository for PostgreSQL
type AlertMatchRepository struct {
	db *DB
}

// NewAlertMatchRepository creates a new PostgreSQL alert match repository
func NewAlertMatchRepository(db *DB) *AlertMatchRepository {
	if db == nil {
		panic("database cannot be nil")
	}
	return &AlertMatchRepository{db: db}
}

// Create inserts a new alert match into the database
func (r *AlertMatchRepository) Create(ctx context.Context, match *domain.AlertMatch) error {
	if match == nil {
		return fmt.Errorf("alert match cannot be nil")
	}

	if match.ID == uuid.Nil {
		return fmt.Errorf("alert match ID cannot be nil")
	}

	if match.AlertID == uuid.Nil {
		return fmt.Errorf("alert ID cannot be nil")
	}

	if match.ArticleID == uuid.Nil {
		return fmt.Errorf("article ID cannot be nil")
	}

	query := `
		INSERT INTO alert_matches (id, alert_id, article_id, priority, matched_at, notified_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (alert_id, article_id) DO NOTHING
	`

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		match.ID,
		match.AlertID,
		match.ArticleID,
		match.Priority,
		match.MatchedAt,
		match.NotifiedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Foreign key violation (23503)
			if pgErr.Code == "23503" {
				if pgErr.ConstraintName == "alert_matches_alert_id_fkey" {
					return fmt.Errorf("invalid alert ID: %w", domainerrors.ErrNotFound)
				}
				if pgErr.ConstraintName == "alert_matches_article_id_fkey" {
					return fmt.Errorf("invalid article ID: %w", domainerrors.ErrNotFound)
				}
				return fmt.Errorf("invalid reference: %w", domainerrors.ErrNotFound)
			}
		}
		return fmt.Errorf("failed to create alert match: %w", err)
	}

	// Check if row was inserted (not a duplicate)
	if result.RowsAffected() == 0 {
		// Duplicate match - not an error, just skip
		return nil
	}

	return nil
}

// GetByAlertID retrieves all matches for an alert
func (r *AlertMatchRepository) GetByAlertID(ctx context.Context, alertID uuid.UUID) ([]*domain.AlertMatch, error) {
	if alertID == uuid.Nil {
		return nil, fmt.Errorf("alert ID cannot be nil")
	}

	query := `
		SELECT
			id,
			alert_id,
			article_id,
			priority,
			matched_at,
			notified_at
		FROM alert_matches
		WHERE alert_id = $1
		ORDER BY matched_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, alertID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alert matches: %w", err)
	}
	defer rows.Close()

	matches := make([]*domain.AlertMatch, 0)

	for rows.Next() {
		var match domain.AlertMatch
		err := rows.Scan(
			&match.ID,
			&match.AlertID,
			&match.ArticleID,
			&match.Priority,
			&match.MatchedAt,
			&match.NotifiedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert match row: %w", err)
		}

		matches = append(matches, &match)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert match rows: %w", err)
	}

	return matches, nil
}

// MarkNotified marks an alert match as notified
func (r *AlertMatchRepository) MarkNotified(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("alert match ID cannot be nil")
	}

	query := `
		UPDATE alert_matches
		SET notified_at = NOW()
		WHERE id = $1 AND notified_at IS NULL
	`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to mark alert match as notified: %w", err)
	}

	if result.RowsAffected() == 0 {
		// Either not found or already notified
		// Check if match exists
		var exists bool
		checkQuery := `SELECT EXISTS(SELECT 1 FROM alert_matches WHERE id = $1)`
		err := r.db.Pool.QueryRow(ctx, checkQuery, id).Scan(&exists)
		if err != nil {
			return fmt.Errorf("failed to check alert match existence: %w", err)
		}

		if !exists {
			return &domainerrors.NotFoundError{
				Resource: "alert_match",
				ID:       id.String(),
			}
		}

		// Match exists but was already notified - not an error
		return nil
	}

	return nil
}
