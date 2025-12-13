package jwt

import (
	"crypto/rsa"
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

const (
	// AccessTokenExpiry is the duration for access token validity
	AccessTokenExpiry = 15 * time.Minute

	// RefreshTokenExpiry is the duration for refresh token validity
	RefreshTokenExpiry = 7 * 24 * time.Hour
)

// TokenPair holds access and refresh tokens
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// Claims represents JWT claims structure
type Claims struct {
	jwt.RegisteredClaims
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
	Role   string    `json:"role"`
}

// Service defines the interface for JWT operations
type Service interface {
	GenerateTokenPair(userID uuid.UUID, email, role string) (*TokenPair, error)
	ValidateAccessToken(tokenString string) (*Claims, error)
	ValidateRefreshToken(tokenString string) (uuid.UUID, error)
}

// service implements the Service interface using RS256 signing
type service struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
	issuer     string
}

// Config holds configuration for JWT service
type Config struct {
	PrivateKeyPath string
	PublicKeyPath  string
	Issuer         string
}

// NewService creates a new JWT service
func NewService(cfg *Config) (Service, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}

	if cfg.PrivateKeyPath == "" {
		return nil, fmt.Errorf("private key path is required")
	}

	if cfg.PublicKeyPath == "" {
		return nil, fmt.Errorf("public key path is required")
	}

	if cfg.Issuer == "" {
		return nil, fmt.Errorf("issuer is required")
	}

	s := &service{
		issuer: cfg.Issuer,
	}

	if err := s.LoadPrivateKey(cfg.PrivateKeyPath); err != nil {
		return nil, fmt.Errorf("failed to load private key: %w", err)
	}

	if err := s.LoadPublicKey(cfg.PublicKeyPath); err != nil {
		return nil, fmt.Errorf("failed to load public key: %w", err)
	}

	return s, nil
}

// LoadPrivateKey loads the RSA private key from file
func (s *service) LoadPrivateKey(path string) error {
	if path == "" {
		return fmt.Errorf("private key path is required")
	}

	keyData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read private key file: %w", err)
	}

	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(keyData)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	s.privateKey = privateKey
	return nil
}

// LoadPublicKey loads the RSA public key from file
func (s *service) LoadPublicKey(path string) error {
	if path == "" {
		return fmt.Errorf("public key path is required")
	}

	keyData, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read public key file: %w", err)
	}

	publicKey, err := jwt.ParseRSAPublicKeyFromPEM(keyData)
	if err != nil {
		return fmt.Errorf("failed to parse public key: %w", err)
	}

	s.publicKey = publicKey
	return nil
}

// GenerateTokenPair generates both access and refresh tokens
func (s *service) GenerateTokenPair(userID uuid.UUID, email, role string) (*TokenPair, error) {
	if userID == uuid.Nil {
		return nil, fmt.Errorf("user ID is required")
	}

	if email == "" {
		return nil, fmt.Errorf("email is required")
	}

	if role == "" {
		return nil, fmt.Errorf("role is required")
	}

	if s.privateKey == nil {
		return nil, fmt.Errorf("private key not loaded")
	}

	now := time.Now()
	accessExpiry := now.Add(AccessTokenExpiry)
	refreshExpiry := now.Add(RefreshTokenExpiry)

	// Generate access token
	accessClaims := &Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accessExpiry),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    s.issuer,
			Subject:   userID.String(),
		},
		UserID: userID,
		Email:  email,
		Role:   role,
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	accessTokenString, err := accessToken.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// Generate refresh token
	refreshClaims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(refreshExpiry),
		IssuedAt:  jwt.NewNumericDate(now),
		NotBefore: jwt.NewNumericDate(now),
		Issuer:    s.issuer,
		Subject:   userID.String(),
		ID:        uuid.New().String(), // Unique ID for refresh token
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(s.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    accessExpiry,
	}, nil
}

// ValidateAccessToken validates and parses an access token
func (s *service) ValidateAccessToken(tokenString string) (*Claims, error) {
	if tokenString == "" {
		return nil, fmt.Errorf("token is required")
	}

	if s.publicKey == nil {
		return nil, fmt.Errorf("public key not loaded")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != s.issuer {
		return nil, fmt.Errorf("invalid issuer: expected %s, got %s", s.issuer, claims.Issuer)
	}

	// Validate expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("token has expired")
	}

	// Validate user ID
	if claims.UserID == uuid.Nil {
		return nil, fmt.Errorf("invalid user ID in token")
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token and returns the user ID
func (s *service) ValidateRefreshToken(tokenString string) (uuid.UUID, error) {
	if tokenString == "" {
		return uuid.Nil, fmt.Errorf("token is required")
	}

	if s.publicKey == nil {
		return uuid.Nil, fmt.Errorf("public key not loaded")
	}

	token, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (interface{}, error) {
		// Verify signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.publicKey, nil
	})

	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return uuid.Nil, fmt.Errorf("token is invalid")
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		return uuid.Nil, fmt.Errorf("invalid token claims")
	}

	// Validate issuer
	if claims.Issuer != s.issuer {
		return uuid.Nil, fmt.Errorf("invalid issuer: expected %s, got %s", s.issuer, claims.Issuer)
	}

	// Validate expiration
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, fmt.Errorf("token has expired")
	}

	// Parse user ID from subject
	if claims.Subject == "" {
		return uuid.Nil, fmt.Errorf("missing subject in token")
	}

	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	return userID, nil
}
