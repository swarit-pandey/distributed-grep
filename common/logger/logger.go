package logger

import (
	"context"
	"log/slog"
	"os"
	"runtime"
)

type Priority int

const (
	DEBUG Priority = iota
	INFO
	WARNING
	ERR
	SILENT Priority = 999
)

type Logger struct {
	*slog.Logger
	level Priority
}

type customHandler struct {
	handler slog.Handler
}

func (h *customHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *customHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &customHandler{h.handler.WithAttrs(attrs)}
}

func (h *customHandler) WithGroup(name string) slog.Handler {
	return &customHandler{h.handler.WithGroup(name)}
}

func (h *customHandler) Handle(ctx context.Context, r slog.Record) error {
	if r.Level < slog.LevelError {
		newR := slog.NewRecord(r.Time, r.Level, r.Message, 0)
		r.Attrs(func(a slog.Attr) bool {
			newR.AddAttrs(a)
			return true
		})
		return h.handler.Handle(ctx, newR)
	}
	return h.handler.Handle(ctx, r)
}

func New() *Logger {
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	})

	handler := &customHandler{baseHandler}

	l := &Logger{
		Logger: slog.New(handler),
		level:  INFO, // default level
	}

	return l
}

func (l *Logger) SetLevel(level Priority) {
	l.level = level
}

func (l *Logger) LogAtLevel(level Priority, msg string, args ...any) {
	if level < l.level {
		return
	}

	switch level {
	case DEBUG:
		l.Debug(msg, args...)
	case INFO:
		l.Info(msg, args...)
	case WARNING:
		l.Warn(msg, args...)
	case ERR:
		l.Error(msg, args...)
	}
}

func (l *Logger) Debug(msg string, args ...any) {
	if l.level <= DEBUG {
		l.Logger.Debug(msg, args...)
	}
}

func (l *Logger) Info(msg string, args ...any) {
	if l.level <= INFO {
		l.Logger.Info(msg, args...)
	}
}

func (l *Logger) Warn(msg string, args ...any) {
	if l.level <= WARNING {
		l.Logger.Warn(msg, args...)
	}
}

func (l *Logger) Error(msg string, args ...any) {
	if l.level <= ERR {
		pc, file, line, ok := runtime.Caller(1)
		if ok {
			args = append(args,
				slog.String("file", file),
				slog.Int("line", line),
				slog.String("func", runtime.FuncForPC(pc).Name()),
			)
		}
		l.Logger.Error(msg, args...)
	}
}

func (l *Logger) WithContext(ctx context.Context) *Logger {
	if ctx == nil {
		return l
	}

	return &Logger{
		Logger: l.Logger.With(slog.String("trace_id", GetTraceIDFromContext(ctx))),
		level:  l.level,
	}
}

func GetTraceIDFromContext(ctx context.Context) string {
	// TODO: Implement based race ID in context
	return "" // placeholder
}
