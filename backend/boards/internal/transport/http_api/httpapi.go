// Package httpapi публикует HTTP-эндпоинты сервиса: только пробы здоровья. Бизнес-API по gRPC (REST для клиентов делает Envoy).
package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"sort"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/metrics"
)

// Pinger проверяет зависимость (Postgres).
type Pinger interface {
	Ping(ctx context.Context) error
}

const readyTimeout = 2 * time.Second

// GET /healthz  процесс жив
// GET /readyz   зависимости доступны (503 и список недоступных)
func New(deps map[string]Pinger) http.Handler {
	mux := http.NewServeMux()

	mux.Handle("GET /metrics", metrics.Handler("boards"))

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), readyTimeout)
		defer cancel()

		var failed []string
		for name, p := range deps {
			if err := p.Ping(ctx); err != nil {
				failed = append(failed, name) // только имя: текст ошибки может содержать адреса и учётные данные
			}
		}
		if len(failed) > 0 {
			sort.Strings(failed)
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"status": "unavailable", "failed": failed})

			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	return mux
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
