@echo off
REM Run all tests (unit + integration) for Oak project
REM Works on Windows cmd.exe and PowerShell
REM
REM Requirements for integration tests:
REM   - Docker must be running
REM   - Port 8081 must be available (for Flink)

setlocal enabledelayedexpansion

echo ========================================
echo   Oak Project - All Tests
echo ========================================
echo.

REM Navigate to project root
cd /d "%~dp0\.."

REM Run unit tests
echo [1/2] Running unit tests...
echo.
call "%~dp0test-unit.bat"
if errorlevel 1 (
    echo.
    echo [ERROR] Unit tests failed. Stopping.
    exit /b 1
)

echo.
echo.

REM Run integration tests
echo [2/2] Running integration tests...
echo.
call "%~dp0test-integration.bat"
if errorlevel 1 (
    echo.
    echo [ERROR] Integration tests failed.
    exit /b 1
)

REM Summary
echo.
echo ========================================
echo   Test Summary
echo ========================================
echo [SUCCESS] Unit tests: PASSED
echo [SUCCESS] Integration tests: PASSED
echo.
echo All tests passed!
exit /b 0
