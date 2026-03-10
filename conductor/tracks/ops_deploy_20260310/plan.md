# Plan for Track 2.2: OpsDeploy Engine

## Phase 1: Security Scanning (Trivy Integration) [checkpoint: ]
**Goal:** Implement the `ScanImage` logic and verification.

- [x] Task: Define `Scanner` interface for `Trivy` integration (Already defined)
- [x] Task: Implement `ScanImage` using `exec.Command` (61c3608)
- [x] Task: Write tests for `ScanImage` with a mock scanner (61c3608)
- [ ] Task: Conductor - User Manual Verification 'Security Scanning' (Protocol in workflow.md)

## Phase 2: Remote Deployment (SSH) [checkpoint: ]
**Goal:** Implement the `RemoteUp` logic using SSH.

- [ ] Task: Define `SSHClient` interface for command execution
- [ ] Task: Implement `RemoteUp` command sequence (pull, up)
- [ ] Task: Write tests for `RemoteUp` using a mock SSH client
- [ ] Task: Conductor - User Manual Verification 'Remote Deployment' (Protocol in workflow.md)

## Phase 3: Engine Orchestration [checkpoint: ]
**Goal:** Connect all pieces in `BuildAndPush` and `Deploy` flows.

- [ ] Task: Implement `BuildAndPush` transitions and logging
- [ ] Task: Add deployment logging to the `AuditLog`
- [ ] Task: Conductor - User Manual Verification 'Engine Orchestration' (Protocol in workflow.md)
