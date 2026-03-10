# Infrastructure & CI/CD Architecture Plan

## Executive Summary

This document outlines a complete infrastructure solution for a 2-server high-availability setup with microservices architecture, designed for a Go development team with ~2,000-10,000 users.

---

## Table of Contents

1. [Problem Statement](#1-problem-statement)
2. [Constraints & Requirements](#2-constraints--requirements)
3. [Architecture Overview](#3-architecture-overview)
4. [Component-Specific HA Strategy](#4-component-specific-ha-strategy)
5. [Secrets Management](#5-secrets-management)
6. [Microservices Migration Strategy](#6-microservices-migration-strategy)
7. [CI/CD Pipeline Design](#7-cicd-pipeline-design)
8. [Custom OpsPilot Tool Proposal](#8-custom-opspilot-tool-proposal)
   - [8.5 OpsProxy Module: Native Reverse Proxy & SSL](#85-opsproxy-module-native-reverse-proxy--ssl)
   - [8.6 OpsDeploy Module: UI-Driven Deployment & Rollbacks](#86-opsdeploy-module-ui-driven-deployment--rollbacks)
   - [8.7 Windows DNS Integration Module](#87-windows-dns-integration-module)
   - [8.8 Managed Management Tool Stack (HA)](#88-managed-management-tool-stack-ha)
   - [8.9 Logging & Observation Module (Websocket Streaming)](#89-logging--observation-module-websocket-streaming)
   - [8.10 OpsVisualizer Module (Topology Map)](#810-opsvisualizer-module-topology-map)
9. [Maintenance Workflow](#9-maintenance-workflow)
10. [Next Steps](#10-next-steps)
11. [Resilience & Governance](#11-resilience--governance)
12. [Advanced Operations & Safety](#12-advanced-operations--safety)

---

## 1. Problem Statement

### Current Situation

- Single server deployment using Docker
- Downtime during hardware maintenance
- Monolithic Go application (Gin framework)
- Uses PostgreSQL, LGTM stack (Redis, RabbitMQ, and MinIO consolidated)

### Goals

- High availability on 2 physical servers
- Zero/smooth downtime for planned maintenance
- Microservices architecture migration
- Multi-environment support (dev/staging/testing/production)
- Secrets management solution

---

## 2. Constraints & Requirements

| Constraint | Implication |
|------------|-------------|
| Only 2 servers available | Cannot use 3-node quorum systems (Kubernetes, etcd cluster) |
| No cloud services | Must be fully self-hosted/on-premise |
| Planned downtime acceptable | Manual/semi-automatic failover is sufficient |
| Multiple environments (4+) | Need environment isolation with easy switching |
| Vault/Infisical planned | Use for secrets management across environments |
| Small data size (~120MB DB, 18K files) | Simplifies backup/restore strategies |
| 1Gbps network (10Gbps planned) | Async replication will work fine |
| Go proficiency | Custom tools can be built in Go |
| Beginner-friendly requirement | Need simple GUI for operations |

---

## 3. Architecture Overview

### Proxmox Virtual Machine Layout

OpsPilot uses **Terraform** to dynamically manage the compute layer. The environment consists of two static "Control Plane" VMs and a pool of dynamic "Application VMs."

| VM Category | VM Name | Role | Lifecycle |
| :--- | :--- | :--- | :--- |
| **Control Plane** | **OpsControl-01/02** | PostgreSQL (HA), OpsPilot, HAProxy/Keepalived. | **Static** |
| **Dynamic Apps** | **AppNode-<env>-<id>** | Isolated VM for a specific environment (Prod, Staging, etc.). | **Dynamic (Terraform)** |

### Server Roles (Logical View)

```
+---------------------------------------------------------------------+
|                      OPSPILOT CONTROL PLANE                         |
+---------------------------------------------------------------------+
|                                                                      |
|   +---------------------+     +---------------------+                |
|   |   OpsControl-01     |     |   OpsControl-02     |                |
|   |   (Active/Primary)  |<--->|   (Standby)        |                |
|   +---------------------+     +---------------------+                |
|   | HAProxy (VIP route) |     | HAProxy (passive)  |                |
|   | Keepalived (VIP)    |     | Keepalived (monitor)|                |
|   +---------------------+     +---------------------+                |
|   | PostgreSQL (Primary)|---->| PostgreSQL (Replica)|               |
|   | OpsPilot (Orchestrator) <---> Terraform + Proxmox API          |
|   +---------------------+     +---------------------+                |
|                                                                      |
+---------------------------------------------------------------------+
          │                                      │
          ▼                                      ▼
+-----------------------+              +-----------------------+
|  Dynamic App VM (Dev) |              | Dynamic App VM (Prod) |
|  [Docker Microservices]|              | [Docker Microservices]|
+-----------------------+              +-----------------------+
```

### Dynamic Environment Mapping

| Environment | Host VM | Provisioning Tool | Isolation |
|-------------|--------|-------------------|-----------|
| Production | Dynamic (Host 2 Default) | Terraform | Dedicated VM |
| Staging | Dynamic (Host 1 Default) | Terraform | Dedicated VM |
| Testing | Dynamic (Host 1) | Terraform | Dedicated VM |
| Dev/Feature | Dynamic (Host 1) | Terraform | Per-Branch VM |


### Core Philosophy

Since planned downtime is acceptable, the architecture uses **Active-Passive** with **manual/semi-automatic failover**. This is perfect for the use case - true automatic failover adds unnecessary complexity.

---

## 4. Component-Specific HA Strategy

### 4.1 PostgreSQL High Availability

| Aspect | Recommendation |
|--------|----------------|
| Setup | Streaming replication (async) |
| Failover | Manual `pg_promote()` + script |
| Connection | Apps connect to HAProxy VIP |
| Configuration | `synchronous_commit = off` for performance |

**Why not Patroni?**
- Requires 3 nodes or external etcd
- For planned maintenance, manual failover is sufficient
- Simpler to maintain and debug

**Failover Process:**

1. Stop PostgreSQL on Server 1
2. Run `pg_promote()` on Server 2
3. Update HAProxy to route to new primary
4. Verify application connectivity

**Backup Strategy:**
- Daily `pg_dump` to local storage on primary
- Nightly copy to standby for disaster recovery
- WAL archiving to standby

### 4.2 Production VM High Availability (Failover & Migration)

Instead of manual OS patching, failover for the dynamic **Production VM** is managed via Terraform re-provisioning.

| Scenario | Recovery Process |
| :--- | :--- |
| **VM Fault** | OpsPilot executes `terraform apply` to re-create the VM on the same host. |
| **Host 2 Fault** | OpsPilot detects failure -> executes `terraform apply -var="target_node=host1"` -> VM spins up on Host 1 -> OpsProxy redirects traffic. |
| **Rollback** | See Section 8.6 for the commit-based rollback UI. |

### 4.3 Resource Consolidation Strategy (Removing External Dependencies)

To simplify the architecture for a small userbase (2,000-10,000 users), external dependencies (Redis, RabbitMQ, MinIO) are consolidated into natively managed Go and PostgreSQL solutions.

| Component | Legacy Tool | New Consolidated Strategy |
|-----------|-------------|---------------------------|
| **Caching / Sessions** | Redis | **PostgreSQL (Shared State) + In-Memory Go Cache (Ephemeral)** |
| **Message Queue** | RabbitMQ | **PostgreSQL `SKIP LOCKED` Job Queue** |
| **Object Storage** | MinIO | **Local Filesystem + OpsPilot Sync (rsync/SCP)** |

**Rationale for Consolidation:**
1. **Caching:** 10k users generate low enough traffic that PostgreSQL can easily handle session state. High-read/low-write data can be cached in-memory directly in the Go application (`sync.Map` or `golang-lru`).
2. **Queues:** PostgreSQL `SELECT ... FOR UPDATE SKIP LOCKED` allows building a high-throughput, transactional queue without the overhead of maintaining RabbitMQ.
3. **Storage:** Since the data size is small (~18K files), storing them locally and relying on OpsPilot to run an `rsync` cron job to the standby server removes the need for MinIO entirely.

### 4.3 HAProxy + Keepalived

**Keepalived Configuration:**

```
Server 1: Master (priority 100)
Server 2: Backup (priority 90)
VIP: 192.168.1.200 (shared IP)
```

**HAProxy Configuration Example:**

```haproxy
# Frontend - Application
frontend app_frontend
    bind 192.168.1.200:80
    default_backend app_servers

backend app_servers
    server server1 192.168.1.10:8080 check
    server server2 192.168.1.11:8080 check backup

# Frontend - PostgreSQL
frontend pg_frontend
    bind 192.168.1.200:5432
    default_backend pg_primary

backend pg_primary
    server pg1 192.168.1.10:5432 check
    server pg2 192.168.1.11:5432 check backup
```

---

## 5. Secrets Management

### Option: HashiCorp Vault (Recommended)

**Structure:**

```
Vault/
├── namespace: dev/
│   ├── kv/database       # DB credentials
│   └── kv/opspilot       # OpsPilot credentials
├── namespace: staging/
├── namespace: testing/
└── namespace: prod/
```

### Application Integration (Go)

```go
import "github.com/hashicorp/vault/api"

// Connect to Vault
config := api.DefaultConfig()
config.Address = os.Getenv("VAULT_ADDR")

client, _ := api.NewClient(config)
client.SetNamespace(os.Getenv("ENVIRONMENT")) // "prod", "staging", etc.

// Fetch secrets
secret, err := client.KVv2("kv").Get(context.Background(), "database")
dbPassword := secret.Data["password"].(string)
```

### Alternative: Environment Files with GitLab CI Variables

For simpler setup, use GitLab CI/CD variables:

```
# GitLab: Settings → CI/CD → Variables
DEV_DB_HOST=192.168.1.10
PROD_DB_HOST=192.168.1.11
SSH_PRIVATE_KEY=<key>
DEPLOY_USER=deploy
```

---

## 6. Microservices Migration Strategy

### Current State

- Single Go monolith with Gin
- Everything in one repository

### Target Project Structure

```
project-root/
├── Makefile
├── .env
├── docker-compose.yml
│
├── services/                       # Each service in own folder
│   ├── user-service/
│   │   ├── main.go
│   │   ├── handler/
│   │   ├── repository/
│   │   ├── config/
│   │   └── Dockerfile
│   ├── order-service/
│   ├── product-service/
│   ├── payment-service/
│   └── notification-service/
│
├── shared/                         # Shared code (not a separate service)
│   ├── config/
│   ├── middleware/
│   ├── vault/
│   └── db/
│
└── scripts/
    ├── deploy-server1.sh
    ├── deploy-server2.sh
    └── failover.sh
```

### Migration Approach

1. **Start with modular monolith**: Keep all services in one Go binary initially
2. **Separate by package**: Group by domain (user/, order/, etc.)
3. **Extract gradually**: When a service needs independent deployment, make it its own `main.go`
4. **Communication**:
   - REST (simple, familiar)
   - gRPC (better performance, needs protobuf)
   - Event-driven (RabbitMQ for async)

### 6.4 Service Discovery & Ingress

OpsPilot manages a two-tier discovery system that avoids the need for external tools like Consul.

1.  **Internal (Microservice to Microservice):**
    - All services for a specific environment (Auth, Master Data, etc.) are deployed in the same dynamic VM.
    - They share a Docker bridge network.
    - **Resolution:** Services use container names (e.g., `http://auth:8080`) via Docker's built-in DNS.
2.  **External (Public Ingress):**
    - When Terraform provisions a VM, OpsPilot captures its IP.
    - **OpsProxy** (Section 8.5) maps the public domain (e.g., `auth-stage.yourdomain.com`) to the current VM IP in the PostgreSQL database.
    - Traffic flows: `User -> VIP -> HAProxy -> OpsProxy -> Dynamic VM IP -> Docker Port`.

---

## 7. CI/CD Architecture: OpsPilot Deployment Engine

Instead of relying on external GitLab Runners or `.gitlab-ci.yml` files, OpsPilot acts as the primary orchestrator for building and deploying code.

```
[ Developer ] --(Push)--> [ GitLab Repository ]
                                  │
                                  │ (Manual Trigger via UI)
                                  ▼
[ OpsPilot Web UI ] <---------- [ Git API ] (Fetch Commits/Branches)
      │
      ├─> [ Build Engine ] (Docker Build & Push)
      │
      └─> [ Deploy Engine ] (SSH to Server -> Pull -> Up)
```

### 8.6 OpsDeploy Module: UI-Driven Deployment & Rollbacks

#### Objective
To provide a self-contained deployment platform within OpsPilot that allows developers to browse any branch/commit and deploy it to any environment (Dev, Staging, Testing, Prod) with one click, while maintaining a full deployment history and rollback capability.

#### Features
- **Git Integration:** Connects to GitLab/GitHub via API tokens to list all branches and commits for the configured project repositories.
- **Environment Selection:** Choose the target environment (S1: Dev/Staging/Test, S2: Prod).
- **History & Logs:** Every deployment is logged with the commit hash, user, timestamp, and full build/deploy logs.
- **One-Click Rollback:** A "Rollback" button on each history entry that instantly redeploys that specific commit hash to the target environment.

#### Deployment State Model (PostgreSQL)
```sql
CREATE TABLE deployments (
    id SERIAL PRIMARY KEY,
    environment VARCHAR(50) NOT NULL, -- prod, staging, dev, testing
    service_name VARCHAR(100) NOT NULL,
    branch VARCHAR(100) NOT NULL,
    commit_hash VARCHAR(40) NOT NULL,
    commit_message TEXT,
    deployed_by VARCHAR(100), -- User name
    status VARCHAR(20),       -- 'pending', 'building', 'success', 'failed'
    logs TEXT,                 -- Captured build/deploy output
    deployed_at TIMESTAMP DEFAULT NOW()
);
```

#### Workflow
1.  **Browse:** User opens OpsPilot "Deploy" tab, selects a service, and sees a list of the latest 50 commits from the current branch (or switches branches).
2.  **Trigger:** User clicks "Deploy" for a specific commit and selects "Production".
3.  **Execute:**
    -   OpsPilot clones the repo (or uses a cached one).
    -   Runs `docker build` using the commit hash as the tag.
    -   Pushes the image to the local/private registry.
    -   SSHs to the target server, updates the `docker-compose.yml` image tag, and runs `docker-compose up -d`.
4.  **Confirm:** OpsPilot polls the service health check until it is "Healthy".

#### Rollback Logic
- Rollback is simply a "Re-deploy" of a known good `commit_hash`.
- Since every deployment is immutably tagged with its commit hash in the registry, the rollback process is identical to a standard deploy but bypasses the "Build" step if the image already exists.
```

### Environment Structure on Servers

```
# Server 1 (Development/Staging/Testing)
/opt/yourapp/
├── dev/
│   ├── docker-compose.yml
│   ├── .env (links to Vault)
│   └── config/
├── staging/
│   ├── docker-compose.yml
│   ├── .env
│   └── config/
└── testing/
    ├── docker-compose.yml
    ├── .env
    └── config/

# Server 2 (Production)
/opt/yourapp/
└── production/
    ├── docker-compose.yml
    ├── .env (links to Vault)
    └── config/
```

### Deployment Workflow

```
1. Developer pushes code to feature branch
2. Pipeline runs: Build → Test (automatic)
3. Manual trigger: Click "Build Dev Image" → "Deploy to Dev"
4. Verify in dev environment
5. Repeat for staging when ready
6. For production: Click "Deploy to Production" (with approval)
```

### GitLab CI/CD Variables

Configure these in: Project → Settings → CI/CD → Variables

| Variable | Description | Protected |
|----------|-------------|-----------|
| SSH_PRIVATE_KEY | SSH key for deployment | Yes |
| DEPLOY_USER | SSH username | Yes |
| DEV_DEPLOY_HOST | Dev server IP | Yes |
| STAGING_DEPLOY_HOST | Staging server IP | Yes |
| PRODUCTION_DEPLOY_HOST | Production server IP | Yes |
| CI_REGISTRY_USER | GitLab registry username | No |
| CI_REGISTRY_PASSWORD | GitLab registry password | Yes |

---

## 8. Custom OpsPilot Tool Proposal

### Why Build Custom?

| Concern with Multi-Tool | Reality |
|------------------------|---------|
| Too many dependencies | 6+ tools to update |
| Breaking on updates | One tool update can break integration |
| Maintenance overhead | Each tool has separate config, logs |
| Beginner-friendly | Steep learning curve for each tool |

### The Case for OpsPilot

- Your requirements are simple - don't need Kubernetes-level complexity
- You're a Go developer - you have the skills
- Maintenance becomes easier - one tool to understand
- Learning curve for team - one UI instead of 6+ tools
- Customization - fits your exact workflow

### Proposed Architecture

OpsPilot is a Go application that acts as an orchestration layer for Terraform and Docker.

```
+---------------------------------------------------------------+
|                      OpsPilot (Go + Gin)                       |
+---------------------------------------------------------------+
|  Web UI (HTML/HTMX)  |  API Layer  |  Scheduler              |
+----------------------+-------------+-------------------------+
|                       Core Modules                             |
|  Terraform Engine |  Deploy Manager  |  Proxy Manager        |
|  Health Monitor   |  Backup Manager  |  Log Aggregator       |
+----------------------+-------------+-------------------------+
|                       Execution Layer                          |
|  hashicorp/tf-exec |  SSH Client  |  Docker Client         |
+----------------------+-------------+---------------------------+
```

### Features

| Feature | Description | Priority |
|---------|-------------|----------|
| Dashboard | Server status, service health, VM IPs | MVP |
| **TF Provisioning** | Dynamic Proxmox VM lifecycle management | MVP |
| Deploy Button | One-click deploy to dynamic VMs | MVP |
| Environment Manager | Define workspaces for each branch/env | MVP |
| Server Health | Docker status, service status | MVP |
| Failover Control | Manual/Auto migration via TF | MVP |
| Secret Management | UI to manage environment variables | Extended |
| Log Viewer | Simple log aggregation | Extended |
| Backup Manager | PostgreSQL backup/restore buttons | Extended |

### 8.1 Terraform Provisioning Module (Internal Developer Platform)

#### Objective
To automate the creation of isolated, per-environment Virtual Machines on Proxmox, providing developers with a "Kubernetes-like" experience without the complexity.

#### Implementation
- **Library:** `github.com/hashicorp/terraform-exec` for native Go execution.
- **Workflow:**
  - OpsPilot clones the Terraform repository or generates a workspace.
  - Generates `terraform.tfvars` with dynamic values (VM ID, IP, Host Node).
  - Runs `terraform init`, `plan`, and `apply`.
  - Captures the VM's private IP from Terraform outputs and saves it to PostgreSQL.

### Complexity Assessment

| Phase | Features | Time to Build |
|-------|----------|---------------|
| Phase 1 | Web UI, Deploy, Health Check, Environment Config | 2-3 weeks |
| Phase 2 | Git Webhooks, Log Viewer, Backup/Restore | 1-2 months |
| Phase 3 | Add features as needed | Ongoing |

### Technology Stack

| Layer | Technology |
|-------|------------|
| Backend | Go + Gin |
| Database | SQLite (embedded, no setup) |
| Frontend | Go templates + HTMX + **Tailwind CSS (Responsive for mobile)** |
| SSH | golang.org/x/crypto/ssh |
| Docker | docker/client-go |
| Scheduling | robfig/cron |

### MVP Data Model

```go
// Server represents a physical server
type Server struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    IP        string `json:"ip"`
    Role      string `json:"role"` // "primary", "standby"
    Status    string `json:"status"` // "healthy", "unhealthy"
}

// Environment represents a deployment environment
type Environment struct {
    ID        string `json:"id"`
    Name      string `json:"name"` // "dev", "staging", "production"
    ServerID  string `json:"server_id"`
    Services  []Service `json:"services"`
}

// Service represents a deployed microservice
type Service struct {
    Name        string `json:"name"`
    Image       string `json:"image"`
    Port        int    `json:"port"`
    Status      string `json:"status"`
    LastDeploy  string `json:"last_deploy"`
}

// Secret represents environment configuration
type Secret struct {
    ID          string `json:"id"`
    Environment string `json:"environment"`
    Key         string `json:"key"`
    Value       string `json:"value"` // encrypted
}
```

### Advantages Over Multi-Tool

| Aspect | Multi-Tool | Custom Tool |
|--------|-----------|-------------|
| Dependencies | 6+ tools | One tool |
| Debugging | Check multiple logs | One place |
| Onboarding | Learn 6+ tools | Learn one UI |
| Failure Points | 6+ integration points | Single point |
| Updates | Coordinating updates | One update |

### 8.5 OpsProxy Module: Native Reverse Proxy & Manual SSL Management

#### Objective
To eliminate the complexity of synchronizing Nginx Proxy Manager (NPM) across two servers by building a native Go-based reverse proxy that uses a shared PostgreSQL database for configuration and SSL certificate storage. This includes a manual management workflow for vendor-supplied wildcard certificates.

#### Core Technologies
- **Proxy Engine:** `net/http/httputil.ReverseProxy` (Go standard library).
- **SSL Management:** Native Go `crypto/tls` implementation with PostgreSQL storage.
- **Protocol Support:** Supports **HTTP/1.1**, **HTTP/2**, and **gRPC** (via `h2c` internal routing).
- **Workflow:** Manual certificate upload with "Test-then-Deploy" logic.

#### Data Model (PostgreSQL)
```sql
-- Proxy Routes
CREATE TABLE proxy_routes (
    id SERIAL PRIMARY KEY,
    domain VARCHAR(255) UNIQUE NOT NULL, -- e.g., app.yourdomain.com
    target_url VARCHAR(255) NOT NULL,    -- e.g., http://localhost:8080
    environment VARCHAR(50),             -- prod, staging, dev
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

-- SSL Storage (Vendor Certificates)
CREATE TABLE certificates (
    id SERIAL PRIMARY KEY,
    label VARCHAR(100),          -- e.g., "Wildcard-2026-Q1"
    cert_data TEXT NOT NULL,      -- fullchain.pem
    key_data TEXT NOT NULL,       -- privkey.pem
    is_production BOOLEAN DEFAULT false, -- If true, used for all prod domains
    created_at TIMESTAMP DEFAULT NOW()
);

-- Test Overrides (Test before Global deployment)
CREATE TABLE cert_test_overrides (
    domain VARCHAR(255) PRIMARY KEY, -- e.g., "test-ssl.yourdomain.com"
    cert_id INTEGER REFERENCES certificates(id)
);
```

#### Manual SSL Workflow (Test-then-Deploy)
1.  **Upload:** Use the OpsPilot UI to upload the `fullchain.pem` and `privkey.pem` provided by the vendor. Label it (e.g., `2026-Update`).
2.  **Test Assignment:** Assign the new certificate to a specific test domain (e.g., `test-check.yourdomain.com`) in the `cert_test_overrides` table.
3.  **Verification:** Visit the test domain to ensure the new certificate is served correctly by OpsProxy.
4.  **Promote:** Click "Promote to Global" in the UI. OpsPilot sets `is_production = false` for the old cert and `is_production = true` for the new cert, triggering an in-memory TLS config reload.

#### Key Advantages
- **Risk Mitigation:** No accidental global SSL breakage; every certificate is tested on a subdomain first.
- **No File Syncing:** Certificates and routes are in the DB. Server 2 instantly has the same configuration as Server 1.
- **Unified Logs:** Proxy access/error logs are natively available to the OpsPilot Log Aggregator.

#### Implementation Snippet (Conceptual)
```go
func (p *OpsProxy) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error) {
    // 1. Check for a test override first
    certID := p.db.GetOverrideID(hello.ServerName)
    
    // 2. If no override, get the global production certificate
    if certID == 0 {
        certID = p.db.GetGlobalProductionCertID()
    }

    // 3. Load from DB and parse
    certRecord := p.db.GetCert(certID)
    tlsCert, err := tls.X509KeyPair([]byte(certRecord.Cert), []byte(certRecord.Key))
    return &tlsCert, err
}
```

---

### Firewall & Ingress Strategy (Minimal Footprint)

To maximize security, only two ports are exposed on the physical firewall to the HAProxy VIP:
- **Port 80 (HTTP):** Redirects to 443.
- **Port 443 (HTTPS):** All traffic (Apps, gRPC, Admin Tools) flows through here.

OpsProxy routes traffic based on the `Host` header, eliminating the need to open separate ports for internal tools.

### 8.7 Windows DNS Integration Module

#### Objective
To automate or verify the creation of DNS records in the Windows Server DNS environment during deployment.

#### Implementation Options
1. **Automated (SSH/PowerShell):** OpsPilot SSHs into the Windows DNS server and executes `Add-DnsServerResourceRecordA`.
2. **Manual (Verification Loop):** OpsPilot displays the required `A` record and waits for the user to click **"Record Added Manually"** before proceeding with the Terraform/Docker deployment.

---

### 8.8 Managed Management Tool Stack (HA)

OpsPilot natively manages the deployment and high availability of critical management tools. These tools are deployed as containers on the **OpsControl** static VMs.

| Tool | Purpose | HA Strategy |
| :--- | :--- | :--- |
| **Slash** | URL Shortener / Entrypoint | Replicated on S1/S2 via OpsProxy routing. |
| **Cloudbeaver** | Database Management (Web) | Deployed on S1, failover to S2. |
| **pgAdmin** | PostgreSQL Administration | Deployed on S1, failover to S2. |
| **Databaseus** | Backup/Restore Orchestrator | Syncs backups to GDrive, Zoho, S3, Local. |

---

### 8.9 Logging & Observation Module (Websocket Streaming)

#### Objective
To provide real-time container log visibility directly in the OpsPilot Web UI without requiring terminal access.

#### Implementation
- **Backend:** Uses `docker/client-go` to open a log stream from the target VM.
- **Transport:** Streams log chunks through a **Websocket** from OpsPilot to the browser.
- **Frontend:** HTMX/Alpine.js renders the live log feed in a terminal-like component.

### 8.10 OpsVisualizer Module (Topology Map)

#### Objective
To provide a real-time, visual, and interactive map of the network architecture as seen in tools like n8n, allowing for instant "at-a-glance" status monitoring.

#### Features
- **Node-Based Layout:** Automatically generated nodes for Firewalls, VIPs, Static Control Plane VMs, Dynamic App VMs, and Docker Containers.
- **Edge Mapping:** Lines showing traffic flow (HTTP/gRPC) and database connections.
- **Live Health Status:** Dynamic status colors based on real-time health checks (Green = Healthy, Red = Down, Blue = Provisioning).
- **Interactive Nodes:** Click any container to instantly view logs, restart it, or see its environment variables.

#### Implementation
- **Backend:** A Go layout generator that builds a node-edge JSON object from PostgreSQL (`proxy_routes`, `deployments`) and Terraform state.
- **Frontend:** Integrated via **Cytoscape.js** or **Go SVG generation**, rendered seamlessly with **HTMX**.

---

## 9. Maintenance Workflow

### Planned Maintenance Process

```
+------------------------------------------------------------------+
|                     PLANNED MAINTENANCE WORKFLOW                  |
+------------------------------------------------------------------+

1. PRE-MAINTENANCE (Day before)
   + Verify replication is working (PostgreSQL lag = 0)
   + Backup PostgreSQL to both servers
   + Note current VIP holder

2. START MAINTENANCE
   a. Stop accepting new connections (optional: HAProxy drain mode)
   b. Run final backup
   c. Stop services on Server 1 (standby)

3. PERFORM MAINTENANCE
   + Update code, OS patches, etc. on Server 1

4. SWITCHOVER (to Server 1)
   a. On Server 2: pg_promote() to make it primary
   b. Update HAProxy: Server 2 is now primary
   c. Test applications

5. UPDATE SERVER 2 (now standby)
   + Perform maintenance
   + Reconfigure as replica of Server 1

6. FAILBACK (optional - can skip and keep Server 2 as primary)
   a. Promote Server 1 back to primary
   b. Update HAProxy
   c. Verify everything works
```

### Automated Script Example

```bash
#!/bin/bash
# failover.sh - Automated failover script

set -e

echo "=== Starting Failover ==="

# Arguments
NEW_PRIMARY=$1  # IP of new primary server
NEW_STANDBY=$2  # IP of new standby server

echo "Promoting $NEW_PRIMARY to primary..."

# SSH to new primary and promote
ssh deploy@$NEW_PRIMARY "
    sudo -u postgres pg_ctl promote -D /var/lib/postgresql/16/main
    echo 'PostgreSQL promoted'
"

echo "Updating HAProxy configuration..."
# Update HAProxy to point to new primary
ssh deploy@$NEW_PRIMARY "
    docker-compose restart haproxy
"

echo "Verifying connectivity..."
sleep 5
# Add verification commands here

echo "=== Failover Complete ==="
echo "Primary: $NEW_PRIMARY"
echo "Standby: $NEW_STANDBY"
```

---

## 10. Next Steps

### Option A: Multi-Tool Implementation

1. Set up Keepalived + HAProxy on both servers
2. Configure PostgreSQL streaming replication
3. Deploy Vault for secrets management
4. Set up GitLab Runner on Server 1
5. Create CI/CD pipelines for each environment
6. Document failover procedures
7. Test failover in staging first

### Option B: Custom OpsPilot Tool

**Phase 1 - MVP (2-3 weeks):**
1. Design architecture and data models
2. Build HTTP server with Gin routing
3. Create HTML dashboard with HTMX
4. Implement SSH deployment commands
5. Add basic health checks
6. Test with staging environment

**Phase 2 (1-2 months):**
1. Git webhook integration
2. Log aggregation viewer
3. Backup/restore UI
4. PostgreSQL backup automation

**Phase 3 (Ongoing):**
1. Add features as needed
2. Improve monitoring
3. Add alerting

---

## 11. Resilience & Governance

### 11.1 High-Availability Docker Registry (Active-Active Autosync)
**Objective:** Ensure the Production VM (Host 2) can always pull images even if OpsControl-01 (Host 1) goes down.
**Implementation:**
- OpsPilot deploys a private Docker Registry container on both OpsControl-01 and OpsControl-02.
- OpsPilot runs a Go routine (Registry Sync Worker) that polls the Registry HTTP API every 60 seconds on both nodes.
- When a new image/tag is pushed to the primary registry, the worker instantly triggers a `docker pull` from the primary and a `docker push` to the standby registry. It also handles image pruning/deletions.

### 11.2 Live Metrics Stream & Time-Series DB
**Objective:** Move beyond static "Green/Red" status to real-time performance streaming and historical metric retention (1+ weeks) for seamless troubleshooting.
**Implementation:**
- **Time-Series Database:** Deploy **VictoriaMetrics** (or Prometheus) inside the OpsControl management stack. It is incredibly lightweight, highly performant, and easily retains a week of metrics with negligible disk space.
- **Data Collection:** Docker container metrics (CPU/RAM/Network) are scraped automatically via OpsPilot and stored in VictoriaMetrics.
- **Live Stream (Zero-Downtime Monitoring):** OpsPilot opens a Websocket to the UI that streams the live `docker stats` output (refreshing per second) so developers can watch CPU utilization during migrations.

### 11.3 Identity, Auth & Audit
**Objective:** Prevent unauthorized or accidental destructive actions (like `terraform destroy`) and track all system mutations.
**Implementation:**
- **Authentication:** All OpsPilot users must authenticate using an Email OTP or a TOTP authenticator app (Google Authenticator/Authy).
- **Granular RBAC:** 
    - A **Master Admin** role has full control over the platform.
    - Master Admin can define custom roles and assign specific permissions per module (e.g., "Allow Deploy to Dev but Deny to Prod").
    - Permissions are enforced at the API level for every module (OpsProxy, OpsDeploy, Terraform, etc.).
- **Immutable Audit Trail:** Every API request that mutates state (Deploy, Delete, Change Cert) is recorded in PostgreSQL: `[Timestamp] [User] [Action] [Target] [IP]`.

### 11.4 PostgreSQL Point-In-Time Recovery (Zero Data Loss)
**Objective:** Protect the OpsPilot Control Plane data (Proxy routes, job queues, deployment history, SSL certs) from corruption or catastrophic failure between the 1-hour backup windows.
**Implementation:**
- **WAL Archiving:** Configure PostgreSQL to stream its Write-Ahead Logs (WAL) continuously using a tool like **pgBackRest** or **WAL-G**.
- **Storage Target:** Ship the WAL segments directly to an external object store (e.g., S3, Google Drive via the `Databaseus` tool) or the Standby server.
- **Recovery:** In the event of a database failure at 10:43 AM, the admin can restore the 10:00 AM base backup and replay the WALs up to exactly 10:42:59 AM, achieving effectively **zero data loss**.

---

## 12. Advanced Operations & Safety

### 12.1 Environment Lifecycle & Resource Quotas
**Objective:** Prevent resource exhaustion on Proxmox hosts due to "abandoned" dynamic VMs.
**Implementation:**
- **TTL (Time-To-Live):** When creating a Dev/Feature environment, users must select an expiry period (e.g., 24h, 3 days). OpsPilot will automatically run `terraform destroy` when the TTL expires unless manually extended.
- **Resource Quotas:** OpsPilot enforces limits on total CPU/RAM that can be requested per user or per environment type to ensure the Production VM is never starved of resources.

### 12.2 Proactive Alerting & Self-Healing
**Objective:** Ensure rapid response to critical failures without constant manual dashboard monitoring.
**Implementation:**
- **Notification Engine:** Integrates with Telegram/Email/Webhooks. Alerts are triggered for:
    - Service transitions from `Healthy` to `Down`.
    - Physical host disk/RAM usage exceeding 90%.
    - Terraform provisioning failures.
- **Self-Healing:** OpsPilot can be configured to attempt one "Auto-Restart" of a container or "Auto-Migration" of a VM before escalating to a human alert.

### 12.3 Automated Security Scanning (Trivy)
**Objective:** Prevent the deployment of vulnerable Docker images to Production.
**Implementation:**
- **Pre-Deploy Scan:** The **OpsDeploy** module integrates **Trivy** (Go-based vulnerability scanner).
- **Blocking Logic:** When an image is built, OpsPilot runs a scan. If "Critical" vulnerabilities are detected, the deployment to Production is blocked, and a detailed report is shown in the UI.

### 12.4 Graceful Migration & Traffic Draining
**Objective:** Achieve true zero-downtime during Production VM migrations between Proxmox hosts.
**Implementation:**
- **HAProxy Orchestration:** Instead of a hard switch, OpsPilot manages a multi-step "Drain" process:
    1. Set old VM to `DRAIN` mode in HAProxy (no new sessions).
    2. Provision new VM on Host 1.
    3. Verify New VM health.
    4. Switch traffic to New VM.
    5. Wait for active connections on the old VM to terminate gracefully before destruction.

---

## Summary Table

| Component | HA Approach | Failover Type | Complexity |
|-----------|-------------|---------------|------------|
| **Compute** | **Dynamic VMs (Terraform)** | **Terraform Re-provision** | **Medium** |
| **Reverse Proxy** | **OpsProxy (Native Go)** | **Active-Active (DB Synced)** | **Low (Unified)** |
| **Service Discovery**| **Docker DNS + OpsProxy** | **Automatic (DB Mapping)** | **Low** |
| **Visibility** | **OpsVisualizer (Graph)** | **N/A (Real-time)** | **Low** |
| PostgreSQL | Streaming replication | Manual (script) | Low |
| State/Queues | Handled by PostgreSQL | DB Failover | Low |
| Local Storage | OpsPilot rsync cron | Manual | Low |
| Vault | Single (PostgreSQL backend) | N/A | Low |
| Microservices | Containerized (Docker) | Dynamic Deploy | Low |
| VIP | Keepalived | Automatic | Low |
| CI/CD | OpsDeploy (Native) | UI-Driven | Low |

---

## Appendix: GitLab Environment Setup

### Create Environments in GitLab

1. Go to **Project → Operations → Environments**
2. Click **New environment**
3. Create:
   - `development` → URL: https://dev.yourdomain.com
   - `staging` → URL: https://staging.yourdomain.com
   - `testing` → URL: https://testing.yourdomain.com
   - `production` → URL: https://yourdomain.com

### Protect Production Environment

1. Go to **Project → Settings → CI/CD → Environments**
2. Edit production environment
3. Enable **Protected** - only maintainers can deploy

### Add Approval for Production

1. Go to **Project → Settings → CI/CD → General pipelines**
2. Enable **Require approval for production**
3. Configure required number of approvals

---

*Document created based on architecture discussions*
*Last updated: 2026-03-05*
