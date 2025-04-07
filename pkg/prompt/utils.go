package prompt

import (
	"fmt"
	"strings"
)

// FormatCodeBlock formats code with the appropriate language tag
func FormatCodeBlock(code string, language string) string {
	if language == "" {
		return fmt.Sprintf("```\n%s\n```", code)
	}
	return fmt.Sprintf("```%s\n%s\n```", language, code)
}

// FormatDiffBlock formats a diff with the appropriate diff tag
func FormatDiffBlock(diff string) string {
	return fmt.Sprintf("```diff\n%s\n```", diff)
}

// TruncatePrompt ensures the prompt doesn't exceed the token limit
func TruncatePrompt(prompt string, maxTokens int) string {
	// Very rough approximation: 1 token ≈ 4 characters for English text
	// This is a simplification; actual tokenization depends on the model
	if maxTokens <= 0 {
		return prompt
	}
	
	maxChars := maxTokens * 4
	if len(prompt) <= maxChars {
		return prompt
	}
	
	// If we need to truncate, preserve the beginning and end of the prompt
	// This is important to keep instructions and context
	preserveChars := maxChars / 5
	
	beginning := prompt[:preserveChars]
	end := prompt[len(prompt)-preserveChars:]
	
	// Add a note about truncation
	middle := "\n\n[... Content truncated due to length constraints ...]\n\n"
	
	return beginning + middle + end
}

// ExtractFileExtension extracts the file extension from a file path
func ExtractFileExtension(filePath string) string {
	parts := strings.Split(filePath, ".")
	if len(parts) < 2 {
		return ""
	}
	return parts[len(parts)-1]
}

// EnhancePromptWithLanguageContext adds language-specific context to the prompt
func EnhancePromptWithLanguageContext(prompt string, language string) string {
	if language == "" {
		return prompt
	}
	
	// Add language-specific context based on the file type
	var languageContext string
	
	switch strings.ToLower(language) {
	case "go":
		languageContext = "When reviewing Go code, consider Go's idioms and best practices such as error handling patterns, proper use of interfaces, and following the Go style guide."
	case "javascript", "typescript":
		languageContext = "When reviewing JavaScript/TypeScript code, look for common issues like type safety, async/await patterns, and modern ES6+ features."
	case "python":
		languageContext = "When reviewing Python code, consider PEP 8 style guidelines, proper exception handling, and Pythonic idioms."
	case "java":
		languageContext = "When reviewing Java code, look for proper exception handling, design patterns, and object-oriented principles."
	case "ruby":
		languageContext = "When reviewing Ruby code, consider Ruby idioms, proper use of blocks, and following the Ruby style guide."
	case "rust":
		languageContext = "When reviewing Rust code, look for proper memory management, use of ownership/borrowing, and idiomatic Rust patterns."
	default:
		// No specific context for other languages
		return prompt
	}
	
	// Insert the language context before the final instructions
	parts := strings.Split(prompt, "Please analyze the changes for:")
	if len(parts) != 2 {
		// If we can't find the marker, just append to the end
		return prompt + "\n\n" + languageContext
	}
	
	return parts[0] + languageContext + "\n\nPlease analyze the changes for:" + parts[1]
}
