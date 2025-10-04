package errors

import (
	"fmt"
	"runtime"
)

// ErrorType represents different types of errors in the system
type ErrorType string

const (
	// Network related errors
	NetworkError     ErrorType = "NETWORK_ERROR"
	ConnectionError  ErrorType = "CONNECTION_ERROR"
	TimeoutError     ErrorType = "TIMEOUT_ERROR"
	
	// Storage related errors
	StorageError     ErrorType = "STORAGE_ERROR"
	FileNotFoundError ErrorType = "FILE_NOT_FOUND"
	CorruptionError  ErrorType = "CORRUPTION_ERROR"
	QuotaExceededError ErrorType = "QUOTA_EXCEEDED"
	
	// Security related errors
	AuthenticationError ErrorType = "AUTHENTICATION_ERROR"
	AuthorizationError  ErrorType = "AUTHORIZATION_ERROR"
	EncryptionError     ErrorType = "ENCRYPTION_ERROR"
	
	// Configuration related errors
	ConfigError      ErrorType = "CONFIG_ERROR"
	ValidationError  ErrorType = "VALIDATION_ERROR"
	
	// General errors
	InternalError    ErrorType = "INTERNAL_ERROR"
	InvalidInputError ErrorType = "INVALID_INPUT"
)

// FileSystemError represents a structured error with context
type FileSystemError struct {
	Type      ErrorType
	Message   string
	Cause     error
	File      string
	Line      int
	Operation string
	Context   map[string]interface{}
}

// Error implements the error interface
func (e *FileSystemError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("[%s] %s: %v", e.Type, e.Message, e.Cause)
	}
	return fmt.Sprintf("[%s] %s", e.Type, e.Message)
}

// Unwrap returns the underlying error
func (e *FileSystemError) Unwrap() error {
	return e.Cause
}

// WithContext adds context information to the error
func (e *FileSystemError) WithContext(key string, value interface{}) *FileSystemError {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithOperation sets the operation that caused the error
func (e *FileSystemError) WithOperation(operation string) *FileSystemError {
	e.Operation = operation
	return e
}

// New creates a new FileSystemError
func New(errorType ErrorType, message string) *FileSystemError {
	_, file, line, _ := runtime.Caller(1)
	return &FileSystemError{
		Type:    errorType,
		Message: message,
		File:    file,
		Line:    line,
		Context: make(map[string]interface{}),
	}
}

// Wrap wraps an existing error with additional context
func Wrap(err error, errorType ErrorType, message string) *FileSystemError {
	_, file, line, _ := runtime.Caller(1)
	return &FileSystemError{
		Type:    errorType,
		Message: message,
		Cause:   err,
		File:    file,
		Line:    line,
		Context: make(map[string]interface{}),
	}
}

// IsType checks if an error is of a specific type
func IsType(err error, errorType ErrorType) bool {
	if fsErr, ok := err.(*FileSystemError); ok {
		return fsErr.Type == errorType
	}
	return false
}

// GetType returns the error type if it's a FileSystemError
func GetType(err error) ErrorType {
	if fsErr, ok := err.(*FileSystemError); ok {
		return fsErr.Type
	}
	return InternalError
}

// IsRetryable determines if an error is retryable
func IsRetryable(err error) bool {
	errorType := GetType(err)
	switch errorType {
	case NetworkError, ConnectionError, TimeoutError:
		return true
	case StorageError:
		return true
	default:
		return false
	}
}

// Convenience functions for common error types

// NewNetworkError creates a new network error
func NewNetworkError(message string) *FileSystemError {
	return New(NetworkError, message)
}

// NewConnectionError creates a new connection error
func NewConnectionError(message string) *FileSystemError {
	return New(ConnectionError, message)
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(message string) *FileSystemError {
	return New(TimeoutError, message)
}

// NewStorageError creates a new storage error
func NewStorageError(message string) *FileSystemError {
	return New(StorageError, message)
}

// NewFileNotFoundError creates a new file not found error
func NewFileNotFoundError(filename string) *FileSystemError {
	return New(FileNotFoundError, fmt.Sprintf("file not found: %s", filename))
}

// NewCorruptionError creates a new corruption error
func NewCorruptionError(message string) *FileSystemError {
	return New(CorruptionError, message)
}

// NewQuotaExceededError creates a new quota exceeded error
func NewQuotaExceededError(message string) *FileSystemError {
	return New(QuotaExceededError, message)
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(message string) *FileSystemError {
	return New(AuthenticationError, message)
}

// NewAuthorizationError creates a new authorization error
func NewAuthorizationError(message string) *FileSystemError {
	return New(AuthorizationError, message)
}

// NewEncryptionError creates a new encryption error
func NewEncryptionError(message string) *FileSystemError {
	return New(EncryptionError, message)
}

// NewConfigError creates a new configuration error
func NewConfigError(message string) *FileSystemError {
	return New(ConfigError, message)
}

// NewValidationError creates a new validation error
func NewValidationError(message string) *FileSystemError {
	return New(ValidationError, message)
}

// NewInternalError creates a new internal error
func NewInternalError(message string) *FileSystemError {
	return New(InternalError, message)
}

// NewInvalidInputError creates a new invalid input error
func NewInvalidInputError(message string) *FileSystemError {
	return New(InvalidInputError, message)
}
