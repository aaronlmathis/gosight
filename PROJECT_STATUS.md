# GoSight Project Status

Last updated: 2025-04-05

## ✅ Completed Milestones

- [x] **Agent ↔ Server gRPC streaming works**
  - CPU metrics collector implemented
  - Worker pool and retry logic in agent sender
- [x] **Server uses TLS**
  - gRPC server now serves over TLS using a server certificate signed by a local CA
  - TLS configuration is fully loaded from `server/config.yaml`
- [x] **Agent connects securely over TLS**
  - Agent validates the server using the CA certificate
  - TLS configuration is fully loaded from `agent/config.yaml`
- [x] **Mutual TLS (mTLS) implemented**
  - Agent presents client certificate during handshake
  - Server validates agent certificate using client CA (same CA used to sign both)
- [x] **Cert generation tooling added**
  - Bash and PowerShell scripts support SAN-based certs
  - Scripts located in `install/`, output to `/certs`
- [x] **gRPC server bootstrapping refactored**
  - Clean `NewGRPCServer(cfg)` returns `*grpc.Server` and `net.Listener`
  - Optional reflection controlled via `debug.enable_reflection` in config
- [x] **Project folder structure audit complete**
  - TLS helpers, sender, and runner logic separated cleanly
  - Internal paths follow Go idioms (e.g. `internal/config`, `internal/sender`)
- [x] **Graceful shutdown and signal handling for gRPC server and agent**
  - Agent and server now exit cleanly on `SIGINT` or `SIGTERM`
- [x] **TLS fingerprint logging**
  - TLS fingerprint or Common Name from connecting agent cert is logged at connect
- [x] **Podman container collector added**
  - Native Podman REST API integration (no Docker dependency)
  - Includes memory, CPU, network, and metadata collection per container
  - Integrated into agent and dashboard
- [x] **Admin Dashboard Mocked-up**
  - Responsive design using JS / TailwindCSS / Flowbite
  - Includes animated graphs / charts
- [X] **Persistent metric storage backend**
  - Implemented MetricStore abstraction layer implemented to support multiple time-series backends
  - Initial backend integration completed using VictoriaMetrics:
    - Prometheus-compatible label formatting
    - Batch write support via /api/v1/import/prometheus
    - Gzip compression and retry logic
    - Tag enrichment with metadata (e.g. hostname, container ID)

## 🔜 In Progress / Next
- [ ] Design and implement auth package with support for:
    - Single Sign-On (SSO) via Google, AWS Cognito, and Azure AD
    - Multi-factor authentication (MFA) including TOTP and hardware tokens (e.g. YubiKey)
    - Role-based access control (RBAC) for dashboard and API endpoints
- [ ] Design and implement UserStore abstraction layer implemented to support multiple SQL backends for storing User / Role / Permission data
- [ ] Finalize front-end html/css/js templates for administration panel
- [ ] Refactor agent and server config structs to support full TLS/mTLS config validation
- [ ] Historical dashboard views and time-series charting
- [ ] Alerting engine and trigger conditions
- [ ] Podman container lifecycle tracking and restart alerting


