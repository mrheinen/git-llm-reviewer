package retry

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/llm"
)

// TestIsLLMErrorRetryable tests the IsLLMErrorRetryable function
func TestIsLLMErrorRetryable(t *testing.T) {
	// Test retryable errors
	retryableErrors := []error{
		context.DeadlineExceeded,
		llm.NewTimeoutError("request timed out"),
		errors.New("rate limit exceeded"),
		errors.New("too many requests"),
		errors.New("received status code 429"),
		errors.New("received status code 500"),
		errors.New("received status code 502"),
		errors.New("received status code 503"),
		errors.New("received status code 504"),
		errors.New("connection reset by peer"),
		errors.New("server error occurred"),
		errors.New("internal server error"),
		errors.New("service unavailable"),
	}

	for _, err := range retryableErrors {
		if !IsLLMErrorRetryable(err) {
			t.Errorf("Expected error to be retryable: %v", err)
		}
	}

	// Test non-retryable errors
	nonRetryableErrors := []error{
		nil,
		errors.New("invalid request"),
		errors.New("bad request"),
		errors.New("received status code 400"),
		errors.New("received status code 401"),
		errors.New("received status code 403"),
		errors.New("received status code 404"),
		llm.NewInvalidRequestError("invalid parameter"),
		llm.NewAuthenticationError("invalid API key"),
	}

	for _, err := range nonRetryableErrors {
		if IsLLMErrorRetryable(err) {
			t.Errorf("Expected error to be non-retryable: %v", err)
		}
	}
}

// TestDoLLMRequest tests the DoLLMRequest function
func TestDoLLMRequest(t *testing.T) {
	// Create a mock HTTP client that fails a few times before succeeding
	attempts := 0
	maxFailures := 2
	mockLLMAPIFunc := func() (*http.Response, error) {
		attempts++
		if attempts <= maxFailures {
			return nil, errors.New("rate limit exceeded")
		}
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       http.NoBody,
		}, nil
	}

	// Create a configuration with retry enabled
	cfg := &config.Config{
		Retry: config.RetryConfig{
			Enabled:       true,
			MaxRetries:    5,
			InitialDelay:  1,  // 1 millisecond for faster tests
			MaxDelay:      10, // 10 milliseconds for faster tests
			BackoffFactor: 2.0,
			JitterFactor:  0.1,
			RetryableErrors: []string{
				"rate limit",
				"timeout",
			},
		},
	}

	// Execute with retry
	resp, err := DoLLMRequest(mockLLMAPIFunc, cfg)

	// Verify results
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if resp == nil {
		t.Error("Expected response, got nil")
	}

	if attempts != maxFailures+1 {
		t.Errorf("Expected %d attempts, got: %d", maxFailures+1, attempts)
	}

	// Test with retry disabled
	attempts = 0
	disabledCfg := &config.Config{
		Retry: config.RetryConfig{
			Enabled: false,
		},
	}

	// Execute without retry
	resp, err = DoLLMRequest(mockLLMAPIFunc, disabledCfg)

	// Verify results
	if err == nil {
		t.Error("Expected error when retry is disabled, got nil")
	}

	if resp != nil {
		t.Errorf("Expected nil response, got: %v", resp)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt when retry is disabled, got: %d", attempts)
	}

	// Test with non-retryable error
	attempts = 0
	nonRetryableMockFunc := func() (*http.Response, error) {
		attempts++
		return nil, errors.New("bad request")
	}

	// Execute with retry
	resp, err = DoLLMRequest(nonRetryableMockFunc, cfg)

	// Verify results
	if err == nil {
		t.Error("Expected error for non-retryable error, got nil")
	}

	if resp != nil {
		t.Errorf("Expected nil response, got: %v", resp)
	}

	if attempts != 1 {
		t.Errorf("Expected 1 attempt for non-retryable error, got: %d", attempts)
	}
}
