package processor

import (
	"context"
	"sync"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/git"
	"github.com/niels/git-llm-review/pkg/parse"
	"github.com/niels/git-llm-review/pkg/progress"
)

// FileInfo represents a file to be processed
type FileInfo struct {
	Path   string // Path relative to repository root
	Status string // Git status
	Type   string // Type of file (staged, unstaged, unified)
}

// FileProcessor is a function that processes a single file and returns a review result or an error
type FileProcessor func(ctx context.Context, file FileInfo) (*parse.ReviewResult, error)

// ConcurrentProcessor processes multiple files concurrently with a configurable concurrency limit
type ConcurrentProcessor struct {
	config          *config.Config
	fileProcessor   FileProcessor
	progressTracker progress.Tracker
}

// NewConcurrentProcessor creates a new concurrent processor with the given configuration and file processor
func NewConcurrentProcessor(config *config.Config, fileProcessor FileProcessor) *ConcurrentProcessor {
	return &ConcurrentProcessor{
		config:          config,
		fileProcessor:   fileProcessor,
		progressTracker: progress.NewConsoleTracker(),
	}
}

// WithProgressTracker sets a custom progress tracker
func (p *ConcurrentProcessor) WithProgressTracker(tracker progress.Tracker) *ConcurrentProcessor {
	p.progressTracker = tracker
	return p
}

// ProcessFiles processes multiple files concurrently and returns the results and errors
func (p *ConcurrentProcessor) ProcessFiles(ctx context.Context, files []FileInfo) (map[string]*parse.ReviewResult, map[string]error) {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize progress tracker
	if p.progressTracker != nil {
		p.progressTracker.Start(len(files))
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, p.config.Concurrency.MaxTasks)

	// Create maps for results and errors with mutex for concurrent access
	var resultsMutex sync.Mutex
	results := make(map[string]*parse.ReviewResult)
	var errorsMutex sync.Mutex
	errors := make(map[string]error)

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start a goroutine for each file
	for _, file := range files {
		wg.Add(1)
		go func(file FileInfo) {
			defer wg.Done()

			// Acquire a semaphore slot
			select {
			case semaphore <- struct{}{}:
				// We got a slot, continue processing
				defer func() { <-semaphore }() // Release the slot when done
			case <-ctx.Done():
				// Context was cancelled, stop processing
				if p.progressTracker != nil {
					p.progressTracker.ErrorFile(file.Path, ctx.Err().Error())
				}
				
				// Safely add to errors map
				errorsMutex.Lock()
				errors[file.Path] = ctx.Err()
				errorsMutex.Unlock()
				
				return
			}

			// Update progress tracker
			if p.progressTracker != nil {
				p.progressTracker.StartFile(file.Path)
			}

			// Process the file
			result, err := p.fileProcessor(ctx, file)
			if err != nil {
				// Update progress tracker
				if p.progressTracker != nil {
					p.progressTracker.ErrorFile(file.Path, err.Error())
				}

				// Safely add to errors map
				errorsMutex.Lock()
				errors[file.Path] = err
				errorsMutex.Unlock()
				
				return
			}

			// Update progress tracker
			if p.progressTracker != nil {
				p.progressTracker.CompleteFile(file.Path, result.GetIssueCount())
			}

			// Safely add to results map
			resultsMutex.Lock()
			results[file.Path] = result
			resultsMutex.Unlock()
		}(file)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Finish progress tracking
	if p.progressTracker != nil {
		p.progressTracker.Finish()
	}

	return results, errors
}

// ProcessFilesWithCallback processes multiple files concurrently and calls the callback function for each result or error
func (p *ConcurrentProcessor) ProcessFilesWithCallback(
	ctx context.Context,
	files []FileInfo,
	resultCallback func(path string, result *parse.ReviewResult),
	errorCallback func(path string, err error),
) {
	// Create a context that can be cancelled
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Initialize progress tracker
	if p.progressTracker != nil {
		p.progressTracker.Start(len(files))
	}

	// Create a semaphore to limit concurrency
	semaphore := make(chan struct{}, p.config.Concurrency.MaxTasks)

	// Create a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	// Start a goroutine for each file
	for _, file := range files {
		wg.Add(1)
		go func(file FileInfo) {
			defer wg.Done()

			// Acquire a semaphore slot
			select {
			case semaphore <- struct{}{}:
				// We got a slot, continue processing
				defer func() { <-semaphore }() // Release the slot when done
			case <-ctx.Done():
				// Context was cancelled, stop processing
				if p.progressTracker != nil {
					p.progressTracker.ErrorFile(file.Path, ctx.Err().Error())
				}
				if errorCallback != nil {
					errorCallback(file.Path, ctx.Err())
				}
				return
			}

			// Update progress tracker
			if p.progressTracker != nil {
				p.progressTracker.StartFile(file.Path)
			}

			// Process the file
			result, err := p.fileProcessor(ctx, file)
			if err != nil {
				// Update progress tracker
				if p.progressTracker != nil {
					p.progressTracker.ErrorFile(file.Path, err.Error())
				}

				// Call the error callback
				if errorCallback != nil {
					errorCallback(file.Path, err)
				}
				return
			}

			// Update progress tracker
			if p.progressTracker != nil {
				p.progressTracker.CompleteFile(file.Path, result.GetIssueCount())
			}

			// Call the result callback
			if resultCallback != nil {
				resultCallback(file.Path, result)
			}
		}(file)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	// Finish progress tracking
	if p.progressTracker != nil {
		p.progressTracker.Finish()
	}
}

// ConvertStagedFilesToFileInfo converts a slice of StagedFile to a slice of FileInfo
func ConvertStagedFilesToFileInfo(stagedFiles []git.StagedFile) []FileInfo {
	files := make([]FileInfo, len(stagedFiles))
	for i, file := range stagedFiles {
		files[i] = FileInfo{
			Path:   file.Path,
			Status: file.Status,
			Type:   "staged",
		}
	}
	return files
}

// ConvertChangedFilesToFileInfo converts a slice of ChangedFile to a slice of FileInfo
func ConvertChangedFilesToFileInfo(changedFiles []git.ChangedFile) []FileInfo {
	files := make([]FileInfo, len(changedFiles))
	for i, file := range changedFiles {
		fileType := "unstaged"
		if file.Staged {
			fileType = "staged"
		}
		files[i] = FileInfo{
			Path:   file.Path,
			Status: file.Status,
			Type:   fileType,
		}
	}
	return files
}

// ConvertUnifiedChangedFilesToFileInfo converts a slice of UnifiedChangedFile to a slice of FileInfo
func ConvertUnifiedChangedFilesToFileInfo(unifiedFiles []git.UnifiedChangedFile) []FileInfo {
	files := make([]FileInfo, len(unifiedFiles))
	for i, file := range unifiedFiles {
		// Determine status based on staged and unstaged statuses
		var status string
		if file.StagedStatus != "" && file.UnstagedStatus != "" {
			// If both staged and unstaged statuses are present, combine them
			status = file.StagedStatus + "+" + file.UnstagedStatus
		} else if file.StagedStatus != "" {
			// If only staged status is present, use it
			status = file.StagedStatus
		} else {
			// If only unstaged status is present, use it
			status = file.UnstagedStatus
		}
		
		files[i] = FileInfo{
			Path:   file.Path,
			Status: status,
			Type:   "unified",
		}
	}
	return files
}