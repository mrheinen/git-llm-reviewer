package workflow

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/git"
	"github.com/niels/git-llm-review/pkg/llm"
	"github.com/niels/git-llm-review/pkg/llm/anthropic"
	"github.com/niels/git-llm-review/pkg/llm/exchangelog"
	"github.com/niels/git-llm-review/pkg/llm/openai"
	"github.com/niels/git-llm-review/pkg/llm/promptlog"
	"github.com/niels/git-llm-review/pkg/logging"
	"github.com/niels/git-llm-review/pkg/output"
	"github.com/niels/git-llm-review/pkg/parse"
	"github.com/niels/git-llm-review/pkg/processor"
	"github.com/niels/git-llm-review/pkg/progress"
)

// Options represents the options for the review workflow
type Options struct {
	ConfigPath      string
	All             bool
	OutputFormat    string
	OutputPath      string
	ProviderName    string
	VerboseOutput   bool
	LogPrompts      bool
	LogFullExchange bool
}

// Statistics represents review statistics
type Statistics struct {
	FilesProcessed  int
	FilesWithErrors int
	TotalIssues     int
	IssuesByType    map[string]int
	IssuesByFile    map[string]int
	Duration        time.Duration
}

// ReviewWorkflow represents the main workflow for the git-llm-review tool
type ReviewWorkflow struct {
	options        Options
	config         *config.Config
	repoDetector   git.RepositoryDetector
	provider       llm.Provider
	terminalOutput *output.TerminalFormatter
	markdownOutput *output.MarkdownFormatter
	progressTracker progress.Tracker
	diffHighlights  []diffHighlight
	pendingDiffs    []pendingDiff
}

type diffHighlight struct {
	marker   string
	diff     string
	filePath string
}

type pendingDiff struct {
	diff     string
	filePath string
	position int
	isLegacy bool
	issueIdx int
}

// NewReviewWorkflow creates a new review workflow with the given options
func NewReviewWorkflow(options Options) (*ReviewWorkflow, error) {
	// Load configuration
	var cfg *config.Config
	var err error

	if options.ConfigPath != "" {
		cfg, err = config.Load(options.ConfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load configuration: %w", err)
		}
	} else {
		cfg = config.Default()
	}

	// Initialize repository detector
	repoDetector := git.NewRepositoryDetector()

	// Initialize prompt logger if requested
	if options.LogPrompts {
		logging.Info("Initializing prompt logger")
		promptLogPath := cfg.Logging.PromptLogPath
		logging.InfoWith("Prompt logging enabled", map[string]interface{}{
			"path": promptLogPath,
		})
		
		if err := promptlog.InitGlobalLogger(true, promptLogPath); err != nil {
			logging.ErrorWith("Failed to initialize prompt logger", map[string]interface{}{
				"error": err.Error(),
				"path":  promptLogPath,
			})
			// Continue without prompt logging, but log the error
		}
	} else {
		logging.Debug("Prompt logging is disabled")
	}
	
	// Initialize exchange logger if requested
	if options.LogFullExchange {
		logging.Info("Initializing exchange logger")
		exchangeLogPath := "exchange.log" // Use a default path or add to config
		logging.InfoWith("Exchange logging enabled", map[string]interface{}{
			"path": exchangeLogPath,
		})
		
		if err := exchangelog.InitGlobalLogger(true, exchangeLogPath); err != nil {
			logging.ErrorWith("Failed to initialize exchange logger", map[string]interface{}{
				"error": err.Error(),
				"path":  exchangeLogPath,
			})
			// Continue without exchange logging, but log the error
		}
	} else {
		logging.Debug("Exchange logging is disabled")
	}

	// Initialize LLM provider
	var provider llm.Provider
	
	providerName := options.ProviderName
	if providerName == "" {
		// If provider name is not specified, use the one from config
		providerName = cfg.LLM.Provider
	}
	
	switch strings.ToLower(providerName) {
	case "openai":
		provider, err = openai.NewProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize OpenAI provider: %w", err)
		}
	case "anthropic":
		provider, err = anthropic.NewProvider(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize Anthropic provider: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}

	// Initialize output formatters
	terminalOutput := output.NewTerminalFormatter(true) // Always use color in the workflow
	markdownOutput := output.NewMarkdownFormatter()
	
	// Initialize progress tracker
	progressTracker := progress.NewConsoleTracker()

	return &ReviewWorkflow{
		options:        options,
		config:         cfg,
		repoDetector:   repoDetector,
		provider:       provider,
		terminalOutput: terminalOutput,
		markdownOutput: markdownOutput,
		progressTracker: progressTracker,
		diffHighlights:  make([]diffHighlight, 0),
		pendingDiffs:    make([]pendingDiff, 0),
	}, nil
}

