package dns

import (
	"context"
	"fmt"
	"net"
	"testing"
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

func TestVerifyDNS(t *testing.T) {
	ctx := context.Background()

	t.Run("Match", func(t *testing.T) {
		mockRes := &MockResolver{MockIPs: []net.IP{net.ParseIP("10.0.0.1")}}
		mgr := NewDNSManager("dns.local", "opspilot.local", nil)
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
		mgr := NewDNSManager("dns.local", "opspilot.local", nil)
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
		mgr := NewDNSManager("dns.local", "opspilot.local", nil)
		mgr.Resolver = mockRes

		_, err := mgr.VerifyDNS(ctx, "app", "10.0.0.1")
		if err == nil {
			t.Error("Expected error, got nil")
		}
	})
}

func TestUpdateRecordA(t *testing.T) {
	mockSSH := &MockSSHClient{MockOutput: "Success"}
	mgr := NewDNSManager("dns.local", "opspilot.local", mockSSH)

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
	mockSSH := &MockSSHClient{MockErr: fmt.Errorf("connection refused"), MockOutput: "SSH Error"}
	mgr := NewDNSManager("dns.local", "opspilot.local", mockSSH)

	err := mgr.UpdateRecordA(context.Background(), "app1", "192.168.1.100")

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if err.Error() != "failed to update DNS record via SSH: connection refused (Output: SSH Error)" {
		t.Errorf("Unexpected error message: %v", err.Error())
	}
}
