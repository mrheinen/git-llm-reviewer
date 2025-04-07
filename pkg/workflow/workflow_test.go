package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/git"
	"github.com/niels/git-llm-review/pkg/llm"
)

// MockRepositoryDetector is a mock implementation of the git.RepositoryDetector interface
type MockRepositoryDetector struct {
	isGitRepositoryFunc     func(string) (bool, error)
	getRepositoryRootFunc   func(string) (string, error)
	getStagedFilesFunc      func(string, *config.Config) ([]git.StagedFile, error)
	getAllChangedFilesFunc  func(string, *config.Config) ([]git.ChangedFile, error)
	getUnifiedChangedFilesFunc func(string, *config.Config) ([]git.UnifiedChangedFile, error)
	getFileDiffFunc         func(string, string, bool) (string, error)
	getUnifiedFileDiffFunc  func(string, git.UnifiedChangedFile) (string, error)
	getDiffWithOptionsFunc  func(string, string, bool, git.DiffOptions) (string, error)
	getFileContentFunc      func(string, string) (string, error)
}

func (m *MockRepositoryDetector) IsGitRepository(path string) (bool, error) {
	return m.isGitRepositoryFunc(path)
}

func (m *MockRepositoryDetector) GetRepositoryRoot(path string) (string, error) {
	return m.getRepositoryRootFunc(path)
}

func (m *MockRepositoryDetector) GetStagedFiles(repoRoot string, cfg *config.Config) ([]git.StagedFile, error) {
	return m.getStagedFilesFunc(repoRoot, cfg)
}

func (m *MockRepositoryDetector) GetAllChangedFiles(repoRoot string, cfg *config.Config) ([]git.ChangedFile, error) {
	return m.getAllChangedFilesFunc(repoRoot, cfg)
}

func (m *MockRepositoryDetector) GetUnifiedChangedFiles(repoRoot string, cfg *config.Config) ([]git.UnifiedChangedFile, error) {
	return m.getUnifiedChangedFilesFunc(repoRoot, cfg)
}

func (m *MockRepositoryDetector) GetFileDiff(repoRoot string, filePath string, staged bool) (string, error) {
	return m.getFileDiffFunc(repoRoot, filePath, staged)
}

func (m *MockRepositoryDetector) GetUnifiedFileDiff(repoRoot string, file git.UnifiedChangedFile) (string, error) {
	return m.getUnifiedFileDiffFunc(repoRoot, file)
}

func (m *MockRepositoryDetector) GetDiffWithOptions(repoRoot string, filePath string, staged bool, options git.DiffOptions) (string, error) {
	return m.getDiffWithOptionsFunc(repoRoot, filePath, staged, options)
}

func (m *MockRepositoryDetector) GetFileContent(repoRoot string, filePath string) (string, error) {
	return m.getFileContentFunc(repoRoot, filePath)
}

// MockLLMProvider is a mock implementation of the llm.Provider interface
type MockLLMProvider struct {
	nameFunc          func() string
	validateConfigFunc func() error
	reviewCodeFunc    func(context.Context, *llm.ReviewRequest) (*llm.ReviewResponse, error)
	getCompletionFunc func(string) (string, error)
}

func (m *MockLLMProvider) Name() string {
	return m.nameFunc()
}

func (m *MockLLMProvider) ValidateConfig() error {
	return m.validateConfigFunc()
}

func (m *MockLLMProvider) ReviewCode(ctx context.Context, request *llm.ReviewRequest) (*llm.ReviewResponse, error) {
	return m.reviewCodeFunc(ctx, request)
}

func (m *MockLLMProvider) GetCompletion(prompt string) (string, error) {
	return m.getCompletionFunc(prompt)
}

