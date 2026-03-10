# Track 2.3: Windows DNS Integration Spec

## Goal
Implement a DNS module that automates the update of Windows DNS records by executing PowerShell commands over SSH, providing a bridge between the Go-based control plane and the Windows-managed DNS infrastructure.

## Components

### 1. DNS Update Engine
- **SSH PowerShell Execution:**
  - Connects to a Windows Jump Host or DNS Server via SSH.
  - Executes `Add-DnsServerResourceRecordA` or similar PowerShell cmdlets.
- **`UpdateDNS` Function:**
  - Signature: `UpdateDNS(domain string, ip string) error`
  - Handles command construction and error parsing from PowerShell output.

### 2. Verification & Manual Fallback
- **`VerifyDNS` Function:**
  - Performs a DNS lookup (using `net.LookupIP`) to confirm the record has propagated.
- **`RequestManualDNS` Function:**
  - If automated update is not possible or fails, blocks the UI flow and prompts the user to manually update the record.
  - Continues once verification succeeds.

### 3. Integration
- **Audit Logging:** Logs all DNS change attempts (success and failure).

## Success Criteria
- [ ] `UpdateDNS` correctly executes PowerShell commands via a mocked SSH client.
- [ ] `VerifyDNS` correctly identifies when a record matches the expected IP.
- [ ] DNS actions are recorded in the `AuditLog`.
- [ ] Code has >80% unit test coverage using interfaces for SSH.
