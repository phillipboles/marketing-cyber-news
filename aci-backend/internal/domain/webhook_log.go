package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// WebhookStatus represents the processing status of a webhook event
type WebhookStatus string

const (
	WebhookStatusPending    WebhookStatus = "pending"
	WebhookStatusProcessing WebhookStatus = "processing"
	WebhookStatusSuccess    WebhookStatus = "success"
	WebhookStatusFailed     WebhookStatus = "failed"
)

// IsValid validates the webhook status value
func (s WebhookStatus) IsValid() bool {
	switch s {
	case WebhookStatusPending, WebhookStatusProcessing, WebhookStatusSuccess, WebhookStatusFailed:
		return true
	default:
		return false
	}
}

// WebhookLog represents a logged webhook event
type WebhookLog struct {
	ID          uuid.UUID     `json:"id"`
	EventType   string        `json:"event_type"`
	Status      WebhookStatus `json:"status"`
	Payload     string        `json:"payload"`
	WorkflowID  *string       `json:"workflow_id,omitempty"`
	ExecutionID *string       `json:"execution_id,omitempty"`
	ErrorMsg    *string       `json:"error_msg,omitempty"`
	ProcessedAt *time.Time    `json:"processed_at,omitempty"`
	CreatedAt   time.Time     `json:"created_at"`
}

// Validate validates the webhook log entity
func (w *WebhookLog) Validate() error {
	if w.ID == uuid.Nil {
		return fmt.Errorf("webhook log ID is required")
	}

	if w.EventType == "" {
		return fmt.Errorf("event_type is required")
	}

	if !w.Status.IsValid() {
		return fmt.Errorf("invalid webhook status")
	}

	if w.Payload == "" {
		return fmt.Errorf("payload is required")
	}

	if w.CreatedAt.IsZero() {
		return fmt.Errorf("created_at is required")
	}

	return nil
}

// MarkProcessing updates the status to processing
func (w *WebhookLog) MarkProcessing() {
	w.Status = WebhookStatusProcessing
}

// MarkSuccess updates the status to success
func (w *WebhookLog) MarkSuccess() {
	w.Status = WebhookStatusSuccess
	now := time.Now()
	w.ProcessedAt = &now
}

// MarkFailed updates the status to failed with error message
func (w *WebhookLog) MarkFailed(errMsg string) {
	w.Status = WebhookStatusFailed
	w.ErrorMsg = &errMsg
	now := time.Now()
	w.ProcessedAt = &now
}

// NewWebhookLog creates a new webhook log entry
func NewWebhookLog(eventType, payload string, workflowID, executionID *string) *WebhookLog {
	return &WebhookLog{
		ID:          uuid.New(),
		EventType:   eventType,
		Status:      WebhookStatusPending,
		Payload:     payload,
		WorkflowID:  workflowID,
		ExecutionID: executionID,
		CreatedAt:   time.Now(),
	}
}
