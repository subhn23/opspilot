# Implementation Plan: Final Production Consolidation

## Phase 1: Network & Observability Depth
**Goal:** Connect monitoring sinks and external DNS providers.

- [x] Task: Implement the SSH-based Windows DNS client in `internal/dns/dns.go`. (711c5ce)
- [x] Task: Update `internal/metrics/metrics.go` to use the real Docker API for stats. (390c0da)
- [x] Task: Implement `PushToVictoriaMetrics` in `internal/metrics/collector.go`. (b8abe7e)
- [ ] Task: Update the Topology Map to include active Docker containers within each VM.

## Phase 2: Production Resilience & UI Finalization
**Goal:** Finalize backup, sync, and real-time monitoring.

- [ ] Task: Configure `archive_command` for Postgres WAL archiving on OpsControl VMs.
- [ ] Task: Finalize the background Registry Autosync worker between Host 1 and Host 2.
- [ ] Task: Build the Live Logs UI component for real-time streaming in the browser.
- [ ] Task: Verify the minimal firewall configuration for port 80/443 routing.
