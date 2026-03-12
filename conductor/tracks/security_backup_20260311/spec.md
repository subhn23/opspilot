# Track 3.3: Security & Backup Resilience Spec

## Goal
Implement automated security scanning for Docker images and robust backup/synchronization mechanisms for high-availability database and container registry layers.

## Components

### 1. Security Scanner (Trivy Integration)
- **Image Scanning:** Implement `ScanImage(image string) (*TrivyReport, error)` to run the Trivy binary against built images.
- **Reporting:** Capture and store scan results, blocking deployments if CRITICAL or HIGH vulnerabilities are detected.
- **UI Feedback:** Display vulnerability status on the deployment and environment dashboards.

### 2. Database Resilience (Postgres WAL Archiving)
- **WAL-E/WAL-G Integration:** Configure PostgreSQL to stream Write-Ahead Logs (WAL) to an external storage layer (e.g., MinIO or a local network share).
- **Point-In-Time Recovery (PITR):** Enable the ability to restore the database to any specific second in the past.
- **Configuration:** Implement `ConfigureWALArchiving() error` to automate the setup via the control plane.

### 3. Registry Synchronization
- **Multi-Node Sync:** Implement `(r *RegistrySync) SyncNodes()` to ensure Host 1 and Host 2 container registries are identical.
- **Periodic Sync:** A background worker that triggers synchronization every 60 seconds to maintain parity.

## Success Criteria
- [ ] Docker images are automatically scanned before deployment.
- [ ] Deployments are blocked by high-severity CVEs.
- [ ] Postgres WAL logs are successfully archived to the backup target.
- [ ] Container registries on both HA nodes are synchronized periodically.
- [ ] Code has >80% unit test coverage for scanning and sync logic.
