# OpsPilot Implementation TODO List

This document tracks the "Last-Mile" tasks required to move from the current architectural scaffold to a fully production-integrated platform.

## 1. Authentication & Security
- [ ] **JWT/Session Implementation:** Finish `AuthMiddleware` in `internal/auth/auth.go` to validate user sessions.
- [ ] **MFA Enrollment UI:** Create a page to show the TOTP QR code for new users.
- [ ] **Trivy Integration:** Ensure the `trivy` binary is installed on the host and finish the `ScanImage` call in `internal/deploy/deploy.go`.

## 2. Infrastructure Orchestration (Terraform)
- [ ] **Template Mirroring:** In `internal/terraform/terraform.go`, implement the logic to copy `terraform/base/*.tf` files into the newly created environment workspaces.
- [ ] **Proxmox Credentials:** Securely pass `PM_API_TOKEN_ID` and `PM_API_TOKEN_SECRET` from environment variables to the `TFEngine`.

## 3. Delivery Pipeline (OpsDeploy)
- [ ] **Git Integration:** Implement the `git clone` and `git checkout` logic within the `BuildAndPush` function.
- [ ] **Registry Credentials:** Add `docker login` capability or use an AWS/GCP helper for pushing to the mirrored registry.
- [ ] **SSH Deployment:** Finish the `RemoteUp` function in `internal/deploy/deploy.go` using `golang.org/x/crypto/ssh` to trigger `docker-compose up` on dynamic VMs.

## 4. Network & DNS
- [ ] **Windows DNS SSH:** Implement the SSH client in `internal/dns/dns.go` to execute PowerShell commands on the Windows DNS server.
- [ ] **Minimal Firewall:** Verify that your physical firewall only routes ports 80/443 to the HAProxy VIP.

## 5. Observability (Metrics & Visualizer)
- [ ] **Real Docker Stats:** Update `internal/metrics/metrics.go` to use the Docker socket/API to fetch real per-second stats instead of mock data.
- [ ] **VictoriaMetrics Sink:** Implement the `PushToVictoriaMetrics` function to persist historical data for the 1-week requirement.
- [ ] **Topology Granularity:** Update `internal/visualizer/visualizer.go` to loop through active Docker containers within each VM and add them as nodes to the map.

## 6. Resilience
- [ ] **Postgres WAL Archiving:** Install `pgBackRest` or `WAL-G` on the OpsControl VMs and configure the `archive_command` in `postgresql.conf`.
- [ ] **Registry Autosync:** Finalize the background worker that triggers `docker pull/push` between Host 1 and Host 2 every 60 seconds.

## 7. UI/UX (Tailwind & HTMX)
- [ ] **Environment Wizard:** Build the HTMX form to allow Master Admins to input VM specs and select branches for new environments.
- [ ] **Audit Viewer:** Create a page to display the `system_audit_logs` table with filtering by user and action.
- [ ] **Live Logs UI:** Build the frontend component to consume the Websocket stream for real-time container logs.
