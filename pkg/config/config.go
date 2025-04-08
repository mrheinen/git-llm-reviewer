package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Extensions  []string         `yaml:"extensions"`
	LLM         LLMConfig        `yaml:"llm"`
	Concurrency ConcurrencyConfig `yaml:"concurrency"`
	Retry       RetryConfig      `yaml:"retry"`
	Logging     LogConfig        `yaml:"logging"`
}

// LLMConfig contains settings for the LLM provider
type LLMConfig struct {
	Provider string `yaml:"provider"`
	APIURL   string `yaml:"api_url"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
	Timeout  int    `yaml:"timeout"` // in seconds
}

// ConcurrencyConfig contains settings for concurrency control
type ConcurrencyConfig struct {
	MaxTasks int `yaml:"max_tasks"`
}

// RetryConfig contains settings for retry behavior
type RetryConfig struct {
	Enabled       bool     `yaml:"enabled"`
	MaxRetries    int      `yaml:"max_retries"`
	InitialDelay  int      `yaml:"initial_delay"` // in milliseconds
	MaxDelay      int      `yaml:"max_delay"`     // in milliseconds
	BackoffFactor float64  `yaml:"backoff_factor"`
	JitterFactor  float64  `yaml:"jitter_factor"`
	RetryableErrors []string `yaml:"retryable_errors"`
}

// LogConfig contains settings for logging
type LogConfig struct {
	LogToFile   bool   `yaml:"log_to_file"`
	LogFilePath string `yaml:"log_file_path"`
	MaxSize     int    `yaml:"max_size"`      // maximum size in megabytes
	MaxBackups  int    `yaml:"max_backups"`   // maximum number of old log files to retain
	MaxAge      int    `yaml:"max_age"`       // maximum number of days to retain old log files
	Compress    bool   `yaml:"compress"`      // compress determines if the rotated log files should be compressed
	PromptLogPath string `yaml:"prompt_log_path"` // path to the file where prompts will be logged
}

// LoadDefault returns a configuration with default values
func LoadDefault() *Config {
	return &Config{
		Extensions: []string{".go", ".c", ".cc", ".proto", ".vue"},
		LLM: LLMConfig{
			Provider: "openai",
			APIURL:   "https://api.openai.com/v1",
			APIKey:   "",
			Model:    "gpt-4",
			Timeout:  300, // 5 minutes
		},
		Concurrency: ConcurrencyConfig{
			MaxTasks: 5,
		},
		Retry: RetryConfig{
			Enabled:        true,
			MaxRetries:     3,
			InitialDelay:   500,
			MaxDelay:       5000,
			BackoffFactor:  2.0,
			JitterFactor:   0.1,
			RetryableErrors: []string{
				"rate limit",
				"timeout",
				"connection reset",
				"connection refused",
				"no response",
				"internal server error",
			},
		},
		Logging: LogConfig{
			LogToFile:     false,
			LogFilePath:   "git-llm-review.log",
			MaxSize:       10,
			MaxBackups:    3,
			MaxAge:        28,
			Compress:      true,
			PromptLogPath: "prompt.log",
		},
	}
}

// Default returns a configuration with default values
// This is an alias for LoadDefault for backward compatibility
func Default() *Config {
	return LoadDefault()
}

// Load reads configuration from a file and merges it with default values
func Load(configPath string) (*Config, error) {
	// Start with default configuration
	cfg := LoadDefault()

	// Read configuration file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Create a temporary config to parse the file
	var fileCfg Config
	if err := yaml.Unmarshal(data, &fileCfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge file configuration with defaults
	if len(fileCfg.Extensions) > 0 {
		cfg.Extensions = fileCfg.Extensions
	}

	// Merge LLM configuration
	if fileCfg.LLM.Provider != "" {
		cfg.LLM.Provider = fileCfg.LLM.Provider
	}
	if fileCfg.LLM.APIURL != "" {
		cfg.LLM.APIURL = fileCfg.LLM.APIURL
	}
	if fileCfg.LLM.APIKey != "" {
		cfg.LLM.APIKey = fileCfg.LLM.APIKey
	}
	if fileCfg.LLM.Model != "" {
		cfg.LLM.Model = fileCfg.LLM.Model
	}
	if fileCfg.LLM.Timeout > 0 {
		cfg.LLM.Timeout = fileCfg.LLM.Timeout
	}
	
	// Check for LLM_API_KEY environment variable to override the API key
	if envKey := os.Getenv("LLM_API_KEY"); envKey != "" {
		cfg.LLM.APIKey = envKey
	}

	// Merge concurrency configuration
	if fileCfg.Concurrency.MaxTasks > 0 {
		cfg.Concurrency.MaxTasks = fileCfg.Concurrency.MaxTasks
	}

	// Merge retry configuration
	if fileCfg.Retry.Enabled {
		cfg.Retry.Enabled = fileCfg.Retry.Enabled
	}
	if fileCfg.Retry.MaxRetries > 0 {
		cfg.Retry.MaxRetries = fileCfg.Retry.MaxRetries
	}
	if fileCfg.Retry.InitialDelay > 0 {
		cfg.Retry.InitialDelay = fileCfg.Retry.InitialDelay
	}
	if fileCfg.Retry.MaxDelay > 0 {
		cfg.Retry.MaxDelay = fileCfg.Retry.MaxDelay
	}
	if fileCfg.Retry.BackoffFactor > 0 {
		cfg.Retry.BackoffFactor = fileCfg.Retry.BackoffFactor
	}
	if fileCfg.Retry.JitterFactor > 0 {
		cfg.Retry.JitterFactor = fileCfg.Retry.JitterFactor
	}
	if len(fileCfg.Retry.RetryableErrors) > 0 {
		cfg.Retry.RetryableErrors = fileCfg.Retry.RetryableErrors
	}

	// Merge logging configuration
	if fileCfg.Logging.LogToFile {
		cfg.Logging.LogToFile = fileCfg.Logging.LogToFile
	}
	if fileCfg.Logging.LogFilePath != "" {
		cfg.Logging.LogFilePath = fileCfg.Logging.LogFilePath
	}
	if fileCfg.Logging.MaxSize > 0 {
		cfg.Logging.MaxSize = fileCfg.Logging.MaxSize
	}
	if fileCfg.Logging.MaxBackups > 0 {
		cfg.Logging.MaxBackups = fileCfg.Logging.MaxBackups
	}
	if fileCfg.Logging.MaxAge > 0 {
		cfg.Logging.MaxAge = fileCfg.Logging.MaxAge
	}
	if fileCfg.Logging.Compress {
		cfg.Logging.Compress = fileCfg.Logging.Compress
	}
	if fileCfg.Logging.PromptLogPath != "" {
		cfg.Logging.PromptLogPath = fileCfg.Logging.PromptLogPath
	}

	return cfg, nil
}

// LoadOrDefault attempts to load configuration from a file
// If the file doesn't exist or can't be parsed, it returns default configuration
func LoadOrDefault(configPath string) *Config {
	cfg, err := Load(configPath)
	if err != nil {
		// Log the error but continue with defaults
		fmt.Fprintf(os.Stderr, "Warning: Failed to load config from %s: %v\n", configPath, err)
		fmt.Fprintf(os.Stderr, "Using default configuration\n")
		cfg = LoadDefault()
		
		// Even with default config, check for LLM_API_KEY environment variable
		if envKey := os.Getenv("LLM_API_KEY"); envKey != "" {
			cfg.LLM.APIKey = envKey
		}
	}
	return cfg
}
