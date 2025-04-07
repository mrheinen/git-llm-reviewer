package git

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/niels/git-llm-review/pkg/config"
)

// Common errors
var (
	ErrNotGitRepository = errors.New("not a Git repository")
	ErrGitNotInstalled  = errors.New("Git executable not found")
)

// StagedFile represents a staged file in the Git repository
type StagedFile struct {
	Path   string // Path relative to repository root
	Status string // Git status (A: added, M: modified, D: deleted, R: renamed)
}

// ChangedFile represents a changed file in the Git repository
type ChangedFile struct {
	Path   string // Path relative to repository root
	Status string // Git status (A: added, M: modified, D: deleted, R: renamed, ??: untracked)
	Staged bool   // Whether the file is staged
}

// UnifiedChangedFile represents a file that may have both staged and unstaged changes
type UnifiedChangedFile struct {
	Path           string // Path relative to repository root
	StagedStatus   string // Status of staged changes (A: added, M: modified, D: deleted, R: renamed)
	UnstagedStatus string // Status of unstaged changes (M: modified, D: deleted, ??: untracked)
}

// RepositoryDetector defines the interface for Git repository detection
type RepositoryDetector interface {
	// IsGitRepository checks if the given directory is within a Git repository
	IsGitRepository(dir string) (bool, error)
	// GetRepositoryRoot returns the root directory of the Git repository
	GetRepositoryRoot(dir string) (string, error)
	// GetStagedFiles returns a list of staged files in the Git repository
	GetStagedFiles(dir string, cfg *config.Config) ([]StagedFile, error)
	// GetAllChangedFiles returns a list of all changed files in the Git repository
	GetAllChangedFiles(dir string, cfg *config.Config) ([]ChangedFile, error)
	// GetUnifiedChangedFiles returns a unified list of all changed files
	GetUnifiedChangedFiles(dir string, cfg *config.Config) ([]UnifiedChangedFile, error)
	// GetFileDiff returns the diff for a specific file
	GetFileDiff(dir string, filePath string, staged bool) (string, error)
	// GetUnifiedFileDiff returns a unified diff for a file that may have both staged and unstaged changes
	GetUnifiedFileDiff(dir string, file UnifiedChangedFile) (string, error)
	// GetDiffWithOptions returns the diff for a specific file with custom options
	GetDiffWithOptions(dir string, filePath string, staged bool, options DiffOptions) (string, error)
	// GetFileContent returns the content of a file from the repository
	GetFileContent(dir string, filePath string) (string, error)
}

// CommandRunner is an interface for running commands
type CommandRunner interface {
	runCommand(name string, args ...string) ([]byte, error)
}

// RealCommandRunner implements CommandRunner using os/exec
type RealCommandRunner struct{}

// runCommand executes a command and returns its output
func (r *RealCommandRunner) runCommand(name string, args ...string) ([]byte, error) {
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	return output, err
}

// RepositoryDetectorImpl provides methods for Git repository detection
type RepositoryDetectorImpl struct {
	cmdRunner CommandRunner
}

// NewRepositoryDetector creates a new RepositoryDetector with a real command runner
func NewRepositoryDetector() *RepositoryDetectorImpl {
	return &RepositoryDetectorImpl{
		cmdRunner: &RealCommandRunner{},
	}
}

