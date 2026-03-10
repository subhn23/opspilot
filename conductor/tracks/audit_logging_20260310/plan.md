# Plan for Track 1.3: Audit & Action Logging

## Phase 1: Data Model ## Phase 1: Data Model & Persistence [checkpoint: ] Persistence [checkpoint: ffc000f]
**Goal:** Define and migrate the `AuditLog` model.

- [x] Task: Verify `AuditLog` model in `internal/models/models.go` (0f6340a)
- [x] Task: Implement/Verify GORM migration for `AuditLog` (24bb379)
- [x] Task: Conductor - User Manual Verification 'Data Model - [ ] Task: Conductor - User Manual Verification 'Data Model & Persistence' (Protocol in workflow.md) Persistence' (Protocol in workflow.md)

## Phase 2: Core Logging Logic [checkpoint: ]
**Goal:** Implement the `LogAction` function and unit tests.

- [x] Task: Write tests for `LogAction` (82e1122)
- [ ] Task: Implement `LogAction` in `internal/auth/auth.go` (or a more suitable location)
- [ ] Task: Conductor - User Manual Verification 'Core Logging Logic' (Protocol in workflow.md)

## Phase 3: System-wide Integration [checkpoint: ]
**Goal:** Integrate logging into existing modules.

- [ ] Task: Add audit logging to `AuthMiddleware` (for failed logins if applicable)
- [ ] Task: Add audit logging to `OpsProxy` route updates
- [ ] Task: Conductor - User Manual Verification 'System-wide Integration' (Protocol in workflow.md)
