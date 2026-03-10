# Plan for Track 1.1: Authentication & Identity Core

## Phase 1: Data Models & Persistence [checkpoint: ]
**Goal:** Define and migrate the database models for users, roles, and permissions.

- [ ] Task: Create User and Role models in `internal/models/models.go`
- [ ] Task: Implement GORM database migration logic
- [ ] Task: Seed initial roles (Admin, Developer, Viewer) and permissions
- [ ] Task: Conductor - User Manual Verification 'Data Models & Persistence' (Protocol in workflow.md)

## Phase 2: Core Authentication Logic [checkpoint: ]
**Goal:** Implement JWT generation/validation and TOTP verification.

- [ ] Task: Write tests for JWT provider logic
- [ ] Task: Implement JWT provider in `internal/auth/auth.go`
- [ ] Task: Write tests for TOTP enrollment and verification
- [ ] Task: Implement TOTP logic (using a standard library like `pquerna/otp`)
- [ ] Task: Conductor - User Manual Verification 'Core Authentication Logic' (Protocol in workflow.md)

## Phase 3: Middleware & Security Integration [checkpoint: ]
**Goal:** Protect routes and enforce RBAC.

- [ ] Task: Write tests for `AuthMiddleware`
- [ ] Task: Implement `AuthMiddleware` in `internal/auth/auth.go`
- [ ] Task: Write tests for `RequirePermission` RBAC middleware
- [ ] Task: Implement `RequirePermission` middleware
- [ ] Task: Conductor - User Manual Verification 'Middleware & Security Integration' (Protocol in workflow.md)
