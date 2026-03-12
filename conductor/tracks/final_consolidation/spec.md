# Specification: Final Production Consolidation

## Goal
Complete the final set of integration and hardening tasks required for a production-ready OpsPilot platform. This includes finalizing network automation, observability depth, data resilience, and real-time UI components.

## Scope
- **Network & DNS**: Finalize Windows DNS integration via SSH.
- **Observability**: Implement real Docker metrics collection, VictoriaMetrics long-term storage, and granular topology visualization.
- **Resilience**: Configure Postgres WAL archiving and finalize registry synchronization.
- **UI/UX**: Build the real-time container log viewer.

## Technical Requirements
- Use `golang.org/x/crypto/ssh` for Windows DNS automation.
- Use Docker SDK for real-time metrics.
- Use VictoriaMetrics API for metrics persistence.
- Use WebSockets for live log streaming to the frontend.
- Use HTMX and Tailwind for UI updates.
