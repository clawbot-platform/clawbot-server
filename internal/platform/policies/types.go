package policies

import (
	"context"
	"encoding/json"
	"time"

	"clawbot-server/internal/platform/store"
)

type Policy struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Version     string          `json:"version"`
	Enabled     bool            `json:"enabled"`
	Description string          `json:"description"`
	RulesJSON   json.RawMessage `json:"rules_json"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateInput struct {
	Name        string          `json:"name"`
	Category    string          `json:"category"`
	Version     string          `json:"version"`
	Enabled     bool            `json:"enabled"`
	Description string          `json:"description"`
	RulesJSON   json.RawMessage `json:"rules_json"`
}

type UpdateInput struct {
	Name        *string          `json:"name,omitempty"`
	Category    *string          `json:"category,omitempty"`
	Version     *string          `json:"version,omitempty"`
	Enabled     *bool            `json:"enabled,omitempty"`
	Description *string          `json:"description,omitempty"`
	RulesJSON   *json.RawMessage `json:"rules_json,omitempty"`
}

type Repository interface {
	List(context.Context, store.DBTX) ([]Policy, error)
	Get(context.Context, store.DBTX, string) (Policy, error)
	Create(context.Context, store.DBTX, CreateInput) (Policy, error)
	Update(context.Context, store.DBTX, Policy) (Policy, error)
}

type Service interface {
	List(context.Context) ([]Policy, error)
	Get(context.Context, string) (Policy, error)
	Create(context.Context, CreateInput, string) (Policy, error)
	Update(context.Context, string, UpdateInput, string) (Policy, error)
}
