package processor

// File status constants
const (
	// FileStatusStaged indicates a file is staged in Git
	FileStatusStaged = "staged"
	// FileStatusUnstaged indicates a file is unstaged in Git
	FileStatusUnstaged = "unstaged"
	// FileStatusUnified indicates a file has both staged and unstaged changes
	FileStatusUnified = "unified"
)
