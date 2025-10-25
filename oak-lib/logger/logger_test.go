package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestMain runs before/after all tests to ensure cleanup
func TestMain(m *testing.M) {
	code := m.Run()
	if err := CloseAll(); err != nil {
		// Log error but don't change exit code - test results are more important
		println("WARNING: CloseAll() error:", err.Error())
	}
	os.Exit(code)
}

func TestNewComponent(t *testing.T) {
	// Use temp directory for tests
	tempDir := t.TempDir()
	cfg := &Config{
		LogDir:   tempDir,
		Format:   FormatText,
		ToStdout: false,
		ToFile:   true,
		BufSize:  100,
	}
	SetGlobalConfig(cfg)
	t.Cleanup(func() {
		if err := CloseAll(); err != nil {
			t.Logf("CloseAll() error: %v", err)
		}
	})

	logger := NewComponent("test")
	if logger == nil {
		t.Fatal("NewComponent() returned nil")
	}

	if logger.component != "test" {
		t.Errorf("component = %s, want test", logger.component)
	}

	// Verify log files were created
	genericPath := filepath.Join(tempDir, "oak.log")
	componentPath := filepath.Join(tempDir, "test.log")

	// Write a log entry
	logger.Infof("test message")

	// Give async logger time to write
	time.Sleep(50 * time.Millisecond)
	logger.Close()

	// Check files exist
	if _, err := os.Stat(genericPath); os.IsNotExist(err) {
		t.Errorf("Generic log file not created: %s", genericPath)
	}

	if _, err := os.Stat(componentPath); os.IsNotExist(err) {
		t.Errorf("Component log file not created: %s", componentPath)
	}
}

