package processor

import (
	"testing"

	"github.com/niels/git-llm-review/pkg/git"
)

func TestConvertStagedFilesToFileInfo(t *testing.T) {
	// Create test staged files
	stagedFiles := []git.StagedFile{
		{Path: "file1.go", Status: "M"},
		{Path: "file2.go", Status: "A"},
	}

	// Convert to FileInfo
	fileInfos := ConvertStagedFilesToFileInfo(stagedFiles)

	// Check results
	if len(fileInfos) != 2 {
		t.Errorf("Expected 2 FileInfo, got %d", len(fileInfos))
	}

	// Check file1.go
	if fileInfos[0].Path != "file1.go" || fileInfos[0].Status != "M" || fileInfos[0].Type != "staged" {
		t.Errorf("Unexpected FileInfo for file1.go: %+v", fileInfos[0])
	}

	// Check file2.go
	if fileInfos[1].Path != "file2.go" || fileInfos[1].Status != "A" || fileInfos[1].Type != "staged" {
		t.Errorf("Unexpected FileInfo for file2.go: %+v", fileInfos[1])
	}
}

func TestConvertChangedFilesToFileInfo(t *testing.T) {
	// Create test changed files
	changedFiles := []git.ChangedFile{
		{Path: "file1.go", Status: "M", Staged: true},
		{Path: "file2.go", Status: "A", Staged: true},
		{Path: "file3.go", Status: "M", Staged: false},
	}

	// Convert to FileInfo
	fileInfos := ConvertChangedFilesToFileInfo(changedFiles)

	// Check results
	if len(fileInfos) != 3 {
		t.Errorf("Expected 3 FileInfo, got %d", len(fileInfos))
	}

	// Check staged files have type "staged"
	for i := 0; i < 2; i++ {
		if fileInfos[i].Type != "staged" {
			t.Errorf("Expected type 'staged' for %s, got %s", fileInfos[i].Path, fileInfos[i].Type)
		}
	}

	// Check unstaged file has type "unstaged"
	if fileInfos[2].Type != "unstaged" {
		t.Errorf("Expected type 'unstaged' for %s, got %s", fileInfos[2].Path, fileInfos[2].Type)
	}
}

func TestConvertUnifiedChangedFilesToFileInfo(t *testing.T) {
	// Create test unified changed files
	unifiedFiles := []git.UnifiedChangedFile{
		{Path: "file1.go", StagedStatus: "M", UnstagedStatus: ""},
		{Path: "file2.go", StagedStatus: "A", UnstagedStatus: ""},
		{Path: "file3.go", StagedStatus: "", UnstagedStatus: "M"},
		{Path: "file4.go", StagedStatus: "M", UnstagedStatus: "M"},
	}

	// Convert to FileInfo
	fileInfos := ConvertUnifiedChangedFilesToFileInfo(unifiedFiles)

	// Check results
	if len(fileInfos) != 4 {
		t.Errorf("Expected 4 FileInfo, got %d", len(fileInfos))
	}

	// Check all files have type "unified"
	for i := 0; i < 4; i++ {
		if fileInfos[i].Type != "unified" {
			t.Errorf("Expected type 'unified' for %s, got %s", fileInfos[i].Path, fileInfos[i].Type)
		}
	}

	// Check file with both staged and unstaged changes has combined status
	if fileInfos[3].Status != "M+M" {
		t.Errorf("Expected status 'M+M' for file4.go, got %s", fileInfos[3].Status)
	}
}
