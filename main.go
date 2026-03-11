package main

import (
	"log"
	"net/http"
	"opspilot/internal/config"
	"opspilot/internal/metrics"
	"opspilot/internal/models"
	"opspilot/internal/visualizer"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/moby/moby/client"
)

func main() {
	// 1. Initialize Database
	db := config.InitDB(nil)

	// 2. Initialize Docker Client
	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Printf("Warning: Failed to initialize Docker client: %v", err)
	}

	// 3. Initialize Metrics
	collector := &metrics.MetricCollector{
		Docker:             dockerCli,
		VictoriaMetricsURL: os.Getenv("VICTORIAMETRICS_URL"),
	}
	streamer := &metrics.MetricStreamer{
		Collector: collector,
	}

	// 4. Initialize Visualizer
	viz := visualizer.NewOpsVisualizer(db)

	// 5. Initialize Router
	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("ui/templates/*")
	r.Static("/static", "./ui/static")

	// 6. Routes
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

	// Live Metrics WebSocket
	r.GET("/ws/metrics/:id", streamer.StreamContainerStats)

	// Historical Metrics API
	r.GET("/api/metrics/history/:id", func(c *gin.Context) {
		id := c.Param("id")
		metric := c.Query("metric") // cpu_usage, memory_usage
		if metric == "" {
			metric = "cpu_usage"
		}

		query := fmt.Sprintf("docker_metrics_%s{container_id=\"%s\"}", metric, id)
		// Last 1 hour
		end := time.Now()
		start := end.Add(-1 * time.Hour)

		data, err := collector.QueryRange(c.Request.Context(), query, start, end, "1m")
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.Data(http.StatusOK, "application/json", []byte(data))
	})

	// 7. Start Server
	port := "8080"
	log.Printf("OpsPilot Control Plane starting on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
