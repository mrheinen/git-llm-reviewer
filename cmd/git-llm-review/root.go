package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/git"
	"github.com/niels/git-llm-review/pkg/llm"
	"github.com/niels/git-llm-review/pkg/llm/exchangelog"
	"github.com/niels/git-llm-review/pkg/llm/promptlog"
	"github.com/niels/git-llm-review/pkg/logging"
	"github.com/niels/git-llm-review/pkg/output"
	"github.com/niels/git-llm-review/pkg/parse"
	"github.com/niels/git-llm-review/pkg/prompt"
	"github.com/niels/git-llm-review/pkg/version"
	"github.com/spf13/cobra"
)

var (
	configPath      string
	outputPath      string
	markdownPath    string
	debug           bool
	all             bool
	showVersion     bool
	noColor         bool
	logPrompts      bool
	logFullExchange bool
	cfg             *config.Config
	repoDetector    git.RepositoryDetector
)

// NewRootCmd creates the root command for git-llm-review
func NewRootCmd() *cobra.Command {
	return NewRootCmdWithDetector(nil)
}

// NewRootCmdWithDetector creates the root command with a custom repository detector
func NewRootCmdWithDetector(detector git.RepositoryDetector) *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   version.AppName,
		Short: version.Description,
		Long: fmt.Sprintf(`%s - %s

A tool that uses LLMs to review code changes in Git repositories.
`, version.AppName, version.Description),
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Initialize the logger with default logging config
			cfg := config.Default()

			// If config file is specified, try to load it first for logging settings
			if configPath != "" {
				loadedConfig := config.LoadOrDefault(configPath)
				cfg = loadedConfig
			}

			logging.InitGlobalLogger(debug, cfg)
			logging.Info("Initializing git-llm-reviewer")

			// Initialize prompt logger if enabled
			if logPrompts {
				if err := promptlog.InitGlobalLogger(true, "prompt.log"); err != nil {
					return fmt.Errorf("failed to initialize prompt logger: %w", err)
				}
				logging.Info("Prompt logging enabled")
			}

			// Initialize exchange logger if enabled
			if logFullExchange {
				if err := exchangelog.InitGlobalLogger(true, "exchange.log"); err != nil {
					return fmt.Errorf("failed to initialize exchange logger: %w", err)
				}
				logging.Info("Full exchange logging enabled")
			}

			if debug {
				logging.Debug("Debug logging enabled")
			}

			// Load configuration
			if configPath != "" {
				logging.InfoWith("Loading configuration", map[string]interface{}{
					"path": configPath,
				})

				// Use LoadOrDefault instead of Load to handle missing files gracefully
				cfg = config.LoadOrDefault(configPath)
				logging.Debug("Configuration loaded successfully")
			} else {
				logging.Info("Using default configuration")
				cfg = config.LoadDefault()
			}

			// Initialize Git repository detector if not provided
			if detector != nil {
				repoDetector = detector
			} else {
				repoDetector = git.NewRepositoryDetector()
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if showVersion {
				fmt.Fprintf(cmd.OutOrStdout(), "%s version %s\n", version.AppName, version.Version)
				return nil
			}

			// Check if we're in a Git repository
			currentDir, err := os.Getwd()
			if err != nil {
				logging.ErrorWith("Failed to get current directory", map[string]interface{}{
					"error": err.Error(),
				})
				return fmt.Errorf("failed to get current directory: %w", err)
			}

			isRepo, err := repoDetector.IsGitRepository(currentDir)
			if err != nil {
				logging.ErrorWith("Failed to check if directory is a Git repository", map[string]interface{}{
					"error": err.Error(),
				})
				return fmt.Errorf("failed to check if directory is a Git repository: %w", err)
			}

			if !isRepo {
				logging.Error("Not in a Git repository")
				return fmt.Errorf("not in a Git repository")
			}

			repoRoot, err := repoDetector.GetRepositoryRoot(currentDir)
			if err != nil {
				logging.ErrorWith("Failed to get repository root", map[string]interface{}{
					"error": err.Error(),
				})
				return fmt.Errorf("failed to get repository root: %w", err)
			}

			logging.InfoWith("Git repository detected", map[string]interface{}{
				"root": repoRoot,
			})

			// Get files to review based on the all flag
			var filesToReview interface{}
			if all {
				// Get unified list of changed files (both staged and unstaged)
				unifiedFiles, err := repoDetector.GetUnifiedChangedFiles(currentDir, cfg)
				if err != nil {
					logging.ErrorWith("Failed to get changed files", map[string]interface{}{
						"error": err.Error(),
					})
					return fmt.Errorf("failed to get changed files: %w", err)
				}

				logging.InfoWith("Found changed files", map[string]interface{}{
					"count": len(unifiedFiles),
				})

				filesToReview = unifiedFiles
			} else {
				// Get staged files
				stagedFiles, err := repoDetector.GetStagedFiles(currentDir, cfg)
				if err != nil {
					logging.ErrorWith("Failed to get staged files", map[string]interface{}{
						"error": err.Error(),
					})
					return fmt.Errorf("failed to get staged files: %w", err)
				}

				logging.InfoWith("Found staged files", map[string]interface{}{
					"count": len(stagedFiles),
				})

				filesToReview = stagedFiles
			}

			// Display parsed flags and configuration
			logging.InfoWith("Command flags", map[string]interface{}{
				"config": configPath,
				"output": outputPath,
				"debug":  debug,
				"all":    all,
			})

			fmt.Fprintf(cmd.OutOrStdout(), "Git repository: %s\n", repoRoot)
			fmt.Fprintf(cmd.OutOrStdout(), "Configuration file: %s\n", configPath)
			if outputPath != "" {
				fmt.Fprintf(cmd.OutOrStdout(), "Output file: %s\n", outputPath)
			}
			if debug {
				fmt.Fprintf(cmd.OutOrStdout(), "Debug mode: enabled\n")

				// In debug mode, print the loaded configuration
				fmt.Fprintf(cmd.OutOrStdout(), "\nLoaded Configuration:\n")
				fmt.Fprintf(cmd.OutOrStdout(), "  Extensions: %v\n", cfg.Extensions)
				fmt.Fprintf(cmd.OutOrStdout(), "  LLM Provider: %s\n", cfg.LLM.Provider)
				fmt.Fprintf(cmd.OutOrStdout(), "  LLM API URL: %s\n", cfg.LLM.APIURL)
				fmt.Fprintf(cmd.OutOrStdout(), "  LLM Model: %s\n", cfg.LLM.Model)
				fmt.Fprintf(cmd.OutOrStdout(), "  LLM Timeout: %d seconds\n", cfg.LLM.Timeout)
				fmt.Fprintf(cmd.OutOrStdout(), "  Max Concurrent Tasks: %d\n", cfg.Concurrency.MaxTasks)

				logging.DebugWith("Configuration details", map[string]interface{}{
					"extensions": cfg.Extensions,
					"provider":   cfg.LLM.Provider,
					"api_url":    cfg.LLM.APIURL,
					"model":      cfg.LLM.Model,
					"timeout":    cfg.LLM.Timeout,
					"max_tasks":  cfg.Concurrency.MaxTasks,
				})
			}

			// Display files to review
			if all {
				unifiedFiles := filesToReview.([]git.UnifiedChangedFile)
				fmt.Fprintf(cmd.OutOrStdout(), "\nAll changed files to review (%d):\n", len(unifiedFiles))
				for _, file := range unifiedFiles {
					statusStr := ""
					if file.StagedStatus != "" && file.UnstagedStatus != "" {
						// Both staged and unstaged changes
						statusStr = file.StagedStatus + file.UnstagedStatus
					} else if file.StagedStatus != "" {
						// Only staged changes
						statusStr = file.StagedStatus + " (staged)"
					} else if file.UnstagedStatus != "" {
						// Only unstaged changes
						statusStr = file.UnstagedStatus + " (unstaged)"
					}
					fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", statusStr, file.Path)
				}

				// Perform code review if there are files to review
				if len(unifiedFiles) > 0 {
					// Generate combined diff for all files
					combinedDiff := ""

					// Create a map to store file contents
					fileContents := make(map[string]string)

					for _, file := range unifiedFiles {
						diff, err := repoDetector.GetUnifiedFileDiff(repoRoot, file)
						if err != nil {
							logging.ErrorWith("Failed to get diff for file", map[string]interface{}{
								"file":  file.Path,
								"error": err.Error(),
							})
							continue
						}

						// Get the full file content
						content, err := repoDetector.GetFileContent(repoRoot, file.Path)
						if err != nil {
							logging.WarnWith("Failed to get file content, will proceed with diff only", map[string]interface{}{
								"file":  file.Path,
								"error": err.Error(),
							})
						} else {
							fileContents[file.Path] = content
						}

						combinedDiff += fmt.Sprintf("File: %s\n", file.Path)
						combinedDiff += diff
						combinedDiff += "\n\n"
					}

					// Create LLM provider based on configuration
					provider, err := llm.CreateProviderFromConfig(cfg)
					if err != nil {
						logging.ErrorWith("Failed to create LLM provider", map[string]interface{}{
							"error": err.Error(),
						})
						return fmt.Errorf("failed to create LLM provider: %w", err)
					}

					// Generate prompt for code review
					reviewPrompt := prompt.GenerateReviewPrompt(combinedDiff, fileContents, cfg.LLM.Provider)

					// Get response from LLM
					logging.Info("Sending code review request to LLM...")
					fmt.Fprintf(cmd.OutOrStdout(), "\nSending code review request to LLM...\n")

					response, err := provider.GetCompletion(reviewPrompt)
					if err != nil {
						logging.ErrorWith("Failed to get LLM response", map[string]interface{}{
							"error": err.Error(),
						})
						return fmt.Errorf("failed to get LLM response: %w", err)
					}

					// Parse the response
					reviewResult := parse.ParseReview(response)

					// Format and display the results
					formatter := output.NewTerminalFormatter(!noColor)
					formattedReview := formatter.FormatReview(reviewResult)

					// Write to output file if specified
					if outputPath != "" {
						err := os.WriteFile(outputPath, []byte(formattedReview), 0644)
						if err != nil {
							logging.ErrorWith("Failed to write output to file", map[string]interface{}{
								"path":  outputPath,
								"error": err.Error(),
							})
							return fmt.Errorf("failed to write output to file: %w", err)
						}
						fmt.Fprintf(cmd.OutOrStdout(), "Review results written to %s\n", outputPath)
					}

					// Generate markdown report if specified
					if markdownPath != "" {
						// Get repository name from the path
						repoName := filepath.Base(repoRoot)

						// Create markdown formatter
						mdFormatter := output.NewMarkdownFormatter()

						// Write markdown report for all files
						for _, file := range unifiedFiles {
							mdOutputPath := markdownPath

							// If multiple files, create separate reports with file names
							if len(unifiedFiles) > 1 {
								ext := filepath.Ext(markdownPath)
								basePath := strings.TrimSuffix(markdownPath, ext)
								mdOutputPath = fmt.Sprintf("%s_%s%s", basePath, filepath.Base(file.Path), ext)
							}

							err := mdFormatter.WriteToFile(reviewResult, file.Path, repoName, mdOutputPath)
							if err != nil {
								logging.ErrorWith("Failed to write markdown report", map[string]interface{}{
									"path":  mdOutputPath,
									"error": err.Error(),
								})
								return fmt.Errorf("failed to write markdown report: %w", err)
							}
							fmt.Fprintf(cmd.OutOrStdout(), "Markdown report written to %s\n", mdOutputPath)
						}
					}

					// Display the results
					fmt.Fprintln(cmd.OutOrStdout(), "\nCode Review Results:")
					fmt.Fprintln(cmd.OutOrStdout(), formattedReview)
				}
			} else {
				stagedFiles := filesToReview.([]git.StagedFile)
				fmt.Fprintf(cmd.OutOrStdout(), "\nStaged files to review (%d):\n", len(stagedFiles))
				for _, file := range stagedFiles {
					fmt.Fprintf(cmd.OutOrStdout(), "  %s %s\n", file.Status, file.Path)
				}

				// Perform code review if there are files to review
				if len(stagedFiles) > 0 {
					// Generate combined diff for all files
					combinedDiff := ""

					// Create a map to store file contents
					fileContents := make(map[string]string)

					for _, file := range stagedFiles {
						diff, err := repoDetector.GetFileDiff(repoRoot, file.Path, true)
						if err != nil {
							logging.ErrorWith("Failed to get diff for file", map[string]interface{}{
								"file":  file.Path,
								"error": err.Error(),
							})
							continue
						}

						// Get the full file content
						content, err := repoDetector.GetFileContent(repoRoot, file.Path)
						if err != nil {
							logging.WarnWith("Failed to get file content, will proceed with diff only", map[string]interface{}{
								"file":  file.Path,
								"error": err.Error(),
							})
						} else {
							fileContents[file.Path] = content
						}

						combinedDiff += fmt.Sprintf("File: %s\n", file.Path)
						combinedDiff += diff
						combinedDiff += "\n\n"
					}

					// Create LLM provider based on configuration
					provider, err := llm.CreateProviderFromConfig(cfg)
					if err != nil {
						logging.ErrorWith("Failed to create LLM provider", map[string]interface{}{
							"error": err.Error(),
						})
						return fmt.Errorf("failed to create LLM provider: %w", err)
					}

					// Generate prompt for code review
					reviewPrompt := prompt.GenerateReviewPrompt(combinedDiff, fileContents, cfg.LLM.Provider)

					// Get response from LLM
					logging.Info("Sending code review request to LLM...")
					fmt.Fprintf(cmd.OutOrStdout(), "\nSending code review request to LLM...\n")

					response, err := provider.GetCompletion(reviewPrompt)
					if err != nil {
						logging.ErrorWith("Failed to get LLM response", map[string]interface{}{
							"error": err.Error(),
						})
						return fmt.Errorf("failed to get LLM response: %w", err)
					}

					// Parse the response
					reviewResult := parse.ParseReview(response)

					// Format and display the results
					formatter := output.NewTerminalFormatter(!noColor)
					formattedReview := formatter.FormatReview(reviewResult)

					// Write to output file if specified
					if outputPath != "" {
						err := os.WriteFile(outputPath, []byte(formattedReview), 0644)
						if err != nil {
							logging.ErrorWith("Failed to write output to file", map[string]interface{}{
								"path":  outputPath,
								"error": err.Error(),
							})
							return fmt.Errorf("failed to write output to file: %w", err)
						}
						fmt.Fprintf(cmd.OutOrStdout(), "Review results written to %s\n", outputPath)
					}

					// Generate markdown report if specified
					if markdownPath != "" {
						// Get repository name from the path
						repoName := filepath.Base(repoRoot)

						// Create markdown formatter
						mdFormatter := output.NewMarkdownFormatter()

						// Write markdown report for all files
						for _, file := range stagedFiles {
							mdOutputPath := markdownPath

							// If multiple files, create separate reports with file names
							if len(stagedFiles) > 1 {
								ext := filepath.Ext(markdownPath)
								basePath := strings.TrimSuffix(markdownPath, ext)
								mdOutputPath = fmt.Sprintf("%s_%s%s", basePath, filepath.Base(file.Path), ext)
							}

							err := mdFormatter.WriteToFile(reviewResult, file.Path, repoName, mdOutputPath)
							if err != nil {
								logging.ErrorWith("Failed to write markdown report", map[string]interface{}{
									"path":  mdOutputPath,
									"error": err.Error(),
								})
								return fmt.Errorf("failed to write markdown report: %w", err)
							}
							fmt.Fprintf(cmd.OutOrStdout(), "Markdown report written to %s\n", mdOutputPath)
						}
					}

					// Display the results
					fmt.Fprintln(cmd.OutOrStdout(), "\nCode Review Results:")
					fmt.Fprintln(cmd.OutOrStdout(), formattedReview)
				}
			}

			// Actual functionality will be implemented in future steps
			logging.Info("No LLM review functionality implemented yet. This is just a CLI framework with configuration, logging, and Git repository detection support.")
			fmt.Fprintf(cmd.OutOrStdout(), "\nNo LLM review functionality implemented yet. This is just a CLI framework with configuration, logging, and Git repository detection support.\n")
			return nil
		},
	}

	// Add flags
	rootCmd.Flags().StringVarP(&configPath, "config", "c", ".git-llm-reviewer.yaml", "Specify configuration file path")
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "", "Specify output file path")
	rootCmd.Flags().StringVarP(&markdownPath, "markdown", "m", "", "Specify markdown report output file path")
	rootCmd.Flags().BoolVarP(&debug, "debug", "d", false, "Enable verbose output")
	rootCmd.Flags().BoolVarP(&all, "all", "a", false, "Check all changes (not just staged)")
	rootCmd.Flags().BoolVarP(&showVersion, "version", "v", false, "Display version information")
	rootCmd.Flags().BoolVarP(&noColor, "no-color", "", false, "Disable color output")
	rootCmd.Flags().BoolVarP(&logPrompts, "log-prompts", "x", false, "Log prompts to prompt.log for debugging")
	rootCmd.Flags().BoolVar(&logFullExchange, "log-full-exchange", false, "Log both prompts and raw LLM responses to exchange.log")

	return rootCmd
}
