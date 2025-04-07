package processor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/niels/git-llm-review/pkg/git"
	"github.com/niels/git-llm-review/pkg/llm"
	"github.com/niels/git-llm-review/pkg/llm/promptlog"
	"github.com/niels/git-llm-review/pkg/logging"
	"github.com/niels/git-llm-review/pkg/parse"
	"github.com/niels/git-llm-review/pkg/prompt"
)

// ReviewFileProcessor creates a FileProcessor that processes a file for code review
func ReviewFileProcessor(
	repoRoot string,
	repoDetector git.RepositoryDetector,
	provider llm.Provider,
	providerName string,
) FileProcessor {
	return func(ctx context.Context, file FileInfo) (*parse.ReviewResult, error) {
		// Log the start of processing
		logging.InfoWith("Processing file for review", map[string]interface{}{
			"file": file.Path,
		})

		// Get the diff for the file
		var diff string
		var err error

		if file.Type == "unstaged" || file.Type == "unified" {
			// Get diff for unstaged or unified file
			unifiedFile := git.UnifiedChangedFile{
				Path:           file.Path,
				StagedStatus:   file.Status,
				UnstagedStatus: file.Status,
			}
			diff, err = repoDetector.GetUnifiedFileDiff(repoRoot, unifiedFile)
		} else {
			// Get diff for staged file
			diff, err = repoDetector.GetFileDiff(repoRoot, file.Path, true)
		}

		if err != nil {
			logging.ErrorWith("Failed to get diff for file", map[string]interface{}{
				"file":  file.Path,
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to get diff for file %s: %w", file.Path, err)
		}

		// Check if the context has been cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue processing
		}

		// Generate prompt for code review
		// Create a proper review request with file diff
		reviewRequest := &llm.ReviewRequest{
			FilePath:    file.Path,
			FileDiff:    diff,
			FileContent: "", // We don't have the full file content here, only the diff
			Options: llm.ReviewOptions{
				Timeout:              30 * time.Second, // Default timeout
				MaxTokens:            1024,
				Temperature:          0.1,
				IncludeExplanations: true,
			},
		}

		// Use the full-featured prompt generation that properly handles templates
		var providerType prompt.ProviderType
		switch strings.ToLower(providerName) {
		case "anthropic":
			providerType = prompt.ProviderAnthropic
		case "openai":
			providerType = prompt.ProviderOpenAI
		default:
			providerType = prompt.ProviderDefault
		}
		
		reviewPrompt := prompt.CreatePrompt(reviewRequest, providerType)

		// Log the prompt if the prompt logger is enabled
		err = promptlog.LogPrompt(provider.Name(), file.Path, reviewPrompt)
		if err != nil {
			// Just log the error but continue with the process
			logging.ErrorWith("Failed to log prompt", map[string]interface{}{
				"file":  file.Path,
				"error": err.Error(),
			})
		}

		// Get response from LLM
		logging.InfoWith("Sending code review request to LLM", map[string]interface{}{
			"file": file.Path,
		})

		response, err := provider.GetCompletion(reviewPrompt)
		if err != nil {
			logging.ErrorWith("Failed to get LLM response", map[string]interface{}{
				"file":  file.Path,
				"error": err.Error(),
			})
			return nil, fmt.Errorf("failed to get LLM response for file %s: %w", file.Path, err)
		}

		// Check if the context has been cancelled
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			// Continue processing
		}

		// Parse the response
		reviewResult := parse.ParseReview(response)

		logging.InfoWith("Completed review for file", map[string]interface{}{
			"file":        file.Path,
			"issue_count": reviewResult.GetIssueCount(),
		})

		return reviewResult, nil
	}
}