package backend

import "github.com/automaat/local-k8s-manager/internal/models"

// Provider defines the interface for Kubernetes cluster management tools
type Provider interface {
	// Name returns the provider's name (e.g., "k3d", "kind", "minikube")
	Name() string

	// IsInstalled checks if the provider's CLI tool is available
	IsInstalled() bool

	// List returns all clusters managed by this provider
	List() ([]models.Cluster, error)

	// Create creates a new cluster with the given name and options
	// Returns the command output and any error
	// If outputChan is provided, output will be streamed line-by-line
	Create(name string, opts CreateOptions, outputChan ...chan<- string) (string, error)

	// Delete removes a cluster
	Delete(name string) error

	// Start starts a stopped cluster
	Start(name string) error

	// Stop stops a running cluster
	Stop(name string) error
}

// CreateOptions contains options for creating a new cluster
type CreateOptions struct {
	Workers int
}
