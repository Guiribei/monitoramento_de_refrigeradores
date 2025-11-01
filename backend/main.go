package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"

	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/tuya"
	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/store"
)



func getenvInt(key string, def int) int {
	if s := os.Getenv(key); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			return v
		}
	}
	return def
}

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT obrigatório")
	}

	// janela (s) será usada como período do coletor também
	windowSec := getenvInt("RATE_WINDOW_SECONDS", 3600)

	// origem p/ CORS (se quiser abrir p/ seu domínio)
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")

		dataDir := os.Getenv("DATA_DIR")
	if dataDir == "" {
		if os.Geteuid() == 0 {
			dataDir = "/var/lib/tuya-backend"
		} else {
			dataDir = "./data"
	    }
	}
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		log.Fatalf("não consegui criar DATA_DIR: %v", err)
	}
	dbPath := filepath.Join(dataDir, "tuya.db")

	st, err := store.Open(dbPath)
	if err != nil {
		log.Fatalf("erro abrindo sqlite: %v", err)
	}
	defer st.Close()

	tc, err := tuya.NewTuyaClientFromEnv()
	if err != nil {
		log.Fatalf("config inválida: %v", err)
	}

	// Coletor: faz 1x no start + a cada windowSec
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		runCollect(ctx, tc, st) // primeira tentativa
		tick := time.NewTicker(time.Duration(windowSec) * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				runCollect(ctx, tc, st)
			}
		}
	}()

	mux := http.NewServeMux()

	// Health
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	// Endpoint público → lê do SQLite (NÃO chama Tuya)
	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, jsonErr{OK: false, Error: "method_not_allowed"})
			return
		}
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Vary", "Origin")
		}

		snap, err := st.GetLatest(r.Context())
		if err != nil {
			writeJSON(w, http.StatusServiceUnavailable, jsonErr{
				OK:          false,
				Error:       "no_data",
				Description: "Ainda não há dados coletados. Tente novamente em instantes.",
			})
			return
		}

		// devolve exatamente o JSON bruto da Tuya (como antes)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Data-Age-ms", strconv.FormatInt(time.Now().UnixMilli()-snap.FetchedAtMs, 10))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(snap.RawJSON)
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           logMiddleware(securityHeaders(mux)),
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    8 << 10,
	}

	log.Printf("listening on :%s", port)
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("server error: %v", err)
	}
}

// Faz a coleta (chama Tuya) e persiste SOMENTE se status 200 e JSON
func runCollect(ctx context.Context, tc *tuya.TuyaClient, st *store.Store) {
	ctx, cancel := context.WithTimeout(ctx, 12*time.Second)
	defer cancel()

	status, body, err := tc.GetDevice(ctx)
	if err != nil {
		log.Printf("[collector] tuya error: %v", err)
		return
	}
	trim := strings.TrimSpace(string(body))
	if status != http.StatusOK || !(strings.HasPrefix(trim, "{") || strings.HasPrefix(trim, "[")) {
		log.Printf("[collector] ignorado: status=%d body_head=%q", status, trim[:min(40, len(trim))])
		return
	}
	if err := st.SaveLatest(context.Background(), []byte(trim), time.Now().UTC()); err != nil {
    log.Printf("[collector] save error: %v", err)
    return
}
	log.Printf("[collector] latest atualizado (%d bytes)", len(trim))
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}