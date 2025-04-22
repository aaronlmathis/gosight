# GoSight Project Status

_Last Updated: April 22, 2025_

GoSight is a modern observability and infrastructure monitoring platform built for system administrators, cloud operators, and DevOps engineers. It provides real-time visibility, secure telemetry collection, and intelligent event-based alerting.

See: our [Contributing Guidelines](https://github.com/aaronlmathis/gosight/blob/main/CONTRIBUTING.md) for full instructions on how to setup a dev environment (including victoria metrics and pgsql containers) for gosight using Make.

---

## Agent Capabilities

- **System Metrics Collection**  
  - Supports CPU, memory, disk usage, disk I/O, and network statistics.  
  - Emits structured metrics with namespaces and dimensions.

- **Container Metrics**  
  - Full support for **Podman** and **Docker** runtime metrics.  
  - Per-container CPU, memory, I/O, uptime, and status.

- **Log Collection**  
  - Journald streaming with cursor tracking.  
  - Auth logs from `/var/log/auth.log` or `/var/log/secure`.  
  - Deduplicated, filtered, and enriched with host metadata.

- **Data Streaming**  
  - Secure **mTLS-encrypted gRPC** streaming.  
  - Supports `SubmitMetrics` and `SubmitStream` API.  
  - Batching, retry logic, and worker pool for resilience and throughput.

---

## Server Capabilities

- **Authentication & Access Control**  
  - Local login with **bcrypt** password hashing and TOTP-based **MFA**.  
  - **SSO** via Google, AWS Cognito, and Azure AD (OIDC).  
  - **JWT-based session system** with secure cookies and expiration.  
  - **RBAC** with assignable roles and fine-grained **permissions**.
  - **Audit/Tracing** with trace-id injection on every api call.

- **Live Observability Dashboard**  
  - Dashboard built with Tailwind + Flowbite.  
  - **Real-time charts** using ApexCharts for CPU, memory, disk, and network.  
  - Interactive host and container explorer.  
  - Expandable metric detail view per endpoint.  
  - Role-restricted UI panels.

- **WebSocket Streaming**  
  - Live updates for metrics and logs to the frontend via `/ws/metrics` and `/ws/logs`.

- **Rules, Events, and Alerts**  
  - Expression-based **Rules** with threshold logic.  
  - **Alerts** trigger actions like webhook/email dispatch. (runbook execution to come)
  - YAML/JSON-based definition and configuration of rules and actions.  
  - Inspired by **CloudWatch Alarms** and **EventBridge**.

- **Storage and Querying**  
  - **VictoriaMetrics** is used for scalable metric storage.  
  - Abstracted interface supports future backends (e.g., InfluxDB, TimescaleDB).  
  - API supports:  
    - Full PromQL-style query (`/api/v1/query`)  
    - RESTful discovery endpoints (`/api/v1/{namespace}/{sub}/dimensions`, etc.)  
    - Latest metric values and historical range queries.
    - Querying Logs, Events, Alerts (`/api/v1/logs` `/api/v1/events` `/api/v1/alerts`)

---

## In Progress / Upcoming

- Full UI completion across all dashboard tabs.  
- Log archive backend selector (File, JSONL, Elasticsearch).  
- Additional collector types (Windows Event Log, Cloud Metrics).  
- Alert routing engine (webhook, email, script targets).  
- CLI tools for audit and manual query.  
- Contributor guide and simplified dev setup via `dev/Makefile`.

---

##  Technologies Used

- **Go** for all backend and agent components.  
- **PostgreSQL** for user/role/session metadata.  
- **VictoriaMetrics** for time-series metrics.  
- **Tailwind CSS + Flowbite** for UI design.  
- **Chart.js / ApexCharts** for real-time graphs.  
- **gRPC**, **mTLS**, **JWT**, **OAuth2**, **TOTP** for secure telemetry and auth.
