package retry

import (
	"errors"
	"time"

	"github.com/niels/git-llm-review/pkg/config"
)

// FromConfig creates retry options from the application configuration
func FromConfig(cfg *config.Config) Options {
	// If retry is disabled, return options with no retries
	if !cfg.Retry.Enabled {
		return Options{
			MaxRetries: 0,
		}
	}

	// Convert string errors to error objects
	retryableErrors := make([]error, 0, len(cfg.Retry.RetryableErrors))
	for _, errStr := range cfg.Retry.RetryableErrors {
		retryableErrors = append(retryableErrors, errors.New(errStr))
	}

	// Create retry options from configuration
	return Options{
		MaxRetries:      cfg.Retry.MaxRetries,
		InitialDelay:    time.Duration(cfg.Retry.InitialDelay) * time.Millisecond,
		MaxDelay:        time.Duration(cfg.Retry.MaxDelay) * time.Millisecond,
		BackoffFactor:   cfg.Retry.BackoffFactor,
		JitterFactor:    cfg.Retry.JitterFactor,
		RetryableErrors: retryableErrors,
	}
}
