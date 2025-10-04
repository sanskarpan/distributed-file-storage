package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

// String returns the string representation of the log level
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

// Logger represents a structured logger
type Logger struct {
	level  LogLevel
	output io.Writer
	prefix string
}

// New creates a new logger with the specified level and output
func New(level LogLevel, output io.Writer, prefix string) *Logger {
	return &Logger{
		level:  level,
		output: output,
		prefix: prefix,
	}
}

// NewDefault creates a default logger that writes to stdout
func NewDefault() *Logger {
	return New(INFO, os.Stdout, "")
}

// SetLevel sets the minimum log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// SetOutput sets the output destination
func (l *Logger) SetOutput(output io.Writer) {
	l.output = output
}

// SetPrefix sets the logger prefix
func (l *Logger) SetPrefix(prefix string) {
	l.prefix = prefix
}

// log writes a log message with the specified level
func (l *Logger) log(level LogLevel, format string, args ...interface{}) {
	if level < l.level {
		return
	}

	// Get caller information
	_, file, line, ok := runtime.Caller(2)
	caller := "unknown"
	if ok {
		// Extract just the filename
		parts := strings.Split(file, "/")
		if len(parts) > 0 {
			caller = fmt.Sprintf("%s:%d", parts[len(parts)-1], line)
		}
	}

	// Format timestamp
	timestamp := time.Now().Format("2006-01-02 15:04:05.000")

	// Build log message
	message := fmt.Sprintf(format, args...)
	
	var logLine string
	if l.prefix != "" {
		logLine = fmt.Sprintf("[%s] %s [%s] [%s] %s\n", 
			timestamp, level.String(), l.prefix, caller, message)
	} else {
		logLine = fmt.Sprintf("[%s] %s [%s] %s\n", 
			timestamp, level.String(), caller, message)
	}

	// Write to output
	l.output.Write([]byte(logLine))

	// If it's a fatal error, exit the program
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(format string, args ...interface{}) {
	l.log(DEBUG, format, args...)
}

// Info logs an info message
func (l *Logger) Info(format string, args ...interface{}) {
	l.log(INFO, format, args...)
}

// Warn logs a warning message
func (l *Logger) Warn(format string, args ...interface{}) {
	l.log(WARN, format, args...)
}

// Error logs an error message
func (l *Logger) Error(format string, args ...interface{}) {
	l.log(ERROR, format, args...)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(format string, args ...interface{}) {
	l.log(FATAL, format, args...)
}

// WithPrefix returns a new logger with the specified prefix
func (l *Logger) WithPrefix(prefix string) *Logger {
	return &Logger{
		level:  l.level,
		output: l.output,
		prefix: prefix,
	}
}

// Global logger instance
var defaultLogger = NewDefault()

// SetGlobalLevel sets the global logger level
func SetGlobalLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// SetGlobalOutput sets the global logger output
func SetGlobalOutput(output io.Writer) {
	defaultLogger.SetOutput(output)
}

// Debug logs a debug message using the global logger
func Debug(format string, args ...interface{}) {
	defaultLogger.Debug(format, args...)
}

// Info logs an info message using the global logger
func Info(format string, args ...interface{}) {
	defaultLogger.Info(format, args...)
}

// Warn logs a warning message using the global logger
func Warn(format string, args ...interface{}) {
	defaultLogger.Warn(format, args...)
}

// Error logs an error message using the global logger
func Error(format string, args ...interface{}) {
	defaultLogger.Error(format, args...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(format string, args ...interface{}) {
	defaultLogger.Fatal(format, args...)
}

// WithPrefix returns a new logger with the specified prefix using the global logger
func WithPrefix(prefix string) *Logger {
	return defaultLogger.WithPrefix(prefix)
}

// Compatibility with standard log package
func init() {
	// Disable standard log package timestamps since we handle them
	log.SetFlags(0)
}
