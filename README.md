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
├── oak-server/                    # Control plane server (planned)
├── oak-sidecar/                   # Data plane agent (planned)
├── oak-lib/                       # Shared libraries
│   └── flink/
│       └── rest-api/              # Flink REST API client (complete)
└── scripts/                       # Test automation scripts
    ├── test-unit.{bat,sh}         # Run unit tests
    ├── test-integration.{bat,sh}  # Run integration tests
    ├── test-all.{bat,sh}          # Run all tests
    └── test-coverage.{bat,sh}     # Generate coverage report
```

## Current Status

Project is in initial development:

- [x] Repository structure with Go workspaces
- [x] Three modules: oak-server, oak-sidecar, oak-lib
- [x] **Flink REST API client** (oak-lib/flink/rest-api)
  - Complete implementation for Flink 1.18-2.1
  - Cluster operations (overview, config, version detection)
  - Job operations (list, get, cancel, config)
  - JAR operations (upload, run, delete)
  - Savepoint operations (trigger, status, stop with savepoint)
  - Metrics collection
  - 84.7% test coverage
  - Full integration tests with Docker
- [ ] Basic server implementation
- [ ] Sidecar agent implementation
- [ ] Web UI
- [ ] Database schema

### Flink REST API Client - Planned Features

Security features to be implemented (see [Flink Security SSL Documentation](https://nightlies.apache.org/flink/flink-docs-release-2.1/docs/deployment/security/security-ssl/)):

- [ ] **SSL/TLS Support**
  - [ ] `WithTLSConfig(tlsConfig *tls.Config)` option
  - [ ] Mutual TLS (mTLS) authentication
  - [ ] Certificate validation
  - [ ] Custom CA certificates
- [ ] **Authentication**
  - [ ] Basic authentication (username/password)
  - [ ] Token-based authentication
  - [ ] Custom authentication headers

## Development

### Prerequisites

- Go 1.21 or later
- Docker (for integration tests)
- Git

### Getting Started

```bash
# Clone repository
git clone https://github.com/oakproject-flink/oak-flink.git
cd oak-flink

# Sync workspace dependencies
go work sync

# Build (when implemented)
go build -o bin/oak-server ./oak-server/cmd/oak-server
go build -o bin/oak-sidecar ./oak-sidecar/cmd/oak-sidecar
```

## Testing

The Oak project includes comprehensive test suites with automated scripts that work on **Windows, Linux, and macOS**.

### Quick Start

**Windows (cmd.exe or PowerShell):**
```cmd
REM Run unit tests (fast, no Docker required)
scripts\test-unit.bat

REM Run integration tests (requires Docker)
scripts\test-integration.bat

REM Run all tests
scripts\test-all.bat

REM Run tests with coverage report
scripts\test-coverage.bat
```

**Linux / macOS / Git Bash:**
```bash
# Run unit tests (fast, no Docker required)
bash scripts/test-unit.sh

# Run integration tests (requires Docker)
bash scripts/test-integration.sh

# Run all tests
bash scripts/test-all.sh

# Run tests with coverage report
bash scripts/test-coverage.sh
```

### Test Scripts

All test scripts are located in the `scripts/` directory and available in both `.bat` (Windows) and `.sh` (Unix) formats:

| Script | Description | Requirements |
|--------|-------------|--------------|
| `test-unit` | Run all unit tests across modules | Go 1.21+ |
| `test-integration` | Run integration tests with real Flink cluster | Docker |
| `test-all` | Run both unit and integration tests | Docker |
| `test-coverage` | Generate HTML coverage report | Go 1.21+ |

### Manual Testing

**Unit tests only:**
```bash
# Test all modules
go test -short ./oak-lib/...

# Test specific package
go test -short ./oak-lib/flink/rest-api/...

# With verbose output
go test -short -v ./oak-lib/...
```

**Integration tests:**
```bash
# Run integration tests (requires Docker)
cd oak-lib/flink/rest-api
go test -tags=integration -v -timeout 10m

# Run specific integration test
go test -tags=integration -v -run TestIntegration_JAROperations
```

**Coverage:**
```bash
# Generate coverage report
go test -short -coverprofile=coverage.out ./oak-lib/...
go tool cover -html=coverage.out -o coverage.html

# View coverage in terminal
go test -short -cover ./oak-lib/...
```

### Integration Test Details

Integration tests for the Flink REST API client:
- Automatically start Flink 2.1.0 cluster using Docker Compose
- Test against real Flink JobManager REST API
- Upload and execute test JAR files
- Validate savepoint operations
- Test concurrent requests and error handling
- Automatic cleanup after tests complete

**Requirements:**
- Docker running
- Port 8081 available
- ~2GB disk space for Flink image

**Note on Test Caching:**
Integration tests use `-count=1` flag to disable Go's test caching, ensuring fresh Docker container startup every time. Unit tests use caching for faster execution.

### Test Coverage

Current test coverage by module:

| Module | Unit Tests | Integration Tests | Coverage |
|--------|-----------|-------------------|----------|
| oak-lib/flink/rest-api | 19 tests | 14 tests | 84.7% |

**Coverage breakdown:**
- Client operations: ✅ Complete
- Cluster operations: ✅ Complete
- Job operations: ✅ Complete
- JAR operations: ✅ Complete
- Savepoint operations: ✅ Complete
- Metrics collection: ✅ Complete

## License

Copyright 2025 Andrei Grigoriu

Licensed under the Apache License, Version 2.0. See [LICENSE](LICENSE) for details.

## Author

Created by Andrei Grigoriu
