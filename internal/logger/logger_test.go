package logger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	// Reset global state
	logFile = nil
	logPath = ""
	enabled = false

	err := Init()
	if err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer Close()

	if !enabled {
		t.Error("Logger should be enabled after Init()")
	}

	if logFile == nil {
		t.Error("logFile should not be nil after Init()")
	}

	expectedPath := filepath.Join(tmpDir, "lkm", "lkm.log")
	if logPath != expectedPath {
		t.Errorf("Expected logPath to be %s, got %s", expectedPath, logPath)
	}

	// Check if log file was created
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		t.Errorf("Log file was not created at %s", logPath)
	}
}

func TestLog(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	logFile = nil
	logPath = ""
	enabled = false

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer Close()

	// Log an entry
	details := map[string]interface{}{
		"test": "value",
		"num":  42,
	}
	Log("test.operation", details)

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 log line, got %d", len(lines))
	}

	// Parse JSON
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	// Verify fields
	if entry["level"] != "info" {
		t.Errorf("Expected level 'info', got %v", entry["level"])
	}
	if entry["operation"] != "test.operation" {
		t.Errorf("Expected operation 'test.operation', got %v", entry["operation"])
	}
	if entry["timestamp"] == nil {
		t.Error("Expected timestamp field")
	}

	detailsMap, ok := entry["details"].(map[string]interface{})
	if !ok {
		t.Fatal("Expected details to be a map")
	}
	if detailsMap["test"] != "value" {
		t.Errorf("Expected details.test to be 'value', got %v", detailsMap["test"])
	}
}

func TestLogError(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	logFile = nil
	logPath = ""
	enabled = false

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer Close()

	// Log an error
	testErr := os.ErrNotExist
	details := map[string]interface{}{
		"file": "/some/path",
	}
	LogError("file.read", testErr, details)

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read log file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 1 {
		t.Fatalf("Expected 1 log line, got %d", len(lines))
	}

	// Parse JSON
	var entry map[string]interface{}
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("Failed to parse log entry: %v", err)
	}

	// Verify fields
	if entry["level"] != "error" {
		t.Errorf("Expected level 'error', got %v", entry["level"])
	}
	if entry["operation"] != "file.read" {
		t.Errorf("Expected operation 'file.read', got %v", entry["operation"])
	}
	if entry["error"] == nil {
		t.Error("Expected error field")
	}
	if !strings.Contains(entry["error"].(string), "does not exist") && !strings.Contains(entry["error"].(string), "no such file") {
		t.Errorf("Expected error message about file not existing, got %v", entry["error"])
	}
}

func TestLogRotation(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	logFile = nil
	logPath = ""
	enabled = false

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}
	defer Close()

	// For this test we artificially trigger rotation by writing directly to the
	// underlying log file handle until its size exceeds maxLogSize, then calling
	// Log once to cause the rotation logic to run.
	largeData := make([]byte, maxLogSize+1000)
	for i := range largeData {
		largeData[i] = 'a'
	}

	if _, err := logFile.Write(largeData); err != nil {
		t.Fatalf("Failed to write large data: %v", err)
	}

	// Now log something, which should trigger rotation
	Log("test.after.rotation", map[string]interface{}{"test": "value"})

	// Verify rotation occurred by checking:
	// 1. Rotated file exists and contains the large data
	rotatedPath := logPath + ".1"
	rotatedInfo, err := os.Stat(rotatedPath)
	if os.IsNotExist(err) {
		t.Fatal("Rotated log file should exist")
	}
	if err != nil {
		t.Fatalf("Failed to stat rotated file: %v", err)
	}
	if rotatedInfo.Size() < maxLogSize {
		t.Errorf("Rotated file should contain the large data, got %d bytes", rotatedInfo.Size())
	}

	// 2. New log file is small and contains the logged entry
	info, err := os.Stat(logPath)
	if err != nil {
		t.Fatalf("Failed to stat new log file: %v", err)
	}
	if info.Size() > 1000 {
		t.Errorf("New log file should be small, got %d bytes", info.Size())
	}

	// 3. New log file contains the entry we just logged
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read new log file: %v", err)
	}
	if !strings.Contains(string(content), "test.after.rotation") {
		t.Error("New log file should contain the entry logged after rotation")
	}
}

func TestLogDisabled(t *testing.T) {
	// Setup with logging disabled
	logFile = nil
	logPath = ""
	enabled = false

	// These should not panic
	Log("test", nil)
	LogError("test", os.ErrNotExist, nil)
}

func TestClose(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)
	logFile = nil
	logPath = ""
	enabled = false

	if err := Init(); err != nil {
		t.Fatalf("Init() failed: %v", err)
	}

	// Close the logger
	Close()

	if enabled {
		t.Error("Logger should be disabled after Close()")
	}

	if logFile != nil {
		t.Error("logFile should be nil after Close()")
	}

	// Logging after close should not panic
	Log("test", nil)
}
