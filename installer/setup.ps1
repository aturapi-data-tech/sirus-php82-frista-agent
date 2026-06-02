# setup.ps1 — install sirus-frista-agent sebagai autostart user-session (TANPA admin).
#
# - Copy file ke %LOCALAPPDATA%\SirusFristaAgent
# - Daftar autostart via HKCU\...\Run (jalan tiap user login)
# - Langsung jalankan agent sekarang
#
# Kenapa user-session (bukan Windows Service)?
#   Agent mengetik ke jendela FRISTA via SendKeys → wajib di desktop user yang login.
#   Windows Service (session 0) tidak punya akses ke desktop.

$ErrorActionPreference = 'Stop'

$AppName   = 'SirusFristaAgent'
$InstallDir = Join-Path $env:LOCALAPPDATA $AppName
$ExeName   = 'sirus-frista-agent.exe'
$SrcDir    = Split-Path -Parent $MyInvocation.MyCommand.Path

Write-Host "=== Install $AppName ===" -ForegroundColor Cyan

# 1. Hentikan instance lama bila masih jalan
Get-Process -Name 'sirus-frista-agent' -ErrorAction SilentlyContinue | Stop-Process -Force -ErrorAction SilentlyContinue
Start-Sleep -Milliseconds 500

# 2. Copy file
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
Copy-Item (Join-Path $SrcDir $ExeName) (Join-Path $InstallDir $ExeName) -Force

# config.json: jangan timpa kalau sudah ada (biar setting user tidak hilang saat update)
$dstConfig = Join-Path $InstallDir 'config.json'
if (Test-Path $dstConfig) {
    Write-Host "config.json sudah ada — TIDAK ditimpa (cek manual bila ada field baru)." -ForegroundColor Yellow
} else {
    Copy-Item (Join-Path $SrcDir 'config.json') $dstConfig -Force
}

$ExePath = Join-Path $InstallDir $ExeName

# 3. Daftar autostart (HKCU Run — per user, tanpa admin)
$runKey = 'HKCU:\Software\Microsoft\Windows\CurrentVersion\Run'
Set-ItemProperty -Path $runKey -Name $AppName -Value "`"$ExePath`""
Write-Host "Autostart terdaftar di HKCU\...\Run" -ForegroundColor Green

# 4. Jalankan sekarang
Start-Process -FilePath $ExePath -WorkingDirectory $InstallDir
Start-Sleep -Seconds 1

Write-Host ""
Write-Host "INSTALL SUKSES." -ForegroundColor Green
Write-Host "  Lokasi : $InstallDir"
Write-Host "  Config : $dstConfig"
Write-Host ""
Write-Host "PENTING: edit config.json → isi username, password, fristaPath, lalu restart agent." -ForegroundColor Yellow

# Buka dashboard + folder config
try {
    $port = (Get-Content $dstConfig -Raw | ConvertFrom-Json).port
    Start-Process "http://localhost:$port"
} catch {}
Start-Process notepad.exe $dstConfig
