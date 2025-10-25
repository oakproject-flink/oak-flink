@echo off
REM install-tools.bat
REM Installs build tools tracked in tools.go

echo Installing build tools (tracked in tools.go)...

go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
if errorlevel 1 (
    echo Failed to install protoc-gen-go
    exit /b 1
)

go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
if errorlevel 1 (
    echo Failed to install protoc-gen-go-grpc
    exit /b 1
)

echo ✅ Build tools installed successfully