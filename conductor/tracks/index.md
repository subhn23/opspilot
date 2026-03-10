# OpsPilot Implementation Index

## [Phase 1: Foundation (Core Control Plane)](./phase-1-foundation/plan.md)
*   Base Application & UI Scaffold (Go + Gin + Tailwind + HTMX)
*   OpsProxy & SSL Management (L7 Routing, Manual SSL Versioning)
*   Audit & Identity (TOTP, RBAC, Action Logging)

## [Phase 2: Dynamic Infrastructure (Terraform & Docker)](./phase-2-dynamic-infra/plan.md)
*   Terraform Orchestration (Proxmox Lifecycle via `terraform-exec`)
*   OpsDeploy Engine (Commit Browsing, Build & Remote Deploy)
*   Windows DNS Module (Automated/Verified Records)

## [Phase 3: Governance & Operations (Visibility & Safety)](./phase-3-governance-ops/plan.md)
*   OpsVisualizer (Interactive Topology Map)
*   OpsMetric (VictoriaMetrics, Live Docker Stats Streaming)
*   Security & Backup (Trivy CVE Scanning, Postgres WAL Archiving)