// IsGitRepository checks if the given directory is within a Git repository
func (d *RepositoryDetectorImpl) IsGitRepository(dir string) (bool, error) {
	// Run git rev-parse --is-inside-work-tree
	output, err := d.cmdRunner.runCommand("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	
	// Handle errors
	if err != nil {
		// Check if Git is not installed
		if errors.Is(err, exec.ErrNotFound) {
			return false, ErrGitNotInstalled
		}
		
		// If the error contains "not a git repository", it's not a Git repository
		if strings.Contains(string(output), "not a git repository") || 
		   strings.Contains(err.Error(), "not a git repository") {
			return false, nil
		}
		
		// For other errors, return the error
		return false, fmt.Errorf("failed to check if directory is a Git repository: %w", err)
	}
	
	// Check the output
	return bytes.Equal(bytes.TrimSpace(output), []byte("true")), nil
}

// GetRepositoryRoot returns the root directory of the Git repository
func (d *RepositoryDetectorImpl) GetRepositoryRoot(dir string) (string, error) {
	// First check if we're in a Git repository
	isRepo, err := d.IsGitRepository(dir)
	if err != nil {
		return "", err
	}
	
	if !isRepo {
		return "", ErrNotGitRepository
	}
	
	// Get the repository root
	output, err := d.cmdRunner.runCommand("git", "-C", dir, "rev-parse", "--show-toplevel")
	if err != nil {
		// Check if Git is not installed
		if errors.Is(err, exec.ErrNotFound) {
			return "", ErrGitNotInstalled
		}
		
		return "", fmt.Errorf("failed to get repository root: %w", err)
	}
	
	// Return the repository root
	return strings.TrimSpace(string(output)), nil
}

// GetStagedFiles returns a list of staged files in the Git repository
// filtered by the configured extensions
func (d *RepositoryDetectorImpl) GetStagedFiles(dir string, cfg *config.Config) ([]StagedFile, error) {
	// Check if Git is installed
	output, err := d.cmdRunner.runCommand("git", "-C", dir, "diff", "--cached", "--name-status")
	if err != nil {
		// Check if Git is not installed
		if errors.Is(err, exec.ErrNotFound) {
			return nil, ErrGitNotInstalled
		}
		
		return nil, fmt.Errorf("failed to get staged files: %w", err)
	}
	
	// Parse the output
	return d.parseStagedFiles(string(output), cfg.Extensions), nil
}

// GetAllChangedFiles returns a list of all changed files in the Git repository
// (both staged and unstaged) filtered by the configured extensions
func (d *RepositoryDetectorImpl) GetAllChangedFiles(dir string, cfg *config.Config) ([]ChangedFile, error) {
	// Check if Git is installed
	output, err := d.cmdRunner.runCommand("git", "-C", dir, "status", "--porcelain")
	if err != nil {
		// Check if Git is not installed
		if errors.Is(err, exec.ErrNotFound) {
			return nil, ErrGitNotInstalled
		}
		
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}
	
	// Parse the output
	return d.parseChangedFiles(string(output), cfg.Extensions), nil
}

// GetUnifiedChangedFiles returns a unified list of all changed files in the Git repository
// with no duplicates, handling files with both staged and unstaged changes
func (d *RepositoryDetectorImpl) GetUnifiedChangedFiles(dir string, cfg *config.Config) ([]UnifiedChangedFile, error) {
	// Check if Git is installed
	output, err := d.cmdRunner.runCommand("git", "-C", dir, "status", "--porcelain")
	if err != nil {
		// Check if Git is not installed
		if errors.Is(err, exec.ErrNotFound) {
			return nil, ErrGitNotInstalled
		}
		
		return nil, fmt.Errorf("failed to get changed files: %w", err)
	}
	
	// Parse the output
	return d.parseUnifiedChangedFiles(string(output), cfg.Extensions), nil
}

// GetFileContent returns the content of a file from the repository
func (d *RepositoryDetectorImpl) GetFileContent(dir string, filePath string) (string, error) {
	// Check if Git is installed
	output, err := d.cmdRunner.runCommand("git", "-C", dir, "show", ":./"+filePath)
	if err != nil {
		// Check if Git is not installed
		if errors.Is(err, exec.ErrNotFound) {
			return "", ErrGitNotInstalled
		}
		
		return "", fmt.Errorf("failed to get file content: %w", err)
	}
	
	// Return the file content
	return string(output), nil
}

// parseStagedFiles parses the output of git diff --cached --name-status
// and returns a list of staged files filtered by the given extensions
func (d *RepositoryDetectorImpl) parseStagedFiles(output string, extensions []string) []StagedFile {
	var files []StagedFile
	
	// Split the output into lines
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		
		// Handle renamed files with arrow notation (R100    old_file.cc -> new_file.cc)
		if strings.HasPrefix(line, "R") && strings.Contains(line, " -> ") {
			parts := strings.Split(line, " -> ")
			if len(parts) != 2 {
				continue
			}
			
			// Extract status and new path
			statusAndOldPath := strings.Fields(parts[0])
			if len(statusAndOldPath) < 2 {
				continue
			}
			
			// We don't need the status here, just using it to check the prefix
			newPath := strings.TrimSpace(parts[1])
			
			// Check if the file has one of the configured extensions
			if !hasExtension(newPath, extensions) {
				continue
			}
			
			// Add the file to the list with status "R"
			files = append(files, StagedFile{
				Path:   newPath,
				Status: "R",
			})
			continue
		}
		
		// Parse the line for normal cases
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		
		status := parts[0]
		path := parts[1]
		
		// Handle renamed files with separate fields
		if strings.HasPrefix(status, "R") && len(parts) >= 3 {
			// For renamed files, use the new name
			path = parts[2]
		}
		
		// Skip deleted files
		if status == "D" {
			continue
		}
		
		// Check if the file has one of the configured extensions
		if !hasExtension(path, extensions) {
			continue
		}
		
		// Add the file to the list
		files = append(files, StagedFile{
			Path:   path,
			Status: string(status[0]), // Use only the first character of the status
		})
	}
	
	return files
}

