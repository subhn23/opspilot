package config

import (
	"os"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestInitDBWithSQLite(t *testing.T) {
	// Test with SQLite dialector
	db := InitDB(sqlite.Open(":memory:"))
	if db == nil {
		t.Fatal("InitDB returned nil")
	}

	if DB == nil {
		t.Fatal("Global DB variable was not set")
	}

	// Verify tables exist
	tables := []string{"users", "roles", "permissions"}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			t.Errorf("Table %s was not created by InitDB", table)
		}
	}
}

func TestGetDialector(t *testing.T) {
	os.Setenv("DB_HOST", "myhost")
	defer os.Unsetenv("DB_HOST")

	dialector := GetDialector()
	if dialector == nil {
		t.Error("GetDialector returned nil")
	}

	if dialector.Name() != "postgres" {
		t.Errorf("Expected postgres dialector, got %s", dialector.Name())
	}
}

func TestAutoMigrate(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open sqlite: %v", err)
	}

	// The target function
	err = AutoMigrate(db)
	if err != nil {
		t.Errorf("AutoMigrate failed: %v", err)
	}

	// Verify tables exist
	tables := []string{"users", "roles", "permissions", "certificates", "proxy_routes", "environments", "deployments", "audit_logs"}
	for _, table := range tables {
		if !db.Migrator().HasTable(table) {
			t.Errorf("Table %s was not created", table)
		}
	}
}

func TestGetEnv(t *testing.T) {
	os.Setenv("TEST_KEY", "test_value")
	defer os.Unsetenv("TEST_KEY")

	val := getEnv("TEST_KEY", "fallback")
	if val != "test_value" {
		t.Errorf("Expected test_value, got %s", val)
	}

	val = getEnv("NON_EXISTENT_KEY", "fallback")
	if val != "fallback" {
		t.Errorf("Expected fallback, got %s", val)
	}
}
