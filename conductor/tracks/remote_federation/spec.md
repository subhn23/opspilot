# Specification: Remote Host & Federation Management

## Goal
Enable OpsPilot to manage deployments across disparate infrastructure, moving away from a strict dependency on a local Proxmox cluster. This allows developers to test deployments on local "Static Hosts" (like a Raspberry Pi or EC2 instance) and enables multi-datacenter management via a Master-Worker architecture.

## Scope
1. **TargetHost Data Model:** Introduce a new model to track underlying physical/virtual machines or remote OpsPilot APIs.
2. **Agentless SSH Management:** Ability to securely store SSH keys and trigger `docker-compose` deployments on raw remote Linux servers.
3. **OpsPilot Federation (API):** Ability for a "Master" OpsPilot to trigger provisioning and deployment tasks on a "Worker" OpsPilot via authenticated REST endpoints.
4. **Environment Abstraction:** Update the `Environment` model to link to a `TargetHost` rather than a hardcoded `HostNode`.

## Architecture Details
- **Type `local_proxmox`**: The current behavior. Uses `terraform-exec` locally.
- **Type `remote_ssh`**: Bypasses Terraform. Uses `golang.org/x/crypto/ssh` to connect directly to the target IP, pull the docker image, and run it.
- **Type `federated_opspilot`**: Acts as an API proxy. The Master sends a JSON payload to the Worker's `/api/federation/deploy` endpoint, and the Worker handles local Proxmox/Docker steps.

## Security Considerations
- SSH keys and API tokens in the `TargetHost` table must be encrypted at rest using a platform-level symmetric key.
- Federation endpoints must enforce strict JWT or Token-based authentication.
