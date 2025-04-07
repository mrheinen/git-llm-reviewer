package llm

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestProviderInterface verifies that a mock provider can implement the Provider interface
func TestProviderInterface(t *testing.T) {
	// Create a mock provider that implements the Provider interface
	mockProvider := &MockProvider{
		name:        "MockLLM",
		apiKey:      "mock-api-key",
		shouldError: false,
	}

	// Verify that the mock provider implements the Provider interface
	var _ Provider = mockProvider // Type assertion to ensure MockProvider implements Provider interface

	// Test the Name method
	name := mockProvider.Name()
	if name != "MockLLM" {
		t.Errorf("Expected provider name 'MockLLM', got '%s'", name)
	}

	// Test the ValidateConfig method
	err := mockProvider.ValidateConfig()
	if err != nil {
		t.Errorf("Expected no error from ValidateConfig, got: %v", err)
	}

	// Test the ValidateConfig method with an error
	mockProvider.shouldError = true
	err = mockProvider.ValidateConfig()
	if err == nil {
		t.Error("Expected error from ValidateConfig when shouldError is true")
	}
	mockProvider.shouldError = false

	// Test the ReviewCode method
	ctx := context.Background()
	request := &ReviewRequest{
		FilePath:    "test.go",
		FileContent: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
		FileDiff:    "diff --git a/test.go b/test.go\nnew file mode 100644\nindex 0000000..1234567\n--- /dev/null\n+++ b/test.go\n@@ -0,0 +1,5 @@\n+package main\n+\n+func main() {\n+\tprintln(\"Hello, World!\")\n+}\n",
		Options: ReviewOptions{
			MaxTokens:  1000,
			Temperature: 0.7,
			Timeout:    30 * time.Second,
		},
	}

	response, err := mockProvider.ReviewCode(ctx, request)
	if err != nil {
		t.Errorf("Expected no error from ReviewCode, got: %v", err)
	}

	if response.Review == "" {
		t.Error("Expected non-empty review in response")
	}

	// Test the ReviewCode method with an error
	mockProvider.shouldError = true
	_, err = mockProvider.ReviewCode(ctx, request)
	if err == nil {
		t.Error("Expected error from ReviewCode when shouldError is true")
	}
	if !errors.Is(err, ErrProviderFailure) {
		t.Errorf("Expected ErrProviderFailure, got: %v", err)
	}
}

// TestReviewRequest verifies that the ReviewRequest structure works as expected
func TestReviewRequest(t *testing.T) {
	// Create a review request
	request := &ReviewRequest{
		FilePath:    "test.go",
		FileContent: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
		FileDiff:    "diff --git a/test.go b/test.go\nnew file mode 100644\nindex 0000000..1234567\n--- /dev/null\n+++ b/test.go\n@@ -0,0 +1,5 @@\n+package main\n+\n+func main() {\n+\tprintln(\"Hello, World!\")\n+}\n",
		Options: ReviewOptions{
			MaxTokens:  1000,
			Temperature: 0.7,
			Timeout:    30 * time.Second,
		},
	}

	// Verify the request fields
	if request.FilePath != "test.go" {
		t.Errorf("Expected FilePath 'test.go', got '%s'", request.FilePath)
	}

	if request.Options.MaxTokens != 1000 {
		t.Errorf("Expected MaxTokens 1000, got %d", request.Options.MaxTokens)
	}

	// Test the Validate method
	err := request.Validate()
	if err != nil {
		t.Errorf("Expected no error from Validate, got: %v", err)
	}

	// Test validation with empty file path
	invalidRequest := &ReviewRequest{
		FilePath:    "",
		FileContent: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
		FileDiff:    "diff --git a/test.go b/test.go\nnew file mode 100644\nindex 0000000..1234567\n--- /dev/null\n+++ b/test.go\n@@ -0,0 +1,5 @@\n+package main\n+\n+func main() {\n+\tprintln(\"Hello, World!\")\n+}\n",
	}

	err = invalidRequest.Validate()
	if err == nil {
		t.Error("Expected error from Validate with empty file path")
	}
	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("Expected ErrInvalidRequest, got: %v", err)
	}
}

