@echo off
REM Run tests with coverage report for Oak project
REM Works on Windows cmd.exe and PowerShell
REM
REM Generates HTML coverage report and displays coverage percentages

setlocal enabledelayedexpansion

echo ========================================
echo   Oak Project - Coverage Report
echo ========================================
echo.

REM Navigate to project root
cd /d "%~dp0\.."

REM Create coverage directory
if not exist coverage mkdir coverage

echo Running tests with coverage (unit tests only)...
echo.

REM Run unit tests with coverage for oak-lib
go test -short -v -coverprofile=coverage\coverage.out -covermode=atomic ./oak-lib/...
if errorlevel 1 (
    echo.
    echo ========================================
    echo   ERROR: Tests failed!
    echo ========================================
    exit /b 1
)

echo.
echo ========================================
echo   SUCCESS: Tests completed
echo ========================================
echo.

REM Display coverage summary
echo Coverage Summary:
go tool cover -func=coverage\coverage.out | findstr "total:"
echo.

REM Generate HTML report
echo Generating HTML coverage report...
go tool cover -html=coverage\coverage.out -o coverage\coverage.html

echo.
echo ========================================
echo   Coverage report generated
echo ========================================
echo.
echo To view the report, open:
echo   coverage\coverage.html

exit /b 0
