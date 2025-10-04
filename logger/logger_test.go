package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestLogLevels(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
	}

	for _, test := range tests {
		if test.level.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.level.String())
		}
	}
}

func TestLoggerOutput(t *testing.T) {
	var buf bytes.Buffer
	logger := New(DEBUG, &buf, "TEST")

	logger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "INFO") {
		t.Error("Expected INFO level in output")
	}
	if !strings.Contains(output, "test message") {
		t.Error("Expected test message in output")
	}
	if !strings.Contains(output, "TEST") {
		t.Error("Expected TEST prefix in output")
	}
}

func TestLoggerLevelFiltering(t *testing.T) {
	var buf bytes.Buffer
	logger := New(WARN, &buf, "")

	logger.Debug("debug message")
	logger.Info("info message")
	logger.Warn("warn message")

	output := buf.String()
	if strings.Contains(output, "debug message") {
		t.Error("Debug message should be filtered out")
	}
	if strings.Contains(output, "info message") {
		t.Error("Info message should be filtered out")
	}
	if !strings.Contains(output, "warn message") {
		t.Error("Warn message should be included")
	}
}

func TestLoggerWithPrefix(t *testing.T) {
	var buf bytes.Buffer
	logger := New(INFO, &buf, "MAIN")
	
	subLogger := logger.WithPrefix("SUB")
	subLogger.Info("test message")

	output := buf.String()
	if !strings.Contains(output, "SUB") {
		t.Error("Expected SUB prefix in output")
	}
}

func TestGlobalLogger(t *testing.T) {
	var buf bytes.Buffer
	SetGlobalOutput(&buf)
	SetGlobalLevel(INFO)

	Info("global test message")

	output := buf.String()
	if !strings.Contains(output, "global test message") {
		t.Error("Expected global test message in output")
	}
}
