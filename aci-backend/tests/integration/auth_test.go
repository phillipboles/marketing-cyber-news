package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/phillipboles/aci-backend/internal/domain"
	"github.com/phillipboles/aci-backend/internal/pkg/crypto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// RegisterRequest represents the registration request payload
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// LoginRequest represents the login request payload
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents the refresh token request payload
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// AuthResponse represents the authentication response data
type AuthResponseData struct {
	User struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
		Role  string `json:"role"`
	} `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

// AuthResponse wraps the authentication response with standard envelope
type AuthResponse struct {
	Data    AuthResponseData `json:"data"`
	Message string           `json:"message,omitempty"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// TestRegister_HappyPath tests successful user registration
// Test ID: T035
func TestRegister_HappyPath(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: Valid registration data
	req := RegisterRequest{
		Email:    "newuser@example.com",
		Password: "SecurePass123!",
		Name:     "New User",
	}

	// When: POST /v1/auth/register
	resp := PostJSON(t, server.BaseURL+"/v1/auth/register", req)
	defer resp.Body.Close()

	// Debug: Log actual response
	bodyBytes, _ := io.ReadAll(resp.Body)
	t.Logf("Response status: %d, body: %s", resp.StatusCode, string(bodyBytes))

	// Then: Should return 201 Created with user and tokens
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "expected 201 Created")

	var authResp AuthResponse
	if err := json.Unmarshal(bodyBytes, &authResp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	// Verify user data
	assert.NotEmpty(t, authResp.Data.User.ID, "user ID should not be empty")
	assert.Equal(t, req.Email, authResp.Data.User.Email, "email should match")
	assert.Equal(t, req.Name, authResp.Data.User.Name, "name should match")
	assert.Equal(t, "user", authResp.Data.User.Role, "default role should be user")

	// Verify tokens
	assert.NotEmpty(t, authResp.Data.AccessToken, "access token should not be empty")
	assert.NotEmpty(t, authResp.Data.RefreshToken, "refresh token should not be empty")
	assert.NotEmpty(t, authResp.Data.ExpiresAt, "expires_at should not be empty")

	// Verify user was created in database
	ctx := context.Background()
	userID, err := uuid.Parse(authResp.Data.User.ID)
	require.NoError(t, err, "user ID should be valid UUID")

	var dbUser domain.User
	err = db.DB.Pool.QueryRow(ctx, "SELECT id, email, name, role FROM users WHERE id = $1", userID).
		Scan(&dbUser.ID, &dbUser.Email, &dbUser.Name, &dbUser.Role)
	require.NoError(t, err, "user should exist in database")

	assert.Equal(t, req.Email, dbUser.Email)
	assert.Equal(t, req.Name, dbUser.Name)
}

// TestRegister_DuplicateEmail tests registration with existing email
// Test ID: T036
func TestRegister_DuplicateEmail(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: An existing user
	existingEmail := "existing@example.com"
	ctx := context.Background()

	passwordHash, err := crypto.HashPassword("ExistingPass123!")
	require.NoError(t, err)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(), existingEmail, passwordHash, "Existing User", "user")
	require.NoError(t, err)

	// When: Attempting to register with same email
	req := RegisterRequest{
		Email:    existingEmail,
		Password: "NewPass123!",
		Name:     "New User",
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/register", req)
	defer resp.Body.Close()

	// Then: Should return 409 Conflict
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "expected 409 Conflict")

	var errResp ErrorResponse
	ReadJSONResponse(t, resp, &errResp)

	assert.Contains(t, errResp.Message, "email", "error should mention email")
	assert.Contains(t, errResp.Message, "already", "error should mention already exists")
}

// TestRegister_InvalidEmail tests registration with invalid email format
func TestRegister_InvalidEmail(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	testCases := []struct {
		name  string
		email string
	}{
		{"missing @", "invalidemail.com"},
		{"missing domain", "user@"},
		{"missing local part", "@domain.com"},
		{"empty", ""},
		{"spaces", "user @domain.com"},
		{"no TLD", "user@domain"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := RegisterRequest{
				Email:    tc.email,
				Password: "ValidPass123!",
				Name:     "Test User",
			}

			resp := PostJSON(t, server.BaseURL+"/v1/auth/register", req)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
				"invalid email should return 400")
		})
	}
}

