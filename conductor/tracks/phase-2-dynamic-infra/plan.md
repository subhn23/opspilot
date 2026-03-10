# Phase 2: Dynamic Infrastructure (Terraform & Docker)

This phase enables dynamic VM provisioning on Proxmox and the UI-driven deployment engine.

## Track 1: Terraform Orchestration
**Goal:** Interface with Proxmox via `terraform-exec`.

### Data Structures
```go
type Environment struct {
    ID          uuid.UUID `gorm:"primaryKey"`
    Name        string    `gorm:"uniqueIndex"`
    Type        string    // Prod, Staging, Dev
    HostNode    string    // Proxmox Node 1 or 2
    VM_ID       int
    IP_Address  string
    TTL         time.Time // Expiry for Dev environments
    Status      string    // Provisioning, Healthy, Failed
}
```

### Key Functions
- `(t *TFEngine) Provision(env *Environment) error`: Runs `terraform apply`.
- `(t *TFEngine) Destroy(env *Environment) error`: Runs `terraform destroy`.
- `(t *TFEngine) Migrate(env *Environment, targetNode string) error`: Handles host-to-host move.

---

## Track 2: OpsDeploy Engine
**Goal:** Multi-branch commit browsing and Docker delivery.

### Data Structures
```go
type Deployment struct {
    ID            uint      `gorm:"primaryKey"`
    EnvironmentID uuid.UUID
    CommitHash    string
    Branch        string
    Status        string    // Building, Pushing, Deploying, Success
    Logs          string    `gorm:"type:text"`
    DeployedAt    time.Time
}
```

### Key Functions
- `(d *Deployer) ListCommits(repo string, branch string) ([]Commit, error)`: Fetches from Git API.
- `(d *Deployer) BuildAndPush(deploy *Deployment) error`: Local Docker build.
- `(d *Deployer) RemoteUp(deploy *Deployment) error`: SSH to VM and run `docker-compose up`.

---

## Track 3: Windows DNS Module
**Goal:** Automate record updates.

### Key Functions
- `UpdateDNS(domain string, ip string) error`: Executes PowerShell via SSH.
- `RequestManualDNS(domain string)`: Blocks UI until record is verified.