// Run executes the review workflow
func (w *ReviewWorkflow) Run(ctx context.Context) (*Statistics, error) {
	startTime := time.Now()
	
	stats := &Statistics{
		IssuesByType: make(map[string]int),
		IssuesByFile: make(map[string]int),
	}

	// Step 1: Detect Git repository
	logging.Info("Detecting Git repository...")
	isRepo, err := w.repoDetector.IsGitRepository(".")
	if err != nil {
		return nil, fmt.Errorf("failed to detect Git repository: %w", err)
	}
	if !isRepo {
		return nil, fmt.Errorf("not a Git repository")
	}

	// Get repository root
	repoRoot, err := w.repoDetector.GetRepositoryRoot(".")
	if err != nil {
		return nil, fmt.Errorf("failed to get repository root: %w", err)
	}
	logging.InfoWith("Git repository detected", map[string]interface{}{
		"root": repoRoot,
	})

	// Step 2: Get repository name from the root directory
	repoName := filepath.Base(repoRoot)
	logging.InfoWith("Repository name detected", map[string]interface{}{
		"name": repoName,
	})

	// Step 3: Detect changed files
	var files []processor.FileInfo
	if w.options.All {
		// Get all changed files (staged and unstaged)
		logging.Info("Detecting all changed files...")
		changedFiles, err := w.repoDetector.GetAllChangedFiles(repoRoot, w.config)
		if err != nil {
			return nil, fmt.Errorf("failed to get changed files: %w", err)
		}
		files = processor.ConvertChangedFilesToFileInfo(changedFiles)
	} else {
		// Get only staged files
		logging.Info("Detecting staged files...")
		stagedFiles, err := w.repoDetector.GetStagedFiles(repoRoot, w.config)
		if err != nil {
			return nil, fmt.Errorf("failed to get staged files: %w", err)
		}
		files = processor.ConvertStagedFilesToFileInfo(stagedFiles)
	}

	if len(files) == 0 {
		logging.Info("No files to review")
		fmt.Println("No files to review")
		return stats, nil
	}

	logging.InfoWith("Found files to review", map[string]interface{}{
		"count": len(files),
	})
	
	// Display files to be reviewed
	fmt.Printf("Found %d files to review:\n", len(files))
	for _, file := range files {
		statusDesc := file.Status
		if file.Type != "" {
			statusDesc = fmt.Sprintf("%s (%s)", file.Status, file.Type)
		}
		fmt.Printf("  %s %s\n", statusDesc, file.Path)
	}

	// Step 4: Create file processor
	fileProcessor := processor.ReviewFileProcessor(
		repoRoot,
		w.repoDetector, 
		w.provider,
		w.provider.Name(),
	)

	// Step 5: Create concurrent processor with progress tracker
	concurrentProcessor := processor.NewConcurrentProcessor(w.config, fileProcessor).
		WithProgressTracker(w.progressTracker)

	// Step 6: Process files concurrently
	logging.InfoWith("Processing files concurrently", map[string]interface{}{
		"concurrency": w.config.Concurrency.MaxTasks,
	})
	fmt.Printf("\nProcessing %d files with concurrency %d...\n", len(files), w.config.Concurrency.MaxTasks)

	// Create a timeout context based on the configuration
	timeoutCtx, cancel := context.WithTimeout(ctx, time.Duration(w.config.LLM.Timeout)*time.Second)
	defer cancel()
	
	results, errors := concurrentProcessor.ProcessFiles(timeoutCtx, files)

	// Step 7: Collect statistics and display results
	stats.FilesProcessed = len(results)
	stats.FilesWithErrors = len(errors)
	stats.Duration = time.Since(startTime)
	
	// Display errors
	if len(errors) > 0 {
		logging.Info("Errors encountered during processing:")
		fmt.Println("\nErrors encountered during processing:")
		for filePath, err := range errors {
			logging.ErrorWith("Error processing file", map[string]interface{}{
				"file":  filePath,
				"error": err.Error(),
			})
			fmt.Printf("  %s: %s\n", filePath, err.Error())
		}
	}

	// Display results in terminal
	for filePath, result := range results {
		issueCount := result.GetIssueCount()
		stats.TotalIssues += issueCount
		stats.IssuesByFile[filePath] = issueCount
		
		for _, issue := range result.Issues {
			issueType := extractIssueType(issue.Title)
			stats.IssuesByType[issueType]++
		}
		
		fmt.Printf("\n=== Review for %s (%d issues) ===\n", filePath, issueCount)
		formattedReview := w.terminalOutput.FormatReview(result)
		
		// Process the formatted review to handle the diff placeholders
		
		// Process the output line by line
		lines := strings.Split(formattedReview, "\n")
		var processedOutput strings.Builder
		
		lineIdx := 0
		for lineIdx < len(lines) {
			line := lines[lineIdx]
			
			// Check for diff markers
			if strings.Contains(line, "<<HIGHLIGHTED_DIFF:") {
				// Log the diff marker found
				logging.DebugWith("Found diff marker", map[string]interface{}{
					"marker": line,
				})
				
				// Extract file path from the marker
				// The marker format is: <<HIGHLIGHTED_DIFF:filepath>>
				parts := strings.Split(line, ":")
				if len(parts) >= 2 {
					filePath := strings.TrimSuffix(parts[1], ">>")
					logging.DebugWith("Extracted file path from marker", map[string]interface{}{
						"filePath": filePath,
					})
					
					// Find the corresponding diff
					var diffContent string
					logging.DebugWith("Looking for diff", map[string]interface{}{
						"availableDiffs": len(result.Diffs),
					})
					for i, diff := range result.Diffs {
						logging.DebugWith("Comparing diff", map[string]interface{}{
							"index": i,
							"diffFile": diff.File,
							"targetFile": filePath,
						})
						if diff.File == filePath {
							diffContent = diff.Diff
							logging.DebugWith("Found matching diff", map[string]interface{}{
								"contentLength": len(diffContent),
							})
							break
						}
					}
					
					// Output a separator and the file path
					fmt.Fprintf(&processedOutput, "\n--- Diff for %s ---\n", filePath)
					
					// Print the formatted text up to this point
					fmt.Print(processedOutput.String())
					processedOutput.Reset()
					
					// Directly highlight the diff to terminal
					if diffContent != "" {
						logging.DebugWith("Highlighting diff", map[string]interface{}{
							"file": filePath,
						})
						w.terminalOutput.HighlightDiff(diffContent, filePath)
					} else {
						logging.WarnWith("No diff content available", map[string]interface{}{
							"file": filePath,
						})
						fmt.Println("[No diff content available]")
					}
					
					// Skip the next line if it's empty (usually follows the marker)
					if lineIdx+1 < len(lines) && lines[lineIdx+1] == "" {
						lineIdx++
					}
				}
			} else if strings.Contains(line, "<<HIGHLIGHTED_DIFF_LEGACY>>") {
				// Handle legacy diffs similarly
				// Find the first issue with a diff
				for i, issue := range result.Issues {
					if issue.Diff != "" {
						// Output a separator
						fmt.Fprintf(&processedOutput, "\n--- Legacy diff for issue %d ---\n", i+1)
						
						// Print the formatted text up to this point
						fmt.Print(processedOutput.String())
						processedOutput.Reset()
						
						// Directly highlight the diff
						w.terminalOutput.HighlightDiff(issue.Diff, issue.File)
						break
					}
				}
				
				// Skip the next line if it's empty
				if lineIdx+1 < len(lines) && lines[lineIdx+1] == "" {
					lineIdx++
				}
			} else {
				// Regular line, add to the output buffer
				processedOutput.WriteString(line)
				processedOutput.WriteString("\n")
			}
			
			lineIdx++
		}
		
		// Print any remaining output
		if processedOutput.Len() > 0 {
			fmt.Print(processedOutput.String())
		}
	}

	// Generate markdown report if requested
	if w.options.OutputPath != "" {
		outputDir := w.options.OutputPath
		
		logging.InfoWith("Generating markdown reports", map[string]interface{}{
			"output_dir": outputDir,
		})
		fmt.Printf("\nGenerating markdown reports to %s...\n", outputDir)

		// Create output directory if it doesn't exist
		if err := os.MkdirAll(outputDir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create output directory: %w", err)
		}

		// Write the report to file for each result
		for filePath, result := range results {
			// Create a safe filename from the file path
			safePath := strings.ReplaceAll(filePath, "/", "_")
			reportPath := filepath.Join(outputDir, fmt.Sprintf("review_%s.md", safePath))
			
			if err := w.markdownOutput.WriteToFile(result, filePath, repoName, reportPath); err != nil {
				return stats, fmt.Errorf("failed to write markdown report for %s: %w", filePath, err)
			}
			fmt.Printf("  Report for %s written to %s\n", filePath, reportPath)
		}

		// Create a summary report
		summaryPath := filepath.Join(outputDir, "summary.md")
		if err := w.writeSummaryReport(summaryPath, results, stats, repoName); err != nil {
			logging.ErrorWith("Failed to write summary report", map[string]interface{}{
				"error": err.Error(),
			})
			// Continue execution even if summary report fails
		} else {
			fmt.Printf("  Summary report written to %s\n", summaryPath)
		}

		logging.InfoWith("Markdown reports generated", map[string]interface{}{
			"output_dir": outputDir,
		})
	}

	// Step 8: Show summary statistics
	fmt.Println("\nSummary:")
	fmt.Printf("Files processed: %d\n", stats.FilesProcessed)
	fmt.Printf("Files with errors: %d\n", stats.FilesWithErrors)
	fmt.Printf("Total issues found: %d\n", stats.TotalIssues)
	
	if len(stats.IssuesByType) > 0 {
		fmt.Println("Issues by type:")
		for issueType, count := range stats.IssuesByType {
			fmt.Printf("  %s: %d\n", issueType, count)
		}
	}
	
	if len(stats.IssuesByFile) > 0 {
		fmt.Println("Issues by file:")
		for filePath, count := range stats.IssuesByFile {
			fmt.Printf("  %s: %d\n", filePath, count)
		}
	}
	
	fmt.Printf("Time taken: %s\n", stats.Duration.Round(time.Second))

	return stats, nil
}

