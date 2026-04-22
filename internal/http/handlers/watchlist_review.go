package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"clawbot-server/internal/watchlistreviewclient"
)

type WatchlistReviewService interface {
	Review(ctx context.Context, req watchlistreviewclient.ReviewRequest, correlationID string) (watchlistreviewclient.ReviewResponse, error)
}

type WatchlistReviewHandler struct {
	service WatchlistReviewService
}

func NewWatchlistReviewHandler(service WatchlistReviewService) *WatchlistReviewHandler {
	return &WatchlistReviewHandler{service: service}
}

func (h *WatchlistReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeWatchlistJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
		return
	}
	if h.service == nil {
		writeWatchlistJSON(w, http.StatusNotImplemented, map[string]any{"error": "watchlist review integration not configured"})
		return
	}

	var req watchlistreviewclient.ReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeWatchlistJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json"})
		return
	}

	correlationID := watchlistCorrelationID(r)
	w.Header().Set("X-Correlation-ID", correlationID)

	resp, err := h.service.Review(r.Context(), req, correlationID)
	if err != nil {
		writeWatchlistJSON(w, http.StatusBadGateway, map[string]any{"error": err.Error()})
		return
	}

	writeWatchlistJSON(w, http.StatusOK, resp)
}

func watchlistCorrelationID(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Correlation-ID")); v != "" {
		return v
	}
	return "corr_" + time.Now().UTC().Format("20060102150405.000000000")
}

func writeWatchlistJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
