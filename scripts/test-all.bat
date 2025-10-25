@echo off
REM Run all tests for Oak project
REM Usage: test-all.bat [--full-compat]
REM   --full-compat: Include multi-version compatibility tests (10-15 min)
REM
REM Requirements for integration tests:
REM   - Docker must be running
REM   - Port 8081 must be available (for Flink)

setlocal enabledelayedexpansion

REM Check for --full-compat flag
set "FULL_COMPAT=0"
if "%~1"=="--full-compat" set "FULL_COMPAT=1"

if %FULL_COMPAT%==1 (
    echo ========================================
    echo   Oak Project - FULL COMPATIBILITY TEST
    echo   ^(Unit + Integration + Multi-Version^)
    echo ========================================
) else (
    echo ========================================
    echo   Oak Project - All Tests
    echo   ^(Unit + Integration^)
    echo ========================================
    echo.
    echo TIP: Use --full-compat to test all Flink versions ^(1.18-2.1^)
)
echo.

REM Navigate to project root
cd /d "%~dp0\.."

REM Run unit tests
echo [1/3] Running unit tests...
echo.
call "%~dp0test-unit.bat"
if errorlevel 1 (
    echo.
    echo [ERROR] Unit tests failed. Stopping.
    exit /b 1
)

echo.
echo.

REM Run integration tests (single version - Flink 2.1)
echo [2/3] Running integration tests ^(Flink 2.1^)...
echo.
call "%~dp0test-integration.bat"
if errorlevel 1 (
    echo.
    echo [ERROR] Integration tests failed.
    exit /b 1
)

echo.
echo.

REM Optionally run multi-version compatibility tests
if %FULL_COMPAT%==1 (
    echo [3/3] Running multi-version compatibility tests...
    echo This will test against ALL Flink versions 1.18-2.1 ^(~10-15 min^)
    echo.
    call "%~dp0test-all-flink-versions.bat"
    if errorlevel 1 (
        echo.
        echo [ERROR] Multi-version compatibility tests failed.
        exit /b 1
    )

    REM Full compat summary
    echo.
    echo ========================================
    echo   FULL COMPATIBILITY TEST SUMMARY
    echo ========================================
    echo [SUCCESS] Unit tests: PASSED
    echo [SUCCESS] Integration tests: PASSED
    echo [SUCCESS] Multi-version tests: PASSED
    echo.
    echo All compatibility tests passed!
) else (
    REM Standard summary
    echo.
    echo ========================================
    echo   Test Summary
    echo ========================================
    echo [SUCCESS] Unit tests: PASSED
    echo [SUCCESS] Integration tests: PASSED
    echo.
    echo All tests passed!
    echo.
    echo Run with --full-compat to test all Flink versions ^(1.18-2.1^)
)

exit /b 0
