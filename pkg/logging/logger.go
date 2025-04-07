package logging

import (
	"io"
	"os"
	"time"

	"github.com/natefinch/lumberjack"
	"github.com/niels/git-llm-review/pkg/config"
	"github.com/rs/zerolog"
)

var (
	// Global logger instance
	globalLogger zerolog.Logger
)

// InitGlobalLogger initializes the global logger with the specified debug level
func InitGlobalLogger(debug bool, cfg *config.Config) {
	var output io.Writer

	if cfg != nil && cfg.Logging.LogToFile {
		// Configure rotating file logger
		fileLogger := &lumberjack.Logger{
			Filename:   cfg.Logging.LogFilePath,
			MaxSize:    cfg.Logging.MaxSize,    // megabytes
			MaxBackups: cfg.Logging.MaxBackups,
			MaxAge:     cfg.Logging.MaxAge,     // days
			Compress:   cfg.Logging.Compress,
		}
		
		if debug {
			// In debug mode, send logs to both file and stderr
			output = io.MultiWriter(fileLogger, os.Stderr)
			Info("Logging to file and stderr: " + cfg.Logging.LogFilePath)
		} else {
			// In non-debug mode, send logs only to file
			output = fileLogger
			// Log initialization message to both stderr and file just this once
			tempLogger := NewLogger(false, os.Stderr)
			tempLogger.Info().Msg("Logging to file only: " + cfg.Logging.LogFilePath)
		}
	} else {
		// If file logging is disabled, only log to stderr in debug mode
		// otherwise discard logs completely
		if debug {
			output = os.Stderr
		} else {
			output = io.Discard // Send logs to nowhere
		}
	}

	globalLogger = NewLogger(debug, output)
}

// NewLogger creates a new zerolog logger with the specified debug level
func NewLogger(debug bool, output io.Writer) zerolog.Logger {
	// If no output is specified, use stderr
	if output == nil {
		output = os.Stderr
	}

	// Set the global log level based on debug flag
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Configure the logger
	logger := zerolog.New(output).
		With().
		Timestamp().
		Caller().
		Logger()

	return logger
}

// Debug logs a message at debug level
func Debug(msg string) {
	globalLogger.Debug().Msg(msg)
}

// Info logs a message at info level
func Info(msg string) {
	globalLogger.Info().Msg(msg)
}

// Warn logs a message at warn level
func Warn(msg string) {
	globalLogger.Warn().Msg(msg)
}

// Error logs a message at error level
func Error(msg string) {
	globalLogger.Error().Msg(msg)
}

// DebugWith logs a message at debug level with additional context
func DebugWith(msg string, fields map[string]interface{}) {
	event := globalLogger.Debug()
	for k, v := range fields {
		event = addField(event, k, v)
	}
	event.Msg(msg)
}

// InfoWith logs a message at info level with additional context
func InfoWith(msg string, fields map[string]interface{}) {
	event := globalLogger.Info()
	for k, v := range fields {
		event = addField(event, k, v)
	}
	event.Msg(msg)
}

// WarnWith logs a message at warn level with additional context
func WarnWith(msg string, fields map[string]interface{}) {
	event := globalLogger.Warn()
	for k, v := range fields {
		event = addField(event, k, v)
	}
	event.Msg(msg)
}

// ErrorWith logs a message at error level with additional context
func ErrorWith(msg string, fields map[string]interface{}) {
	event := globalLogger.Error()
	for k, v := range fields {
		event = addField(event, k, v)
	}
	event.Msg(msg)
}

// GetLogger returns the global logger instance
func GetLogger() zerolog.Logger {
	return globalLogger
}

// WithComponent returns a logger with the component field set
func WithComponent(component string) zerolog.Logger {
	return globalLogger.With().Str("component", component).Logger()
}

// addField adds a field to the log event based on its type
func addField(event *zerolog.Event, key string, value interface{}) *zerolog.Event {
	switch v := value.(type) {
	case string:
		return event.Str(key, v)
	case int:
		return event.Int(key, v)
	case int64:
		return event.Int64(key, v)
	case float64:
		return event.Float64(key, v)
	case bool:
		return event.Bool(key, v)
	case time.Time:
		return event.Time(key, v)
	case []string:
		return event.Strs(key, v)
	case []int:
		return event.Ints(key, v)
	case []bool:
		return event.Bools(key, v)
	case error:
		return event.Err(v).Str(key, v.Error())
	default:
		return event.Interface(key, v)
	}
}
