package git

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/niels/git-llm-review/pkg/config"
)

// mockCommandRunner is used to mock exec.Command for testing
type mockCommandRunner struct {
	outputFunc func(string, ...string) ([]byte, error)
}

func (m *mockCommandRunner) runCommand(name string, args ...string) ([]byte, error) {
	return m.outputFunc(name, args...)
}

// TestIsGitRepository tests the IsGitRepository function
func TestIsGitRepository(t *testing.T) {
	// Test case 1: Valid Git repository
	mockRunner := &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[0] == "-C" && args[2] == "rev-parse" && args[3] == "--is-inside-work-tree" {
				return []byte("true"), nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector := &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	isRepo, err := detector.IsGitRepository(".")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !isRepo {
		t.Error("Expected directory to be a Git repository")
	}

	// Test case 2: Not a Git repository
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[0] == "-C" && args[2] == "rev-parse" && args[3] == "--is-inside-work-tree" {
				return nil, errors.New("fatal: not a git repository")
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	isRepo, err = detector.IsGitRepository(".")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if isRepo {
		t.Error("Expected directory not to be a Git repository")
	}

	// Test case 3: Git command not found
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			return nil, exec.ErrNotFound
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	isRepo, err = detector.IsGitRepository(".")
	if err == nil {
		t.Error("Expected error when Git is not installed")
	}
	if isRepo {
		t.Error("Expected directory not to be a Git repository when Git is not installed")
	}
}

// TestGetRepositoryRoot tests the GetRepositoryRoot function
func TestGetRepositoryRoot(t *testing.T) {
	// Test case 1: Valid Git repository
	mockRunner := &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[0] == "-C" && args[2] == "rev-parse" {
				if args[3] == "--is-inside-work-tree" {
					return []byte("true"), nil
				}
				if args[3] == "--show-toplevel" {
					return []byte("/path/to/repo"), nil
				}
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector := &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	root, err := detector.GetRepositoryRoot(".")
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if root != "/path/to/repo" {
		t.Errorf("Expected repository root to be '/path/to/repo', got: %s", root)
	}

	// Test case 2: Not a Git repository
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[0] == "-C" && args[2] == "rev-parse" {
				return nil, errors.New("fatal: not a git repository")
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	root, err = detector.GetRepositoryRoot(".")
	if err == nil {
		t.Error("Expected error when not in a Git repository")
	}
	if !strings.Contains(err.Error(), "not a Git repository") {
		t.Errorf("Expected error message to contain 'not a Git repository', got: %v", err)
	}
	if root != "" {
		t.Errorf("Expected empty repository root, got: %s", root)
	}

	// Test case 3: Git command not found
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			return nil, exec.ErrNotFound
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	root, err = detector.GetRepositoryRoot(".")
	if err == nil {
		t.Error("Expected error when Git is not installed")
	}
	if !strings.Contains(err.Error(), "Git executable not found") {
		t.Errorf("Expected error message to contain 'Git executable not found', got: %v", err)
	}
	if root != "" {
		t.Errorf("Expected empty repository root, got: %s", root)
	}
}

// TestIntegration performs integration tests with a real temporary directory
func TestIntegration(t *testing.T) {
	// Skip if git is not installed
	_, err := exec.LookPath("git")
	if err != nil {
		t.Skip("Git not installed, skipping integration test")
	}

	// Create a temporary directory
	tempDir, err := os.MkdirTemp("", "git-llm-review-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a nested directory structure
	nestedDir := filepath.Join(tempDir, "level1", "level2")
	err = os.MkdirAll(nestedDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create nested directory: %v", err)
	}

	// Change to the temp directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)
	
	// Initialize a Git repository in the temp directory
	cmd := exec.Command("git", "init")
	cmd.Dir = tempDir
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to initialize Git repository: %v", err)
	}

	// Configure Git user for the test repository
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tempDir
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to configure Git user name: %v", err)
	}

	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = tempDir
	err = cmd.Run()
	if err != nil {
		t.Fatalf("Failed to configure Git user email: %v", err)
	}

	// Create a real detector
	detector := NewRepositoryDetector()

	// Test IsGitRepository from the repository root
	isRepo, err := detector.IsGitRepository(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !isRepo {
		t.Error("Expected tempDir to be a Git repository")
	}

	// Test IsGitRepository from a nested directory
	isRepo, err = detector.IsGitRepository(nestedDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if !isRepo {
		t.Error("Expected nestedDir to be in a Git repository")
	}

	// Test GetRepositoryRoot from the repository root
	root, err := detector.GetRepositoryRoot(tempDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	// Compare canonical paths to handle any symlinks or relative paths
	expectedRoot, err := filepath.EvalSymlinks(tempDir)
	if err != nil {
		t.Fatalf("Failed to get canonical path for tempDir: %v", err)
	}
	actualRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatalf("Failed to get canonical path for root: %v", err)
	}
	
	if actualRoot != expectedRoot {
		t.Errorf("Expected repository root to be '%s', got: '%s'", expectedRoot, actualRoot)
	}

	// Test GetRepositoryRoot from a nested directory
	root, err = detector.GetRepositoryRoot(nestedDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if actualRoot != expectedRoot {
		t.Errorf("Expected repository root to be '%s', got: '%s'", expectedRoot, actualRoot)
	}

	// Test with a non-Git directory
	nonGitDir, err := os.MkdirTemp("", "non-git-dir")
	if err != nil {
		t.Fatalf("Failed to create non-Git directory: %v", err)
	}
	defer os.RemoveAll(nonGitDir)

	isRepo, err = detector.IsGitRepository(nonGitDir)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	if isRepo {
		t.Error("Expected nonGitDir not to be a Git repository")
	}

	root, err = detector.GetRepositoryRoot(nonGitDir)
	if err == nil {
		t.Error("Expected error when not in a Git repository")
	}
	if !strings.Contains(err.Error(), "not a Git repository") {
		t.Errorf("Expected error message to contain 'not a Git repository', got: %v", err)
	}
}

// TestGetStagedFiles tests the GetStagedFiles function
func TestGetStagedFiles(t *testing.T) {
	// Sample git diff --cached output
	sampleDiffOutput := `A       file1.go
M       file2.go
D       file3.go
R100    old_file.cc -> new_file.cc
A       path/to/file4.go
M       path/to/file5.vue
A       path/to/file6.txt
A       path/to/file7.proto
`

	// Test case 1: Valid staged files with extension filtering
	mockRunner := &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "diff" && args[3] == "--cached" {
				return []byte(sampleDiffOutput), nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector := &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	// Create a configuration with specific extensions
	cfg := &config.Config{
		Extensions: []string{".go", ".cc", ".proto"},
	}

	files, err := detector.GetStagedFiles(".", cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that only files with the specified extensions are included
	expectedFiles := []StagedFile{
		{Path: "file1.go", Status: "A"},
		{Path: "file2.go", Status: "M"},
		{Path: "new_file.cc", Status: "R"},
		{Path: "path/to/file4.go", Status: "A"},
		{Path: "path/to/file7.proto", Status: "A"},
	}

	// Sort both slices for consistent comparison
	sortStagedFiles(files)
	sortStagedFiles(expectedFiles)

	if !reflect.DeepEqual(files, expectedFiles) {
		t.Errorf("Expected files %v, got: %v", expectedFiles, files)
	}

	// Test case 2: No staged files
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "diff" && args[3] == "--cached" {
				return []byte(""), nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	files, err = detector.GetStagedFiles(".", cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected empty file list, got: %v", files)
	}

	// Test case 3: Git command error
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "diff" && args[3] == "--cached" {
				return nil, errors.New("git command failed")
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	_, err = detector.GetStagedFiles(".", cfg)
	if err == nil {
		t.Error("Expected error when Git command fails")
	}

	// Test case 4: Git not installed
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			return nil, exec.ErrNotFound
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	_, err = detector.GetStagedFiles(".", cfg)
	if err == nil {
		t.Error("Expected error when Git is not installed")
	}
	if !strings.Contains(err.Error(), "Git executable not found") {
		t.Errorf("Expected error message to contain 'Git executable not found', got: %v", err)
	}
}

// TestGetAllChangedFiles tests the GetAllChangedFiles function
func TestGetAllChangedFiles(t *testing.T) {
	// Sample git status --porcelain output
	sampleStatusOutput := ` M file1.go
M  file2.go
A  file3.go
?? file4.go
 D file5.cc
D  file6.cc
R  old_file.vue -> new_file.vue
MM file7.proto
`

	// Test case 1: Valid changed files with extension filtering
	mockRunner := &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "status" && args[3] == "--porcelain" {
				return []byte(sampleStatusOutput), nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector := &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	// Create a configuration with specific extensions
	cfg := &config.Config{
		Extensions: []string{".go", ".cc", ".proto"},
	}

	files, err := detector.GetAllChangedFiles(".", cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that only files with the specified extensions are included
	expectedFiles := []ChangedFile{
		{Path: "file1.go", Status: "M", Staged: false},
		{Path: "file2.go", Status: "M", Staged: true},
		{Path: "file3.go", Status: "A", Staged: true},
		{Path: "file4.go", Status: "??", Staged: false},
		{Path: "file5.cc", Status: "D", Staged: false},
		{Path: "file6.cc", Status: "D", Staged: true},
		{Path: "file7.proto", Status: "MM", Staged: true},
	}

	// Sort both slices for consistent comparison
	sortChangedFiles(files)
	sortChangedFiles(expectedFiles)

	if !reflect.DeepEqual(files, expectedFiles) {
		t.Errorf("Expected files %v, got: %v", expectedFiles, files)
	}

	// Test case 2: No changed files
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "status" && args[3] == "--porcelain" {
				return []byte(""), nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	files, err = detector.GetAllChangedFiles(".", cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected empty file list, got: %v", files)
	}

	// Test case 3: Git command error
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "status" && args[3] == "--porcelain" {
				return nil, errors.New("git command failed")
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	_, err = detector.GetAllChangedFiles(".", cfg)
	if err == nil {
		t.Error("Expected error when Git command fails")
	}

	// Test case 4: Git not installed
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			return nil, exec.ErrNotFound
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	_, err = detector.GetAllChangedFiles(".", cfg)
	if err == nil {
		t.Error("Expected error when Git is not installed")
	}
	if !strings.Contains(err.Error(), "Git executable not found") {
		t.Errorf("Expected error message to contain 'Git executable not found', got: %v", err)
	}
}

// TestGetUnifiedChangedFiles tests the GetUnifiedChangedFiles function
func TestGetUnifiedChangedFiles(t *testing.T) {
	// Sample git status --porcelain output with duplicate files and mixed statuses
	sampleStatusOutput := ` M file1.go
M  file1.go
MM file2.go
A  file3.go
?? file4.go
 D file5.cc
D  file6.cc
R  old_file.vue -> new_file.vue
`

	// Test case 1: Unified changed files with extension filtering
	mockRunner := &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "status" && args[3] == "--porcelain" {
				return []byte(sampleStatusOutput), nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector := &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	// Create a configuration with specific extensions
	cfg := &config.Config{
		Extensions: []string{".go", ".cc"},
	}

	files, err := detector.GetUnifiedChangedFiles(".", cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that only files with the specified extensions are included
	// and that duplicates are handled correctly
	expectedFiles := []UnifiedChangedFile{
		{Path: "file1.go", StagedStatus: "M", UnstagedStatus: "M"},
		{Path: "file2.go", StagedStatus: "M", UnstagedStatus: "M"},
		{Path: "file3.go", StagedStatus: "A", UnstagedStatus: ""},
		{Path: "file4.go", StagedStatus: "", UnstagedStatus: "??"},
		{Path: "file5.cc", StagedStatus: "", UnstagedStatus: "D"},
		{Path: "file6.cc", StagedStatus: "D", UnstagedStatus: ""},
	}

	// Sort both slices for consistent comparison
	sortUnifiedChangedFiles(files)
	sortUnifiedChangedFiles(expectedFiles)

	if !reflect.DeepEqual(files, expectedFiles) {
		t.Errorf("Expected files %v, got: %v", expectedFiles, files)
	}

	// Test case 2: No changed files
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "status" && args[3] == "--porcelain" {
				return []byte(""), nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	files, err = detector.GetUnifiedChangedFiles(".", cfg)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if len(files) != 0 {
		t.Errorf("Expected empty file list, got: %v", files)
	}

	// Test case 3: Git command error
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "status" && args[3] == "--porcelain" {
				return nil, errors.New("git command failed")
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	_, err = detector.GetUnifiedChangedFiles(".", cfg)
	if err == nil {
		t.Error("Expected error when Git command fails")
	}
}

// sortStagedFiles sorts a slice of StagedFile by Path
func sortStagedFiles(files []StagedFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

// sortChangedFiles sorts a slice of ChangedFile by Path
func sortChangedFiles(files []ChangedFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}

// sortUnifiedChangedFiles sorts a slice of UnifiedChangedFile by Path
func sortUnifiedChangedFiles(files []UnifiedChangedFile) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].Path < files[j].Path
	})
}
