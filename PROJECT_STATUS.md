## ✅ Current Status (as of 2025-04-02)

### Security
- [x] TLS and optional mTLS fully supported via `config.yaml`
- [x] Logs SHA256 fingerprint and Common Name (CN) of connecting agent certificates

### Agent
- [x] CPU, Memory, Disk, and Network collectors implemented using `gopsutil`
- [x] Collector registry and runtime loader (`registry.go`)
- [x] Configurable agent via `config.yaml`, CLI flags, and ENV vars
- [x] Periodic collection loop using `runner.go`
- [x] `model.Meta` system implemented for sending host/container/application metadata
- [x] User-defined custom tags via `yaml` and CLI `--tags` flag
- [x] `utils.MergeMaps()` utility to merge agent-level and runtime tags
- [x] Agent uses gRPC streaming (`SubmitStream`) with TLS/mTLS support
- [x] Protobuf schema updated with full `Meta` struct and map tags
- [x] Metrics packaged and sent as `MetricPayload` via gRPC stream
- [x] `convert.go` updated to include full `Meta` field serialization

### Server
- [x] gRPC server accepts streamed metrics with mTLS (optional)
- [x] Graceful shutdown and signal handling using `grpcServer.GracefulStop()`
- [x] Modular storage backend with current support for VictoriaMetrics
- [x] VictoriaMetrics batch writer implemented with:
  - gzip compression  
  - retry/backoff logic  
  - worker pool  
  - full meta label conversion
- [x] Dynamic Prometheus-style label generation from `model.Meta`
- [x] Internal debug and metrics dashboard logic tested via `/api/v1/*`
- [x] Custom CLI script built to audit stored metrics and metadata

---

## 🛣️ Roadmap Items

1. **Historical Storage & Querying**  
   Store timestamped metrics in VictoriaMetrics and later allow query via Go API or dashboard UI.

2. **Container & Kubernetes Collectors**  
   Add native Podman collector (REST), Docker (optional), and later K8s metadata tagging.

3. **Dashboard UI**  
   Tailwind-powered dark theme dashboard already prototyped. Integrate metric graphs and alert banners.

4. **Alerting & Triggers**  
   Allow users to define thresholds or PromQL-style rules to trigger actions or webhook notifications.

5. **Additional System Collectors**
   - `loadavg.go` — load average (Linux only)  
   - `uptime.go` — simple metric for system uptime  
   - `processes.go` — total processes, running, sleeping, etc.  
   - `filesystem.go` — count of mounted volumes, failed mounts  

---

## 🔧 In Progress / Next Goals

- [ ] Refactor agent and server config structs to support full TLS/mTLS config validation  
- [ ] Historical dashboard views and time-series charting  
