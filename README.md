# Teeter — High-Performance L7 Load Balancer

Teeter is a lightweight, production-grade **Layer 7 (Application Layer) Load Balancer and Reverse Proxy** written in Go. It provides advanced traffic management features including path-based routing, multiple balancing strategies, active health checks, and a live Admin API, all configured via a simple YAML file.

## 🚀 Introduction

Teeter is designed to sit in front of your microservices (e.g., a Java backend and a React frontend) and route traffic based on URL prefixes. It ensures high availability by automatically detecting failed backends and retrying requests across healthy nodes.

### Key Features:
- **Path-Based Routing**: Route by URL prefix (longest-match first).
- **Multiple Strategies**: Round Robin, Weighted Round Robin, and Least Connections.
- **Resilience**: Active health checks, Circuit Breaker, and Exponential Backoff retries.
- **Traffic Shaping**: Token-bucket rate limiting and worker-pool request queuing.
- **Live Admin API**: Inspect status and add backends at runtime without restarts.
- **Metrics & Observability**: Native Prometheus instrumentation and pre-configured Grafana dashboards.

## 📋 Prerequisites

- **Go 1.23.2+**
- **YAML** configuration file (`config.yaml`)

## ⚡ Usage

### 1. Clone & Build
```bash
git clone https://github.com/SiddhantTotade/Teeter.git
cd Teeter
```

### 2. Run the Load Balancer
You can run Teeter by pointing it to your configuration file:

```bash
go run lb/cmd/lb/main.go -config config.yaml
```

### 3. Accessing the Services
- **Proxy Entry Point**: `http://localhost:1996` (based on default config)
- **Admin API**: `http://localhost:1997/status`
- **Prometheus Metrics**: `http://localhost:1997/metrics`

### 💡 Manual Network Migration
When your IP address changes (e.g., switching WiFi), you must manually update the backend URLs in `config.yaml`. See the [Manual Migration Guide](file:///C:/Users/cools/.gemini/antigravity/brain/a4c62a47-464b-43c3-a6e9-1ec1e99fa8a6/manual_migration_guide.md) for a full list of files.

```yaml
# Manual IP configuration
backends:
  - url: "http://192.168.1.18:1994"
```

## ⚙️ Configuration Reference

Teeter is configured via `config.yaml`. Below is an example setup for a common Java + React full-stack application:

```yaml
port: 1996         # The main entry point for your application
admin_port: 1997    # Port for the Admin API

routes:
  # Route all /api requests to the Java backend
  - prefix: "/api"
    strategy: "round_robin"
    backends:
      - url: "http://127.0.0.1:8080"
        timeout: "3s"

  # Route all other traffic to the React frontend
  - prefix: "/"
    strategy: "least_connections"
    backends:
      - url: "http://127.0.0.1:3000"
        timeout: "10s"
```

### Route Matching Rules
Routes are matched by **longest prefix first**. This means specific routes like `/api/auth` will be matched before a generic `/api` or the catch-all `/` route.

## 📊 Admin API

The Admin API provides real-time observability into your cluster.

- **GET /status**: Returns a JSON object containing the health status, alive state, and failure counts for all backends.
- **POST /add-backend**: Dynamically adds a new backend URL to an existing route without requiring a service restart.

```bash
# Example: Add a new server to the /api route
curl -X POST http://localhost:1997/add-backend \
  -H "Content-Type: application/json" \
  -d '{"prefix": "/api", "url": "http://127.0.0.1:8081"}'
```

## 📈 Monitoring & Observability

Teeter is built with observability in mind. Every request, latency, and health check is exported for analysis.

### Docker Compose Stack
The easiest way to run the full monitoring suite is via Docker Compose:

```bash
docker compose up --build
```

This will launch:
- **Teeter**: The load balancer.
- **Prometheus**: Collected metrics at `http://localhost:9090`.
- **Grafana**: A pre-configured dashboard at `http://localhost:3005`.

### Custom Dashboard
Teeter includes a pre-provisioned Grafana dashboard that visualizes:
- **RPS (Requests Per Second)**: Real-time traffic throughput.
- **Latencies**: p50 and p99 quantiles.
- **Backend Health**: Visual indicators of which backends are online.

---

*Teeter — Balancing your traffic with ease.*
