package prompt

import (
	"strings"
	"testing"

	"github.com/niels/git-llm-review/pkg/llm"
)

func TestCreatePrompt(t *testing.T) {
	// Test file content and diff
	fileContent := `package main

import "fmt"

func main() {
	fmt.Println("Hello, World!")
}
`
	fileDiff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,7 @@
 package main
 
+import "fmt"
+
 func main() {
-	println("Hello, World!")
+	fmt.Println("Hello, World!")
 }
`
	filePath := "main.go"

	// Create review request
	request := &llm.ReviewRequest{
		FilePath:    filePath,
		FileContent: fileContent,
		FileDiff:    fileDiff,
		Options: llm.ReviewOptions{
			AdditionalInstructions: "Focus on Go best practices",
			IncludeExplanations:    true,
		},
	}

	// Test default provider
	t.Run("Default provider", func(t *testing.T) {
		prompt := CreatePrompt(request, ProviderDefault)
		
		// Check that prompt includes all necessary components
		if !strings.Contains(prompt, fileContent) {
			t.Errorf("Prompt does not contain file content")
		}
		
		if !strings.Contains(prompt, fileDiff) {
			t.Errorf("Prompt does not contain file diff")
		}
		
		if !strings.Contains(prompt, filePath) {
			t.Errorf("Prompt does not contain file path")
		}
		
		if !strings.Contains(prompt, "Focus on Go best practices") {
			t.Errorf("Prompt does not contain additional instructions")
		}
		
		// Check that prompt includes required instructions
		if !strings.Contains(prompt, "Potential bugs or logic errors") {
			t.Errorf("Prompt does not contain instruction for bugs/errors")
		}
		
		if !strings.Contains(prompt, "Code style issues") {
			t.Errorf("Prompt does not contain instruction for code style")
		}
		
		if !strings.Contains(prompt, "Performance or efficiency concerns") {
			t.Errorf("Prompt does not contain instruction for performance")
		}
		
		// Check formatting markers
		if !strings.Contains(prompt, "FULL FILE CONTENT:") {
			t.Errorf("Prompt does not contain file content marker")
		}
		
		if !strings.Contains(prompt, "DIFF (CHANGES MADE):") {
			t.Errorf("Prompt does not contain diff marker")
		}
	})

	// Test OpenAI provider
	t.Run("OpenAI provider", func(t *testing.T) {
		prompt := CreatePrompt(request, ProviderOpenAI)
		
		// Check OpenAI-specific formatting
		if !strings.Contains(prompt, "You are a code review assistant") {
			t.Errorf("OpenAI prompt does not contain appropriate system message")
		}
		
		// Check that JSON format is mentioned
		if !strings.Contains(prompt, "JSON format") {
			t.Errorf("OpenAI prompt does not mention JSON format")
		}
	})

	// Test Anthropic provider
	t.Run("Anthropic provider", func(t *testing.T) {
		prompt := CreatePrompt(request, ProviderAnthropic)
		
		// Check Anthropic-specific formatting
		if !strings.Contains(prompt, "Human:") {
			t.Errorf("Anthropic prompt does not contain Human: marker")
		}
		
		// Check that JSON format is mentioned
		if !strings.Contains(prompt, "JSON format") {
			t.Errorf("Anthropic prompt does not mention JSON format")
		}
	})
}

func TestFormatCodeForProvider(t *testing.T) {
	code := "func main() {\n\tfmt.Println(\"Hello, World!\")\n}"
	
	// Test default formatting
	defaultFormatted := FormatCodeForProvider(code, ProviderDefault)
	if !strings.Contains(defaultFormatted, "```") {
		t.Errorf("Default code formatting does not use code blocks")
	}
	
	// Test OpenAI formatting
	openaiFormatted := FormatCodeForProvider(code, ProviderOpenAI)
	if !strings.Contains(openaiFormatted, "```") {
		t.Errorf("OpenAI code formatting does not use code blocks")
	}
	
	// Test Anthropic formatting
	anthropicFormatted := FormatCodeForProvider(code, ProviderAnthropic)
	if !strings.Contains(anthropicFormatted, "```") {
		t.Errorf("Anthropic code formatting does not use code blocks")
	}
}

func TestFormatDiffForProvider(t *testing.T) {
	diff := `diff --git a/main.go b/main.go
index 1234567..abcdefg 100644
--- a/main.go
+++ b/main.go
@@ -1,5 +1,7 @@
 package main
 
+import "fmt"
+
 func main() {
-	println("Hello, World!")
+	fmt.Println("Hello, World!")
 }
`
	
	// Test default formatting
	defaultFormatted := FormatDiffForProvider(diff, ProviderDefault)
	if !strings.Contains(defaultFormatted, "```diff") {
		t.Errorf("Default diff formatting does not use diff code blocks")
	}
	
	// Test OpenAI formatting
	openaiFormatted := FormatDiffForProvider(diff, ProviderOpenAI)
	if !strings.Contains(openaiFormatted, "```diff") {
		t.Errorf("OpenAI diff formatting does not use diff code blocks")
	}
	
	// Test Anthropic formatting
	anthropicFormatted := FormatDiffForProvider(diff, ProviderAnthropic)
	if !strings.Contains(anthropicFormatted, "```diff") {
		t.Errorf("Anthropic diff formatting does not use diff code blocks")
	}
}
