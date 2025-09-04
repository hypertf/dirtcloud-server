package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirtError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *DirtError
		expected string
	}{
		{
			name: "simple error",
			err: &DirtError{
				Code:    ErrorCodeNotFound,
				Message: "resource not found",
			},
			expected: "NOT_FOUND: resource not found",
		},
		{
			name: "error with details",
			err: &DirtError{
				Code:    ErrorCodeInvalidInput,
				Message: "validation failed",
				Details: map[string]interface{}{"field": "name"},
			},
			expected: "INVALID_INPUT: validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestNewError(t *testing.T) {
	tests := []struct {
		name     string
		code     string
		message  string
		details  []map[string]interface{}
		expected *DirtError
	}{
		{
			name:    "without details",
			code:    ErrorCodeNotFound,
			message: "not found",
			expected: &DirtError{
				Code:    ErrorCodeNotFound,
				Message: "not found",
				Details: nil,
			},
		},
		{
			name:    "with details",
			code:    ErrorCodeInvalidInput,
			message: "invalid",
			details: []map[string]interface{}{{"field": "name"}},
			expected: &DirtError{
				Code:    ErrorCodeInvalidInput,
				Message: "invalid",
				Details: map[string]interface{}{"field": "name"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewError(tt.code, tt.message, tt.details...)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNotFoundError(t *testing.T) {
	err := NotFoundError("project", "123")
	
	assert.Equal(t, ErrorCodeNotFound, err.Code)
	assert.Equal(t, "project not found", err.Message)
	assert.Equal(t, map[string]interface{}{
		"resource":   "project",
		"identifier": "123",
	}, err.Details)
}

func TestAlreadyExistsError(t *testing.T) {
	err := AlreadyExistsError("project", "name", "test")
	
	assert.Equal(t, ErrorCodeAlreadyExists, err.Code)
	assert.Equal(t, "project with name 'test' already exists", err.Message)
	assert.Equal(t, map[string]interface{}{
		"resource": "project",
		"field":    "name",
		"value":    "test",
	}, err.Details)
}

func TestInvalidInputError(t *testing.T) {
	details := map[string]interface{}{"field": "cpu", "min": 1}
	err := InvalidInputError("CPU must be positive", details)
	
	assert.Equal(t, ErrorCodeInvalidInput, err.Code)
	assert.Equal(t, "CPU must be positive", err.Message)
	assert.Equal(t, details, err.Details)
}

func TestForeignKeyViolationError(t *testing.T) {
	err := ForeignKeyViolationError("project", "id", "123")
	
	assert.Equal(t, ErrorCodeForeignKeyViolation, err.Code)
	assert.Equal(t, "Referenced project with id '123' does not exist", err.Message)
	assert.Equal(t, map[string]interface{}{
		"resource": "project",
		"field":    "id",
		"value":    "123",
	}, err.Details)
}

func TestInternalError(t *testing.T) {
	err := InternalError("something went wrong")
	
	assert.Equal(t, ErrorCodeInternalError, err.Code)
	assert.Equal(t, "something went wrong", err.Message)
	assert.Nil(t, err.Details)
}

func TestUnauthorizedError(t *testing.T) {
	err := UnauthorizedError("invalid token")
	
	assert.Equal(t, ErrorCodeUnauthorized, err.Code)
	assert.Equal(t, "invalid token", err.Message)
	assert.Nil(t, err.Details)
}

func TestTooManyRequestsError(t *testing.T) {
	err := TooManyRequestsError("rate limited")
	
	assert.Equal(t, ErrorCodeTooManyRequests, err.Code)
	assert.Equal(t, "rate limited", err.Message)
	assert.Nil(t, err.Details)
}

func TestServiceUnavailableError(t *testing.T) {
	err := ServiceUnavailableError("service down")
	
	assert.Equal(t, ErrorCodeServiceUnavailable, err.Code)
	assert.Equal(t, "service down", err.Message)
	assert.Nil(t, err.Details)
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "not found error",
			err:      NotFoundError("project", "123"),
			expected: true,
		},
		{
			name:     "invalid input error",
			err:      InvalidInputError("bad input", nil),
			expected: false,
		},
		{
			name:     "generic error",
			err:      assert.AnError,
			expected: false,
		},
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsNotFound(tt.err))
		})
	}
}

func TestIsAlreadyExists(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "already exists error",
			err:      AlreadyExistsError("project", "name", "test"),
			expected: true,
		},
		{
			name:     "not found error",
			err:      NotFoundError("project", "123"),
			expected: false,
		},
		{
			name:     "generic error",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsAlreadyExists(tt.err))
		})
	}
}

func TestIsForeignKeyViolation(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "foreign key violation error",
			err:      ForeignKeyViolationError("project", "id", "123"),
			expected: true,
		},
		{
			name:     "not found error",
			err:      NotFoundError("project", "123"),
			expected: false,
		},
		{
			name:     "generic error",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsForeignKeyViolation(tt.err))
		})
	}
}

func TestIsInvalidInput(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "invalid input error",
			err:      InvalidInputError("bad input", nil),
			expected: true,
		},
		{
			name:     "not found error",
			err:      NotFoundError("project", "123"),
			expected: false,
		},
		{
			name:     "generic error",
			err:      assert.AnError,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, IsInvalidInput(tt.err))
		})
	}
}