package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
)

// LogLevel represents the logging level
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp  string                 `json:"timestamp"`
	Level      string                 `json:"level"`
	Logger     string                 `json:"logger"`
	Message    string                 `json:"message"`
	Module     string                 `json:"module,omitempty"`
	Function   string                 `json:"function,omitempty"`
	Line       int                    `json:"line,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	Method     string                 `json:"method,omitempty"`
	Path       string                 `json:"path,omitempty"`
	StatusCode int                    `json:"status_code,omitempty"`
	Duration   float64                `json:"duration,omitempty"`
	ClientIP   string                 `json:"client_ip,omitempty"`
	UserAgent  string                 `json:"user_agent,omitempty"`
	UserID     string                 `json:"user_id,omitempty"`
	Error      string                 `json:"error,omitempty"`
	ErrorType  string                 `json:"error_type,omitempty"`
	Extra      map[string]interface{} `json:"extra,omitempty"`
}

// Logger represents a structured logger
type Logger struct {
	name         string
	level        LogLevel
	logFile      *os.File
	appFile      *os.File
	errorFile    *os.File
	perfFile     *os.File
	businessFile *os.File
}

var (
	// Global logger instance
	globalLogger *Logger
)

// NewLogger creates a new structured logger
func NewLogger(name string, level LogLevel, logDir string) (*Logger, error) {
	// Ensure log directory exists
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logger := &Logger{
		name:  name,
		level: level,
	}

	// Open log files
	var err error
	logger.logFile, err = os.OpenFile(
		filepath.Join(logDir, "go_app.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open app log file: %w", err)
	}

	logger.errorFile, err = os.OpenFile(
		filepath.Join(logDir, "go_error.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open error log file: %w", err)
	}

	logger.perfFile, err = os.OpenFile(
		filepath.Join(logDir, "go_performance.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open performance log file: %w", err)
	}

	logger.businessFile, err = os.OpenFile(
		filepath.Join(logDir, "go_business.log"),
		os.O_CREATE|os.O_WRONLY|os.O_APPEND,
		0644,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to open business log file: %w", err)
	}

	return logger, nil
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	globalLogger = logger
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return globalLogger
}

// writeLog writes a log entry to the appropriate file
func (l *Logger) writeLog(level LogLevel, entry LogEntry, file *os.File) {
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	entry.Logger = l.name
	entry.Level = level.String()

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	fmt.Fprintln(file, string(data))
}

// log writes a log entry if the level is appropriate
func (l *Logger) log(level LogLevel, message string, extra map[string]interface{}) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Message: message,
		Extra:   extra,
	}

	switch level {
	case DEBUG, INFO:
		l.writeLog(level, entry, l.logFile)
	case WARN, ERROR, FATAL:
		l.writeLog(level, entry, l.errorFile)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, extra ...map[string]interface{}) {
	extraMap := mergeExtra(extra...)
	l.log(DEBUG, message, extraMap)
}

// Info logs an info message
func (l *Logger) Info(message string, extra ...map[string]interface{}) {
	extraMap := mergeExtra(extra...)
	l.log(INFO, message, extraMap)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, extra ...map[string]interface{}) {
	extraMap := mergeExtra(extra...)
	l.log(WARN, message, extraMap)
}

// Error logs an error message
func (l *Logger) Error(message string, err error, extra ...map[string]interface{}) {
	extraMap := mergeExtra(extra...)
	if err != nil {
		extraMap["error"] = err.Error()
		extraMap["error_type"] = fmt.Sprintf("%T", err)
	}
	l.log(ERROR, message, extraMap)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(message string, err error, extra ...map[string]interface{}) {
	extraMap := mergeExtra(extra...)
	if err != nil {
		extraMap["error"] = err.Error()
		extraMap["error_type"] = fmt.Sprintf("%T", err)
	}
	l.log(FATAL, message, extraMap)
	os.Exit(1)
}

// LogRequest logs a request with context
func (l *Logger) LogRequest(c *fiber.Ctx, duration float64, extra ...map[string]interface{}) {
	extraMap := mergeExtra(extra...)
	extraMap["method"] = c.Method()
	extraMap["path"] = c.Path()
	extraMap["status_code"] = c.Response().StatusCode()
	extraMap["duration"] = duration
	extraMap["client_ip"] = c.IP()
	extraMap["user_agent"] = c.Get("User-Agent")

	// Extract request ID if available
	if requestID := c.Get("X-Request-ID"); requestID != "" {
		extraMap["request_id"] = requestID
	}

	// Extract user ID if available (from JWT claims)
	if userID := c.Locals("user_id"); userID != nil {
		extraMap["user_id"] = userID
	}

	l.Info("Request completed", extraMap)
}

// LogPerformance logs performance metrics
func (l *Logger) LogPerformance(operation string, duration float64, extra ...map[string]interface{}) {
	extraMap := mergeExtra(extra...)
	extraMap["operation"] = operation
	extraMap["duration"] = duration

	entry := LogEntry{
		Message: fmt.Sprintf("Performance: %s", operation),
		Extra:   extraMap,
	}

	l.writeLog(INFO, entry, l.perfFile)
}

// LogBusinessEvent logs business events
func (l *Logger) LogBusinessEvent(event string, extra ...map[string]interface{}) {
	extraMap := mergeExtra(extra...)
	extraMap["event"] = event

	entry := LogEntry{
		Message: fmt.Sprintf("Business Event: %s", event),
		Extra:   extraMap,
	}

	l.writeLog(INFO, entry, l.businessFile)
}

// Close closes all log files
func (l *Logger) Close() error {
	var err error
	if l.logFile != nil {
		if closeErr := l.logFile.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if l.errorFile != nil {
		if closeErr := l.errorFile.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if l.perfFile != nil {
		if closeErr := l.perfFile.Close(); closeErr != nil {
			err = closeErr
		}
	}
	if l.businessFile != nil {
		if closeErr := l.businessFile.Close(); closeErr != nil {
			err = closeErr
		}
	}
	return err
}

// mergeExtra merges multiple extra maps
func mergeExtra(extra ...map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})
	for _, m := range extra {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}

// Convenience functions for global logger
func Debug(message string, extra ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Debug(message, extra...)
	}
}

func Info(message string, extra ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Info(message, extra...)
	}
}

func Warn(message string, extra ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(message, extra...)
	}
}

func Error(message string, err error, extra ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Error(message, err, extra...)
	}
}

func Fatal(message string, err error, extra ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.Fatal(message, err, extra...)
	}
}

// Formatted convenience functions for replacing log.Printf calls
func Infof(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Info(fmt.Sprintf(format, args...))
	}
}

func Warnf(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.Warn(fmt.Sprintf(format, args...))
	}
}

func Errorf(format string, args ...interface{}) {
	if globalLogger != nil {
		globalLogger.log(ERROR, fmt.Sprintf(format, args...), nil)
	}
}

func LogRequest(c *fiber.Ctx, duration float64, extra ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.LogRequest(c, duration, extra...)
	}
}

func LogPerformance(operation string, duration float64, extra ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.LogPerformance(operation, duration, extra...)
	}
}

func LogBusinessEvent(event string, extra ...map[string]interface{}) {
	if globalLogger != nil {
		globalLogger.LogBusinessEvent(event, extra...)
	}
}
