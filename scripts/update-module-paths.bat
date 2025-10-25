@echo off
REM update-module-paths.bat
REM Updates all Go module paths when forking the repository
REM Usage: scripts\update-module-paths.bat github.com/yourorg/yourrepo

setlocal enabledelayedexpansion

if "%~1"=="" (
    echo Usage: %0 ^<new-module-path^>
    echo Example: %0 github.com/mycompany/oak-flink
    echo Example: %0 gitlab.com/myorg/flink-platform
    exit /b 1
)

set "OLD_PATH=github.com/oakproject-flink/oak-flink"
set "NEW_PATH=%~1"

echo Updating module paths from %OLD_PATH% to %NEW_PATH%
echo ==================================================

REM Create backup directory
for /f "tokens=1-4 delims=/ " %%a in ('date /t') do set "BACKUP_DATE=%%c%%a%%b"
for /f "tokens=1-2 delims=: " %%a in ('time /t') do set "BACKUP_TIME=%%a%%b"
set "BACKUP_DIR=.backup-%BACKUP_DATE%-%BACKUP_TIME%"

echo Creating backup in %BACKUP_DIR%
mkdir "%BACKUP_DIR%" 2>nul

REM Find and update all relevant files
echo Updating files...

REM Update go.mod files
for /r %%f in (go.mod) do (
    if not "%%~pf"=="%CD%\node_modules\" (
        if not "%%~pf"=="%CD%\vendor\" (
            REM Create backup
            xcopy "%%f" "%BACKUP_DIR%\%%~pf" /Y /Q /I >nul 2>&1

            REM Update file using PowerShell
            powershell -Command "(Get-Content '%%f' -Raw) -replace '%OLD_PATH:\=\\%', '%NEW_PATH:\=\\%' | Set-Content '%%f' -NoNewline"
            echo   ✓ %%f
        )
    )
)

REM Update .go files
for /r %%f in (*.go) do (
    if not "%%~pf"=="%CD%\node_modules\" (
        if not "%%~pf"=="%CD%\vendor\" (
            REM Create backup
            xcopy "%%f" "%BACKUP_DIR%\%%~pf" /Y /Q /I >nul 2>&1

            REM Update file using PowerShell
            powershell -Command "(Get-Content '%%f' -Raw) -replace '%OLD_PATH:\=\\%', '%NEW_PATH:\=\\%' | Set-Content '%%f' -NoNewline"
            echo   ✓ %%f
        )
    )
)

REM Update .proto files
for /r %%f in (*.proto) do (
    if not "%%~pf"=="%CD%\node_modules\" (
        if not "%%~pf"=="%CD%\vendor\" (
            REM Create backup
            xcopy "%%f" "%BACKUP_DIR%\%%~pf" /Y /Q /I >nul 2>&1

            REM Update file using PowerShell
            powershell -Command "(Get-Content '%%f' -Raw) -replace '%OLD_PATH:\=\\%', '%NEW_PATH:\=\\%' | Set-Content '%%f' -NoNewline"
            echo   ✓ %%f
        )
    )
)

echo.
echo ==================================================
echo ✅ Module paths updated successfully!
echo.
echo Next steps:
echo 1. Run: go work sync
echo 2. Run: go mod tidy in each module directory
echo 3. Test: go build ./...
echo.
echo Backup created in: %BACKUP_DIR%
echo If anything goes wrong, restore from backup

endlocal
