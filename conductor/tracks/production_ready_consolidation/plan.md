# Implementation Plan: Production-Ready Consolidation

## Phase 1: Authentication & UI Foundation
**Goal:** Solidify user security and provide essential monitoring UI.

- [x] Task: Complete `AuthMiddleware` in `internal/auth/auth.go` to validate JWT sessions. (ab291a1)
- [x] Task: Create `ui/templates/mfa_enroll.html` for TOTP QR code display. (0564ce8)
- [x] Task: Build the Audit Viewer page to display `system_audit_logs`. (2255308)
- [x] Task: Implement the HTMX-based Environment Wizard for VM provisioning. (24460d7)

## Phase 2: Infrastructure & Deployment Logic
**Goal:** Finalize the Terraform and Docker delivery pipelines.

- [ ] Task: Implement Terraform template mirroring logic in `internal/terraform/terraform.go`.
- [ ] Task: Securely pass Proxmox credentials to `TFEngine`.
- [ ] Task: Implement `git clone` and `git checkout` in `BuildAndPush`.
- [ ] Task: Add `docker login` and push logic for the mirrored registry.
- [ ] Task: Finish `RemoteUp` using `golang.org/x/crypto/ssh` for VM deployment.

## Phase 3: Observability & Network Integration
**Goal:** Connect monitoring sinks and external DNS providers.

- [ ] Task: Implement the SSH-based Windows DNS client in `internal/dns/dns.go`.
- [ ] Task: Update `internal/metrics/metrics.go` to use the real Docker API for stats.
- [ ] Task: Implement `PushToVictoriaMetrics` in `internal/metrics/collector.go`.
- [ ] Task: Update the Topology Map to include active Docker containers.
- [ ] Task: Build the Live Logs UI component for real-time streaming.

## Phase 4: Production Resilience & Hardening
**Goal:** Finalize backup, sync, and security scanning.

- [ ] Task: Finalize the `ScanImage` call in `internal/deploy/deploy.go` with Trivy binary.
- [ ] Task: Configure `archive_command` for Postgres WAL archiving on OpsControl VMs.
- [ ] Task: Finalize the background Registry Autosync worker between Host 1 and Host 2.
- [ ] Task: Verify the minimal firewall configuration for port 80/443 routing.
