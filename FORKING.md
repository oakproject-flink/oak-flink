# Forking Oak

This document explains how to fork the Oak project and customize it for your organization.

## Why Update Module Paths?

Oak uses Go workspaces with multiple modules (`api`, `oak-lib`, `oak-server`, `oak-agent`). Each module has a module path defined in its `go.mod` file. When you fork the repository, you'll want to update these paths to point to your own organization.

**Current module paths:**
```
github.com/oakproject-flink/oak-flink/api
github.com/oakproject-flink/oak-flink/oak-lib
github.com/oakproject-flink/oak-flink/oak-server
github.com/oakproject-flink/oak-flink/oak-agent
```

**After forking to your org:**
```
github.com/yourorg/oak-flink/api
github.com/yourorg/oak-flink/oak-lib
github.com/yourorg/oak-flink/oak-server
github.com/yourorg/oak-flink/oak-agent
```

## Quick Start

We provide scripts to automate this process:

### Linux / macOS / Git Bash

```bash
# 1. Fork the repository on GitHub/GitLab
# 2. Clone your fork
git clone https://github.com/yourorg/oak-flink.git
cd oak-flink

# 3. Run the update script
bash scripts/update-module-paths.sh github.com/yourorg/oak-flink

# 4. Sync workspace and update dependencies
go work sync
cd api && go mod tidy && cd ..
cd oak-lib && go mod tidy && cd ..
cd oak-server && go mod tidy && cd ..
cd oak-agent && go mod tidy && cd ..

# 5. Verify everything builds
go build ./...

# 6. Commit changes
git add .
git commit -m "Update module paths for forked repository"
git push origin main
```

### Windows (PowerShell or cmd.exe)

```cmd
REM 1. Fork the repository on GitHub/GitLab
REM 2. Clone your fork
git clone https://github.com/yourorg/oak-flink.git
cd oak-flink

REM 3. Run the update script
scripts\update-module-paths.bat github.com/yourorg/oak-flink

REM 4. Sync workspace and update dependencies
go work sync
cd api && go mod tidy && cd ..
cd oak-lib && go mod tidy && cd ..
cd oak-server && go mod tidy && cd ..
cd oak-agent && go mod tidy && cd ..

REM 5. Verify everything builds
go build ./...

REM 6. Commit changes
git add .
git commit -m "Update module paths for forked repository"
git push origin main
```

## What the Script Does

The `update-module-paths` script:

1. **Creates a backup** of all Go and proto files
2. **Updates all occurrences** of the old module path in:
   - `go.mod` files (module declarations)
   - `*.go` files (import statements)
   - `*.proto` files (go_package options)
3. **Shows progress** as it updates each file
4. **Saves backup** in `.backup-YYYYMMDD-HHMMSS/` directory

## Manual Method (Alternative)

If you prefer to update manually:

### 1. Update go.mod Files

**api/go.mod:**
```go
module github.com/yourorg/oak-flink/api
```

**oak-lib/go.mod:**
```go
module github.com/yourorg/oak-flink/oak-lib
```

**oak-server/go.mod:**
```go
module github.com/yourorg/oak-flink/oak-server
```

**oak-agent/go.mod:**
```go
module github.com/yourorg/oak-flink/oak-agent
```

### 2. Update Import Statements

Find and replace in all `.go` files:
- Find: `github.com/oakproject-flink/oak-flink`
- Replace: `github.com/yourorg/oak-flink`

Example import before:
```go
import (
    oakv1 "github.com/oakproject-flink/oak-flink/api/proto/oak/v1"
    "github.com/oakproject-flink/oak-flink/oak-lib/certs"
)
```

Example import after:
```go
import (
    oakv1 "github.com/yourorg/oak-flink/api/proto/oak/v1"
    "github.com/yourorg/oak-flink/oak-lib/certs"
)
```

### 3. Update Proto Files

Update `go_package` options in `api/proto/oak/v1/*.proto`:

Before:
```protobuf
option go_package = "github.com/oakproject-flink/oak-flink/api/proto/oak/v1;oakv1";
```

After:
```protobuf
option go_package = "github.com/yourorg/oak-flink/api/proto/oak/v1;oakv1";
```

### 4. Sync and Tidy

```bash
go work sync
cd api && go mod tidy && cd ..
cd oak-lib && go mod tidy && cd ..
cd oak-server && go mod tidy && cd ..
cd oak-agent && go mod tidy && cd ..
```

## Using a Custom Domain (Advanced)

Instead of GitHub/GitLab URLs, you can use a custom domain like `oak.yourcompany.com`:

```
oak.yourcompany.com/api
oak.yourcompany.com/oak-lib
oak.yourcompany.com/oak-server
oak.yourcompany.com/oak-agent
```

**Requirements:**
1. DNS record for your domain
2. HTTP server responding to `https://oak.yourcompany.com/module?go-get=1` with:
   ```html
   <meta name="go-import" content="oak.yourcompany.com git https://github.com/yourorg/oak-flink">
   ```

**Benefits:**
- Cleaner import paths
- Independence from Git hosting provider
- Professional appearance

**See:** https://go.dev/ref/mod#serving-from-proxy

## Troubleshooting

### Error: "module not found"

```bash
# Clear Go cache
go clean -modcache

# Re-sync workspace
go work sync

# Re-download dependencies
cd api && go mod download && cd ..
cd oak-lib && go mod download && cd ..
cd oak-server && go mod download && cd ..
cd oak-agent && go mod download && cd ..
```

### Error: "import path mismatch"

Make sure ALL occurrences of the old path are replaced. Re-run the script or use a text editor's find-in-files feature.

### Backup restore

If something goes wrong:

**Linux/macOS:**
```bash
# Restore from backup (replace TIMESTAMP with actual timestamp)
cp -r .backup-YYYYMMDD-HHMMSS/* .
```

**Windows:**
```cmd
REM Restore from backup (replace TIMESTAMP with actual timestamp)
xcopy .backup-YYYYMMDD-HHMMSS\* . /S /Y
```

## CI/CD Considerations

After forking, update:
- Docker image names in `Dockerfile.*`
- Container registry URLs in CI/CD pipelines
- Helm chart repository references
- Documentation URLs

## Contributing Back

If you make improvements and want to contribute back to the original Oak project:

1. Create patches against the original module paths
2. Submit a pull request to `github.com/oakproject-flink/oak-flink`
3. The maintainers will review and merge

## Questions?

- Open an issue on GitHub
- Check the main README.md
- See CLAUDE.md for development guidelines
