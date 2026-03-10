# Track 2.1: Terraform Orchestration Engine Spec

## Goal
Implement a reliable orchestration engine that interfaces with Proxmox via `terraform-exec` to dynamically provision, destroy, and manage the lifecycle of Virtual Machines for different environments.

## Components

### 1. Terraform Engine (`TFEngine`)
- **Wrapper around `terraform-exec`:** Provides a Go-native interface to execute Terraform commands (`init`, `apply`, `destroy`, `output`).
- **Workspace Management:** Dynamically creates working directories for each environment and copies base Terraform templates into them.
- **State Management:** (Local for now, potentially remote later) Manages the `terraform.tfstate` for each provisioned environment within its workspace.

### 2. Environment Lifecycle Methods
- **`Provision(env *Environment)`:**
  - Sets up the workspace.
  - Injects dynamic variables (e.g., `vm_name`, `target_node`, `vm_id`).
  - Runs `terraform apply`.
  - Parses outputs to extract the assigned IP address.
  - Updates the `Environment` status in the database (e.g., `PROVISIONING` -> `HEALTHY`).
- **`Destroy(env *Environment)`:**
  - Runs `terraform destroy` in the environment's workspace.
  - Updates the `Environment` status to `DESTROYED`.
- **`Migrate(env *Environment, targetNode string)`:** (Future/Advanced)
  - Updates the `target_node` variable and applies changes to move a VM between physical hosts.

## Success Criteria
- [ ] `TFEngine` can successfully initialize a workspace and run `terraform init`.
- [ ] `Provision` successfully executes an apply and correctly parses the output IP address.
- [ ] `Destroy` successfully tears down the resources.
- [ ] Environment statuses are correctly updated in the database during operations.
- [ ] Unit/Integration tests simulate the `terraform-exec` behavior or mock it effectively.
- [ ] Code has >80% test coverage.
