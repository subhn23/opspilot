# Plan for Track 2.1: Terraform Orchestration Engine

## Phase 1: Environment Models & Setup [checkpoint: ]
**Goal:** Define the `Environment` model and prepare the Terraform workspace manager.

- [x] Task: Verify/Create `Environment` model in `internal/models/models.go` (d3198d5)
- [ ] Task: Ensure `AutoMigrate` includes the `Environment` model
- [ ] Task: Conductor - User Manual Verification 'Environment Models & Setup' (Protocol in workflow.md)

## Phase 2: TFEngine Core & Testing Strategy [checkpoint: ]
**Goal:** Build the `TFEngine` structure and establish a testable interface.

- [ ] Task: Define a `TerraformClient` interface to wrap `tfexec` methods for mocking
- [ ] Task: Write tests for `Provision` using a mock `TerraformClient`
- [ ] Task: Write tests for `Destroy` using a mock `TerraformClient`
- [ ] Task: Conductor - User Manual Verification 'TFEngine Core & Testing Strategy' (Protocol in workflow.md)

## Phase 3: Terraform Execution Integration [checkpoint: ]
**Goal:** Implement the concrete `terraform-exec` calls within the engine.

- [ ] Task: Implement `Provision` using `tfexec` calls in `internal/terraform/terraform.go`
- [ ] Task: Implement `Destroy` using `tfexec` calls
- [ ] Task: Implement `setupTF` workspace preparation logic
- [ ] Task: Conductor - User Manual Verification 'Terraform Execution Integration' (Protocol in workflow.md)
