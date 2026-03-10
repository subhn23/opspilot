package visualizer

import (
	"opspilot/internal/models"

	"gorm.io/gorm"
)

type Node struct {
	ID    string            `json:"id"`
	Label string            `json:"label"`
	Type  string            `json:"type"`  // Firewall, VM, Container
	State string            `json:"state"` // Healthy, Down, Provisioning
	Meta  map[string]string `json:"meta"`
}

type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"` // HTTP, gRPC, DB
}

type OpsVisualizer struct {
	DB *gorm.DB
}

func NewOpsVisualizer(db *gorm.DB) *OpsVisualizer {
	return &OpsVisualizer{DB: db}
}

// BuildTopology scans the DB and returns the Graph structure
func (v *OpsVisualizer) BuildTopology() ([]Node, []Edge) {
	var nodes []Node
	var edges []Edge

	// 1. Add static Control Plane Nodes
	nodes = append(nodes, Node{ID: "fw-01", Label: "Physical Firewall", Type: "firewall", State: "healthy"})
	nodes = append(nodes, Node{ID: "vip-01", Label: "HAProxy VIP", Type: "network", State: "healthy"})
	edges = append(edges, Edge{Source: "fw-01", Target: "vip-01", Label: "HTTPS"})

	// 2. Add Dynamic Environments (VMs)
	var envs []models.Environment
	v.DB.Find(&envs)
	for _, env := range envs {
		nodes = append(nodes, Node{
			ID:    env.ID.String(),
			Label: env.Name,
			Type:  "vm",
			State: env.Status,
			Meta:  map[string]string{"ip": env.IPAddress},
		})
		edges = append(edges, Edge{Source: "vip-01", Target: env.ID.String(), Label: "Proxy"})
	}

	// 3. Add individual Services (Containers)
	// Conceptual: Loop through active deployments per environment

	return nodes, edges
}
