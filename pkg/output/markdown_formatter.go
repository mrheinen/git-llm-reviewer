package output

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/niels/git-llm-review/pkg/parse"
)

// MarkdownFormatter formats review results as markdown
type MarkdownFormatter struct{}

// NewMarkdownFormatter creates a new markdown formatter
func NewMarkdownFormatter() *MarkdownFormatter {
	return &MarkdownFormatter{}
}

// FormatReview formats a review result as markdown
func (f *MarkdownFormatter) FormatReview(result *parse.ReviewResult, filePath, repoName string) string {
	var sb strings.Builder

	// Add header and metadata
	sb.WriteString("# Code Review Report\n\n")
	sb.WriteString(fmt.Sprintf("## File: %s\n\n", filePath))
	sb.WriteString(fmt.Sprintf("Repository: %s\n\n", repoName))
	sb.WriteString(fmt.Sprintf("Generated on: %s\n\n", time.Now().Format(time.RFC1123)))

	// Handle empty result
	if result == nil || len(result.Issues) == 0 {
		sb.WriteString("No issues found in this file.\n")
		return sb.String()
	}

	// Group issues by type
	issuesByType := f.groupIssuesByType(result.Issues)

	// Sort issue types for consistent output
	var types []string
	for issueType := range issuesByType {
		types = append(types, issueType)
	}
	sort.Strings(types)

	// Process each issue type
	for _, issueType := range types {
		issues := issuesByType[issueType]
		sb.WriteString(fmt.Sprintf("## %s Issues\n\n", issueType))

		for _, issue := range issues {
			// Add issue title as header
			sb.WriteString(fmt.Sprintf("### %s\n\n", issue.Title))
			
			// Add explanation
			sb.WriteString(fmt.Sprintf("%s\n\n", issue.Explanation))
			
			// Add diff if available
			if issue.Diff != "" {
				// Format the diff properly for markdown
				formattedDiff := f.formatDiff(issue.Diff)
				sb.WriteString(fmt.Sprintf("%s\n\n", formattedDiff))
			}
		}
	}

	return sb.String()
}

// WriteToFile writes the formatted review to a file
func (f *MarkdownFormatter) WriteToFile(result *parse.ReviewResult, filePath, repoName, outputPath string) error {
	// Format the review
	markdown := f.FormatReview(result, filePath, repoName)

	// Create the directory if it doesn't exist
	dir := filepath.Dir(outputPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write to file
	if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	return nil
}

// groupIssuesByType groups issues by their type (Bug, Style, etc.)
func (f *MarkdownFormatter) groupIssuesByType(issues []parse.Issue) map[string][]parse.Issue {
	result := make(map[string][]parse.Issue)

	for _, issue := range issues {
		// Extract issue type from title
		issueType := f.extractIssueType(issue.Title)
		
		// Add to the appropriate group
		result[issueType] = append(result[issueType], issue)
	}

	return result
}

// extractIssueType extracts the issue type from the title
func (f *MarkdownFormatter) extractIssueType(title string) string {
	// Default type if we can't extract one
	defaultType := "General"

	// Check if title follows the "Type: Description" format
	parts := strings.SplitN(title, ":", 2)
	if len(parts) < 2 {
		return defaultType
	}

	// Extract and clean the type
	issueType := strings.TrimSpace(parts[0])
	
	// Common issue types
	switch strings.ToLower(issueType) {
	case IssueTypeBug:
		return IssueTypeBugDisplay
	case IssueTypeStyle, "formatting":
		return IssueTypeStyleDisplay
	case IssueTypePerformance, IssueTypePerf, "efficiency":
		return IssueTypePerformanceDisplay
	case IssueTypeSecurity:
		return IssueTypeSecurityDisplay
	case IssueTypeMaintainability, IssueTypeMaintenance, IssueTypeMaintain, "readability":
		return IssueTypeMaintainabilityDisplay
	default:
		// If it's capitalized, it's probably a valid type
		if len(issueType) > 0 && issueType[0] >= 'A' && issueType[0] <= 'Z' {
			return issueType
		}
		return defaultType
	}
}

// formatDiff formats a diff for markdown
func (f *MarkdownFormatter) formatDiff(diff string) string {
	// Clean up diff format
	// Some LLMs might include triple backticks in their output
	diffContent := diff
	
	// Remove triple backticks if they exist
	re := regexp.MustCompile("^```diff\n")
	diffContent = re.ReplaceAllString(diffContent, "")
	
	re = regexp.MustCompile("```$")
	diffContent = re.ReplaceAllString(diffContent, "")
	
	// Ensure diff has proper markdown code block format
	if !strings.HasPrefix(diffContent, "```diff\n") {
		diffContent = "```diff\n" + diffContent
	}
	
	// Ensure diff ends with a newline before the closing backticks
	if !strings.HasSuffix(diffContent, "\n```") {
		if strings.HasSuffix(diffContent, "```") {
			// Remove trailing backticks and add newline
			diffContent = strings.TrimSuffix(diffContent, "```") + "\n```"
		} else {
			// Just add closing backticks with newline
			diffContent = diffContent + "\n```"
		}
	}
	
	return diffContent
}
