package metrics

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type MetricStreamer struct{}

// StreamStats opens a websocket to stream live 'docker stats' data
func (s *MetricStreamer) StreamStats(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	// Conceptual Loop: Scrape docker stats and send to UI
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// Mock data for now
		stats := map[string]interface{}{
			"cpu":    "12.5%",
			"memory": "256MB / 2GB",
			"time":   time.Now().Format(time.Kitchen),
		}

		if err := conn.WriteJSON(stats); err != nil {
			return
		}
	}
}

// PushToVictoriaMetrics (Conceptual)
func (s *MetricStreamer) PushToVictoriaMetrics() {
	// Send metrics to http://victoriametrics:8428/api/v1/import
}
