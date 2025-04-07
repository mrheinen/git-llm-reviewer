package git

import (
	"errors"
	"strings"
	"testing"
)

func TestGetFileDiff(t *testing.T) {
	// Test case 1: Staged file diff
	sampleStagedDiff := `diff --git a/file.go b/file.go
index 1234567..abcdefg 100644
--- a/file.go
+++ b/file.go
@@ -10,7 +10,7 @@ func Example() {
 	// Some code
-	fmt.Println("old line")
+	fmt.Println("new line")
 	// More code
 }
`

	mockRunner := &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "diff" && args[3] == "--cached" {
				return []byte(sampleStagedDiff), nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector := &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	diff, err := detector.GetFileDiff(".", "file.go", true)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !strings.Contains(diff, "old line") || !strings.Contains(diff, "new line") {
		t.Errorf("Expected diff to contain the changed lines, got: %s", diff)
	}

	// Test case 2: Unstaged file diff
	sampleUnstagedDiff := `diff --git a/file.go b/file.go
index abcdefg..9876543 100644
--- a/file.go
+++ b/file.go
@@ -15,3 +15,6 @@ func Example() {
 	// More code
+
+// New function added
+func NewFunction() {
+}
`

	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "diff" && len(args) == 4 {
				return []byte(sampleUnstagedDiff), nil
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	diff, err = detector.GetFileDiff(".", "file.go", false)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if !strings.Contains(diff, "New function added") {
		t.Errorf("Expected diff to contain the added function, got: %s", diff)
	}

	// Test case 3: Error handling - Git command failure
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("git command failed")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	_, err = detector.GetFileDiff(".", "file.go", true)
	if err == nil {
		t.Error("Expected error when Git command fails")
	}
}

func TestGetUnifiedFileDiff(t *testing.T) {
	// Test case 1: File with both staged and unstaged changes
	sampleStagedDiff := `diff --git a/file.go b/file.go
index 1234567..abcdefg 100644
--- a/file.go
+++ b/file.go
@@ -10,7 +10,7 @@ func Example() {
 	// Some code
-	fmt.Println("old line")
+	fmt.Println("new line")
 	// More code
 }
`

	sampleUnstagedDiff := `diff --git a/file.go b/file.go
index abcdefg..9876543 100644
--- a/file.go
+++ b/file.go
@@ -15,3 +15,6 @@ func Example() {
 	// More code
+
+// New function added
+func NewFunction() {
+}
`

	mockRunner := &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			if name == "git" && len(args) >= 3 && args[2] == "diff" {
				if len(args) > 3 && args[3] == "--cached" {
					return []byte(sampleStagedDiff), nil
				} else {
					return []byte(sampleUnstagedDiff), nil
				}
			}
			return nil, errors.New("unexpected command")
		},
	}

	detector := &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	unifiedFile := UnifiedChangedFile{
		Path:           "file.go",
		StagedStatus:   "M",
		UnstagedStatus: "M",
	}

	diff, err := detector.GetUnifiedFileDiff(".", unifiedFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that the unified diff contains both staged and unstaged changes
	if !strings.Contains(diff, "old line") || !strings.Contains(diff, "new line") || 
	   !strings.Contains(diff, "New function added") {
		t.Errorf("Expected unified diff to contain both staged and unstaged changes, got: %s", diff)
	}

	// Test case 2: File with only staged changes
	unifiedFile = UnifiedChangedFile{
		Path:           "file.go",
		StagedStatus:   "M",
		UnstagedStatus: "",
	}

	diff, err = detector.GetUnifiedFileDiff(".", unifiedFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that the diff contains only staged changes
	if !strings.Contains(diff, "old line") || !strings.Contains(diff, "new line") {
		t.Errorf("Expected diff to contain staged changes, got: %s", diff)
	}

	if strings.Contains(diff, "New function added") {
		t.Errorf("Expected diff not to contain unstaged changes, got: %s", diff)
	}

	// Test case 3: File with only unstaged changes
	unifiedFile = UnifiedChangedFile{
		Path:           "file.go",
		StagedStatus:   "",
		UnstagedStatus: "M",
	}

	diff, err = detector.GetUnifiedFileDiff(".", unifiedFile)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	// Check that the diff contains only unstaged changes
	if !strings.Contains(diff, "New function added") {
		t.Errorf("Expected diff to contain unstaged changes, got: %s", diff)
	}

	if strings.Contains(diff, "old line") || strings.Contains(diff, "new line") {
		t.Errorf("Expected diff not to contain staged changes, got: %s", diff)
	}

	// Test case 4: Error handling - Git command failure
	mockRunner = &mockCommandRunner{
		outputFunc: func(name string, args ...string) ([]byte, error) {
			return nil, errors.New("git command failed")
		},
	}

	detector = &RepositoryDetectorImpl{
		cmdRunner: mockRunner,
	}

	_, err = detector.GetUnifiedFileDiff(".", unifiedFile)
	if err == nil {
		t.Error("Expected error when Git command fails")
	}
}

func TestFormatDiff(t *testing.T) {
	// Test case 1: Format a standard Git diff
	rawDiff := `diff --git a/file.go b/file.go
index 1234567..abcdefg 100644
--- a/file.go
+++ b/file.go
@@ -10,7 +10,7 @@ func Example() {
 	// Some code
-	fmt.Println("old line")
+	fmt.Println("new line")
 	// More code
 }
`

	formattedDiff := formatDiff(rawDiff)
	
	// Check that the formatted diff contains the important parts
	if !strings.Contains(formattedDiff, "-	fmt.Println(\"old line\")") ||
	   !strings.Contains(formattedDiff, "+	fmt.Println(\"new line\")") {
		t.Errorf("Expected formatted diff to contain the changed lines, got: %s", formattedDiff)
	}

	// Test case 2: Format a diff with added file
	addedFileDiff := `diff --git a/new_file.go b/new_file.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/new_file.go
@@ -0,0 +1,10 @@
+package main
+
+import "fmt"
+
+func NewFunction() {
+	fmt.Println("This is a new file")
+}
`

	formattedAddedDiff := formatDiff(addedFileDiff)
	
	// Check that the formatted diff shows the entire new file
	if !strings.Contains(formattedAddedDiff, "+package main") ||
	   !strings.Contains(formattedAddedDiff, "+func NewFunction()") {
		t.Errorf("Expected formatted diff to show the entire new file, got: %s", formattedAddedDiff)
	}

	// Test case 3: Format a diff with deleted file
	deletedFileDiff := `diff --git a/deleted_file.go b/deleted_file.go
deleted file mode 100644
index 1234567..0000000
--- a/deleted_file.go
+++ /dev/null
@@ -1,10 +0,0 @@
-package main
-
-import "fmt"
-
-func OldFunction() {
-	fmt.Println("This file will be deleted")
-}
`

	formattedDeletedDiff := formatDiff(deletedFileDiff)
	
	// Check that the formatted diff shows the entire deleted file
	if !strings.Contains(formattedDeletedDiff, "-package main") ||
	   !strings.Contains(formattedDeletedDiff, "-func OldFunction()") {
		t.Errorf("Expected formatted diff to show the entire deleted file, got: %s", formattedDeletedDiff)
	}

	// Test case 4: Format an empty diff
	emptyDiff := ""
	formattedEmptyDiff := formatDiff(emptyDiff)
	
	if formattedEmptyDiff != "No changes detected." {
		t.Errorf("Expected formatted empty diff to indicate no changes, got: %s", formattedEmptyDiff)
	}
}
