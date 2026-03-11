# Track 3.2: OpsMetric Spec

## Goal
Implement a real-time metrics collection and visualization system using VictoriaMetrics for time-series storage and Docker stats for live tracking.

## Components

### 1. Metric Collector (Backend)
- **VictoriaMetrics Integration:** Send scraped metrics (CPU, Memory, Network) to a local VictoriaMetrics instance using the InfluxDB line protocol or Prometheus remote write.
- **Docker Stats Scraping:** Implement `(m *MetricCollector) Scrape()` to fetch real-time usage data from the Docker daemon for all active containers.
- **Periodic Scrape Loop:** A background worker that runs every 5-10 seconds to collect and store metrics.

### 2. Live Stats Streaming (WebSocket)
- **WebSocket Handler:** Implement `StreamContainerStats(containerID string, conn *websocket.Conn)` using Gin and `gorilla/websocket`.
- **Real-time Push:** Directly stream per-second per-container performance monitoring to the browser.

### 3. UI Dashboard (Frontend)
- **Live Graphs:** (Conceptual) Use a lightweight charting library or simple HTML/CSS bars (HTMX-driven) to show live CPU/Mem usage.
- **Historical View:** Query VictoriaMetrics for historical trends.

## Success Criteria
- [ ] `MetricCollector` correctly scrapes data from Docker and stores it in VictoriaMetrics.
- [ ] WebSockets push live container stats to the UI without page refresh.
- [ ] UI displays real-time and historical graphs for selected containers.
- [ ] Code has >80% unit test coverage for collection logic.
