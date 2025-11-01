package collector

import(
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/store"
	"github.com/Guiribei/monitoramento_de_refrigeradores/backend/tuya"
)

func RunCollect(ctx context.Context, tc *tuya.TuyaClient, st *store.Store) {
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