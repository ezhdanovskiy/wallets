package httperr

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		format     string
		args       []interface{}
		expected   *Error
	}{
		{
			name:       "basic error",
			statusCode: http.StatusBadRequest,
			format:     "test error",
			args:       nil,
			expected: &Error{
				Message:    "test error",
				StatusCode: http.StatusBadRequest,
			},
		},
		{
			name:       "error with formatting",
			statusCode: http.StatusNotFound,
			format:     "item %s not found with id %d",
			args:       []interface{}{"wallet", 123},
			expected: &Error{
				Message:    "item wallet not found with id 123",
				StatusCode: http.StatusNotFound,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := New(tt.statusCode, tt.format, tt.args...)
			assert.Equal(t, tt.expected.Message, err.Message)
			assert.Equal(t, tt.expected.StatusCode, err.StatusCode)
			assert.Nil(t, err.Err)
		})
	}
}

func TestError(t *testing.T) {
	tests := []struct {
		name     string
		err      *Error
		expected string
	}{
		{
			name: "error without wrapped error",
			err: &Error{
				Message:    "test error",
				StatusCode: http.StatusBadRequest,
			},
			expected: "test error",
		},
		{
			name: "error with wrapped error",
			err: &Error{
				Message:    "test error",
				StatusCode: http.StatusInternalServerError,
				Err:        errors.New("wrapped error"),
			},
			expected: "test error: wrapped error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestWrap(t *testing.T) {
	originalErr := &Error{
		Message:    "original error",
		StatusCode: http.StatusBadRequest,
	}

	wrappedErr := errors.New("underlying error")
	newErr := originalErr.Wrap(wrappedErr)

	assert.Equal(t, originalErr.Message, newErr.Message)
	assert.Equal(t, originalErr.StatusCode, newErr.StatusCode)
	assert.Equal(t, wrappedErr, newErr.Err)
}

func TestWrapFunction(t *testing.T) {
	underlyingErr := errors.New("database error")
	err := Wrap(underlyingErr, http.StatusInternalServerError, "failed to get item %s", "wallet")

	assert.Equal(t, "failed to get item wallet", err.Message)
	assert.Equal(t, http.StatusInternalServerError, err.StatusCode)
	assert.Equal(t, underlyingErr, err.Err)
	assert.Equal(t, "failed to get item wallet: database error", err.Error())
}