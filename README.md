# Sirus Frista Agent

Local agent (Go) yang membuka aplikasi **FRISTA** (Face Recognition BPJS Kesehatan) di PC pendaftaran, **login otomatis**, dan **mengetik nomor BPJS peserta** — dipicu dari tombol "Scan Wajah" di web Sirus (Daftar RJ). Tujuannya: petugas tidak perlu buka FRISTA manual + ketik ulang kartu BPJS tiap pasien datang.

Saudara dari [`sirus-print-agent`](../sirus-php82-print-agent), pola arsitektur sama (Go HTTP agent di localhost, dipanggil browser via fetch).

## Konsep

```
[ Browser Sirus / Daftar RJ ] ──fetch──▶ [ http://localhost:9998/buka-frista ]
        (klik "Scan Wajah")                        │  body: { bpjsId }
                                                    ▼ (sirus-frista-agent.exe)
                                          PowerShell + SendKeys:
                                            1. Start frista.exe
                                            2. ketik username + password → Login
                                            3. ketik ID BPJS peserta → Enter
                                                    │
                                                    ▼
                                          Jendela FRISTA siap scan wajah
```

## ⚠️ Beda penting dengan print-agent: BUKAN Windows Service

`print-agent` cukup kirim file ke spooler → aman jalan sebagai service (session 0).

`frista-agent` harus **mengetik ke jendela GUI** via SendKeys → **wajib jalan di sesi desktop user yang login**. Windows Service (session 0) terisolasi dari desktop dan tidak bisa. Karena itu installer mendaftarkannya sebagai **autostart user-session** (`HKCU\...\Run`), bukan nssm service. Tidak butuh admin.

## Stack

- **Go 1.18+** — single-binary, no runtime
- **PowerShell + System.Windows.Forms.SendKeys** — built-in Windows, tanpa dependency tambahan (mis. AutoHotkey)

## Endpoint

| Method | Path | Body | Keterangan |
|---|---|---|---|
| GET | `/` | — | Dashboard debug |
| GET | `/health` | — | Liveness check |
| POST | `/buka-frista` | `{ "bpjsId": "0001234567890" }` | Buka FRISTA + login + ketik BPJS. `bpjsId` boleh kosong (cuma buka+login). |

## config.json

```json
{
    "port": 9998,
    "fristaPath": "C:\\frista\\frista.exe",
    "username": "ISIKAN_USERNAME",
    "password": "ISIKAN_PASSWORD",
    "allowedOrigins": ["http://localhost", "http://sirus.local"],
    "loginWindowTitle": "Login Frista",
    "mainWindowTitle": "Frista",
    "launchWaitMs": 12000,
    "stepDelayMs": 300,
    "submitAfterBpjs": true
}
```

| Field | Fungsi |
|---|---|
| `fristaPath` | Lokasi `frista.exe` |
| `username` / `password` | Kredensial login, diketik otomatis |
| `loginWindowTitle` | Prefix judul jendela login (utk fokus window). Dari screenshot: `Login Frista` |
| `mainWindowTitle` | Prefix judul jendela utama setelah login (tempat input BPJS) |
| `launchWaitMs` | Maks tunggu jendela muncul |
| `stepDelayMs` | Jeda antar ketik (naikkan bila field belum sempat fokus) |
| `submitAfterBpjs` | Tekan Enter setelah ketik BPJS |

## Build

```bash
./build-windows.sh    # hasil di dist/ — copy ke PC, dobel-klik setup.bat
```

## Integrasi web Sirus

Helper JS global (mis. di `resources/js/app.js`):

```js
window.bukaFrista = async function (bpjsId) {
    try {
        const res = await fetch('http://localhost:9998/buka-frista', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ bpjsId }),
        });
        const data = await res.json();
        Livewire.dispatch('toast', { type: data.ok ? 'success' : 'error', message: data.msg });
    } catch (e) {
        Livewire.dispatch('toast', { type: 'error', message: 'Frista agent tidak aktif di PC ini. Hubungi IT.' });
    }
};
```

Tombol di baris pasien Daftar RJ:

```blade
<button type="button" x-data x-on:click="bukaFrista(@js($pasien->nokartu_bpjs))">
    Scan Wajah
</button>
```

## Status / TODO

- [x] Buka FRISTA + auto-login (username/password) + ketik ID BPJS
- [ ] **Perlu diverifikasi live**: alur layar SETELAH login — judul jendela utama & field input BPJS (apakah fokus otomatis, perlu klik/Tab dulu, perlu Enter). Sesuaikan `mainWindowTitle` / `stepDelayMs` / `submitAfterBpjs` setelah lihat layar asli.
- [ ] Opsi upgrade ke AutoHotkey/UIAutomation bila SendKeys kurang andal (fokus field).
