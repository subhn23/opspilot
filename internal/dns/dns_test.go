package dns

import (
	"context"
	"fmt"
	"net"
	"opspilot/internal/models"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type MockSSHClient struct {
	LastCommand string
	MockOutput  string
	MockErr     error
}

func (m *MockSSHClient) RunCommand(ctx context.Context, addr, command string) (string, error) {
	m.LastCommand = command
	return m.MockOutput, m.MockErr
}

type MockResolver struct {
	MockIPs []net.IP
	MockErr error
}

func (m *MockResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return m.MockIPs, m.MockErr
}

func setupTestDB() *gorm.DB {
	db, _ := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	db.AutoMigrate(&models.AuditLog{})
	return db
}

func TestVerifyDNS(t *testing.T) {
	ctx := context.Background()
	db := setupTestDB()

	t.Run("Match", func(t *testing.T) {
		mockRes := &MockResolver{MockIPs: []net.IP{net.ParseIP("10.0.0.1")}}
		mgr := NewDNSManager(db, "dns.local", "opspilot.local", nil)
		mgr.Resolver = mockRes

		verified, err := mgr.VerifyDNS(ctx, "app", "10.0.0.1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if !verified {
			t.Error("Expected verification to pass")
		}
	})

	t.Run("No Match", func(t *testing.T) {
		mockRes := &MockResolver{MockIPs: []net.IP{net.ParseIP("10.0.0.2")}}
		mgr := NewDNSManager(db, "dns.local", "opspilot.local", nil)
		mgr.Resolver = mockRes

		verified, err := mgr.VerifyDNS(ctx, "app", "10.0.0.1")
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		}
		if verified {
			t.Error("Expected verification to fail")
		}
	})

	t.Run("Error", func(t *testing.T) {
		mockRes := &MockResolver{MockErr: fmt.Errorf("timeout")}
		mgr := NewDNSManager(db, "dns.local", "opspilot.local", nil)
		mgr.Resolver = mockRes

		_, err := mgr.VerifyDNS(ctx, "app", "10.0.0.1")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestUpdateRecordA(t *testing.T) {
	db := setupTestDB()
	mockSSH := &MockSSHClient{MockOutput: "Success"}
	mgr := NewDNSManager(db, "dns.local", "opspilot.local", mockSSH)

	err := mgr.UpdateRecordA(context.Background(), "app1", "192.168.1.100")

	if err != nil {
		t.Fatalf("UpdateRecordA failed: %v", err)
	}

	expectedCmd := "Add-DnsServerResourceRecordA -Name 'app1' -ZoneName 'opspilot.local' -IPv4Address '192.168.1.100' -AllowUpdateAny"
	if mockSSH.LastCommand != expectedCmd {
		t.Errorf("Expected command:\n%s\nGot:\n%s", expectedCmd, mockSSH.LastCommand)
	}
}

func TestUpdateRecordAFailure(t *testing.T) {
	db := setupTestDB()
	mockSSH := &MockSSHClient{MockErr: fmt.Errorf("connection refused"), MockOutput: "SSH Error"}
	mgr := NewDNSManager(db, "dns.local", "opspilot.local", mockSSH)

	err := mgr.UpdateRecordA(context.Background(), "app1", "192.168.1.100")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "failed to update DNS record via SSH: connection refused (Output: SSH Error)" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}

func TestEndToEndFlow(t *testing.T) {
	db := setupTestDB()
	ctx := context.Background()
	mockSSH := &MockSSHClient{MockOutput: "Success"}
	mockRes := &MockResolver{MockIPs: []net.IP{net.ParseIP("10.0.0.100")}}

	mgr := NewDNSManager(db, "dns.local", "opspilot.local", mockSSH)
	mgr.Resolver = mockRes

	// 1. Update
	err := mgr.UpdateRecordA(ctx, "test-app", "10.0.0.100")
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	// 2. Verify
	verified, err := mgr.VerifyDNS(ctx, "test-app", "10.0.0.100")
	if err != nil || !verified {
		t.Fatalf("Verification failed: %v (verified=%v)", err, verified)
	}

	// 3. Check Audit Logs
	var logs []models.AuditLog
	db.Find(&logs)
	if len(logs) != 2 {
		t.Errorf("Expected 2 audit logs, got %d", len(logs))
	}
}
