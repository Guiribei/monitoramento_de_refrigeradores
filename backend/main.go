package main

import (
	"context"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/joho/godotenv"

	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/collector"
	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/store"
	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/tuya"
	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/server"
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

	windowSec := getenvInt("RATE_WINDOW_SECONDS", 3600)

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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		collector.RunCollect(ctx, tc, st)
		tick := time.NewTicker(time.Duration(windowSec) * time.Second)
		defer tick.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tick.C:
				collector.RunCollect(ctx, tc, st)
			}
		}
	}()

	server.Run(tc, st, port)
}
