package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/limiter"
	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/tuya"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		os.Exit(1)
	}

	windowSec := 3600
	if s := os.Getenv("RATE_WINDOW_SECONDS"); s != "" {
		if v, err := strconv.Atoi(s); err == nil && v > 0 {
			windowSec = v
		}
	}
	limiter := limiter.NewRateLimiter(time.Duration(windowSec) * time.Second)

	tc, err := tuya.NewTuyaClientFromEnv()
	if err != nil {
		log.Fatalf("config inválida: %v", err)
	}

	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/dale", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, jsonErr{OK: false, Error: "method_not_allowed"})
			return
		}
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Vary", "Origin")
		}

		allowed, retry := limiter.Allow()
		if !allowed {
			retrySec := int(math.Ceil(retry.Seconds()))
			w.Header().Set("Retry-After", strconv.Itoa(retrySec))
			w.Header().Set("X-RateLimit-Window-Seconds", strconv.Itoa(int((time.Duration(windowSec)*time.Second)/time.Second)))
			resetAt := time.Now().Add(retry).UTC().Format(time.RFC3339)
			w.Header().Set("X-RateLimit-Reset-At", resetAt)

			writeJSON(w, http.StatusTooManyRequests, jsonErr{
				OK:          false,
				Error:       "rate_limited",
				Description: "A rota /dale só pode ser chamada uma vez a cada janela de tempo.",
			})
			return
		}

		ctx, cancel := context.WithTimeout(r.Context(), 12*time.Second)
		defer cancel()

		status, body, err := tc.GetDevice(ctx)
		if err != nil {
			writeJSON(w, http.StatusBadGateway, jsonErr{OK: false, Error: "tuya_upstream_error", Description: err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(status)

		trim := strings.TrimSpace(string(body))
		if strings.HasPrefix(trim, "{") || strings.HasPrefix(trim, "[") {
			_, _ = w.Write(body)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"ok":     status >= 200 && status < 300,
			"status": status,
			"body":   string(body),
		})
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
