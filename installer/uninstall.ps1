# uninstall.ps1 — hapus autostart + file sirus-frista-agent.
$ErrorActionPreference = 'SilentlyContinue'

$AppName    = 'SirusFristaAgent'
$InstallDir = Join-Path $env:LOCALAPPDATA $AppName

Write-Host "=== Uninstall $AppName ===" -ForegroundColor Cyan

# Stop proses
Get-Process -Name 'sirus-frista-agent' -ErrorAction SilentlyContinue | Stop-Process -Force
Start-Sleep -Milliseconds 500

# Hapus autostart
$runKey = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run'
Remove-ItemProperty -Path $runKey -Name $AppName -ErrorAction SilentlyContinue
Write-Host "Autostart dihapus." -ForegroundColor Green

# Hapus folder
if (Test-Path $InstallDir) {
    Remove-Item -Recurse -Force $InstallDir
    Write-Host "Folder $InstallDir dihapus." -ForegroundColor Green
}

Write-Host "UNINSTALL SELESAI." -ForegroundColor Green
