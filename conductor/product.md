# Initial Concept
OpsPilot is a self-hosted, Go-based orchestrator for infrastructure and CI/CD, designed to manage Proxmox Virtual Machines and Docker-based microservices in a high-availability setup.

# Product Definition

## Project Goal
OpsPilot aims to provide a "Kubernetes-like" experience for developers and IT teams who manage their own on-premise infrastructure. It streamlines the lifecycle of dynamic VMs, simplifies the deployment pipeline, and offers real-time observability through an interactive topology map.

## Target Users
- **DevOps/Platform Teams:** Managing environments and CI/CD for their organizations.
- **Individual Developers:** Self-hosters looking for an easy way to deploy and manage projects.
- **IT Administrators:** Organizations seeking a self-hosted alternative to cloud platforms.

## Core Goals
- **Dynamic VM Lifecycle:** Efficient provisioning and teardown of Proxmox-based VMs using Terraform.
- **Simplified CI/CD Pipeline:** One-click deployment from Git branches directly to target environments.
- **Observability & Visualization:** Real-time monitoring and an interactive topology map for clear stack status.

## Key Features
- **Distributed Control Plane:** Ability to manage agentless remote SSH hosts and orchestrate multi-datacenter setups via OpsPilot Federation (Master-Worker APIs).
- **Audit & Identity:** Robust authentication with JWT and TOTP MFA, coupled with a granular RBAC system and immutable action logging.
- **OpsProxy & SSL:** Native Go-based reverse proxy with integrated SSL certificate management.
- **Automated Windows DNS:** Automated management of A records via PowerShell over SSH, ensuring consistent internal naming.
- **OpsVisualizer Map:** Interactive node-based map for visualizing the network and service architecture.
- **OpsMetric (Live Monitoring):** Real-time performance metrics (CPU/Memory) for Docker containers using VictoriaMetrics and WebSocket streaming.
- **Security & Backup:** Automated Trivy CVE scanning for Docker images and robust Postgres WAL archiving for point-in-time recovery.
- **Multi-Env Support:** Robust support for Production, Staging, Testing, and Feature environments.

## Constraints & Requirements
- **Self-Hosted/On-Premise:** No external cloud dependencies; must run entirely on-site.
- **2-Server HA Architecture:** Designed for high availability on a minimal 2-server physical setup.
- **Inferred Tech Stack Compliance:** Built with Go, PostgreSQL, Gin, Tailwind CSS, and Terraform.
