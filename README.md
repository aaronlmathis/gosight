# GoSight - Lightweight Self-Hosted Monitoring Agent

GoSight is a minimalist system metrics exporter and dashboard built in Go. Designed for Linux/cloud administrators and DevOps engineers, it monitors system health and exposes metrics over HTTP or Prometheus format.

## 🔍 Features

- Concurrent metric collection (CPU, memory, disk, network)
- HTTP API and embedded dashboard
- Optional Prometheus or gRPC exporter
- Configurable via YAML or ENV
- Graceful shutdowns with context
- Docker-ready and systemd-compatible

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- Linux system (for metrics)
- \[Optional\] Docker & Prometheus

### Installation

```bash
git clone https://github.com/yourname/gosight.git
cd gosight
go build -o gosight
```

### Run

```bash
./gosight --config config.yaml
```

### Docker

```bash
docker build -t gosight .
docker run -p 8080:8080 --rm gosight
```

## 📊 Metrics

| Metric         | Description              |
|----------------|--------------------------|
| cpu_usage      | Percent used (avg)       |
| memory_usage   | RAM used / total         |
| disk_usage     | % used per mount         |
| net_traffic    | Bytes sent/received      |

## ⚙️ Configuration

```yaml
server:
  port: 8080
metrics:
  interval_seconds: 5
exporters:
  prometheus: true
  dashboard: true
```

## 📂 Project Structure

```
gosight/
├── cmd/
│   └── main.go
├── internal/
│   ├── collector/
│   ├── exporter/
│   ├── config/
│   └── utils/
├── web/
│   └── static/
│   └── templates/
├── config.yaml
├── Dockerfile
├── go.mod
└── README.md
```

## 🧠 Concepts Demonstrated

- Goroutines, channels, worker pools
- HTTP server & REST endpoints
- File parsing (/proc, /sys)
- Signal handling and graceful shutdown
- Modular design with interfaces
- Optional gRPC, WebSocket streaming

## 📜 License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.
