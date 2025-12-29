package backend

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/automaat/local-k8s-manager/internal/models"
)

// MinikubeProvider implements the Provider interface for minikube
type MinikubeProvider struct{}

// NewMinikubeProvider creates a new minikube provider
func NewMinikubeProvider() *MinikubeProvider {
	return &MinikubeProvider{}
}

// Name returns the provider name
func (p *MinikubeProvider) Name() string {
	return "minikube"
}

// IsInstalled checks if minikube is installed
func (p *MinikubeProvider) IsInstalled() bool {
	return IsCommandAvailable("minikube")
}

// minikubeProfileList represents the JSON structure from minikube profile list
type minikubeProfileList struct {
	Valid   []minikubeProfile `json:"valid"`
	Invalid []minikubeProfile `json:"invalid"`
}

type minikubeProfile struct {
	Name   string         `json:"Name"`
	Status string         `json:"Status"`
	Config minikubeConfig `json:"Config"`
}

type minikubeConfig struct {
	Nodes []minikubeNode `json:"Nodes"`
}

type minikubeNode struct {
	Name string `json:"Name"`
}

// List returns all minikube clusters
func (p *MinikubeProvider) List() ([]models.Cluster, error) {
	output, err := Exec("minikube", "profile", "list", "-o", "json")
	if err != nil {
		// minikube returns error when no profiles exist
		// check if output is empty or contains error message
		if len(output) == 0 {
			return []models.Cluster{}, nil
		}
		// Try to parse anyway, might have valid profiles with errors
	}

	var profileList minikubeProfileList
	if err := json.Unmarshal(output, &profileList); err != nil {
		return nil, fmt.Errorf("failed to parse minikube output: %w", err)
	}

	clusters := make([]models.Cluster, 0, len(profileList.Valid))
	for _, profile := range profileList.Valid {
		status := parseMinikubeStatus(profile.Status)
		nodes := len(profile.Config.Nodes)

		clusters = append(clusters, models.Cluster{
			Name:      profile.Name,
			Provider:  "minikube",
			Status:    status,
			Nodes:     nodes,
			CreatedAt: time.Time{}, // minikube doesn't provide creation time in list
		})
	}

	return clusters, nil
}

func parseMinikubeStatus(status string) models.Status {
	switch status {
	case "Running":
		return models.StatusRunning
	case "Stopped":
		return models.StatusStopped
	default:
		return models.StatusUnknown
	}
}

// Create creates a new minikube cluster
func (p *MinikubeProvider) Create(name string, opts CreateOptions) (string, error) {
	args := []string{"start", "--profile", name}
	if opts.Workers > 0 {
		// minikube uses --nodes for total node count (control-plane + workers)
		// so we add 1 for the control-plane node
		args = append(args, "--nodes", fmt.Sprintf("%d", opts.Workers+1))
	}

	output, err := Exec("minikube", args...)
	outputStr := string(output)
	if err != nil {
		if len(output) > 0 {
			if dockerErr := ParseDockerError(outputStr); dockerErr != "" {
				return outputStr, fmt.Errorf("%s", dockerErr)
			}
			return outputStr, fmt.Errorf("failed to create minikube cluster: %s", outputStr)
		}
		return outputStr, fmt.Errorf("failed to create minikube cluster: %w", err)
	}
	return outputStr, nil
}

// Delete deletes a minikube cluster
func (p *MinikubeProvider) Delete(name string) error {
	output, err := Exec("minikube", "delete", "-p", name)
	if err != nil {
		if len(output) > 0 {
			if dockerErr := ParseDockerError(string(output)); dockerErr != "" {
				return fmt.Errorf("%s", dockerErr)
			}
			return fmt.Errorf("failed to delete minikube cluster: %s", string(output))
		}
		return fmt.Errorf("failed to delete minikube cluster: %w", err)
	}
	return nil
}

// Start starts a minikube cluster
func (p *MinikubeProvider) Start(name string) error {
	output, err := Exec("minikube", "start", "-p", name)
	if err != nil {
		if len(output) > 0 {
			if dockerErr := ParseDockerError(string(output)); dockerErr != "" {
				return fmt.Errorf("%s", dockerErr)
			}
			return fmt.Errorf("failed to start minikube cluster: %s", string(output))
		}
		return fmt.Errorf("failed to start minikube cluster: %w", err)
	}
	return nil
}

// Stop stops a minikube cluster
func (p *MinikubeProvider) Stop(name string) error {
	output, err := Exec("minikube", "stop", "-p", name)
	if err != nil {
		if len(output) > 0 {
			if dockerErr := ParseDockerError(string(output)); dockerErr != "" {
				return fmt.Errorf("%s", dockerErr)
			}
			return fmt.Errorf("failed to stop minikube cluster: %s", string(output))
		}
		return fmt.Errorf("failed to stop minikube cluster: %w", err)
	}
	return nil
}
