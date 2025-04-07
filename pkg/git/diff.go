package git

import (
	"fmt"
	"regexp"
	"strings"
)

// DiffOptions represents options for generating diffs
type DiffOptions struct {
	ContextLines int  // Number of context lines to include in the diff
	ColorOutput  bool // Whether to include color codes in the output
}

// DefaultDiffOptions returns the default options for diff generation
func DefaultDiffOptions() DiffOptions {
	return DiffOptions{
		ContextLines: 3,
		ColorOutput:  false,
	}
}

// GetFileDiff returns the diff for a specific file
// If staged is true, it returns the diff for staged changes
// If staged is false, it returns the diff for unstaged changes
func (d *RepositoryDetectorImpl) GetFileDiff(dir string, filePath string, staged bool) (string, error) {
	var args []string
	
	// Base arguments
	args = append(args, "-C", dir, "diff")
	
	// Add --cached flag for staged changes
	if staged {
		args = append(args, "--cached")
	}
	
	// Add the file path
	args = append(args, filePath)
	
	// Run the git diff command
	output, err := d.cmdRunner.runCommand("git", args...)
	if err != nil {
		return "", fmt.Errorf("failed to get diff for file %s: %w", filePath, err)
	}
	
	// Format the diff for better readability
	return formatDiff(string(output)), nil
}

// GetUnifiedFileDiff returns a unified diff for a file that may have both staged and unstaged changes
func (d *RepositoryDetectorImpl) GetUnifiedFileDiff(dir string, file UnifiedChangedFile) (string, error) {
	var stagedDiff, unstagedDiff string
	var err error
	
	// Get staged changes if any
	if file.StagedStatus != "" {
		stagedDiff, err = d.GetFileDiff(dir, file.Path, true)
		if err != nil {
			return "", fmt.Errorf("failed to get staged diff for file %s: %w", file.Path, err)
		}
	}
	
	// Get unstaged changes if any
	if file.UnstagedStatus != "" {
		unstagedDiff, err = d.GetFileDiff(dir, file.Path, false)
		if err != nil {
			return "", fmt.Errorf("failed to get unstaged diff for file %s: %w", file.Path, err)
		}
	}
	
	// Combine the diffs with appropriate headers
	var result strings.Builder
	
	if file.StagedStatus != "" && file.UnstagedStatus != "" {
		// Both staged and unstaged changes
		result.WriteString(fmt.Sprintf("=== Staged changes (%s) ===\n", file.StagedStatus))
		result.WriteString(stagedDiff)
		result.WriteString(fmt.Sprintf("\n=== Unstaged changes (%s) ===\n", file.UnstagedStatus))
		result.WriteString(unstagedDiff)
	} else if file.StagedStatus != "" {
		// Only staged changes
		result.WriteString(fmt.Sprintf("=== Staged changes (%s) ===\n", file.StagedStatus))
		result.WriteString(stagedDiff)
	} else if file.UnstagedStatus != "" {
		// Only unstaged changes
		result.WriteString(fmt.Sprintf("=== Unstaged changes (%s) ===\n", file.UnstagedStatus))
		result.WriteString(unstagedDiff)
	} else {
		// No changes (shouldn't happen)
		result.WriteString("No changes detected.")
	}
	
	return result.String(), nil
}

// GetDiffWithOptions returns the diff for a specific file with custom options
func (d *RepositoryDetectorImpl) GetDiffWithOptions(dir string, filePath string, staged bool, options DiffOptions) (string, error) {
	var args []string
	
	// Base arguments
	args = append(args, "-C", dir, "diff")
	
	// Add --cached flag for staged changes
	if staged {
		args = append(args, "--cached")
	}
	
	// Add context lines option
	args = append(args, fmt.Sprintf("-U%d", options.ContextLines))
	
	// Add color option
	if options.ColorOutput {
		args = append(args, "--color")
	} else {
		args = append(args, "--no-color")
	}
	
	// Add the file path
	args = append(args, filePath)
	
	// Run the git diff command
	output, err := d.cmdRunner.runCommand("git", args...)
	if err != nil {
		return "", fmt.Errorf("failed to get diff for file %s: %w", filePath, err)
	}
	
	// Format the diff for better readability
	return formatDiff(string(output)), nil
}

// formatDiff formats a Git diff for better readability
// It removes unnecessary header information and focuses on the actual changes
func formatDiff(diff string) string {
	if diff == "" {
		return "No changes detected."
	}
	
	// Split the diff into lines
	lines := strings.Split(diff, "\n")
	
	// Process the diff to make it more readable
	var result []string
	inHeader := true
	
	// Regular expression to match the @@ line which indicates the start of a hunk
	hunkHeaderRegex := regexp.MustCompile(`^@@ -\d+,\d+ \+\d+,\d+ @@`)
	
	for _, line := range lines {
		// Skip the first few lines of header information
		if inHeader {
			if hunkHeaderRegex.MatchString(line) {
				inHeader = false
			}
		}
		
		// Include the line if we're past the header or if it's a hunk header
		if !inHeader || hunkHeaderRegex.MatchString(line) {
			result = append(result, line)
		}
	}
	
	// Join the lines back together
	return strings.Join(result, "\n")
}

// GetFileStatus returns a human-readable description of the file status
func GetFileStatus(status string) string {
	switch status {
	case "A":
		return "Added"
	case "M":
		return "Modified"
	case "D":
		return "Deleted"
	case "R":
		return "Renamed"
	case "C":
		return "Copied"
	case "U":
		return "Updated but unmerged"
	case "??":
		return "Untracked"
	case "!!":
		return "Ignored"
	case "AM":
		return "Added with modifications"
	case "MM":
		return "Modified with additional modifications"
	case "RM":
		return "Renamed with modifications"
	default:
		return status
	}
}
