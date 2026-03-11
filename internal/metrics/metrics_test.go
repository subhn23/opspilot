package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/moby/moby/api/types/container"
)

func TestMetricStreamer_StreamContainerStats(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("MockDataFallback", func(t *testing.T) {
		streamer := &MetricStreamer{}
		r := gin.New()
		r.GET("/ws/metrics/:id", streamer.StreamContainerStats)
		server := httptest.NewServer(r)
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/metrics/test-id"
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

		if msg["container_id"] != "test-id" {
			t.Errorf("Expected container_id test-id, got %v", msg["container_id"])
		}
		if msg["cpu"] != 12.5 {
			t.Errorf("Expected cpu 12.5, got %v", msg["cpu"])
		}
	})

	t.Run("WithRealCollector", func(t *testing.T) {
		mockDocker := &MockDockerClient{
			Containers: []container.Summary{
				{ID: "real-id", Names: []string{"/real-container"}},
			},
			Stats: container.StatsResponse{
				CPUStats: container.CPUStats{
					CPUUsage:    container.CPUUsage{TotalUsage: 200},
					SystemUsage: 2000,
					OnlineCPUs:  1,
				},
				PreCPUStats: container.CPUStats{
					CPUUsage:    container.CPUUsage{TotalUsage: 100},
					SystemUsage: 1000,
				},
				MemoryStats: container.MemoryStats{Usage: 1024},
			},
		}

		collector := &MetricCollector{Docker: mockDocker}
		streamer := &MetricStreamer{Collector: collector}

		r := gin.New()
		r.GET("/ws/metrics/:id", streamer.StreamContainerStats)
		server := httptest.NewServer(r)
		defer server.Close()

		wsURL := "ws" + strings.TrimPrefix(server.URL, "http") + "/ws/metrics/real-id"
		dialer := websocket.Dialer{}
		conn, _, err := dialer.Dial(wsURL, nil)
		if err != nil {
			t.Fatalf("Failed to connect to WebSocket: %v", err)
		}
		defer conn.Close()

		// Read first message
		var msg map[string]interface{}
		err = conn.ReadJSON(&msg)
		if err != nil {
			t.Fatalf("Failed to read JSON: %v", err)
		}

		if msg["container_id"] != "real-id" {
			t.Errorf("Expected container_id real-id, got %v", msg["container_id"])
		}
		if msg["cpu"] != 10.0 {
			t.Errorf("Expected cpu 10.0, got %v", msg["cpu"])
		}
		if msg["memory"] != 1024.0 { // JSON numbers are float64
			t.Errorf("Expected memory 1024, got %v", msg["memory"])
		}
	})
}

// Redefine MockDockerClient for metrics_test.go if needed, 
// but since it's in the same package 'metrics', it should be available 
// if defined in collector_test.go. 
// Wait, collector_test.go is only included during 'go test'.
// I'll check if they are in the same package. Yes they are.
