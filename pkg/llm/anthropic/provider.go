package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/extractor"
	"github.com/niels/git-llm-review/pkg/llm"
	"github.com/niels/git-llm-review/pkg/llm/promptlog"
	"github.com/niels/git-llm-review/pkg/logging"
	"github.com/niels/git-llm-review/pkg/prompt"
	"github.com/niels/git-llm-review/pkg/retry"
)

const (
	// Default Anthropic API URL
	defaultAPIURL = "https://api.anthropic.com"
	// Anthropic API version
	apiVersion = "2023-06-01"
)

// Provider implements the llm.Provider interface for Anthropic
type Provider struct {
	apiKey  string
	model   string
	baseURL string
	client  *http.Client
	config  *config.Config
}

// NewProvider creates a new Anthropic provider
func NewProvider(cfg *config.Config) (llm.Provider, error) {
	// Create a new provider
	provider := &Provider{
		apiKey:  cfg.LLM.APIKey,
		model:   cfg.LLM.Model,
		baseURL: cfg.LLM.APIURL,
		client:  &http.Client{},
		config:  cfg,
	}

	// Set default API URL if not specified
	if provider.baseURL == "" {
		provider.baseURL = defaultAPIURL
	}

	// Set timeout for the HTTP client
	if cfg.LLM.Timeout > 0 {
		provider.client.Timeout = time.Duration(cfg.LLM.Timeout) * time.Second
	}

	// Validate configuration
	if err := provider.ValidateConfig(); err != nil {
		return nil, err
	}

	return provider, nil
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig() error {
	// Validate API key
	if p.apiKey == "" {
		return llm.NewAuthenticationError("Anthropic API key is required")
	}

	// Validate model
	if p.model == "" {
		return llm.NewInvalidRequestError("Anthropic model is required")
	}

	return nil
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return "Anthropic"
}

// ReviewCode sends a code review request to Anthropic and returns the response
func (p *Provider) ReviewCode(ctx context.Context, request *llm.ReviewRequest) (*llm.ReviewResponse, error) {
	// Validate request
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// Create the messages for the API request
	// Get system message from centralized system prompts
	systemMessage := prompt.GetSystemPrompt(prompt.ProviderAnthropic, prompt.SystemPromptReview)
	
	// Create user message with prompt
	userPrompt := prompt.CreatePrompt(request, prompt.ProviderAnthropic)
	
	// Create messages
	messages := []map[string]interface{}{
		{
			"role":    "system",
			"content": systemMessage,
		},
		{
			"role":    "user",
			"content": userPrompt,
		},
	}

	// Define the FindDefinitionForType tool for the LLM to call
	tools := []map[string]interface{}{
		{
			"name": "FindDefinitionForType",
			"description": "Searches for and extracts a type definition by name in the codebase",
			"input_schema": map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"typeName": map[string]interface{}{
						"type": "string",
						"description": "The name of the type to find definition for",
					},
				},
				"required": []string{"typeName"},
			},
		},
	}

	// Create the API request
	apiRequest := map[string]interface{}{
		"model":       p.model,
		"messages":    messages,
		"tools":       tools,
		"max_tokens":  request.Options.MaxTokens,
		"temperature": request.Options.Temperature,
	}

	// Log the prompt if enabled
	userMessageContent := ""
	for _, msg := range messages {
		if role, ok := msg["role"].(string); ok && role == "user" {
			if content, ok := msg["content"].(string); ok {
				userMessageContent = content
				break
			}
		}
	}
	if userMessageContent != "" {
		if err := promptlog.LogPrompt(p.Name(), request.FilePath, userMessageContent); err != nil {
			logging.ErrorWith("Failed to log prompt", map[string]interface{}{
				"error": err.Error(),
			})
			// Continue without prompt logging, but log the error
		}
	}

	// Create a variable to store the API response
	var apiResponse map[string]interface{}

	// Convert the request to JSON
	requestBody, err := json.Marshal(apiRequest)
	if err != nil {
		return nil, llm.NewInvalidRequestError(fmt.Sprintf("failed to marshal request: %v", err))
	}

	// Function to create a new request with the same body
	createRequest := func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			fmt.Sprintf("%s/v1/messages", p.baseURL),
			bytes.NewBuffer(requestBody),
		)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Api-Key", p.apiKey)
		req.Header.Set("Anthropic-Version", apiVersion)
		return req, nil
	}

	// Function to make the API call with retry
	makeAPICall := func() (*http.Response, error) {
		req, err := createRequest()
		if err != nil {
			return nil, err
		}
		return p.client.Do(req)
	}

	// Make the API call with retry
	var resp *http.Response

	// Check if retry is enabled in the config
	if p.config != nil && p.config.Retry.Enabled {
		// Use the retry utility
		resp, err = retry.DoLLMRequest(makeAPICall, p.config)
	} else {
		// Make a single API call without retry
		resp, err = makeAPICall()
	}

	if err != nil {
		return nil, llm.NewProviderError(fmt.Sprintf("failed to send request: %v", err), err)
	}
	defer resp.Body.Close()

	// Check for error status code
	if resp.StatusCode != http.StatusOK {
		// Try to parse error response
		var errorResp map[string]interface{}
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &errorResp); err == nil {
			// Extract error message
			if errObj, ok := errorResp["error"].(map[string]interface{}); ok {
				if msg, ok := errObj["message"].(string); ok {
					return nil, llm.NewProviderError(fmt.Sprintf("API error: %s", msg), nil)
				}
			}
		}
		
		// Fallback error message
		return nil, llm.NewProviderError(fmt.Sprintf("API error: %d", resp.StatusCode), nil)
	}

	// Parse the response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, llm.NewProviderError(fmt.Sprintf("failed to read response body: %v", err), err)
	}
	
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, llm.NewInvalidRequestError(fmt.Sprintf("failed to parse response: %v", err))
	}

	// Handle tool calls and recursively process responses
	var handleToolCalls func() (*llm.ReviewResponse, error)
	handleToolCalls = func() (*llm.ReviewResponse, error) {
		// Check if the response contains a tool call
		var toolCalls []map[string]interface{}
		if content, exists := apiResponse["content"].([]interface{}); exists {
			for _, contentBlock := range content {
				if block, ok := contentBlock.(map[string]interface{}); ok {
					if block["type"] == "tool_call" {
						if toolCall, ok := block["tool_call"].(map[string]interface{}); ok {
							toolCalls = append(toolCalls, toolCall)
						}
					}
				}
			}
		}
		
		// If no tool calls, extract the text and return it
		if len(toolCalls) == 0 {
			// Get the completion from the API response
			completionContent, err := extractReviewText(apiResponse)
			if err != nil {
				return nil, err
			}
			
			// Process the response and return the result
			return &llm.ReviewResponse{
				Review:   completionContent,
				Metadata: map[string]interface{}{
					"model": p.model,
				},
			}, nil
		}
		
		// Process the first tool call
		toolCall := toolCalls[0]
		toolName, ok := toolCall["name"].(string)
		if !ok {
			return nil, llm.NewProviderError("invalid tool call format", nil)
		}
		
		// Currently we only support FindDefinitionForType
		if toolName == "FindDefinitionForType" {
			// Extract the parameters
			params, ok := toolCall["parameters"].(map[string]interface{})
			if !ok {
				return nil, llm.NewProviderError("invalid tool parameters", nil)
			}
			
			typeName, ok := params["typeName"].(string)
			if !ok {
				return nil, llm.NewProviderError("missing typeName parameter", nil)
			}
			
			// Get the extractor
			var codeExtractor *extractor.CodeExtractor
			var language extractor.Language

			if request.Extractor != nil {
				codeExtractor = request.Extractor
			} else {
				// Infer language from file extension
				ext := filepath.Ext(request.FilePath)
				switch strings.ToLower(ext) {
				case ".go":
					language = extractor.Go
				case ".js":
					language = extractor.JavaScript
				case ".py":
					language = extractor.Python
				default:
					return nil, llm.NewProviderError(fmt.Sprintf("unsupported file extension: %s", ext), nil)
				}

				// Get the working directory
				workDir, err := os.Getwd()
				if err != nil {
					return nil, llm.NewProviderError(fmt.Sprintf("failed to get working directory: %v", err), err)
				}

				// Create a new code extractor
				codeExtractor, err = extractor.NewCodeExtractor(language, workDir)
				if err != nil {
					return nil, llm.NewProviderError(fmt.Sprintf("failed to create code extractor: %v", err), err)
				}
			}
			
			// Call the function
			typeDefinition, filePath, err := codeExtractor.FindDefinitionForType(typeName)
			var toolResponse map[string]interface{}
			if err != nil {
				toolResponse = map[string]interface{}{
					"error": err.Error(),
				}
			} else {
				toolResponse = map[string]interface{}{
					"typeDefinition": typeDefinition,
					"filePath":       filePath,
				}
			}
			
			// Convert tool response to JSON
			toolResponseJSON, err := json.Marshal(toolResponse)
			if err != nil {
				return nil, llm.NewProviderError("failed to marshal tool response", err)
			}
			
			// Add the assistant's tool call message
			messages = append(messages, map[string]interface{}{
				"role": "assistant",
				"content": []map[string]interface{}{
					{
						"type": "tool_call",
						"tool_call": map[string]interface{}{
							"id": toolCall["id"],
							"name": "FindDefinitionForType",
							"parameters": params,
						},
					},
				},
			})
			
			// Add the tool response message
			messages = append(messages, map[string]interface{}{
				"role": "tool",
				"tool_call_id": toolCall["id"],
				"name": "FindDefinitionForType",
				"content": string(toolResponseJSON),
			})
			
			// Make another API call with the updated messages
			apiRequest["messages"] = messages
			
			// Convert the request to JSON
			requestBody, err = json.Marshal(apiRequest)
			if err != nil {
				return nil, llm.NewInvalidRequestError(fmt.Sprintf("failed to marshal request: %v", err))
			}
			
			// Make another API call with retry
			resp, err = retry.DoLLMRequest(makeAPICall, p.config)
			if err != nil {
				return nil, llm.NewProviderError(fmt.Sprintf("failed to send request: %v", err), err)
			}
			defer resp.Body.Close()
			
			// Read the response body
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, llm.NewProviderError(fmt.Sprintf("failed to read response body: %v", err), err)
			}
			
			// Parse the response
			if err := json.Unmarshal(body, &apiResponse); err != nil {
				return nil, llm.NewProviderError(fmt.Sprintf("failed to parse response: %v", err), err)
			}
			
			// Recursively handle more tool calls if needed
			return handleToolCalls()
		}
		
		return nil, llm.NewProviderError(fmt.Sprintf("unsupported tool: %s", toolName), nil)
	}

	// Process the response and return the result
	return handleToolCalls()
}

