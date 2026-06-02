@echo off
REM uninstall.bat — hapus sirus-frista-agent. Dobel-klik.
powershell -NoProfile -ExecutionPolicy Bypass -File "%~dp0uninstall.ps1"
echo.
pause
