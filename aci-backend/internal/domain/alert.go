package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type AlertType string

const (
	AlertTypeKeyword  AlertType = "keyword"
	AlertTypeCategory AlertType = "category"
	AlertTypeSeverity AlertType = "severity"
	AlertTypeVendor   AlertType = "vendor"
	AlertTypeCVE      AlertType = "cve"
)

// IsValid validates the alert type value
func (t AlertType) IsValid() bool {
	switch t {
	case AlertTypeKeyword, AlertTypeCategory, AlertTypeSeverity, AlertTypeVendor, AlertTypeCVE:
		return true
	default:
		return false
	}
}

// Alert represents a user-configured alert
type Alert struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Name      string    `json:"name"`
	Type      AlertType `json:"type"`
	Value     string    `json:"value"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Statistics (populated on query)
	MatchCount int `json:"match_count,omitempty"`
}

// Validate performs validation on the Alert
func (a *Alert) Validate() error {
	if a.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}

	if a.Name == "" {
		return fmt.Errorf("name is required")
	}

	if !a.Type.IsValid() {
		return fmt.Errorf("invalid alert type")
	}

	if a.Value == "" {
		return fmt.Errorf("value is required")
	}

	// Type-specific validation
	switch a.Type {
	case AlertTypeSeverity:
		severity := Severity(a.Value)
		if !severity.IsValid() {
			return fmt.Errorf("invalid severity value for alert")
		}
	case AlertTypeCategory:
		if _, err := uuid.Parse(a.Value); err != nil {
			return fmt.Errorf("category alert value must be a valid UUID")
		}
	}

	return nil
}

// Matches checks if the alert matches the given article
func (a *Alert) Matches(article *Article) bool {
	if article == nil {
		return false
	}

	if !a.IsActive {
		return false
	}

	switch a.Type {
	case AlertTypeKeyword:
		return article.ContainsKeyword(a.Value)

	case AlertTypeCategory:
		categoryID, err := uuid.Parse(a.Value)
		if err != nil {
			return false
		}
		return article.CategoryID == categoryID

	case AlertTypeSeverity:
		return strings.EqualFold(string(article.Severity), a.Value)

	case AlertTypeVendor:
		return article.HasVendor(a.Value)

	case AlertTypeCVE:
		return article.HasCVE(a.Value)

	default:
		return false
	}
}

// AlertMatch records when an alert matches an article
type AlertMatch struct {
	ID         uuid.UUID  `json:"id"`
	AlertID    uuid.UUID  `json:"alert_id"`
	ArticleID  uuid.UUID  `json:"article_id"`
	Priority   string     `json:"priority"` // critical, high, normal
	MatchedAt  time.Time  `json:"matched_at"`
	NotifiedAt *time.Time `json:"notified_at,omitempty"`

	// Populated on query
	Alert   *Alert   `json:"alert,omitempty"`
	Article *Article `json:"article,omitempty"`
}

// Validate performs validation on the AlertMatch
func (m *AlertMatch) Validate() error {
	if m.AlertID == uuid.Nil {
		return fmt.Errorf("alert_id is required")
	}

	if m.ArticleID == uuid.Nil {
		return fmt.Errorf("article_id is required")
	}

	if m.Priority == "" {
		return fmt.Errorf("priority is required")
	}

	validPriorities := map[string]bool{
		"critical": true,
		"high":     true,
		"normal":   true,
	}

	if !validPriorities[m.Priority] {
		return fmt.Errorf("priority must be critical, high, or normal")
	}

	return nil
}

// IsNotified returns true if the match has been notified
func (m *AlertMatch) IsNotified() bool {
	return m.NotifiedAt != nil
}

// MarkNotified marks the match as notified
func (m *AlertMatch) MarkNotified() {
	now := time.Now()
	m.NotifiedAt = &now
}

// DeterminePriority determines the priority based on article severity
func DeterminePriority(article *Article) string {
	if article == nil {
		return "normal"
	}

	switch article.Severity {
	case SeverityCritical:
		return "critical"
	case SeverityHigh:
		return "high"
	default:
		return "normal"
	}
}

// AlertFilter represents query parameters for filtering alerts
type AlertFilter struct {
	UserID   uuid.UUID
	Type     *AlertType
	IsActive *bool
	Page     int
	PageSize int
}

// NewAlertFilter returns a filter with default values
func NewAlertFilter(userID uuid.UUID) *AlertFilter {
	return &AlertFilter{
		UserID:   userID,
		Page:     1,
		PageSize: 20,
	}
}

// Validate validates the filter parameters
func (f *AlertFilter) Validate() error {
	if f.UserID == uuid.Nil {
		return fmt.Errorf("user_id is required")
	}

	if f.Page < 1 {
		return fmt.Errorf("page must be at least 1")
	}

	if f.PageSize < 1 {
		return fmt.Errorf("page_size must be at least 1")
	}

	if f.PageSize > 100 {
		return fmt.Errorf("page_size cannot exceed 100")
	}

	if f.Type != nil && !f.Type.IsValid() {
		return fmt.Errorf("invalid alert type")
	}

	return nil
}

// Offset calculates the offset for pagination
func (f *AlertFilter) Offset() int {
	return (f.Page - 1) * f.PageSize
}

// AlertMatchFilter represents query parameters for filtering alert matches
type AlertMatchFilter struct {
	AlertID   *uuid.UUID
	ArticleID *uuid.UUID
	UserID    *uuid.UUID
	Priority  *string
	Notified  *bool
	DateFrom  *time.Time
	DateTo    *time.Time
	Page      int
	PageSize  int
}

// NewAlertMatchFilter returns a filter with default values
func NewAlertMatchFilter() *AlertMatchFilter {
	return &AlertMatchFilter{
		Page:     1,
		PageSize: 20,
	}
}

// Validate validates the filter parameters
func (f *AlertMatchFilter) Validate() error {
	if f.Page < 1 {
		return fmt.Errorf("page must be at least 1")
	}

	if f.PageSize < 1 {
		return fmt.Errorf("page_size must be at least 1")
	}

	if f.PageSize > 100 {
		return fmt.Errorf("page_size cannot exceed 100")
	}

	if f.Priority != nil {
		validPriorities := map[string]bool{
			"critical": true,
			"high":     true,
			"normal":   true,
		}

		if !validPriorities[*f.Priority] {
			return fmt.Errorf("priority must be critical, high, or normal")
		}
	}

	if f.DateFrom != nil && f.DateTo != nil && f.DateFrom.After(*f.DateTo) {
		return fmt.Errorf("date_from cannot be after date_to")
	}

	return nil
}

// Offset calculates the offset for pagination
func (f *AlertMatchFilter) Offset() int {
	return (f.Page - 1) * f.PageSize
}
