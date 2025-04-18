# 🤝 Contributing to GoSight

Welcome to GoSight — a secure, modern observability platform for cloud-native systems. Whether you're a backend engineer, frontend designer, or DevOps wizard, we're excited to collaborate with you.

---

## 📦 Project Overview

GoSight is a modular observability tool written in Go. It includes:

- **Agent**: Collects metrics/logs (Go, gRPC, systemd, gopsutil)
- **Server**: Handles secure metric ingestion, stores data in VictoriaMetrics
- **UI**: Admin dashboard (Tailwind, Flowbite, Chart.js, Go templates)
- **Auth**: Local login + SSO + MFA + RBAC

---

## 🛠️ Quickstart

```bash
cd dev
./setup_dev_env.sh
# or
make
```

This will:
- Clone the repo
- Start PostgreSQL + VictoriaMetrics in containers
- Load the schema and generate certs
- Build agent/server with configs in `./configs`

Credentials:
- **Postgres**: `gosight` / `devpassword`
- **Default User**: `dev` / `password`

Edit configs here:
- `configs/agent.yaml`
- `configs/server.yaml`

---

## 💡 Contributor Focus Areas

| Area              | Skills Needed              | Example Tasks                         |
|-------------------|----------------------------|---------------------------------------|
| Frontend UI       | JS, Tailwind, Flowbite     | Metrics charts, filters, alerts       |
| Agent Collectors  | Go, Linux internals        | Disk, I/O, cgroups, container stats   |
| Server API        | Go, gRPC, PromQL, JSON API | Add endpoints, query performance      |
| Auth/Security     | Go, JWT, MFA, OAuth2       | GitHub SSO, role auditing             |
| DevOps/Infra      | Bash, Postgres, Docker     | Helm charts, CI, log indexing         |

---

## 📁 Repo Structure

```
gosight/
├── agent/         # Collectors, config, streaming logic
├── server/        # HTTP + gRPC server, auth, dashboards
├── shared/        # Protobufs, models, common utils
├── dev/           # Dev tools, setup scripts, schema
```

---

## 🚀 Getting Involved

1. Fork and clone the repo
2. Create a branch: `git checkout -b feat/your-feature`
3. Code, test, commit
4. Push and open a PR

---

## ✅ Good First Issues

- [ ] Convert dashboard table filters to Flowbite components
- [ ] Add `/api/logs` endpoint for log ingestion
- [ ] Improve structured logging with request trace IDs
- [ ] Add test fixtures for metrics and logs

---

## 🧪 Tips for Testing

```bash
# Run agent
GOSIGHT_AGENT_CONFIG=./configs/agent.yaml ./agent/gosight-agent

# Run server
GOSIGHT_SERVER_CONFIG=./configs/server.yaml ./server/gosight

# Query VictoriaMetrics
curl 'http://localhost:8428/api/v1/series?match[]=container.cpu.usage'
```

---

## 🙏 Thanks
We appreciate your interest and contributions to GoSight. Join the mission to build a secure, pluggable, modern observability tool.

> Maintainer: [@aaronlmathis](https://github.com/aaronlmathis)
