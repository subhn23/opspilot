# Plan for Track 1.2: OpsProxy & SSL Management

## Phase 1: L7 Routing & Core Proxy [checkpoint: 5607663]
**Goal:** Implement and test the basic reverse proxying logic based on database routes.

- [x] Task: Write tests for `ServeHTTP` routing logic (34aea3f)
- [x] Task: Refactor/Enhance `ServeHTTP` in `internal/proxy/proxy.go` for better error handling (f2bfd55)
- [x] Task: Conductor - User Manual Verification 'L7 Routing & Core Proxy' (Protocol in workflow.md)

## Phase 2: Dynamic SSL Management [checkpoint: 9de2d23]
**Goal:** Implement and test dynamic certificate loading and SNI support.

- [x] Task: Write tests for `GetCertificate` logic (including overrides) (7bf987e)
- [x] Task: Refactor/Enhance `GetCertificate` and `parseCert` in `internal/proxy/proxy.go` (7786165)
- [x] Task: Conductor - User Manual Verification 'Dynamic SSL Management' (Protocol in workflow.md)

## Phase 3: Integration ## Phase 3: Integration & Resilience [checkpoint: ] Resilience [checkpoint: 83993b5]
**Goal:** Ensure the proxy is robust and integrates well with the rest of the system.

- [x] Task: Add logging and basic health check support to `OpsProxy` (0ed1357)
- [x] Task: Verify hot-reloading by updating DB routes/certs during operation (bac149e)
- [x] Task: Conductor - User Manual Verification 'Integration - [ ] Task: Conductor - User Manual Verification 'Integration & Resilience' (Protocol in workflow.md) Resilience' (Protocol in workflow.md)
