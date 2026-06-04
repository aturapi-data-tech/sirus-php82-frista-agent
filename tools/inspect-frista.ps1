# inspect-frista.ps1 - diagnostik struktur jendela & kontrol FRISTA.
#
# TUJUAN: cari tahu FRISTA dibuat pakai framework apa (Delphi / .NET WinForms / VB6 / dll)
#         dan dapatkan handle + class name + control-id tiap kotak isian (Username,
#         Password, BPJS). Data ini dipakai untuk implementasi "isi kolom walau
#         FRISTA di-minimize" via PostMessage WM_SETTEXT atau UI Automation
#         (tanpa merebut fokus / tanpa SendKeys).
#
# CARA PAKAI (di PC yang ada FRISTA-nya):
#   1. Buka FRISTA sampai jendela LOGIN muncul (jangan login dulu).
#   2. Klik kanan file ini -> "Run with PowerShell"  (atau jalankan dari PowerShell)
#   3. Hasil tampil di layar DAN tersimpan ke: inspect-frista-output.txt (di folder ini)
#   4. Kirim isi file txt itu ke saya.
#   5. (Opsional) Ulangi setelah login, saat jendela UTAMA + kolom BPJS muncul.
#
# Script ini READ-ONLY: hanya membaca struktur jendela, tidak mengubah apa pun.

$ErrorActionPreference = 'Continue'
$OutFile = Join-Path (Split-Path -Parent $MyInvocation.MyCommand.Path) 'inspect-frista-output.txt'

# Bersihkan output lama
"" | Out-File -FilePath $OutFile -Encoding utf8

function Tee2([string]$msg) {
    Write-Host $msg
    $msg | Out-File -FilePath $OutFile -Append -Encoding utf8
}

# ---- P/Invoke Win32 untuk enumerasi child window ----
$sig = @'
using System;
using System.Text;
using System.Collections.Generic;
using System.Runtime.InteropServices;

public class Win32Enum {
    public delegate bool EnumChildProc(IntPtr hWnd, IntPtr lParam);

    [DllImport("user32.dll")]
    public static extern bool EnumChildWindows(IntPtr hWndParent, EnumChildProc lpEnumFunc, IntPtr lParam);
    [DllImport("user32.dll", CharSet=CharSet.Auto)]
    public static extern int GetClassName(IntPtr hWnd, StringBuilder lpClassName, int nMaxCount);
    [DllImport("user32.dll", CharSet=CharSet.Auto)]
    public static extern int GetWindowText(IntPtr hWnd, StringBuilder lpString, int nMaxCount);
    [DllImport("user32.dll")]
    public static extern int GetWindowTextLength(IntPtr hWnd);
    [DllImport("user32.dll")]
    public static extern int GetDlgCtrlID(IntPtr hWnd);
    [DllImport("user32.dll")]
    public static extern bool IsWindowVisible(IntPtr hWnd);

    public static List<string> Children = new List<string>();

    public static bool Collect(IntPtr hWnd, IntPtr lParam) {
        var cls = new StringBuilder(256);
        GetClassName(hWnd, cls, cls.Capacity);
        int len = GetWindowTextLength(hWnd);
        var txt = new StringBuilder(len + 1);
        GetWindowText(hWnd, txt, txt.Capacity);
        int id = GetDlgCtrlID(hWnd);
        bool vis = IsWindowVisible(hWnd);
        Children.Add(string.Format("  hwnd=0x{0:X8}  id={1,-6} vis={2,-5} class='{3}'  text='{4}'",
            hWnd.ToInt64(), id, vis, cls.ToString(), txt.ToString()));
        return true;
    }
    public static void Dump(IntPtr parent) {
        Children.Clear();
        EnumChildWindows(parent, Collect, IntPtr.Zero);
    }
}
'@
if (-not ([System.Management.Automation.PSTypeName]'Win32Enum').Type) {
    Add-Type -TypeDefinition $sig
}

Tee2 "================ INSPECT FRISTA ================"
Tee2 ("Waktu  : " + (Get-Date))
Tee2 ("Komputer: " + $env:COMPUTERNAME + "  User: " + $env:USERNAME)
Tee2 ""

# ---- 1) Daftar proses yang punya jendela ----
Tee2 "---- 1) PROSES DENGAN JENDELA (cari yang FRISTA) ----"
Get-Process | Where-Object { $_.MainWindowTitle -ne '' } | ForEach-Object {
    Tee2 ("  PID={0,-6} proc='{1}'  title='{2}'" -f $_.Id, $_.ProcessName, $_.MainWindowTitle)
}
Tee2 ""

# ---- 2) Untuk tiap jendela top-level yang judulnya mengandung 'frista', dump child controls ----
Tee2 "---- 2) STRUKTUR KONTROL (child windows) ----"
$targets = Get-Process | Where-Object { $_.MainWindowTitle -match 'frista' -and $_.MainWindowHandle -ne 0 }
if (-not $targets) {
    Tee2 "  (tidak ada jendela ber-judul 'frista' yang terdeteksi - pastikan FRISTA terbuka)"
} else {
    foreach ($p in $targets) {
        Tee2 ("  == Jendela: proc='{0}' title='{1}' hwnd=0x{2:X8} ==" -f $p.ProcessName, $p.MainWindowTitle, $p.MainWindowHandle.ToInt64())
        [Win32Enum]::Dump($p.MainWindowHandle)
        if ([Win32Enum]::Children.Count -eq 0) {
            Tee2 "    (tidak ada child window - kemungkinan WPF/elektron, perlu UI Automation di bawah)"
        }
        foreach ($line in [Win32Enum]::Children) { Tee2 $line }
        Tee2 ""
    }
}

# ---- 3) UI Automation tree (paling informatif untuk .NET / WPF) ----
Tee2 "---- 3) UI AUTOMATION TREE (Name / ControlType / AutomationId / value-able) ----"
try {
    Add-Type -AssemblyName UIAutomationClient
    Add-Type -AssemblyName UIAutomationTypes
    $root = [System.Windows.Automation.AutomationElement]::RootElement
    # ambil semua top-level window lalu filter judul mengandung frista
    $wins = $root.FindAll([System.Windows.Automation.TreeScope]::Children,
        [System.Windows.Automation.Condition]::TrueCondition)
    foreach ($w in $wins) {
        $name = $w.Current.Name
        if ($name -notmatch 'frista') { continue }
        Tee2 ("  == UIA Window: '{0}'  (class={1}) ==" -f $name, $w.Current.ClassName)
        $descendants = $w.FindAll([System.Windows.Automation.TreeScope]::Descendants,
            [System.Windows.Automation.Condition]::TrueCondition)
        foreach ($e in $descendants) {
            $c = $e.Current
            $hasValue = $false
            try { $null = $e.GetCurrentPattern([System.Windows.Automation.ValuePattern]::Pattern); $hasValue = $true } catch {}
            Tee2 ("    type={0,-14} valueable={1,-5} autoId='{2}' name='{3}' class='{4}'" -f `
                $c.ControlType.ProgrammaticName.Replace('ControlType.',''), $hasValue, $c.AutomationId, $c.Name, $c.ClassName)
        }
        Tee2 ""
    }
} catch {
    Tee2 ("  UI Automation gagal: " + $_.Exception.Message)
}

Tee2 "================ SELESAI ================"
Tee2 ("Output lengkap tersimpan di: " + $OutFile)
Write-Host ""
Write-Host "Kirim file ini ke saya: $OutFile" -ForegroundColor Green
Read-Host "Tekan ENTER untuk menutup"
