package response

import (
	"net/http"
)

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorBody `json:"error"`
}

// ErrorBody contains error details
type ErrorBody struct {
	Code      string      `json:"code"`
	Message   string      `json:"message"`
	Details   interface{} `json:"details,omitempty"`
	RequestID string      `json:"request_id,omitempty"`
}

// Standard error codes
const (
	ErrCodeBadRequest      = "BAD_REQUEST"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeTooManyRequests = "TOO_MANY_REQUESTS"
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeServiceDown     = "SERVICE_UNAVAILABLE"
)

// ErrorWithDetails sends an error response with additional details and request ID
func ErrorWithDetails(w http.ResponseWriter, status int, code, message string, details interface{}, requestID string) {
	errResp := ErrorResponse{
		Error: ErrorBody{
			Code:      code,
			Message:   message,
			Details:   details,
			RequestID: requestID,
		},
	}

	JSON(w, status, errResp)
}

// BadRequest sends a 400 Bad Request error response
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, ErrCodeBadRequest, message)
}

// BadRequestWithDetails sends a 400 Bad Request error response with details
func BadRequestWithDetails(w http.ResponseWriter, message string, details interface{}, requestID string) {
	ErrorWithDetails(w, http.StatusBadRequest, ErrCodeBadRequest, message, details, requestID)
}

// Unauthorized sends a 401 Unauthorized error response
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Authentication required"
	}
	Error(w, http.StatusUnauthorized, ErrCodeUnauthorized, message)
}

// Forbidden sends a 403 Forbidden error response
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Access denied"
	}
	Error(w, http.StatusForbidden, ErrCodeForbidden, message)
}

// NotFound sends a 404 Not Found error response
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	Error(w, http.StatusNotFound, ErrCodeNotFound, message)
}

// Conflict sends a 409 Conflict error response
func Conflict(w http.ResponseWriter, message string) {
	Error(w, http.StatusConflict, ErrCodeConflict, message)
}

// TooManyRequests sends a 429 Too Many Requests error response
func TooManyRequests(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Rate limit exceeded"
	}
	Error(w, http.StatusTooManyRequests, ErrCodeTooManyRequests, message)
}

// InternalError sends a 500 Internal Server Error response
func InternalError(w http.ResponseWriter, message string, requestID string) {
	if message == "" {
		message = "An unexpected error occurred"
	}
	ErrorWithDetails(w, http.StatusInternalServerError, ErrCodeInternal, message, nil, requestID)
}

// ServiceUnavailable sends a 503 Service Unavailable error response
func ServiceUnavailable(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Service temporarily unavailable"
	}
	Error(w, http.StatusServiceUnavailable, ErrCodeServiceDown, message)
}

// ValidationError sends a 422 Unprocessable Entity error response with validation details
func ValidationError(w http.ResponseWriter, details interface{}, requestID string) {
	ErrorWithDetails(
		w,
		http.StatusUnprocessableEntity,
		ErrCodeValidation,
		"Validation failed",
		details,
		requestID,
	)
}
