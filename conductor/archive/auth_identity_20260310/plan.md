# Plan for Track 1.1: Authentication & Identity Core

## Phase 1: Data Models & Persistence [checkpoint: 39c37f3]
**Goal:** Define and migrate the database models for users, roles, and permissions.

- [x] Task: Create User and Role models in `internal/models/models.go` (7d38b69)
- [x] Task: Implement GORM database migration logic (ec0b604)
- [x] Task: Seed initial roles (Admin, Developer, Viewer) and permissions (6326e23)
- [x] Task: Conductor - User Manual Verification 'Data Models & Persistence' (Protocol in workflow.md)

## Phase 2: Core Authentication Logic [checkpoint: 5646273]
**Goal:** Implement JWT generation/validation and TOTP verification.

- [x] Task: Write tests for JWT provider logic (996bb8a)
- [x] Task: Implement JWT provider in `internal/auth/auth.go` (6caac10)
- [x] Task: Write tests for TOTP enrollment and verification (cc30881)
- [x] Task: Implement TOTP logic (using a standard library like `pquerna/otp`) (639ce6a)
- [x] Task: Conductor - User Manual Verification 'Core Authentication Logic' (Protocol in workflow.md)

## Phase 3: Middleware & Security Integration [checkpoint: a239c16]
**Goal:** Protect routes and enforce RBAC.

- [x] Task: Write tests for `AuthMiddleware` (81d93b9)
- [x] Task: Implement `AuthMiddleware` in `internal/auth/auth.go` (6ca60eb)
- [x] Task: Write tests for `RequirePermission` RBAC middleware (b3d0289)
- [x] Task: Implement `RequirePermission` middleware (4c8b082)
- [x] Task: Conductor - User Manual Verification 'Middleware & Security Integration' (Protocol in workflow.md)

## Phase: Review Fixes
- [x] Task: Apply review suggestions a133c71
