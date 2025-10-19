@echo off
setlocal enabledelayedexpansion

echo ========================================
echo   Oak Project - All Flink Versions Test
echo ========================================
echo.

REM Check if Docker is running
docker info >nul 2>&1
if errorlevel 1 (
    echo ERROR: Docker is not running
    echo Please start Docker Desktop and try again
    exit /b 1
)
echo Docker is running

echo.
echo This will test against ALL supported Flink versions:
echo   - Flink 1.18.1
echo   - Flink 1.19.1
echo   - Flink 1.20.0
echo   - Flink 2.0.1
echo   - Flink 2.1.0
echo.
echo WARNING: This will take 10-15 minutes to complete
echo Each version will start a Docker container, run tests, and clean up
echo.
pause

echo.
echo ========================================
echo Running Multi-Version Integration Tests
echo ========================================
echo.

cd oak-lib\flink\rest-api
go test -tags=integration_versions -v -timeout 30m -count=1

if errorlevel 1 (
    echo.
    echo ERROR: Some tests failed
    exit /b 1
)

echo.
echo ========================================
echo All Flink version tests passed!
echo ========================================
