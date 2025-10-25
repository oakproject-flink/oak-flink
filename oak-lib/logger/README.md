# Oak Logger

Component-based structured logging for Oak with minimal performance overhead.

## Features

- ✅ **Component-based logging**: Each component gets its own log file
- ✅ **Generic log file**: All logs also go to `oak.log`
- ✅ **Async logging**: Buffered channels for lock-free hot path
- ✅ **Multiple formats**: Text (human-readable) or JSON (machine-parsable)
- ✅ **Configurable fields**: Choose what to include in logs
- ✅ **Dual output**: Logs to both stdout and files
- ⏳ **Log rotation**: Planned (by size or time)
- ⏳ **Sampling**: Planned (for high-frequency logs)

## Usage

### Basic Usage

```go
import "github.com/oakproject-flink/oak-flink/oak-lib/logger"

// Create a component-specific logger
log := logger.NewComponent("grpc")

// Log at different levels
log.Infof("Server started on port %d", 9090)
log.Warnf("Retrying connection after %v", delay)
log.Errorf("Failed to connect: %v", err)
log.Debugf("Processing message: %+v", msg) // Only if debug enabled
```

### Configuration

```go
// Set global configuration before creating loggers
logger.SetGlobalConfig(&logger.Config{
    LogDir:   "/var/log/oak",
    Format:   logger.FormatJSON,  // or FormatText
    Debug:    true,
    Fields:   []string{"timestamp", "level", "component", "message"},
    ToStdout: true,
    ToFile:   true,
    BufSize:  5000, // Async buffer size
})

// Then create loggers
log := logger.NewComponent("server")
```

### Shutdown

```go
// In main() defer:
defer logger.CloseAll()
```

## Log Files

Logs are written to:

```
/var/log/oak/              (Linux, if writable)
  ├── oak.log              (all components)
  ├── grpc.log             (grpc component only)
  ├── server.log           (server component only)
  └── agent.log            (agent component only)

%ProgramData%\Oak\logs\    (Windows)
  ├── oak.log
  ├── grpc.log
  └── ...

./logs/                    (fallback)
  ├── oak.log
  └── ...
```

## Log Formats

### Text Format (Human-Readable)

```
2025-10-25 14:23:45.123 [INFO ] [grpc] Server started on port 9090
2025-10-25 14:23:46.456 [WARN ] [grpc] Retrying connection after 5s
2025-10-25 14:23:47.789 [ERROR] [grpc] Failed to connect: connection refused
```

### JSON Format (Machine-Parsable)

```json
{"timestamp":"2025-10-25T14:23:45.123Z","level":"INFO","component":"grpc","message":"Server started on port 9090"}
{"timestamp":"2025-10-25T14:23:46.456Z","level":"WARN","component":"grpc","message":"Retrying connection after 5s"}
{"timestamp":"2025-10-25T14:23:47.789Z","level":"ERROR","component":"grpc","message":"Failed to connect: connection refused"}
```

## Configurable Fields

Available fields:
- `timestamp` - Log timestamp
- `level` - Log level (INFO, WARN, ERROR, DEBUG, FATAL)
- `component` - Component name
- `message` - Log message
- `caller` - File:line (future)

Example:
```go
logger.SetGlobalConfig(&logger.Config{
    Fields: []string{"timestamp", "level", "component", "message"},
})
```

## Environment Variables

- `OAK_LOG_DIR` - Override log directory
- `OAK_DEBUG` - Enable debug logging (set to "true")

## Performance

- **Lock-free hot path**: Uses buffered channels instead of mutexes
- **Async I/O**: Log writing happens in background goroutine
- **Non-blocking**: Drops messages if buffer full (prevents blocking)
- **Typical overhead**: ~100-200ns per log call

## Future Features

### Log Rotation (Planned)

```go
logger.SetGlobalConfig(&logger.Config{
    Rotation: &logger.RotationConfig{
        MaxSize:    100 * 1024 * 1024,  // 100MB
        MaxAge:     7 * 24 * time.Hour, // 7 days
        MaxBackups: 5,
        Compress:   true,
    },
})
```

### Sampling (Planned)

For high-frequency logs, sample to reduce volume:

```go
log.SampleInfof(100, "Processed request %d", reqID) // Log 1 in 100
```

### Structured Fields (Planned)

```go
log.WithFields(map[string]interface{}{
    "user_id": 123,
    "request_id": "abc-def",
}).Infof("User logged in")
```

## Integration Examples

### With gRPC Server

```go
type grpcLogger struct {
    log *logger.Logger
}

func NewGRPCLogger() *grpcLogger {
    return &grpcLogger{log: logger.NewComponent("grpc")}
}

func (l *grpcLogger) Infof(format string, args ...interface{}) {
    l.log.Infof(format, args...)
}
// ... implement other methods
```

### With HTTP Server (Echo)

```go
log := logger.NewComponent("http")

e := echo.New()
e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
    LogStatus: true,
    LogURI:    true,
    LogError:  true,
    HandleError: true,
    LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
        if v.Error != nil {
            log.Errorf("%s %s - %v", v.Method, v.URI, v.Error)
        } else {
            log.Infof("%s %s - %d", v.Method, v.URI, v.Status)
        }
        return nil
    },
}))
```

## Best Practices

1. **Create one logger per component**: Helps with debugging and log organization
2. **Use appropriate log levels**:
   - `DEBUG`: Verbose, only in development
   - `INFO`: Normal operational messages
   - `WARN`: Something unexpected but not critical
   - `ERROR`: Operation failed but system continues
   - `FATAL`: System cannot continue
3. **Include context**: Log relevant IDs, names, etc.
4. **Avoid logging PII**: Don't log passwords, tokens, personal data
5. **Use JSON format in production**: Easier for log aggregation systems

## Testing

In tests, you can disable file logging:

```go
logger.SetGlobalConfig(&logger.Config{
    ToFile:   false,
    ToStdout: true,
    Format:   logger.FormatText,
})
```

Or capture logs:

```go
// TODO: Add test helper for capturing logs
```
