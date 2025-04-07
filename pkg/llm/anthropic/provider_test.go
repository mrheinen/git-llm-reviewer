package anthropic

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/llm"
)

// TestNewProvider tests the creation of a new Anthropic provider
func TestNewProvider(t *testing.T) {
	// Test case 1: Valid configuration
	cfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "anthropic",
			APIURL:   "https://api.anthropic.com",
			APIKey:   "test-api-key",
			Model:    "claude-3-opus-20240229",
			Timeout:  300,
		},
	}

	provider, err := NewProvider(cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if provider == nil {
		t.Error("Expected provider to be created")
	}

	if provider.Name() != "Anthropic" {
		t.Errorf("Expected provider name 'Anthropic', got: %s", provider.Name())
	}

	// Test case 2: Missing API key
	invalidCfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "anthropic",
			APIURL:   "https://api.anthropic.com",
			APIKey:   "",
			Model:    "claude-3-opus-20240229",
			Timeout:  300,
		},
	}

	_, err = NewProvider(invalidCfg)
	if err == nil {
		t.Error("Expected error for missing API key")
	}
	if !errors.Is(err, llm.ErrAuthenticationFailure) {
		t.Errorf("Expected authentication error, got: %v", err)
	}

	// Test case 3: Invalid model
	invalidModelCfg := &config.Config{
		LLM: config.LLMConfig{
			Provider: "anthropic",
			APIURL:   "https://api.anthropic.com",
			APIKey:   "test-api-key",
			Model:    "",
			Timeout:  300,
		},
	}

	_, err = NewProvider(invalidModelCfg)
	if err == nil {
		t.Error("Expected error for invalid model")
	}
	if !errors.Is(err, llm.ErrInvalidRequest) {
		t.Errorf("Expected invalid request error, got: %v", err)
	}
}

// TestValidateConfig tests the validation of the provider configuration
func TestValidateConfig(t *testing.T) {
	// Create a provider with valid configuration
	provider := &Provider{
		apiKey: "test-api-key",
		model:  "claude-3-opus-20240229",
	}

	// Test valid configuration
	err := provider.ValidateConfig()
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Test invalid API key
	provider.apiKey = ""
	err = provider.ValidateConfig()
	if err == nil {
		t.Error("Expected error for missing API key")
	}
	if !errors.Is(err, llm.ErrAuthenticationFailure) {
		t.Errorf("Expected authentication error, got: %v", err)
	}

	// Test invalid model
	provider.apiKey = "test-api-key"
	provider.model = ""
	err = provider.ValidateConfig()
	if err == nil {
		t.Error("Expected error for invalid model")
	}
	if !errors.Is(err, llm.ErrInvalidRequest) {
		t.Errorf("Expected invalid request error, got: %v", err)
	}
}

// MockAnthropicServer creates a mock server for testing Anthropic API calls
func MockAnthropicServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check request method
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got: %s", r.Method)
		}

		// Check authentication
		authHeader := r.Header.Get("X-Api-Key")
		if authHeader != "test-api-key" {
			t.Errorf("Expected X-Api-Key header 'test-api-key', got: %s", authHeader)
		}

		// Check Anthropic-Version header
		versionHeader := r.Header.Get("Anthropic-Version")
		if versionHeader == "" {
			t.Error("Expected Anthropic-Version header to be set")
		}

		// Parse request body
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			t.Errorf("Error decoding request body: %v", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		// Check that the request contains the expected content
		messages, ok := requestBody["messages"].([]interface{})
		if !ok || len(messages) < 2 {
			t.Errorf("Expected at least 2 messages, got: %v", messages)
		}

		// Check model
		model, ok := requestBody["model"].(string)
		if !ok || model != "claude-3-opus-20240229" {
			t.Errorf("Expected model 'claude-3-opus-20240229', got: %v", model)
		}

		// Simulate successful response
		response := map[string]interface{}{
			"id":      "msg_0123456789",
			"type":    "message",
			"role":    "assistant",
			"model":   "claude-3-opus-20240229",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "This code looks good, but you should use fmt.Println instead of println.",
				},
			},
			"usage": map[string]interface{}{
				"input_tokens":  100,
				"output_tokens": 50,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
}

