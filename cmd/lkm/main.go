package main

import (
	"fmt"
	"os"

	"github.com/automaat/local-k8s-manager/internal/backend"
	"github.com/automaat/local-k8s-manager/internal/logger"
)

func main() {
	// Initialize logger
	if err := logger.Init(); err != nil {
		fmt.Printf("Warning: logging disabled: %v\n", err)
	}
	defer logger.Close()

	providers := []backend.Provider{
		backend.NewK3dProvider(),
		backend.NewKindProvider(),
		backend.NewMinikubeProvider(),
	}

	// Filter to installed providers
	var available []backend.Provider
	for _, p := range providers {
		if p.IsInstalled() {
			available = append(available, p)
		}
	}

	if len(available) == 0 {
		fmt.Println("No cluster tools found.")
		fmt.Println("Install one of: k3d, kind, minikube")
		logger.LogError("startup", fmt.Errorf("no cluster tools found"), nil)
		logger.Close()
		os.Exit(1)
	}

	// Log startup
	providerNames := make([]string, len(available))
	for i, p := range available {
		providerNames[i] = p.Name()
	}
	logger.Log("startup", map[string]interface{}{
		"providers": providerNames,
	})

	fmt.Println("lkm - Local Kubernetes Manager")
	fmt.Printf("Available providers: %v\n", providerNames)
	fmt.Println("TUI coming soon...")
}
