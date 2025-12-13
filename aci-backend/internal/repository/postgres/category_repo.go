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

type categoryRepository struct {
	db *DB
}

// NewCategoryRepository creates a new PostgreSQL category repository
func NewCategoryRepository(db *DB) repository.CategoryRepository {
	if db == nil {
		panic("database cannot be nil")
	}
	return &categoryRepository{db: db}
}

// Create creates a new category
func (r *categoryRepository) Create(ctx context.Context, category *domain.Category) error {
	if category == nil {
		return fmt.Errorf("category cannot be nil")
	}

	if err := category.Validate(); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}

	query := `
		INSERT INTO categories (id, name, slug, description, color, icon, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := r.db.Pool.Exec(ctx, query,
		category.ID,
		category.Name,
		category.Slug,
		category.Description,
		category.Color,
		category.Icon,
		category.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create category: %w", err)
	}

	return nil
}

// GetByID retrieves a category by ID
func (r *categoryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("category ID cannot be nil")
	}

	query := `
		SELECT id, name, slug, description, color, icon, created_at
		FROM categories
		WHERE id = $1
	`

	category := &domain.Category{}
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.Color,
		&category.Icon,
		&category.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}

	return category, nil
}

// GetBySlug retrieves a category by slug
func (r *categoryRepository) GetBySlug(ctx context.Context, slug string) (*domain.Category, error) {
	if slug == "" {
		return nil, fmt.Errorf("slug cannot be empty")
	}

	query := `
		SELECT id, name, slug, description, color, icon, created_at
		FROM categories
		WHERE slug = $1
	`

	category := &domain.Category{}
	err := r.db.Pool.QueryRow(ctx, query, slug).Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.Description,
		&category.Color,
		&category.Icon,
		&category.CreatedAt,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("category not found with slug %s: %w", slug, err)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get category by slug: %w", err)
	}

	return category, nil
}

// List retrieves all categories
func (r *categoryRepository) List(ctx context.Context) ([]*domain.Category, error) {
	query := `
		SELECT id, name, slug, description, color, icon, created_at
		FROM categories
		ORDER BY name ASC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	categories := make([]*domain.Category, 0)
	for rows.Next() {
		category := &domain.Category{}
		err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.Color,
			&category.Icon,
			&category.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating categories: %w", err)
	}

	return categories, nil
}

// Update updates an existing category
func (r *categoryRepository) Update(ctx context.Context, category *domain.Category) error {
	if category == nil {
		return fmt.Errorf("category cannot be nil")
	}

	if err := category.Validate(); err != nil {
		return fmt.Errorf("invalid category: %w", err)
	}

	query := `
		UPDATE categories
		SET name = $2, slug = $3, description = $4, color = $5, icon = $6
		WHERE id = $1
	`

	cmdTag, err := r.db.Pool.Exec(ctx, query,
		category.ID,
		category.Name,
		category.Slug,
		category.Description,
		category.Color,
		category.Icon,
	)

	if err != nil {
		return fmt.Errorf("failed to update category: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}

// Delete deletes a category by ID
func (r *categoryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("category ID cannot be nil")
	}

	query := `DELETE FROM categories WHERE id = $1`

	cmdTag, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}

	return nil
}
