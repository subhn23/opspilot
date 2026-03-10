# Plan for Track 2.3: Windows DNS Integration

## Phase 1: DNS Update Engine [checkpoint: 3787ff3]
**Goal:** Implement the SSH-based PowerShell update logic.

- [x] Task: Define `DNSManager` struct and `SSHClient` dependency in `internal/dns/dns.go` (Already complete)
- [x] Task: Implement `UpdateDNS` logic using PowerShell command generation (3300974)
- [x] Task: Write tests for `UpdateDNS` with a mock SSH client (3300974)
- [x] Task: Conductor - User Manual Verification 'DNS Update Engine' (Protocol in workflow.md)

## Phase 2: Verification & Fallback [checkpoint: b5318e7]
**Goal:** Implement DNS lookup verification and manual fallback.

- [x] Task: Implement `VerifyDNS` using `net.LookupIP` (d9fe993)
- [x] Task: Implement `RequestManualDNS` blocking logic (mocked UI trigger) (d9fe993)
- [x] Task: Write tests for `VerifyDNS` (d9fe993)
- [x] Task: Conductor - User Manual Verification 'Verification & Fallback' (Protocol in workflow.md)

## Phase 3: Integration & Logging [checkpoint: 1e0bbbd]
**Goal:** Connect to the audit log and finalize integration.

- [x] Task: Add audit logging to `UpdateDNS` and `VerifyDNS` (Already complete)
- [x] Task: Verify end-to-end flow with mocked dependencies (46f55bb)
- [x] Task: Conductor - User Manual Verification 'Integration & Logging' (Protocol in workflow.md)

## Phase: Review Fixes
- [x] Task: Apply review suggestions a4b4dec
