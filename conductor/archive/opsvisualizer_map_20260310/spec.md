# Track 3.1: OpsVisualizer Map Spec

## Goal
Implement an interactive visual map of the entire infrastructure stack, showing physical servers, virtual machines, containers, and their networking relationships in real-time.

## Components

### 1. Topology Builder
- **Graph Generation:** Logic to query the database (Environments, Deployments, ProxyRoutes) and build a graph of Nodes and Edges.
- **Node Types:** `Firewall`, `VM`, `Container`.
- **Status Mapping:** Map internal resource status to visual indicators (`Green`, `Red`, `Yellow`).

### 2. Real-time Streaming
- **WebSocket Integration:** Use Gin and `gorilla/websocket` to push topology changes to the frontend.
- **Incremental Updates:** Efficiently stream only changed nodes/edges to minimize bandwidth.

### 3. UI Visualization
- **n8n-style Canvas:** (Conceptual) A node-based UI where users can see the relationship between services.
- **Interactive Details:** Clicking a node shows detailed metadata (IP, Status, Logs).

## Success Criteria
- [ ] `BuildTopology` correctly constructs a graph representing the current DB state.
- [ ] WebSocket server correctly handles connections and heartbeats.
- [ ] UI receives and renders the topology graph.
- [ ] Changes in the database are automatically pushed to connected clients.
- [ ] Code has >80% unit test coverage for graph logic.
