package parse

import (
	"reflect"
	"testing"
)

func TestParseReview(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     *ReviewResult
	}{
		{
			name: "Valid JSON response with multiple issues",
			response: `{
  "issues": [
    {
      "title": "Use fmt.Println instead of println",
      "explanation": "The code is using the built-in println function, which is meant for debugging. For production code, you should use fmt.Println instead.",
      "diff": "@@ -4,5 +4,5 @@\n import \"fmt\"\n \n func main() {\n-\tprintln(\"Hello, World!\")\n+\tfmt.Println(\"Hello, World!\")\n }"
    },
    {
      "title": "Missing error handling",
      "explanation": "The function doesn't handle potential errors from the file operations.",
      "diff": "@@ -10,7 +10,10 @@\n func readFile(filename string) string {\n-\tdata, _ := ioutil.ReadFile(filename)\n+\tdata, err := ioutil.ReadFile(filename)\n+\tif err != nil {\n+\t\tlog.Fatalf(\"failed to read file: %v\", err)\n+\t}\n \treturn string(data)\n }"
    }
  ]
}`,
			want: &ReviewResult{
				Issues: []Issue{
					{
						Title:       "Use fmt.Println instead of println",
						Explanation: "The code is using the built-in println function, which is meant for debugging. For production code, you should use fmt.Println instead.",
						Diff:        "@@ -4,5 +4,5 @@\n import \"fmt\"\n \n func main() {\n-\tprintln(\"Hello, World!\")\n+\tfmt.Println(\"Hello, World!\")\n }",
					},
					{
						Title:       "Missing error handling",
						Explanation: "The function doesn't handle potential errors from the file operations.",
						Diff:        "@@ -10,7 +10,10 @@\n func readFile(filename string) string {\n-\tdata, _ := ioutil.ReadFile(filename)\n+\tdata, err := ioutil.ReadFile(filename)\n+\tif err != nil {\n+\t\tlog.Fatalf(\"failed to read file: %v\", err)\n+\t}\n \treturn string(data)\n }",
					},
				},
			},
		},
		{
			name: "JSON with text before and after",
			response: `I've reviewed your code and found some issues:

{
  "issues": [
    {
      "title": "Use fmt.Println instead of println",
      "explanation": "The code is using the built-in println function, which is meant for debugging.",
      "diff": "@@ -4,5 +4,5 @@\n import \"fmt\"\n \n func main() {\n-\tprintln(\"Hello, World!\")\n+\tfmt.Println(\"Hello, World!\")\n }"
    }
  ]
}

Let me know if you have any questions!`,
			want: &ReviewResult{
				Issues: []Issue{
					{
						Title:       "Use fmt.Println instead of println",
						Explanation: "The code is using the built-in println function, which is meant for debugging.",
						Diff:        "@@ -4,5 +4,5 @@\n import \"fmt\"\n \n func main() {\n-\tprintln(\"Hello, World!\")\n+\tfmt.Println(\"Hello, World!\")\n }",
					},
				},
			},
		},
		{
			name:     "Empty response",
			response: "",
			want: &ReviewResult{
				Issues: []Issue{},
			},
		},
		{
			name: "Invalid JSON",
			response: `{
  "issues": [
    {
      "title": "Use fmt.Println instead of println",
      "explanation": "The code is using the built-in println function, which is meant for debugging.",
      INVALID JSON HERE
    }
  ]
}`,
			want: &ReviewResult{
				Issues: []Issue{},
			},
		},
		{
			name: "Non-JSON response",
			response: `Here are some issues I found:

1. Use fmt.Println instead of println
   The code is using the built-in println function, which is meant for debugging.

2. Missing error handling
   The function doesn't handle potential errors from the file operations.`,
			want: &ReviewResult{
				Issues: []Issue{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseReview(tt.response)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseReview() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReviewResult_GetIssueCount(t *testing.T) {
	tests := []struct {
		name   string
		result *ReviewResult
		want   int
	}{
		{
			name: "Multiple issues",
			result: &ReviewResult{
				Issues: []Issue{
					{Title: "Issue 1"},
					{Title: "Issue 2"},
				},
			},
			want: 2,
		},
		{
			name: "Single issue",
			result: &ReviewResult{
				Issues: []Issue{
					{Title: "Issue 1"},
				},
			},
			want: 1,
		},
		{
			name: "No issues",
			result: &ReviewResult{
				Issues: []Issue{},
			},
			want: 0,
		},
		{
			name:   "Nil result",
			result: nil,
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.GetIssueCount(); got != tt.want {
				t.Errorf("ReviewResult.GetIssueCount() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReviewResult_HasIssues(t *testing.T) {
	tests := []struct {
		name   string
		result *ReviewResult
		want   bool
	}{
		{
			name: "With issues",
			result: &ReviewResult{
				Issues: []Issue{
					{
						Title:       "Issue 1",
						Explanation: "Explanation 1",
					},
				},
			},
			want: true,
		},
		{
			name: "No issues",
			result: &ReviewResult{
				Issues: []Issue{},
			},
			want: false,
		},
		{
			name:   "Nil result",
			result: nil,
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.GetIssueCount() > 0; got != tt.want {
				t.Errorf("ReviewResult.GetIssueCount() > 0 = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestReviewResult_String(t *testing.T) {
	tests := []struct {
		name   string
		result *ReviewResult
		want   string
	}{
		{
			name: "Multiple issues with diffs",
			result: &ReviewResult{
				Issues: []Issue{
					{
						Title:       "Issue 1",
						Explanation: "Explanation 1",
						Diff:        "- old code\n+ new code",
					},
					{
						Title:       "Issue 2",
						Explanation: "Explanation 2",
						Diff:        "- another old\n+ another new",
					},
				},
			},
			want: "Found 2 issues:\n\nIssue 1: Issue 1\nExplanation: Explanation 1\nSuggested changes:\n```diff\n- old code\n+ new code\n```\n\nIssue 2: Issue 2\nExplanation: Explanation 2\nSuggested changes:\n```diff\n- another old\n+ another new\n```\n\n",
		},
		{
			name: "Issue without diff",
			result: &ReviewResult{
				Issues: []Issue{
					{
						Title:       "Issue 1",
						Explanation: "Explanation 1",
					},
				},
			},
			want: "Found 1 issues:\n\nIssue 1: Issue 1\nExplanation: Explanation 1\n\n",
		},
		{
			name: "No issues",
			result: &ReviewResult{
				Issues: []Issue{},
			},
			want: "No issues found.",
		},
		{
			name:   "Nil result",
			result: nil,
			want:   "No issues found.",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.result.String(); got != tt.want {
				t.Errorf("ReviewResult.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
