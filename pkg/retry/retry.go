package retry

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// RetryFunc is a function that can be retried
type RetryFunc func() (interface{}, error)

// IsRetryableFunc is a function that determines if an error is retryable
type IsRetryableFunc func(error) bool

// Options configures the retry behavior
type Options struct {
	// MaxRetries is the maximum number of retry attempts (not including the initial attempt)
	MaxRetries int

	// InitialDelay is the delay before the first retry
	InitialDelay time.Duration

	// MaxDelay is the maximum delay between retries
	MaxDelay time.Duration

	// BackoffFactor is the factor by which the delay increases after each retry
	BackoffFactor float64

	// JitterFactor adds randomness to the delay (0.0 = no jitter, 1.0 = 100% jitter)
	JitterFactor float64

	// RetryableErrors is a list of errors that are considered retryable
	RetryableErrors []error

	// IsRetryableFunc is a function that determines if an error is retryable
	// If provided, this takes precedence over RetryableErrors
	IsRetryableFunc IsRetryableFunc

	// Logger is a function that logs retry attempts
	Logger func(format string, args ...interface{})
}

// DefaultOptions returns default retry options
func DefaultOptions() Options {
	return Options{
		MaxRetries:    3,
		InitialDelay:  100 * time.Millisecond,
		MaxDelay:      5 * time.Second,
		BackoffFactor: 2.0,
		JitterFactor:  0.2,
		Logger:        func(format string, args ...interface{}) { fmt.Printf(format+"\n", args...) },
	}
}

// Do executes the given function with retry logic
func Do(fn RetryFunc, opts Options) (interface{}, error) {
	var result interface{}
	var err error
	var delay time.Duration = 0

	// Initialize random source for jitter
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	// Set default logger if not provided
	if opts.Logger == nil {
		opts.Logger = func(format string, args ...interface{}) { fmt.Printf(format+"\n", args...) }
	}

	// Try the function up to MaxRetries+1 times
	for attempt := 0; attempt <= opts.MaxRetries; attempt++ {
		// Execute the function
		result, err = fn()

		// If no error or non-retryable error, return immediately
		if err == nil {
			if attempt > 0 {
				opts.Logger("Retry successful on attempt %d", attempt+1)
			}
			return result, nil
		}

		// Check if the error is retryable
		if !isRetryable(err, opts) {
			opts.Logger("Non-retryable error: %v", err)
			return nil, err
		}

		// If this was the last attempt, return the error
		if attempt == opts.MaxRetries {
			opts.Logger("Max retries exceeded (%d attempts): %v", attempt+1, err)
			return nil, err
		}

		// Calculate delay for the next retry
		if attempt == 0 {
			delay = opts.InitialDelay
		} else {
			// Apply exponential backoff
			delay = time.Duration(float64(delay) * opts.BackoffFactor)
			
			// Cap at MaxDelay
			if delay > opts.MaxDelay {
				delay = opts.MaxDelay
			}
		}

		// Apply jitter
		if opts.JitterFactor > 0 {
			jitter := float64(delay) * opts.JitterFactor
			delay = time.Duration(float64(delay) + (rnd.Float64()*jitter*2 - jitter))
		}

		opts.Logger("Retry attempt %d after %v: %v", attempt+1, delay, err)

		// Wait before the next attempt
		time.Sleep(delay)
	}

	// This should never be reached due to the return in the loop
	return nil, errors.New("unexpected error in retry logic")
}

// IsRetryable checks if an error is retryable based on the provided retryable errors
func IsRetryable(err error, retryableErrors []error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	for _, retryableErr := range retryableErrors {
		if strings.Contains(errMsg, retryableErr.Error()) {
			return true
		}
	}

	return false
}

// isRetryable checks if an error is retryable based on the options
func isRetryable(err error, opts Options) bool {
	// If a custom function is provided, use it
	if opts.IsRetryableFunc != nil {
		return opts.IsRetryableFunc(err)
	}

	// Otherwise, use the list of retryable errors
	return IsRetryable(err, opts.RetryableErrors)
}