// TestReviewResponse verifies that the ReviewResponse structure works as expected
func TestReviewResponse(t *testing.T) {
	// Create a review response
	response := &ReviewResponse{
		Review:     "This code looks good, but you should use fmt.Println instead of println.",
		Confidence: 0.95,
		Metadata: map[string]interface{}{
			"model":      "gpt-4",
			"token_count": 150,
		},
	}

	// Verify the response fields
	if response.Review != "This code looks good, but you should use fmt.Println instead of println." {
		t.Errorf("Expected specific review text, got '%s'", response.Review)
	}

	if response.Confidence != 0.95 {
		t.Errorf("Expected Confidence 0.95, got %f", response.Confidence)
	}

	// Test the IsEmpty method
	if response.IsEmpty() {
		t.Error("Expected IsEmpty to return false for non-empty response")
	}

	// Test IsEmpty with an empty response
	emptyResponse := &ReviewResponse{}
	if !emptyResponse.IsEmpty() {
		t.Error("Expected IsEmpty to return true for empty response")
	}
}

// TestErrorTypes verifies that the error types are appropriate and useful
func TestErrorTypes(t *testing.T) {
	// Test ErrInvalidRequest
	err := NewInvalidRequestError("missing file path")
	if !errors.Is(err, ErrInvalidRequest) {
		t.Errorf("Expected error to be ErrInvalidRequest, got: %v", err)
	}
	if err.Error() != "invalid request: missing file path" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}

	// Test ErrProviderFailure
	err = NewProviderError("API request failed", nil)
	if !errors.Is(err, ErrProviderFailure) {
		t.Errorf("Expected error to be ErrProviderFailure, got: %v", err)
	}
	if err.Error() != "provider error: API request failed" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}

	// Test ErrAuthenticationFailure
	err = NewAuthenticationError("invalid API key")
	if !errors.Is(err, ErrAuthenticationFailure) {
		t.Errorf("Expected error to be ErrAuthenticationFailure, got: %v", err)
	}
	if err.Error() != "authentication error: invalid API key" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}

	// Test ErrTimeout
	err = NewTimeoutError("request timed out after 30s")
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("Expected error to be ErrTimeout, got: %v", err)
	}
	if err.Error() != "timeout error: request timed out after 30s" {
		t.Errorf("Expected specific error message, got: %s", err.Error())
	}

	// Test wrapping errors
	originalErr := errors.New("original error")
	wrappedErr := NewProviderError("API request failed", originalErr)
	if !errors.Is(wrappedErr, ErrProviderFailure) {
		t.Errorf("Expected error to be ErrProviderFailure, got: %v", wrappedErr)
	}
	if !errors.Is(wrappedErr, originalErr) {
		t.Errorf("Expected wrapped error to contain original error")
	}
}

// MockProvider is a mock implementation of the Provider interface for testing
type MockProvider struct {
	name        string
	apiKey      string
	shouldError bool
}

// Name returns the name of the provider
func (m *MockProvider) Name() string {
	return m.name
}

// ValidateConfig validates the provider configuration
func (m *MockProvider) ValidateConfig() error {
	if m.shouldError {
		return NewProviderError("mock validation error", nil)
	}
	return nil
}

// ReviewCode performs a code review using the mock provider
func (m *MockProvider) ReviewCode(ctx context.Context, request *ReviewRequest) (*ReviewResponse, error) {
	if m.shouldError {
		return nil, NewProviderError("mock review error", nil)
	}

	// Simulate a successful review
	return &ReviewResponse{
		Review:     "This is a mock review of the code. It looks good!",
		Confidence: 0.9,
		Metadata: map[string]interface{}{
			"provider": m.name,
			"mock":     true,
		},
	}, nil
}

// GetCompletion sends a prompt to the mock provider and returns a completion
func (p *MockProvider) GetCompletion(prompt string) (string, error) {
	if p.shouldError {
		return "", NewProviderError("mock provider error", errors.New("mock error"))
	}
	
	// Return a mock completion
	return "This is a mock completion response for: " + prompt, nil
}

// SetAPIKey sets the API key for the provider
func (m *MockProvider) SetAPIKey(apiKey string) {
	m.apiKey = apiKey
}
