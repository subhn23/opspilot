# Implementation Plan: Remote Host & Federation Management

## Phase 1: Data Models & Migration [checkpoint: bba7f86]
**Goal:** Establish the `TargetHost` entity and update `Environment`.

- [x] Task: Create `TargetHost` model in `internal/models/models.go` with types (`local_proxmox`, `remote_ssh`, `federated_opspilot`). (112e78d)
- [x] Task: Update the `Environment` model to belong to a `TargetHost`. (177815a)
- [x] Task: Create a database migration script/seed to convert existing `HostNode` strings into `TargetHost` records. (5552988)
- [x] Task: Implement a utility package for AES encryption/decryption of `AuthData` (SSH keys/tokens) at rest. (4a284d4)

## Phase 2: Agentless SSH Deployments [checkpoint: 36d8a2b]
**Goal:** Allow deployments to raw Linux machines without OpsPilot/Proxmox.

- [x] Task: Create a UI page (`ui/templates/hosts.html`) to manage (Add/Edit/Delete) Static Hosts and input SSH keys. (7008c85)
- [x] Task: Add API endpoints in `main.go` for managing `TargetHost` records. (2c09061)
- [x] Task: Update the Provisioning Wizard (`env_wizard.html`) to select from available `TargetHost` records instead of hardcoded strings. (ec8ea4d)
- [x] Task: Refactor `Deployer.RemoteUp` to dynamically load the SSH configuration from the `TargetHost` record (decrypting the key) instead of relying on local environment variables. (ffa970c)

## Phase 3: OpsPilot Federation
**Goal:** Enable Master-Worker communication between OpsPilot instances.

- [x] Task: Add a UI section to register a "Federated OpsPilot" (URL and API Token). (36a5ee6)
- [x] Task: Create `/api/federation/deploy` endpoint on the Worker node to accept incoming deployment payloads and trigger local execution. (5448efc)
- [x] Task: Implement a `FederatedClient` in `internal/deploy/` that the Master node uses to forward `BuildAndPush` and `RemoteUp` commands to the Worker node via HTTP POST. (7061f4c)
- [x] Task: Add logic to `Deployer` to check the `TargetHost.Type`. If it's `federated_opspilot`, route the request through the `FederatedClient` instead of local Docker/SSH. (f885e76)
