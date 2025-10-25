@echo off
REM Run integration tests against all Flink versions (1.18-2.1)
REM Usage: test-all-flink-versions.bat [-y|--yes]
REM   -y, --yes: Skip confirmation prompt (for CI/automation)
REM
REM Environment variables:
REM   CI=true or CONTINUOUS_INTEGRATION=true: Auto-skip prompt
REM
REM Requirements:
REM   - Docker must be running
REM   - Port 8081 must be available
REM   - 10-15 minutes to complete

setlocal enabledelayedexpansion

REM Check for -y or --yes flag, or CI environment
set "SKIP_PROMPT=0"
if "%~1"=="-y" set "SKIP_PROMPT=1"
if "%~1"=="--yes" set "SKIP_PROMPT=1"
if "%CI%"=="true" set "SKIP_PROMPT=1"
if "%CONTINUOUS_INTEGRATION%"=="true" set "SKIP_PROMPT=1"
if "%GITHUB_ACTIONS%"=="true" set "SKIP_PROMPT=1"

echo ========================================
echo   Oak Project - All Flink Versions Test
echo ========================================
echo.

REM Check if Docker is running
echo Checking Docker...
docker info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not running
    echo Please start Docker Desktop and try again
    exit /b 1
)
echo [OK] Docker is running

echo.
echo This will test against ALL supported Flink versions:
echo   - Flink 1.18.1
echo   - Flink 1.19.1
echo   - Flink 1.20.0
echo   - Flink 2.0.0
echo   - Flink 2.1.0
echo.
echo WARNING: This will take 10-15 minutes to complete
echo Each version will start a Docker container, run tests, and clean up
echo.

REM Skip prompt in CI or with -y flag
if %SKIP_PROMPT%==1 (
    echo [CI MODE] Skipping confirmation prompt
    echo.
) else (
    echo Press any key to continue or Ctrl+C to cancel...
    pause >nul
)

echo.
echo ========================================
echo Running Multi-Version Integration Tests
echo ========================================
echo.

cd /d "%~dp0\..\oak-lib\flink\rest-api"
go test -tags=integration_versions -v -timeout 30m -count=1

if errorlevel 1 (
    echo.
    echo ========================================
    echo [ERROR] Some tests failed
    echo ========================================
    exit /b 1
)

echo.
echo ========================================
echo [SUCCESS] All Flink version tests passed!
echo ========================================
exit /b 0