// parseChangedFiles parses the output of git status --porcelain
// and returns a list of changed files filtered by the given extensions
func (d *RepositoryDetectorImpl) parseChangedFiles(output string, extensions []string) []ChangedFile {
	var files []ChangedFile
	
	// Split the output into lines
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		// Parse the line
		if len(line) < 3 {
			continue
		}
		
		// First two characters are the status
		statusX := line[0]
		statusY := line[1]
		path := strings.TrimSpace(line[2:])
		
		// Handle renamed files
		if (statusX == 'R' || statusY == 'R') && strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			if len(parts) == 2 {
				path = parts[1] // Use the new name
			}
		}
		
		// Check if the file has one of the configured extensions
		if !hasExtension(path, extensions) {
			continue
		}
		
		// Determine the status and whether the file is staged
		status := string(statusX) + string(statusY)
		staged := false
		
		// If the status in the index (X) is not a space, the file is staged
		if statusX != ' ' && statusX != '?' {
			staged = true
		}
		
		// Simplify the status for common cases
		if statusX == ' ' && statusY != ' ' {
			status = string(statusY)
		} else if statusX != ' ' && statusY == ' ' {
			status = string(statusX)
		} else if statusX == '?' && statusY == '?' {
			status = "??"
		}
		
		// Add the file to the list
		files = append(files, ChangedFile{
			Path:   path,
			Status: status,
			Staged: staged,
		})
	}
	
	return files
}

// parseUnifiedChangedFiles parses the output of git status --porcelain
// and returns a unified list of changed files with no duplicates
func (d *RepositoryDetectorImpl) parseUnifiedChangedFiles(output string, extensions []string) []UnifiedChangedFile {
	// Use a map to avoid duplicates
	fileMap := make(map[string]*UnifiedChangedFile)
	
	// Split the output into lines
	lines := strings.Split(output, "\n")
	
	for _, line := range lines {
		if line == "" {
			continue
		}
		
		// Parse the line
		if len(line) < 3 {
			continue
		}
		
		// First two characters are the status
		statusX := line[0] // Index status (staged)
		statusY := line[1] // Working tree status (unstaged)
		path := strings.TrimSpace(line[2:])
		
		// Handle renamed files
		if (statusX == 'R' || statusY == 'R') && strings.Contains(path, " -> ") {
			parts := strings.Split(path, " -> ")
			if len(parts) == 2 {
				path = parts[1] // Use the new name
			}
		}
		
		// Check if the file has one of the configured extensions
		if !hasExtension(path, extensions) {
			continue
		}
		
		// Get or create the unified file entry
		unifiedFile, exists := fileMap[path]
		if !exists {
			unifiedFile = &UnifiedChangedFile{
				Path: path,
			}
			fileMap[path] = unifiedFile
		}
		
		// Update staged status
		if statusX != ' ' && statusX != '?' {
			unifiedFile.StagedStatus = string(statusX)
		}
		
		// Update unstaged status
		if statusY != ' ' {
			unifiedFile.UnstagedStatus = string(statusY)
		}
		
		// Special handling for untracked files
		if statusX == '?' && statusY == '?' {
			unifiedFile.UnstagedStatus = "??"
		}
	}
	
	// Convert map to slice
	var files []UnifiedChangedFile
	for _, file := range fileMap {
		files = append(files, *file)
	}
	
	return files
}

// hasExtension checks if the given file path has one of the specified extensions
func hasExtension(path string, extensions []string) bool {
	// If no extensions are specified, include all files
	if len(extensions) == 0 {
		return true
	}
	
	ext := filepath.Ext(path)
	for _, e := range extensions {
		if ext == e {
			return true
		}
	}
	
	return false
}
