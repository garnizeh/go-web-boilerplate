// Package logger provides support for initializing the log system.
package logger

import (
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"runtime"
	"time"

	"log/slog"
)

// Logger represents a logger for logging information.
type Logger struct {
	handler slog.Handler
}

// New constructs a new log for application use.
func New(w io.Writer, minLevel Level, serviceName string) *Logger {
	return new(w, minLevel, serviceName, Events{})
}

// NewWithEvents constructs a new log for application use with events.
func NewWithEvents(w io.Writer, minLevel Level, serviceName string, events Events) *Logger {
	return new(w, minLevel, serviceName, events)
}

// NewWithHandler returns a new log for application use with the underlying
// handler.
func NewWithHandler(h slog.Handler) *Logger {
	return &Logger{handler: h}
}

// NewStdLogger returns a standard library Logger that wraps the slog Logger.
func NewStdLogger(logger *Logger, level Level) *log.Logger {
	return slog.NewLogLogger(logger.handler, slog.Level(level))
}

// Debug logs at LevelDebug with the given context.
func (log *Logger) Debug(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelDebug, 3, msg, args...)
}

// Debugc logs the information at the specified call stack position.
func (log *Logger) Debugc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelDebug, caller, msg, args...)
}

// Info logs at LevelInfo with the given context.
func (log *Logger) Info(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelInfo, 3, msg, args...)
}

// Infoc logs the information at the specified call stack position.
func (log *Logger) Infoc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelInfo, caller, msg, args...)
}

// Warn logs at LevelWarn with the given context.
func (log *Logger) Warn(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelWarn, 3, msg, args...)
}

// Warnc logs the information at the specified call stack position.
func (log *Logger) Warnc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelWarn, caller, msg, args...)
}

// Error logs at LevelError with the given context.
func (log *Logger) Error(ctx context.Context, msg string, args ...any) {
	log.write(ctx, LevelError, 3, msg, args...)
}

// Errorc logs the information at the specified call stack position.
func (log *Logger) Errorc(ctx context.Context, caller int, msg string, args ...any) {
	log.write(ctx, LevelError, caller, msg, args...)
}

func (log *Logger) write(ctx context.Context, level Level, caller int, msg string, args ...any) {
	slogLevel := slog.Level(level)

	if !log.handler.Enabled(ctx, slogLevel) {
		return
	}

	var pcs [1]uintptr
	runtime.Callers(caller, pcs[:])

	r := slog.NewRecord(time.Now(), slogLevel, msg, pcs[0])

	args = append(args, "trace_id", log.getTraceID(ctx))
	r.Add(args...)

	if err := log.handler.Handle(ctx, r); err != nil {
		panic(err)
	}
}

type ctxKey int

const Key ctxKey = 1

// Values represent state for each request.
type Values struct {
	TraceID   string
	RequestID string
	Now       time.Time
}

func (log *Logger) getTraceID(ctx context.Context) string {
	v, ok := ctx.Value(Key).(*Values)
	if !ok {
		return "00000000-0000-0000-0000-000000000000"
	}

	return v.TraceID
}

func new(w io.Writer, minLevel Level, serviceName string, events Events) *Logger {
	// Convert the file name to just the name.ext when this key/value will be logged.
	f := func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.SourceKey {
			if source, ok := a.Value.Any().(*slog.Source); ok {
				v := fmt.Sprintf("%s:%d", filepath.Base(source.File), source.Line)
				return slog.Attr{Key: "file", Value: slog.StringValue(v)}
			}
		}

		return a
	}

	// Construct the slog JSON handler for use.
	opts := slog.HandlerOptions{
		AddSource:   true,
		Level:       slog.Level(minLevel),
		ReplaceAttr: f,
	}
	handler := slog.Handler(slog.NewJSONHandler(w, &opts))

	// If events are to be processed, wrap the JSON handler around the custom log handler.
	if events.Debug != nil || events.Info != nil || events.Warn != nil || events.Error != nil {
		handler = newLogHandler(handler, events)
	}

	// Attributes to add to every log.
	attrs := []slog.Attr{
		{Key: "service", Value: slog.StringValue(serviceName)},
	}

	// Add those attributes and capture the final handler.
	handler = handler.WithAttrs(attrs)

	return &Logger{handler: handler}
}
