package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"
)

const (
	agentName    = "sirus-frista-agent"
	agentVersion = "1.0.0"
)

func main() {
	// Lokasi config relatif terhadap binary (working dir bisa beda dengan letak .exe,
	// mis. saat dijalankan via shortcut Startup / Task Scheduler).
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("cannot get exe path: %v", err)
	}
	exeDir := filepath.Dir(exePath)
	configPath := filepath.Join(exeDir, "config.json")

	// Saat dev, fallback ke ./config.json
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configPath = "config.json"
	}

	// Binary di-build dengan -H windowsgui (tanpa console), jadi stdout/stderr
	// tidak ke mana-mana. Alihkan log ke file di samping .exe supaya tetap bisa
	// di-debug di lapangan. Append (tidak hapus log lama). Bila gagal buka file,
	// biarkan log default (mis. saat dev di terminal).
	logPath := filepath.Join(exeDir, "agent.log")
	if lf, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		log.SetOutput(lf)
		defer lf.Close()
	}
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)

	config, err := loadConfig(configPath)
	if err != nil {
		log.Fatalf("load config gagal: %v", err)
	}
	log.Printf("Config loaded from %s — frista: %s", configPath, config.FristaPath)

	// Routing
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleIndex(config))
	mux.HandleFunc("/health", handleHealth)
	mux.HandleFunc("/buka-frista", handleBukaFrista(config))

	handler := logMiddleware(corsMiddleware(config.AllowedOrigins, mux))

	addr := fmt.Sprintf("127.0.0.1:%d", config.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 60 * time.Second, // automation GUI bisa makan waktu (tunggu jendela)
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	go func() {
		log.Printf("%s v%s listening at http://%s", agentName, agentVersion, addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Graceful shutdown via signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
	log.Println("Bye.")
}
