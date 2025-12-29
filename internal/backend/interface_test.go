package backend

import (
	"testing"
)

func TestCreateOptions(t *testing.T) {
	opts := CreateOptions{
		Workers: 3,
	}

	if opts.Workers != 3 {
		t.Errorf("expected Workers to be 3, got %d", opts.Workers)
	}
}

func TestProviderNames(t *testing.T) {
	tests := []struct {
		name     string
		provider Provider
	}{
		{"k3d provider", NewK3dProvider()},
		{"kind provider", NewKindProvider()},
		{"minikube provider", NewMinikubeProvider()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name := tt.provider.Name()
			if name == "" {
				t.Errorf("Provider.Name() returned empty string")
			}
		})
	}
}

func TestProviderIsInstalled(t *testing.T) {
	providers := []Provider{
		NewK3dProvider(),
		NewKindProvider(),
		NewMinikubeProvider(),
	}

	for _, p := range providers {
		t.Run(p.Name(), func(t *testing.T) {
			// Just verify it doesn't panic
			_ = p.IsInstalled()
		})
	}
}
