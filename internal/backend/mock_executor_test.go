package backend_test

import (
	"fmt"
)

// MockExecutor is a mock implementation of CommandExecutor for testing
type MockExecutor struct {
	ExecFunc               func(name string, args ...string) ([]byte, error)
	IsCommandAvailableFunc func(name string) bool
}

func (m *MockExecutor) Exec(name string, args ...string) ([]byte, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(name, args...)
	}
	return nil, fmt.Errorf("not implemented")
}

func (m *MockExecutor) IsCommandAvailable(name string) bool {
	if m.IsCommandAvailableFunc != nil {
		return m.IsCommandAvailableFunc(name)
	}
	return false
}

