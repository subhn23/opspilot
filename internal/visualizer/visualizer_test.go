package visualizer

import (
	"opspilot/internal/models"
	"testing"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Environment{}, &models.Deployment{})
	return db
}

func TestBuildTopology(t *testing.T) {
	db := setupTestDB()
	v := NewOpsVisualizer(db)

	t.Run("Empty DB", func(t *testing.T) {
		nodes, edges := v.BuildTopology()
		// Static nodes: fw-01, vip-01
		if len(nodes) != 2 {
			t.Errorf("Expected 2 nodes, got %d", len(nodes))
		}
		if len(edges) != 1 {
			t.Errorf("Expected 1 edge, got %d", len(edges))
		}
	})

	t.Run("With Environments and Deployments", func(t *testing.T) {
		env := models.Environment{
			Name:      "prod-api",
			Status:    "HEALTHY",
			IPAddress: "10.0.0.10",
		}
		if err := db.Create(&env).Error; err != nil {
			t.Fatalf("Failed to create env: %v", err)
		}

		deploy := models.Deployment{
			EnvironmentID: env.ID,
			CommitHash:    "deadbeef123",
			Branch:        "main",
			Status:        "SUCCESS",
			DeployedAt:    time.Now(),
		}
		if err := db.Create(&deploy).Error; err != nil {
			t.Fatalf("Failed to create deploy: %v", err)
		}

		nodes, edges := v.BuildTopology()
		// Nodes: 2 static + 1 vm + 1 container = 4
		if len(nodes) != 4 {
			t.Errorf("Expected 4 nodes, got %d", len(nodes))
		}
		// Edges: 1 static + 1 proxy + 1 docker = 3
		if len(edges) != 3 {
			t.Errorf("Expected 3 edges, got %d", len(edges))
		}

		// Verify container node info
		foundContainer := false
		for _, n := range nodes {
			if n.Type == "container" {
				foundContainer = true
				if n.Metadata["hash"] != "deadbeef123" {
					t.Errorf("Expected hash deadbeef123, got %s", n.Metadata["hash"])
				}
			}
		}
		if !foundContainer {
			t.Error("Container node not found in topology")
		}
	})
}
