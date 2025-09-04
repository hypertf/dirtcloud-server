package domain

import (
	"fmt"
)

// Error codes
const (
	ErrorCodeNotFound           = "NOT_FOUND"
	ErrorCodeAlreadyExists      = "ALREADY_EXISTS"
	ErrorCodeInvalidInput       = "INVALID_INPUT"
	ErrorCodeForeignKeyViolation = "FOREIGN_KEY_VIOLATION"
	ErrorCodeInternalError      = "INTERNAL_ERROR"
	ErrorCodeUnauthorized       = "UNAUTHORIZED"
	ErrorCodeTooManyRequests    = "TOO_MANY_REQUESTS"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
)

// DirtError represents a domain error with structured information
type DirtError struct {
	Code    string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

func (e *DirtError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// NewError creates a new domain error
func NewError(code, message string, details ...map[string]interface{}) *DirtError {
	var d map[string]interface{}
	if len(details) > 0 {
		d = details[0]
	}
	return &DirtError{
		Code:    code,
		Message: message,
		Details: d,
	}
}

// NotFoundError creates a not found error
func NotFoundError(resource string, identifier string) *DirtError {
	return NewError(ErrorCodeNotFound, fmt.Sprintf("%s not found", resource), map[string]interface{}{
		"resource":   resource,
		"identifier": identifier,
	})
}

// AlreadyExistsError creates an already exists error
func AlreadyExistsError(resource string, field string, value string) *DirtError {
	return NewError(ErrorCodeAlreadyExists, fmt.Sprintf("%s with %s '%s' already exists", resource, field, value), map[string]interface{}{
		"resource": resource,
		"field":    field,
		"value":    value,
	})
}

// InvalidInputError creates an invalid input error
func InvalidInputError(message string, details map[string]interface{}) *DirtError {
	return NewError(ErrorCodeInvalidInput, message, details)
}

// ForeignKeyViolationError creates a foreign key violation error
func ForeignKeyViolationError(resource string, field string, value string) *DirtError {
	return NewError(ErrorCodeForeignKeyViolation, fmt.Sprintf("Referenced %s with %s '%s' does not exist", resource, field, value), map[string]interface{}{
		"resource": resource,
		"field":    field,
		"value":    value,
	})
}

// InternalError creates an internal error
func InternalError(message string) *DirtError {
	return NewError(ErrorCodeInternalError, message)
}

// UnauthorizedError creates an unauthorized error
func UnauthorizedError(message string) *DirtError {
	return NewError(ErrorCodeUnauthorized, message)
}

// TooManyRequestsError creates a too many requests error
func TooManyRequestsError(message string) *DirtError {
	return NewError(ErrorCodeTooManyRequests, message)
}

// ServiceUnavailableError creates a service unavailable error
func ServiceUnavailableError(message string) *DirtError {
	return NewError(ErrorCodeServiceUnavailable, message)
}

// IsNotFound checks if error is a not found error
func IsNotFound(err error) bool {
	if dirtErr, ok := err.(*DirtError); ok {
		return dirtErr.Code == ErrorCodeNotFound
	}
	return false
}

// IsAlreadyExists checks if error is an already exists error
func IsAlreadyExists(err error) bool {
	if dirtErr, ok := err.(*DirtError); ok {
		return dirtErr.Code == ErrorCodeAlreadyExists
	}
	return false
}

// IsForeignKeyViolation checks if error is a foreign key violation error
func IsForeignKeyViolation(err error) bool {
	if dirtErr, ok := err.(*DirtError); ok {
		return dirtErr.Code == ErrorCodeForeignKeyViolation
	}
	return false
}

// IsInvalidInput checks if error is an invalid input error
func IsInvalidInput(err error) bool {
	if dirtErr, ok := err.(*DirtError); ok {
		return dirtErr.Code == ErrorCodeInvalidInput
	}
	return false
}