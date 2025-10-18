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
go test -short -v ./oak-lib/...
if errorlevel 1 (
    echo.
    echo ========================================
    echo   ERROR: Unit tests failed!
    echo ========================================
    exit /b 1
)

REM Future: Add oak-server and oak-sidecar when they exist
REM echo [*] Testing oak-server...
REM echo.
REM go test -short -v ./oak-server/...
REM if errorlevel 1 (
REM     echo.
REM     echo ========================================
REM     echo   ERROR: Unit tests failed!
REM     echo ========================================
REM     exit /b 1
REM )

echo.
echo ========================================
echo   SUCCESS: All unit tests passed!
echo ========================================
exit /b 0