// extractReviewText extracts the review text from the API response
func extractReviewText(response map[string]interface{}) (string, error) {
	// Log the full response for debugging
	responseBytes, _ := json.Marshal(response)
	log.Printf("Raw Anthropic response: %s", string(responseBytes))

	// Extract content from the response
	content, ok := response["content"].([]interface{})
	if !ok || len(content) == 0 {
		log.Printf("Invalid response format: missing or empty content array")
		return "", llm.NewInvalidRequestError("invalid response format: missing content")
	}

	// Extract text from all content blocks (Anthropic may split the response)
	var fullText strings.Builder
	
	for i, contentBlock := range content {
		block, ok := contentBlock.(map[string]interface{})
		if !ok {
			log.Printf("Content block %d is not an object", i)
			continue
		}

		// Check content type
		contentType, ok := block["type"].(string)
		if !ok || contentType != "text" {
			log.Printf("Content block %d has invalid or missing type", i)
			continue
		}

		// Extract text
		text, ok := block["text"].(string)
		if !ok {
			log.Printf("Content block %d is missing text field", i)
			continue
		}

		fullText.WriteString(text)
	}

	extractedText := fullText.String()
	log.Printf("Extracted text from Anthropic response: %s", extractedText)
	
	// If we couldn't extract any text, return an error
	if extractedText == "" {
		return "", llm.NewInvalidRequestError("invalid response format: no text content found")
	}

	return extractedText, nil
}

