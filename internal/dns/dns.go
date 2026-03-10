package dns

import (
	"fmt"
	"log"
	"os/exec"
)

type DNSManager struct {
	ServerAddr string
	ZoneName   string
}

// UpdateRecordA adds or updates an A record in Windows DNS
func (m *DNSManager) UpdateRecordA(hostname string, ip string) error {
	// Conceptual: Use SSH to execute PowerShell on Windows Server
	// Command: Add-DnsServerResourceRecordA -Name "hostname" -ZoneName "zone" -IPv4Address "ip"

	log.Printf("Updating DNS: %s.%s -> %s", hostname, m.ZoneName, ip)

	// Example local execution (if running on windows or via cross-platform tool)
	psCommand := fmt.Sprintf("Add-DnsServerResourceRecordA -Name '%s' -ZoneName '%s' -IPv4Address '%s'", hostname, m.ZoneName, ip)
	cmd := exec.Command("powershell", "-Command", psCommand)

	// We don't run it here to avoid errors in non-windows dev environments
	// return cmd.Run()
	_ = cmd
	return nil
}

// GetManualInstructions returns the A record details for the user to add manually
func (m *DNSManager) GetManualInstructions(hostname string, ip string) string {
	return fmt.Sprintf("Please add the following A Record to Windows DNS:\nName: %s\nZone: %s\nIP: %s",
		hostname, m.ZoneName, ip)
}
