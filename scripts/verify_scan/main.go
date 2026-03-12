package main

import (
	"context"
	"fmt"
	"log"
	"opspilot/internal/config"
	"opspilot/internal/deploy"
	"opspilot/internal/models"

	"gorm.io/driver/sqlite"
)

// MockScanner for manual verification
type MockScanner struct {
	Safe   bool
	Report string
}

func (m *MockScanner) Scan(ctx context.Context, imageName string) (bool, string, error) {
	return m.Safe, m.Report, nil
}

func main() {
	// 1. Setup in-memory DB for verification
	db := config.InitDB(sqlite.Open(":memory:"))
	db.AutoMigrate(&models.Deployment{}, &models.AuditLog{})

	ctx := context.Background()

	t.Run("Verification: Safe Scan", func() {
		scanner := &MockScanner{Safe: true, Report: "No issues found"}
		deployer := deploy.NewDeployer(db)
		deployer.Scanner = scanner

		d := &models.Deployment{CommitHash: "safe123"}
		db.Create(d)

		fmt.Println("Attempting deployment with safe image...")
		err := deployer.BuildAndPush(ctx, d)
		if err != nil {
			log.Fatalf("Unexpected error: %v", err)
		}

		var updated models.Deployment
		db.First(&updated, d.ID)
		fmt.Printf("Deployment Status: %s\n", updated.Status)
		if updated.Status != "PUSHED" {
			log.Fatalf("Expected PUSHED, got %s", updated.Status)
		}
	})

	fmt.Println("---")

	t.Run("Verification: Unsafe Scan", func() {
		scanner := &MockScanner{Safe: false, Report: "CRITICAL: Log4Shell found"}
		deployer := deploy.NewDeployer(db)
		deployer.Scanner = scanner

		d := &models.Deployment{CommitHash: "unsafe456"}
		db.Create(d)

		fmt.Println("Attempting deployment with unsafe image...")
		err := deployer.BuildAndPush(ctx, d)
		if err == nil {
			log.Fatal("Expected error, but deployment proceeded")
		}
		fmt.Printf("Deployment blocked as expected: %v\n", err)

		var updated models.Deployment
		db.First(&updated, d.ID)
		fmt.Printf("Deployment Status: %s\n", updated.Status)
		if updated.Status != "FAILED_SECURITY" {
			log.Fatalf("Expected FAILED_SECURITY, got %s", updated.Status)
		}

		// Verify Audit Log
		var auditEntry models.AuditLog
		err = db.Where("action = ?", "SECURITY_FAILURE").First(&auditEntry).Error
		if err != nil {
			log.Fatalf("Failed to find audit log entry for SECURITY_FAILURE: %v", err)
		}
		fmt.Printf("Audit Log Found: Action=%s, Target=%s, Payload=%s\n", 
			auditEntry.Action, auditEntry.Target, auditEntry.Payload)
	})

	fmt.Println("\nManual verification of Security Scanning logic PASSED.")
}

// Minimal t.Run helper for script
type T struct{}
func (t *T) Run(name string, f func()) {
	fmt.Printf("Running %s...\n", name)
	f()
}
var t = &T{}
