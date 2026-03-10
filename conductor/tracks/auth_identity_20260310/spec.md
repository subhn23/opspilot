# Track 1.1: Authentication & Identity Core Spec

## Goal
Implement the foundation for user authentication and authorization in OpsPilot.

## Components

### 1. User and Role Models
- **User Model (GORM):**
  - `ID`: uuid.UUID (primaryKey)
  - `Email`: string (uniqueIndex)
  - `PasswordHash`: string
  - `TOTPSecret`: string
  - `RoleID`: uuid.UUID
  - `CreatedAt`: time.Time
- **Role Model:**
  - `ID`: uuid.UUID (primaryKey)
  - `Name`: string (unique, e.g., Admin, Developer, Viewer)
  - `Permissions`: []Permission (many2many:role_permissions)

### 2. Authentication Logic
- **JWT Provider:** Logic to generate and validate JWTs with claims (User ID, Role ID).
- **TOTP Engine:** Logic to generate a new TOTP secret (for QR enrollment) and verify current passcodes.
- **Login Flow:** User provides email/password; if correct, check for TOTP. If TOTP is enabled and verified, issue JWT.

### 3. Middleware
- **AuthMiddleware:** Validates the JWT in the `Authorization` header or cookie and populates the request context with the user identity.
- **RBAC Middleware:** Checks if the authenticated user has the required permission for the specific route.

## Success Criteria
- [ ] Users can successfully authenticate with email/password and (if enabled) TOTP.
- [ ] JWT tokens are correctly issued and validated.
- [ ] Routes are protected by `AuthMiddleware`.
- [ ] RBAC correctly restricts access based on role permissions.
- [ ] New code has >80% unit test coverage.
