# Phase 3: Governance & Operations (Visibility & Safety)

This phase adds the final layer of n8n-style visualization, real-time metrics, and security.

## Track 1: OpsVisualizer (Topology Map)
**Goal:** Interactive visual map of the entire stack.

### Data Structures
```go
type Node struct {
    ID       string `json:"id"`
    Label    string `json:"label"`
    Type     string `json:"type"`   // Firewall, VM, Container
    Status   string `json:"status"` // Green, Red, Yellow
    Metadata map[string]string `json:"metadata"`
}

type Edge struct {
    Source string `json:"source"`
    Target string `json:"target"`
    Label  string `json:"label"` // HTTP, gRPC, DB
}
```

### Key Functions
- `BuildTopology() ([]Node, []Edge)`: Generates graph data from DB/Terraform.
- `StreamTopologyUpdates(conn *websocket.Conn)`: Pushes live map changes to UI.

---

## Track 2: OpsMetric (Time-Series & Health)
**Goal:** VictoriaMetrics integration and live docker stats.

### Key Functions
- `(m *MetricCollector) Scrape()`: Fetches `docker stats` and pushes to VictoriaMetrics.
- `StreamContainerStats(containerID string, conn *websocket.Conn)`: Real-time per-second performance monitoring in browser.

---

## Track 3: Security & Backup (Trivy & WAL)
**Goal:** CVE Scanning and Point-In-Time Recovery.

### Key Functions
- `ScanImage(image string) (*TrivyReport, error)`: Runs Trivy binary.
- `ConfigureWALArchiving() error`: Sets up Postgres to stream logs to external storage.
- `(r *RegistrySync) SyncNodes()`: Ensures Host 1 and Host 2 registries are identical every 60s.
