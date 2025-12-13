package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/phillipboles/aci-backend/internal/domain"
	domainerrors "github.com/phillipboles/aci-backend/internal/domain/errors"
)

// AlertRepository implements repository.AlertRepository for PostgreSQL
type AlertRepository struct {
	db *DB
}

// NewAlertRepository creates a new PostgreSQL alert repository
func NewAlertRepository(db *DB) *AlertRepository {
	if db == nil {
		panic("database cannot be nil")
	}
	return &AlertRepository{db: db}
}

// Create inserts a new alert into the database
func (r *AlertRepository) Create(ctx context.Context, alert *domain.Alert) error {
	if alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}

	if alert.ID == uuid.Nil {
		return fmt.Errorf("alert ID cannot be nil")
	}

	if alert.UserID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	query := `
		INSERT INTO alerts (id, user_id, name, type, value, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		alert.ID,
		alert.UserID,
		alert.Name,
		alert.Type,
		alert.Value,
		alert.IsActive,
		alert.CreatedAt,
		alert.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Unique constraint violation (23505)
			if pgErr.Code == "23505" {
				return fmt.Errorf("alert already exists: %w", domainerrors.ErrConflict)
			}
			// Foreign key violation (23503)
			if pgErr.Code == "23503" {
				return fmt.Errorf("invalid user ID: %w", domainerrors.ErrNotFound)
			}
		}
		return fmt.Errorf("failed to create alert: %w", err)
	}

	return nil
}

// GetByID retrieves an alert by its ID
func (r *AlertRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Alert, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("alert ID cannot be nil")
	}

	query := `
		SELECT
			a.id,
			a.user_id,
			a.name,
			a.type,
			a.value,
			a.is_active,
			a.created_at,
			a.updated_at,
			COALESCE(COUNT(am.id), 0) as match_count
		FROM alerts a
		LEFT JOIN alert_matches am ON a.id = am.alert_id
		WHERE a.id = $1
		GROUP BY a.id, a.user_id, a.name, a.type, a.value, a.is_active, a.created_at, a.updated_at
	`

	var alert domain.Alert
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&alert.ID,
		&alert.UserID,
		&alert.Name,
		&alert.Type,
		&alert.Value,
		&alert.IsActive,
		&alert.CreatedAt,
		&alert.UpdatedAt,
		&alert.MatchCount,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domainerrors.NotFoundError{
				Resource: "alert",
				ID:       id.String(),
			}
		}
		return nil, fmt.Errorf("failed to get alert by ID: %w", err)
	}

	return &alert, nil
}

// GetByUserID retrieves all alerts for a user with match counts
func (r *AlertRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Alert, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	query := `
		SELECT
			a.id,
			a.user_id,
			a.name,
			a.type,
			a.value,
			a.is_active,
			a.created_at,
			a.updated_at,
			COALESCE(COUNT(am.id), 0) as match_count
		FROM alerts a
		LEFT JOIN alert_matches am ON a.id = am.alert_id
		WHERE a.user_id = $1
		GROUP BY a.id, a.user_id, a.name, a.type, a.value, a.is_active, a.created_at, a.updated_at
		ORDER BY a.created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query alerts by user ID: %w", err)
	}
	defer rows.Close()

	alerts := make([]*domain.Alert, 0)

	for rows.Next() {
		var alert domain.Alert
		err := rows.Scan(
			&alert.ID,
			&alert.UserID,
			&alert.Name,
			&alert.Type,
			&alert.Value,
			&alert.IsActive,
			&alert.CreatedAt,
			&alert.UpdatedAt,
			&alert.MatchCount,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert row: %w", err)
		}

		alerts = append(alerts, &alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert rows: %w", err)
	}

	return alerts, nil
}

// Update updates an existing alert
func (r *AlertRepository) Update(ctx context.Context, alert *domain.Alert) error {
	if alert == nil {
		return fmt.Errorf("alert cannot be nil")
	}

	if alert.ID == uuid.Nil {
		return fmt.Errorf("alert ID cannot be nil")
	}

	query := `
		UPDATE alerts
		SET name = $2, value = $3, is_active = $4, updated_at = $5
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		alert.ID,
		alert.Name,
		alert.Value,
		alert.IsActive,
		alert.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &domainerrors.NotFoundError{
			Resource: "alert",
			ID:       alert.ID.String(),
		}
	}

	return nil
}

// Delete removes an alert from the database
func (r *AlertRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("alert ID cannot be nil")
	}

	query := `DELETE FROM alerts WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete alert: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &domainerrors.NotFoundError{
			Resource: "alert",
			ID:       id.String(),
		}
	}

	return nil
}

// GetActiveAlerts retrieves all active alerts across all users
func (r *AlertRepository) GetActiveAlerts(ctx context.Context) ([]*domain.Alert, error) {
	query := `
		SELECT
			id,
			user_id,
			name,
			type,
			value,
			is_active,
			created_at,
			updated_at
		FROM alerts
		WHERE is_active = true
		ORDER BY created_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query active alerts: %w", err)
	}
	defer rows.Close()

	alerts := make([]*domain.Alert, 0)

	for rows.Next() {
		var alert domain.Alert
		err := rows.Scan(
			&alert.ID,
			&alert.UserID,
			&alert.Name,
			&alert.Type,
			&alert.Value,
			&alert.IsActive,
			&alert.CreatedAt,
			&alert.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan alert row: %w", err)
		}

		alerts = append(alerts, &alert)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating alert rows: %w", err)
	}

	return alerts, nil
}
