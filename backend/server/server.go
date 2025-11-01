package server

import (
	"errors"
	"net/http"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/store"
	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/tuya"
	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/utils"
)

func Run(tc *tuya.TuyaClient, st *store.Store, port string) {
	allowedOrigin := os.Getenv("ALLOWED_ORIGIN")

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	mux.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			utils.WriteJSON(w, http.StatusMethodNotAllowed, utils.JsonErr{OK: false, Error: "method_not_allowed"})
			return
		}
		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Vary", "Origin")
		}

		snap, err := st.GetLatest(r.Context())
		if err != nil {
			utils.WriteJSON(w, http.StatusServiceUnavailable, utils.JsonErr{
				OK:          false,
				Error:       "no_data",
				Description: "Ainda não há dados coletados. Tente novamente em instantes.",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Header().Set("X-Data-Age-ms", strconv.FormatInt(time.Now().UnixMilli()-snap.FetchedAtMs, 10))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(snap.RawJSON)
	})

	srv := &http.Server{
		Addr:              ":" + port,
		Handler:           utils.LogMiddleware(utils.SecurityHeaders(mux)),
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