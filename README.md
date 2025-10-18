# Oak - Flink Orchestration Platform for Kubernetes

⚠️ **Work in Progress** - This project is in early development. Features listed below are planned ideas, not yet implemented.

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go)](https://go.dev/)

Oak aims to be a Flink orchestration platform for Kubernetes providing centralized management of Flink clusters across multiple environments.

## Planned Features

- Unified control plane for managing Flink jobs across clusters
- Centralized monitoring dashboard
- Savepoint management
- Job deployment automation
- Autoscaling based on Flink metrics
- Health monitoring and alerting

## Planned Architecture

Two-tier design:

- **Oak Server**: Centralized control plane (Web UI, API, database)
- **Oak Sidecar**: Per-cluster agent monitoring Flink jobs in Kubernetes Application mode

Communication via gRPC.

## Project Structure

This repository uses Go workspaces with three modules:

```
oak-flink/
├── oak-server/      # Control plane server
├── oak-sidecar/     # Data plane agent
└── oak-lib/         # Shared libraries (Flink client, K8s utils, etc.)
```

## Current Status

Project is in initial setup phase:

- [x] Repository structure with Go workspaces
- [x] Three modules: oak-server, oak-sidecar, oak-lib
- [ ] Basic server implementation
- [ ] Sidecar agent implementation
- [ ] Flink REST API client
- [ ] Web UI
- [ ] Database schema

## Development

```bash
# Clone repository
git clone https://github.com/oakproject-flink/oak-flink.git
cd oak-flink

# Build (when implemented)
go build -o bin/oak-server ./oak-server/cmd/oak-server
go build -o bin/oak-sidecar ./oak-sidecar/cmd/oak-sidecar

# Run tests
go test ./...
```

## License

Copyright 2025 Andrei Grigoriu

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Author

Created by Andrei Grigoriu
