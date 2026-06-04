package main

import (
	"encoding/json"
	"os"
)

// Config = struktur config.json yang dibaca saat startup.
//
// Field:
//   Port             : port HTTP local agent (default 9998)
//   FristaPath       : path absolute ke frista.exe
//   Username         : username login Frista (diketik otomatis)
//   Password         : password login Frista (diketik otomatis)
//   AllowedOrigins   : whitelist origin browser yang boleh akses agent (CORS)
//   LoginWindowTitle : judul (prefix) jendela login Frista — utk AppActivate
//   MainWindowTitle  : judul (prefix) jendela utama Frista setelah login
//   LaunchWaitMs     : maksimal tunggu jendela muncul setelah launch (ms)
//   LoginProbeMs     : tunggu jendela login saat FRISTA SUDAH jalan (ms). Pendek
//                      saja — kalau jendela login tidak muncul secepat ini, berarti
//                      sudah login, jadi langsung lanjut (tidak buang waktu nunggu
//                      penuh LaunchWaitMs untuk jendela login yang tidak ada).
//   StepDelayMs      : jeda antar langkah ketik (ms) — biar field sempat fokus
//   SubmitAfterBpjs  : tekan Enter setelah mengetik ID BPJS peserta
type Config struct {
	Port             int      `json:"port"`
	FristaPath       string   `json:"fristaPath"`
	Username         string   `json:"username"`
	Password         string   `json:"password"`
	AllowedOrigins   []string `json:"allowedOrigins"`
	LoginWindowTitle string   `json:"loginWindowTitle"`
	MainWindowTitle  string   `json:"mainWindowTitle"`
	LaunchWaitMs     int      `json:"launchWaitMs"`
	LoginProbeMs     int      `json:"loginProbeMs"`
	StepDelayMs      int      `json:"stepDelayMs"`
	SubmitAfterBpjs  bool     `json:"submitAfterBpjs"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c Config
	if err := json.Unmarshal(data, &c); err != nil {
		return nil, err
	}

	// Default
	if c.Port == 0 {
		c.Port = 9998
	}
	if c.LoginWindowTitle == "" {
		c.LoginWindowTitle = "Login Frista"
	}
	if c.MainWindowTitle == "" {
		c.MainWindowTitle = "Frista"
	}
	if c.LaunchWaitMs == 0 {
		c.LaunchWaitMs = 12000
	}
	if c.LoginProbeMs == 0 {
		c.LoginProbeMs = 1500
	}
	if c.StepDelayMs == 0 {
		c.StepDelayMs = 180
	}
	return &c, nil
}
