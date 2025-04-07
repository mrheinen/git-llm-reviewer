package main

import (
	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/git"
)

// MockRepositoryDetector is a mock implementation of the RepositoryDetector interface for testing
type MockRepositoryDetector struct {
	IsGitRepositoryFunc      func(dir string) (bool, error)
	GetRepositoryRootFunc    func(dir string) (string, error)
	GetStagedFilesFunc       func(dir string, cfg *config.Config) ([]git.StagedFile, error)
	GetAllChangedFilesFunc   func(dir string, cfg *config.Config) ([]git.ChangedFile, error)
	GetUnifiedChangedFilesFunc func(dir string, cfg *config.Config) ([]git.UnifiedChangedFile, error)
	GetFileDiffFunc          func(dir string, filePath string, staged bool) (string, error)
	GetUnifiedFileDiffFunc   func(dir string, file git.UnifiedChangedFile) (string, error)
	GetDiffWithOptionsFunc   func(dir string, filePath string, staged bool, options git.DiffOptions) (string, error)
	GetFileContentFunc       func(dir string, filePath string) (string, error)
}

// IsGitRepository implements the RepositoryDetector interface
func (m *MockRepositoryDetector) IsGitRepository(dir string) (bool, error) {
	if m.IsGitRepositoryFunc != nil {
		return m.IsGitRepositoryFunc(dir)
	}
	return true, nil
}

// GetRepositoryRoot implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetRepositoryRoot(dir string) (string, error) {
	if m.GetRepositoryRootFunc != nil {
		return m.GetRepositoryRootFunc(dir)
	}
	return "/mock/repo/root", nil
}

// GetStagedFiles implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetStagedFiles(dir string, cfg *config.Config) ([]git.StagedFile, error) {
	if m.GetStagedFilesFunc != nil {
		return m.GetStagedFilesFunc(dir, cfg)
	}
	return []git.StagedFile{}, nil
}

// GetAllChangedFiles implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetAllChangedFiles(dir string, cfg *config.Config) ([]git.ChangedFile, error) {
	if m.GetAllChangedFilesFunc != nil {
		return m.GetAllChangedFilesFunc(dir, cfg)
	}
	return []git.ChangedFile{}, nil
}

// GetUnifiedChangedFiles implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetUnifiedChangedFiles(dir string, cfg *config.Config) ([]git.UnifiedChangedFile, error) {
	if m.GetUnifiedChangedFilesFunc != nil {
		return m.GetUnifiedChangedFilesFunc(dir, cfg)
	}
	return []git.UnifiedChangedFile{}, nil
}

// GetFileDiff implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetFileDiff(dir string, filePath string, staged bool) (string, error) {
	if m.GetFileDiffFunc != nil {
		return m.GetFileDiffFunc(dir, filePath, staged)
	}
	return "mock diff content", nil
}

// GetUnifiedFileDiff implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetUnifiedFileDiff(dir string, file git.UnifiedChangedFile) (string, error) {
	if m.GetUnifiedFileDiffFunc != nil {
		return m.GetUnifiedFileDiffFunc(dir, file)
	}
	return "mock unified diff content", nil
}

// GetDiffWithOptions implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetDiffWithOptions(dir string, filePath string, staged bool, options git.DiffOptions) (string, error) {
	if m.GetDiffWithOptionsFunc != nil {
		return m.GetDiffWithOptionsFunc(dir, filePath, staged, options)
	}
	return "mock diff with options content", nil
}

// GetFileContent implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetFileContent(dir string, filePath string) (string, error) {
	if m.GetFileContentFunc != nil {
		return m.GetFileContentFunc(dir, filePath)
	}
	return "mock file content for " + filePath, nil
}
