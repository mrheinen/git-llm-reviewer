package prompt

import (
	"strings"
	"path/filepath"
)

// GenerateReviewPrompt creates a prompt for code review based on the provided diff and file content
func GenerateReviewPrompt(diff string, fileContent map[string]string, provider string) string {
	// Get the appropriate template based on the provider
	var template string
	switch strings.ToLower(provider) {
	case "anthropic":
		template = AnthropicTemplate
	default:
		template = OpenAITemplate
	}

	// Create the final prompt by replacing placeholders
	var result strings.Builder

	// Split the diff by files 
	diffParts := strings.Split(diff, "File: ")
	
	// The first element is empty because diff starts with "File: ", so we skip it
	if len(diffParts) > 1 {
		for i := 1; i < len(diffParts); i++ {
			diffPart := diffParts[i]
			
			// Extract file path - it's the first line of each part
			lines := strings.SplitN(diffPart, "\n", 2)
			if len(lines) < 2 {
				continue
			}
			
			filePath := strings.TrimSpace(lines[0])
			fileDiff := lines[1]
			
			// Get file content if available
			content, hasContent := fileContent[filePath]
			
			// Get language based on file extension
			ext := filepath.Ext(filePath)
			language := ""
			if ext != "" {
				language = ext[1:] // remove the dot
			}
			
			// Replace placeholders in the template
			filePrompt := template
			filePrompt = strings.ReplaceAll(filePrompt, "{{.FilePath}}", filePath)
			filePrompt = strings.ReplaceAll(filePrompt, "{{.FileDiff}}", fileDiff)
			filePrompt = strings.ReplaceAll(filePrompt, "{{.Language}}", language)
			
			if hasContent {
				filePrompt = strings.ReplaceAll(filePrompt, "{{.FileContent}}", content)
				filePrompt = strings.ReplaceAll(filePrompt, "{{.HasFileContent}}", "true")
			} else {
				filePrompt = strings.ReplaceAll(filePrompt, "{{.FileContent}}", "")
				filePrompt = strings.ReplaceAll(filePrompt, "{{.HasFileContent}}", "false")
			}
			
			result.WriteString(filePrompt)
			result.WriteString("\n\n")
		}
	}
	
	return result.String()
}
