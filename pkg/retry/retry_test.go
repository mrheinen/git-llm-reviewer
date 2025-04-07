package retry

import (
	"errors"
	"testing"
	"time"
)

// TestRetrySuccess tests that a function is retried until it succeeds
func TestRetrySuccess(t *testing.T) {
	// Create a function that fails a few times before succeeding
	attempts := 0
	maxFailures := 2
	testFunc := func() (interface{}, error) {
		attempts++
		if attempts <= maxFailures {
			return nil, errors.New("temporary error")
		}
		return "success", nil
	}

	// Configure retry options
	opts := Options{
		MaxRetries:      5,
		InitialDelay:    1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		RetryableErrors: []error{errors.New("temporary error")},
	}

	// Execute with retry
	result, err := Do(testFunc, opts)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected result 'success', got: %v", result)
	}

	if attempts != maxFailures+1 {
		t.Errorf("Expected %d attempts, got: %d", maxFailures+1, attempts)
	}
}

// TestRetryMaxAttemptsExceeded tests that retry gives up after max attempts
func TestRetryMaxAttemptsExceeded(t *testing.T) {
	// Create a function that always fails
	attempts := 0
	testFunc := func() (interface{}, error) {
		attempts++
		return nil, errors.New("temporary error")
	}

	// Configure retry options
	opts := Options{
		MaxRetries:      3,
		InitialDelay:    1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		RetryableErrors: []error{errors.New("temporary error")},
	}

	// Execute with retry
	_, err := Do(testFunc, opts)

	// Verify results
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if attempts != opts.MaxRetries+1 {
		t.Errorf("Expected %d attempts, got: %d", opts.MaxRetries+1, attempts)
	}
}

// TestRetryNonRetryableError tests that retry stops immediately on non-retryable errors
func TestRetryNonRetryableError(t *testing.T) {
	// Create a function that fails with a non-retryable error
	attempts := 0
	testFunc := func() (interface{}, error) {
		attempts++
		return nil, errors.New("non-retryable error")
	}

	// Configure retry options
	opts := Options{
		MaxRetries:      3,
		InitialDelay:    1 * time.Millisecond,
		MaxDelay:        10 * time.Millisecond,
		BackoffFactor:   2.0,
		JitterFactor:    0.1,
		RetryableErrors: []error{errors.New("temporary error")},
	}

	// Execute with retry
	_, err := Do(testFunc, opts)

	// Verify results
	if err == nil {
		t.Error("Expected error, got nil")
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got: %d", attempts)
	}
}

// TestRetryBackoff tests that retry uses exponential backoff
func TestRetryBackoff(t *testing.T) {
	// Create a function that always fails
	attempts := 0
	delays := make([]time.Duration, 0)
	lastAttemptTime := time.Now()

	testFunc := func() (interface{}, error) {
		currentTime := time.Now()
		if attempts > 0 {
			delays = append(delays, currentTime.Sub(lastAttemptTime))
		}
		lastAttemptTime = currentTime
		attempts++
		return nil, errors.New("temporary error")
	}

	// Configure retry options with predictable jitter
	opts := Options{
		MaxRetries:      3,
		InitialDelay:    10 * time.Millisecond,
		MaxDelay:        100 * time.Millisecond,
		BackoffFactor:   2.0,
		JitterFactor:    0.0, // No jitter for predictable testing
		RetryableErrors: []error{errors.New("temporary error")},
	}

	// Execute with retry
	_, _ = Do(testFunc, opts)

	// Verify results
	if len(delays) != 3 {
		t.Errorf("Expected 3 delays, got: %d", len(delays))
	}

	// Check that delays are increasing exponentially
	if len(delays) >= 3 {
		if delays[1] < delays[0] {
			t.Errorf("Expected delay to increase, but got: %v, %v", delays[0], delays[1])
		}
		if delays[2] < delays[1] {
			t.Errorf("Expected delay to increase, but got: %v, %v", delays[1], delays[2])
		}

		// Check approximate exponential growth (with some tolerance for timing variations)
		expectedRatio := opts.BackoffFactor
		actualRatio1 := float64(delays[1]) / float64(delays[0])
		actualRatio2 := float64(delays[2]) / float64(delays[1])

		tolerance := 0.5 // Allow for some timing variation
		if actualRatio1 < expectedRatio-tolerance || actualRatio1 > expectedRatio+tolerance {
			t.Errorf("Expected delay ratio around %.1f, got: %.1f", expectedRatio, actualRatio1)
		}
		if actualRatio2 < expectedRatio-tolerance || actualRatio2 > expectedRatio+tolerance {
			t.Errorf("Expected delay ratio around %.1f, got: %.1f", expectedRatio, actualRatio2)
		}
	}
}

