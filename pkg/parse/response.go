package parse

import (
	"fmt"
	"strings"
)

// ReviewResult represents the parsed result of a code review
type ReviewResult struct {
	Issues []Issue
	Diffs  []FileDiff
}

// ParseReview parses a review response and returns a ReviewResult
func ParseReview(response string) *ReviewResult {
	if response == "" {
		return &ReviewResult{Issues: []Issue{}, Diffs: []FileDiff{}}
	}

	// Parse as JSON (since we are always expecting valid JSON now)
	jsonResult, err := ParseJSONReview(response)
	if err == nil && jsonResult != nil && (len(jsonResult.Issues) > 0 || len(jsonResult.Diffs) > 0) {
		// Ensure we have the required fields in each issue
		jsonResult.Issues = cleanupIssues(jsonResult.Issues)
		return &ReviewResult{
			Issues: jsonResult.Issues,
			Diffs:  jsonResult.Diffs,
		}
	}

	// If parsing failed, return empty result
	return &ReviewResult{Issues: []Issue{}, Diffs: []FileDiff{}}
}

// GetIssueCount returns the number of issues in the review result
func (r *ReviewResult) GetIssueCount() int {
	if r == nil {
		return 0
	}
	return len(r.Issues)
}

// GetDiffCount returns the number of consolidated diffs in the review result
func (r *ReviewResult) GetDiffCount() int {
	if r == nil {
		return 0
	}
	return len(r.Diffs)
}

// String returns a string representation of the review result
func (r *ReviewResult) String() string {
	if r == nil || (len(r.Issues) == 0 && len(r.Diffs) == 0) {
		return "No issues found."
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d issues:\n\n", len(r.Issues)))

	// Print issues without grouping by file to match test expectations
	for i, issue := range r.Issues {
		sb.WriteString(fmt.Sprintf("Issue %d: %s\n", i+1, issue.Title))
		sb.WriteString(fmt.Sprintf("Explanation: %s\n", issue.Explanation))
		
		// If there's a diff, show it
		if issue.Diff != "" {
			sb.WriteString("Suggested changes:\n")
			sb.WriteString("```diff\n")
			sb.WriteString(issue.Diff)
			sb.WriteString("\n```\n")
		}
		
		sb.WriteString("\n")
	}
	
	// Add the consolidated diffs
	if len(r.Diffs) > 0 {
		sb.WriteString("\nConsolidated diffs by file:\n")
		sb.WriteString("========================================\n\n")
		
		for _, diff := range r.Diffs {
			sb.WriteString(fmt.Sprintf("File: %s\n", diff.File))
			sb.WriteString("----------------------------------------\n")
			sb.WriteString("```diff\n")
			sb.WriteString(diff.Diff)
			sb.WriteString("\n```\n\n")
		}
	}

	return sb.String()
}

// cleanupIssues ensures that all issues have the required fields
func cleanupIssues(issues []Issue) []Issue {
	for i := range issues {
		// Ensure Title has content
		if issues[i].Title == "" {
			issues[i].Title = "Unnamed Issue"
		}
		
		// Ensure Explanation has content
		if issues[i].Explanation == "" {
			issues[i].Explanation = "No explanation provided."
		}
		
		// Make sure File field is properly set (optional)
		if issues[i].File == "" {
			issues[i].File = "General"
		}
		
		// Diff can be empty, no need to modify
	}
	return issues
}
