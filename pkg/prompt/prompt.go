package prompt

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/niels/git-llm-review/pkg/llm"
)

// ProviderType represents the type of LLM provider
type ProviderType int

const (
	// ProviderDefault is the default provider type
	ProviderDefault ProviderType = iota
	// ProviderOpenAI is the OpenAI provider type
	ProviderOpenAI
	// ProviderAnthropic is the Anthropic provider type
	ProviderAnthropic
)

// CreatePrompt creates a prompt for the given request and provider
func CreatePrompt(request *llm.ReviewRequest, providerType ProviderType) string {
	// Create template data
	data := NewTemplateData(request)
	
	// Select the appropriate template based on provider type
	var tmplStr string
	switch providerType {
	case ProviderOpenAI:
		tmplStr = OpenAITemplate
	case ProviderAnthropic:
		tmplStr = AnthropicTemplate
	default:
		tmplStr = DefaultTemplate
	}
	
	// Parse and execute the template
	tmpl, err := template.New("prompt").Parse(tmplStr)
	if err != nil {
		// Fallback to simple string formatting if template parsing fails
		return fmt.Sprintf("Review code in %s. File content: %s. Changes: %s", 
			request.FilePath, request.FileContent, request.FileDiff)
	}
	
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		// Fallback to simple string formatting if template execution fails
		return fmt.Sprintf("Review code in %s. File content: %s. Changes: %s", 
			request.FilePath, request.FileContent, request.FileDiff)
	}
	
	return buf.String()
}

// FormatCodeForProvider formats code for the given provider
func FormatCodeForProvider(code string, providerType ProviderType) string {
	// All providers use markdown code blocks
	return fmt.Sprintf("```\n%s\n```", code)
}

// FormatDiffForProvider formats a diff for the given provider
func FormatDiffForProvider(diff string, providerType ProviderType) string {
	// All providers use markdown diff code blocks
	return fmt.Sprintf("```diff\n%s\n```", diff)
}

// GetLanguageFromFilePath determines the programming language from the file path
func GetLanguageFromFilePath(filePath string) string {
	// Extract extension
	parts := strings.Split(filePath, ".")
	if len(parts) < 2 {
		return ""
	}
	
	ext := parts[len(parts)-1]
	
	// Map extension to language
	switch strings.ToLower(ext) {
	case "go":
		return "go"
	case "js":
		return "javascript"
	case "ts":
		return "typescript"
	case "py":
		return "python"
	case "java":
		return "java"
	case "rb":
		return "ruby"
	case "php":
		return "php"
	case "c", "cpp", "cc":
		return "cpp"
	case "cs":
		return "csharp"
	case "rs":
		return "rust"
	case "swift":
		return "swift"
	case "kt":
		return "kotlin"
	case "sh":
		return "bash"
	case "html":
		return "html"
	case "css":
		return "css"
	case "json":
		return "json"
	case "yaml", "yml":
		return "yaml"
	case "md":
		return "markdown"
	default:
		return ""
	}
}
