package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// ContextKey type for context keys
type ContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey ContextKey = "request_id"
	// UserIDKey is the context key for user ID
	UserIDKey ContextKey = "user_id"
	// OrgIDKey is the context key for organization ID
	OrgIDKey ContextKey = "org_id"
	// TraceIDKey is the context key for trace ID
	TraceIDKey ContextKey = "trace_id"
	// SpanIDKey is the context key for span ID
	SpanIDKey ContextKey = "span_id"
)

// Config holds logger configuration
type Config struct {
	Level      string // debug, info, warn, error
	Format     string // json, console
	Output     string // stdout, stderr, file path
	AppName    string
	AppVersion string
	Env        string
}

// Logger wraps zerolog with additional functionality
type Logger struct {
	zl zerolog.Logger
}

var defaultLogger *Logger

// Init initializes the default logger
func Init(cfg Config) *Logger {
	var output io.Writer

	switch cfg.Output {
	case "stderr":
		output = os.Stderr
	case "stdout", "":
		output = os.Stdout
	default:
		// File output
		file, err := os.OpenFile(cfg.Output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			output = os.Stdout
		} else {
			output = file
		}
	}

	// Set format
	if cfg.Format == "console" {
		output = zerolog.ConsoleWriter{
			Out:        output,
			TimeFormat: time.RFC3339,
		}
	}

	// Parse level
	level := parseLevel(cfg.Level)

	// Create logger with default fields
	zl := zerolog.New(output).
		Level(level).
		With().
		Timestamp().
		Str("app", cfg.AppName).
		Str("version", cfg.AppVersion).
		Str("env", cfg.Env).
		Logger()

	defaultLogger = &Logger{zl: zl}
	return defaultLogger
}

// Default returns the default logger
func Default() *Logger {
	if defaultLogger == nil {
		// Initialize with defaults
		Init(Config{
			Level:   "info",
			Format:  "json",
			Output:  "stdout",
			AppName: "ads-analytics",
		})
	}
	return defaultLogger
}

// parseLevel converts string level to zerolog.Level
func parseLevel(level string) zerolog.Level {
	switch level {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "fatal":
		return zerolog.FatalLevel
	case "panic":
		return zerolog.PanicLevel
	default:
		return zerolog.InfoLevel
	}
}

// WithContext returns a logger with context values
func (l *Logger) WithContext(ctx context.Context) *Logger {
	newLogger := l.zl.With().Logger()

	if requestID, ok := ctx.Value(RequestIDKey).(string); ok && requestID != "" {
		newLogger = newLogger.With().Str("request_id", requestID).Logger()
	}

	if userID, ok := ctx.Value(UserIDKey).(string); ok && userID != "" {
		newLogger = newLogger.With().Str("user_id", userID).Logger()
	}

	if orgID, ok := ctx.Value(OrgIDKey).(string); ok && orgID != "" {
		newLogger = newLogger.With().Str("org_id", orgID).Logger()
	}

	if traceID, ok := ctx.Value(TraceIDKey).(string); ok && traceID != "" {
		newLogger = newLogger.With().Str("trace_id", traceID).Logger()
	}

	if spanID, ok := ctx.Value(SpanIDKey).(string); ok && spanID != "" {
		newLogger = newLogger.With().Str("span_id", spanID).Logger()
	}

	return &Logger{zl: newLogger}
}

// With returns a logger with additional fields
func (l *Logger) With() *LogContext {
	return &LogContext{ctx: l.zl.With()}
}

// LogContext wraps zerolog.Context for fluent field addition
type LogContext struct {
	ctx zerolog.Context
}

// Str adds a string field
func (c *LogContext) Str(key, val string) *LogContext {
	c.ctx = c.ctx.Str(key, val)
	return c
}

// Int adds an int field
func (c *LogContext) Int(key string, val int) *LogContext {
	c.ctx = c.ctx.Int(key, val)
	return c
}

