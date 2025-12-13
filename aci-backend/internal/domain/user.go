package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// UserRole represents the role of a user in the system
type UserRole string

const (
	RoleUser  UserRole = "user"
	RoleAdmin UserRole = "admin"
)

// IsValid checks if the user role is valid
func (r UserRole) IsValid() error {
	if r != RoleUser && r != RoleAdmin {
		return fmt.Errorf("invalid user role: %s", r)
	}
	return nil
}

// String returns the string representation of the role
func (r UserRole) String() string {
	return string(r)
}

// User represents a user in the system
type User struct {
	ID            uuid.UUID  `json:"id"`
	Email         string     `json:"email"`
	PasswordHash  string     `json:"-"`
	Name          string     `json:"name"`
	Role          UserRole   `json:"role"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	EmailVerified bool       `json:"email_verified"`
	LastLoginAt   *time.Time `json:"last_login_at,omitempty"`
}

// Validate validates the user entity
func (u *User) Validate() error {
	if u.ID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if u.Email == "" {
		return fmt.Errorf("email is required")
	}

	if u.Name == "" {
		return fmt.Errorf("name is required")
	}

	if err := u.Role.IsValid(); err != nil {
		return fmt.Errorf("invalid role: %w", err)
	}

	if u.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	if u.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}

	return nil
}

// IsAdmin returns true if the user has admin role
func (u *User) IsAdmin() bool {
	return u.Role == RoleAdmin
}

// UpdateLastLogin updates the last login timestamp
func (u *User) UpdateLastLogin() {
	now := time.Now()
	u.LastLoginAt = &now
	u.UpdatedAt = now
}

// UserPreferences represents user notification preferences
type UserPreferences struct {
	ID                     uuid.UUID `json:"id"`
	UserID                 uuid.UUID `json:"user_id"`
	EmailNotifications     bool      `json:"email_notifications"`
	PushNotifications      bool      `json:"push_notifications"`
	DailyDigest            bool      `json:"daily_digest"`
	WeeklyDigest           bool      `json:"weekly_digest"`
	SecurityAlerts         bool      `json:"security_alerts"`
	TrendingTopics         bool      `json:"trending_topics"`
	SavedSearchAlerts      bool      `json:"saved_search_alerts"`
	BreakingNewsAlerts     bool      `json:"breaking_news_alerts"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

// Validate validates the user preferences
func (p *UserPreferences) Validate() error {
	if p.ID == uuid.Nil {
		return fmt.Errorf("preference ID is required")
	}

	if p.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if p.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	if p.UpdatedAt.IsZero() {
		return fmt.Errorf("updated_at is required")
	}

	return nil
}

// NewDefaultPreferences creates default preferences for a user
func NewDefaultPreferences(userID uuid.UUID) *UserPreferences {
	now := time.Now()
	return &UserPreferences{
		ID:                     uuid.New(),
		UserID:                 userID,
		EmailNotifications:     true,
		PushNotifications:      true,
		DailyDigest:            false,
		WeeklyDigest:           true,
		SecurityAlerts:         true,
		TrendingTopics:         true,
		SavedSearchAlerts:      true,
		BreakingNewsAlerts:     true,
		CreatedAt:              now,
		UpdatedAt:              now,
	}
}

// RefreshToken represents a refresh token for authentication
type RefreshToken struct {
	ID           uuid.UUID  `json:"id"`
	UserID       uuid.UUID  `json:"user_id"`
	Token        string     `json:"token"`
	ExpiresAt    time.Time  `json:"expires_at"`
	CreatedAt    time.Time  `json:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	LastUsedAt   *time.Time `json:"last_used_at,omitempty"`
	IPAddress    string     `json:"ip_address"`
	UserAgent    string     `json:"user_agent"`
}

// Validate validates the refresh token
func (t *RefreshToken) Validate() error {
	if t.ID == uuid.Nil {
		return fmt.Errorf("token ID is required")
	}

	if t.UserID == uuid.Nil {
		return fmt.Errorf("user ID is required")
	}

	if t.Token == "" {
		return fmt.Errorf("token is required")
	}

	if t.ExpiresAt.IsZero() {
		return fmt.Errorf("expires_at is required")
	}

	if t.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	return nil
}

// IsExpired checks if the token has expired
func (t *RefreshToken) IsExpired() bool {
	return time.Now().After(t.ExpiresAt)
}

// IsRevoked checks if the token has been revoked
func (t *RefreshToken) IsRevoked() bool {
	return t.RevokedAt != nil
}

// IsValid checks if the token is valid (not expired and not revoked)
func (t *RefreshToken) IsValid() bool {
	return !t.IsExpired() && !t.IsRevoked()
}

// Revoke marks the token as revoked
func (t *RefreshToken) Revoke() {
	now := time.Now()
	t.RevokedAt = &now
}

// UpdateLastUsed updates the last used timestamp
func (t *RefreshToken) UpdateLastUsed() {
	now := time.Now()
	t.LastUsedAt = &now
}
