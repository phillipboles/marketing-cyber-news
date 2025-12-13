package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// AuditLog represents an audit trail entry for admin actions
type AuditLog struct {
	ID           uuid.UUID   `json:"id"`
	UserID       *uuid.UUID  `json:"user_id,omitempty"`
	UserEmail    *string     `json:"user_email,omitempty"` // Denormalized for query performance
	Action       string      `json:"action"`
	ResourceType string      `json:"resource_type"`
	ResourceID   *uuid.UUID  `json:"resource_id,omitempty"`
	OldValue     interface{} `json:"old_value,omitempty"`
	NewValue     interface{} `json:"new_value,omitempty"`
	IPAddress    *string     `json:"ip_address,omitempty"`
	UserAgent    *string     `json:"user_agent,omitempty"`
	CreatedAt    time.Time   `json:"created_at"`
}

// Validate validates the audit log entity
func (a *AuditLog) Validate() error {
	if a.ID == uuid.Nil {
		return fmt.Errorf("audit log ID is required")
	}

	if a.Action == "" {
		return fmt.Errorf("action is required")
	}

	if len(a.Action) > 100 {
		return fmt.Errorf("action must not exceed 100 characters")
	}

	if a.ResourceType == "" {
		return fmt.Errorf("resource type is required")
	}

	if len(a.ResourceType) > 100 {
		return fmt.Errorf("resource type must not exceed 100 characters")
	}

	if a.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	return nil
}

// AuditLogFilter represents filter criteria for listing audit logs
type AuditLogFilter struct {
	UserID       *uuid.UUID
	Action       *string
	ResourceType *string
	ResourceID   *uuid.UUID
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

// Validate validates the audit log filter
func (f *AuditLogFilter) Validate() error {
	if f.Limit < 0 {
		return fmt.Errorf("limit must be non-negative")
	}

	if f.Offset < 0 {
		return fmt.Errorf("offset must be non-negative")
	}

	if f.StartDate != nil && f.EndDate != nil && f.StartDate.After(*f.EndDate) {
		return fmt.Errorf("start_date must be before end_date")
	}

	return nil
}

// NewAuditLog creates a new audit log entry
func NewAuditLog(
	userID *uuid.UUID,
	action string,
	resourceType string,
	resourceID *uuid.UUID,
	oldValue interface{},
	newValue interface{},
	ipAddress *string,
	userAgent *string,
) *AuditLog {
	return &AuditLog{
		ID:           uuid.New(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		OldValue:     oldValue,
		NewValue:     newValue,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		CreatedAt:    time.Now(),
	}
}
