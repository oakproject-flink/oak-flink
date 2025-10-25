# Oak Flink

A Flink orchestration platform for Kubernetes with centralized management and monitoring.

## Project Structure

This project uses Go workspaces with four modules:

- **api/** - Protocol Buffers / gRPC API definitions
- **oak-lib/** - Shared library (Flink REST API client, certs, logging, gRPC utilities)
- **oak-server/** - Control plane server (manages multiple clusters)
- **oak-agent/** - Kubernetes operator (per-cluster, monitors Flink jobs)

## Prerequisites

- Go 1.21 or higher
- Docker (for integration tests)

## Setup

```bash
# Clone repository
git clone https://github.com/oakproject-flink/oak-flink.git
cd oak-flink

# Initialize workspace
go work sync
```

## Building

```bash
# Build all modules
go build ./...

# Build specific binaries
go build -o bin/oak-server ./oak-server/cmd/oak-server
go build -o bin/oak-agent ./oak-agent/cmd/oak-agent
```

## Testing

```bash
# Unit tests (fast, no Docker required)
./scripts/test-unit.sh        # Linux/macOS/Git Bash
scripts\test-unit.bat         # Windows

# Integration tests (requires Docker)
./scripts/test-integration.sh
scripts\test-integration.bat

# All tests
./scripts/test-all.sh
scripts\test-all.bat
```

## Forking

To fork to your organization:

```bash
# Linux/macOS/Git Bash
bash scripts/update-module-paths.sh github.com/yourorg/oak-flink

# Windows
scripts\update-module-paths.bat github.com/yourorg/oak-flink
```

See [FORKING.md](FORKING.md) for details.

## License

Copyright 2025 Andrei Grigoriu

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.
