package metrics

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type MetricStreamer struct {
	Collector *MetricCollector
}

// StreamContainerStats opens a websocket to stream live 'docker stats' data for a specific container
func (s *MetricStreamer) StreamContainerStats(c *gin.Context) {
	containerID := c.Param("id")
	if containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "container id is required"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	if s.Collector == nil || s.Collector.Docker == nil {
		log.Printf("StreamContainerStats: Collector or Docker client is nil")
		return
	}

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	statsChan, errChan := s.Collector.StreamStats(ctx, containerID)

	for {
		select {
		case m, ok := <-statsChan:
			if !ok {
				return
			}
			stats := map[string]interface{}{
				"container_id": m.ContainerID,
				"cpu":          m.CPUUsage,
				"memory":       m.MemoryUsage,
				"time":         m.Timestamp.Format(time.Kitchen),
			}
			if err := conn.WriteJSON(stats); err != nil {
				return
			}
		case err := <-errChan:
			if err != nil {
				log.Printf("StreamContainerStats: error from Docker: %v", err)
				return
			}
		case <-c.Request.Context().Done():
			return
		}
	}
}