// TestRegister_WeakPassword tests registration with weak passwords
func TestRegister_WeakPassword(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	testCases := []struct {
		name     string
		password string
	}{
		{"too short", "Pass1!"},
		{"empty", ""},
		{"no uppercase", "password123!"},
		{"no lowercase", "PASSWORD123!"},
		{"no number", "Password!"},
		{"no special char", "Password123"},
		{"only 7 chars", "Pass12!"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			req := RegisterRequest{
				Email:    fmt.Sprintf("user-%s@example.com", tc.name),
				Password: tc.password,
				Name:     "Test User",
			}

			resp := PostJSON(t, server.BaseURL+"/v1/auth/register", req)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusBadRequest, resp.StatusCode,
				"weak password should return 400")
		})
	}
}

// TestLogin_HappyPath tests successful login with correct credentials
// Test ID: T037
func TestLogin_HappyPath(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: A registered user
	email := "loginuser@example.com"
	password := "LoginPass123!"
	name := "Login User"

	ctx := context.Background()
	passwordHash, err := crypto.HashPassword(password)
	require.NoError(t, err)

	userID := uuid.New()
	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)",
		userID, email, passwordHash, name, "user")
	require.NoError(t, err)

	// When: POST /v1/auth/login with correct credentials
	req := LoginRequest{
		Email:    email,
		Password: password,
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/login", req)
	defer resp.Body.Close()

	// Then: Should return 200 OK with user and tokens
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 OK")

	var authResp AuthResponse
	ReadJSONResponse(t, resp, &authResp)

	// Verify user data
	assert.Equal(t, userID.String(), authResp.Data.User.ID)
	assert.Equal(t, email, authResp.Data.User.Email)
	assert.Equal(t, name, authResp.Data.User.Name)

	// Verify tokens
	assert.NotEmpty(t, authResp.Data.AccessToken)
	assert.NotEmpty(t, authResp.Data.RefreshToken)
	assert.NotEmpty(t, authResp.Data.ExpiresAt)

	// Verify refresh token was stored
	var tokenCount int
	err = db.DB.Pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM refresh_tokens WHERE user_id = $1 AND revoked_at IS NULL",
		userID).Scan(&tokenCount)
	require.NoError(t, err)
	assert.Equal(t, 1, tokenCount, "should have one active refresh token")
}

// TestLogin_WrongPassword tests login with incorrect password
// Test ID: T038
func TestLogin_WrongPassword(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: A registered user
	email := "wrongpass@example.com"
	correctPassword := "CorrectPass123!"

	ctx := context.Background()
	passwordHash, err := crypto.HashPassword(correctPassword)
	require.NoError(t, err)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(), email, passwordHash, "Test User", "user")
	require.NoError(t, err)

	// When: POST /v1/auth/login with wrong password
	req := LoginRequest{
		Email:    email,
		Password: "WrongPassword123!",
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/login", req)
	defer resp.Body.Close()

	// Then: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")

	var errResp ErrorResponse
	ReadJSONResponse(t, resp, &errResp)

	assert.Contains(t, errResp.Message, "credentials", "error should mention invalid credentials")
}

// TestLogin_NonexistentUser tests login with non-existent email
func TestLogin_NonexistentUser(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// When: POST /v1/auth/login with non-existent email
	req := LoginRequest{
		Email:    "nonexistent@example.com",
		Password: "SomePass123!",
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/login", req)
	defer resp.Body.Close()

	// Then: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")

	var errResp ErrorResponse
	ReadJSONResponse(t, resp, &errResp)

	assert.Contains(t, errResp.Message, "credentials", "error should mention invalid credentials")
}

// TestRefresh_HappyPath tests successful token refresh
// Test ID: T039
func TestRefresh_HappyPath(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: A logged-in user with valid refresh token
	ctx := context.Background()
	userID := uuid.New()
	email := "refresh@example.com"

	passwordHash, err := crypto.HashPassword("RefreshPass123!")
	require.NoError(t, err)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)",
		userID, email, passwordHash, "Refresh User", "user")
	require.NoError(t, err)

	// Generate refresh token
	refreshToken, err := crypto.GenerateToken()
	require.NoError(t, err)
	tokenHash := crypto.HashToken(refreshToken)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)",
		uuid.New(), userID, tokenHash, time.Now().Add(7*24*time.Hour))
	require.NoError(t, err)

	// When: POST /v1/auth/refresh with valid refresh token
	req := RefreshRequest{
		RefreshToken: refreshToken,
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/refresh", req)
	defer resp.Body.Close()

	// Then: Should return 200 OK with new tokens
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 OK")

	var authResp AuthResponse
	ReadJSONResponse(t, resp, &authResp)

	// Verify new tokens
	assert.NotEmpty(t, authResp.Data.AccessToken, "should have new access token")
	assert.NotEmpty(t, authResp.Data.RefreshToken, "should have new refresh token")
	assert.NotEqual(t, refreshToken, authResp.Data.RefreshToken, "refresh token should be rotated")

	// Verify old token was revoked
	var revokedAt *time.Time
	err = db.DB.Pool.QueryRow(ctx,
		"SELECT revoked_at FROM refresh_tokens WHERE token_hash = $1",
		tokenHash).Scan(&revokedAt)
	require.NoError(t, err)
	assert.NotNil(t, revokedAt, "old token should be revoked")
}

