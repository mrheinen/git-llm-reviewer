package output

import (
	"strings"
	"testing"

	"github.com/niels/git-llm-review/pkg/parse"
)

func TestTerminalFormatter_FormatReview(t *testing.T) {
	// Test cases
	tests := []struct {
		name           string
		result         *parse.ReviewResult
		useColor       bool
		expectContains []string
		expectNotContains []string
	}{
		{
			name: "No issues",
			result: &parse.ReviewResult{
				Issues: []parse.Issue{},
			},
			useColor: true,
			expectContains: []string{
				"No issues found",
			},
		},
		{
			name: "Single issue with diff",
			result: &parse.ReviewResult{
				Issues: []parse.Issue{
					{
						Title:       "Bug: Potential null pointer dereference",
						Explanation: "This code might cause a null pointer dereference if the variable is not initialized.",
						Diff:        "+\tif x != nil {\n \tx.Method()\n+\t}",
					},
				},
			},
			useColor: true,
			expectContains: []string{
				"Bug",
				"Potential null pointer dereference",
				"This code might cause a null pointer dereference",
				"if x != nil",
				"\033[32m+", // Green color for additions
			},
		},
		{
			name: "Style issue without diff",
			result: &parse.ReviewResult{
				Issues: []parse.Issue{
					{
						Title:       "Style: Inconsistent naming convention",
						Explanation: "Variable names should use camelCase for consistency with the rest of the codebase.",
						Diff:        "",
					},
				},
			},
			useColor: true,
			expectContains: []string{
				"Style",
				"Inconsistent naming convention",
				"Variable names should use camelCase",
			},
			expectNotContains: []string{
				"Suggested changes:",
			},
		},
		{
			name: "Multiple issues with different types",
			result: &parse.ReviewResult{
				Issues: []parse.Issue{
					{
						Title:       "Bug: Missing error check",
						Explanation: "The error returned by this function is not being checked.",
						Diff:        "+\tif err != nil {\n+\t\treturn err\n+\t}",
					},
					{
						Title:       "Efficiency: Unnecessary allocation",
						Explanation: "This code is allocating memory unnecessarily, which could be avoided.",
						Diff:        "-\tresult := make([]string, 0)\n+\tvar result []string",
					},
				},
			},
			useColor: true,
			expectContains: []string{
				"Bug",
				"Efficiency",
				"Missing error check",
				"Unnecessary allocation",
				"\033[32m+", // Green color for additions
				"\033[31m-", // Red color for deletions
			},
		},
		{
			name: "No color output",
			result: &parse.ReviewResult{
				Issues: []parse.Issue{
					{
						Title:       "Bug: Missing error check",
						Explanation: "The error returned by this function is not being checked.",
						Diff:        "+\tif err != nil {\n+\t\treturn err\n+\t}",
					},
				},
			},
			useColor: false,
			expectContains: []string{
				"Bug",
				"Missing error check",
				"if err != nil",
			},
			expectNotContains: []string{
				"\033[", // No ANSI escape codes
			},
		},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewTerminalFormatter(tt.useColor)
			output := formatter.FormatReview(tt.result)

			// Check expected content
			for _, expected := range tt.expectContains {
				if !strings.Contains(output, expected) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s", expected, output)
				}
			}

			// Check not expected content
			for _, notExpected := range tt.expectNotContains {
				if strings.Contains(output, notExpected) {
					t.Errorf("Expected output to NOT contain %q, but it did.\nOutput: %s", notExpected, output)
				}
			}
		})
	}
}

