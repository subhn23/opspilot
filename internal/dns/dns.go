package dns

import (
	"context"
	"fmt"
	"log"
	"net"
	"opspilot/internal/audit"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SSHClient abstracts remote command execution
type SSHClient interface {
	RunCommand(ctx context.Context, addr, command string) (string, error)
}

// Resolver abstracts DNS lookup logic
type Resolver interface {
	LookupIP(ctx context.Context, network, host string) ([]net.IP, error)
}

// DefaultResolver uses the standard net package
type DefaultResolver struct{}

func (r *DefaultResolver) LookupIP(ctx context.Context, network, host string) ([]net.IP, error) {
	return net.DefaultResolver.LookupIP(ctx, network, host)
}

type DNSManager struct {
	DB         *gorm.DB
	ServerAddr string
	ZoneName   string
	SSH        SSHClient
	Resolver   Resolver
}

// NewDNSManager creates a new DNSManager
func NewDNSManager(db *gorm.DB, serverAddr, zoneName string, ssh SSHClient) *DNSManager {
	return &DNSManager{
		DB:         db,
		ServerAddr: serverAddr,
		ZoneName:   zoneName,
		SSH:        ssh,
		Resolver:   &DefaultResolver{},
	}
}

// UpdateRecordA adds or updates an A record in Windows DNS via PowerShell over SSH
func (m *DNSManager) UpdateRecordA(ctx context.Context, hostname string, ip string) error {
	log.Printf("DNS: Updating %s.%s -> %s on %s", hostname, m.ZoneName, ip, m.ServerAddr)

	// PowerShell command to add/update record
	psCommand := fmt.Sprintf("Add-DnsServerResourceRecordA -Name '%s' -ZoneName '%s' -IPv4Address '%s' -AllowUpdateAny",
		hostname, m.ZoneName, ip)

	output, err := m.SSH.RunCommand(ctx, m.ServerAddr, psCommand)
	if err != nil {
		audit.LogAction(m.DB, uuid.Nil, "DNS_UPDATE_FAILURE", hostname, m.ServerAddr, fmt.Sprintf("Error: %v, Output: %s", err, output))
		return fmt.Errorf("failed to update DNS record via SSH: %w (Output: %s)", err, output)
	}

	audit.LogAction(m.DB, uuid.Nil, "DNS_UPDATE_SUCCESS", hostname, m.ServerAddr, "DNS record updated successfully via PowerShell")
	log.Printf("DNS: Successfully updated record for %s", hostname)
	return nil
}

// VerifyDNS performs a DNS lookup to confirm the record matches the expected IP
func (m *DNSManager) VerifyDNS(ctx context.Context, hostname string, expectedIP string) (bool, error) {
	fqdn := fmt.Sprintf("%s.%s", hostname, m.ZoneName)
	ips, err := m.Resolver.LookupIP(ctx, "ip", fqdn)
	if err != nil {
		return false, fmt.Errorf("lookup failed for %s: %w", fqdn, err)
	}

	for _, ip := range ips {
		if ip.String() == expectedIP {
			audit.LogAction(m.DB, uuid.Nil, "DNS_VERIFY_SUCCESS", hostname, fqdn, "DNS record verified")
			return true, nil
		}
	}

	audit.LogAction(m.DB, uuid.Nil, "DNS_VERIFY_FAILURE", hostname, fqdn, "DNS record mismatch")
	return false, nil
}

// GetManualInstructions returns the A record details for the user to add manually
func (m *DNSManager) GetManualInstructions(hostname string, ip string) string {
	return fmt.Sprintf("Please add the following A Record to Windows DNS:\nName: %s\nZone: %s\nIP: %s",
		hostname, m.ZoneName, ip)
}

// RequestManualDNS blocks until the DNS record is manually verified or timeout occurs
func (m *DNSManager) RequestManualDNS(ctx context.Context, hostname string, expectedIP string, timeout time.Duration) error {
	log.Printf("DNS: Automated update failed or requested manual fallback. Blocking for verification of %s", hostname)
	log.Println(m.GetManualInstructions(hostname, expectedIP))

	deadline := time.Now().Add(timeout)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if time.Now().After(deadline) {
				return fmt.Errorf("timeout waiting for manual DNS verification for %s", hostname)
			}

			verified, err := m.VerifyDNS(ctx, hostname, expectedIP)
			if err == nil && verified {
				log.Printf("DNS: Record %s manually verified successfully.", hostname)
				return nil
			}
		}
	}
}
