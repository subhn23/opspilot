# Plan for Track 3.3: Security & Backup Resilience

## Phase 1: Security Scanning [checkpoint: df8965e]
**Goal:** Implement automated image scanning with Trivy.

- [x] Task: Implement `internal/deploy/scanner.go` with `ScanImage` using Trivy binary. (0fe2e62)
- [x] Task: Integrate `ScanImage` into the `Deployer.BuildAndPush` workflow. (0fe2e62)
- [x] Task: Write unit tests for `ScanImage` with mocked Trivy output. (0fe2e62)
- [x] Task: Conductor - User Manual Verification 'Security Scanning' (Protocol in workflow.md). (df8965e)

## Phase 2: Database Resilience [checkpoint: 02e7eaa]
**Goal:** Implement Postgres WAL archiving.

- [x] Task: Implement `internal/config/backup.go` with `ConfigureWALArchiving` logic. (702abaa)
- [x] Task: Add CLI/API support to trigger and verify WAL archiving status. (702abaa)
- [x] Task: Write tests for backup configuration logic. (702abaa)
- [x] Task: Conductor - User Manual Verification 'Database Resilience' (Protocol in workflow.md). (02e7eaa)

## Phase 3: Registry Synchronization [checkpoint: 7ce76ee]
**Goal:** Implement container registry synchronization between nodes.

- [x] Task: Implement `internal/deploy/registry_sync.go` with `SyncNodes` method. (28f0ee9)
- [x] Task: Setup a background worker to trigger `SyncNodes` every 60 seconds. (28f0ee9)
- [x] Task: Write unit tests for registry synchronization logic. (28f0ee9)
- [x] Task: Conductor - User Manual Verification 'Registry Synchronization' (Protocol in workflow.md). (7ce76ee)
