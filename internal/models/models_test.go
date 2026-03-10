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
	if roleCount != 1 {
		t.Errorf("Expected 1 role, got %d", roleCount)
	}

	var permCount int64
	db.Model(&Permission{}).Count(&permCount)
	if permCount != 6 {
		t.Errorf("Expected 6 permissions, got %d", permCount)
	}

	// Second seed (should be idempotent)
	SeedSystemData(db)
	db.Model(&Role{}).Count(&roleCount)
	if roleCount != 1 {
		t.Errorf("Expected 1 role after second seed, got %d", roleCount)
	}
}