// TestRefresh_RevokedToken tests refresh with revoked token
// Test ID: T040
func TestRefresh_RevokedToken(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: A revoked refresh token
	ctx := context.Background()
	userID := uuid.New()

	passwordHash, err := crypto.HashPassword("RevokedPass123!")
	require.NoError(t, err)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)",
		userID, "revoked@example.com", passwordHash, "Revoked User", "user")
	require.NoError(t, err)

	// Create revoked refresh token
	refreshToken, err := crypto.GenerateToken()
	require.NoError(t, err)
	tokenHash := crypto.HashToken(refreshToken)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked_at) VALUES ($1, $2, $3, $4, $5)",
		uuid.New(), userID, tokenHash, time.Now().Add(7*24*time.Hour), time.Now())
	require.NoError(t, err)

	// When: POST /v1/auth/refresh with revoked token
	req := RefreshRequest{
		RefreshToken: refreshToken,
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/refresh", req)
	defer resp.Body.Close()

	// Then: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")

	var errResp ErrorResponse
	ReadJSONResponse(t, resp, &errResp)

	assert.Contains(t, errResp.Message, "token", "error should mention invalid token")
}

// TestRefresh_ExpiredToken tests refresh with expired token
func TestRefresh_ExpiredToken(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: An expired refresh token
	ctx := context.Background()
	userID := uuid.New()

	passwordHash, err := crypto.HashPassword("ExpiredPass123!")
	require.NoError(t, err)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)",
		userID, "expired@example.com", passwordHash, "Expired User", "user")
	require.NoError(t, err)

	// Create expired refresh token
	refreshToken, err := crypto.GenerateToken()
	require.NoError(t, err)
	tokenHash := crypto.HashToken(refreshToken)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)",
		uuid.New(), userID, tokenHash, time.Now().Add(-1*time.Hour))
	require.NoError(t, err)

	// When: POST /v1/auth/refresh with expired token
	req := RefreshRequest{
		RefreshToken: refreshToken,
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/refresh", req)
	defer resp.Body.Close()

	// Then: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")
}

// TestLogout_InvalidatesToken tests logout revokes refresh token
// Test ID: T041
func TestLogout_InvalidatesToken(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: A logged-in user with refresh token
	ctx := context.Background()
	userID := uuid.New()
	email := "logout@example.com"

	passwordHash, err := crypto.HashPassword("LogoutPass123!")
	require.NoError(t, err)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)",
		userID, email, passwordHash, "Logout User", "user")
	require.NoError(t, err)

	// Create refresh token
	refreshToken, err := crypto.GenerateToken()
	require.NoError(t, err)
	tokenHash := crypto.HashToken(refreshToken)

	tokenID := uuid.New()
	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at) VALUES ($1, $2, $3, $4)",
		tokenID, userID, tokenHash, time.Now().Add(7*24*time.Hour))
	require.NoError(t, err)

	// Generate access token for authorization
	tokens, err := server.JWTService.GenerateTokenPair(userID, email, "user")
	require.NoError(t, err)

	// When: POST /v1/auth/logout
	resp := PostJSON(t, server.BaseURL+"/v1/auth/logout", nil)
	resp.Request.Header.Set("Authorization", "Bearer "+tokens.AccessToken)
	defer resp.Body.Close()

	// Then: Should return 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 OK")

	// Verify token was revoked
	var revokedAt *time.Time
	err = db.DB.Pool.QueryRow(ctx,
		"SELECT revoked_at FROM refresh_tokens WHERE id = $1",
		tokenID).Scan(&revokedAt)
	require.NoError(t, err)
	assert.NotNil(t, revokedAt, "refresh token should be revoked")

	// Subsequent refresh should fail
	req := RefreshRequest{
		RefreshToken: refreshToken,
	}

	resp2 := PostJSON(t, server.BaseURL+"/v1/auth/refresh", req)
	defer resp2.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp2.StatusCode,
		"refresh with revoked token should fail")
}

