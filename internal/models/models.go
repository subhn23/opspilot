package models

import (
	"opspilot/internal/events"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// User represents a platform administrator or developer
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	TOTPSecret   string    `gorm:"not null"` // TOTP Secret
	RoleID       uuid.UUID `gorm:"type:uuid;not null"`
	Role         Role      `gorm:"foreignKey:RoleID"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

// Role represents a group of permissions (Admin, Developer, etc.)
type Role struct {
	ID          uuid.UUID    `gorm:"type:uuid;primaryKey"`
	Name        string       `gorm:"uniqueIndex;not null"`
	Permissions []Permission `gorm:"many2many:role_permissions;"`
	CreatedAt   time.Time
}

// Permission defines a specific action allowed in a module
type Permission struct {
	ID     uint   `gorm:"primaryKey"`
	Slug   string `gorm:"uniqueIndex;not null"` // e.g., "proxy:write", "deploy:prod"
	Module string `gorm:"not null"`             // e.g., "OpsProxy", "OpsDeploy"
}

// Certificate stores vendor wildcard or manual SSL certs
type Certificate struct {
	ID           uint   `gorm:"primaryKey"`
	Label        string `gorm:"uniqueIndex;not null"`
	FullChain    string `gorm:"type:text;not null"`
	PrivateKey   string `gorm:"type:text;not null"`
	IsProduction bool   `gorm:"default:false"`
	CreatedAt    time.Time
}

// ProxyRoute defines Layer 7 routing rules
type ProxyRoute struct {
	ID        uint   `gorm:"primaryKey"`
	Domain    string `gorm:"uniqueIndex;not null"`
	TargetURL string `gorm:"not null"`
	Protocol  string `gorm:"default:'HTTP'"` // HTTP, gRPC
	IsActive  bool   `gorm:"default:true"`
	CreatedAt time.Time
}

// Test Overrides (Test before Global deployment)
type CertTestOverride struct {
	Domain    string `gorm:"primaryKey"` // e.g., "test-ssl.yourdomain.com"
	CertID    uint   `gorm:"not null"`
	CreatedAt time.Time
}

// Environment represents a dynamically provisioned VM environment
type Environment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	Name      string    `gorm:"uniqueIndex;not null"` // e.g., "staging-auth"
	Type      string    `gorm:"not null"`             // prod, staging, dev
	HostNode  string    `gorm:"not null"`             // host1 or host2
	VMID      int       `gorm:"uniqueIndex"`
	IPAddress string
	TTL       *time.Time // Expiry for Dev environments
	Status    string     `gorm:"default:'PROVISIONING'"` // Provisioning, Healthy, Failed, Destroyed
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Deployment tracks the history of code delivery to environments
type Deployment struct {
	ID            uint      `gorm:"primaryKey"`
	EnvironmentID uuid.UUID `gorm:"type:uuid"`
	Environment   Environment
	CommitHash    string `gorm:"not null"`
	Branch        string `gorm:"not null"`
	ContainerID   string // Real Docker container ID
	Status        string // Building, Pushing, Deploying, Success, Failed
	Logs          string `gorm:"type:text"`
	DeployedBy    string
	DeployedAt    time.Time
}

// AuditLog captures every mutating action for compliance
type AuditLog struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid"`
	Action    string    `gorm:"not null"`  // e.g., "DEPLOY", "DELETE_VM"
	Target    string    `gorm:"not null"`  // e.g., "VM-Prod-01"
	Payload   string    `gorm:"type:text"` // JSON payload
	IPAddress string
	CreatedAt time.Time
}

// Node represents a visual element in the topology map
type Node struct {
	ID       string            `json:"id"`
	Label    string            `json:"label"`
	Type     string            `json:"type"`   // Firewall, VM, Container
	Status   string            `json:"status"` // Green, Red, Yellow
	Metadata map[string]string `json:"metadata"`
}

// Edge represents a connection between nodes in the topology map
type Edge struct {
	Source string `json:"source"`
	Target string `json:"target"`
	Label  string `json:"label"` // HTTP, gRPC, DB
}

// BeforeCreate hook to generate UUIDs
func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return
}

func (r *Role) BeforeCreate(tx *gorm.DB) (err error) {
	if r.ID == uuid.Nil {
		r.ID = uuid.New()
	}
	return
}

func (e *Environment) BeforeCreate(tx *gorm.DB) (err error) {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return
}

func (e *Environment) AfterSave(tx *gorm.DB) (err error) {
	events.Notify()
	return
}

func (d *Deployment) AfterSave(tx *gorm.DB) (err error) {
	events.Notify()
	return
}
