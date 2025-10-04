package retry

import (
	"context"
	"testing"
	"time"

	"github.com/anthdm/foreverstore/errors"
)

func TestRetrySuccess(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 3 {
			return errors.NewNetworkError("temporary failure")
		}
		return nil
	}

	config := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	ctx := context.Background()
	err := Do(ctx, config, fn)

	if err != nil {
		t.Errorf("Expected success after retries, got error: %v", err)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryFailure(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		return errors.NewNetworkError("persistent failure")
	}

	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	ctx := context.Background()
	err := Do(ctx, config, fn)

	if err == nil {
		t.Error("Expected failure after max attempts")
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetryNonRetryableError(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		return errors.NewValidationError("non-retryable error")
	}

	config := DefaultRetryConfig()
	ctx := context.Background()
	err := Do(ctx, config, fn)

	if err == nil {
		t.Error("Expected error to be returned")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got %d", attempts)
	}
}

func TestRetryContextCancellation(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		return errors.NewNetworkError("network error")
	}

	config := RetryConfig{
		MaxAttempts:  10,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
		Multiplier:   2.0,
		Jitter:       false,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := Do(ctx, config, fn)

	if err != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded, got: %v", err)
	}

	// Should have attempted at least once
	if attempts < 1 {
		t.Errorf("Expected at least 1 attempt, got %d", attempts)
	}
}

func TestRetryWithTimeout(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 2 {
			return errors.NewNetworkError("temporary failure")
		}
		return nil
	}

	config := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 1 * time.Millisecond,
		MaxDelay:     10 * time.Millisecond,
		Multiplier:   2.0,
		Jitter:       false,
	}

	err := DoWithTimeout(100*time.Millisecond, config, fn)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestRetrySimple(t *testing.T) {
	attempts := 0
	fn := func() error {
		attempts++
		if attempts < 2 {
			return errors.NewNetworkError("temporary failure")
		}
		return nil
	}

	err := DoSimple(fn)

	if err != nil {
		t.Errorf("Expected success, got error: %v", err)
	}

	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxAttempts != 3 {
		t.Errorf("Expected 3 max attempts, got %d", config.MaxAttempts)
	}

	if config.InitialDelay != 100*time.Millisecond {
		t.Errorf("Expected 100ms initial delay, got %v", config.InitialDelay)
	}

	if config.MaxDelay != 5*time.Second {
		t.Errorf("Expected 5s max delay, got %v", config.MaxDelay)
	}

	if config.Multiplier != 2.0 {
		t.Errorf("Expected 2.0 multiplier, got %f", config.Multiplier)
	}

	if !config.Jitter {
		t.Error("Expected jitter to be enabled")
	}
}
