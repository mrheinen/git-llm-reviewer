package openai

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/extractor"
	"github.com/niels/git-llm-review/pkg/llm"
	"github.com/niels/git-llm-review/pkg/llm/exchangelog"
	"github.com/niels/git-llm-review/pkg/llm/promptlog"
	"github.com/niels/git-llm-review/pkg/logging"
	"github.com/niels/git-llm-review/pkg/prompt"
	"github.com/niels/git-llm-review/pkg/util"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

// Provider implements the llm.Provider interface for OpenAI
type Provider struct {
	client openai.Client
	model  string
	config *config.Config
}

// NewProvider creates a new OpenAI provider with the given configuration
func NewProvider(cfg *config.Config) (llm.Provider, error) {
	// Validate API key
	if cfg.LLM.APIKey == "" {
		return nil, llm.NewAuthenticationError("OpenAI API key is required")
	}

	// Validate model
	if cfg.LLM.Model == "" {
		return nil, llm.NewProviderError("OpenAI model is required", llm.ErrConfigurationError)
	}

	// Set the OpenAI API options
	options := []option.RequestOption{
		option.WithAPIKey(cfg.LLM.APIKey),
	}

	// Set custom API URL if provided
	if cfg.LLM.APIURL != "" {
		options = append(options, option.WithBaseURL(cfg.LLM.APIURL))
	}

	// Create the OpenAI client
	client := openai.NewClient(options...)

	return &Provider{
		client: client,
		model:  cfg.LLM.Model,
		config: cfg,
	}, nil
}

// Name returns the name of the provider
func (p *Provider) Name() string {
	return "OpenAI"
}

// ValidateConfig validates the provider configuration
func (p *Provider) ValidateConfig() error {
	if p.model == "" {
		return llm.NewProviderError("OpenAI model is required", llm.ErrConfigurationError)
	}

	return nil
}

// ReviewCode performs a code review using the OpenAI API
func (p *Provider) ReviewCode(ctx context.Context, request *llm.ReviewRequest) (*llm.ReviewResponse, error) {
	// Validate the request
	if request == nil {
		return nil, errors.New("request is nil")
	}

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, request.Options.Timeout)
	defer cancel()

	// Get the user prompt from the prompt package
	userPrompt := prompt.CreatePrompt(request, prompt.ProviderOpenAI)

	// Create the messages for the chat completion
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage(prompt.GetSystemPrompt(prompt.ProviderOpenAI, prompt.SystemPromptReview)),
		openai.UserMessage(userPrompt),
	}

	// Define the FindDefinitionForType function as a tool
	tools := []openai.ChatCompletionToolParam{
		{
			Function: openai.FunctionDefinitionParam{
				Name:        "FindDefinitionForType",
				Description: openai.String("Searches for and extracts a type definition by name in the codebase"),
				Parameters: openai.FunctionParameters{
					"type": "object",
					"properties": map[string]interface{}{
						"typeName": map[string]interface{}{
							"type":        "string",
							"description": "The name of the type to find definition for",
						},
					},
					"required": []string{"typeName"},
				},
			},
		},
	}

	// Prepare the chat completion parameters
	params := openai.ChatCompletionNewParams{
		Messages: messages,
		Model:    p.model,
		Tools:    tools,
	}

	// Set temperature using the openai helper function
	params.Temperature = openai.Opt(request.Options.Temperature)

	// Set max tokens if specified
	if request.Options.MaxTokens > 0 {
		// Use the Int helper function from the SDK
		params.MaxTokens = openai.Int(int64(request.Options.MaxTokens))
	}

	// Log the prompt if enabled
	if err := promptlog.LogPrompt(p.Name(), request.FilePath, userPrompt); err != nil {
		logging.ErrorWith("Failed to log prompt", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue without prompt logging, but log the error
	}

	// Make the initial API call
	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return nil, llm.NewProviderError(fmt.Sprintf("failed to get chat completion: %v", err), err)
	}

	// Process the completion and check for tool calls
	for len(completion.Choices) > 0 && len(completion.Choices[0].Message.ToolCalls) > 0 {
		// Extract and process the tool calls
		hasToolCalls := false

		for _, toolCall := range completion.Choices[0].Message.ToolCalls {
			if toolCall.Function.Name == "FindDefinitionForType" {
				hasToolCalls = true

				// Extract the typeName parameter
				var args map[string]string
				if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
					return nil, llm.NewProviderError(fmt.Sprintf("failed to parse function arguments: %v", err), err)
				}

				typeName, ok := args["typeName"]
				if !ok {
					return nil, llm.NewProviderError("typeName parameter is required", nil)
				}

				// Create a code extractor if not available in the request
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
				var toolResult string
				if err != nil {
					toolResult = fmt.Sprintf("Error finding type definition: %s", err.Error())
				} else {
					resultMap := map[string]string{
						"typeDefinition": typeDefinition,
						"filePath":       filePath,
					}
					resultJSON, err := json.Marshal(resultMap)
					if err != nil {
						return nil, llm.NewProviderError(fmt.Sprintf("failed to marshal function result: %v", err), err)
					}
					toolResult = string(resultJSON)
				}

				// Add the assistant's message with the tool call to the messages
				params.Messages = append(params.Messages, completion.Choices[0].Message.ToParam())

				// Add the tool response to the messages
				params.Messages = append(params.Messages, openai.ToolMessage(toolResult, toolCall.ID))
			}
		}

		// If no tool calls were processed, break the loop
		if !hasToolCalls {
			break
		}

		// Make another API call with the updated messages
		completion, err = p.client.Chat.Completions.New(ctx, params)
		if err != nil {
			return nil, llm.NewProviderError(fmt.Sprintf("failed to get chat completion: %v", err), err)
		}
	}

	// Extract the final content from the completion
	content := completion.Choices[0].Message.Content

	// Remove <think>...</think> tags if present
	content = util.RemoveThinkTags(content)

	// Create the response with metadata
	return &llm.ReviewResponse{
		Review:     content,
		Confidence: 1.0, // OpenAI doesn't provide confidence scores
		Metadata: map[string]interface{}{
			"model":             p.model,
			"token_count":       completion.Usage.TotalTokens,
			"prompt_tokens":     completion.Usage.PromptTokens,
			"completion_tokens": completion.Usage.CompletionTokens,
			"finish_reason":     completion.Choices[0].FinishReason,
		},
	}, nil
}

