# Track 1.3: Audit & Action Logging Spec

## Goal
Implement a centralized, immutable audit logging system to record all user-driven mutations and system events in OpsPilot.

## Components

### 1. Audit Log Model
- **AuditLog Model (GORM):**
  - `ID`: uint (primaryKey)
  - `UserID`: uuid.UUID (index)
  - `Action`: string (e.g., "VM_CREATE", "PROXY_UPDATE")
  - `Target`: string (e.g., name of the VM or domain)
  - `Payload`: string (JSON representation of the change, optional)
  - `IPAddress`: string
  - `CreatedAt`: time.Time (index)

### 2. Logging Functionality
- **`LogAction` Function:** A utility function to record events to the database.
  - Signature: `LogAction(db *gorm.DB, userID uuid.UUID, action string, target string, ip string, payload string)`
- **Middleware Integration:** (Future) Hook into Gin routes to automatically log certain actions.

## Success Criteria
- [ ] `AuditLog` table is correctly migrated in the database.
- [ ] `LogAction` correctly persists events to the database.
- [ ] All required fields (User, Action, Target, IP, Timestamp) are captured.
- [ ] Unit tests verify the persistence and retrieval of logs.
- [ ] Code has >80% unit test coverage.
