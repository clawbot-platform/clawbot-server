package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"clawbot-server/internal/platform/store"
)

type envelope map[string]any

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code string, message string) {
	writeJSON(w, status, envelope{
		"error": envelope{
			"code":    code,
			"message": message,
		},
	})
}

func decodeJSON(r *http.Request, out any) error {
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	return decoder.Decode(out)
}

func actorFromRequest(r *http.Request) string {
	actor := strings.TrimSpace(r.Header.Get("X-Actor"))
	if actor == "" {
		return "control-plane"
	}
	return actor
}

func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrNotFound):
		writeError(w, http.StatusNotFound, "not_found", err.Error())
	default:
		writeError(w, http.StatusBadRequest, "bad_request", err.Error())
	}
}
