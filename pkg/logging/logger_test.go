package logging

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestLogLevels(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	
	// Initialize logger with debug enabled and custom writer
	logger := NewLogger(true, &buf)
	
	// Test debug level (should be visible with debug enabled)
	logger.Debug().Msg("debug message")
	output := buf.String()
	buf.Reset()
	
	if !strings.Contains(output, "debug message") {
		t.Errorf("Debug log should contain 'debug message', got: %s", output)
	}
	if !strings.Contains(output, `"level":"debug"`) {
		t.Errorf("Debug log should have debug level, got: %s", output)
	}
	
	// Test info level
	logger.Info().Msg("info message")
	output = buf.String()
	buf.Reset()
	
	if !strings.Contains(output, "info message") {
		t.Errorf("Info log should contain 'info message', got: %s", output)
	}
	if !strings.Contains(output, `"level":"info"`) {
		t.Errorf("Info log should have info level, got: %s", output)
	}
	
	// Test warn level
	logger.Warn().Msg("warn message")
	output = buf.String()
	buf.Reset()
	
	if !strings.Contains(output, "warn message") {
		t.Errorf("Warn log should contain 'warn message', got: %s", output)
	}
	if !strings.Contains(output, `"level":"warn"`) {
		t.Errorf("Warn log should have warn level, got: %s", output)
	}
	
	// Test error level
	logger.Error().Msg("error message")
	output = buf.String()
	buf.Reset()
	
	if !strings.Contains(output, "error message") {
		t.Errorf("Error log should contain 'error message', got: %s", output)
	}
	if !strings.Contains(output, `"level":"error"`) {
		t.Errorf("Error log should have error level, got: %s", output)
	}
}

func TestDebugDisabled(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	
	// Initialize logger with debug disabled
	logger := NewLogger(false, &buf)
	
	// Debug messages should not appear
	logger.Debug().Msg("debug message")
	output := buf.String()
	buf.Reset()
	
	if strings.Contains(output, "debug message") {
		t.Errorf("Debug log should not be visible when debug is disabled, got: %s", output)
	}
	
	// Other levels should still appear
	logger.Info().Msg("info message")
	output = buf.String()
	buf.Reset()
	
	if !strings.Contains(output, "info message") {
		t.Errorf("Info log should be visible when debug is disabled, got: %s", output)
	}
}

func TestStructuredLogging(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	
	// Initialize logger with debug enabled
	logger := NewLogger(true, &buf)
	
	// Log with structured fields
	logger.Info().
		Str("file", "main.go").
		Int("line", 42).
		Msg("structured log message")
	
	output := buf.String()
	buf.Reset()
	
	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v", err)
	}
	
	// Check structured fields
	if file, ok := logEntry["file"].(string); !ok || file != "main.go" {
		t.Errorf("Expected file field to be 'main.go', got: %v", logEntry["file"])
	}
	
	if line, ok := logEntry["line"].(float64); !ok || int(line) != 42 {
		t.Errorf("Expected line field to be 42, got: %v", logEntry["line"])
	}
	
	if msg, ok := logEntry["message"].(string); !ok || msg != "structured log message" {
		t.Errorf("Expected message to be 'structured log message', got: %v", logEntry["message"])
	}
}

func TestContextualLogging(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	
	// Initialize logger with debug enabled
	logger := NewLogger(true, &buf)
	
	// Create a sub-logger with context
	contextLogger := logger.With().
		Str("component", "git").
		Str("operation", "diff").
		Logger()
	
	// Log with the contextual logger
	contextLogger.Info().Msg("contextual log message")
	
	output := buf.String()
	buf.Reset()
	
	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v", err)
	}
	
	// Check contextual fields
	if component, ok := logEntry["component"].(string); !ok || component != "git" {
		t.Errorf("Expected component field to be 'git', got: %v", logEntry["component"])
	}
	
	if operation, ok := logEntry["operation"].(string); !ok || operation != "diff" {
		t.Errorf("Expected operation field to be 'diff', got: %v", logEntry["operation"])
	}
}

func TestHelperFunctions(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	
	// Use a direct approach: create a logger with our buffer and set it as the global logger
	testLogger := NewLogger(true, &buf)
	globalLogger = testLogger
	
	// Test helper functions
	Debug("debug helper message")
	output := buf.String()
	buf.Reset()
	
	if !strings.Contains(output, "debug helper message") {
		t.Errorf("Debug helper should log 'debug helper message', got: %s", output)
	}
	
	Info("info helper message")
	output = buf.String()
	buf.Reset()
	
	if !strings.Contains(output, "info helper message") {
		t.Errorf("Info helper should log 'info helper message', got: %s", output)
	}
	
	Warn("warn helper message")
	output = buf.String()
	buf.Reset()
	
	if !strings.Contains(output, "warn helper message") {
		t.Errorf("Warn helper should log 'warn helper message', got: %s", output)
	}
	
	Error("error helper message")
	output = buf.String()
	buf.Reset()
	
	if !strings.Contains(output, "error helper message") {
		t.Errorf("Error helper should log 'error helper message', got: %s", output)
	}
}

func TestTimestampFormat(t *testing.T) {
	// Capture log output
	var buf bytes.Buffer
	
	// Initialize logger with debug enabled
	logger := NewLogger(true, &buf)
	
	// Log a message
	logger.Info().Msg("timestamp test")
	
	output := buf.String()
	buf.Reset()
	
	// Parse JSON output
	var logEntry map[string]interface{}
	if err := json.Unmarshal([]byte(output), &logEntry); err != nil {
		t.Fatalf("Failed to parse log output as JSON: %v", err)
	}
	
	// Check timestamp format (should be RFC3339)
	if timestamp, ok := logEntry["time"].(string); ok {
		_, err := time.Parse(time.RFC3339, timestamp)
		if err != nil {
			t.Errorf("Timestamp should be in RFC3339 format, got: %s", timestamp)
		}
	} else {
		t.Errorf("Log entry should contain a timestamp field")
	}
}
