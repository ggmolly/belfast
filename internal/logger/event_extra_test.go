package logger

import (
	"bytes"
	"errors"
	"io"
	"os"
	"sync"
	"testing"
)

func resetLogLevelState() {
	logLevelOnce = sync.Once{}
	logLevel = LOG_LEVEL_INFO
}

func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	os.Stdout = writer

	fn()

	_ = writer.Close()
	os.Stdout = original

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, reader)
	_ = reader.Close()
	return buf.String()
}

func TestEntryLogMethods(t *testing.T) {
	resetLogLevelState()
	t.Setenv("LOG_LEVEL", "debug")

	output := captureStdout(t, func() {
		entry := Scope("Test")
		entry.Debug("debug message")
		entry.Warn("warn message")
		entry.Error("error message")
	})

	if output == "" {
		t.Fatalf("expected log output")
	}
}

func TestShouldLogLevelsFromEnv(t *testing.T) {
	resetLogLevelState()
	t.Setenv("LOG_LEVEL", "warning")
	if !shouldLog(LOG_LEVEL_WARN) {
		t.Fatalf("expected warn to be logged with warning level")
	}
	if shouldLog(LOG_LEVEL_DEBUG) {
		t.Fatalf("expected debug to not be logged with warning level")
	}

	resetLogLevelState()
	t.Setenv("LOG_LEVEL", "unknown")
	if !shouldLog(LOG_LEVEL_INFO) {
		t.Fatalf("expected default log level to be info")
	}
}

func TestShouldLogEnvCases(t *testing.T) {
	testCases := []struct {
		name     string
		level    string
		expected int
	}{
		{name: "debug", level: "debug", expected: LOG_LEVEL_DEBUG},
		{name: "info", level: "info", expected: LOG_LEVEL_INFO},
		{name: "warn", level: "warn", expected: LOG_LEVEL_WARN},
		{name: "error", level: "error", expected: LOG_LEVEL_ERROR},
		{name: "default", level: "", expected: LOG_LEVEL_INFO},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			resetLogLevelState()
			t.Setenv("LOG_LEVEL", tt.level)
			_ = shouldLog(LOG_LEVEL_INFO)
			if logLevel != tt.expected {
				t.Fatalf("expected log level %d, got %d", tt.expected, logLevel)
			}
		})
	}
}

func TestLogSkipsWhenBelowLevel(t *testing.T) {
	resetLogLevelState()
	t.Setenv("LOG_LEVEL", "error")

	output := captureStdout(t, func() {
		Log("Test", "message", LOG_LEVEL_WARN)
	})

	if output != "" {
		t.Fatalf("expected no log output, got %q", output)
	}
}

func TestSpanErrorPath(t *testing.T) {
	resetLogLevelState()
	t.Setenv("LOG_LEVEL", "debug")

	output := captureStdout(t, func() {
		span := Span("Test", Field{Key: "key", Value: "value"})
		span(errors.New("boom"))
	})

	if output == "" {
		t.Fatalf("expected span to log an error")
	}
}

func TestLevelColorUnknown(t *testing.T) {
	if levelColor(999) != unknownColor {
		t.Fatalf("expected unknown color for unknown level")
	}
}
