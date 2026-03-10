package visualizer

import (
	"fmt"
	"log"
	"net/http"
	"opspilot/internal/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // For dev, allow all origins
	},
}

type OpsVisualizer struct {
	DB *gorm.DB
}

func NewOpsVisualizer(db *gorm.DB) *OpsVisualizer {
	return &OpsVisualizer{DB: db}
}

// StreamTopologyUpdates handles WebSocket connections and streams graph data
func (v *OpsVisualizer) StreamTopologyUpdates(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Failed: %v", err)
		return
	}
	defer conn.Close()

	log.Printf("Visualizer: Client connected from %s", c.ClientIP())

	// Initial Push
	nodes, edges := v.BuildTopology()
	if err := conn.WriteJSON(gin.H{"nodes": nodes, "edges": edges}); err != nil {
		return
	}

	// Simple Polling Loop for now (Phase 3 will make it event-driven)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			nodes, edges := v.BuildTopology()
			if err := conn.WriteJSON(gin.H{"nodes": nodes, "edges": edges}); err != nil {
				log.Printf("WebSocket Write Error: %v", err)
				return
			}
		case <-c.Request.Context().Done():
			log.Println("Visualizer: Client disconnected")
			return
		}
	}
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
