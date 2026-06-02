package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

type APIResponse struct {
	OK  bool   `json:"ok"`
	Msg string `json:"msg,omitempty"`
}

type BukaRequest struct {
	BpjsID string `json:"bpjsId"` // nomor kartu BPJS peserta (opsional — kosong = cuma buka+login)
}

// GET /health — liveness check
func handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":      true,
		"agent":   agentName,
		"version": agentVersion,
		"time":    time.Now().Format(time.RFC3339),
	})
}

// POST /buka-frista — buka FRISTA, auto-login, ketik ID BPJS peserta
//
// Body JSON: { "bpjsId": "0001234567890" }
func handleBukaFrista(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			respondError(w, http.StatusMethodNotAllowed, "POST only")
			return
		}

		var req BukaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
			return
		}

		bpjsID := sanitizeBpjs(req.BpjsID)

		log.Printf("[BUKA-FRISTA] bpjsId=%q", bpjsID)

		if err := jalankanFrista(config, bpjsID); err != nil {
			respondError(w, http.StatusInternalServerError, err.Error())
			return
		}

		msg := "FRISTA dibuka & login otomatis"
		if bpjsID != "" {
			msg = fmt.Sprintf("FRISTA dibuka, login, dan ID BPJS %s dimasukkan", bpjsID)
		}
		writeJSON(w, http.StatusOK, APIResponse{OK: true, Msg: msg})
	}
}

// sanitizeBpjs membersihkan nomor BPJS: hanya sisakan digit (anti-injeksi keystroke).
func sanitizeBpjs(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// GET / — dashboard debug sederhana
func handleIndex(config *Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprintf(w, indexHTML,
			agentName, agentVersion,
			config.Port,
			config.Port,
			config.FristaPath,
			config.LoginWindowTitle,
			config.MainWindowTitle,
			config.Port, // sample fetch URL
		)
	}
}

