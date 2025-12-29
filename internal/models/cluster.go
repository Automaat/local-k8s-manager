package models

import "time"

// Cluster represents a local Kubernetes cluster
type Cluster struct {
	Name      string
	Provider  string
	Status    Status
	Nodes     int
	CreatedAt time.Time
}

// Status represents the current state of a cluster
type Status string

const (
	StatusRunning Status = "running"
	StatusStopped Status = "stopped"
	StatusUnknown Status = "unknown"
	StatusError   Status = "error"
)
