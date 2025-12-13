package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/phillipboles/aci-backend/internal/domain/entities"
	domainerrors "github.com/phillipboles/aci-backend/internal/domain/errors"
)

// UserRepository implements repository.UserRepository for PostgreSQL
type UserRepository struct {
	db *DB
}

// NewUserRepository creates a new PostgreSQL user repository
func NewUserRepository(db *DB) *UserRepository {
	if db == nil {
		panic("database cannot be nil")
	}
	return &UserRepository{db: db}
}

// Create inserts a new user into the database
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if user.ID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	if user.Email == "" {
		return fmt.Errorf("user email is required")
	}

	query := `
		INSERT INTO users (id, email, password_hash, name, role, email_verified, created_at, updated_at, last_login_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := r.db.Pool.Exec(
		ctx,
		query,
		user.ID,
		user.Email,
		user.PasswordHash,
		user.Name,
		user.Role,
		user.EmailVerified,
		user.CreatedAt,
		user.UpdatedAt,
		user.LastLoginAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			// Unique constraint violation (23505)
			if pgErr.Code == "23505" {
				if pgErr.ConstraintName == "users_email_key" {
					return &domainerrors.ConflictError{
						Resource: "user",
						Field:    "email",
						Value:    user.Email,
					}
				}
				return fmt.Errorf("user already exists: %w", domainerrors.ErrConflict)
			}
		}
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID retrieves a user by their ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	if id == uuid.Nil {
		return nil, fmt.Errorf("user ID cannot be nil")
	}

	query := `
		SELECT id, email, password_hash, name, role, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE id = $1
	`

	var user entities.User
	err := r.db.Pool.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Role,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domainerrors.NotFoundError{
				Resource: "user",
				ID:       id.String(),
			}
		}
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by their email address
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if email == "" {
		return nil, fmt.Errorf("email cannot be empty")
	}

	query := `
		SELECT id, email, password_hash, name, role, email_verified, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`

	var user entities.User
	err := r.db.Pool.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.PasswordHash,
		&user.Name,
		&user.Role,
		&user.EmailVerified,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.LastLoginAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &domainerrors.NotFoundError{
				Resource: "user",
				ID:       email,
			}
		}
		return nil, fmt.Errorf("failed to get user by email: %w", err)
	}

	return &user, nil
}

// Update updates an existing user's information
func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	if user == nil {
		return fmt.Errorf("user cannot be nil")
	}

	if user.ID == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	user.UpdatedAt = time.Now()

	query := `
		UPDATE users
		SET name = $2, email_verified = $3, updated_at = $4, role = $5
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(
		ctx,
		query,
		user.ID,
		user.Name,
		user.EmailVerified,
		user.UpdatedAt,
		user.Role,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &domainerrors.NotFoundError{
			Resource: "user",
			ID:       user.ID.String(),
		}
	}

	return nil
}

// UpdateLastLogin updates the user's last login timestamp
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	now := time.Now()
	query := `
		UPDATE users
		SET last_login_at = $2
		WHERE id = $1
	`

	result, err := r.db.Pool.Exec(ctx, query, id, now)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &domainerrors.NotFoundError{
			Resource: "user",
			ID:       id.String(),
		}
	}

	return nil
}

// Delete removes a user from the database
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return fmt.Errorf("user ID cannot be nil")
	}

	query := `DELETE FROM users WHERE id = $1`

	result, err := r.db.Pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	if result.RowsAffected() == 0 {
		return &domainerrors.NotFoundError{
			Resource: "user",
			ID:       id.String(),
		}
	}

	return nil
}
