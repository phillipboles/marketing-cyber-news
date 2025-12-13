package service

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/domain/entities"
	domainerrors "github.com/phillipboles/aci-backend/internal/domain/errors"
	"github.com/phillipboles/aci-backend/internal/pkg/crypto"
	"github.com/phillipboles/aci-backend/internal/pkg/jwt"
	"github.com/phillipboles/aci-backend/internal/repository"
)

const (
	minPasswordLength = 8
	minNameLength     = 2
)

var (
	// emailRegex is a basic email validation pattern
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
)

// AuthService handles authentication business logic
type AuthService struct {
	userRepo  UserRepoInterface
	tokenRepo repository.RefreshTokenRepository
	jwtSvc    jwt.Service
}

// NewAuthService creates a new authentication service
func NewAuthService(
	userRepo UserRepoInterface,
	tokenRepo repository.RefreshTokenRepository,
	jwtSvc jwt.Service,
) *AuthService {
	if userRepo == nil {
		panic("userRepo cannot be nil")
	}
	if tokenRepo == nil {
		panic("tokenRepo cannot be nil")
	}
	if jwtSvc == nil {
		panic("jwtSvc cannot be nil")
	}

	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtSvc:    jwtSvc,
	}
}

// Register creates a new user account with validation and password hashing
func (s *AuthService) Register(ctx context.Context, email, password, name string) (*entities.User, *jwt.TokenPair, error) {
	// Validate email
	if err := s.validateEmail(email); err != nil {
		return nil, nil, err
	}

	// Validate password strength
	if err := s.validatePassword(password); err != nil {
		return nil, nil, err
	}

	// Validate name
	if err := s.validateName(name); err != nil {
		return nil, nil, err
	}

	// Check if email already exists
	_, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil {
		// User found - email conflict
		return nil, nil, &domainerrors.ConflictError{
			Resource: "user",
			Field:    "email",
			Value:    email,
		}
	}

	// If error is not NotFound, it's an actual error
	var notFoundErr *domainerrors.NotFoundError
	if err != nil && !errors.As(err, &notFoundErr) {
		return nil, nil, fmt.Errorf("failed to check existing user: %w", err)
	}

	// Hash password
	passwordHash, err := crypto.HashPassword(password)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user entity
	user := entities.NewUser(email, passwordHash, name)

	// Persist user
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Generate token pair
	tokenPair, err := s.generateAndStoreTokens(ctx, user, "", "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user, tokenPair, nil
}

// Login authenticates user credentials and returns tokens
func (s *AuthService) Login(ctx context.Context, email, password string) (*entities.User, *jwt.TokenPair, error) {
	if email == "" {
		return nil, nil, &domainerrors.ValidationError{
			Field:   "email",
			Message: "email is required",
		}
	}

	if password == "" {
		return nil, nil, &domainerrors.ValidationError{
			Field:   "password",
			Message: "password is required",
		}
	}

	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		// Return generic unauthorized error to prevent email enumeration
		return nil, nil, fmt.Errorf("invalid credentials: %w", domainerrors.ErrUnauthorized)
	}

	// Verify password
	if !crypto.CheckPassword(password, user.PasswordHash) {
		return nil, nil, fmt.Errorf("invalid credentials: %w", domainerrors.ErrUnauthorized)
	}

	// Update last login timestamp
	if err := s.userRepo.UpdateLastLogin(ctx, user.ID); err != nil {
		// Log error but don't fail login
		// In production, use proper logger
		_ = err
	}

	// Generate token pair
	tokenPair, err := s.generateAndStoreTokens(ctx, user, "", "")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate tokens: %w", err)
	}

	return user, tokenPair, nil
}

// Refresh generates new token pair from valid refresh token (token rotation)
func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (*jwt.TokenPair, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required: %w", domainerrors.ErrUnauthorized)
	}

	// Validate refresh token format with JWT service
	userID, err := s.jwtSvc.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", domainerrors.ErrUnauthorized)
	}

	// Hash the token to look it up in database
	tokenHash := crypto.HashToken(refreshToken)

	// Get refresh token record from database
	storedToken, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, fmt.Errorf("refresh token not found or expired: %w", domainerrors.ErrUnauthorized)
	}

	// Verify token belongs to the user from JWT claims
	if storedToken.UserID != userID {
		return nil, fmt.Errorf("token user mismatch: %w", domainerrors.ErrUnauthorized)
	}

	// Get user details for new token
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	// Revoke old refresh token (token rotation security)
	if err := s.tokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		// Log error but continue - we'll still issue new token
		_ = err
	}

	// Generate new token pair
	tokenPair, err := s.generateAndStoreTokens(ctx, user, "", "")
	if err != nil {
		return nil, fmt.Errorf("failed to generate new tokens: %w", err)
	}

	return tokenPair, nil
}

