package parse

import (
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/niels/git-llm-review/pkg/util"
)

// JSONReviewResult represents the parsed result of a code review in JSON format
type JSONReviewResult struct {
	Issues []Issue    `json:"issues"`
	Diffs  []FileDiff `json:"diffs,omitempty"`
}

// Issue represents a single issue found in the code review
type Issue struct {
	Title       string `json:"title"`
	Explanation string `json:"explanation"`
	Diff        string `json:"diff,omitempty"` // Kept for backward compatibility
	File        string `json:"file,omitempty"` // New field to track which file the issue belongs to
}

// FileDiff represents a consolidated diff for a single file
type FileDiff struct {
	File string `json:"file"`
	Diff string `json:"diff"`
}

// ParseJSONReview parses a JSON review response from an LLM and extracts structured information
func ParseJSONReview(response string) (*JSONReviewResult, error) {
	// If response is empty, return empty result
	if response == "" {
		return &JSONReviewResult{Issues: []Issue{}}, nil
	}

	// First, remove any <think>...</think> tags if present
	response = util.RemoveThinkTags(response)

	// Clean the response to remove code blocks and trim whitespace
	cleanedResponse := cleanResponse(response)

	var result JSONReviewResult
	err := json.Unmarshal([]byte(cleanedResponse), &result)
	if err != nil {
		// Log error and JSON content (truncated if necessary)
		maxLogLength := 1000
		jsonToLog := cleanedResponse
		if len(jsonToLog) > maxLogLength {
			jsonToLog = jsonToLog[:maxLogLength] + "... [truncated]"
		}
		log.Printf("JSON parsing failed: %v\nJSON content: %s", err, jsonToLog)
		
		// Try one more approach - extract JSON from the response
		// This can happen if the LLM added extra text before or after the JSON
		content := extractJSONContent(response)
		if content != "" && content != cleanedResponse {
			var secondAttempt JSONReviewResult
			if err2 := json.Unmarshal([]byte(content), &secondAttempt); err2 == nil {
				// If this worked, use this result
				log.Printf("Second parsing attempt succeeded: %d issues", len(secondAttempt.Issues))
				return &secondAttempt, nil
			}
		}
		
		return &JSONReviewResult{Issues: []Issue{}}, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	// Log successful parse
	log.Printf("Successfully parsed response as JSON: %d issues, %d diffs",
		len(result.Issues), len(result.Diffs))

	// Post-process: ensure unique diffs per file by keeping only the first diff for each file
	if len(result.Diffs) > 0 {
		uniqueDiffs := make(map[string]FileDiff)
		for _, diff := range result.Diffs {
			if _, exists := uniqueDiffs[diff.File]; !exists {
				uniqueDiffs[diff.File] = diff
			}
		}

		// Convert back to slice
		result.Diffs = make([]FileDiff, 0, len(uniqueDiffs))
		for _, diff := range uniqueDiffs {
			result.Diffs = append(result.Diffs, diff)
		}
	}

	return &result, nil
}

// cleanResponse trims whitespace and removes markdown code block markers
func cleanResponse(response string) string {
	// Remove any markdown code block markers that might be wrapping the JSON
	response = regexp.MustCompile("(?s)^\\s*```json\\s*\\n(.*?)\\n\\s*```\\s*$").ReplaceAllString(response, "$1")
	response = regexp.MustCompile("(?s)^\\s*```\\s*\\n(.*?)\\n\\s*```\\s*$").ReplaceAllString(response, "$1")

	// Trim any leading/trailing whitespace
	return strings.TrimSpace(response)
}

// truncateForLogging truncates a string to the specified max length for logging purposes
func truncateForLogging(input string, maxLength int) string {
	if len(input) <= maxLength {
		return input
	}
	return input[:maxLength] + "... [truncated]"
}

// extractJSONContent extracts JSON content from a response that might contain additional text
// This function is kept for backward compatibility with tests
func extractJSONContent(response string) string {
	// Clean the response first
	response = cleanResponse(response)

	// Try to find a complete JSON object
	start := strings.Index(response, "{")
	if start == -1 {
		return ""
	}

	// Find the matching closing brace, accounting for quoted strings
	depth := 1
	inString := false
	escaped := false

	for end := start + 1; end < len(response); end++ {
		c := response[end]

		// Handle string escape sequences
		if escaped {
			// This character is escaped, so ignore it for parsing purposes
			escaped = false
			continue
		}

		// Check for escape character
		if c == '\\' {
			escaped = true
			continue
		}

		// Handle quote characters to track whether we're inside a string
		if c == '"' {
			inString = !inString
			continue
		}

		// Only process braces if we're not inside a string
		if !inString {
			switch c {
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return response[start : end+1]
				}
			}
		}
	}

	return "" // No matching JSON found
}