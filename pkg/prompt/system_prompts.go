package prompt

// SystemPromptType represents the type of system prompt
type SystemPromptType string

const (
	// SystemPromptReview is the standard system prompt for code reviews
	SystemPromptReview SystemPromptType = "review"
	// SystemPromptExplain is a prompt for code explanation
	SystemPromptExplain SystemPromptType = "explain"
)

// GetSystemPrompt returns the system prompt for the specified provider and type
func GetSystemPrompt(providerType ProviderType, promptType SystemPromptType) string {
	// Get the base prompt by type
	basePrompt := getBaseSystemPrompt(promptType)

	// Apply provider-specific modifications if needed
	switch providerType {
	case ProviderOpenAI:
		return addOpenAISpecifics(basePrompt, promptType)
	case ProviderAnthropic:
		return addAnthropicSpecifics(basePrompt, promptType)
	default:
		return basePrompt
	}
}

// getBaseSystemPrompt returns the base system prompt for the specified type
func getBaseSystemPrompt(promptType SystemPromptType) string {
	switch promptType {
	case SystemPromptReview:
		return `
		
You are a code review assistant. 
Please review the code and provide helpful feedback on bugs, code style and performance issues. 
Take a special interest in security issues and help us fix them.

IMPORTANT: You MUST format your ENTIRE response as a valid JSON object with the following structure:
{
  "issues": [
    {
      "title": "Issue title",
      "explanation": "Detailed explanation",
      "file": "path/to/file.ext"
    }
  ],
  "diffs": [
    {
      "file": "path/to/file.ext",
      "diff": "@@ line numbers @@
code changes"
    }
  ]
}

Output raw JSON and do not use markdown code blocks or other wrapping.

Do NOT include any text before or after the JSON. Your entire response must be ONLY valid JSON.`

	case SystemPromptExplain:
		return `You are a code explanation assistant. Please explain the provided code in a clear and concise manner.`

	default:
		return ""
	}
}

// addOpenAISpecifics adds OpenAI-specific modifications to the system prompt
func addOpenAISpecifics(basePrompt string, promptType SystemPromptType) string {
	switch promptType {
	case SystemPromptReview:
		return basePrompt + `

You have access to the FindDefinitionForType function to get more context about types you see in the code. When you encounter a type that you need more information about, use this function to look up its definition. This will help you provide more accurate and insightful reviews.`
	default:
		return basePrompt
	}
}

// addAnthropicSpecifics adds Anthropic-specific modifications to the system prompt
func addAnthropicSpecifics(basePrompt string, promptType SystemPromptType) string {
	// Currently no Anthropic-specific modifications needed
	return basePrompt
}
