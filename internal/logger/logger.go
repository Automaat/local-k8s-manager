package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	maxLogSize  = 10 * 1024 * 1024 // 10MB
	maxLogFiles = 3
)

var (
	logFile *os.File
	logPath string
	mu      sync.Mutex
	enabled bool
)

// Init initializes the logger and creates the log file
// Logs are written to ~/.local/share/lkm/lkm.log or $XDG_DATA_HOME/lkm/lkm.log
func Init() error {
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to get home directory: %w", err)
		}
		dataDir = filepath.Join(home, ".local", "share")
	}

	logDir := filepath.Join(dataDir, "lkm")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logPath = filepath.Join(logDir, "lkm.log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	logFile = f
	enabled = true
	return nil
}

// Log writes an operation log entry
func Log(operation string, details map[string]interface{}) {
	if !enabled {
		return
	}

	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     "info",
		"operation": operation,
		"details":   details,
	}

	writeEntry(entry)
}

// LogError writes an error log entry
func LogError(operation string, err error, details map[string]interface{}) {
	if !enabled {
		return
	}

	entry := map[string]interface{}{
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"level":     "error",
		"operation": operation,
		"error":     err.Error(),
		"details":   details,
	}

	writeEntry(entry)
}

func writeEntry(entry map[string]interface{}) {
	mu.Lock()
	defer mu.Unlock()

	if logFile == nil {
		return
	}

	// Check if rotation is needed
	if err := rotateIfNeeded(); err != nil {
		fmt.Fprintf(os.Stderr, "lkm: failed to rotate log: %v\n", err)
	}

	data, err := json.Marshal(entry)
	if err != nil {
		fmt.Fprintf(os.Stderr, "lkm: failed to marshal log entry: %v\n", err)
		return
	}

	if _, err := logFile.Write(data); err != nil {
		fmt.Fprintf(os.Stderr, "lkm: failed to write log entry: %v\n", err)
		return
	}
	if _, err := logFile.Write([]byte("\n")); err != nil {
		fmt.Fprintf(os.Stderr, "lkm: failed to write log newline: %v\n", err)
	}
}

// rotateIfNeeded checks if the log file exceeds maxLogSize and rotates if necessary
// Must be called with mu held
func rotateIfNeeded() error {
	if logPath == "" || logFile == nil {
		return nil
	}

	info, err := logFile.Stat()
	if err != nil {
		return err
	}

	if info.Size() < maxLogSize {
		return nil
	}

	// Close current log file
	if err := logFile.Close(); err != nil {
		return err
	}

	// Rotate existing log files
	for i := maxLogFiles - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", logPath, i)
		newPath := fmt.Sprintf("%s.%d", logPath, i+1)
		_ = os.Rename(oldPath, newPath)
	}

	// Move current log to .1
	if err := os.Rename(logPath, logPath+".1"); err != nil {
		return err
	}

	// Remove oldest log file if it exists
	oldestPath := fmt.Sprintf("%s.%d", logPath, maxLogFiles+1)
	_ = os.Remove(oldestPath)

	// Open new log file
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	logFile = f
	return nil
}

// Close closes the log file
func Close() {
	mu.Lock()
	defer mu.Unlock()

	if logFile != nil {
		_ = logFile.Close()
		logFile = nil
	}
	enabled = false
}
