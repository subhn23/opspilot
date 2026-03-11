package metrics

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func TestMetricStreamer_StreamContainerStats(t *testing.T) {
	gin.SetMode(gin.TestMode)
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

	if _, ok := msg["cpu"]; !ok {
		t.Error("Expected cpu in message")
	}
}
