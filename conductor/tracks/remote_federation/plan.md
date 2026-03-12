# Implementation Plan: Remote Host & Federation Management

## Phase 1: Data Models & Migration
**Goal:** Establish the `TargetHost` entity and update `Environment`.

- [ ] Task: Create `TargetHost` model in `internal/models/models.go` with types (`local_proxmox`, `remote_ssh`, `federated_opspilot`).
- [ ] Task: Update the `Environment` model to belong to a `TargetHost`.
- [ ] Task: Create a database migration script/seed to convert existing `HostNode` strings into `TargetHost` records.
- [ ] Task: Implement a utility package for AES encryption/decryption of `AuthData` (SSH keys/tokens) at rest.

## Phase 2: Agentless SSH Deployments
**Goal:** Allow deployments to raw Linux machines without OpsPilot/Proxmox.

- [ ] Task: Create a UI page (`ui/templates/hosts.html`) to manage (Add/Edit/Delete) Static Hosts and input SSH keys.
- [ ] Task: Add API endpoints in `main.go` for managing `TargetHost` records.
- [ ] Task: Update the Provisioning Wizard (`env_wizard.html`) to select from available `TargetHost` records instead of hardcoded strings.
- [ ] Task: Refactor `Deployer.RemoteUp` to dynamically load the SSH configuration from the `TargetHost` record (decrypting the key) instead of relying on local environment variables.

## Phase 3: OpsPilot Federation
**Goal:** Enable Master-Worker communication between OpsPilot instances.

- [ ] Task: Add a UI section to register a "Federated OpsPilot" (URL and API Token).
- [ ] Task: Create `/api/federation/deploy` endpoint on the Worker node to accept incoming deployment payloads and trigger local execution.
- [ ] Task: Implement a `FederatedClient` in `internal/deploy/` that the Master node uses to forward `BuildAndPush` and `RemoteUp` commands to the Worker node via HTTP POST.
- [ ] Task: Add logic to `Deployer` to check the `TargetHost.Type`. If it's `federated_opspilot`, route the request through the `FederatedClient` instead of local Docker/SSH.
