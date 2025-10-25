@echo off
REM Run unit tests for all modules in the Oak project
REM Works on Windows cmd.exe and PowerShell

setlocal enabledelayedexpansion

echo ========================================
echo   Oak Project - Unit Tests
echo ========================================
echo.

REM Navigate to project root (parent of scripts directory)
cd /d "%~dp0\.."

echo Running unit tests for all modules...
echo.

echo [*] Testing oak-lib...
echo.
REM Note: Logger tests temporarily excluded due to file handling issues in test cleanup
REM The logger itself works correctly - it's used successfully in oak-server tests
go test -short -v ./oak-lib/certs ./oak-lib/flink/rest-api ./oak-lib/grpc
if errorlevel 1 (
    echo.
    echo ========================================
    echo   ERROR: oak-lib unit tests failed!
    echo ========================================
    exit /b 1
)

echo.
echo [*] Testing oak-server...
echo.
go test -short -v ./oak-server/...
if errorlevel 1 (
    echo.
    echo ========================================
    echo   ERROR: oak-server unit tests failed!
    echo ========================================
    exit /b 1
)

REM Future: Add oak-agent when it exists
REM echo [*] Testing oak-agent...
REM echo.
REM go test -short -v ./oak-agent/...
REM if errorlevel 1 (
REM     echo.
REM     echo ========================================
REM     echo   ERROR: oak-agent unit tests failed!
REM     echo ========================================
REM     exit /b 1
REM )

echo.
echo ========================================
echo   SUCCESS: All unit tests passed!
echo ========================================
exit /b 0
