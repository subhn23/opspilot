package visualizer

import (
	"net/http/httptest"
	"opspilot/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
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

	t.Run("With Various Statuses", func(t *testing.T) {
		db := setupTestDB()
		v := NewOpsVisualizer(db)

		envs := []models.Environment{
			{Name: "healthy-env", Status: "HEALTHY", IPAddress: "10.0.0.1", VMID: 101},
			{Name: "failed-env", Status: "FAILED", IPAddress: "10.0.0.2", VMID: 102},
			{Name: "destroyed-env", Status: "DESTROYED", IPAddress: "10.0.0.3", VMID: 103},
			{Name: "pending-env", Status: "PROVISIONING", IPAddress: "10.0.0.4", VMID: 104},
		}

		for _, env := range envs {
			db.Create(&env)
		}

		nodes, _ := v.BuildTopology()
		// 2 static + 4 envs = 6
		if len(nodes) != 6 {
			t.Errorf("Expected 6 nodes, got %d", len(nodes))
		}

		statusMap := make(map[string]string)
		for _, n := range nodes {
			if n.Type == "vm" {
				statusMap[n.Label] = n.Status
			}
		}

		if statusMap["healthy-env"] != "Green" {
			t.Errorf("Expected healthy-env to be Green, got %s", statusMap["healthy-env"])
		}
		if statusMap["failed-env"] != "Red" {
			t.Errorf("Expected failed-env to be Red, got %s", statusMap["failed-env"])
		}
		if statusMap["destroyed-env"] != "Red" {
			t.Errorf("Expected destroyed-env to be Red, got %s", statusMap["destroyed-env"])
		}
		if statusMap["pending-env"] != "Yellow" {
			t.Errorf("Expected pending-env to be Yellow, got %s", statusMap["pending-env"])
		}
	})

	t.Run("With Deployments", func(t *testing.T) {
		db := setupTestDB()
		v := NewOpsVisualizer(db)

		env := models.Environment{
			Name:      "prod-api",
			Status:    "HEALTHY",
			IPAddress: "10.0.0.10",
		}
		db.Create(&env)

		deploy := models.Deployment{
			EnvironmentID: env.ID,
			CommitHash:    "deadbeef123",
			Branch:        "main",
			Status:        "SUCCESS",
			DeployedAt:    time.Now(),
		}
		db.Create(&deploy)

		nodes, edges := v.BuildTopology()
		// 2 static + 1 vm + 1 container = 4
		if len(nodes) != 4 {
			t.Errorf("Expected 4 nodes, got %d", len(nodes))
		}
		// 1 static + 1 proxy + 1 docker = 3
		if len(edges) != 3 {
			t.Errorf("Expected 3 edges, got %d", len(edges))
		}
	})
}

func TestStreamTopologyUpdates(t *testing.T) {
	gin.SetMode(gin.TestMode)

	setup := func() (*gorm.DB, *OpsVisualizer, *httptest.Server, string) {
		db := setupTestDB()
		v := NewOpsVisualizer(db)
		r := gin.New()
		r.GET("/ws", v.StreamTopologyUpdates)
		server := httptest.NewServer(r)
		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws"
		return db, v, server, wsURL
	}

	t.Run("Connect and Receive Initial Data", func(t *testing.T) {
		_, _, server, wsURL := setup()
		defer server.Close()

		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		var msg map[string]interface{}
		err = conn.ReadJSON(&msg)
		if err != nil {
			t.Fatalf("Failed to read JSON: %v", err)
		}

		if _, ok := msg["nodes"]; !ok {
			t.Error("Expected nodes in message")
		}
	})

	t.Run("Receive Live Update on Notify", func(t *testing.T) {
		db, v, server, wsURL := setup()
		defer server.Close()

		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		// Read initial data
		var initialMsg map[string]interface{}
		_ = conn.ReadJSON(&initialMsg)

		// Create a new environment to trigger update
		env := models.Environment{Name: "new-env", Status: "HEALTHY", VMID: 200}
		db.Create(&env)
		v.Notify()

		// Read second message
		var updateMsg map[string]interface{}
		err = conn.ReadJSON(&updateMsg)
		if err != nil {
			t.Fatalf("Failed to read JSON update: %v", err)
		}

		nodes := updateMsg["nodes"].([]interface{})
		// 2 static + 1 new env = 3
		if len(nodes) != 3 {
			t.Errorf("Expected 3 nodes after update, got %d", len(nodes))
		}
	})

	t.Run("Handle Client Disconnection", func(t *testing.T) {
		_, v, server, wsURL := setup()
		defer server.Close()

		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}

		// Wait for registration
		time.Sleep(50 * time.Millisecond)
		if len(v.Hub.clients) != 1 {
			t.Errorf("Expected 1 registered client, got %d", len(v.Hub.clients))
		}

		// Close connection
		conn.Close()
		
		// We need to trigger a write failure or wait for context cancellation
		// In Gin TestMode, context cancellation might not happen immediately
		// But let's try to Notify to trigger a write on a closed connection
		v.Notify()
		
		time.Sleep(100 * time.Millisecond)
		// Client should be removed after failed write or unregister
		if len(v.Hub.clients) != 0 {
			t.Errorf("Expected 0 registered clients, got %d", len(v.Hub.clients))
		}
	})
}
