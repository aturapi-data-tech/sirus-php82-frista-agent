@echo off
REM setup.bat — installer sirus-frista-agent (user-session, TANPA admin)
REM Dobel-klik file ini.

echo ================================================
echo   Install Sirus Frista Agent (autostart user)
echo ================================================
echo.

powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0setup.ps1"

echo.
pause
