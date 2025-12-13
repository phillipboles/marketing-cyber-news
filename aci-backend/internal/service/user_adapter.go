package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/domain/entities"
)

// UserRepoInterface defines what we need from user repository
// This works around the mismatch between repository.UserRepository (uses domain.User)
// and the actual postgres implementation (uses entities.User)
type UserRepoInterface interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	UpdateLastLogin(ctx context.Context, id uuid.UUID) error
}
