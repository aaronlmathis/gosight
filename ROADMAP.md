# GoSight Roadmap 

Built with Go. Driven by open standards. Aimed at replacing bloatware.

---

## Phase 1: Core Foundations (Complete)
> _“Build it right, not rushed.”_

- [x] Agent/Server gRPC Streaming
- [x] TLS Encryption, MFA, and SSO (Google, AWS, Azure)
- [x] Role-Based Access Control (RBAC)
- [x] VictoriaMetrics backend (scalable time-series DB)
- [x] Real-time dashboard with tabs, charts, container & host metrics

---

## Phase 2: Logs + Metrics Fusion
> _“Correlate the ‘why’ with the ‘when.’”_

- [ ] Unified log ingestion (JSON, syslog, journald)
- [ ] Pluggable log storage (disk/SQLite → Elastic or Loki)
- [ ] Log query API (filter by time, severity, service)
- [ ] Logs linked to metrics via `endpoint_id` and timestamp
- [ ] `/logs` page: live search, filter, export, chart overlays

---

## Phase 3: Visual Intelligence
> _“Make data visible, filterable, and actionable.”_

- [ ] Saved dashboards and queries
- [ ] Dashboard filters (host, label, metric, etc.)
- [ ] Multi-host comparisons (e.g. CPU/mem across agents)
- [ ] Anomaly detection helpers (spikes, gaps)
- [ ] (Optional) Drag-and-drop dashboard layout

---

## Phase 4: Alerts + Actions
> _“Observability isn’t useful if you’re not notified.”_

- [ ] Threshold alert rules (e.g. `mem.used > 80% for 5m`)
- [ ] Webhook + email alert delivery
- [ ] `/alerts` view with dismiss/archive
- [ ] Chart annotations showing alert triggers

---

## Phase 5: Multi-Tenant + Enterprise Readiness
> _“Let teams run side by side.”_

- [ ] Organizations / Projects / Scoping
- [ ] RBAC with resource ownership + audit log
- [ ] API tokens (scopes, expiration)
- [ ] User quotas / resource limits
- [ ] Helm chart or containerized deployment

---

## Bonus Phase: Delight & Ecosystem
> _“You had me at dark mode.”_

- [x] Dark/light theme toggle
- [ ] CLI / TUI client
- [ ] WebSocket stream viewer (live payload inspector)
- [ ] Embedded console / tmux log replay
- [ ] Plugin/event hook system

---

## North Star Goals

- Secure, modular observability stack in Go
- Simple, self-hosted alternative to Datadog/Splunk
- Works offline or in air-gapped environments
- No JVM, no vendor lock-in, no bloat

---

_Contributors welcome — especially for frontend, UX, and visualization!_
