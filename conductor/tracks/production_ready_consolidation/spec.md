# Specification: Production-Ready Consolidation

## Goal
Finalize the "last-mile" integration tasks required to move OpsPilot from an architectural scaffold to a production-hardened platform. This track consolidates pending work across authentication, infrastructure orchestration, delivery pipelines, observability, and resilience.

## Scope
- **Authentication:** Finalize JWT/Session middleware and MFA enrollment UI.
- **Infrastructure:** Complete Terraform template mirroring and secure credential handling.
- **Deployment:** Implement Git integration, Registry authentication, and SSH-based remote deployment.
- **Network:** Finalize Windows DNS integration via SSH.
- **Observability:** Implement real Docker metrics collection and VictoriaMetrics integration.
- **Resilience:** Finalize Postgres WAL archiving and Registry autosync workers.
- **UI/UX:** Build the Environment Wizard, Audit Viewer, and Live Logs interface.

## Technical Requirements
- Use `internal/auth` for JWT and TOTP logic.
- Use `internal/terraform` and `hashicorp/tf-exec` for orchestration.
- Use `internal/deploy` for Git, Docker, and SSH deployment logic.
- Use `internal/metrics` and VictoriaMetrics for time-series data.
- Use HTMX and Tailwind for all new UI components.
