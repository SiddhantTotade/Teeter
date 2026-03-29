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

## ⚙️ Configuration Example

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

---

*Teeter — Balancing your traffic with ease.*
