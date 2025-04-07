package progress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/fatih/color"
)

// Status represents the status of a file being processed
type Status string

const (
	// StatusPending indicates the file is waiting to be processed
	StatusPending Status = "pending"
	// StatusProcessing indicates the file is currently being processed
	StatusProcessing Status = "processing"
	// StatusCompleted indicates the file has been processed successfully
	StatusCompleted Status = "completed"
	// StatusError indicates an error occurred while processing the file
	StatusError Status = "error"
)

// FileProgress represents the progress of a single file
type FileProgress struct {
	Path       string
	Status     Status
	StartTime  time.Time
	EndTime    time.Time
	Message    string
	IssueCount int
}

// Tracker is an interface for tracking progress of file processing
type Tracker interface {
	// Start initializes the progress tracker with the total number of files
	Start(totalFiles int)
	// StartFile marks a file as being processed
	StartFile(path string)
	// CompleteFile marks a file as completed with optional issue count
	CompleteFile(path string, issueCount int)
	// ErrorFile marks a file as having an error with an error message
	ErrorFile(path string, message string)
	// Finish completes the progress tracking
	Finish()
}

// ConsoleTracker implements Tracker for console output
type ConsoleTracker struct {
	mu           sync.Mutex
	writer       io.Writer
	totalFiles   int
	fileProgress map[string]*FileProgress
	startTime    time.Time
	spinner      []string
	spinnerIndex int
	lastUpdate   time.Time
	completed    int
	errors       int
}

// NewConsoleTracker creates a new console progress tracker
func NewConsoleTracker() *ConsoleTracker {
	return &ConsoleTracker{
		writer:       os.Stdout,
		fileProgress: make(map[string]*FileProgress),
		spinner:      []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"},
		lastUpdate:   time.Now(),
	}
}

// WithWriter sets the writer for the console tracker
func (t *ConsoleTracker) WithWriter(writer io.Writer) *ConsoleTracker {
	t.writer = writer
	return t
}

// Start initializes the progress tracker with the total number of files
func (t *ConsoleTracker) Start(totalFiles int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.totalFiles = totalFiles
	t.startTime = time.Now()
	t.completed = 0
	t.errors = 0

	fmt.Fprintf(t.writer, "Starting code review of %d files...\n", totalFiles)
}

// StartFile marks a file as being processed
func (t *ConsoleTracker) StartFile(path string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.fileProgress[path] = &FileProgress{
		Path:      path,
		Status:    StatusProcessing,
		StartTime: time.Now(),
	}

	t.updateProgress()
}

// CompleteFile marks a file as completed with optional issue count
func (t *ConsoleTracker) CompleteFile(path string, issueCount int) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if progress, ok := t.fileProgress[path]; ok {
		progress.Status = StatusCompleted
		progress.EndTime = time.Now()
		progress.IssueCount = issueCount
	} else {
		t.fileProgress[path] = &FileProgress{
			Path:       path,
			Status:     StatusCompleted,
			StartTime:  time.Now(),
			EndTime:    time.Now(),
			IssueCount: issueCount,
		}
	}

	t.completed++
	t.updateProgress()
}

// ErrorFile marks a file as having an error with an error message
func (t *ConsoleTracker) ErrorFile(path string, message string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if progress, ok := t.fileProgress[path]; ok {
		progress.Status = StatusError
		progress.EndTime = time.Now()
		progress.Message = message
	} else {
		t.fileProgress[path] = &FileProgress{
			Path:      path,
			Status:    StatusError,
			StartTime: time.Now(),
			EndTime:   time.Now(),
			Message:   message,
		}
	}

	t.errors++
	t.updateProgress()
}

// Finish completes the progress tracking
func (t *ConsoleTracker) Finish() {
	t.mu.Lock()
	defer t.mu.Unlock()

	duration := time.Since(t.startTime).Round(time.Second)
	fmt.Fprintf(t.writer, "\nCode review completed in %s\n", duration)
	fmt.Fprintf(t.writer, "Processed %d files: %d completed, %d errors\n", 
		t.totalFiles, t.completed, t.errors)
}

// updateProgress updates the progress display
func (t *ConsoleTracker) updateProgress() {
	// Only update at most 10 times per second to avoid flickering
	if time.Since(t.lastUpdate) < 100*time.Millisecond {
		return
	}
	t.lastUpdate = time.Now()

	// Clear the current line
	fmt.Fprint(t.writer, "\r\033[K")

	// Calculate progress percentage
	progress := float64(t.completed+t.errors) / float64(t.totalFiles)
	progressBar := createProgressBar(progress, 20)

	// Get the current spinner frame
	t.spinnerIndex = (t.spinnerIndex + 1) % len(t.spinner)
	spinnerChar := t.spinner[t.spinnerIndex]

	// Find currently processing files
	var processingFiles []string
	for path, progress := range t.fileProgress {
		if progress.Status == StatusProcessing {
			processingFiles = append(processingFiles, path)
		}
	}

	// Format the progress message
	progressMsg := fmt.Sprintf("%s %s %d/%d files (%.0f%%)", 
		spinnerChar, progressBar, t.completed+t.errors, t.totalFiles, progress*100)

	// Add currently processing files
	if len(processingFiles) > 0 {
		currentFile := processingFiles[0]
		if len(currentFile) > 30 {
			// Truncate long filenames
			currentFile = "..." + currentFile[len(currentFile)-27:]
		}
		progressMsg += fmt.Sprintf(" | Processing: %s", color.CyanString(currentFile))
	}

	fmt.Fprint(t.writer, progressMsg)
}

// createProgressBar creates a visual progress bar
func createProgressBar(progress float64, width int) string {
	completed := int(progress * float64(width))
	remaining := width - completed

	bar := "["
	bar += color.GreenString(repeat("=", completed))
	if remaining > 0 {
		bar += ">"
		bar += repeat(" ", remaining-1)
	}
	bar += "]"

	return bar
}

// repeat repeats a string n times
func repeat(s string, n int) string {
	if n <= 0 {
		return ""
	}
	result := ""
	for i := 0; i < n; i++ {
		result += s
	}
	return result
}