package dns

import (
	"context"
	"fmt"
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
