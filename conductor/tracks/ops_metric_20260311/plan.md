# Plan for Track 3.2: OpsMetric

## Phase 1: Metric Collection Engine [checkpoint: ]
**Goal:** Implement the logic to scrape Docker stats and push to VictoriaMetrics.

- [x] Task: Implement `internal/metrics/collector.go` with `MetricCollector` struct and `Scrape` method. (300ae5d)
- [ ] Task: Integrate VictoriaMetrics client (e.g., using InfluxDB line protocol over HTTP).
- [ ] Task: Write unit tests for `MetricCollector` with mocked Docker daemon and VictoriaMetrics server.
- [ ] Task: Conductor - User Manual Verification 'Collection Engine' (Protocol in workflow.md).

## Phase 2: Live Stats Streaming [checkpoint: ]
**Goal:** Stream live metrics to the frontend via WebSockets.

- [ ] Task: Implement WebSocket handler `StreamContainerStats` in `internal/metrics/metrics.go`.
- [ ] Task: Integrate WebSocket handler with Gin router in `main.go`.
- [ ] Task: Write tests for WebSocket streaming logic.
- [ ] Task: Conductor - User Manual Verification 'Live Stats Streaming' (Protocol in workflow.md).

## Phase 3: UI Integration [checkpoint: ]
**Goal:** Display real-time and historical graphs on the dashboard.

- [ ] Task: Create UI components for live container stats (HTMX + Alpine.js or simple JS).
- [ ] Task: Implement a historical metrics query API endpoint.
- [ ] Task: Final end-to-end verification of the full metrics pipeline.
- [ ] Task: Conductor - User Manual Verification 'UI Integration' (Protocol in workflow.md).
