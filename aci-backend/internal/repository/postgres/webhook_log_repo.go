package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/repository"
)

type webhookLogRepository struct {
	db *DB
}

// NewWebhookLogRepository creates a new PostgreSQL webhook log repository
func NewWebhookLogRepository(db *DB) repository.WebhookLogRepository {
	if db == nil {
		panic("database cannot be nil")
	}
	return &webhookLogRepository{db: db}
}

// Create creates a new webhook log entry
func (r *webhookLogRepository) Create(ctx context.Context, log *domain.WebhookLog) error {
	if log == nil {
		return fmt.Errorf("webhook log cannot be nil")
	}

	if err := log.Validate(); err != nil {
		return fmt.Errorf("invalid webhook log: %w", err)
	}

	query := `
		INSERT INTO webhook_logs (id, event_type, status, payload, workflow_id, execution_id, error_msg, processed_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		log.ID,
		log.EventType,
		log.Status,
		log.Payload,
		log.WorkflowID,
		log.ExecutionID,
		log.ErrorMsg,
		log.ProcessedAt,
		log.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create webhook log: %w", err)
	}

	return nil
}

// GetByID retrieves a webhook log by ID
func (r *webhookLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WebhookLog, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("webhook log ID cannot be nil")
	}

	query := `
		SELECT id, event_type, status, payload, workflow_id, execution_id, error_msg, processed_at, created_at
		FROM webhook_logs
		WHERE id = $1
	`

	log := &domain.WebhookLog{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&log.ID,
		&log.EventType,
		&log.Status,
		&log.Payload,
		&log.WorkflowID,
		&log.ExecutionID,
		&log.ErrorMsg,
		&log.ProcessedAt,
		&log.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("webhook log not found: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get webhook log: %w", err)
	}

	return log, nil
}

// Update updates an existing webhook log
func (r *webhookLogRepository) Update(ctx context.Context, log *domain.WebhookLog) error {
	if log == nil {
		return fmt.Errorf("webhook log cannot be nil")
	}

	if err := log.Validate(); err != nil {
		return fmt.Errorf("invalid webhook log: %w", err)
	}

	query := `
		UPDATE webhook_logs
		SET status = $2, error_msg = $3, processed_at = $4
		WHERE id = $1
	`

	cmdTag, err := r.db.Pool.Exec(ctx, query,
		log.ID,
		log.Status,
		log.ErrorMsg,
		log.ProcessedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update webhook log: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("webhook log not found")
	}

	return nil
}

// List retrieves webhook logs with pagination
func (r *webhookLogRepository) List(ctx context.Context, limit, offset int) ([]*domain.WebhookLog, error) {
	if limit <= 0 {
		return nil, fmt.Errorf("limit must be positive")
	}

	if offset < 0 {
		return nil, fmt.Errorf("offset cannot be negative")
	}

	query := `
		SELECT id, event_type, status, payload, workflow_id, execution_id, error_msg, processed_at, created_at
		FROM webhook_logs
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list webhook logs: %w", err)
	}
	defer rows.Close()

	logs := make([]*domain.WebhookLog, 0)
	for rows.Next() {
		log := &domain.WebhookLog{}
		err := rows.Scan(
			&log.ID,
			&log.EventType,
			&log.Status,
			&log.Payload,
			&log.WorkflowID,
			&log.ExecutionID,
			&log.ErrorMsg,
			&log.ProcessedAt,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan webhook log: %w", err)
		}
		logs = append(logs, log)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating webhook logs: %w", err)
	}

	return logs, nil
}