func TestReviewWorkflow(t *testing.T) {
	// Create a temporary directory for test files
	tempDir, err := os.MkdirTemp("", "git-llm-review-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a test config file
	configPath := filepath.Join(tempDir, "config.yaml")
	configContent := `
llm:
  provider: openai
  api_key: test-api-key
  model: gpt-4
  timeout: 30
concurrency:
  max_tasks: 2
extensions:
  - .go
  - .js
  - .py
`
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Create test config
	testConfig := &config.Config{
		Extensions: []string{".go", ".js", ".py"},
		LLM: config.LLMConfig{
			Provider: "openai",
			APIKey:   "test-api-key",
			Model:    "gpt-4",
			Timeout:  30,
		},
		Concurrency: config.ConcurrencyConfig{
			MaxTasks: 2,
		},
	}

	// Create mock repository detector
	mockRepoDetector := &MockRepositoryDetector{
		isGitRepositoryFunc: func(path string) (bool, error) {
			return true, nil
		},
		getRepositoryRootFunc: func(path string) (string, error) {
			return "/mock/repo/root", nil
		},
		getStagedFilesFunc: func(repoRoot string, cfg *config.Config) ([]git.StagedFile, error) {
			return []git.StagedFile{
				{Path: "file1.go", Status: "M"},
				{Path: "file2.js", Status: "A"},
			}, nil
		},
		getAllChangedFilesFunc: func(repoRoot string, cfg *config.Config) ([]git.ChangedFile, error) {
			return []git.ChangedFile{
				{Path: "file1.go", Status: "M", Staged: true},
				{Path: "file2.js", Status: "A", Staged: true},
				{Path: "file3.py", Status: "M", Staged: false},
			}, nil
		},
		getUnifiedChangedFilesFunc: func(repoRoot string, cfg *config.Config) ([]git.UnifiedChangedFile, error) {
			return []git.UnifiedChangedFile{
				{Path: "file1.go", StagedStatus: "M", UnstagedStatus: ""},
				{Path: "file2.js", StagedStatus: "A", UnstagedStatus: ""},
				{Path: "file3.py", StagedStatus: "", UnstagedStatus: "M"},
			}, nil
		},
		getFileDiffFunc: func(repoRoot string, filePath string, staged bool) (string, error) {
			return "mock diff for " + filePath, nil
		},
		getUnifiedFileDiffFunc: func(repoRoot string, file git.UnifiedChangedFile) (string, error) {
			return "mock unified diff for " + file.Path, nil
		},
		getDiffWithOptionsFunc: func(repoRoot string, filePath string, staged bool, options git.DiffOptions) (string, error) {
			return "mock diff with options for " + filePath, nil
		},
		getFileContentFunc: func(repoRoot string, filePath string) (string, error) {
			return "mock file content for " + filePath, nil
		},
	}

	// Create mock LLM provider
	mockLLMProvider := &MockLLMProvider{
		nameFunc: func() string {
			return "mock-provider"
		},
		validateConfigFunc: func() error {
			return nil
		},
		reviewCodeFunc: func(ctx context.Context, request *llm.ReviewRequest) (*llm.ReviewResponse, error) {
			return &llm.ReviewResponse{
				Review: `{
					"issues": [
						{
							"title": "Mock bug issue",
							"explanation": "This is a mock bug issue for testing",
							"diff": "mock diff"
						},
						{
							"title": "Mock style issue",
							"explanation": "This is a mock style issue for testing",
							"diff": "mock diff"
						}
					]
				}`,
				Confidence: 0.95,
				Metadata: map[string]interface{}{
					"model": "mock-model",
				},
			}, nil
		},
		getCompletionFunc: func(prompt string) (string, error) {
			// Return a mock JSON response
			return `{
				"issues": [
					{
						"title": "Mock bug issue",
						"explanation": "This is a mock bug issue for testing",
						"diff": "mock diff"
					},
					{
						"title": "Mock style issue",
						"explanation": "This is a mock style issue for testing",
						"diff": "mock diff"
					}
				]
			}`, nil
		},
	}

	t.Run("Process staged files", func(t *testing.T) {
		// Create workflow options
		options := Options{
			ConfigPath:    configPath,
			All:           false,
			OutputFormat:  "terminal",
			OutputPath:    "",
			ProviderName:  "openai",
			VerboseOutput: false,
		}

		// Create the workflow
		workflow, err := NewReviewWorkflow(options)
		if err != nil {
			t.Fatalf("Failed to create workflow: %v", err)
		}

		// Override dependencies with mocks
		workflow.repoDetector = mockRepoDetector
		workflow.provider = mockLLMProvider
		workflow.config = testConfig

		// Run the workflow
		stats, err := workflow.Run(context.Background())
		if err != nil {
			t.Fatalf("Workflow failed: %v", err)
		}
		
		// Verify statistics
		if stats.FilesProcessed != 2 {
			t.Errorf("Expected 2 files processed, got %d", stats.FilesProcessed)
		}
	})

	t.Run("Process all changed files", func(t *testing.T) {
		// Create workflow options
		options := Options{
			ConfigPath:    configPath,
			All:           true,
			OutputFormat:  "terminal",
			OutputPath:    "",
			ProviderName:  "openai",
			VerboseOutput: false,
		}

		// Create the workflow
		workflow, err := NewReviewWorkflow(options)
		if err != nil {
			t.Fatalf("Failed to create workflow: %v", err)
		}

		// Override dependencies with mocks
		workflow.repoDetector = mockRepoDetector
		workflow.provider = mockLLMProvider
		workflow.config = testConfig

		// Run the workflow
		stats, err := workflow.Run(context.Background())
		if err != nil {
			t.Fatalf("Workflow failed: %v", err)
		}
		
		// Verify statistics
		if stats.FilesProcessed != 3 {
			t.Errorf("Expected 3 files processed, got %d", stats.FilesProcessed)
		}
	})

	t.Run("Generate markdown report", func(t *testing.T) {
		// Create output directory
		outputDir := filepath.Join(tempDir, "reports")
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			t.Fatalf("Failed to create output directory: %v", err)
		}

		// Create workflow options
		options := Options{
			ConfigPath:    configPath,
			All:           false,
			OutputFormat:  "markdown",
			OutputPath:    outputDir,
			ProviderName:  "openai",
			VerboseOutput: false,
		}

		// Create the workflow
		workflow, err := NewReviewWorkflow(options)
		if err != nil {
			t.Fatalf("Failed to create workflow: %v", err)
		}

		// Override dependencies with mocks
		workflow.repoDetector = mockRepoDetector
		workflow.provider = mockLLMProvider
		workflow.config = testConfig

		// Run the workflow
		stats, err := workflow.Run(context.Background())
		if err != nil {
			t.Fatalf("Workflow failed: %v", err)
		}
		
		// Verify statistics
		if stats.FilesProcessed != 2 {
			t.Errorf("Expected 2 files processed, got %d", stats.FilesProcessed)
		}
		
		// Verify report files were created (one for each processed file)
		// Since we have multiple files, we should have separate reports for each file
		reportFile1 := filepath.Join(outputDir, "review_file1.go.md")
		reportFile2 := filepath.Join(outputDir, "review_file2.js.md")
		summaryFile := filepath.Join(outputDir, "summary.md")
		
		if _, err := os.Stat(reportFile1); os.IsNotExist(err) {
			t.Errorf("Report file was not created: %s", reportFile1)
		}
		
		if _, err := os.Stat(reportFile2); os.IsNotExist(err) {
			t.Errorf("Report file was not created: %s", reportFile2)
		}
		
		if _, err := os.Stat(summaryFile); os.IsNotExist(err) {
			t.Errorf("Summary report file was not created: %s", summaryFile)
		}
	})
}
