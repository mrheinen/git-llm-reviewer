package processor

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/git"
	"github.com/niels/git-llm-review/pkg/llm"
)

// MockRepositoryDetector is a mock implementation of git.RepositoryDetector
type MockRepositoryDetector struct {
	GetFileDiffFunc    func(dir string, filePath string, staged bool) (string, error)
	GetFileContentFunc func(dir string, filePath string) (string, error)
}

func (m *MockRepositoryDetector) IsGitRepository(dir string) (bool, error) {
	return true, nil
}

func (m *MockRepositoryDetector) GetRepositoryRoot(dir string) (string, error) {
	return "/repo/root", nil
}

func (m *MockRepositoryDetector) GetStagedFiles(dir string, cfg *config.Config) ([]git.StagedFile, error) {
	return nil, nil
}

func (m *MockRepositoryDetector) GetAllChangedFiles(dir string, cfg *config.Config) ([]git.ChangedFile, error) {
	return nil, nil
}

func (m *MockRepositoryDetector) GetUnifiedChangedFiles(dir string, cfg *config.Config) ([]git.UnifiedChangedFile, error) {
	return nil, nil
}

func (m *MockRepositoryDetector) GetFileDiff(dir string, filePath string, staged bool) (string, error) {
	if m.GetFileDiffFunc != nil {
		return m.GetFileDiffFunc(dir, filePath, staged)
	}
	return "", nil
}

func (m *MockRepositoryDetector) GetUnifiedFileDiff(dir string, file git.UnifiedChangedFile) (string, error) {
	return "", nil
}

func (m *MockRepositoryDetector) GetDiffWithOptions(dir string, filePath string, staged bool, options git.DiffOptions) (string, error) {
	return "", nil
}

func (m *MockRepositoryDetector) GetFileContent(dir string, filePath string) (string, error) {
	if m.GetFileContentFunc != nil {
		return m.GetFileContentFunc(dir, filePath)
	}
	return "", nil
}

// MockLLMProvider is a mock implementation of llm.Provider
type MockLLMProvider struct {
	GetCompletionFunc func(prompt string) (string, error)
}

func (m *MockLLMProvider) Name() string {
	return "MockProvider"
}

func (m *MockLLMProvider) ValidateConfig() error {
	return nil
}

func (m *MockLLMProvider) ReviewCode(ctx context.Context, request *llm.ReviewRequest) (*llm.ReviewResponse, error) {
	return nil, nil
}

func (m *MockLLMProvider) GetCompletion(prompt string) (string, error) {
	if m.GetCompletionFunc != nil {
		return m.GetCompletionFunc(prompt)
	}
	return "", nil
}

func TestReviewFileProcessor_WithFileContent(t *testing.T) {
	// Create mocks
	mockDetector := &MockRepositoryDetector{
		GetFileDiffFunc: func(dir string, filePath string, staged bool) (string, error) {
			return "diff --git a/test.go b/test.go\nindex 1234567..abcdef 100644\n--- a/test.go\n+++ b/test.go\n@@ -1,3 +1,3 @@\n package test\n-func hello() {}\n+func hello() { println(\"hello\") }\n", nil
		},
		GetFileContentFunc: func(dir string, filePath string) (string, error) {
			return "package test\nfunc hello() { println(\"hello\") }\n", nil
		},
	}

	// Track if file content was included in the prompt
	fileContentIncluded := false

	mockProvider := &MockLLMProvider{
		GetCompletionFunc: func(prompt string) (string, error) {
			// Check if the prompt contains the file content
			if prompt != "" {
				// Check if we have proper file content in the prompt
				fileContentIncluded = !strings.Contains(prompt, "FULL FILE CONTENT (STAGED VERSION):\n```go\n\n```") &&
					strings.Contains(prompt, "package test\nfunc hello() { println(\"hello\") }")
			}
			return "{\"issues\":[]}", nil
		},
	}

	// Create the file processor
	processor := ReviewFileProcessor(
		"/repo/root",
		mockDetector,
		mockProvider,
		"openai",
	)

	// Create a test file info
	fileInfo := FileInfo{
		Path:   "test.go",
		Type:   "staged",
		Status: "M",
	}

	// Process the file
	_, err := processor(context.Background(), fileInfo)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}

	// Assert that file content was included
	if !fileContentIncluded {
		t.Error("File content was not included in the prompt")
	}
}

func TestReviewFileProcessor_MissingFileContent(t *testing.T) {
	// Create mocks
	mockDetector := &MockRepositoryDetector{
		GetFileDiffFunc: func(dir string, filePath string, staged bool) (string, error) {
			return "diff --git a/test.go b/test.go\nindex 1234567..abcdef 100644\n--- a/test.go\n+++ b/test.go\n@@ -1,3 +1,3 @@\n package test\n-func hello() {}\n+func hello() { println(\"hello\") }\n", nil
		},
		GetFileContentFunc: func(dir string, filePath string) (string, error) {
			// Simulate error getting file content
			return "", errors.New("file not found")
		},
	}

	// Create processor that should still work even with file content error
	mockProvider := &MockLLMProvider{
		GetCompletionFunc: func(prompt string) (string, error) {
			return "{\"issues\":[]}", nil
		},
	}

	// Create the file processor
	processor := ReviewFileProcessor(
		"/repo/root",
		mockDetector,
		mockProvider,
		"openai",
	)

	// Create a test file info
	fileInfo := FileInfo{
		Path:   "test.go",
		Type:   "staged",
		Status: "M",
	}

	// Process the file - it should still work even with missing file content
	result, err := processor(context.Background(), fileInfo)
	if err != nil {
		t.Fatalf("Failed to process file: %v", err)
	}

	// Make sure we got a valid result even with missing file content
	if result == nil {
		t.Error("Expected non-nil result even with missing file content")
	}
}
