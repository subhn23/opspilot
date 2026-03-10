# Phase 1: Foundation (Core Control Plane)

This phase focuses on setting up the base OpsPilot application, the high-availability PostgreSQL layer, and the native reverse proxy.

## Track 1: Base Application & UI Scaffold
**Goal:** Establish the Go + Gin + Tailwind + HTMX foundation.

### Data Structures
```go
type User struct {
    ID        uuid.UUID `gorm:"primaryKey"`
    Email     string    `gorm:"uniqueIndex"`
    Secret    string    // TOTP Secret
    RoleID    uuid.UUID
    CreatedAt time.Time
}

type Role struct {
    ID          uuid.UUID `gorm:"primaryKey"`
    Name        string    `gorm:"unique"` // Admin, Developer, Viewer
    Permissions []Permission `gorm:"many2many:role_permissions;"`
}
```

### Key Functions
- `SetupRouter() *gin.Engine`: Configures Gin with Tailwind/HTMX templates.
- `InitDB() *gorm.DB`: Connects to PostgreSQL with HA connection string.

---

## Track 2: OpsProxy & SSL Management
**Goal:** Native L7 routing and manual SSL certificate versioning.

### Data Structures
```go
type Certificate struct {
    ID           uint   `gorm:"primaryKey"`
    Label        string `gorm:"uniqueIndex"`
    FullChain    string
    PrivateKey   string
    IsProduction bool
}

type ProxyRoute struct {
    ID          uint   `gorm:"primaryKey"`
    Domain      string `gorm:"uniqueIndex"`
    TargetURL   string
    Protocol    string // HTTP, gRPC
    IsActive    bool
}
```

### Key Functions
- `(p *OpsProxy) Start()`: Launches the HTTPS server.
- `(p *OpsProxy) GetCertificate(hello *tls.ClientHelloInfo) (*tls.Certificate, error)`: Logic for Test-then-Deploy.
- `(p *OpsProxy) ReloadTLSConfig()`: Hot-reloads certs from DB without restart.

---

## Track 3: Audit & Identity
**Goal:** TOTP Authentication and Action Logging.

### Key Functions
- `VerifyTOTP(passcode string, secret string) bool`: Validates MFA.
- `LogAction(c *gin.Context, action string, target string)`: Records mutation to `audit_logs` table.
