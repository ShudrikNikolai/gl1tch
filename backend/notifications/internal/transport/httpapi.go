// Package transport отдаёт SSE и health-эндпоинты notifications.
package transport

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/boards"
	"github.com/ShudrikNikolai/gl1tch/backend/notifications/internal/hub"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/jwtx"
	"github.com/ShudrikNikolai/gl1tch/backend/shared/pkg/metrics"
)

const (
	heartbeatEvery = 15 * time.Second
	checkTimeout   = 2 * time.Second
)

type Handler struct {
	verifier *jwtx.Verifier
	boards   *boards.Client
	hub      *hub.Hub
	log      *slog.Logger
}

func New(verifier *jwtx.Verifier, boardsClient *boards.Client, h *hub.Hub, log *slog.Logger) http.Handler {
	api := &Handler{verifier: verifier, boards: boardsClient, hub: h, log: log}
	mux := http.NewServeMux()
	mux.Handle("GET /metrics", metrics.Handler("notifications"))
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("GET /readyz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("GET /v1/threads/{thread_id}/stream", api.stream)

	return mux
}

func (h *Handler) stream(w http.ResponseWriter, r *http.Request) {
	userID, err := h.userID(r)
	if err != nil {
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "unauthorized"})
		return
	}

	threadID := r.PathValue("thread_id")
	if threadID == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "thread_id is required"})
		return
	}

	checkCtx, cancel := context.WithTimeout(r.Context(), checkTimeout)
	posted, err := h.boards.HasUserPosted(checkCtx, threadID, userID)
	cancel()
	if err != nil {
		h.log.WarnContext(r.Context(), "check thread participation failed",
			slog.String("thread_id", threadID), slog.String("user_id", userID), slog.Any("error", err))
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "boards unavailable"})
		return
	}
	if !posted {
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "user has not posted in this thread"})
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": "streaming is not supported"})
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "retry: 3000\n\n")
	flusher.Flush()

	client := h.hub.Subscribe(threadID)
	defer client.Close()

	heartbeat := time.NewTicker(heartbeatEvery)
	defer heartbeat.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-heartbeat.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			flusher.Flush()
		case evt, ok := <-client.Events():
			if !ok {
				return
			}
			if _, err := fmt.Fprintf(w, "id: %s\nevent: post.created\ndata: %s\n\n", evt.ID, evt.Data); err != nil {
				return
			}
			flusher.Flush()
		}
	}
}

func (h *Handler) userID(r *http.Request) (string, error) {
	const prefix = "Bearer "
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, prefix) {
		return "", fmt.Errorf("missing bearer token")
	}

	claims, err := h.verifier.Verify(r.Context(), strings.TrimSpace(strings.TrimPrefix(auth, prefix)))
	if err != nil {
		return "", err
	}

	return claims.Subject, nil
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}
