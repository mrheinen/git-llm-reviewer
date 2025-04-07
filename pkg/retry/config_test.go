package retry

import (
	"testing"
	"time"

	"github.com/niels/git-llm-review/pkg/config"
)

// TestFromConfig tests the conversion from application config to retry options
func TestFromConfig(t *testing.T) {
	// Test with retry enabled
	cfg := &config.Config{
		Retry: config.RetryConfig{
			Enabled:       true,
			MaxRetries:    5,
			InitialDelay:  200,
			MaxDelay:      10000,
			BackoffFactor: 3.0,
			JitterFactor:  0.3,
			RetryableErrors: []string{
				"timeout",
				"connection reset",
			},
		},
	}

	opts := FromConfig(cfg)

	// Verify options
	if opts.MaxRetries != 5 {
		t.Errorf("Expected MaxRetries=5, got: %d", opts.MaxRetries)
	}
	if opts.InitialDelay != 200*time.Millisecond {
		t.Errorf("Expected InitialDelay=200ms, got: %v", opts.InitialDelay)
	}
	if opts.MaxDelay != 10000*time.Millisecond {
		t.Errorf("Expected MaxDelay=10000ms, got: %v", opts.MaxDelay)
	}
	if opts.BackoffFactor != 3.0 {
		t.Errorf("Expected BackoffFactor=3.0, got: %v", opts.BackoffFactor)
	}
	if opts.JitterFactor != 0.3 {
		t.Errorf("Expected JitterFactor=0.3, got: %v", opts.JitterFactor)
	}
	if len(opts.RetryableErrors) != 2 {
		t.Errorf("Expected 2 retryable errors, got: %d", len(opts.RetryableErrors))
	}

	// Test with retry disabled
	disabledCfg := &config.Config{
		Retry: config.RetryConfig{
			Enabled: false,
		},
	}

	disabledOpts := FromConfig(disabledCfg)

	// Verify options
	if disabledOpts.MaxRetries != 0 {
		t.Errorf("Expected MaxRetries=0 when disabled, got: %d", disabledOpts.MaxRetries)
	}
}