const indexHTML = `<!doctype html>
<html lang="id"><head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Sirus Frista Agent</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; }
body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
       background: #f5f7fa; color: #1f2937; padding: 2rem; line-height: 1.5; }
.container { max-width: 820px; margin: 0 auto; }
h1 { font-size: 1.6rem; margin-bottom: 0.25rem; color: #0f172a; }
.subtitle { color: #64748b; margin-bottom: 1.5rem; }
.card { background: #fff; border: 1px solid #e2e8f0; border-radius: 12px;
        padding: 1.25rem; margin-bottom: 1rem; box-shadow: 0 1px 3px rgba(0,0,0,0.05); }
h2 { font-size: 1.05rem; margin-bottom: 0.75rem; color: #334155; }
.kv { display: grid; grid-template-columns: 200px 1fr; gap: 0.5rem 1rem; font-size: 0.9rem; }
.kv .k { color: #64748b; }
.kv .v { font-family: ui-monospace, Consolas, monospace; word-break: break-all; }
.badge { display: inline-block; padding: 0.15rem 0.55rem; border-radius: 999px;
         font-size: 0.75rem; font-weight: 600; background: #dcfce7; color: #15803d; }
button { font-family: inherit; cursor: pointer; padding: 0.45rem 0.9rem;
         border: 1px solid #cbd5e1; background: #fff; border-radius: 6px;
         font-size: 0.85rem; color: #334155; }
button.primary { background: #0d9488; border-color: #0d9488; color: #fff; }
button.primary:hover { background: #0f766e; }
input[type=text] { font-family: ui-monospace, monospace; font-size: 0.9rem;
                   padding: 0.45rem 0.6rem; border: 1px solid #cbd5e1; border-radius: 6px; width: 100%%; }
label { display: block; font-size: 0.8rem; font-weight: 600; margin: 0.5rem 0 0.25rem; color: #334155; }
.actions { display: flex; gap: 0.5rem; margin-top: 0.75rem; }
.output { background: #0f172a; color: #e2e8f0; padding: 1rem; border-radius: 8px;
          font-size: 0.78rem; min-height: 70px; font-family: ui-monospace, monospace;
          white-space: pre-wrap; word-break: break-all; }
.note { background: #fef3c7; border-left: 4px solid #f59e0b; padding: 0.75rem 1rem;
        border-radius: 6px; font-size: 0.85rem; color: #78350f; margin-bottom: 1rem; }
pre { background: #0f172a; color: #e2e8f0; padding: 1rem; border-radius: 8px;
      font-size: 0.78rem; overflow-x: auto; font-family: ui-monospace, monospace; }
</style>
</head><body><div class="container">

<h1>%s <span class="badge">v%s — RUNNING</span></h1>
<p class="subtitle">Local agent — listening at <code>http://127.0.0.1:%d</code></p>

<div class="note">
<strong>Halaman ini cuma dashboard debug.</strong> Agent dipanggil dari Sirus via JavaScript fetch.
Agent WAJIB jalan di sesi desktop user yang login (bukan Windows Service) supaya bisa mengetik ke jendela FRISTA.
</div>

<div class="card">
<h2>Konfigurasi Aktif</h2>
<div class="kv">
<div class="k">Port</div><div class="v">%d</div>
<div class="k">Frista Path</div><div class="v">%s</div>
<div class="k">Login Window Title</div><div class="v">%s</div>
<div class="k">Main Window Title</div><div class="v">%s</div>
</div>
</div>

<div class="card">
<h2>Test Buka FRISTA</h2>
<label>ID BPJS Peserta (kosongkan untuk cuma buka + login)</label>
<input type="text" id="bpjsId" placeholder="0001234567890">
<div class="actions">
<button class="primary" onclick="testBuka()">Buka FRISTA Sekarang</button>
<button onclick="callApi('GET','/health')">Cek Health</button>
</div>
</div>

<div class="card">
<h2>Output</h2>
<div class="output" id="output">(klik tombol di atas untuk lihat response)</div>
</div>

<div class="card">
<h2>Sample Integrasi — Sirus (Daftar RJ)</h2>
<pre>window.bukaFrista = async function (bpjsId) {
    try {
        const res = await fetch('http://localhost:%d/buka-frista', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ bpjsId }),
        });
        const data = await res.json();
        Livewire.dispatch('toast', {
            type: data.ok ? 'success' : 'error', message: data.msg });
    } catch (e) {
        Livewire.dispatch('toast', { type: 'error',
            message: 'Frista agent tidak aktif di PC ini. Hubungi IT.' });
    }
};

// Tombol "Scan Wajah" di baris pasien Daftar RJ:
// &lt;button x-on:click="bukaFrista(@js($pasien-&gt;nokartu_bpjs))"&gt;Scan Wajah&lt;/button&gt;</pre>
</div>

</div>
<script>
async function callApi(method, path, body) {
    const out = document.getElementById('output');
    out.textContent = method + ' ' + path + ' ...';
    try {
        const opts = { method };
        if (body) { opts.headers = { 'Content-Type': 'application/json' }; opts.body = JSON.stringify(body); }
        const res = await fetch(path, opts);
        const data = await res.json();
        out.textContent = '[' + res.status + '] ' + JSON.stringify(data, null, 2);
    } catch (e) { out.textContent = 'ERROR: ' + e.message; }
}
function testBuka() {
    const bpjsId = document.getElementById('bpjsId').value.trim();
    callApi('POST', '/buka-frista', { bpjsId });
}
</script>
</body></html>
`

/* =============================== HELPERS =============================== */

func respondError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, APIResponse{OK: false, Msg: msg})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

/* =============================== MIDDLEWARE =============================== */

func corsMiddleware(allowed []string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		for _, a := range allowed {
			if a == "*" || a == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				break
			}
		}
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Max-Age", "3600")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func logMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s (%v)", r.RemoteAddr, r.Method, r.URL.Path, time.Since(start))
	})
}
