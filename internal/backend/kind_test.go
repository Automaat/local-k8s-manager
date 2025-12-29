package backend

import (
	"testing"
)

func TestKindProvider_Name(t *testing.T) {
	p := NewKindProvider()
	if p.Name() != "kind" {
		t.Errorf("expected name 'kind', got '%s'", p.Name())
	}
}

func TestKindProvider_IsInstalled(t *testing.T) {
	p := NewKindProvider()
	result := p.IsInstalled()
	_ = result
}

func TestNewKindProvider(t *testing.T) {
	p := NewKindProvider()
	if p == nil {
		t.Error("NewKindProvider() returned nil")
	}
}