// Int64 adds an int64 field
func (c *LogContext) Int64(key string, val int64) *LogContext {
	c.ctx = c.ctx.Int64(key, val)
	return c
}

// Float64 adds a float64 field
func (c *LogContext) Float64(key string, val float64) *LogContext {
	c.ctx = c.ctx.Float64(key, val)
	return c
}

// Bool adds a bool field
func (c *LogContext) Bool(key string, val bool) *LogContext {
	c.ctx = c.ctx.Bool(key, val)
	return c
}

// Err adds an error field
func (c *LogContext) Err(err error) *LogContext {
	c.ctx = c.ctx.Err(err)
	return c
}

// Dur adds a duration field
func (c *LogContext) Dur(key string, val time.Duration) *LogContext {
	c.ctx = c.ctx.Dur(key, val)
	return c
}

// Time adds a time field
func (c *LogContext) Time(key string, val time.Time) *LogContext {
	c.ctx = c.ctx.Time(key, val)
	return c
}

// Interface adds an interface field
func (c *LogContext) Interface(key string, val interface{}) *LogContext {
	c.ctx = c.ctx.Interface(key, val)
	return c
}

// Logger returns the logger with accumulated fields
func (c *LogContext) Logger() *Logger {
	return &Logger{zl: c.ctx.Logger()}
}

// Debug logs at debug level
func (l *Logger) Debug() *LogEvent {
	return &LogEvent{event: l.zl.Debug()}
}

// Info logs at info level
func (l *Logger) Info() *LogEvent {
	return &LogEvent{event: l.zl.Info()}
}

// Warn logs at warn level
func (l *Logger) Warn() *LogEvent {
	return &LogEvent{event: l.zl.Warn()}
}

// Error logs at error level
func (l *Logger) Error() *LogEvent {
	return &LogEvent{event: l.zl.Error()}
}

// Fatal logs at fatal level and exits
func (l *Logger) Fatal() *LogEvent {
	return &LogEvent{event: l.zl.Fatal()}
}

// Panic logs at panic level and panics
func (l *Logger) Panic() *LogEvent {
	return &LogEvent{event: l.zl.Panic()}
}

// LogEvent wraps zerolog.Event for fluent logging
type LogEvent struct {
	event *zerolog.Event
}

// Str adds a string field
func (e *LogEvent) Str(key, val string) *LogEvent {
	e.event = e.event.Str(key, val)
	return e
}

// Int adds an int field
func (e *LogEvent) Int(key string, val int) *LogEvent {
	e.event = e.event.Int(key, val)
	return e
}

// Int64 adds an int64 field
func (e *LogEvent) Int64(key string, val int64) *LogEvent {
	e.event = e.event.Int64(key, val)
	return e
}

// Float64 adds a float64 field
func (e *LogEvent) Float64(key string, val float64) *LogEvent {
	e.event = e.event.Float64(key, val)
	return e
}

// Bool adds a bool field
func (e *LogEvent) Bool(key string, val bool) *LogEvent {
	e.event = e.event.Bool(key, val)
	return e
}

// Err adds an error field
func (e *LogEvent) Err(err error) *LogEvent {
	e.event = e.event.Err(err)
	return e
}

// Dur adds a duration field
func (e *LogEvent) Dur(key string, val time.Duration) *LogEvent {
	e.event = e.event.Dur(key, val)
	return e
}

// Time adds a time field
func (e *LogEvent) Time(key string, val time.Time) *LogEvent {
	e.event = e.event.Time(key, val)
	return e
}

// Interface adds an interface field
func (e *LogEvent) Interface(key string, val interface{}) *LogEvent {
	e.event = e.event.Interface(key, val)
	return e
}

// Msg sends the log with a message
func (e *LogEvent) Msg(msg string) {
	e.event.Msg(msg)
}

// Msgf sends the log with a formatted message
func (e *LogEvent) Msgf(format string, v ...interface{}) {
	e.event.Msgf(format, v...)
}

// Send sends the log without a message
func (e *LogEvent) Send() {
	e.event.Send()
}
