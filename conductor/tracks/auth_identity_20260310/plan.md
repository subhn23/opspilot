# Plan for Track 1.1: Authentication & Identity Core

## Phase 1: Data Models & Persistence [checkpoint: 39c37f3]
**Goal:** Define and migrate the database models for users, roles, and permissions.

- [x] Task: Create User and Role models in `internal/models/models.go` (7d38b69)
- [x] Task: Implement GORM database migration logic (ec0b604)
- [x] Task: Seed initial roles (Admin, Developer, Viewer) and permissions (6326e23)
- [x] Task: Conductor - User Manual Verification 'Data Models & Persistence' (Protocol in workflow.md)

## Phase 2: Core Authentication Logic [checkpoint: ]
**Goal:** Implement JWT generation/validation and TOTP verification.

- [x] Task: Write tests for JWT provider logic (996bb8a)
- [x] Task: Implement JWT provider in `internal/auth/auth.go` (6caac10)
- [x] Task: Write tests for TOTP enrollment and verification (cc30881)
- [x] Task: Implement TOTP logic (using a standard library like `pquerna/otp`) (639ce6a)
- [ ] Task: Conductor - User Manual Verification 'Core Authentication Logic' (Protocol in workflow.md)

## Phase 3: Middleware & Security Integration [checkpoint: ]
**Goal:** Protect routes and enforce RBAC.

- [ ] Task: Write tests for `AuthMiddleware`
- [ ] Task: Implement `AuthMiddleware` in `internal/auth/auth.go`
- [ ] Task: Write tests for `RequirePermission` RBAC middleware
- [ ] Task: Implement `RequirePermission` middleware
- [ ] Task: Conductor - User Manual Verification 'Middleware & Security Integration' (Protocol in workflow.md)
