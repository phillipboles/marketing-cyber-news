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

type sourceRepository struct {
	db *DB
}

// NewSourceRepository creates a new PostgreSQL source repository
func NewSourceRepository(db *DB) repository.SourceRepository {
	if db == nil {
		panic("database cannot be nil")
	}
	return &sourceRepository{db: db}
}

// Create creates a new source
func (r *sourceRepository) Create(ctx context.Context, source *domain.Source) error {
	if source == nil {
		return fmt.Errorf("source cannot be nil")
	}

	if err := source.Validate(); err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	query := `
		INSERT INTO sources (id, name, url, description, is_active, trust_score, last_scraped_at, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		source.ID,
		source.Name,
		source.URL,
		source.Description,
		source.IsActive,
		source.TrustScore,
		source.LastScrapedAt,
		source.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	return nil
}

// GetByID retrieves a source by ID
func (r *sourceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Source, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("source ID cannot be nil")
	}

	query := `
		SELECT id, name, url, description, is_active, trust_score, last_scraped_at, created_at
		FROM sources
		WHERE id = $1
	`

	source := &domain.Source{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Description,
		&source.IsActive,
		&source.TrustScore,
		&source.LastScrapedAt,
		&source.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("source not found: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	return source, nil
}

// GetByURL retrieves a source by URL
func (r *sourceRepository) GetByURL(ctx context.Context, url string) (*domain.Source, error) {
	if url == "" {
		return nil, fmt.Errorf("URL cannot be empty")
	}

	query := `
		SELECT id, name, url, description, is_active, trust_score, last_scraped_at, created_at
		FROM sources
		WHERE url = $1
	`

	source := &domain.Source{}
	err := r.db.Pool.QueryRow(ctx, query, url).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Description,
		&source.IsActive,
		&source.TrustScore,
		&source.LastScrapedAt,
		&source.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("source not found with URL %s: %w", url, err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get source by URL: %w", err)
	}

	return source, nil
}

// GetByName retrieves a source by name
func (r *sourceRepository) GetByName(ctx context.Context, name string) (*domain.Source, error) {
	if name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	query := `
		SELECT id, name, url, description, is_active, trust_score, last_scraped_at, created_at
		FROM sources
		WHERE name = $1
	`

	source := &domain.Source{}
	err := r.db.Pool.QueryRow(ctx, query, name).Scan(
		&source.ID,
		&source.Name,
		&source.URL,
		&source.Description,
		&source.IsActive,
		&source.TrustScore,
		&source.LastScrapedAt,
		&source.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("source not found with name %s: %w", name, err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get source by name: %w", err)
	}

	return source, nil
}

// List retrieves all sources, optionally filtering by active status
func (r *sourceRepository) List(ctx context.Context, activeOnly bool) ([]*domain.Source, error) {
	query := `
		SELECT id, name, url, description, is_active, trust_score, last_scraped_at, created_at
		FROM sources
	`

	if activeOnly {
		query += ` WHERE is_active = true`
	}

	query += ` ORDER BY name ASC`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list sources: %w", err)
	}
	defer rows.Close()

	sources := make([]*domain.Source, 0)
	for rows.Next() {
		source := &domain.Source{}
		err := rows.Scan(
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
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		sources = append(sources, source)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating sources: %w", err)
	}

	return sources, nil
}

// Update updates an existing source
func (r *sourceRepository) Update(ctx context.Context, source *domain.Source) error {
	if source == nil {
		return fmt.Errorf("source cannot be nil")
	}

	if err := source.Validate(); err != nil {
		return fmt.Errorf("invalid source: %w", err)
	}

	query := `
		UPDATE sources
		SET name = $2, url = $3, description = $4, is_active = $5, trust_score = $6, last_scraped_at = $7
		WHERE id = $1
	`

	cmdTag, err := r.db.Pool.Exec(ctx, query,
		source.ID,
		source.Name,
		source.URL,
		source.Description,
		source.IsActive,
		source.TrustScore,
		source.LastScrapedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to update source: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("source not found")
	}

	return nil
}

// Delete deletes a source by ID
func (r *sourceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("source ID cannot be nil")
	}

	query := `DELETE FROM sources WHERE id = $1`

	cmdTag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("source not found")
	}

	return nil
}
