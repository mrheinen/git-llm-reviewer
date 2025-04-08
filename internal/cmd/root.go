package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/niels/git-llm-review/pkg/config"
	"github.com/niels/git-llm-review/pkg/git"
	"github.com/niels/git-llm-review/pkg/logging"
	"github.com/niels/git-llm-review/pkg/version"
	"github.com/niels/git-llm-review/pkg/workflow"
	"github.com/spf13/cobra"
)

var (
	configPath      string
	outputDir       string
	debug           bool
	all             bool
	showVersion     bool
	providerName    string
	verbose         bool
	logPrompts      bool
	logFullExchange bool
	cfg             *config.Config
	repoDetector    git.RepositoryDetector
)

// NewRootCmd creates the root command for git-llm-review
func NewRootCmd() *cobra.Command {
	return NewRootCmdWithDetector(git.NewRepositoryDetector())
}

// NewRootCmdWithDetector creates the root command with a custom repository detector
// This is primarily used for testing
func NewRootCmdWithDetector(detector git.RepositoryDetector) *cobra.Command {
	repoDetector = detector
	
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
			logging.Info("Initializing git-llm-review")
			
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
				// Use default configuration
				logging.Info("Using default configuration")
				cfg = config.Default()
			}
			
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			// Check if we should just show the version
			if showVersion {
				fmt.Fprintln(cmd.OutOrStdout(), version.GetVersionInfo())
				return nil
			}
			
			// Create workflow options
			options := workflow.Options{
				ConfigPath:      configPath,
				All:             all,
				OutputFormat:    "terminal",
				OutputPath:      outputDir,
				ProviderName:    providerName,
				VerboseOutput:   verbose,
				LogPrompts:      logPrompts,
				LogFullExchange: logFullExchange,
			}
			
			// Create and run workflow
			reviewWorkflow, err := workflow.NewReviewWorkflow(options)
			if err != nil {
				logging.ErrorWith("Failed to create review workflow", map[string]interface{}{
					"error": err.Error(),
				})
				return fmt.Errorf("failed to create review workflow: %w", err)
			}
			
			// Run the workflow
			ctx := context.Background()
			stats, err := reviewWorkflow.Run(ctx)
			if err != nil {
				logging.ErrorWith("Review workflow failed", map[string]interface{}{
					"error": err.Error(),
				})
				return fmt.Errorf("review workflow failed: %w", err)
			}
			
			// Log completion
			logging.InfoWith("Review workflow completed", map[string]interface{}{
				"files_processed":   stats.FilesProcessed,
				"files_with_errors": stats.FilesWithErrors,
				"total_issues":      stats.TotalIssues,
				"duration":          stats.Duration.String(),
			})
			
			return nil
		},
	}
	
	// Add flags
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "", "Path to configuration file")
	rootCmd.PersistentFlags().StringVarP(&outputDir, "output-dir", "o", "", "Directory to write review reports")
	rootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "Enable debug mode")
	rootCmd.PersistentFlags().BoolVarP(&all, "all", "a", false, "Review all changed files (both staged and unstaged)")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "v", false, "Show version information")
	rootCmd.PersistentFlags().StringVarP(&providerName, "provider", "p", "", "LLM provider to use (overrides config)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&logPrompts, "log-prompts", "x", false, "Log prompts to prompt.log for debugging")
	rootCmd.PersistentFlags().BoolVar(&logFullExchange, "log-full-exchange", false, "Log both prompts and raw LLM responses to exchange.log")
	
	return rootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := NewRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}