func TestTerminalFormatter_formatIssueType(t *testing.T) {
	tests := []struct {
		title          string
		useColor       bool
		expectedPrefix string
		expectedColor  string
	}{
		{
			title:          "Bug: Something wrong",
			useColor:       true,
			expectedPrefix: "Bug",
			expectedColor:  "\033[1;31m", // Bold red
		},
		{
			title:          "Style: Formatting issue",
			useColor:       true,
			expectedPrefix: "Style",
			expectedColor:  "\033[1;33m", // Bold yellow
		},
		{
			title:          "Efficiency: Performance problem",
			useColor:       true,
			expectedPrefix: "Efficiency",
			expectedColor:  "\033[1;36m", // Bold cyan
		},
		{
			title:          "Security: Potential vulnerability",
			useColor:       true,
			expectedPrefix: "Security",
			expectedColor:  "\033[1;35m", // Bold magenta
		},
		{
			title:          "Other: Something else",
			useColor:       true,
			expectedPrefix: "Other",
			expectedColor:  "\033[1;37m", // Bold white
		},
		{
			title:          "No specific type",
			useColor:       true,
			expectedPrefix: "Issue",
			expectedColor:  "\033[1;37m", // Bold white
		},
		{
			title:          "Bug: Something wrong",
			useColor:       false,
			expectedPrefix: "Bug",
			expectedColor:  "", // No color
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			formatter := NewTerminalFormatter(tt.useColor)
			output := formatter.formatIssueType(tt.title)

			// Check prefix
			if !strings.Contains(output, tt.expectedPrefix) {
				t.Errorf("Expected output to contain prefix %q, but it didn't.\nOutput: %s", tt.expectedPrefix, output)
			}

			// Check color
			if tt.useColor && !strings.Contains(output, tt.expectedColor) {
				t.Errorf("Expected output to contain color code %q, but it didn't.\nOutput: %s", tt.expectedColor, output)
			}

			if !tt.useColor && strings.Contains(output, "\033[") {
				t.Errorf("Expected output to NOT contain color codes, but it did.\nOutput: %s", output)
			}
		})
	}
}

func TestTerminalFormatter_colorizeText(t *testing.T) {
	tests := []struct {
		text      string
		colorCode string
		useColor  bool
		expected  string
	}{
		{
			text:      "Hello",
			colorCode: "\033[31m", // Red
			useColor:  true,
			expected:  "\033[31mHello\033[0m",
		},
		{
			text:      "World",
			colorCode: "\033[32m", // Green
			useColor:  true,
			expected:  "\033[32mWorld\033[0m",
		},
		{
			text:      "No Color",
			colorCode: "\033[31m", // Red
			useColor:  false,
			expected:  "No Color",
		},
	}

	for _, tt := range tests {
		t.Run(tt.text, func(t *testing.T) {
			formatter := NewTerminalFormatter(tt.useColor)
			output := formatter.colorizeText(tt.text, tt.colorCode)

			if output != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, output)
			}
		})
	}
}

func TestTerminalFormatter_simpleColorizeDiff(t *testing.T) {
	tests := []struct {
		name     string
		diff     string
		useColor bool
		expected []string
		notExpected []string
	}{
		{
			name:     "With color",
			diff:     "+added line\n-removed line\n unchanged line",
			useColor: true,
			expected: []string{
				"\033[32m+added line\033[0m",
				"\033[31m-removed line\033[0m",
				" unchanged line",
			},
		},
		{
			name:     "Without color",
			diff:     "+added line\n-removed line\n unchanged line",
			useColor: false,
			expected: []string{
				"+added line",
				"-removed line",
				" unchanged line",
			},
			notExpected: []string{
				"\033[32m",
				"\033[31m",
				"\033[0m",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewTerminalFormatter(tt.useColor)
			output := formatter.simpleColorizeDiff(tt.diff)

			for _, exp := range tt.expected {
				if !strings.Contains(output, exp) {
					t.Errorf("Expected output to contain %q, but it didn't.\nOutput: %s", exp, output)
				}
			}

			for _, notExp := range tt.notExpected {
				if strings.Contains(output, notExp) {
					t.Errorf("Expected output to NOT contain %q, but it did.\nOutput: %s", notExp, output)
				}
			}
		})
	}
}