package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Format represents the log output format
type Format string

const (
	FormatText Format = "text" // Human-readable text
	FormatJSON Format = "json" // JSON lines
)

// Config holds logger configuration
type Config struct {
	LogDir    string            // Base log directory (default: /var/log/oak or ./logs)
	Format    Format            // Output format (text or json)
	Debug     bool              // Enable debug logging
	Fields    []string          // Fields to include (timestamp, level, component, message, caller)
	ToStdout  bool              // Log to stdout (default: true)
	ToFile    bool              // Log to file (default: true)
	BufSize   int               // Async buffer size (default: 1000)
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		LogDir:   getDefaultLogDir(),
		Format:   FormatText,
		Debug:    os.Getenv("OAK_DEBUG") == "true",
		Fields:   []string{"timestamp", "level", "component", "message"},
		ToStdout: true,
		ToFile:   true,
		BufSize:  1000,
	}
}

// Logger provides component-based structured logging
type Logger struct {
	component string
	config    *Config

	// Output files
	genericFile   *os.File // oak.log
	componentFile *os.File // <component>.log

	// Async logging (minimal lock contention)
	logChan chan *logEntry
	wg      sync.WaitGroup
	closeCh chan struct{}
	once    sync.Once
}

// logEntry represents a single log entry
type logEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Level     string            `json:"level"`
	Component string            `json:"component"`
	Message   string            `json:"message"`
	Caller    string            `json:"caller,omitempty"` // Future: for debugging
	Extra     map[string]string `json:"extra,omitempty"`  // Future: additional fields
}

var (
	globalConfig     = DefaultConfig()
	genericLogFile   *os.File
	componentLoggers = make(map[string]*Logger)
	globalMu         sync.RWMutex // Use RWMutex for better read performance
)

// SetGlobalConfig sets the global logger configuration
func SetGlobalConfig(cfg *Config) {
	globalMu.Lock()
	defer globalMu.Unlock()
	globalConfig = cfg
}

// NewComponent creates or retrieves a logger for a specific component
// Components get their own log files: <logdir>/<component>.log
// All logs also go to: <logdir>/oak.log (generic)
func NewComponent(component string) *Logger {
	globalMu.Lock()
	defer globalMu.Unlock()

	// Return existing logger if already created
	if logger, exists := componentLoggers[component]; exists {
		return logger
	}

	// Create new logger
	logger := &Logger{
		component: component,
		config:    globalConfig,
		logChan:   make(chan *logEntry, globalConfig.BufSize),
		closeCh:   make(chan struct{}),
	}

	// Initialize log files
	if globalConfig.ToFile {
		// Initialize generic log file (while holding globalMu)
		if err := initGenericLogFile(globalConfig.LogDir); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Failed to initialize generic log file: %v\n", err)
		}

		// Initialize component-specific log file
		if err := logger.initLogFiles(); err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Failed to initialize log files for %s: %v\n", component, err)
		}
	}

	// Start async log worker
	logger.wg.Add(1)
	go logger.logWorker()

	// Cache the logger
	componentLoggers[component] = logger

	return logger
}

// initGenericLogFile initializes the shared generic log file (called with globalMu held)
func initGenericLogFile(logDir string) error {
	if genericLogFile != nil {
		return nil
	}

	// Create log directory
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", logDir, err)
	}

	genericPath := filepath.Join(logDir, "oak.log")
	file, err := os.OpenFile(genericPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open generic log file: %w", err)
	}
	genericLogFile = file
	return nil
}

// initLogFiles creates the generic and component-specific log files
func (l *Logger) initLogFiles() error {
	// Create log directory
	if err := os.MkdirAll(l.config.LogDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory %s: %w", l.config.LogDir, err)
	}

	// Generic file is already opened in NewComponent (which holds globalMu)
	l.genericFile = genericLogFile

	// Open component-specific log file
	componentPath := filepath.Join(l.config.LogDir, fmt.Sprintf("%s.log", l.component))
	file, err := os.OpenFile(componentPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open component log file: %w", err)
	}
	l.componentFile = file

	return nil
}

// logWorker processes log entries asynchronously (no locks on hot path)
func (l *Logger) logWorker() {
	defer l.wg.Done()

	for {
		select {
		case entry := <-l.logChan:
			l.writeEntry(entry)
		case <-l.closeCh:
			// Drain remaining entries
			for len(l.logChan) > 0 {
				l.writeEntry(<-l.logChan)
			}
			return
		}
	}
}

// writeEntry writes a log entry to all configured outputs
func (l *Logger) writeEntry(entry *logEntry) {
	var formatted string

	// Format the entry
	switch l.config.Format {
	case FormatJSON:
		formatted = l.formatJSON(entry)
	case FormatText:
		formatted = l.formatText(entry)
	default:
		formatted = l.formatText(entry)
	}

	// Write to stdout
	if l.config.ToStdout {
		fmt.Fprintln(os.Stdout, formatted)
	}

	// Write to files
	if l.config.ToFile {
		if l.genericFile != nil {
			fmt.Fprintln(l.genericFile, formatted)
		}
		if l.componentFile != nil {
			fmt.Fprintln(l.componentFile, formatted)
		}
	}
}