// TestReviewCode tests the code review functionality
func TestReviewCode(t *testing.T) {
	// Create a mock server to simulate Anthropic API
	server := MockAnthropicServer(t)
	defer server.Close()

	// Create a provider using the mock server
	provider := &Provider{
		apiKey:  "test-api-key",
		model:   "claude-3-opus-20240229",
		baseURL: server.URL,
		client:  &http.Client{},
	}

	// Create a review request
	request := &llm.ReviewRequest{
		FilePath:    "test.go",
		FileContent: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
		FileDiff:    "diff --git a/test.go b/test.go\nnew file mode 100644\nindex 0000000..1234567\n--- /dev/null\n+++ b/test.go\n@@ -0,0 +1,5 @@\n+package main\n+\n+func main() {\n+\tprintln(\"Hello, World!\")\n+}\n",
		Options: llm.ReviewOptions{
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     30 * time.Second,
		},
	}

	// Test successful review
	ctx := context.Background()
	response, err := provider.ReviewCode(ctx, request)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Review != "This code looks good, but you should use fmt.Println instead of println." {
		t.Errorf("Expected specific review text, got: %s", response.Review)
	}

	// Check metadata
	if response.Metadata == nil {
		t.Error("Expected metadata in response")
	} else {
		model, ok := response.Metadata["model"]
		if !ok || model != "claude-3-opus-20240229" {
			t.Errorf("Expected model 'claude-3-opus-20240229' in metadata, got: %v", model)
		}

		if _, hasTokenCount := response.Metadata["token_count"]; !hasTokenCount {
			t.Errorf("Expected token_count in metadata, got none")
		}
	}
}

// MockTimeoutServer creates a mock server that delays response to trigger timeout
func MockTimeoutServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Sleep longer than the timeout
		time.Sleep(200 * time.Millisecond)

		// Simulate successful response (this should not be reached due to timeout)
		response := map[string]interface{}{
			"id":      "msg_0123456789",
			"type":    "message",
			"role":    "assistant",
			"model":   "claude-3-opus-20240229",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "This code looks good.",
				},
			},
			"usage": map[string]interface{}{
				"input_tokens":  100,
				"output_tokens": 50,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
	}))
}

// TestReviewCodeTimeout tests the timeout behavior
func TestReviewCodeTimeout(t *testing.T) {
	// Create a mock server that delays response to trigger timeout
	server := MockTimeoutServer(t)
	defer server.Close()

	// Create a provider using the mock server
	provider := &Provider{
		apiKey:  "test-api-key",
		model:   "claude-3-opus-20240229",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 100 * time.Millisecond},
	}

	// Create a review request with a very short timeout
	request := &llm.ReviewRequest{
		FilePath:    "test.go",
		FileContent: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
		FileDiff:    "diff --git a/test.go b/test.go\nnew file mode 100644\nindex 0000000..1234567\n--- /dev/null\n+++ b/test.go\n@@ -0,0 +1,5 @@\n+package main\n+\n+func main() {\n+\tprintln(\"Hello, World!\")\n+}\n",
		Options: llm.ReviewOptions{
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     100 * time.Millisecond, // Very short timeout
		},
	}

	// Test timeout
	ctx := context.Background()
	_, err := provider.ReviewCode(ctx, request)
	if err == nil {
		t.Error("Expected timeout error")
	}
	if !errors.Is(err, llm.ErrTimeout) {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

// MockErrorServer creates a mock server that returns an error
func MockErrorServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate API error
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"type":    "invalid_request_error",
				"message": "Invalid API key",
			},
		})
	}))
}

