package git

import (
	"github.com/niels/git-llm-review/pkg/config"
)

// MockRepositoryDetector is a mock implementation of the RepositoryDetector
// for testing purposes
type MockRepositoryDetector struct {
	// IsGitRepositoryResult is the result to return from IsGitRepository
	IsGitRepositoryResult bool
	// IsGitRepositoryError is the error to return from IsGitRepository
	IsGitRepositoryError error
	// GetRepositoryRootResult is the result to return from GetRepositoryRoot
	GetRepositoryRootResult string
	// GetRepositoryRootError is the error to return from GetRepositoryRoot
	GetRepositoryRootError error
	// GetStagedFilesResult is the result to return from GetStagedFiles
	GetStagedFilesResult []StagedFile
	// GetStagedFilesError is the error to return from GetStagedFiles
	GetStagedFilesError error
	// GetAllChangedFilesResult is the result to return from GetAllChangedFiles
	GetAllChangedFilesResult []ChangedFile
	// GetAllChangedFilesError is the error to return from GetAllChangedFiles
	GetAllChangedFilesError error
	// GetUnifiedChangedFilesResult is the result to return from GetUnifiedChangedFiles
	GetUnifiedChangedFilesResult []UnifiedChangedFile
	// GetUnifiedChangedFilesError is the error to return from GetUnifiedChangedFiles
	GetUnifiedChangedFilesError error
}

// NewMockRepositoryDetector creates a new MockRepositoryDetector with default success values
func NewMockRepositoryDetector() *MockRepositoryDetector {
	return &MockRepositoryDetector{
		IsGitRepositoryResult:       true,
		GetRepositoryRootResult:     "/mock/repo/root",
		GetStagedFilesResult:        []StagedFile{},
		GetAllChangedFilesResult:    []ChangedFile{},
		GetUnifiedChangedFilesResult: []UnifiedChangedFile{},
	}
}

// IsGitRepository implements the RepositoryDetector interface
func (m *MockRepositoryDetector) IsGitRepository(dir string) (bool, error) {
	return m.IsGitRepositoryResult, m.IsGitRepositoryError
}

// GetRepositoryRoot implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetRepositoryRoot(dir string) (string, error) {
	return m.GetRepositoryRootResult, m.GetRepositoryRootError
}

// GetStagedFiles implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetStagedFiles(dir string, cfg *config.Config) ([]StagedFile, error) {
	return m.GetStagedFilesResult, m.GetStagedFilesError
}

// GetAllChangedFiles implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetAllChangedFiles(dir string, cfg *config.Config) ([]ChangedFile, error) {
	return m.GetAllChangedFilesResult, m.GetAllChangedFilesError
}

// GetUnifiedChangedFiles implements the RepositoryDetector interface
func (m *MockRepositoryDetector) GetUnifiedChangedFiles(dir string, cfg *config.Config) ([]UnifiedChangedFile, error) {
	return m.GetUnifiedChangedFilesResult, m.GetUnifiedChangedFilesError
}
