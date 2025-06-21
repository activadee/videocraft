# VideoCraft Logger Package - Structured Logging with Security & Performance

The logger package provides comprehensive structured logging capabilities built on Go's slog framework. It offers multiple output formats, security-aware logging, and performance-optimized implementations for VideoCraft's observability needs.

## 📝 Logger Architecture

```mermaid
graph TB
    subgraph "Logger Interface Layer"
        INTERFACE[Logger Interface]
        METHODS[Standard Methods]
        FORMATTED[Formatted Methods]
        STRUCTURED[Structured Methods]
    end
    
    subgraph "Implementation Layer"
        SLOG_IMPL[slog Implementation]
        HANDLER_FACTORY[Handler Factory]
        CONTEXT_SUPPORT[Context Support]
        NOOP_IMPL[No-op Implementation]
    end
    
    subgraph "Handler Types"
        TEXT_HANDLER[Text Handler]
        JSON_HANDLER[JSON Handler]
        CUSTOM_HANDLER[Custom Writer Handler]
        MULTI_HANDLER[Multi-destination Handler]
    end
    
    subgraph "Output Destinations"
        STDOUT[Standard Output]
        STDERR[Standard Error]
        FILE_OUTPUT[File Output]
        REMOTE_LOGGING[Remote Logging]
    end
    
    subgraph "Log Levels"
        DEBUG[Debug (-4)]
        INFO[Info (0)]
        WARN[Warning (4)]
        ERROR[Error (8)]
        FATAL[Fatal (12)]
    end
    
    subgraph "Features"
        SOURCE_INFO[Source Information]
        FIELD_SUPPORT[Structured Fields]
        ERROR_ENRICHMENT[Error Enrichment]
        PERFORMANCE[Performance Optimized]
    end
    
    INTERFACE --> METHODS
    INTERFACE --> FORMATTED
    INTERFACE --> STRUCTURED
    
    METHODS --> SLOG_IMPL
    FORMATTED --> SLOG_IMPL
    STRUCTURED --> SLOG_IMPL
    
    SLOG_IMPL --> HANDLER_FACTORY
    SLOG_IMPL --> CONTEXT_SUPPORT
    INTERFACE --> NOOP_IMPL
    
    HANDLER_FACTORY --> TEXT_HANDLER
    HANDLER_FACTORY --> JSON_HANDLER
    HANDLER_FACTORY --> CUSTOM_HANDLER
    HANDLER_FACTORY --> MULTI_HANDLER
    
    TEXT_HANDLER --> STDOUT
    JSON_HANDLER --> STDOUT
    CUSTOM_HANDLER --> FILE_OUTPUT
    MULTI_HANDLER --> REMOTE_LOGGING
    
    SLOG_IMPL --> DEBUG
    SLOG_IMPL --> INFO
    SLOG_IMPL --> WARN
    SLOG_IMPL --> ERROR
    SLOG_IMPL --> FATAL
    
    SLOG_IMPL --> SOURCE_INFO
    STRUCTURED --> FIELD_SUPPORT
    STRUCTURED --> ERROR_ENRICHMENT
    SLOG_IMPL --> PERFORMANCE
    
    style INTERFACE fill:#e3f2fd
    style SLOG_IMPL fill:#f3e5f5
    style STRUCTURED fill:#e8f5e8
    style SOURCE_INFO fill:#fff3e0
```

## 🔧 Logger Interface Design

### Comprehensive Logging Interface

```go
type Logger interface {
    // Standard logging methods for simple messages
    Debug(args ...interface{})
    Info(args ...interface{})
    Warn(args ...interface{})
    Error(args ...interface{})
    Fatal(args ...interface{})

    // Formatted logging methods with printf-style formatting
    Debugf(format string, args ...interface{})
    Infof(format string, args ...interface{})
    Warnf(format string, args ...interface{})
    Errorf(format string, args ...interface{})
    Fatalf(format string, args ...interface{})

    // Structured logging methods for enriched context
    WithField(key string, value interface{}) Logger
    WithFields(fields map[string]interface{}) Logger
    WithError(err error) Logger
}
```

### Logger Implementation Structure

```go
type logger struct {
    slog *slog.Logger
}

// Main constructor with sensible defaults
func New(level string) Logger {
    // Parse log level with safe fallback
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
        logLevel = slog.LevelInfo // Safe default for production
    }

    // Create handler options with source information
    opts := &slog.HandlerOptions{
        Level:     logLevel,
        AddSource: true, // Include source file and line number for debugging
    }

    // Create human-readable text handler for development
    handler := slog.NewTextHandler(os.Stdout, opts)
    return &logger{slog: slog.New(handler)}
}
```

## 🎯 Advanced Logger Factories

### Production-Ready JSON Logger

