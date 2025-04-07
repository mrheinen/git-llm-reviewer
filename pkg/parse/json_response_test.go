package parse

import (
	"reflect"
	"testing"
)

func TestParseJSONReview(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     *JSONReviewResult
		wantErr  bool
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
    },
    {
      "title": "Inefficient string concatenation",
      "explanation": "Using the + operator in a loop for string concatenation is inefficient. Consider using strings.Builder instead.",
      "diff": "@@ -20,9 +20,11 @@\n func concatenateStrings(items []string) string {\n-\tresult := \"\"\n+\tvar builder strings.Builder\n \tfor _, item := range items {\n-\t\tresult += item\n+\t\tbuilder.WriteString(item)\n \t}\n-\treturn result\n+\treturn builder.String()\n }"
    }
  ]
}`,
			want: &JSONReviewResult{
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
					{
						Title:       "Inefficient string concatenation",
						Explanation: "Using the + operator in a loop for string concatenation is inefficient. Consider using strings.Builder instead.",
						Diff:        "@@ -20,9 +20,11 @@\n func concatenateStrings(items []string) string {\n-\tresult := \"\"\n+\tvar builder strings.Builder\n \tfor _, item := range items {\n-\t\tresult += item\n+\t\tbuilder.WriteString(item)\n \t}\n-\treturn result\n+\treturn builder.String()\n }",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Valid JSON response with no issues",
			response: `{
  "issues": []
}`,
			want: &JSONReviewResult{
				Issues: []Issue{},
			},
			wantErr: false,
		},
		{
			name:     "Empty response",
			response: "",
			want: &JSONReviewResult{
				Issues: []Issue{},
			},
			wantErr: false,
		},
		{
			name: "JSON with extra text before and after",
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
			want: &JSONReviewResult{
				Issues: []Issue{
					{
						Title:       "Use fmt.Println instead of println",
						Explanation: "The code is using the built-in println function, which is meant for debugging.",
						Diff:        "@@ -4,5 +4,5 @@\n import \"fmt\"\n \n func main() {\n-\tprintln(\"Hello, World!\")\n+\tfmt.Println(\"Hello, World!\")\n }",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "Invalid JSON",
			response: `{
  "issues": [
    {
      "title": "Use fmt.Println instead of println",
      "explanation": "The code is using the built-in println function, which is meant for debugging.",
      "diff": "@@ -4,5 +4,5 @@\n import \"fmt\"\n \n func main() {\n-\tprintln(\"Hello, World!\")\n+\tfmt.Println(\"Hello, World!\")\n }"
    },
    INVALID JSON HERE
  ]
}`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseJSONReview(tt.response)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONReview() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ParseJSONReview() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractJSONContent(t *testing.T) {
	tests := []struct {
		name     string
		response string
		want     string
	}{
		{
			name: "Extract JSON from clean response",
			response: `{
  "issues": [
    {
      "title": "Test issue",
      "explanation": "Test explanation",
      "diff": "Test diff"
    }
  ]
}`,
			want: `{
  "issues": [
    {
      "title": "Test issue",
      "explanation": "Test explanation",
      "diff": "Test diff"
    }
  ]
}`,
		},
		{
			name: "Extract JSON with text before and after",
			response: `Here's my review:

{
  "issues": [
    {
      "title": "Test issue",
      "explanation": "Test explanation",
      "diff": "Test diff"
    }
  ]
}

Let me know if you have any questions!`,
			want: `{
  "issues": [
    {
      "title": "Test issue",
      "explanation": "Test explanation",
      "diff": "Test diff"
    }
  ]
}`,
		},
		{
			name:     "No JSON content",
			response: "This response doesn't contain any JSON.",
			want:     "",
		},
		{
			name: "Nested JSON objects",
			response: `{
  "issues": [
    {
      "title": "Test issue",
      "explanation": "Test explanation with { nested braces }",
      "diff": "Test diff",
      "metadata": {
        "severity": "high",
        "category": "bug"
      }
    }
  ]
}`,
			want: `{
  "issues": [
    {
      "title": "Test issue",
      "explanation": "Test explanation with { nested braces }",
      "diff": "Test diff",
      "metadata": {
        "severity": "high",
        "category": "bug"
      }
    }
  ]
}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJSONContent(tt.response)
			if got != tt.want {
				t.Errorf("extractJSONContent() = %v, want %v", got, tt.want)
			}
		})
	}
}
