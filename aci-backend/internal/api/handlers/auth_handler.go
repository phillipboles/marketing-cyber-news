package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/rs/zerolog/log"

	"github.com/phillipboles/aci-backend/internal/api/middleware"
	"github.com/phillipboles/aci-backend/internal/api/response"
	"github.com/phillipboles/aci-backend/internal/domain/entities"
	domainerrors "github.com/phillipboles/aci-backend/internal/domain/errors"
	"github.com/phillipboles/aci-backend/internal/service"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new authentication handler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	if authService == nil {
		panic("authService cannot be nil")
	}
	return &AuthHandler{
		authService: authService,
	}
}

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

// LogoutRequest represents the logout request payload
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token,omitempty"`
	AllDevices   bool   `json:"all_devices"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User         UserDTO  `json:"user"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token"`
	ExpiresAt    string   `json:"expires_at"`
}

// UserDTO represents user data transfer object
type UserDTO struct {
	ID            string  `json:"id"`
	Email         string  `json:"email"`
	Name          string  `json:"name"`
	Role          string  `json:"role"`
	EmailVerified bool    `json:"email_verified"`
	LastLoginAt   *string `json:"last_login_at,omitempty"`
}

// TokenResponse represents token refresh response
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

// Register handles user registration
// POST /v1/auth/register
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		requestID := middleware.GetRequestID(r.Context())
		response.BadRequestWithDetails(w, "Invalid request body", nil, requestID)
		return
	}

	user, tokens, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		h.handleAuthError(w, r, err)
		return
	}

	authResp := AuthResponse{
		User:         h.userToDTO(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Created(w, authResp)
}

// Login handles user authentication
// POST /v1/auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		requestID := middleware.GetRequestID(r.Context())
		response.BadRequestWithDetails(w, "Invalid request body", nil, requestID)
		return
	}

	user, tokens, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.handleAuthError(w, r, err)
		return
	}

	authResp := AuthResponse{
		User:         h.userToDTO(user),
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, authResp)
}

// Refresh handles token refresh
// POST /v1/auth/refresh
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		requestID := middleware.GetRequestID(r.Context())
		response.BadRequestWithDetails(w, "Invalid request body", nil, requestID)
		return
	}

	if req.RefreshToken == "" {
		response.BadRequest(w, "refresh_token is required")
		return
	}

	tokens, err := h.authService.Refresh(r.Context(), req.RefreshToken)
	if err != nil {
		h.handleAuthError(w, r, err)
		return
	}

	tokenResp := TokenResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
		ExpiresAt:    tokens.ExpiresAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, tokenResp)
}

// Logout handles user logout
// POST /v1/auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		requestID := middleware.GetRequestID(r.Context())
		response.BadRequestWithDetails(w, "Invalid request body", nil, requestID)
		return
	}

	// If logging out all devices, get user ID from JWT context
	if req.AllDevices {
		claims, ok := middleware.GetUserFromContext(r.Context())
		if !ok {
			response.Unauthorized(w, "Authentication required")
			return
		}

		if err := h.authService.LogoutAll(r.Context(), claims.UserID); err != nil {
			h.handleAuthError(w, r, err)
			return
		}
	} else {
		// Logout single device with refresh token
		if req.RefreshToken == "" {
			response.BadRequest(w, "refresh_token is required when all_devices is false")
			return
		}

		if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
			h.handleAuthError(w, r, err)
			return
		}
	}

	response.SuccessWithMessage(w, nil, "Logged out successfully")
}


// handleAuthError handles authentication-specific errors
func (h *AuthHandler) handleAuthError(w http.ResponseWriter, r *http.Request, err error) {
	requestID := middleware.GetRequestID(r.Context())

	// Handle validation errors
	var validationErr *domainerrors.ValidationError
	if errors.As(err, &validationErr) {
		response.BadRequestWithDetails(w, validationErr.Error(), nil, requestID)
		return
	}

	// Handle conflict errors (email already exists)
	var conflictErr *domainerrors.ConflictError
	if errors.As(err, &conflictErr) {
		response.Conflict(w, conflictErr.Error())
		return
	}

	// Handle unauthorized errors
	if errors.Is(err, domainerrors.ErrUnauthorized) {
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	// Handle not found errors
	var notFoundErr *domainerrors.NotFoundError
	if errors.As(err, &notFoundErr) {
		response.NotFound(w, notFoundErr.Error())
		return
	}

	// Generic internal error - log for debugging
	log.Error().
		Err(err).
		Str("request_id", requestID).
		Msg("Unhandled error in auth handler")
	response.InternalError(w, "An unexpected error occurred", requestID)
}
// userToDTO converts entities.User to DTO
func (h *AuthHandler) userToDTO(u *entities.User) UserDTO {
	dto := UserDTO{
		ID:            u.ID.String(),
		Email:         u.Email,
		Name:          u.Name,
		Role:          string(u.Role),
		EmailVerified: u.EmailVerified,
	}

	if u.LastLoginAt != nil {
		lastLogin := u.LastLoginAt.Format("2006-01-02T15:04:05Z07:00")
		dto.LastLoginAt = &lastLogin
	}

	return dto
}
