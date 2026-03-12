package models

import (
	"reflect"
	"testing"

	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestUserFields(t *testing.T) {
	u := User{}
	val := reflect.ValueOf(u)
	typ := val.Type()

	requiredFields := []string{"PasswordHash", "TOTPSecret"}
	for _, field := range requiredFields {
		_, found := typ.FieldByName(field)
		if !found {
			t.Errorf("User struct is missing required field: %s", field)
		}
	}
}

func TestUserBeforeCreate(t *testing.T) {
	u := User{}
	err := u.BeforeCreate(nil)
	if err != nil {
		t.Errorf("BeforeCreate failed: %v", err)
	}
	if u.ID == uuid.Nil {
		t.Error("User ID was not generated")
	}
}

func TestRoleBeforeCreate(t *testing.T) {
	r := Role{}
	err := r.BeforeCreate(nil)
	if err != nil {
		t.Errorf("BeforeCreate failed: %v", err)
	}
	if r.ID == uuid.Nil {
		t.Error("Role ID was not generated")
	}
}

func TestEnvironmentBeforeCreate(t *testing.T) {
	e := Environment{}
	err := e.BeforeCreate(nil)
	if err != nil {
		t.Errorf("BeforeCreate failed: %v", err)
	}
	if e.ID == uuid.Nil {
		t.Error("Environment ID was not generated")
	}
}

func TestSeedSystemData(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}

	err = db.AutoMigrate(&Permission{}, &Role{})
	if err != nil {
		t.Fatalf("Failed to migrate: %v", err)
	}

	// First seed
	SeedSystemData(db)

	var roleCount int64
	db.Model(&Role{}).Count(&roleCount)
	if roleCount != 3 {
		t.Errorf("Expected 3 roles (Admin, Developer, Viewer), got %d", roleCount)
	}

	roles := []string{"Master Admin", "Developer", "Viewer"}
	for _, name := range roles {
		var r Role
		if err := db.Where("name = ?", name).First(&r).Error; err != nil {
			t.Errorf("Role %s not found", name)
		}
	}

	var permCount int64
	db.Model(&Permission{}).Count(&permCount)
	if permCount < 6 {
		t.Errorf("Expected at least 6 permissions, got %d", permCount)
	}

	// Second seed (should be idempotent)
	SeedSystemData(db)
	db.Model(&Role{}).Count(&roleCount)
	if roleCount != 3 {
		t.Errorf("Expected 3 roles after second seed, got %d", roleCount)
	}
}

func TestTargetHostModel(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&TargetHost{})

	host := TargetHost{
		Name:     "Test Host",
		Type:     "remote_ssh",
		Endpoint: "1.2.3.4",
		AuthData: "encrypted-key",
	}

	if err := db.Create(&host).Error; err != nil {
		t.Fatalf("Failed to create TargetHost: %v", err)
	}

	if host.ID == uuid.Nil {
		t.Error("TargetHost ID was not generated")
	}

	var saved TargetHost
	db.First(&saved, host.ID)
	if saved.Name != "Test Host" {
		t.Errorf("Expected name 'Test Host', got %s", saved.Name)
	}
}

func TestEnvironmentTargetHostLink(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&TargetHost{}, &Environment{})

	host := TargetHost{Name: "Main Server", Type: "local_proxmox"}
	db.Create(&host)

	env := Environment{
		Name:         "Dev Env",
		Type:         "dev",
		TargetHostID: &host.ID,
	}

	if err := db.Create(&env).Error; err != nil {
		t.Fatalf("Failed to create Environment with TargetHostID: %v", err)
	}

	var savedEnv Environment
	db.Preload("TargetHost").First(&savedEnv, env.ID)
	if savedEnv.TargetHostID == nil || *savedEnv.TargetHostID != host.ID {
		t.Error("TargetHostID was not saved correctly")
	}
	if savedEnv.TargetHost.Name != "Main Server" {
		t.Errorf("Expected associated host name 'Main Server', got %s", savedEnv.TargetHost.Name)
	}
}

func TestHostNodeMigration(t *testing.T) {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&TargetHost{}, &Environment{})

	// Create legacy environments
	env1 := Environment{Name: "Legacy 1", HostNode: "host1", VMID: 1001}
	env2 := Environment{Name: "Legacy 2", HostNode: "host2", VMID: 1002}
	env3 := Environment{Name: "Legacy 3", HostNode: "host1", VMID: 1003}
	db.Create(&env1)
	db.Create(&env2)
	db.Create(&env3)

	// Run migration
	MigrateHostNodesToTargetHosts(db)

	// Verify TargetHosts were created
	var hosts []TargetHost
	db.Find(&hosts)
	if len(hosts) != 2 {
		t.Errorf("Expected 2 TargetHosts (host1, host2), got %d", len(hosts))
	}

	// Verify Environments are linked
	var updatedEnv1 Environment
	db.Preload("TargetHost").First(&updatedEnv1, env1.ID)
	if updatedEnv1.TargetHostID == nil || updatedEnv1.TargetHost.Name != "host1" {
		t.Error("env1 was not linked correctly to host1")
	}

	var updatedEnv2 Environment
	db.Preload("TargetHost").First(&updatedEnv2, env2.ID)
	if updatedEnv2.TargetHostID == nil || updatedEnv2.TargetHost.Name != "host2" {
		t.Error("env2 was not linked correctly to host2")
	}
}
