@echo off
REM Run integration tests for Oak project
REM Works on Windows cmd.exe and PowerShell
REM
REM Requirements:
REM   - Docker must be running
REM   - Port 8081 must be available (for Flink)

setlocal enabledelayedexpansion

echo ========================================
echo   Oak Project - Integration Tests
echo ========================================
echo.

REM Navigate to project root
cd /d "%~dp0\.."

REM Check if Docker is running
echo Checking Docker...
docker ps >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not running!
    echo Please start Docker and try again.
    exit /b 1
)
echo [OK] Docker is running
echo.

echo Running integration tests...
echo (This will start a Flink cluster in Docker)
echo.

REM Run integration tests with tags (-count=1 disables test caching)
go test -tags=integration -v -timeout 10m -count=1 ./oak-lib/flink/rest-api/...
if errorlevel 1 (
    echo.
    echo ========================================
    echo   ERROR: Integration tests failed!
    echo ========================================
    exit /b 1
)

echo.
echo ========================================
echo   SUCCESS: All integration tests passed!
echo ========================================
exit /b 0
