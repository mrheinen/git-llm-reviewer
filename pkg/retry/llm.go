package retry

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/llm"
)

// Common retryable errors for LLM API calls
var (
	// Network errors
	ErrTimeout        = errors.New("timeout")
	ErrConnectionReset = errors.New("connection reset")
	
	// Rate limiting errors
	ErrRateLimit      = errors.New("rate limit")
	ErrTooManyRequests = errors.New("too many requests")
	
	// Server errors
	ErrServerError    = errors.New("server error")
	ErrInternalServerError = errors.New("internal server error")
	ErrServiceUnavailable = errors.New("service unavailable")
)

// DefaultLLMRetryableErrors returns a list of common retryable errors for LLM API calls
func DefaultLLMRetryableErrors() []error {
	return []error{
		ErrTimeout,
		ErrConnectionReset,
		ErrRateLimit,
		ErrTooManyRequests,
		ErrServerError,
		ErrInternalServerError,
		ErrServiceUnavailable,
	}
}

// IsLLMErrorRetryable checks if an LLM error is retryable
func IsLLMErrorRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for context deadline exceeded
	if errors.Is(err, context.DeadlineExceeded) {
		return true
	}

	// Check for specific error types
	if errors.Is(err, llm.ErrTimeout) {
		return true
	}

	// Check for HTTP status codes in error message
	errMsg := err.Error()
	if strings.Contains(errMsg, "429") || // Too Many Requests
		strings.Contains(errMsg, "500") || // Internal Server Error
		strings.Contains(errMsg, "502") || // Bad Gateway
		strings.Contains(errMsg, "503") || // Service Unavailable
		strings.Contains(errMsg, "504") {  // Gateway Timeout
		return true
	}

	// Check for common error messages
	for _, retryableErr := range DefaultLLMRetryableErrors() {
		if strings.Contains(strings.ToLower(errMsg), strings.ToLower(retryableErr.Error())) {
			return true
		}
	}

	// Check for rate limit errors specifically
	if strings.Contains(strings.ToLower(errMsg), "rate limit") ||
	   strings.Contains(strings.ToLower(errMsg), "rate_limit") ||
	   strings.Contains(strings.ToLower(errMsg), "too many requests") {
		return true
	}

	return false
}

// LLMAPIFunc is a function that makes an API call to an LLM provider
type LLMAPIFunc func() (*http.Response, error)

// DoLLMRequest executes an LLM API request with retry logic
func DoLLMRequest(fn LLMAPIFunc, cfg *config.Config) (*http.Response, error) {
	// Create retry options from config
	opts := FromConfig(cfg)
	
	// Set custom IsRetryable function for LLM errors
	opts.IsRetryableFunc = IsLLMErrorRetryable
	
	// Set custom logger if not already set
	if opts.Logger == nil {
		opts.Logger = func(format string, args ...interface{}) {
			fmt.Printf("[LLM Retry] "+format+"\n", args...)
		}
	}
	
	// Execute with retry
	result, err := Do(func() (interface{}, error) {
		resp, err := fn()
		
		// Check for rate limit response (status code 429)
		if resp != nil && resp.StatusCode == http.StatusTooManyRequests {
			return nil, errors.New("rate limit exceeded")
		}
		
		return resp, err
	}, opts)
	
	if err != nil {
		return nil, err
	}
	
	// Convert result to http.Response
	resp, ok := result.(*http.Response)
	if !ok {
		return nil, fmt.Errorf("unexpected result type: %T", result)
	}
	
	return resp, nil
}