// Logout invalidates a specific refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return fmt.Errorf("refresh token is required")
	}

	// Validate token format
	_, err := s.jwtSvc.ValidateRefreshToken(refreshToken)
	if err != nil {
		// Token invalid or expired - already unusable, no need to error
		return nil
	}

	// Hash token to look it up
	tokenHash := crypto.HashToken(refreshToken)

	// Get token from database
	storedToken, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		// Token not found or already revoked - consider success
		return nil
	}

	// Revoke the token
	if err := s.tokenRepo.Revoke(ctx, storedToken.ID); err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}

	return nil
}

// LogoutAll invalidates all refresh tokens for a user
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if err := s.tokenRepo.RevokeAllForUser(ctx, userID); err != nil {
		return fmt.Errorf("failed to revoke all tokens: %w", err)
	}

	return nil
}

// generateAndStoreTokens creates JWT pair and stores refresh token in database
func (s *AuthService) generateAndStoreTokens(
	ctx context.Context,
	user *entities.User,
	ipAddress string,
	userAgent string,
) (*jwt.TokenPair, error) {
	// Generate JWT token pair
	tokenPair, err := s.jwtSvc.GenerateTokenPair(user.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, fmt.Errorf("failed to generate token pair: %w", err)
	}

	// Hash refresh token before storing
	tokenHash := crypto.HashToken(tokenPair.RefreshToken)

	// Create refresh token record
	refreshToken := &domain.RefreshToken{
		ID:         uuid.New(),
		UserID:     user.ID,
		Token:      tokenHash, // Store hash, not plain token
		ExpiresAt:  time.Now().Add(jwt.RefreshTokenExpiry),
		CreatedAt:  time.Now(),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}

	// Store refresh token in database
	if err := s.tokenRepo.Create(ctx, refreshToken); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	return tokenPair, nil
}

// validateEmail checks email format and requirements
func (s *AuthService) validateEmail(email string) error {
	if email == "" {
		return &domainerrors.ValidationError{
			Field:   "email",
			Message: "email is required",
		}
	}

	email = strings.TrimSpace(strings.ToLower(email))

	if !emailRegex.MatchString(email) {
		return &domainerrors.ValidationError{
			Field:   "email",
			Message: "invalid email format",
		}
	}

	return nil
}

// validatePassword checks password strength requirements
func (s *AuthService) validatePassword(password string) error {
	if password == "" {
		return &domainerrors.ValidationError{
			Field:   "password",
			Message: "password is required",
		}
	}

	if len(password) < minPasswordLength {
		return &domainerrors.ValidationError{
			Field:   "password",
			Message: fmt.Sprintf("password must be at least %d characters", minPasswordLength),
		}
	}

	// Check for at least one uppercase letter
	hasUpper := false
	for _, char := range password {
		if char >= 'A' && char <= 'Z' {
			hasUpper = true
			break
		}
	}

	if !hasUpper {
		return &domainerrors.ValidationError{
			Field:   "password",
			Message: "password must contain at least one uppercase letter",
		}
	}

	// Check for at least one lowercase letter
	hasLower := false
	for _, char := range password {
		if char >= 'a' && char <= 'z' {
			hasLower = true
			break
		}
	}

	if !hasLower {
		return &domainerrors.ValidationError{
			Field:   "password",
			Message: "password must contain at least one lowercase letter",
		}
	}

	// Check for at least one digit
	hasDigit := false
	for _, char := range password {
		if char >= '0' && char <= '9' {
			hasDigit = true
			break
		}
	}

	if !hasDigit {
		return &domainerrors.ValidationError{
			Field:   "password",
			Message: "password must contain at least one digit",
		}
	}

	return nil
}

// validateName checks name requirements
func (s *AuthService) validateName(name string) error {
	if name == "" {
		return &domainerrors.ValidationError{
			Field:   "name",
			Message: "name is required",
		}
	}

	name = strings.TrimSpace(name)

	if len(name) < minNameLength {
		return &domainerrors.ValidationError{
			Field:   "name",
			Message: fmt.Sprintf("name must be at least %d characters", minNameLength),
		}
	}

	return nil
}
