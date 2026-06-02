package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// jalankanFrista melakukan otomatisasi:
//   1. Launch frista.exe bila belum berjalan
//   2. Tunggu jendela login muncul, ketik username + password, tekan Login
//   3. Tunggu jendela utama, ketik ID BPJS peserta (opsional Enter)
//
// Implementasi pakai PowerShell + System.Windows.Forms.SendKeys karena:
//   - Built-in Windows (tanpa dependency tambahan)
//   - Bisa fokus jendela target via WScript.Shell AppActivate
//
// CATATAN PENTING: SendKeys mengetik ke jendela yang sedang fokus, jadi agent ini
// HARUS berjalan di sesi desktop user yang login (Startup / Task Scheduler "at logon"),
// BUKAN sebagai Windows Service session-0 (tidak punya akses desktop).
func jalankanFrista(config *Config, bpjsID string) error {
	script := buildPowerShellScript(config, bpjsID)

	// Tulis script ke temp .ps1 (lebih aman dari -Command untuk script panjang/escape)
	tmp := filepath.Join(os.TempDir(),
		"frista-agent-"+time.Now().Format("20060102-150405")+".ps1")
	if err := os.WriteFile(tmp, []byte(script), 0644); err != nil {
		return fmt.Errorf("tulis script temp gagal: %w", err)
	}
	defer func() {
		go func(p string) {
			time.Sleep(2 * time.Second)
			os.Remove(p)
		}(tmp)
	}()

	cmd := exec.Command("powershell",
		"-NoProfile",
		"-ExecutionPolicy", "Bypass",
		"-File", tmp,
	)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell gagal: %v — output: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// buildPowerShellScript menyusun script otomatisasi.
// Nilai dari config diselipkan sebagai variabel PowerShell (di-escape single-quote).
func buildPowerShellScript(config *Config, bpjsID string) string {
	submit := "$false"
	if config.SubmitAfterBpjs {
		submit = "$true"
	}

	return fmt.Sprintf(`
$ErrorActionPreference = 'Stop'
Add-Type -AssemblyName System.Windows.Forms

$FristaPath  = '%s'
$User        = '%s'
$Pass        = '%s'
$BpjsId      = '%s'
$LoginTitle  = '%s'
$MainTitle   = '%s'
$LaunchWait  = %d
$StepDelay   = %d
$SubmitBpjs  = %s
$ProcName    = [System.IO.Path]::GetFileNameWithoutExtension($FristaPath)

# Escape karakter spesial SendKeys: + ^ %% ~ ( ) { } [ ]
function Esc([string]$s) {
    if ($null -eq $s) { return '' }
    $sb = New-Object System.Text.StringBuilder
    foreach ($c in $s.ToCharArray()) {
        if ('+^%%~(){}[]'.IndexOf($c) -ge 0) { [void]$sb.Append('{' + $c + '}') }
        else { [void]$sb.Append($c) }
    }
    return $sb.ToString()
}

$shell = New-Object -ComObject WScript.Shell

# Tunggu sebuah jendela (by title prefix) bisa diaktifkan. Return $true bila ketemu.
function WaitWindow([string]$title, [int]$timeoutMs) {
    $elapsed = 0
    while ($elapsed -lt $timeoutMs) {
        if ($shell.AppActivate($title)) { Start-Sleep -Milliseconds 250; return $true }
        Start-Sleep -Milliseconds 300
        $elapsed += 300
    }
    return $false
}

# 1. Launch frista bila belum jalan
$running = Get-Process -Name $ProcName -ErrorAction SilentlyContinue
if (-not $running) {
    if (-not (Test-Path $FristaPath)) { Write-Error "frista.exe tidak ditemukan: $FristaPath"; exit 2 }
    Start-Process -FilePath $FristaPath -WorkingDirectory ([System.IO.Path]::GetDirectoryName($FristaPath))
}

# 2. Login — hanya bila jendela login muncul (kalau sudah login, dilewati)
if (WaitWindow $LoginTitle $LaunchWait) {
    [void]$shell.AppActivate($LoginTitle)
    Start-Sleep -Milliseconds $StepDelay
    if ($User.Length -gt 0) {
        [System.Windows.Forms.SendKeys]::SendWait((Esc $User))
        Start-Sleep -Milliseconds $StepDelay
        [System.Windows.Forms.SendKeys]::SendWait('{TAB}')
        Start-Sleep -Milliseconds $StepDelay
        [System.Windows.Forms.SendKeys]::SendWait((Esc $Pass))
        Start-Sleep -Milliseconds $StepDelay
        # Form FRISTA tidak submit lewat Enter di field password (tombol Login
        # bukan AcceptButton). Jadi: TAB ke tombol Login, lalu SPASI untuk klik
        # (tombol yang fokus diaktifkan dengan Space, bukan Enter).
        [System.Windows.Forms.SendKeys]::SendWait('{TAB}')
        Start-Sleep -Milliseconds $StepDelay
        [System.Windows.Forms.SendKeys]::SendWait(' ')
    }
}

# 3. Ketik ID BPJS di jendela utama (bila ada ID & jendela utama muncul)
if ($BpjsId.Length -gt 0) {
    if (WaitWindow $MainTitle $LaunchWait) {
        [void]$shell.AppActivate($MainTitle)
        Start-Sleep -Milliseconds $StepDelay
        [System.Windows.Forms.SendKeys]::SendWait((Esc $BpjsId))
        if ($SubmitBpjs) {
            Start-Sleep -Milliseconds $StepDelay
            [System.Windows.Forms.SendKeys]::SendWait('{ENTER}')
        }
    } else {
        Write-Output 'WARN: jendela utama Frista tidak terdeteksi untuk input BPJS'
    }
}

Write-Output 'OK'
`,
		psQuote(config.FristaPath),
		psQuote(config.Username),
		psQuote(config.Password),
		psQuote(bpjsID),
		psQuote(config.LoginWindowTitle),
		psQuote(config.MainWindowTitle),
		config.LaunchWaitMs,
		config.StepDelayMs,
		submit,
	)
}

// psQuote meng-escape string untuk dipakai di dalam single-quote PowerShell.
// Di PowerShell, single-quote di-escape dengan menggandakannya ('').
func psQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}
