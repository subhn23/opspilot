# Plan for Track 2.3: Windows DNS Integration

## Phase 1: DNS Update Engine [checkpoint: 3787ff3]
**Goal:** Implement the SSH-based PowerShell update logic.

- [x] Task: Define `DNSManager` struct and `SSHClient` dependency in `internal/dns/dns.go` ( Already complete)
- [x] Task: Implement `UpdateDNS` logic using PowerShell command generation (3300974)
- [x] Task: Write tests for `UpdateDNS` with a mock SSH client (3300974)
- [x] Task: Conductor - User Manual Verification 'DNS Update Engine' (Protocol in workflow.md)

## Phase 2: Verification & Verification [checkpoint: ]
**Goal:** Implement DNS lookup verification and manual fallback.

- [ ] Task: Implement `VerifyDNS` using `net.LookupIP`
- [ ] Task: Implement `RequestManualDNS` blocking logic (mocked UI trigger)
- [ ] Task: Write tests for `VerifyDNS`
- [ ] Task: Conductor - User Manual Verification 'Verification & Fallback' (Protocol in workflow.md)

## Phase 3: Integration & Logging [checkpoint: ]
**Goal:** Connect to the audit log and finalize integration.

- [ ] Task: Add audit logging to `UpdateDNS` and `VerifyDNS`
- [ ] Task: Verify end-to-end flow with mocked dependencies
- [ ] Task: Conductor - User Manual Verification 'Integration & Logging' (Protocol in workflow.md)