// GetCompletion sends a prompt to the Anthropic API and returns the completion
func (p *Provider) GetCompletion(prompt string) (string, error) {
	// Create request body
	requestBody := map[string]interface{}{
		"model":      p.model,
		"max_tokens": 4096,
		"messages": []map[string]interface{}{
			{
				"role":    "user",
				"content": prompt,
			},
		},
	}

	// Convert request to JSON
	requestJSON, err := json.Marshal(requestBody)
	if err != nil {
		return "", llm.NewProviderError("failed to marshal request", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", p.baseURL+"/v1/messages", bytes.NewBuffer(requestJSON))
	if err != nil {
		return "", llm.NewProviderError("failed to create request", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", apiVersion)

	// Create retry options
	retryOpts := retry.DefaultOptions()
	retryOpts.MaxRetries = 3
	retryOpts.InitialDelay = 1 * time.Second
	retryOpts.MaxDelay = 10 * time.Second

	// Send request with retry
	var responseBody []byte
	result, err := retry.Do(func() (interface{}, error) {
		resp, err := p.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()
		
		// Read response body
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		
		// Check for error response
		if resp.StatusCode != http.StatusOK {
			var errorResponse struct {
				Error struct {
					Message string `json:"message"`
					Type    string `json:"type"`
				} `json:"error"`
			}
			if err := json.Unmarshal(body, &errorResponse); err == nil && errorResponse.Error.Message != "" {
				return nil, fmt.Errorf("Anthropic API error: %s", errorResponse.Error.Message)
			}
			return nil, fmt.Errorf("Anthropic API error: %d", resp.StatusCode)
		}
		
		return body, nil
	}, retryOpts)

	if err != nil {
		return "", llm.NewProviderError("failed to send request", err)
	}
	
	responseBody = result.([]byte)

	// Parse response
	var responseData map[string]interface{}
	if err := json.Unmarshal(responseBody, &responseData); err != nil {
		return "", llm.NewProviderError("failed to parse response", err)
	}

	// Extract content from response
	content, err := extractReviewText(responseData)
	if err != nil {
		return "", llm.NewProviderError("failed to extract content from response", err)
	}

	return content, nil
}

// init registers the Anthropic provider
func init() {
	// Register the Anthropic provider factory
	llm.RegisterProviderFactory("anthropic", func(cfg map[string]interface{}) (llm.Provider, error) {
		// Convert the generic config map to our specific config structure
		apiKey, ok := cfg["api_key"].(string)
		if !ok || apiKey == "" {
			return nil, llm.NewAuthenticationError("Anthropic API key is required")
		}
		
		model, ok := cfg["model"].(string)
		if !ok || model == "" {
			model = "claude-3-opus-20240229" // Default to Claude 3 Opus if not specified
		}
		
		apiURL, ok := cfg["api_url"].(string)
		if !ok || apiURL == "" {
			apiURL = "https://api.anthropic.com" // Default API URL
		}
		
		timeout, ok := cfg["timeout"].(int)
		if !ok || timeout <= 0 {
			timeout = 300 // Default timeout of 5 minutes
		}
		
		// Create a config object
		config := &config.Config{
			LLM: config.LLMConfig{
				Provider: "anthropic",
				APIURL:   apiURL,
				APIKey:   apiKey,
				Model:    model,
				Timeout:  timeout,
			},
		}
		
		return NewProvider(config)
	})
}
