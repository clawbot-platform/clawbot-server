package runs

import (
	"context"
	"encoding/json"
	"time"

	"clawbot-server/internal/platform/store"
)

type Run struct {
	ID           string          `json:"id"`
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Status       string          `json:"status"`
	ScenarioType string          `json:"scenario_type"`
	CreatedBy    string          `json:"created_by"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
	StartedAt    *time.Time      `json:"started_at,omitempty"`
	CompletedAt  *time.Time      `json:"completed_at,omitempty"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
}

type CreateInput struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	Status       string          `json:"status"`
	ScenarioType string          `json:"scenario_type"`
	CreatedBy    string          `json:"created_by"`
	MetadataJSON json.RawMessage `json:"metadata_json"`
}

type UpdateInput struct {
	Name         *string          `json:"name,omitempty"`
	Description  *string          `json:"description,omitempty"`
	Status       *string          `json:"status,omitempty"`
	ScenarioType *string          `json:"scenario_type,omitempty"`
	MetadataJSON *json.RawMessage `json:"metadata_json,omitempty"`
	StartedAt    *time.Time       `json:"started_at,omitempty"`
	CompletedAt  *time.Time       `json:"completed_at,omitempty"`
}

type Repository interface {
	List(context.Context, store.DBTX) ([]Run, error)
	Get(context.Context, store.DBTX, string) (Run, error)
	Create(context.Context, store.DBTX, CreateInput) (Run, error)
	Update(context.Context, store.DBTX, Run) (Run, error)
}

type Service interface {
	List(context.Context) ([]Run, error)
	Get(context.Context, string) (Run, error)
	Create(context.Context, CreateInput, string) (Run, error)
	Update(context.Context, string, UpdateInput, string) (Run, error)
}
