# Implementation Plan: Final Production Consolidation

## Phase 1: Network & Observability Depth [checkpoint: 66b4aef]
**Goal:** Connect monitoring sinks and external DNS providers.

- [x] Task: Implement the SSH-based Windows DNS client in `internal/dns/dns.go`. (711c5ce)
- [x] Task: Update `internal/metrics/metrics.go` to use the real Docker API for stats. (390c0da)
- [x] Task: Update `PushToVictoriaMetrics` in `internal/metrics/collector.go`. (b8abe7e)
- [x] Task: Update the Topology Map to include active Docker containers within each VM. (b612442)

## Phase 2: Production Resilience & UI Finalization
**Goal:** Finalize backup, sync, and real-time monitoring.

- [x] Task: Configure `archive_command` for Postgres WAL archiving on OpsControl VMs. (e2f4ca9)
- [x] Task: Finalize the background Registry Autosync worker between Host 1 and Host 2. (50031ad)
- [ ] Task: Build the Live Logs UI component for real-time streaming in the browser.
- [ ] Task: Verify the minimal firewall configuration for port 80/443 routing.
