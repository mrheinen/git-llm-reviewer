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
	Issues []Issue     `json:"issues"`
	Diffs  []FileDiff  `json:"diffs,omitempty"`
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
	response = cleanResponse(response)

	// Escape special characters in the JSON that might cause parsing issues
	processedResponse := preprocessSpecialChars(response)
	log.Printf("Processed response for JSON parsing, length: %d", len(processedResponse))

	// Try direct parsing of the processed response
	var directResult JSONReviewResult
	directErr := json.Unmarshal([]byte(processedResponse), &directResult)
	if directErr == nil && len(directResult.Issues) > 0 {
		log.Printf("Successfully parsed response directly as JSON: %d issues, %d diffs", 
			len(directResult.Issues), len(directResult.Diffs))
		return &directResult, nil
	}

	if directErr != nil {
		log.Printf("Direct JSON parsing failed: %v", directErr)
	}

	// If direct parsing failed, try the lenient approach
	lenientResult, lenientErr := ParseJSONReviewLenient(response)
	if lenientErr == nil && len(lenientResult.Issues) > 0 {
		return lenientResult, nil
	}

	// We should not reach here if both direct parsing and lenient parsing failed
	// Return a sensible default
	return &JSONReviewResult{Issues: []Issue{}}, fmt.Errorf("failed to parse JSON response")
}

// ParseJSONReviewLenient tries to parse a review result with more lenient extraction
func ParseJSONReviewLenient(response string) (*JSONReviewResult, error) {
	// Simple fallback: Just extract individual issues using regex
	re := regexp.MustCompile(`(?s)"title"\s*:\s*"([^"]+)".*?"explanation"\s*:\s*"([^"]+)"`)
	matches := re.FindAllStringSubmatch(response, -1)
	
	if len(matches) > 0 {
		log.Printf("Found %d issues with simple regex", len(matches))
		var issues []Issue
		for _, m := range matches {
			if len(m) >= 3 {
				issues = append(issues, Issue{
					Title:       strings.TrimSpace(m[1]),
					Explanation: strings.TrimSpace(m[2]),
				})
			}
		}

		// Also try to extract diffs if possible
		re = regexp.MustCompile(`(?s)"file"\s*:\s*"([^"]+)".*?"diff"\s*:\s*"(.*?)"(?:\s*\}|\s*,)`)
		diffMatches := re.FindAllStringSubmatch(response, -1)
		
		var diffs []FileDiff
		if len(diffMatches) > 0 {
			log.Printf("Found %d diffs with simple regex", len(diffMatches))
			for _, m := range diffMatches {
				if len(m) >= 3 {
					// Unescape the diff content (may contain escaped newlines, tabs, etc.)
					diffContent := m[2]
					diffContent = strings.ReplaceAll(diffContent, "\\n", "\n")
					diffContent = strings.ReplaceAll(diffContent, "\\t", "\t")
					diffContent = strings.ReplaceAll(diffContent, "\\\"", "\"")
					
					diffs = append(diffs, FileDiff{
						File: strings.TrimSpace(m[1]),
						Diff: diffContent,
					})
					log.Printf("Extracted diff for file %s with length %d", m[1], len(diffContent))
				}
			}
		}

		return &JSONReviewResult{Issues: issues, Diffs: diffs}, nil
	}

	// No issues found
	return &JSONReviewResult{Issues: []Issue{}}, fmt.Errorf("no issues found with lenient parsing")
}

// cleanResponse trims whitespace and removes unwanted characters from the response
func cleanResponse(response string) string {
	// Remove any markdown code block markers that might be wrapping the JSON
	response = regexp.MustCompile("(?s)^\\s*```json\\s*\\n(.*?)\\n\\s*```\\s*$").ReplaceAllString(response, "$1")
	response = regexp.MustCompile("(?s)^\\s*```\\s*\\n(.*?)\\n\\s*```\\s*$").ReplaceAllString(response, "$1")
	
	// Trim any leading/trailing whitespace
	return strings.TrimSpace(response)
}

// preprocessSpecialChars handles special characters in JSON strings that would cause parsing errors
func preprocessSpecialChars(input string) string {
	// Find all JSON string literals (text between quotes, handling escaped quotes)
	strLiteralRegex := regexp.MustCompile(`"((?:\\.|[^"])*?)"`) // matches content inside quotes, handling escaped quotes
	
	// Replace all string literals with properly escaped versions
	result := strLiteralRegex.ReplaceAllStringFunc(input, func(match string) string {
		// Remove quotes to process just the content
		content := match[1 : len(match)-1]
		
		// Escape problematic characters
		content = strings.ReplaceAll(content, "\t", "\\t") // Tab
		content = strings.ReplaceAll(content, "\r", "\\r") // Carriage return
		content = strings.ReplaceAll(content, "\n", "\\n") // Newline
		content = strings.ReplaceAll(content, "\b", "\\b") // Backspace
		content = strings.ReplaceAll(content, "\f", "\\f") // Form feed
		// Ensure quotes are properly escaped
		content = strings.ReplaceAll(content, "\\\"", "\\\"") // Already escaped quotes should be preserved
		content = strings.ReplaceAll(content, "\"", "\\\"") // Unescaped quotes should be escaped
		
		// Put quotes back
		return "\"" + content + "\""
	})
	
	// Fix any common JSON structural issues
	// Remove trailing commas before closing brackets
	result = regexp.MustCompile(`,(\s*)(\}|\])`).ReplaceAllString(result, "$1$2")
	
	log.Printf("Preprocessed JSON for special characters, original length: %d, new length: %d", len(input), len(result))
	if len(result) != len(input) {
		log.Printf("Special characters were processed in the JSON")
	}
	
	return result
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
	
	// Find the matching closing brace
	depth := 1
	for end := start + 1; end < len(response); end++ {
		switch response[end] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return response[start : end+1]
			}
		}
	}
	
	return "" // No matching JSON found
}


