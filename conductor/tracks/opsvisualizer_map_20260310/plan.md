# Plan for Track 3.1: OpsVisualizer Map

## Phase 1: Topology Engine [checkpoint: 30b327a]
**Goal:** Implement the logic to build the graph data from the database.

- [x] Task: Define Node and Edge models in `internal/models/models.go` (if not already present) (bca49f4)
- [x] Task: Implement `BuildTopology` logic in `internal/visualizer/visualizer.go` (7699322)
- [x] Task: Write tests for `BuildTopology` with various DB states (58d633c)
- [x] Task: Conductor - User Manual Verification 'Topology Engine' (Protocol in workflow.md)

## Phase 2: WebSocket Streaming [checkpoint: 2fc6169]
**Goal:** Stream topology updates to the frontend.

- [x] Task: Implement WebSocket handler in `internal/visualizer/visualizer.go` (1cbc952)
- [x] Task: Integrate WebSocket handler with Gin router (45f7251)
- [x] Task: Write tests for WebSocket connection management (21ddc9a)
- [x] Task: Conductor - User Manual Verification 'WebSocket Streaming' (Protocol in workflow.md)

## Phase 3: Live Integration [checkpoint: ae59699]
**Goal:** Ensure the system pushes live updates on database changes.

- [x] Task: Implement a simple pub/sub or hook system to trigger topology refreshes (39208c0)
- [x] Task: Verify end-to-end flow from DB mutation to WebSocket message (4fde7ee)
- [x] Task: Conductor - User Manual Verification 'Live Integration' (Protocol in workflow.md) (8e41f00)
