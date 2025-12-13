package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/phillipboles/aci-backend/internal/domain"
)

// AuditLogRepository implements repository.AuditLogRepository interface
type AuditLogRepository struct {
	db *sql.DB
}

// NewAuditLogRepository creates a new audit log repository instance
func NewAuditLogRepository(db *sql.DB) *AuditLogRepository {
	if db == nil {
		panic("db cannot be nil")
	}

	return &AuditLogRepository{db: db}
}

// Create inserts a new audit log entry
func (r *AuditLogRepository) Create(ctx context.Context, log *domain.AuditLog) error {
	if log == nil {
		return fmt.Errorf("audit log cannot be nil")
	}

	if err := log.Validate(); err != nil {
		return fmt.Errorf("invalid audit log: %w", err)
	}

	query := `
		INSERT INTO audit_logs (
			id,
			user_id,
			action,
			resource_type,
			resource_id,
			old_value,
			new_value,
			ip_address,
			user_agent,
			created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`

	// Marshal old_value and new_value to JSON
	var oldValueJSON, newValueJSON []byte
	var err error

	if log.OldValue != nil {
		oldValueJSON, err = json.Marshal(log.OldValue)
		if err != nil {
			return fmt.Errorf("failed to marshal old_value: %w", err)
		}
	}

	if log.NewValue != nil {
		newValueJSON, err = json.Marshal(log.NewValue)
		if err != nil {
			return fmt.Errorf("failed to marshal new_value: %w", err)
		}
	}

	_, err = r.db.ExecContext(
		ctx,
		query,
		log.ID,
		log.UserID,
		log.Action,
		log.ResourceType,
		log.ResourceID,
		oldValueJSON,
		newValueJSON,
		log.IPAddress,
		log.UserAgent,
		log.CreatedAt,
	)

	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code {
			case "23503": // Foreign key violation
				return fmt.Errorf("user not found: %w", err)
			}
		}
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// List retrieves audit logs with filtering and pagination
func (r *AuditLogRepository) List(ctx context.Context, filter *domain.AuditLogFilter) ([]*domain.AuditLog, int, error) {
	if filter == nil {
		return nil, 0, fmt.Errorf("filter cannot be nil")
	}

	if err := filter.Validate(); err != nil {
		return nil, 0, fmt.Errorf("invalid filter: %w", err)
	}

	// Build WHERE clauses
	whereClauses := []string{}
	args := []interface{}{}
	argCount := 1

	if filter.UserID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.user_id = $%d", argCount))
		args = append(args, *filter.UserID)
		argCount++
	}

	if filter.Action != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.action = $%d", argCount))
		args = append(args, *filter.Action)
		argCount++
	}

	if filter.ResourceType != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.resource_type = $%d", argCount))
		args = append(args, *filter.ResourceType)
		argCount++
	}

	if filter.ResourceID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.resource_id = $%d", argCount))
		args = append(args, *filter.ResourceID)
		argCount++
	}

	if filter.StartDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.created_at >= $%d", argCount))
		args = append(args, *filter.StartDate)
		argCount++
	}

	if filter.EndDate != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("al.created_at <= $%d", argCount))
		args = append(args, *filter.EndDate)
		argCount++
	}

	whereClause := ""
	if len(whereClauses) > 0 {
		whereClause = "WHERE " + whereClauses[0]
		for _, clause := range whereClauses[1:] {
			whereClause += " AND " + clause
		}
	}

	// Count total matching records
	countQuery := `
		SELECT COUNT(*)
		FROM audit_logs al
		` + whereClause

	var totalCount int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalCount); err != nil {
		return nil, 0, fmt.Errorf("failed to count audit logs: %w", err)
	}

	// Retrieve paginated results with user email
	query := `
		SELECT
			al.id,
			al.user_id,
			u.email,
			al.action,
			al.resource_type,
			al.resource_id,
			al.old_value,
			al.new_value,
			al.ip_address,
			al.user_agent,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		` + whereClause + `
		ORDER BY al.created_at DESC
		LIMIT $` + fmt.Sprintf("%d", argCount) + ` OFFSET $` + fmt.Sprintf("%d", argCount+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	logs := make([]*domain.AuditLog, 0)
	for rows.Next() {
		log := &domain.AuditLog{}
		var oldValueJSON, newValueJSON []byte

		err := rows.Scan(
			&log.ID,
			&log.UserID,
			&log.UserEmail,
			&log.Action,
			&log.ResourceType,
			&log.ResourceID,
			&oldValueJSON,
			&newValueJSON,
			&log.IPAddress,
			&log.UserAgent,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan audit log: %w", err)
		}

		// Unmarshal JSON values
		if oldValueJSON != nil {
			if err := json.Unmarshal(oldValueJSON, &log.OldValue); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal old_value: %w", err)
			}
		}

		if newValueJSON != nil {
			if err := json.Unmarshal(newValueJSON, &log.NewValue); err != nil {
				return nil, 0, fmt.Errorf("failed to unmarshal new_value: %w", err)
			}
		}

		logs = append(logs, log)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating audit logs: %w", err)
	}

	return logs, totalCount, nil
}

// GetByID retrieves a single audit log by ID
func (r *AuditLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.AuditLog, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("audit log ID is required")
	}

	query := `
		SELECT
			al.id,
			al.user_id,
			u.email,
			al.action,
			al.resource_type,
			al.resource_id,
			al.old_value,
			al.new_value,
			al.ip_address,
			al.user_agent,
			al.created_at
		FROM audit_logs al
		LEFT JOIN users u ON al.user_id = u.id
		WHERE al.id = $1
	`

	log := &domain.AuditLog{}
	var oldValueJSON, newValueJSON []byte

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&log.UserID,
		&log.UserEmail,
		&log.Action,
		&log.ResourceType,
		&log.ResourceID,
		&oldValueJSON,
		&newValueJSON,
		&log.IPAddress,
		&log.UserAgent,
		&log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("audit log not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get audit log: %w", err)
	}

	// Unmarshal JSON values
	if oldValueJSON != nil {
		if err := json.Unmarshal(oldValueJSON, &log.OldValue); err != nil {
			return nil, fmt.Errorf("failed to unmarshal old_value: %w", err)
		}
	}

	if newValueJSON != nil {
		if err := json.Unmarshal(newValueJSON, &log.NewValue); err != nil {
			return nil, fmt.Errorf("failed to unmarshal new_value: %w", err)
		}
	}

	return log, nil
}
