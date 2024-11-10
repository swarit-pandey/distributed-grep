package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"testing"
)

type testHandler struct {
	buf    *bytes.Buffer
	format slog.Handler
}

func newTestHandler(buf *bytes.Buffer) *testHandler {
	return &testHandler{
		buf:    buf,
		format: slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}),
	}
}

func (h *testHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.format.Enabled(ctx, level)
}

func (h *testHandler) Handle(ctx context.Context, r slog.Record) error {
	return h.format.Handle(ctx, r)
}

func (h *testHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &testHandler{h.buf, h.format.WithAttrs(attrs)}
}

func (h *testHandler) WithGroup(name string) slog.Handler {
	return &testHandler{h.buf, h.format.WithGroup(name)}
}

func createTestLogger(buf *bytes.Buffer) *Logger {
	handler := &customHandler{newTestHandler(buf)}
	return &Logger{
		Logger: slog.New(handler),
		level:  INFO,
	}
}

func parseLogOutput(output string) (map[string]interface{}, error) {
	var result map[string]interface{}
	err := json.Unmarshal([]byte(output), &result)
	return result, err
}

func TestLoggerLevels(t *testing.T) {
	tests := []struct {
		name          string
		level         Priority
		logFunc       func(*Logger)
		expectedLevel string
		shouldLog     bool
	}{
		{
			name:  "debug not logged at info level",
			level: INFO,
			logFunc: func(l *Logger) {
				l.Debug("debug message")
			},
			expectedLevel: "DEBUG",
			shouldLog:     false,
		},
		{
			name:  "info logged at info level",
			level: INFO,
			logFunc: func(l *Logger) {
				l.Info("info message")
			},
			expectedLevel: "INFO",
			shouldLog:     true,
		},
		{
			name:  "warning logged at info level",
			level: INFO,
			logFunc: func(l *Logger) {
				l.Warn("warn message")
			},
			expectedLevel: "WARN",
			shouldLog:     true,
		},
		{
			name:  "error logged at info level",
			level: INFO,
			logFunc: func(l *Logger) {
				l.Error("error message")
			},
			expectedLevel: "ERROR",
			shouldLog:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := createTestLogger(buf)
			logger.SetLevel(tt.level)

			tt.logFunc(logger)

			output := buf.String()
			if tt.shouldLog && output == "" {
				t.Errorf("expected log output, got none")
			}
			if !tt.shouldLog && output != "" {
				t.Errorf("expected no log output, got: %s", output)
			}

			if tt.shouldLog {
				logData, err := parseLogOutput(output)
				if err != nil {
					t.Fatalf("failed to parse log output: %v", err)
				}

				if level, ok := logData["level"].(string); !ok || level != tt.expectedLevel {
					t.Errorf("expected level %s, got %s", tt.expectedLevel, level)
				}
			}
		})
	}
}

func TestErrorSourceInfo(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := createTestLogger(buf)

	logger.Error("error message")

	logData, err := parseLogOutput(buf.String())
	if err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	if _, ok := logData["file"].(string); !ok {
		t.Error("file information missing from error log")
	}
	if _, ok := logData["line"].(float64); !ok {
		t.Error("line information missing from error log")
	}
	if _, ok := logData["func"].(string); !ok {
		t.Error("function information missing from error log")
	}
}

func TestLogAtLevel(t *testing.T) {
	tests := []struct {
		name      string
		level     Priority
		logLevel  Priority
		message   string
		shouldLog bool
	}{
		{
			name:      "log at same level",
			level:     INFO,
			logLevel:  INFO,
			message:   "test message",
			shouldLog: true,
		},
		{
			name:      "log at higher level",
			level:     INFO,
			logLevel:  WARNING,
			message:   "test message",
			shouldLog: true,
		},
		{
			name:      "log at lower level",
			level:     WARNING,
			logLevel:  INFO,
			message:   "test message",
			shouldLog: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := createTestLogger(buf)
			logger.SetLevel(tt.level)

			logger.LogAtLevel(tt.logLevel, tt.message)

			hasOutput := buf.String() != ""
			if hasOutput != tt.shouldLog {
				t.Errorf("expected shouldLog=%v, got output: %v", tt.shouldLog, hasOutput)
			}
		})
	}
}

func TestWithContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := createTestLogger(buf)

	nilCtxLogger := logger.WithContext(nil)
	if nilCtxLogger != logger {
		t.Error("WithContext(nil) should return same logger")
	}

	ctx := context.Background()
	ctxLogger := logger.WithContext(ctx)
	if ctxLogger == logger {
		t.Error("WithContext should return new logger instance")
	}
}

func TestLoggerAttributes(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := createTestLogger(buf)

	logger.Info("test message",
		"string_key", "string_value",
		"int_key", 42,
		"bool_key", true,
	)

	logData, err := parseLogOutput(buf.String())
	if err != nil {
		t.Fatalf("failed to parse log output: %v", err)
	}

	expectedAttrs := map[string]interface{}{
		"string_key": "string_value",
		"int_key":    float64(42),
		"bool_key":   true,
	}

	for k, v := range expectedAttrs {
		if logData[k] != v {
			t.Errorf("expected %s=%v, got %v", k, v, logData[k])
		}
	}
}
