package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Logger interface {
	Debug(args ...interface{})
	Info(args ...interface{})
	Warn(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warnf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger
}

type logger struct {
	slog *slog.Logger
}

func New(level string) Logger {
	// Parse log level
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Create handler options
	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true, // Add source file and line number
	}

	// Create text handler for readable output
	handler := slog.NewTextHandler(os.Stdout, opts)

	return &logger{slog: slog.New(handler)}
}

// NewJSON creates a logger with JSON output format
func NewJSON(level string) Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	return &logger{slog: slog.New(handler)}
}

// NewWithWriter creates a logger with custom writer
func NewWithWriter(level string, writer io.Writer, format string) Logger {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}

	var handler slog.Handler
	if format == "json" {
		handler = slog.NewJSONHandler(writer, opts)
	} else {
		handler = slog.NewTextHandler(writer, opts)
	}

	return &logger{slog: slog.New(handler)}
}

func (l *logger) Debug(args ...interface{}) {
	l.slog.Debug(formatArgs(args...))
}

func (l *logger) Info(args ...interface{}) {
	l.slog.Info(formatArgs(args...))
}

func (l *logger) Warn(args ...interface{}) {
	l.slog.Warn(formatArgs(args...))
}

func (l *logger) Error(args ...interface{}) {
	l.slog.Error(formatArgs(args...))
}

func (l *logger) Fatal(args ...interface{}) {
	l.slog.Error(formatArgs(args...))
	os.Exit(1)
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.slog.Debug(fmt.Sprintf(format, args...))
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.slog.Info(fmt.Sprintf(format, args...))
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.slog.Warn(fmt.Sprintf(format, args...))
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.slog.Error(fmt.Sprintf(format, args...))
}

func (l *logger) Fatalf(format string, args ...interface{}) {
	l.slog.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func (l *logger) WithField(key string, value interface{}) Logger {
	return &logger{slog: l.slog.With(key, value)}
}

func (l *logger) WithFields(fields map[string]interface{}) Logger {
	args := make([]interface{}, 0, len(fields)*2)
	for k, v := range fields {
		args = append(args, k, v)
	}
	return &logger{slog: l.slog.With(args...)}
}

// WithError adds an error field to the logger
func (l *logger) WithError(err error) Logger {
	return l.WithField("error", err.Error())
}

// formatArgs converts variadic args to a single string message
func formatArgs(args ...interface{}) string {
	if len(args) == 0 {
		return ""
	}
	if len(args) == 1 {
		return fmt.Sprint(args[0])
	}
	return fmt.Sprintln(args...)
}

// noopLogger implements Logger interface but does nothing
type noopLogger struct{}

func (nl *noopLogger) Debug(args ...interface{})                       {}
func (nl *noopLogger) Info(args ...interface{})                        {}
func (nl *noopLogger) Warn(args ...interface{})                        {}
func (nl *noopLogger) Error(args ...interface{})                       {}
func (nl *noopLogger) Fatal(args ...interface{})                       {}
func (nl *noopLogger) Debugf(format string, args ...interface{})       {}
func (nl *noopLogger) Infof(format string, args ...interface{})        {}
func (nl *noopLogger) Warnf(format string, args ...interface{})        {}
func (nl *noopLogger) Errorf(format string, args ...interface{})       {}
func (nl *noopLogger) Fatalf(format string, args ...interface{})       {}
func (nl *noopLogger) WithField(key string, value interface{}) Logger  { return nl }
func (nl *noopLogger) WithFields(fields map[string]interface{}) Logger { return nl }
func (nl *noopLogger) WithError(err error) Logger                      { return nl }

// NewNoop creates a no-op logger for testing
func NewNoop() Logger {
	return &noopLogger{}
}

// NewFromConfig creates a logger based on configuration
func NewFromConfig(level, format string) Logger {
	if format == "json" {
		return NewJSON(level)
	}
	return New(level)
}

// Context-aware logging methods for advanced use cases

// DebugContext logs a debug message with context
func (l *logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
	l.slog.DebugContext(ctx, msg, args...)
}

// InfoContext logs an info message with context
func (l *logger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
	l.slog.InfoContext(ctx, msg, args...)
}

// WarnContext logs a warning message with context
func (l *logger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
	l.slog.WarnContext(ctx, msg, args...)
}

// ErrorContext logs an error message with context
func (l *logger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
	l.slog.ErrorContext(ctx, msg, args...)
}
