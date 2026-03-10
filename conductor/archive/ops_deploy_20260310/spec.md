# Track 2.2: OpsDeploy Engine Spec

## Goal
Implement a deployment engine capable of building Docker images, scanning them for vulnerabilities using Trivy, and deploying them to Proxmox VMs via SSH.

## Components

### 1. Image Builder & Scanner
- **`BuildAndPush`:**
  - Triggers local Docker builds.
  - Pushes images to a local/mirrored registry.
- **`ScanImage` (Trivy Integration):**
  - Executes Trivy binary to scan for CRITICAL and HIGH vulnerabilities.
  - Rejects deployment if vulnerabilities are found.

### 2. Remote Deployment (`RemoteUp`)
- **SSH Command Execution:**
  - Uses `golang.org/x/crypto/ssh` to connect to the target VM.
  - Performs the following steps:
    1.  `docker login` (if needed).
    2.  `docker pull <new_image>`.
    3.  Updates or generates `docker-compose.yml`.
    4.  `docker-compose up -d`.
- **Status Tracking:** Updates the `Deployment` model status (`BUILDING`, `SCANNING`, `DEPLOYING`, `SUCCESS`, `FAILED`).

### 3. Commit Integration
- **`ListCommits`:** (Future) Fetch recent commits from a Git provider to allow users to select what to deploy.

## Success Criteria
- [ ] `ScanImage` correctly identifies (mocked or real) vulnerabilities.
- [ ] `BuildAndPush` correctly transitions deployment status.
- [ ] `RemoteUp` successfully executes a sequence of commands via a mocked SSH client.
- [ ] Deployment logs are captured and persisted to the database.
- [ ] Code has >80% unit test coverage using interfaces for external dependencies (Docker, Trivy, SSH).
