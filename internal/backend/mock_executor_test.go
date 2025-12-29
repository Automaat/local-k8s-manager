package backend_test

import (
	"encoding/json"
	"fmt"

	"github.com/automaat/local-k8s-manager/internal/backend"
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

// Helper functions for common mock responses
func mockK3dListResponse() []byte {
	clusters := []backend.K3dCluster{
		{
			Name: "test-cluster",
			Nodes: []backend.K3dNode{
				{Name: "node1", Role: "server", State: backend.K3dNodeState{Running: true}},
				{Name: "node2", Role: "agent", State: backend.K3dNodeState{Running: true}},
			},
			ServerCount: 1,
		},
	}
	data, _ := json.Marshal(clusters)
	return data
}

func mockMinikubeListResponse() []byte {
	response := map[string]interface{}{
		"valid": []map[string]interface{}{
			{
				"Name":   "test-cluster",
				"Status": "Running",
				"Nodes": []map[string]interface{}{
					{"Name": "node1"},
					{"Name": "node2"},
				},
			},
		},
	}
	data, _ := json.Marshal(response)
	return data
}
