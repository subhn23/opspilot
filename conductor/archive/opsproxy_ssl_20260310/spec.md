# Track 1.2: OpsProxy & SSL Management Spec

## Goal
Implement a native Go-based L7 reverse proxy with dynamic SSL certificate management and hot-reloading capabilities.

## Components

### 1. L7 Routing Engine
- **Dynamic Routing:** Routes traffic based on the `Host` header using rules stored in the database (`ProxyRoute` model).
- **Protocol Support:** Initial support for HTTP/HTTPS proxying to backend services (Docker containers or VMs).
- **Health Checks:** (Future) Basic check if target is reachable.

### 2. SSL Management ("Test-then-Deploy")
- **Certificate Storage:** Stores full chains and private keys in the database (`Certificate` model).
- **Dynamic SNI:** Implements `GetCertificate` to serve the correct certificate based on the incoming Server Name Indication (SNI).
- **Versioning/Overrides:** Supports `CertTestOverride` to allow testing a new certificate on a specific domain before promoting it to global production status.

### 3. Hot-Reloading
- **In-Memory Cache:** (Optional/Future) Cache active routes and certs for performance.
- **On-Demand Loading:** The current implementation in `internal/proxy/proxy.go` loads from the DB per request/handshake, which effectively provides hot-reloading.

## Success Criteria
- [ ] OpsProxy can successfully route traffic to a backend service based on domain.
- [ ] SSL handshake succeeds using certificates loaded from the database.
- [ ] `CertTestOverride` correctly prioritizes a test certificate over the global production one.
- [ ] OpsProxy handles missing routes with a 401/404 response.
- [ ] Code has >80% unit test coverage.
