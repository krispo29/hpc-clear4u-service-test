package errors

import (
	"fmt"
	"net/http"

	"hpc-express-service/constant"

	"github.com/go-chi/render"
)

// Custom error types for MAWB system integration

// MAWBNotFoundError represents an error when MAWB Info is not found
type MAWBNotFoundError struct {
	UUID string
}

func (e MAWBNotFoundError) Error() string {
	return fmt.Sprintf("MAWB Info not found: %s", e.UUID)
}

// CargoManifestNotFoundError represents an error when Cargo Manifest is not found
type CargoManifestNotFoundError struct {
	MAWBUUID string
}

func (e CargoManifestNotFoundError) Error() string {
	return fmt.Sprintf("Cargo Manifest not found for MAWB: %s", e.MAWBUUID)
}

// DraftMAWBNotFoundError represents an error when Draft MAWB is not found
type DraftMAWBNotFoundError struct {
	MAWBUUID string
}

func (e DraftMAWBNotFoundError) Error() string {
	return fmt.Sprintf("Draft MAWB not found for MAWB: %s", e.MAWBUUID)
}

// ValidationError represents validation errors for business rule violations
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("validation error: %s", e.Message)
}

// NewValidationError creates a new validation error
func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

// BusinessRuleError represents business logic rule violations
type BusinessRuleError struct {
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

func (e BusinessRuleError) Error() string {
	return fmt.Sprintf("business rule violation '%s': %s", e.Rule, e.Message)
}

// NewBusinessRuleError creates a new business rule error
func NewBusinessRuleError(rule, message string) BusinessRuleError {
	return BusinessRuleError{
		Rule:    rule,
		Message: message,
	}
}

// DatabaseError represents database operation errors
type DatabaseError struct {
	Operation string
	Err       error
}

func (e DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Err)
}

// NewDatabaseError creates a new database error
func NewDatabaseError(operation string, err error) DatabaseError {
	return DatabaseError{
		Operation: operation,
		Err:       err,
	}
}

// ErrResponse represents the standard error response structure
type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText string `json:"status,omitempty"`  // user-level status message
	AppCode    int64  `json:"code,omitempty"`    // application-specific error code
	Message    string `json:"message,omitempty"` // application-level error message, for debugging
}

func (e *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, e.HTTPStatusCode)
	return nil
}

// MapErrorToHTTPResponse maps custom errors to appropriate HTTP responses
func MapErrorToHTTPResponse(err error) render.Renderer {
	switch e := err.(type) {
	case MAWBNotFoundError, CargoManifestNotFoundError, DraftMAWBNotFoundError:
		return &ErrResponse{
			Err:            err,
			HTTPStatusCode: http.StatusNotFound,
			AppCode:        constant.CodeError,
			StatusText:     "Not Found",
			Message:        e.Error(),
		}
	case ValidationError:
		return &ErrResponse{
			Err:            err,
			HTTPStatusCode: http.StatusBadRequest,
			AppCode:        constant.CodeError,
			StatusText:     "Validation Error",
			Message:        e.Error(),
		}
	case BusinessRuleError:
		return &ErrResponse{
			Err:            err,
			HTTPStatusCode: http.StatusBadRequest,
			AppCode:        constant.CodeError,
			StatusText:     "Business Rule Violation",
			Message:        e.Error(),
		}
	case DatabaseError:
		return &ErrResponse{
			Err:            err,
			HTTPStatusCode: http.StatusInternalServerError,
			AppCode:        constant.CodeError,
			StatusText:     "Internal Server Error",
			Message:        "Database operation failed",
		}
	default:
		// Handle standard errors and unknown errors
		return &ErrResponse{
			Err:            err,
			HTTPStatusCode: http.StatusInternalServerError,
			AppCode:        constant.CodeError,
			StatusText:     "Internal Server Error",
			Message:        "An unexpected error occurred",
		}
	}
}

// Helper functions for creating common errors

// NewMAWBNotFoundError creates a new MAWB not found error
func NewMAWBNotFoundError(uuid string) MAWBNotFoundError {
	return MAWBNotFoundError{UUID: uuid}
}

// NewCargoManifestNotFoundError creates a new cargo manifest not found error
func NewCargoManifestNotFoundError(mawbUUID string) CargoManifestNotFoundError {
	return CargoManifestNotFoundError{MAWBUUID: mawbUUID}
}

// NewDraftMAWBNotFoundError creates a new draft MAWB not found error
func NewDraftMAWBNotFoundError(mawbUUID string) DraftMAWBNotFoundError {
	return DraftMAWBNotFoundError{MAWBUUID: mawbUUID}
}
