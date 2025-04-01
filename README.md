# GoSight

GoSight is a high-performance, modular, and vendor-agnostic observability platform written in Go. It includes an agent that collects system metrics and meta data and a server that aggregates, stores, and exposes those metrics securely over gRPC.
** This project is still in early development. Not available for production yet, and many features yet to be included.**
## 🌐 Project Overview

- 🔧 Written in pure Go for speed and portability
- 📦 Modular collector architecture (CPU, memory, disk, network, container)
- 🔐 Secure with full TLS and mutual TLS (mTLS) support
- 📊 Built-in web dashboard (HTML/JS)
- 🧰 Cross-platform: runs on Linux, Windows, and containers

## 🧪 Components

### Agent
- Collects system metrics
- Sends them over gRPC (TLS/mTLS) to the server
- Configurable via `agent/config.yaml`

### Server
- Accepts incoming metrics
- Verifies client identity (mTLS)
- Exposes metrics and dashboards
- Configurable via `server/config.yaml`

---

## 🚀 Quick Start (Dev)

```bash
# From project root
go run ./server/cmd &
go run ./agent/cmd
```

Ensure you’ve generated valid certificates before starting.

---

## 🔐 TLS / mTLS Setup

Certs live in the `/certs` directory. You can regenerate everything using:

```bash
# Linux/macOS
./install/generate_certs_with_san.sh

# Windows PowerShell
./install/generate_certs_with_san.ps1
```

Update paths in `config.yaml` files accordingly.

---

## 📂 Folder Structure (Core)

```
/agent/         - Agent source code and CLI
/server/        - Server source code and CLI
/shared/        - Shared models and proto definitions
/certs/         - TLS and mTLS certificates
/install/       - Cert generation scripts
```

---

## 🛠 Build

```bash
go build -o gosight-agent ./agent/cmd
go build -o gosight-server ./server/cmd
```

---

## 📋 License

GoSight is licensed under the [GPL-3.0-or-later](https://www.gnu.org/licenses/gpl-3.0.html).
