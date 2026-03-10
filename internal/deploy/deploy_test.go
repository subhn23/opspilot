package deploy

import (
	"context"
	"opspilot/internal/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockScanner struct {
	Safe   bool
	Report string
	Err    error
}

func (m *MockScanner) Scan(ctx context.Context, imageName string) (bool, string, error) {
	return m.Safe, m.Report, m.Err
}

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.Deployment{})
	return db
}

func TestScanImage(t *testing.T) {
	ctx := context.Background()
	
	t.Run("Safe Image", func(t *testing.T) {
		mock := &MockScanner{Safe: true, Report: "Clean"}
		deployer := &Deployer{Scanner: mock}
		
		safe, report, err := deployer.ScanImage(ctx, "test-image")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !safe {
			t.Error("Expected image to be safe")
		}
		if report != "Clean" {
			t.Errorf("Expected report 'Clean', got %s", report)
		}
	})

	t.Run("Unsafe Image", func(t *testing.T) {
		mock := &MockScanner{Safe: false, Report: "Vulnerability Found"}
		deployer := &Deployer{Scanner: mock}
		
		safe, report, err := deployer.ScanImage(ctx, "unsafe-image")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if safe {
			t.Error("Expected image to be unsafe")
		}
		if report != "Vulnerability Found" {
			t.Errorf("Expected report 'Vulnerability Found', got %s", report)
		}
	})
}
