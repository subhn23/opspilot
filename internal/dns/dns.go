package dns

import (
	"context"
	"fmt"
	"log"
)

// SSHClient abstracts remote command execution
type SSHClient interface {
	RunCommand(ctx context.Context, addr, command string) (string, error)
}

type DNSManager struct {
	ServerAddr string
	ZoneName   string
	SSH        SSHClient
}

// NewDNSManager creates a new DNSManager
func NewDNSManager(serverAddr, zoneName string, ssh SSHClient) *DNSManager {
	return &DNSManager{
		ServerAddr: serverAddr,
		ZoneName:   zoneName,
		SSH:        ssh,
	}
}

// UpdateRecordA adds or updates an A record in Windows DNS via PowerShell over SSH
func (m *DNSManager) UpdateRecordA(ctx context.Context, hostname string, ip string) error {
	log.Printf("DNS: Updating %s.%s -> %s on %s", hostname, m.ZoneName, ip, m.ServerAddr)

	// PowerShell command to add/update record
	// We use -AllowUpdateAny to handle existing records if needed, or check existence first.
	// For simplicity, we'll try to add it.
	psCommand := fmt.Sprintf("Add-DnsServerResourceRecordA -Name '%s' -ZoneName '%s' -IPv4Address '%s' -AllowUpdateAny", 
		hostname, m.ZoneName, ip)

	output, err := m.SSH.RunCommand(ctx, m.ServerAddr, psCommand)
	if err != nil {
		return fmt.Errorf("failed to update DNS record via SSH: %w (Output: %s)", err, output)
	}

	log.Printf("DNS: Successfully updated record for %s", hostname)
	return nil
}

// GetManualInstructions returns the A record details for the user to add manually
func (m *DNSManager) GetManualInstructions(hostname string, ip string) string {
	return fmt.Sprintf("Please add the following A Record to Windows DNS:\nName: %s\nZone: %s\nIP: %s",
		hostname, m.ZoneName, ip)
}
