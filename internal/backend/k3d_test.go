package backend

import (
	"testing"
)

func TestK3dProvider_Name(t *testing.T) {
	p := NewK3dProvider()
	if p.Name() != "k3d" {
		t.Errorf("expected name 'k3d', got '%s'", p.Name())
	}
}

func TestK3dProvider_IsInstalled(t *testing.T) {
	p := NewK3dProvider()
	// Just verify it returns a boolean without panicking
	result := p.IsInstalled()
	_ = result // Use the result to avoid unused variable warning
}

func TestNewK3dProvider(t *testing.T) {
	p := NewK3dProvider()
	if p == nil {
		t.Error("NewK3dProvider() returned nil")
	}
}
