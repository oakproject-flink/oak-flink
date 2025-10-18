# Oak - Flink Orchestration Platform for Kubernetes

> Strong, reliable Flink orchestration for Kubernetes - where your streaming squirrel feels at home

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

Oak is a comprehensive Flink orchestration platform for Kubernetes that provides centralized management of Flink clusters across multiple environments. It combines monitoring, deployment, autoscaling, and lifecycle management into a unified control plane with distributed sidecars.

## Why Oak?

Just as oak trees provide strong, reliable shelter and produce acorns (the squirrel's resource), Oak provides strong, reliable orchestration and resource management for your Flink jobs.

### What Oak Solves

- **Fragmented tooling** - Consolidates deployment, monitoring, scaling, and savepoint management
- **No multi-cluster visibility** - Unified view of all Flink jobs across dev/staging/prod
- **Manual savepoint workflows** - Automated savepoint creation, management, and restoration
- **Complex deployments** - Simplified Flink job deployment to Kubernetes
- **Reactive scaling only** - Intelligent, proactive autoscaling based on Flink-specific metrics
- **Limited observability** - Deep insights into job health, checkpoints, and backpressure

## Features

- 🎯 **Unified Control Plane** - Manage all Flink jobs across all clusters from one place
- 📊 **Centralized Monitoring** - Real-time dashboard for metrics, health, and alerts
- 💾 **Visual Savepoint Management** - Browse, create, restore savepoints from Web UI
- 🚀 **Simplified Deployment** - Deploy Flink jobs using existing Helm charts via UI or CLI
- ⚖️ **Intelligent Autoscaling** - Flink-aware autoscaling with backpressure, lag, and custom metrics
- 🏥 **Health Management** - Automated recovery, checkpoint monitoring, job restart policies

## Architecture

Oak uses a **two-tier architecture**:

### Oak Server (Centralized Control Plane)
- Centralized history server, orchestrator, and monitoring tool
- Manages multiple Kubernetes clusters across environments
- Aggregates metrics and makes autoscaling decisions
- Provides Web UI and API

### Oak Sidecar (Per-Cluster Data Plane Agent)
- Deployed once per Kubernetes cluster
- Monitors multiple single-job Flink clusters (Application mode)
- Collects metrics and executes scaling operations
- Reports to Oak Server via gRPC

```
┌──────────────────────────────────────┐
│        Oak Server (Control)          │
│    Web UI | API | PostgreSQL         │
└────────────────┬─────────────────────┘
                 │ gRPC
         ┌───────┼───────┐
         ▼       ▼       ▼
    ┌────────┐ ┌────────┐ ┌────────┐
    │ Oak    │ │ Oak    │ │ Oak    │
    │Sidecar │ │Sidecar │ │Sidecar │
    ├────────┤ ├────────┤ ├────────┤
    │ Flink  │ │ Flink  │ │ Flink  │
    │ Jobs   │ │ Jobs   │ │ Jobs   │
    └────────┘ └────────┘ └────────┘
     Dev Cluster Staging  Production
```

## Project Structure

This repository uses Go workspaces with three modules:

```
oak-flink/
├── oak-server/      # Control plane server
├── oak-sidecar/     # Data plane agent
└── oak-lib/         # Shared libraries (Flink client, K8s utils, etc.)
```

## Getting Started

### Prerequisites

- Go 1.21+
- Docker
- kubectl
- kind (for local development)
- PostgreSQL 15+
- Redis 7+

### Quick Start

1. Clone the repository:
```bash
git clone https://github.com/oakproject-flink/oak-flink.git
cd oak-flink
```

2. Start dependencies:
```bash
docker run -d --name oak-postgres \
  -e POSTGRES_DB=oak \
  -e POSTGRES_USER=oak \
  -e POSTGRES_PASSWORD=oak \
  -p 5432:5432 postgres:15

docker run -d --name oak-redis -p 6379:6379 redis:7
```

3. Build and run Oak Server:
```bash
cd oak-server
go run ./cmd/oak-server --config config.yaml
```

4. Build and run Oak Sidecar:
```bash
cd oak-sidecar
go run ./cmd/oak-sidecar --config config.yaml
```

5. Access the Web UI at http://localhost:8080

## Development

### Building

```bash
# Build all binaries
go build -o bin/oak-server ./oak-server/cmd/oak-server
go build -o bin/oak-sidecar ./oak-sidecar/cmd/oak-sidecar
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Test specific module
go test ./oak-lib/flink/...
```

### Project Status

Currently in **Phase 1: MVP** (Monitoring & Manual Scaling)

- [x] Project setup with Go workspaces
- [x] Apache 2.0 License
- [ ] Oak Server with gRPC API
- [ ] Oak Sidecar with metrics collection
- [ ] Basic Web UI
- [ ] Manual scaling via CLI
- [ ] PostgreSQL schema

See [OAK_PROJECT.md](OAK_PROJECT.md) for detailed roadmap.

## Documentation

- [CLAUDE.md](CLAUDE.md) - Development guide for contributors
- [OAK_PROJECT.md](OAK_PROJECT.md) - Detailed project vision and roadmap

## License

Copyright 2025 Cristian-Andrei Grigoriu

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Contributing

Contributions are welcome! This is an open-source project under active development.

## Author

Created by [Cristian-Andrei Grigoriu](https://github.com/yourusername)

## Acknowledgments

Oak is inspired by the Flink squirrel mascot and the strength of oak trees. Built for the Kubernetes and Apache Flink communities.
