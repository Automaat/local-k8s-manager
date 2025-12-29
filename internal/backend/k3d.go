package backend

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/automaat/local-k8s-manager/internal/models"
)

// K3dProvider implements the Provider interface for k3d
type K3dProvider struct{}

// NewK3dProvider creates a new k3d provider
func NewK3dProvider() *K3dProvider {
	return &K3dProvider{}
}

// Name returns the provider name
func (p *K3dProvider) Name() string {
	return "k3d"
}

// IsInstalled checks if k3d is installed
func (p *K3dProvider) IsInstalled() bool {
	return IsCommandAvailable("k3d")
}

// k3dCluster represents the JSON structure from k3d cluster list
type k3dCluster struct {
	Name        string    `json:"name"`
	Nodes       []k3dNode `json:"nodes"`
	ServerCount int       `json:"serversCount"`
}

type k3dNode struct {
	Name  string `json:"name"`
	Role  string `json:"role"`
	State struct {
		Running bool `json:"running"`
	} `json:"state"`
}

// List returns all k3d clusters
func (p *K3dProvider) List() ([]models.Cluster, error) {
	output, err := Exec("k3d", "cluster", "list", "-o", "json")
	if err != nil {
		return nil, fmt.Errorf("failed to list k3d clusters: %w", err)
	}

	var k3dClusters []k3dCluster
	if err := json.Unmarshal(output, &k3dClusters); err != nil {
		return nil, fmt.Errorf("failed to parse k3d output: %w", err)
	}

	clusters := make([]models.Cluster, 0, len(k3dClusters))
	for _, c := range k3dClusters {
		status := models.StatusStopped
		runningNodes := 0
		for _, node := range c.Nodes {
			if node.State.Running {
				runningNodes++
			}
		}
		if runningNodes == len(c.Nodes) && len(c.Nodes) > 0 {
			status = models.StatusRunning
		} else if runningNodes > 0 {
			status = models.StatusUnknown
		}

		clusters = append(clusters, models.Cluster{
			Name:      c.Name,
			Provider:  "k3d",
			Status:    status,
			Nodes:     len(c.Nodes),
			CreatedAt: time.Time{}, // k3d doesn't provide creation time
		})
	}

	return clusters, nil
}

// Create creates a new k3d cluster
func (p *K3dProvider) Create(name string, opts CreateOptions) error {
	args := []string{"cluster", "create", name}
	if opts.Workers > 0 {
		args = append(args, "--agents", strconv.Itoa(opts.Workers))
	}

	_, err := Exec("k3d", args...)
	if err != nil {
		return fmt.Errorf("failed to create k3d cluster: %w", err)
	}
	return nil
}

// Delete deletes a k3d cluster
func (p *K3dProvider) Delete(name string) error {
	_, err := Exec("k3d", "cluster", "delete", name)
	if err != nil {
		return fmt.Errorf("failed to delete k3d cluster: %w", err)
	}
	return nil
}

// Start starts a k3d cluster
func (p *K3dProvider) Start(name string) error {
	_, err := Exec("k3d", "cluster", "start", name)
	if err != nil {
		return fmt.Errorf("failed to start k3d cluster: %w", err)
	}
	return nil
}

// Stop stops a k3d cluster
func (p *K3dProvider) Stop(name string) error {
	_, err := Exec("k3d", "cluster", "stop", name)
	if err != nil {
		return fmt.Errorf("failed to stop k3d cluster: %w", err)
	}
	return nil
}