// TestProtectedEndpoint_WithValidToken tests accessing protected resource with valid token
func TestProtectedEndpoint_WithValidToken(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: A valid access token
	ctx := context.Background()
	userID := uuid.New()
	email := "protected@example.com"

	passwordHash, err := crypto.HashPassword("ProtectedPass123!")
	require.NoError(t, err)

	_, err = db.DB.Pool.Exec(ctx,
		"INSERT INTO users (id, email, password_hash, name, role) VALUES ($1, $2, $3, $4, $5)",
		userID, email, passwordHash, "Protected User", "user")
	require.NoError(t, err)

	tokens, err := server.JWTService.GenerateTokenPair(userID, email, "user")
	require.NoError(t, err)

	// When: GET /v1/users/me with valid token
	resp := GetJSON(t, server.BaseURL+"/v1/users/me", tokens.AccessToken)
	defer resp.Body.Close()

	// Then: Should return 200 OK
	assert.Equal(t, http.StatusOK, resp.StatusCode, "expected 200 OK")
}

// TestProtectedEndpoint_WithExpiredToken tests accessing protected resource with expired token
func TestProtectedEndpoint_WithExpiredToken(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: An expired access token (simulated with invalid token)
	expiredToken := "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE1MTYyMzkwMjJ9.invalid"

	// When: GET /v1/users/me with expired token
	resp := GetJSON(t, server.BaseURL+"/v1/users/me", expiredToken)
	defer resp.Body.Close()

	// Then: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")
}

// TestProtectedEndpoint_WithoutToken tests accessing protected resource without token
func TestProtectedEndpoint_WithoutToken(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// When: GET /v1/users/me without token
	resp := GetJSON(t, server.BaseURL+"/v1/users/me", "")
	defer resp.Body.Close()

	// Then: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")
}

// TestRegister_EmptyName tests registration with empty name
func TestRegister_EmptyName(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// When: POST /v1/auth/register with empty name
	req := RegisterRequest{
		Email:    "emptyname@example.com",
		Password: "ValidPass123!",
		Name:     "",
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/register", req)
	defer resp.Body.Close()

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 Bad Request")
}

// TestRegister_NameTooShort tests registration with name less than 2 characters
func TestRegister_NameTooShort(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// When: POST /v1/auth/register with single character name
	req := RegisterRequest{
		Email:    "shortname@example.com",
		Password: "ValidPass123!",
		Name:     "A",
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/register", req)
	defer resp.Body.Close()

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 Bad Request")
}

// TestLogin_EmptyEmail tests login with empty email
func TestLogin_EmptyEmail(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// When: POST /v1/auth/login with empty email
	req := LoginRequest{
		Email:    "",
		Password: "SomePass123!",
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/login", req)
	defer resp.Body.Close()

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 Bad Request")
}

// TestLogin_EmptyPassword tests login with empty password
func TestLogin_EmptyPassword(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// When: POST /v1/auth/login with empty password
	req := LoginRequest{
		Email:    "test@example.com",
		Password: "",
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/login", req)
	defer resp.Body.Close()

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 Bad Request")
}

// TestRefresh_InvalidToken tests refresh with malformed token
func TestRefresh_InvalidToken(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// When: POST /v1/auth/refresh with invalid token
	req := RefreshRequest{
		RefreshToken: "invalid-token-format",
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/refresh", req)
	defer resp.Body.Close()

	// Then: Should return 401 Unauthorized
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expected 401 Unauthorized")
}

// TestRefresh_EmptyToken tests refresh with empty token
func TestRefresh_EmptyToken(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// When: POST /v1/auth/refresh with empty token
	req := RefreshRequest{
		RefreshToken: "",
	}

	resp := PostJSON(t, server.BaseURL+"/v1/auth/refresh", req)
	defer resp.Body.Close()

	// Then: Should return 400 Bad Request
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode, "expected 400 Bad Request")
}

// TestConcurrentRegistrations tests multiple concurrent registrations
func TestConcurrentRegistrations(t *testing.T) {
	// Setup
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	server := SetupTestServer(t, db)
	defer TeardownTestServer(t, server)

	// Given: Multiple concurrent registration requests
	concurrency := 10
	done := make(chan bool, concurrency)
	successCount := 0

	for i := 0; i < concurrency; i++ {
		go func(index int) {
			req := RegisterRequest{
				Email:    fmt.Sprintf("concurrent%d@example.com", index),
				Password: "ConcurrentPass123!",
				Name:     fmt.Sprintf("Concurrent User %d", index),
			}

			resp := PostJSON(t, server.BaseURL+"/v1/auth/register", req)
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusCreated {
				successCount++
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		<-done
	}

	// Then: All registrations should succeed
	assert.Equal(t, concurrency, successCount, "all concurrent registrations should succeed")
}