// formatText formats an entry as human-readable text
func (l *Logger) formatText(entry *logEntry) string {
	// Build output based on configured fields
	var result string

	for i, field := range l.config.Fields {
		if i > 0 {
			result += " "
		}

		switch field {
		case "timestamp":
			result += entry.Timestamp.Format("2006-01-02 15:04:05.000")
		case "level":
			result += fmt.Sprintf("[%-5s]", entry.Level)
		case "component":
			result += fmt.Sprintf("[%s]", entry.Component)
		case "message":
			result += entry.Message
		case "caller":
			if entry.Caller != "" {
				result += fmt.Sprintf("(%s)", entry.Caller)
			}
		}
	}

	return result
}

// formatJSON formats an entry as JSON
func (l *Logger) formatJSON(entry *logEntry) string {
	// Build JSON object with only configured fields
	obj := make(map[string]interface{})

	for _, field := range l.config.Fields {
		switch field {
		case "timestamp":
			obj["timestamp"] = entry.Timestamp.Format(time.RFC3339Nano)
		case "level":
			obj["level"] = entry.Level
		case "component":
			obj["component"] = entry.Component
		case "message":
			obj["message"] = entry.Message
		case "caller":
			if entry.Caller != "" {
				obj["caller"] = entry.Caller
			}
		}
	}

	// Add extra fields if present
	if len(entry.Extra) > 0 {
		obj["extra"] = entry.Extra
	}

	// Marshal to JSON (ignore error for now)
	data, _ := json.Marshal(obj)
	return string(data)
}

// log enqueues a log entry (non-blocking, lock-free)
func (l *Logger) log(level string, format string, args ...interface{}) {
	// Don't log debug messages if debug is disabled
	if level == "DEBUG" && !l.config.Debug {
		return
	}

	entry := &logEntry{
		Timestamp: time.Now(),
		Level:     level,
		Component: l.component,
		Message:   fmt.Sprintf(format, args...),
	}

	// Non-blocking send (drop if buffer full to avoid blocking)
	select {
	case l.logChan <- entry:
	default:
		// Buffer full, drop message (alternative: could block or write to stderr)
		fmt.Fprintf(os.Stderr, "WARNING: Log buffer full for component %s, dropping message\n", l.component)
	}
}

// Public logging methods
func (l *Logger) Infof(format string, args ...interface{})  { l.log("INFO", format, args...) }
func (l *Logger) Errorf(format string, args ...interface{}) { l.log("ERROR", format, args...) }
func (l *Logger) Warnf(format string, args ...interface{})  { l.log("WARN", format, args...) }
func (l *Logger) Debugf(format string, args ...interface{}) { l.log("DEBUG", format, args...) }

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log("FATAL", format, args...)
	l.Close()
	os.Exit(1)
}

// Close closes the logger and flushes remaining entries
func (l *Logger) Close() error {
	var closeErr error
	l.once.Do(func() {
		close(l.closeCh)
		l.wg.Wait()

		// Close component file (generic file is shared, close in CloseAll)
		if l.componentFile != nil {
			if err := l.componentFile.Close(); err != nil {
				closeErr = err
			}
		}
	})
	return closeErr
}

// CloseAll closes all component loggers and the generic log file
// Returns the first error encountered, but attempts to close all loggers
func CloseAll() error {
	globalMu.Lock()
	defer globalMu.Unlock()

	var firstErr error

	// Close all component loggers
	for _, logger := range componentLoggers {
		if err := logger.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// Close generic log file
	if genericLogFile != nil {
		if err := genericLogFile.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
		genericLogFile = nil
	}

	// Clear component logger cache
	componentLoggers = make(map[string]*Logger)

	return firstErr
}

// Helper: get default log directory
func getDefaultLogDir() string {
	if dir := os.Getenv("OAK_LOG_DIR"); dir != "" {
		return dir
	}

	// Platform-specific defaults
	if os.PathSeparator == '\\' {
		// Windows: %ProgramData%\Oak\logs or .\logs
		if programData := os.Getenv("ProgramData"); programData != "" {
			return filepath.Join(programData, "Oak", "logs")
		}
		return "logs"
	}

	// Unix-like: /var/log/oak if writable, else ./logs
	if isWritable("/var/log") {
		return "/var/log/oak"
	}
	return "logs"
}

// Helper: check if directory is writable
func isWritable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || !info.IsDir() {
		return false
	}

	testFile := filepath.Join(path, ".oak-write-test")
	file, err := os.Create(testFile)
	if err != nil {
		return false
	}
	file.Close()
	os.Remove(testFile)
	return true
}