package errors

import "fmt"

// Domain errors - clean, semantic error types for business logic

var (
	ErrNotFound      = fmt.Errorf("resource not found")
	ErrUnauthorized  = fmt.Errorf("unauthorized access")
	ErrForbidden     = fmt.Errorf("forbidden operation")
	ErrConflict      = fmt.Errorf("resource conflict")
	ErrInvalidInput  = fmt.Errorf("invalid input")
	ErrInternal      = fmt.Errorf("internal error")
)

// NotFoundError represents a resource not found error
type NotFoundError struct {
	Resource string
	ID       string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("%s not found: %s", e.Resource, e.ID)
}

// ValidationError represents a validation failure
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation failed for %s: %s", e.Field, e.Message)
}

// ConflictError represents a resource conflict
type ConflictError struct {
	Resource string
	Field    string
	Value    string
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("%s already exists with %s: %s", e.Resource, e.Field, e.Value)
}
