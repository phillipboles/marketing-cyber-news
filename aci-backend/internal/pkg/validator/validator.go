package validator

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-playground/validator/v10"
)

// Validator wraps go-playground/validator with custom validations
type Validator struct {
	v *validator.Validate
}

// ValidationErrors represents structured validation errors
type ValidationErrors struct {
	Errors []FieldError `json:"errors"`
}

// FieldError represents a single field validation error
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error implements the error interface
func (v *ValidationErrors) Error() string {
	if len(v.Errors) == 0 {
		return "validation failed"
	}

	var messages []string
	for _, err := range v.Errors {
		messages = append(messages, fmt.Sprintf("%s: %s", err.Field, err.Message))
	}

	return strings.Join(messages, "; ")
}

// New creates a new validator with custom validations
//
// Custom validations:
// - "password" - minimum 8 characters, at least 1 uppercase, 1 lowercase, 1 digit
// - "slug" - alphanumeric with hyphens only, no leading/trailing hyphens
// - "cve" - CVE-YYYY-NNNNN format (e.g., CVE-2024-12345)
//
// Example usage:
//
//	type User struct {
//	    Email    string `validate:"required,email"`
//	    Password string `validate:"required,password"`
//	    Username string `validate:"required,slug"`
//	}
//
//	v := validator.New()
//	if err := v.Validate(user); err != nil {
//	    // Handle validation errors
//	}
func New() *Validator {
	v := validator.New()

	// Register custom password validation
	v.RegisterValidation("password", validatePassword)

	// Register custom slug validation
	v.RegisterValidation("slug", validateSlug)

	// Register custom CVE format validation
	v.RegisterValidation("cve", validateCVE)

	return &Validator{v: v}
}

// Validate validates a struct and returns structured errors
//
// Returns ValidationErrors with detailed field-level error messages.
// Returns nil if validation passes.
//
// Example:
//
//	err := v.Validate(user)
//	if err != nil {
//	    if valErr, ok := err.(*ValidationErrors); ok {
//	        for _, fieldErr := range valErr.Errors {
//	            fmt.Printf("%s: %s\n", fieldErr.Field, fieldErr.Message)
//	        }
//	    }
//	}
func (v *Validator) Validate(i interface{}) error {
	if i == nil {
		return fmt.Errorf("cannot validate nil value")
	}

	err := v.v.Struct(i)
	if err == nil {
		return nil
	}

	validationErrs, ok := err.(validator.ValidationErrors)
	if !ok {
		return fmt.Errorf("validation failed: %w", err)
	}

	var fieldErrors []FieldError
	for _, err := range validationErrs {
		fieldErrors = append(fieldErrors, FieldError{
			Field:   err.Field(),
			Message: getErrorMessage(err),
		})
	}

	return &ValidationErrors{Errors: fieldErrors}
}

// validatePassword ensures password meets security requirements
// - Minimum 8 characters
// - At least 1 uppercase letter
// - At least 1 lowercase letter
// - At least 1 digit
func validatePassword(fl validator.FieldLevel) bool {
	password := fl.Field().String()

	if len(password) < 8 {
		return false
	}

	var (
		hasUpper = false
		hasLower = false
		hasDigit = false
	)

	for _, char := range password {
		switch {
		case char >= 'A' && char <= 'Z':
			hasUpper = true
		case char >= 'a' && char <= 'z':
			hasLower = true
		case char >= '0' && char <= '9':
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}

// validateSlug ensures slug is URL-friendly
// - Only lowercase letters, numbers, and hyphens
// - No leading or trailing hyphens
// - No consecutive hyphens
func validateSlug(fl validator.FieldLevel) bool {
	slug := fl.Field().String()

	if slug == "" {
		return false
	}

	// Check for valid characters only
	slugPattern := regexp.MustCompile(`^[a-z0-9]+(?:-[a-z0-9]+)*$`)
	return slugPattern.MatchString(slug)
}

// validateCVE ensures CVE ID follows correct format
// Format: CVE-YYYY-NNNNN where YYYY is year and NNNNN is at least 4 digits
// Examples: CVE-2024-12345, CVE-2023-0001
func validateCVE(fl validator.FieldLevel) bool {
	cve := fl.Field().String()

	cvePattern := regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)
	return cvePattern.MatchString(cve)
}

// getErrorMessage returns a user-friendly error message for validation errors
func getErrorMessage(err validator.FieldError) string {
	field := err.Field()
	tag := err.Tag()

	switch tag {
	case "required":
		return fmt.Sprintf("%s is required", field)
	case "email":
		return fmt.Sprintf("%s must be a valid email address", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s characters", field, err.Param())
	case "max":
		return fmt.Sprintf("%s must not exceed %s characters", field, err.Param())
	case "url":
		return fmt.Sprintf("%s must be a valid URL", field)
	case "password":
		return fmt.Sprintf("%s must be at least 8 characters with 1 uppercase, 1 lowercase, and 1 digit", field)
	case "slug":
		return fmt.Sprintf("%s must contain only lowercase letters, numbers, and hyphens", field)
	case "cve":
		return fmt.Sprintf("%s must be in CVE-YYYY-NNNNN format", field)
	case "oneof":
		return fmt.Sprintf("%s must be one of: %s", field, err.Param())
	case "gt":
		return fmt.Sprintf("%s must be greater than %s", field, err.Param())
	case "gte":
		return fmt.Sprintf("%s must be greater than or equal to %s", field, err.Param())
	case "lt":
		return fmt.Sprintf("%s must be less than %s", field, err.Param())
	case "lte":
		return fmt.Sprintf("%s must be less than or equal to %s", field, err.Param())
	default:
		return fmt.Sprintf("%s failed validation: %s", field, tag)
	}
}
