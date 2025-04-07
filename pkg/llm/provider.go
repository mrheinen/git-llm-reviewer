package llm

import (
	"context"
	"errors"
	"fmt"
	"time"
	
	"github.com/niels/git-llm-review/pkg/extractor"
)

// Common error types for LLM operations
var (
	ErrInvalidRequest       = errors.New("invalid request")
	ErrProviderFailure      = errors.New("provider error")
	ErrAuthenticationFailure = errors.New("authentication error")
	ErrTimeout              = errors.New("timeout error")
	ErrUnsupportedProvider  = errors.New("unsupported provider")
	ErrConfigurationError   = errors.New("configuration error")
)

// Provider defines the interface that all LLM providers must implement
type Provider interface {
	// Name returns the name of the provider
	Name() string
	
	// ValidateConfig validates the provider configuration
	ValidateConfig() error
	
	// ReviewCode performs a code review using the provider
	ReviewCode(ctx context.Context, request *ReviewRequest) (*ReviewResponse, error)
	
	// GetCompletion sends a prompt to the LLM and returns the completion
	GetCompletion(prompt string) (string, error)
}

// ReviewRequest represents a request to review code
type ReviewRequest struct {
	// FilePath is the path to the file being reviewed
	FilePath string
	
	// FileContent is the full content of the file
	FileContent string
	
	// FileDiff is the diff of the changes made to the file
	FileDiff string
	
	// Options contains additional options for the review
	Options ReviewOptions
	
	// Extractor provides access to code extraction functionality
	Extractor *extractor.CodeExtractor
}

// ReviewOptions contains options for the code review
type ReviewOptions struct {
	// MaxTokens is the maximum number of tokens to generate
	MaxTokens int
	
	// Temperature controls the randomness of the output (0.0-1.0)
	Temperature float64
	
	// Timeout is the maximum time to wait for a response
	Timeout time.Duration
	
	// AdditionalInstructions provides extra guidance to the LLM
	AdditionalInstructions string
	
	// IncludeExplanations indicates whether to include explanations in the review
	IncludeExplanations bool
}

// ReviewResponse represents the response from a code review
type ReviewResponse struct {
	// Review is the text of the code review
	Review string
	
	// Confidence is a measure of the provider's confidence in the review (0.0-1.0)
	Confidence float64
	
	// Metadata contains additional information about the review
	Metadata map[string]interface{}
}

// IsEmpty returns true if the response is empty
func (r *ReviewResponse) IsEmpty() bool {
	return r.Review == ""
}

// Validate validates the review request
func (r *ReviewRequest) Validate() error {
	if r.FilePath == "" {
		return NewInvalidRequestError("file path is required")
	}
	
	// FileContent can be empty for deleted files
	
	// FileDiff can be empty for new files or when only reviewing content
	
	// Validate options
	if r.Options.MaxTokens < 0 {
		return NewInvalidRequestError("max tokens must be non-negative")
	}
	
	if r.Options.Temperature < 0 || r.Options.Temperature > 1.0 {
		return NewInvalidRequestError("temperature must be between 0.0 and 1.0")
	}
	
	return nil
}

// ProviderError represents an error from an LLM provider
type ProviderError struct {
	msg string
	err error
}

// Error returns the error message
func (e *ProviderError) Error() string {
	return e.msg
}

// Unwrap returns the underlying error
func (e *ProviderError) Unwrap() error {
	return e.err
}

// NewProviderError creates a new provider error
func NewProviderError(msg string, err error) error {
	return &ProviderError{
		msg: fmt.Sprintf("provider error: %s", msg),
		err: errors.Join(ErrProviderFailure, err),
	}
}

// InvalidRequestError represents an error due to an invalid request
type InvalidRequestError struct {
	msg string
	err error
}

// Error returns the error message
func (e *InvalidRequestError) Error() string {
	return e.msg
}

// Unwrap returns the underlying error
func (e *InvalidRequestError) Unwrap() error {
	return e.err
}

// NewInvalidRequestError creates a new invalid request error
func NewInvalidRequestError(msg string) error {
	return &InvalidRequestError{
		msg: fmt.Sprintf("invalid request: %s", msg),
		err: ErrInvalidRequest,
	}
}

// AuthenticationError represents an authentication failure
type AuthenticationError struct {
	msg string
	err error
}

// Error returns the error message
func (e *AuthenticationError) Error() string {
	return e.msg
}

// Unwrap returns the underlying error
func (e *AuthenticationError) Unwrap() error {
	return e.err
}

// NewAuthenticationError creates a new authentication error
func NewAuthenticationError(msg string) error {
	return &AuthenticationError{
		msg: fmt.Sprintf("authentication error: %s", msg),
		err: ErrAuthenticationFailure,
	}
}

// TimeoutError represents a timeout error
type TimeoutError struct {
	msg string
	err error
}

// Error returns the error message
func (e *TimeoutError) Error() string {
	return e.msg
}

// Unwrap returns the underlying error
func (e *TimeoutError) Unwrap() error {
	return e.err
}

// NewTimeoutError creates a new timeout error
func NewTimeoutError(msg string) error {
	return &TimeoutError{
		msg: fmt.Sprintf("timeout error: %s", msg),
		err: ErrTimeout,
	}
}

// ProviderFactory is a function type that creates a new provider
type ProviderFactory func(config map[string]interface{}) (Provider, error)

// Global provider registry
var globalRegistry = NewProviderRegistry()

// RegisterProviderFactory registers a provider factory with the global registry
func RegisterProviderFactory(name string, factory ProviderFactory) {
	globalRegistry.Register(name, factory)
}

// CreateProvider creates a new provider with the given name and configuration
// using the global registry
func CreateProvider(name string, config map[string]interface{}) (Provider, error) {
	return globalRegistry.Create(name, config)
}

// GetAvailableProviders returns a list of available provider names
// from the global registry
func GetAvailableProviders() []string {
	return globalRegistry.GetAvailableProviders()
}

// ProviderRegistry maintains a registry of available providers
type ProviderRegistry struct {
	factories map[string]ProviderFactory
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		factories: make(map[string]ProviderFactory),
	}
}

// Register registers a provider factory with the registry
func (r *ProviderRegistry) Register(name string, factory ProviderFactory) {
	r.factories[name] = factory
}

// Create creates a new provider with the given name and configuration
func (r *ProviderRegistry) Create(name string, config map[string]interface{}) (Provider, error) {
	factory, ok := r.factories[name]
	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedProvider, name)
	}
	
	return factory(config)
}

// GetAvailableProviders returns a list of available provider names
func (r *ProviderRegistry) GetAvailableProviders() []string {
	providers := make([]string, 0, len(r.factories))
	for name := range r.factories {
		providers = append(providers, name)
	}
	return providers
}
