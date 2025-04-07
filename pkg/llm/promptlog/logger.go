package promptlog

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	// DefaultLogFile is the default file path for prompt logging
	DefaultLogFile = "prompt.log"
)

// Logger is responsible for logging prompts to a file
type Logger struct {
	file    *os.File
	enabled bool
	mu      sync.Mutex
}

// NewLogger creates a new prompt logger
func NewLogger(enabled bool, logFile string) (*Logger, error) {
	if !enabled {
		return &Logger{enabled: false}, nil
	}

	// If no log file is specified, use the default
	if logFile == "" {
		logFile = DefaultLogFile
	}

	// Create directories if they don't exist
	dir := filepath.Dir(logFile)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create log directory: %w", err)
		}
	}

	// Open log file in append mode, create if it doesn't exist
	file, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	return &Logger{
		file:    file,
		enabled: true,
	}, nil
}

// Close closes the log file
func (l *Logger) Close() error {
	if l.file != nil {
		return l.file.Close()
	}
	return nil
}

// LogPrompt logs a prompt to the log file
func (l *Logger) LogPrompt(provider, filePath, prompt string) error {
	if !l.enabled || l.file == nil {
		return nil
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Format the log entry
	timestamp := time.Now().Format(time.RFC3339)
	header := fmt.Sprintf("\n\n===== PROMPT LOG [%s] =====\n", timestamp)
	metadata := fmt.Sprintf("Provider: %s\nFile: %s\n\n", provider, filePath)
	footer := "\n===== END PROMPT LOG =====\n"

	// Write the log entry
	if _, err := l.file.WriteString(header); err != nil {
		return err
	}
	if _, err := l.file.WriteString(metadata); err != nil {
		return err
	}
	if _, err := l.file.WriteString(prompt); err != nil {
		return err
	}
	if _, err := l.file.WriteString(footer); err != nil {
		return err
	}

	return l.file.Sync()
}

// Global logger instance
var globalLogger *Logger = &Logger{enabled: false}
var once sync.Once

// InitGlobalLogger initializes the global prompt logger
func InitGlobalLogger(enabled bool, logFile string) error {
	var err error
	once.Do(func() {
		globalLogger, err = NewLogger(enabled, logFile)
	})
	return err
}

// GetGlobalLogger returns the global prompt logger
func GetGlobalLogger() *Logger {
	return globalLogger
}

// LogPrompt logs a prompt using the global logger
func LogPrompt(provider, filePath, prompt string) error {
	return globalLogger.LogPrompt(provider, filePath, prompt)
}
