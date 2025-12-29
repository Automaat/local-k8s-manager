package backend

import (
	"bytes"
	"fmt"
	"strings"
	"time"

	"github.com/automaat/local-k8s-manager/internal/models"
)

// KindProvider implements the Provider interface for kind
type KindProvider struct{}

// NewKindProvider creates a new kind provider
func NewKindProvider() *KindProvider {
	return &KindProvider{}
}

// Name returns the provider name
func (p *KindProvider) Name() string {
	return "kind"
}

// IsInstalled checks if kind is installed
func (p *KindProvider) IsInstalled() bool {
	return IsCommandAvailable("kind")
}

// List returns all kind clusters
func (p *KindProvider) List() ([]models.Cluster, error) {
	output, err := Exec("kind", "get", "clusters")

	// kind may return "No kind clusters found" with either exit code 0 or non-zero
	// in both cases we should return an empty list instead of an error
	if bytes.Contains(output, []byte("No kind clusters found")) || len(output) == 0 {
		return []models.Cluster{}, nil
	}

	if err != nil {
		return nil, fmt.Errorf("failed to list kind clusters: %w", err)
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	clusters := make([]models.Cluster, 0, len(lines))

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// kind clusters are always running (no stop/start support)
		clusters = append(clusters, models.Cluster{
			Name:      line,
			Provider:  "kind",
			Status:    models.StatusRunning,
			Nodes:     0, // kind get clusters output does not include node count; 0 means "unknown"
			CreatedAt: time.Time{},
		})
	}

	return clusters, nil
}

// Create creates a new kind cluster
func (p *KindProvider) Create(name string, opts CreateOptions, outputChan ...chan<- string) (string, error) {
	args := []string{"create", "cluster", "--name", name}

	// kind doesn't have a simple worker flag, would need config file
	// ignoring worker count for now as per plan (CLI form only)

	var output []byte
	var outputStr string
	var err error

	// Use streaming if channel provided
	if len(outputChan) > 0 && outputChan[0] != nil {
		outputStr, err = ExecStreaming("kind", args, outputChan[0])
	} else {
		output, err = Exec("kind", args...)
		outputStr = string(output)
	}

	if err != nil {
		if len(outputStr) > 0 {
			if dockerErr := ParseDockerError(outputStr); dockerErr != "" {
				return outputStr, fmt.Errorf("%s", dockerErr)
			}
			return outputStr, fmt.Errorf("failed to create kind cluster: %s", outputStr)
		}
		return outputStr, fmt.Errorf("failed to create kind cluster: %w", err)
	}
	return outputStr, nil
}

// Delete deletes a kind cluster
func (p *KindProvider) Delete(name string) error {
	output, err := Exec("kind", "delete", "cluster", "--name", name)
	if err != nil {
		if len(output) > 0 {
			if dockerErr := ParseDockerError(string(output)); dockerErr != "" {
				return fmt.Errorf("%s", dockerErr)
			}
			return fmt.Errorf("failed to delete kind cluster: %s", string(output))
		}
		return fmt.Errorf("failed to delete kind cluster: %w", err)
	}
	return nil
}

// Start is not supported by kind
func (p *KindProvider) Start(name string) error {
	return fmt.Errorf("kind does not support starting clusters")
}

// Stop is not supported by kind
func (p *KindProvider) Stop(name string) error {
	return fmt.Errorf("kind does not support stopping clusters")
}
