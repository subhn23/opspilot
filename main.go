package main

import (
	"log"
	"net/http"
	"opspilot/internal/config"
	"opspilot/internal/models"
	"opspilot/internal/visualizer"

	"github.com/gin-gonic/gin"
)

func main() {
	// 1. Initialize Database
	db := config.InitDB(nil)

	// 2. Initialize Visualizer
	viz := visualizer.NewOpsVisualizer(db)

	// 3. Initialize Router
	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("ui/templates/*")
	r.Static("/static", "./ui/static")

	// 4. Routes
	r.GET("/", func(c *gin.Context) {
		var envCount int64
		db.Model(&models.Environment{}).Count(&envCount)

		c.HTML(http.StatusOK, "index.html", gin.H{
			"Title":     "Dashboard",
			"EnvsCount": envCount,
		})
	})

	// Topology API (HTMX support)
	r.GET("/api/topology", func(c *gin.Context) {
		nodes, edges := viz.BuildTopology()
		c.JSON(http.StatusOK, gin.H{
			"nodes": nodes,
			"edges": edges,
		})
	})

	// Topology WebSocket (Real-time updates)
	r.GET("/ws/topology", viz.StreamTopologyUpdates)

	// 5. Start Server
	port := "8080"
	log.Printf("OpsPilot Control Plane starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
