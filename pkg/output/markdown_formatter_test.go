package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/niels/git-llm-review/pkg/parse"
)

func TestMarkdownFormatter_FormatReview(t *testing.T) {
	// Create a sample review result
	result := &parse.ReviewResult{
		Issues: []parse.Issue{
			{
				Title:       "Bug: Potential null pointer dereference",
				Explanation: "This code might cause a null pointer dereference if the variable is not initialized.",
				Diff:        "```diff\n-foo.Bar()\n+if foo != nil {\n+\tfoo.Bar()\n+}```",
			},
			{
				Title:       "Style: Inconsistent naming convention",
				Explanation: "The variable name doesn't follow the camelCase convention used elsewhere.",
				Diff:        "```diff\n-var User_name string\n+var userName string```",
			},
		},
	}

	// Create a formatter
	formatter := NewMarkdownFormatter()

	// Format the review
	markdown := formatter.FormatReview(result, "test.go", "example-repo")

	// Verify markdown contains expected elements
	t.Run("Contains headers", func(t *testing.T) {
		if !strings.Contains(markdown, "# Code Review Report") {
			t.Error("Markdown should contain a main header")
		}
		if !strings.Contains(markdown, "## File: test.go") {
			t.Error("Markdown should contain a file header")
		}
	})

	t.Run("Contains metadata", func(t *testing.T) {
		if !strings.Contains(markdown, "Repository: example-repo") {
			t.Error("Markdown should contain repository information")
		}
		if !strings.Contains(markdown, "Generated on:") {
			t.Error("Markdown should contain generation timestamp")
		}
	})

	t.Run("Contains issue details", func(t *testing.T) {
		if !strings.Contains(markdown, "### Bug: Potential null pointer dereference") {
			t.Error("Markdown should contain issue title as header")
		}
		if !strings.Contains(markdown, "This code might cause a null pointer") {
			t.Error("Markdown should contain issue explanation")
		}
	})

	t.Run("Contains properly formatted code blocks", func(t *testing.T) {
		if !strings.Contains(markdown, "```diff") {
			t.Error("Markdown should contain diff code blocks")
		}
		// Check that the diff is properly formatted without extra backticks
		if strings.Contains(markdown, "```diff\n-foo.Bar()\n+if foo != nil {\n+\tfoo.Bar()\n+}```") {
			t.Error("Markdown should properly format diff code blocks (remove trailing backticks)")
		}
		if !strings.Contains(markdown, "```diff\n-foo.Bar()\n+if foo != nil {\n+\tfoo.Bar()\n+}\n```") {
			t.Error("Markdown should properly format diff code blocks (add newline before closing backticks)")
		}
	})

	// Test empty result
	emptyResult := &parse.ReviewResult{
		Issues: []parse.Issue{},
	}
	emptyMarkdown := formatter.FormatReview(emptyResult, "test.go", "example-repo")
	
	t.Run("Handles empty results", func(t *testing.T) {
		if !strings.Contains(emptyMarkdown, "No issues found") {
			t.Error("Markdown should indicate when no issues are found")
		}
	})
}

func TestMarkdownFormatter_WriteToFile(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "markdown-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample review result
	result := &parse.ReviewResult{
		Issues: []parse.Issue{
			{
				Title:       "Bug: Potential null pointer dereference",
				Explanation: "This code might cause a null pointer dereference if the variable is not initialized.",
				Diff:        "```diff\n-foo.Bar()\n+if foo != nil {\n+\tfoo.Bar()\n+}```",
			},
		},
	}

	// Create a formatter
	formatter := NewMarkdownFormatter()

	// Test writing to a file
	outputPath := filepath.Join(tempDir, "review.md")
	err = formatter.WriteToFile(result, "test.go", "example-repo", outputPath)
	if err != nil {
		t.Fatalf("Failed to write markdown to file: %v", err)
	}

	// Verify the file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file should exist")
	}

	// Read the file content
	content, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	// Verify content
	if !strings.Contains(string(content), "# Code Review Report") {
		t.Error("Output file should contain markdown content")
	}

	// Test writing to a non-existent directory
	nonExistentDir := filepath.Join(tempDir, "non-existent")
	outputPath = filepath.Join(nonExistentDir, "review.md")
	err = formatter.WriteToFile(result, "test.go", "example-repo", outputPath)
	if err != nil {
		t.Fatalf("Failed to write markdown to file in non-existent directory: %v", err)
	}

	// Verify the directory was created
	if _, err := os.Stat(nonExistentDir); os.IsNotExist(err) {
		t.Error("Output directory should be created if it doesn't exist")
	}

	// Verify the file exists
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Error("Output file should exist in the created directory")
	}
}

func TestMarkdownFormatter_GroupIssuesByType(t *testing.T) {
	// Create a sample review result with different issue types
	result := &parse.ReviewResult{
		Issues: []parse.Issue{
			{
				Title:       "Bug: Potential null pointer dereference",
				Explanation: "This code might cause a null pointer dereference if the variable is not initialized.",
				Diff:        "```diff\n-foo.Bar()\n+if foo != nil {\n+\tfoo.Bar()\n+}```",
			},
			{
				Title:       "Style: Inconsistent naming convention",
				Explanation: "The variable name doesn't follow the camelCase convention used elsewhere.",
				Diff:        "```diff\n-var User_name string\n+var userName string```",
			},
			{
				Title:       "Bug: Array index out of bounds",
				Explanation: "This code might access an array index that is out of bounds.",
				Diff:        "```diff\n-arr[len(arr)]\n+arr[len(arr)-1]```",
			},
		},
	}

	// Create a formatter
	formatter := NewMarkdownFormatter()

	// Format the review
	markdown := formatter.FormatReview(result, "test.go", "example-repo")

	// Verify markdown groups issues by type
	t.Run("Groups issues by type", func(t *testing.T) {
		if !strings.Contains(markdown, "## Bug Issues") {
			t.Error("Markdown should group bug issues under a 'Bug Issues' header")
		}
		if !strings.Contains(markdown, "## Style Issues") {
			t.Error("Markdown should group style issues under a 'Style Issues' header")
		}
	})

	// Verify issues are listed under their respective type headers
	t.Run("Lists issues under type headers", func(t *testing.T) {
		// Find the position of the "Bug Issues" header
		bugHeaderPos := strings.Index(markdown, "## Bug Issues")
		styleHeaderPos := strings.Index(markdown, "## Style Issues")
		
		if bugHeaderPos == -1 || styleHeaderPos == -1 {
			t.Fatal("Could not find expected headers")
		}
		
		// Check that bug issues are listed under the bug header
		bugSection := markdown[bugHeaderPos:styleHeaderPos]
		if !strings.Contains(bugSection, "### Bug: Potential null pointer dereference") {
			t.Error("Bug section should contain the null pointer issue")
		}
		if !strings.Contains(bugSection, "### Bug: Array index out of bounds") {
			t.Error("Bug section should contain the array index issue")
		}
		
		// Check that style issues are listed under the style header
		styleSection := markdown[styleHeaderPos:]
		if !strings.Contains(styleSection, "### Style: Inconsistent naming convention") {
			t.Error("Style section should contain the naming convention issue")
		}
	})
}
