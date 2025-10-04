package retry

import (
	"context"
	"time"

	"github.com/anthdm/foreverstore/errors"
	"github.com/anthdm/foreverstore/logger"
)

// RetryConfig holds configuration for retry operations
type RetryConfig struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
	Jitter       bool
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
		Jitter:       true,
	}
}

// RetryableFunc is a function that can be retried
type RetryableFunc func() error

// Do executes a function with retry logic
func Do(ctx context.Context, config RetryConfig, fn RetryableFunc) error {
	var lastErr error
	delay := config.InitialDelay

	for attempt := 1; attempt <= config.MaxAttempts; attempt++ {
		// Check if context is cancelled
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Execute the function
		err := fn()
		if err == nil {
			if attempt > 1 {
				logger.Info("Operation succeeded after %d attempts", attempt)
			}
			return nil
		}

		lastErr = err
		
		// Check if the error is retryable
		if !errors.IsRetryable(err) {
			logger.Debug("Error is not retryable: %v", err)
			return err
		}

		// Don't sleep on the last attempt
		if attempt == config.MaxAttempts {
			break
		}

		logger.Warn("Attempt %d/%d failed: %v, retrying in %v", 
			attempt, config.MaxAttempts, err, delay)

		// Sleep with context cancellation support
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(delay):
		}

		// Calculate next delay with exponential backoff
		delay = time.Duration(float64(delay) * config.Multiplier)
		if delay > config.MaxDelay {
			delay = config.MaxDelay
		}

		// Add jitter if enabled
		if config.Jitter {
			jitter := time.Duration(float64(delay) * 0.1)
			delay += time.Duration(float64(jitter) * (2*time.Now().UnixNano()%2 - 1))
		}
	}

	logger.Error("Operation failed after %d attempts: %v", config.MaxAttempts, lastErr)
	return lastErr
}

// DoWithTimeout executes a function with retry logic and a timeout
func DoWithTimeout(timeout time.Duration, config RetryConfig, fn RetryableFunc) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return Do(ctx, config, fn)
}

// DoSimple executes a function with default retry configuration
func DoSimple(fn RetryableFunc) error {
	return DoWithTimeout(30*time.Second, DefaultRetryConfig(), fn)
}
