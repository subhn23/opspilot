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
	db.AutoMigrate(&models.Deployment{}, &models.AuditLog{})
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

type MockSSHClient struct {
	CommandsRun []string
	MockOutput  string
	MockErr     error
}

func (m *MockSSHClient) RunCommand(ctx context.Context, addr, command string) (string, error) {
	m.CommandsRun = append(m.CommandsRun, command)
	return m.MockOutput, m.MockErr
}

func TestRemoteUp(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	mockSSH := &MockSSHClient{MockOutput: "Done"}
	deployer := &Deployer{DB: db, SSH: mockSSH}

	deploy := &models.Deployment{
		CommitHash: "abc1234",
	}
	db.Create(deploy)

	err := deployer.RemoteUp(ctx, deploy, "10.0.0.50")

	if err != nil {
		t.Fatalf("RemoteUp failed: %v", err)
	}

	if len(mockSSH.CommandsRun) != 2 {
		t.Errorf("Expected 2 commands, ran %d", len(mockSSH.CommandsRun))
	}

	var updated models.Deployment
	db.First(&updated, deploy.ID)
	if updated.Status != "SUCCESS" {
		t.Errorf("Expected status SUCCESS, got %s", updated.Status)
	}

	// Verify Audit Log
	var auditEntry models.AuditLog
	err = db.Where("action = ?", "DEPLOY_SUCCESS").First(&auditEntry).Error
	if err != nil {
		t.Errorf("Failed to find audit log entry: %v", err)
	}
}

func TestBuildAndPush(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	mockScanner := &MockScanner{Safe: true, Report: "All good"}
	deployer := &Deployer{DB: db, Scanner: mockScanner}

	deploy := &models.Deployment{
		CommitHash: "feat123",
	}
	db.Create(deploy)

	err := deployer.BuildAndPush(ctx, deploy)

	if err != nil {
		t.Fatalf("BuildAndPush failed: %v", err)
	}

	var updated models.Deployment
	db.First(&updated, deploy.ID)
	if updated.Status != "PUSHED" {
		t.Errorf("Expected status PUSHED, got %s", updated.Status)
	}
}
