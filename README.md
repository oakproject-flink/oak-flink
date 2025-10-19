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

REM Run integration tests (requires Docker, Flink 2.1)
scripts\test-integration.bat

REM Run all tests (unit + integration)
scripts\test-all.bat

REM Test against ALL Flink versions 1.18-2.1 (10-15 min)
scripts\test-all-flink-versions.bat

REM Run tests with coverage report
scripts\test-coverage.bat
```

**Linux / macOS / Git Bash:**
```bash
# Run unit tests (fast, no Docker required)
bash scripts/test-unit.sh

# Run integration tests (requires Docker, Flink 2.1)
bash scripts/test-integration.sh

# Run all tests (unit + integration)
bash scripts/test-all.sh

# Test against ALL Flink versions 1.18-2.1 (10-15 min)
bash scripts/test-all-flink-versions.sh

# Run tests with coverage report
bash scripts/test-coverage.sh
```

### Test Scripts

All test scripts are located in the `scripts/` directory and available in both `.bat` (Windows) and `.sh` (Unix) formats:

| Script | Description | Requirements | Duration |
|--------|-------------|--------------|----------|
| `test-unit` | Run all unit tests across modules | Go 1.21+ | ~15s |
| `test-integration` | Run integration tests with Flink 2.1 | Docker | ~30s |
| `test-all` | Run both unit and integration tests | Docker | ~45s |
| `test-all-flink-versions` | **Test against ALL Flink versions (1.18-2.1)** | Docker | **10-15 min** |
| `test-coverage` | Generate HTML coverage report | Go 1.21+ | ~15s |

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
# Run integration tests against Flink 2.1 (requires Docker)
cd oak-lib/flink/rest-api
go test -tags=integration -v -timeout 10m

# Run specific integration test
go test -tags=integration -v -run TestIntegration_JAROperations

# Test against ALL Flink versions 1.18-2.1 (10-15 minutes)
go test -tags=integration_versions -v -timeout 30m
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

### Multi-Version Integration Tests

The `test-all-flink-versions` suite tests the REST API client against **all supported Flink versions sequentially**:

**Supported Versions:**
- ✅ Flink 1.18.1 (API version: 1.18-1.19)
- ✅ Flink 1.19.1 (API version: 1.18-1.19)
- ✅ Flink 1.20.0 (API version: 2.0+)
- ✅ Flink 2.0.1 (API version: 2.0+)
- ✅ Flink 2.1.0 (API version: 2.0+)

**What it does:**
- Starts each Flink version in Docker sequentially
- Runs comprehensive test suite for each version:
  - Cluster overview and configuration
  - Version detection and API compatibility
  - Job management (list, get, cancel)
  - JAR operations (upload, run, delete)
  - Error handling
- Automatically cleans up containers after each version
- Takes ~10-15 minutes total (2-3 min per version)

**Use cases:**
- Verify compatibility across all supported Flink versions
- Test API changes between versions
- Ensure version detection works correctly
- Run before releases to validate multi-version support

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
