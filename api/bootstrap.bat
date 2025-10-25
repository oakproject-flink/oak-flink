@echo off
REM bootstrap.bat
REM Bootstrap the API module for new developers

echo ========================================
echo Bootstrapping Oak API module...
echo ========================================
echo.

REM Install tools
call install-tools.bat
if errorlevel 1 (
    echo Bootstrap failed during tool installation
    exit /b 1
)

echo.
echo Running code generation...
protoc -I. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/oak/v1/*.proto
if errorlevel 1 (
    echo Failed to generate proto code
    echo Make sure protoc is properly installed with: winget install protoc
    exit /b 1
)

echo.
echo ========================================
echo ✅ Bootstrap complete! You're ready to develop.
echo ========================================