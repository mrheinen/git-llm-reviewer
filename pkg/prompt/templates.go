package prompt

import (
	"github.com/niels/git-llm-review/pkg/llm"
)

// Templates for different providers

// DefaultTemplate is the standard template for code reviews
const DefaultTemplate = `You are a code review assistant. Please review the following code file and the changes made to it.

File: {{.FilePath}}

FULL FILE CONTENT:
` + "```{{.Language}}\n{{.FileContent}}\n```" + `

DIFF (CHANGES MADE):
` + "```diff\n{{.FileDiff}}\n```" + `

Please analyze the changes for:
1. Potential bugs or logic errors
2. Code style issues
3. Performance or efficiency concerns
4. Security vulnerabilities
5. Maintainability and readability

Provide your response in JSON format with the following structure:
{
  "issues": [
    {
      "title": "Brief issue title",
      "explanation": "Detailed explanation of the issue",
      "diff": "Code diff showing the fix (if applicable)"
    }
  ]
}

{{.AdditionalInstructions}}
`

// OpenAITemplate is the template for OpenAI code review requests
const OpenAITemplate = `You are a code review assistant. Please review the following code changes and provide feedback.

Your response should be in JSON format with the following structure:
{
  "issues": [
    {
      "title": "Issue title (e.g., 'Bug: Potential null pointer', 'Style: Inconsistent naming')",
      "explanation": "Detailed explanation of the issue",
      "file": "The file path where the issue is found"
    }
  ],
  "diffs": [
    {
      "file": "The file path",
      "diff": "Consolidated diff showing all suggested fixes for this file"
    }
  ]
}

IMPORTANT: Group all issues by file and provide only ONE consolidated diff per file that addresses all issues for that file. Do not include separate diffs for each issue. This ensures all fixes can be applied without conflicts.

IMPORTANT: The FULL FILE CONTENT below shows the STAGED version of the file. Make your diff suggestions AGAINST THIS STAGED VERSION, not against the original pre-staged version. The diff shown is for context only.

IMPORTANT: in your diff only propose changes that you know will work. Do not assume the existance of certain values and or keys unless you know they exist.
File: {{.FilePath}}

FULL FILE CONTENT (STAGED VERSION):
` + "```{{.Language}}\n{{.FileContent}}\n```" + `

DIFF (CHANGES MADE FROM ORIGINAL TO STAGED):
` + "```diff\n{{.FileDiff}}\n```" + `

Remember to focus on:
1. Bugs and potential issues
2. Code style and best practices
3. Performance concerns
4. Security vulnerabilities
5. Maintainability and readability

Use the full file context to provide a thorough analysis of the changes.
Provide your response in the JSON format specified above.

{{.AdditionalInstructions}}
`

// AnthropicTemplate is the template for Anthropic code review requests
const AnthropicTemplate = `Human: You are a code review assistant. Please review the following code changes and provide feedback.

Your response should be in JSON format with the following structure:
{
  "issues": [
    {
      "title": "Issue title (e.g., 'Bug: Potential null pointer', 'Style: Inconsistent naming')",
      "explanation": "Detailed explanation of the issue",
      "file": "The file path where the issue is found"
    }
  ],
  "diffs": [
    {
      "file": "The file path",
      "diff": "Consolidated diff showing all suggested fixes for this file"
    }
  ]
}

IMPORTANT: Group all issues by file and provide only ONE consolidated diff per file that addresses all issues for that file. Do not include separate diffs for each issue. This ensures all fixes can be applied without conflicts.

IMPORTANT: The FULL FILE CONTENT below shows the STAGED version of the file. Make your diff suggestions AGAINST THIS STAGED VERSION, not against the original pre-staged version. The diff shown is for context only.

File: {{.FilePath}}

FULL FILE CONTENT (STAGED VERSION):
` + "```{{.Language}}\n{{.FileContent}}\n```" + `

DIFF (CHANGES MADE FROM ORIGINAL TO STAGED):
` + "```diff\n{{.FileDiff}}\n```" + `

Remember to focus on:
1. Bugs and potential issues
2. Code style and best practices
3. Performance concerns
4. Security vulnerabilities
5. Maintainability and readability

Use the full file context to provide a thorough analysis of the changes.
Provide your response in the JSON format specified above.

{{.AdditionalInstructions}}
`

// TemplateData holds the data to be used in templates
type TemplateData struct {
	FilePath               string
	FileContent            string
	FileDiff               string
	Language               string
	AdditionalInstructions string
	HasFileContent         bool
}

// NewTemplateData creates a new TemplateData from a review request
func NewTemplateData(request *llm.ReviewRequest) *TemplateData {
	// Default values
	data := &TemplateData{
		FilePath:       request.FilePath,
		FileContent:    request.FileContent,
		FileDiff:       request.FileDiff,
		HasFileContent: request.FileContent != "",
	}

	// Detect language from file extension if not provided
	if data.Language == "" {
		data.Language = GetLanguageFromFilePath(request.FilePath)
	}

	return data
}
