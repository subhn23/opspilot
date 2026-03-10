package visualizer

import (
	"fmt"
	"opspilot/internal/models"

	"gorm.io/gorm"
)

type OpsVisualizer struct {
	DB *gorm.DB
}

func NewOpsVisualizer(db *gorm.DB) *OpsVisualizer {
	return &OpsVisualizer{DB: db}
}

// BuildTopology scans the DB and returns the Graph structure
func (v *OpsVisualizer) BuildTopology() ([]models.Node, []models.Edge) {
	var nodes []models.Node
	var edges []models.Edge

	// 1. Add static Control Plane Nodes
	nodes = append(nodes, models.Node{ID: "fw-01", Label: "Physical Firewall", Type: "firewall", Status: "Green"})
	nodes = append(nodes, models.Node{ID: "vip-01", Label: "HAProxy VIP", Type: "network", Status: "Green"})
	edges = append(edges, models.Edge{Source: "fw-01", Target: "vip-01", Label: "HTTPS"})

	// 2. Add Dynamic Environments (VMs)
	var envs []models.Environment
	v.DB.Find(&envs)
	for _, env := range envs {
		status := "Yellow"
		if env.Status == "HEALTHY" {
			status = "Green"
		} else if env.Status == "FAILED" || env.Status == "DESTROYED" {
			status = "Red"
		}

		nodes = append(nodes, models.Node{
			ID:       env.ID.String(),
			Label:    env.Name,
			Type:     "vm",
			Status:   status,
			Metadata: map[string]string{"ip": env.IPAddress, "type": env.Type},
		})
		edges = append(edges, models.Edge{Source: "vip-01", Target: env.ID.String(), Label: "Proxy"})

		// 3. Add individual Services (Containers) per environment
		var latestDeploy models.Deployment
		if err := v.DB.Where("environment_id = ? AND status = ?", env.ID, "SUCCESS").Order("deployed_at desc").First(&latestDeploy).Error; err == nil {
			nodeID := fmt.Sprintf("svc-%d", latestDeploy.ID)
			nodes = append(nodes, models.Node{
				ID:       nodeID,
				Label:    fmt.Sprintf("App (%s)", latestDeploy.CommitHash[:7]),
				Type:     "container",
				Status:   "Green",
				Metadata: map[string]string{"hash": latestDeploy.CommitHash, "branch": latestDeploy.Branch},
			})
			edges = append(edges, models.Edge{Source: env.ID.String(), Target: nodeID, Label: "Docker"})
		}
	}

	return nodes, edges
}
