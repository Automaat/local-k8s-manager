package backend

import (
	"os/exec"
	"strings"
	"time"

	"github.com/automaat/local-k8s-manager/internal/logger"
)

// CommandExecutor is an interface for executing commands
type CommandExecutor interface {
	Exec(name string, args ...string) ([]byte, error)
	IsCommandAvailable(name string) bool
}

// DefaultExecutor is the default implementation of CommandExecutor
type DefaultExecutor struct{}

// Exec executes a command and returns its combined output
func (e *DefaultExecutor) Exec(name string, args ...string) ([]byte, error) {
	start := time.Now()
	cmd := exec.Command(name, args...)
	output, err := cmd.CombinedOutput()
	duration := time.Since(start)

	details := map[string]interface{}{
		"command":  name,
		"args":     args,
		"duration": duration.String(),
	}

	if err != nil {
		logger.LogError("exec", err, details)
	} else {
		logger.Log("exec", details)
	}

	return output, err
}

// IsCommandAvailable checks if a command is available in PATH
func (e *DefaultExecutor) IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

// Global executor instance (can be replaced for testing)
var executor CommandExecutor = &DefaultExecutor{}

// Exec is a convenience function that uses the global executor
func Exec(name string, args ...string) ([]byte, error) {
	return executor.Exec(name, args...)
}

// IsCommandAvailable is a convenience function that uses the global executor
func IsCommandAvailable(name string) bool {
	return executor.IsCommandAvailable(name)
}

// SetExecutor sets the global executor (for testing)
func SetExecutor(e CommandExecutor) {
	executor = e
}

// ParseDockerError checks if an error message indicates Docker is not running
// and returns a user-friendly error message if so
func ParseDockerError(output string) string {
	lowerOutput := strings.ToLower(output)

	// Check for common Docker daemon errors
	if strings.Contains(lowerOutput, "cannot connect to the docker daemon") ||
		strings.Contains(lowerOutput, "is the docker daemon running") ||
		strings.Contains(lowerOutput, "docker: not found") ||
		strings.Contains(lowerOutput, "docker daemon is not running") {
		return "Docker is not running. Please start Docker and try again."
	}

	return ""
}