```go
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

    // Create structured JSON handler for production logging
    handler := slog.NewJSONHandler(os.Stdout, opts)
    return &logger{slog: slog.New(handler)}
}
```

### Flexible Writer-Based Logger

```go
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
```

### Configuration-Driven Logger Factory

```go
func NewFromConfig(level, format string) Logger {
    if format == "json" {
        return NewJSON(level)
    }
    return New(level)
}
```

## 📊 Standard Logging Methods

### Simple Logging Implementation

```go
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

// Efficient argument formatting
func formatArgs(args ...interface{}) string {
    if len(args) == 0 {
        return ""
    }
    if len(args) == 1 {
        return fmt.Sprint(args[0])
    }
    return fmt.Sprintln(args...)
}
```

### Formatted Logging Methods

```go
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
```

## 🏗️ Structured Logging Implementation

### Field-Based Structured Logging

```go
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

func (l *logger) WithError(err error) Logger {
    return l.WithField("error", err.Error())
}
```

### Context-Aware Logging

```go
// Context-aware logging methods for advanced use cases
func (l *logger) DebugContext(ctx context.Context, msg string, args ...interface{}) {
    l.slog.DebugContext(ctx, msg, args...)
}

func (l *logger) InfoContext(ctx context.Context, msg string, args ...interface{}) {
    l.slog.InfoContext(ctx, msg, args...)
}

func (l *logger) WarnContext(ctx context.Context, msg string, args ...interface{}) {
    l.slog.WarnContext(ctx, msg, args...)
}

func (l *logger) ErrorContext(ctx context.Context, msg string, args ...interface{}) {
    l.slog.ErrorContext(ctx, msg, args...)
}
```

## 🧪 Testing Support

### No-op Logger for Testing

```go
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

func NewNoop() Logger {
    return &noopLogger{}
}
```

## 📈 Performance Characteristics

### Efficient Log Level Checking

The logger leverages slog's built-in level checking to avoid expensive operations when logging is disabled:

```go
// slog automatically checks log level before processing
// No string formatting occurs if level is disabled
logger.Debugf("Processing file: %s with size: %d", filename, size)
```

### Memory-Efficient Field Handling

```go
func (l *logger) WithFields(fields map[string]interface{}) Logger {
    // Pre-allocate slice to avoid memory reallocations
    args := make([]interface{}, 0, len(fields)*2)
    for k, v := range fields {
        args = append(args, k, v)
    }
    return &logger{slog: l.slog.With(args...)}
}
```

### Optimized Argument Formatting

```go
func formatArgs(args ...interface{}) string {
    if len(args) == 0 {
        return ""
    }
    if len(args) == 1 {
        // Avoid unnecessary string operations for single arguments
        return fmt.Sprint(args[0])
    }
    return fmt.Sprintln(args...)
}
```

## 💼 Usage Patterns

### Basic Logging

```go
logger := logger.New("info")

// Simple messages
logger.Info("Service starting")
logger.Error("Connection failed")

// Formatted messages
logger.Infof("Processing video with ID: %s", videoID)
logger.Errorf("Failed to process video %s: %v", videoID, err)
```

### Structured Logging

```go
// Single field
logger.WithField("user_id", "user_123").Info("User action performed")

// Multiple fields
logger.WithFields(map[string]interface{}{
    "job_id":     "job_456",
    "duration":   "30s",
    "file_size":  "10MB",
    "quality":    "high",
}).Info("Video generation completed")

// Error context
logger.WithError(err).Error("Video processing failed")
```

### Production Logging

```go
// JSON format for production
logger := logger.NewJSON("info")

// Service lifecycle
logger.WithFields(map[string]interface{}{
    "service":    "videocraft",
    "version":    "0.0.1",
    "port":       8080,
}).Info("Service started")

// Request tracking
requestLogger := logger.WithFields(map[string]interface{}{
    "request_id":   "req_789",
    "user_id":      "user_123",
    "endpoint":     "/api/videos",
    "method":       "POST",
})

requestLogger.Info("Request received")
requestLogger.WithField("duration", "500ms").Info("Request completed")
```

### Security Logging

```go
// Security events with structured data
logger.WithFields(map[string]interface{}{
    "event_type":     "security_violation",
    "violation_type": "path_traversal",
    "source_ip":      clientIP,
    "user_agent":     userAgent,
    "attempted_path": suspiciousPath,
}).Error("Security violation detected")

// Audit logging
logger.WithFields(map[string]interface{}{
    "action":     "file_access",
    "user_id":    userID,
    "resource":   filename,
    "permission": "read",
    "granted":    true,
}).Info("File access granted")
```

## 🔧 Configuration

### Logger Configuration

