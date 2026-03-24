package audit

import (
	"context"
	"encoding/json"
	"time"

	"clawbot-server/internal/platform/store"
)

type Event struct {
	ID         string          `json:"id"`
	EventType  string          `json:"event_type"`
	EntityType string          `json:"entity_type"`
	EntityID   string          `json:"entity_id"`
	Actor      string          `json:"actor"`
	Payload    json.RawMessage `json:"payload_json"`
	CreatedAt  time.Time       `json:"created_at"`
}

type Repository interface {
	Create(context.Context, store.DBTX, Event) error
}

type Recorder interface {
	Record(context.Context, store.DBTX, Event) error
}
