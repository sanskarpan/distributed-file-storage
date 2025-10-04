package errors

import (
	"errors"
	"testing"
)

func TestFileSystemError(t *testing.T) {
	err := New(NetworkError, "connection failed")
	
	if err.Type != NetworkError {
		t.Errorf("Expected NetworkError, got %v", err.Type)
	}
	
	if err.Message != "connection failed" {
		t.Errorf("Expected 'connection failed', got %s", err.Message)
	}
	
	expectedStr := "[NETWORK_ERROR] connection failed"
	if err.Error() != expectedStr {
		t.Errorf("Expected '%s', got '%s'", expectedStr, err.Error())
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	wrappedErr := Wrap(originalErr, StorageError, "storage operation failed")
	
	if wrappedErr.Type != StorageError {
		t.Errorf("Expected StorageError, got %v", wrappedErr.Type)
	}
	
	if wrappedErr.Cause != originalErr {
		t.Error("Expected wrapped error to contain original error")
	}
	
	if wrappedErr.Unwrap() != originalErr {
		t.Error("Expected Unwrap to return original error")
	}
}

func TestErrorWithContext(t *testing.T) {
	err := New(ValidationError, "invalid input").
		WithContext("field", "username").
		WithContext("value", "test@example.com").
		WithOperation("user_creation")
	
	if err.Context["field"] != "username" {
		t.Error("Expected context field to be set")
	}
	
	if err.Context["value"] != "test@example.com" {
		t.Error("Expected context value to be set")
	}
	
	if err.Operation != "user_creation" {
		t.Error("Expected operation to be set")
	}
}

func TestIsType(t *testing.T) {
	err := New(NetworkError, "network failed")
	
	if !IsType(err, NetworkError) {
		t.Error("Expected IsType to return true for NetworkError")
	}
	
	if IsType(err, StorageError) {
		t.Error("Expected IsType to return false for StorageError")
	}
	
	// Test with regular error
	regularErr := errors.New("regular error")
	if IsType(regularErr, NetworkError) {
		t.Error("Expected IsType to return false for regular error")
	}
}

func TestGetType(t *testing.T) {
	err := New(TimeoutError, "timeout occurred")
	
	if GetType(err) != TimeoutError {
		t.Errorf("Expected TimeoutError, got %v", GetType(err))
	}
	
	// Test with regular error
	regularErr := errors.New("regular error")
	if GetType(regularErr) != InternalError {
		t.Errorf("Expected InternalError for regular error, got %v", GetType(regularErr))
	}
}

func TestIsRetryable(t *testing.T) {
	tests := []struct {
		errorType ErrorType
		retryable bool
	}{
		{NetworkError, true},
		{ConnectionError, true},
		{TimeoutError, true},
		{StorageError, true},
		{AuthenticationError, false},
		{ValidationError, false},
		{CorruptionError, false},
	}

	for _, test := range tests {
		err := New(test.errorType, "test error")
		if IsRetryable(err) != test.retryable {
			t.Errorf("Expected %v to be retryable=%v", test.errorType, test.retryable)
		}
	}
}

func TestConvenienceFunctions(t *testing.T) {
	tests := []struct {
		name     string
		createFn func() *FileSystemError
		expected ErrorType
	}{
		{"NetworkError", func() *FileSystemError { return NewNetworkError("test") }, NetworkError},
		{"ConnectionError", func() *FileSystemError { return NewConnectionError("test") }, ConnectionError},
		{"TimeoutError", func() *FileSystemError { return NewTimeoutError("test") }, TimeoutError},
		{"StorageError", func() *FileSystemError { return NewStorageError("test") }, StorageError},
		{"FileNotFoundError", func() *FileSystemError { return NewFileNotFoundError("test.txt") }, FileNotFoundError},
		{"CorruptionError", func() *FileSystemError { return NewCorruptionError("test") }, CorruptionError},
		{"AuthenticationError", func() *FileSystemError { return NewAuthenticationError("test") }, AuthenticationError},
		{"ConfigError", func() *FileSystemError { return NewConfigError("test") }, ConfigError},
		{"ValidationError", func() *FileSystemError { return NewValidationError("test") }, ValidationError},
		{"InternalError", func() *FileSystemError { return NewInternalError("test") }, InternalError},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.createFn()
			if err.Type != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, err.Type)
			}
		})
	}
}