func TestLogLevels(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &Config{
		LogDir:   tempDir,
		Format:   FormatText,
		ToStdout: false,
		ToFile:   true,
		Debug:    true, // Enable debug
		BufSize:  100,
	}
	SetGlobalConfig(cfg)

	logger := NewComponent("levels-test")
	defer logger.Close()

	// Write all log levels
	logger.Infof("info message")
	logger.Warnf("warn message")
	logger.Errorf("error message")
	logger.Debugf("debug message")

	// Give async logger time to write
	time.Sleep(50 * time.Millisecond)

	// Read component log file
	logPath := filepath.Join(tempDir, "levels-test.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logStr := string(content)

	// Verify all levels are present
	if !strings.Contains(logStr, "[INFO]") {
		t.Error("INFO level not found in logs")
	}
	if !strings.Contains(logStr, "[WARN]") {
		t.Error("WARN level not found in logs")
	}
	if !strings.Contains(logStr, "[ERROR]") {
		t.Error("ERROR level not found in logs")
	}
	if !strings.Contains(logStr, "[DEBUG]") {
		t.Error("DEBUG level not found in logs")
	}

	// Verify messages
	if !strings.Contains(logStr, "info message") {
		t.Error("Info message not found in logs")
	}
	if !strings.Contains(logStr, "warn message") {
		t.Error("Warn message not found in logs")
	}
	if !strings.Contains(logStr, "error message") {
		t.Error("Error message not found in logs")
	}
	if !strings.Contains(logStr, "debug message") {
		t.Error("Debug message not found in logs")
	}
}

func TestDebugDisabled(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &Config{
		LogDir:   tempDir,
		Format:   FormatText,
		ToStdout: false,
		ToFile:   true,
		Debug:    false, // Disable debug
		BufSize:  100,
	}
	SetGlobalConfig(cfg)

	logger := NewComponent("debug-test")
	defer logger.Close()

	logger.Infof("info message")
	logger.Debugf("debug message should not appear")

	time.Sleep(50 * time.Millisecond)

	// Read log file
	logPath := filepath.Join(tempDir, "debug-test.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logStr := string(content)

	if !strings.Contains(logStr, "info message") {
		t.Error("Info message should be in logs")
	}
	if strings.Contains(logStr, "debug message") {
		t.Error("Debug message should NOT be in logs when debug is disabled")
	}
}

func TestJSONFormat(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &Config{
		LogDir:   tempDir,
		Format:   FormatJSON,
		ToStdout: false,
		ToFile:   true,
		Fields:   []string{"timestamp", "level", "component", "message"},
		BufSize:  100,
	}
	SetGlobalConfig(cfg)

	logger := NewComponent("json-test")
	defer logger.Close()

	logger.Infof("json test message")

	time.Sleep(50 * time.Millisecond)

	// Read log file
	logPath := filepath.Join(tempDir, "json-test.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	logStr := string(content)

	// Verify JSON format
	if !strings.Contains(logStr, `"level":"INFO"`) {
		t.Error("JSON should contain level field")
	}
	if !strings.Contains(logStr, `"component":"json-test"`) {
		t.Error("JSON should contain component field")
	}
	if !strings.Contains(logStr, `"message":"json test message"`) {
		t.Error("JSON should contain message field")
	}
	if !strings.Contains(logStr, `"timestamp"`) {
		t.Error("JSON should contain timestamp field")
	}
}

func TestComponentSeparation(t *testing.T) {
	tempDir := t.TempDir()
	cfg := &Config{
		LogDir:   tempDir,
		Format:   FormatText,
		ToStdout: false,
		ToFile:   true,
		BufSize:  100,
	}
	SetGlobalConfig(cfg)

	// Create loggers for different components
	logger1 := NewComponent("component1")
	logger2 := NewComponent("component2")
	defer logger1.Close()
	defer logger2.Close()

	logger1.Infof("message from component 1")
	logger2.Infof("message from component 2")

	time.Sleep(50 * time.Millisecond)

	// Read component1.log
	content1, err := os.ReadFile(filepath.Join(tempDir, "component1.log"))
	if err != nil {
		t.Fatalf("Failed to read component1.log: %v", err)
	}

	// Read component2.log
	content2, err := os.ReadFile(filepath.Join(tempDir, "component2.log"))
	if err != nil {
		t.Fatalf("Failed to read component2.log: %v", err)
	}

	// Verify component1.log only has component1 messages
	if !strings.Contains(string(content1), "message from component 1") {
		t.Error("component1.log should contain component 1 message")
	}
	if strings.Contains(string(content1), "message from component 2") {
		t.Error("component1.log should NOT contain component 2 message")
	}

	// Verify component2.log only has component2 messages
	if !strings.Contains(string(content2), "message from component 2") {
		t.Error("component2.log should contain component 2 message")
	}
	if strings.Contains(string(content2), "message from component 1") {
		t.Error("component2.log should NOT contain component 1 message")
	}

	// Verify generic log has both
	genericContent, err := os.ReadFile(filepath.Join(tempDir, "oak.log"))
	if err != nil {
		t.Fatalf("Failed to read oak.log: %v", err)
	}

	if !strings.Contains(string(genericContent), "message from component 1") {
		t.Error("oak.log should contain component 1 message")
	}
	if !strings.Contains(string(genericContent), "message from component 2") {
		t.Error("oak.log should contain component 2 message")
	}
}

func TestLoggerSingleton(t *testing.T) {
	cfg := &Config{
		LogDir:   t.TempDir(),
		Format:   FormatText,
		ToStdout: false,
		ToFile:   false, // Don't create files for this test
		BufSize:  100,
	}
	SetGlobalConfig(cfg)

	// Create same component twice
	logger1 := NewComponent("singleton")
	logger2 := NewComponent("singleton")

	// Should return same instance
	if logger1 != logger2 {
		t.Error("NewComponent should return same instance for same component name")
	}
}

func TestConcurrentLogging(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping concurrent test in short mode")
	}

	tempDir := t.TempDir()
	cfg := &Config{
		LogDir:   tempDir,
		Format:   FormatText,
		ToStdout: false,
		ToFile:   true,
		BufSize:  1000,
	}
	SetGlobalConfig(cfg)

	logger := NewComponent("concurrent")
	defer logger.Close()

	// Log from multiple goroutines
	const numGoroutines = 10
	const messagesPerGoroutine = 100

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < messagesPerGoroutine; j++ {
				logger.Infof("goroutine %d message %d", id, j)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	time.Sleep(100 * time.Millisecond)

	// Read log file
	logPath := filepath.Join(tempDir, "concurrent.log")
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	// Count lines
	lines := strings.Split(string(content), "\n")
	nonEmptyLines := 0
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			nonEmptyLines++
		}
	}

	expectedLines := numGoroutines * messagesPerGoroutine
	if nonEmptyLines != expectedLines {
		t.Errorf("Expected %d log lines, got %d", expectedLines, nonEmptyLines)
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()

	if cfg.Format != FormatText {
		t.Errorf("Default format = %v, want FormatText", cfg.Format)
	}

	if !cfg.ToStdout {
		t.Error("Default should log to stdout")
	}

	if !cfg.ToFile {
		t.Error("Default should log to file")
	}

	if cfg.BufSize != 1000 {
		t.Errorf("Default buffer size = %d, want 1000", cfg.BufSize)
	}

	if len(cfg.Fields) == 0 {
		t.Error("Default should have fields configured")
	}
}

func TestGetDefaultLogDir(t *testing.T) {
	// Save and restore env
	oldLogDir := os.Getenv("OAK_LOG_DIR")
	defer os.Setenv("OAK_LOG_DIR", oldLogDir)

	// Test with OAK_LOG_DIR set
	os.Setenv("OAK_LOG_DIR", "/custom/log/dir")
	dir := getDefaultLogDir()
	if dir != "/custom/log/dir" {
		t.Errorf("getDefaultLogDir() = %s, want /custom/log/dir", dir)
	}

	// Test without OAK_LOG_DIR
	os.Unsetenv("OAK_LOG_DIR")
	dir = getDefaultLogDir()
	if dir == "" {
		t.Error("getDefaultLogDir() should return non-empty string")
	}
}