```yaml
logging:
  level: "info"                    # debug, info, warn, error
  format: "json"                   # text, json
  add_source: true                 # Include source file and line number
  output: "stdout"                 # stdout, stderr, or file path
  
development:
  level: "debug"
  format: "text"
  add_source: true
  
production:
  level: "info"
  format: "json"
  add_source: false                # Disable for performance in production
```

### Environment-Based Configuration

```go
func NewFromEnvironment() Logger {
    level := os.Getenv("LOG_LEVEL")
    if level == "" {
        level = "info"
    }
    
    format := os.Getenv("LOG_FORMAT")
    if format == "" {
        format = "text"
    }
    
    return NewFromConfig(level, format)
}
```

## 🧪 Testing Strategy

### Unit Tests

```go
func TestLogger_BasicLogging(t *testing.T) {
    var buf bytes.Buffer
    logger := NewWithWriter("debug", &buf, "text")
    
    logger.Info("test message")
    
    output := buf.String()
    assert.Contains(t, output, "test message")
    assert.Contains(t, output, "level=INFO")
}

func TestLogger_StructuredLogging(t *testing.T) {
    var buf bytes.Buffer
    logger := NewWithWriter("info", &buf, "json")
    
    logger.WithFields(map[string]interface{}{
        "user_id": "123",
        "action":  "test",
    }).Info("test message")
    
    output := buf.String()
    assert.Contains(t, output, `"user_id":"123"`)
    assert.Contains(t, output, `"action":"test"`)
    assert.Contains(t, output, `"level":"INFO"`)
}

func TestLogger_ErrorContext(t *testing.T) {
    var buf bytes.Buffer
    logger := NewWithWriter("error", &buf, "json")
    
    err := errors.New("test error")
    logger.WithError(err).Error("operation failed")
    
    output := buf.String()
    assert.Contains(t, output, `"error":"test error"`)
    assert.Contains(t, output, "operation failed")
}

func TestLogger_LogLevels(t *testing.T) {
    var buf bytes.Buffer
    
    // Test with warn level - debug and info should be filtered
    logger := NewWithWriter("warn", &buf, "text")
    
    logger.Debug("debug message")  // Should not appear
    logger.Info("info message")    // Should not appear
    logger.Warn("warn message")    // Should appear
    logger.Error("error message") // Should appear
    
    output := buf.String()
    assert.NotContains(t, output, "debug message")
    assert.NotContains(t, output, "info message")
    assert.Contains(t, output, "warn message")
    assert.Contains(t, output, "error message")
}

func TestLogger_NoopLogger(t *testing.T) {
    logger := NewNoop()
    
    // Should not panic or produce output
    logger.Debug("debug")
    logger.Info("info")
    logger.Error("error")
    
    // Chaining should work
    enriched := logger.WithField("key", "value").WithError(errors.New("test"))
    enriched.Error("test message")
    
    // No assertions needed - just ensuring no panics or output
}
```

### Performance Benchmarks

```go
func BenchmarkLogger_SimpleLogging(b *testing.B) {
    logger := NewWithWriter("info", io.Discard, "text")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Info("benchmark message")
    }
}

func BenchmarkLogger_FormattedLogging(b *testing.B) {
    logger := NewWithWriter("info", io.Discard, "text")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.Infof("benchmark message %d", i)
    }
}

func BenchmarkLogger_StructuredLogging(b *testing.B) {
    logger := NewWithWriter("info", io.Discard, "json")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        logger.WithFields(map[string]interface{}{
            "iteration": i,
            "component": "benchmark",
            "action":    "test",
        }).Info("benchmark message")
    }
}

func BenchmarkLogger_DisabledLevel(b *testing.B) {
    logger := NewWithWriter("error", io.Discard, "text")
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        // These should be filtered out and very fast
        logger.Debug("debug message")
        logger.Info("info message")
    }
}
```

### Integration Testing

```go
func TestLogger_Integration(t *testing.T) {
    // Test with real file output
    tmpFile, err := os.CreateTemp("", "logger_test_*.log")
    require.NoError(t, err)
    defer os.Remove(tmpFile.Name())
    defer tmpFile.Close()
    
    logger := NewWithWriter("info", tmpFile, "json")
    
    // Write various log types
    logger.Info("service started")
    logger.WithField("user_id", "123").Info("user action")
    logger.WithError(errors.New("test error")).Error("operation failed")
    
    // Flush and read back
    tmpFile.Sync()
    content, err := os.ReadFile(tmpFile.Name())
    require.NoError(t, err)
    
    output := string(content)
    assert.Contains(t, output, "service started")
    assert.Contains(t, output, `"user_id":"123"`)
    assert.Contains(t, output, "test error")
}
```

---

**Related Documentation:**
- [Shared Packages Overview](../CLAUDE.md)
- [Errors Package](../errors/CLAUDE.md)
- [Core Services](../../core/CLAUDE.md)
- [API Layer](../../api/CLAUDE.md)