// TestRetryWithJitter tests that retry adds jitter to delays
func TestRetryWithJitter(t *testing.T) {
	// Create a function that always fails
	const numRuns = 10
	baseDelays := make([]time.Duration, numRuns)
	jitterDelays := make([]time.Duration, numRuns)

	// First, measure delays without jitter
	for i := 0; i < numRuns; i++ {
		attempts := 0
		lastAttemptTime := time.Now()

		testFunc := func() (interface{}, error) {
			currentTime := time.Now()
			if attempts > 0 {
				baseDelays[i] = currentTime.Sub(lastAttemptTime)
				return "success", nil
			}
			lastAttemptTime = currentTime
			attempts++
			return nil, errors.New("temporary error")
		}

		opts := Options{
			MaxRetries:      3,
			InitialDelay:    20 * time.Millisecond,
			MaxDelay:        100 * time.Millisecond,
			BackoffFactor:   2.0,
			JitterFactor:    0.0, // No jitter
			RetryableErrors: []error{errors.New("temporary error")},
		}

		_, _ = Do(testFunc, opts)
	}

	// Then, measure delays with jitter
	for i := 0; i < numRuns; i++ {
		attempts := 0
		lastAttemptTime := time.Now()

		testFunc := func() (interface{}, error) {
			currentTime := time.Now()
			if attempts > 0 {
				jitterDelays[i] = currentTime.Sub(lastAttemptTime)
				return "success", nil
			}
			lastAttemptTime = currentTime
			attempts++
			return nil, errors.New("temporary error")
		}

		opts := Options{
			MaxRetries:      3,
			InitialDelay:    20 * time.Millisecond,
			MaxDelay:        100 * time.Millisecond,
			BackoffFactor:   2.0,
			JitterFactor:    0.5, // 50% jitter
			RetryableErrors: []error{errors.New("temporary error")},
		}

		_, _ = Do(testFunc, opts)
	}

	// Check that jitter delays have some variation
	allSame := true
	for i := 1; i < numRuns; i++ {
		if jitterDelays[i] != jitterDelays[0] {
			allSame = false
			break
		}
	}

	if allSame {
		t.Error("Expected jitter to cause delay variation, but all delays were identical")
	}
}

// TestRetryIsRetryable tests the IsRetryable function
func TestRetryIsRetryable(t *testing.T) {
	// Define retryable errors
	retryableErrors := []error{
		errors.New("timeout"),
		errors.New("connection reset"),
	}

	// Test retryable errors
	if !IsRetryable(errors.New("timeout"), retryableErrors) {
		t.Error("Expected 'timeout' to be retryable")
	}
	if !IsRetryable(errors.New("connection reset"), retryableErrors) {
		t.Error("Expected 'connection reset' to be retryable")
	}

	// Test non-retryable error
	if IsRetryable(errors.New("bad request"), retryableErrors) {
		t.Error("Expected 'bad request' to be non-retryable")
	}
}

// TestRetryWithCustomIsRetryableFunc tests retry with a custom IsRetryable function
func TestRetryWithCustomIsRetryableFunc(t *testing.T) {
	// Create a function that fails with different errors
	attempts := 0
	testFunc := func() (interface{}, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("retryable-1")
		}
		if attempts == 2 {
			return nil, errors.New("retryable-2")
		}
		return "success", nil
	}

	// Configure retry options with a custom IsRetryable function
	opts := Options{
		MaxRetries:    3,
		InitialDelay:  1 * time.Millisecond,
		MaxDelay:      10 * time.Millisecond,
		BackoffFactor: 2.0,
		JitterFactor:  0.1,
		IsRetryableFunc: func(err error) bool {
			return err.Error() == "retryable-1" || err.Error() == "retryable-2"
		},
	}

	// Execute with retry
	result, err := Do(testFunc, opts)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result != "success" {
		t.Errorf("Expected result 'success', got: %v", result)
	}

	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got: %d", attempts)
	}
}
