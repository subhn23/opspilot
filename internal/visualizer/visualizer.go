package visualizer

import (
	"fmt"
	"log"
	"net/http"
	"opspilot/internal/events"
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

type Hub struct {
	clients    map[chan bool]bool
	broadcast  chan bool
	register   chan chan bool
	unregister chan chan bool
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[chan bool]bool),
		broadcast:  make(chan bool),
		register:   make(chan chan bool),
		unregister: make(chan chan bool),
	}
}

func (h *Hub) Run() {
	// Subscribe to global event bus
	eventCh := events.GlobalBus.Subscribe()
	defer events.GlobalBus.Unsubscribe(eventCh)

	for {
		select {
		case <-eventCh:
			// Trigger a broadcast when global event occurs
			go func() { h.broadcast <- true }()
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client)
			}
		case <-h.broadcast:
			for client := range h.clients {
				select {
				case client <- true:
				default:
					close(client)
					delete(h.clients, client)
				}
			}
		}
	}
}

type OpsVisualizer struct {
	DB  *gorm.DB
	Hub *Hub
}

func NewOpsVisualizer(db *gorm.DB) *OpsVisualizer {
	h := NewHub()
	go h.Run()
	return &OpsVisualizer{
		DB:  db,
		Hub: h,
	}
}

// Notify triggers a refresh for all connected clients
func (v *OpsVisualizer) Notify() {
	v.Hub.broadcast <- true
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

	// Register client
	updateChan := make(chan bool, 1)
	v.Hub.register <- updateChan
	defer func() {
		v.Hub.unregister <- updateChan
	}()

	// Initial Push
	nodes, edges := v.BuildTopology()
	if err := conn.WriteJSON(gin.H{"nodes": nodes, "edges": edges}); err != nil {
		return
	}

	// Ticker for heartbeat/keepalive
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-updateChan:
			log.Println("Visualizer: Pushing live update")
			nodes, edges := v.BuildTopology()
			if err := conn.WriteJSON(gin.H{"nodes": nodes, "edges": edges}); err != nil {
				return
			}
		case <-ticker.C:
			// Ping to keep connection alive
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-c.Request.Context().Done():
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
		var deployments []models.Deployment
		v.DB.Where("environment_id = ? AND status = ?", env.ID, "SUCCESS").Find(&deployments)
		for _, deploy := range deployments {
		        nodeID := fmt.Sprintf("svc-%d", deploy.ID)
		        nodes = append(nodes, models.Node{
		                ID:       nodeID,
		                Label:    fmt.Sprintf("App (%s)", deploy.CommitHash[:7]),
		                Type:     "container",
		                Status:   "Green",
		                Metadata: map[string]string{"hash": deploy.CommitHash, "branch": deploy.Branch, "container_id": deploy.ContainerID},
		        })
		        edges = append(edges, models.Edge{Source: env.ID.String(), Target: nodeID, Label: "Docker"})
		}

	}

	return nodes, edges
}