// GetCompletion sends a prompt to the OpenAI API and returns the completion
func (p *Provider) GetCompletion(prompt string) (string, error) {
	// Create context
	ctx, cancel := context.WithTimeout(context.Background(), 600*time.Second)
	defer cancel()

	// Log the prompt if prompt logging is enabled
	if err := promptlog.LogPrompt(p.Name(), "prompt", prompt); err != nil {
		logging.WarnWith("Failed to log prompt", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue without prompt logging, but log the error
	}

	// Create chat completion params
	params := openai.ChatCompletionNewParams{
		Messages: []openai.ChatCompletionMessageParamUnion{
			openai.UserMessage(prompt),
		},
		Model: p.model,
	}

	// Set temperature
	params.Temperature = openai.Opt(0.7)

	// Make the API request
	completion, err := p.client.Chat.Completions.New(ctx, params)
	if err != nil {
		return "", llm.NewProviderError(fmt.Sprintf("failed to get chat completion: %v", err), err)
	}

	// Check if we have any choices
	if len(completion.Choices) == 0 {
		return "", llm.NewProviderError("no completion choices returned", nil)
	}

	// Extract content
	content := completion.Choices[0].Message.Content

	// Remove <think>...</think> tags if present
	content = util.RemoveThinkTags(content)

	// Log the full exchange if exchange logging is enabled
	if err := exchangelog.LogExchange(p.Name(), "prompt", prompt, content); err != nil {
		logging.WarnWith("Failed to log exchange", map[string]interface{}{
			"error": err.Error(),
		})
		// Continue without exchange logging, but log the error
	}

	// Return the processed content
	return content, nil
}

// init registers the OpenAI provider with the provider registry
func init() {
	// Register the OpenAI provider factory
	llm.RegisterProviderFactory("openai", func(cfg map[string]interface{}) (llm.Provider, error) {
		// Convert the generic config map to our specific config structure
		apiKey, ok := cfg["api_key"].(string)
		if !ok || apiKey == "" {
			return nil, llm.NewAuthenticationError("OpenAI API key is required")
		}

		model, ok := cfg["model"].(string)
		if !ok || model == "" {
			model = "gpt-4o" // Default to GPT-4o if not specified
		}

		apiURL, ok := cfg["api_url"].(string)
		if !ok || apiURL == "" {
			apiURL = "https://api.openai.com/v1" // Default API URL
		}

		timeout, ok := cfg["timeout"].(int)
		if !ok || timeout <= 0 {
			timeout = 300 // Default timeout of 5 minutes
		}

		// Create a config object
		config := &config.Config{
			LLM: config.LLMConfig{
				Provider: "openai",
				APIURL:   apiURL,
				APIKey:   apiKey,
				Model:    model,
				Timeout:  timeout,
			},
			Retry: config.RetryConfig{
				Enabled:      true,
				MaxRetries:   3,
				InitialDelay: 1,
				MaxDelay:     10,
			},
		}

		return NewProvider(config)
	})
}
