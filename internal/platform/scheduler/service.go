package scheduler

import (
	"context"
	"encoding/json"
	"fmt"

	"clawbot-server/internal/platform/audit"
	"clawbot-server/internal/platform/store"
)

type Signal struct {
	RunID   string `json:"run_id"`
	RunName string `json:"run_name"`
	Status  string `json:"status"`
	Actor   string `json:"actor"`
	Reason  string `json:"reason"`
}

type Service interface {
	RecordRunIntent(context.Context, store.DBTX, Signal) error
}

type PlaceholderService struct {
	audits audit.Recorder
}

func NewPlaceholderService(audits audit.Recorder) *PlaceholderService {
	return &PlaceholderService{audits: audits}
}

func (s *PlaceholderService) RecordRunIntent(ctx context.Context, q store.DBTX, signal Signal) error {
	payload, err := json.Marshal(map[string]any{
		"run_id":   signal.RunID,
		"run_name": signal.RunName,
		"status":   signal.Status,
		"reason":   signal.Reason,
		"note":     "This scheduler records run intent only. Execution belongs to downstream integrations.",
	})
	if err != nil {
		return fmt.Errorf("marshal scheduler intent payload: %w", err)
	}

	return s.audits.Record(ctx, q, audit.Event{
		EventType:  "scheduler.intent.recorded",
		EntityType: "run",
		EntityID:   signal.RunID,
		Actor:      signal.Actor,
		Payload:    payload,
	})
}