// extractIssueType extracts the issue type from the title
func extractIssueType(title string) string {
	// Common issue types
	issueTypes := []string{
		"Bug", "Error", "Security", "Performance", 
		"Style", "Documentation", "Optimization", 
		"Refactoring", "Testing", "Maintainability",
	}
	
	// Check if the title starts with any of the issue types
	for _, issueType := range issueTypes {
		if strings.HasPrefix(title, issueType+":") || 
		   strings.HasPrefix(title, "["+issueType+"]") ||
		   strings.HasPrefix(title, issueType+" -") {
			return issueType
		}
	}
	
	// Default to "General" if no specific type is found
	return "General"
}

// writeSummaryReport writes a summary report for all the results
func (w *ReviewWorkflow) writeSummaryReport(filePath string, results map[string]*parse.ReviewResult, stats *Statistics, repoName string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create summary report file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	defer writer.Flush()

	// Write header
	fmt.Fprintf(writer, "# Code Review Summary for %s\n\n", repoName)
	fmt.Fprintf(writer, "Generated on: %s\n\n", time.Now().Format(time.RFC1123))
	
	// Write statistics
	fmt.Fprintf(writer, "## Statistics\n\n")
	fmt.Fprintf(writer, "- Files processed: %d\n", stats.FilesProcessed)
	fmt.Fprintf(writer, "- Files with errors: %d\n", stats.FilesWithErrors)
	fmt.Fprintf(writer, "- Total issues found: %d\n", stats.TotalIssues)
	fmt.Fprintf(writer, "- Time taken: %s\n\n", stats.Duration.Round(time.Second))
	
	// Write issues by type
	if len(stats.IssuesByType) > 0 {
		fmt.Fprintf(writer, "### Issues by Type\n\n")
		for issueType, count := range stats.IssuesByType {
			fmt.Fprintf(writer, "- %s: %d\n", issueType, count)
		}
		fmt.Fprintf(writer, "\n")
	}
	
	// Write issues by file
	if len(stats.IssuesByFile) > 0 {
		fmt.Fprintf(writer, "### Issues by File\n\n")
		for filePath, count := range stats.IssuesByFile {
			fmt.Fprintf(writer, "- [%s](%s): %d\n", filePath, strings.ReplaceAll(filePath, "/", "_")+".md", count)
		}
		fmt.Fprintf(writer, "\n")
	}
	
	// Write file summaries
	fmt.Fprintf(writer, "## File Summaries\n\n")
	for filePath, result := range results {
		fmt.Fprintf(writer, "### [%s](%s)\n\n", filePath, strings.ReplaceAll(filePath, "/", "_")+".md")
		
		if len(result.Issues) == 0 {
			fmt.Fprintf(writer, "No issues found\n\n")
			continue
		}
		
		fmt.Fprintf(writer, "Found %d issues:\n\n", len(result.Issues))
		for _, issue := range result.Issues {
			fmt.Fprintf(writer, "1. **%s**  \n", issue.Title)
			fmt.Fprintf(writer, "   %s\n\n", issue.Explanation)
		}
	}
	
	return nil
}