// TestReviewCodeAPIError tests handling of API errors
func TestReviewCodeAPIError(t *testing.T) {
	// Create a mock server that returns an error
	server := MockErrorServer(t)
	defer server.Close()

	// Create a provider using the mock server
	provider := &Provider{
		apiKey:  "invalid-api-key",
		model:   "claude-3-opus-20240229",
		baseURL: server.URL,
		client:  &http.Client{},
	}

	// Create a review request
	request := &llm.ReviewRequest{
		FilePath:    "test.go",
		FileContent: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
		FileDiff:    "diff --git a/test.go b/test.go\nnew file mode 100644\nindex 0000000..1234567\n--- /dev/null\n+++ b/test.go\n@@ -0,0 +1,5 @@\n+package main\n+\n+func main() {\n+\tprintln(\"Hello, World!\")\n+}\n",
		Options: llm.ReviewOptions{
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     30 * time.Second,
		},
	}

	// Test API error
	ctx := context.Background()
	_, err := provider.ReviewCode(ctx, request)
	if err == nil {
		t.Error("Expected API error")
	}
	if !errors.Is(err, llm.ErrAuthenticationFailure) {
		t.Errorf("Expected authentication failure error, got: %v", err)
	}
}

// TestReviewCodeWithRetry tests the retry behavior for API calls
func TestReviewCodeWithRetry(t *testing.T) {
	// Use a counter to track retry attempts
	attempts := 0
	maxFailures := 2

	// Create a mock server that fails twice before succeeding
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Consume the request body to avoid "ContentLength with Body length 0" errors
		io.ReadAll(r.Body)
		r.Body.Close()

		attempts++
		if attempts <= maxFailures {
			// Return a rate limit error for the first few attempts
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"type":    "rate_limit_error",
					"message": "Rate limit exceeded, please try again later",
				},
			})
			return
		}

		// Return success on the third attempt
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "msg_0123456789",
			"type":    "message",
			"role":    "assistant",
			"model":   "claude-3-opus-20240229",
			"content": []map[string]interface{}{
				{
					"type": "text",
					"text": "This code looks good, but you should use fmt.Println instead of println.",
				},
			},
			"usage": map[string]interface{}{
				"input_tokens":  100,
				"output_tokens": 50,
			},
		})
	}))
	defer server.Close()

	// Create a config with retry enabled
	cfg := &config.Config{
		Retry: config.RetryConfig{
			Enabled:       true,
			MaxRetries:    3,
			InitialDelay:  50,
			MaxDelay:      1000,
			BackoffFactor: 2.0,
			JitterFactor:  0.1,
			RetryableErrors: []string{
				"timeout",
				"rate limit",
				"too many requests",
				"server error",
			},
		},
	}

	// Create a provider using the mock server
	provider := &Provider{
		apiKey:  "test-api-key",
		model:   "claude-3-opus-20240229",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 5 * time.Second},
		config:  cfg,
	}

	// Create a review request
	request := &llm.ReviewRequest{
		FilePath:    "test.go",
		FileContent: "package main\n\nfunc main() {\n\tprintln(\"Hello, World!\")\n}\n",
		FileDiff:    "diff --git a/test.go b/test.go\nnew file mode 100644\nindex 0000000..1234567\n--- /dev/null\n+++ b/test.go\n@@ -0,0 +1,5 @@\n+package main\n+\n+func main() {\n+\tprintln(\"Hello, World!\")\n+}\n",
		Options: llm.ReviewOptions{
			MaxTokens:   1000,
			Temperature: 0.7,
			Timeout:     10 * time.Second,
		},
	}

	// Test successful review with retry
	ctx := context.Background()
	response, err := provider.ReviewCode(ctx, request)
	if err != nil {
		t.Errorf("Expected no error with retry, got: %v", err)
	}

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Review != "This code looks good, but you should use fmt.Println instead of println." {
		t.Errorf("Expected specific review text, got: %s", response.Review)
	}

	// Verify that the correct number of attempts were made
	if attempts != maxFailures+1 {
		t.Errorf("Expected %d attempts, got: %d", maxFailures+1, attempts)
	}
}
