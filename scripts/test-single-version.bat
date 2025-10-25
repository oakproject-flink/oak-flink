@echo off
REM Run integration tests against a specific Flink version
REM Usage: test-single-version.bat [VERSION]
REM   VERSION: 1.18, 1.19, 1.20, 2.0, 2.1 (default: 2.1)
REM
REM Requirements:
REM   - Docker must be running
REM   - Port 8081 must be available

setlocal enabledelayedexpansion

REM Default to Flink 2.1 if no argument provided
set "VERSION=%~1"
if "%VERSION%"=="" set "VERSION=2.1"

echo ========================================
echo   Oak - Single Version Test
echo   Flink %VERSION%
echo ========================================
echo.

REM Validate version
if "%VERSION%"=="1.18" (
    set "COMPOSE_FILE=docker-compose-1.18.yml"
    set "FULL_VERSION=1.18.1"
) else if "%VERSION%"=="1.19" (
    set "COMPOSE_FILE=docker-compose-1.19.yml"
    set "FULL_VERSION=1.19.1"
) else if "%VERSION%"=="1.20" (
    set "COMPOSE_FILE=docker-compose-1.20.yml"
    set "FULL_VERSION=1.20.0"
) else if "%VERSION%"=="2.0" (
    set "COMPOSE_FILE=docker-compose-2.0.yml"
    set "FULL_VERSION=2.0.0"
) else if "%VERSION%"=="2.1" (
    set "COMPOSE_FILE=docker-compose-2.1.yml"
    set "FULL_VERSION=2.1.0"
) else (
    echo [ERROR] Invalid version: %VERSION%
    echo.
    echo Valid versions: 1.18, 1.19, 1.20, 2.0, 2.1
    echo Usage: test-single-version.bat [VERSION]
    exit /b 1
)

REM Check Docker
echo Checking Docker...
docker info >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Docker is not running
    echo Please start Docker and try again
    exit /b 1
)
echo [OK] Docker is running
echo.

REM Navigate to test directory
cd /d "%~dp0\..\oak-lib\flink\rest-api"

echo Starting Flink %FULL_VERSION% cluster...
docker compose -f testdata\%COMPOSE_FILE% up -d
if errorlevel 1 (
    echo [ERROR] Failed to start Flink cluster
    exit /b 1
)

REM Wait for Flink to be ready
echo Waiting for Flink to start...
timeout /t 10 /nobreak >nul

REM Run integration tests (single version)
echo.
echo Running integration tests against Flink %FULL_VERSION%...
echo.
go test -tags=integration -v -timeout 10m -count=1
set TEST_RESULT=%errorlevel%

REM Cleanup
echo.
echo Stopping Flink cluster...
docker compose -f testdata\%COMPOSE_FILE% down -v >nul 2>&1

REM Exit with test result
if %TEST_RESULT% equ 0 (
    echo.
    echo ========================================
    echo [SUCCESS] Flink %FULL_VERSION% tests passed!
    echo ========================================
    exit /b 0
) else (
    echo.
    echo ========================================
    echo [FAILED] Flink %FULL_VERSION% tests failed
    echo ========================================
    exit /b 1
)
