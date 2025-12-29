package backend

import (
	"os/exec"
	"time"

	"github.com/automaat/local-k8s-manager/internal/logger"
)

// Exec executes a command and returns its combined output
func Exec(name string, args ...string) ([]byte, error) {
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
func IsCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}
