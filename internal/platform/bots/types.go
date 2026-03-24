package bots

import (
	"context"
	"encoding/json"
	"time"

	"clawbot-server/internal/platform/store"
)

type Bot struct {
	ID         string          `json:"id"`
	Name       string          `json:"name"`
	Role       string          `json:"role"`
	Runtime    string          `json:"runtime"`
	Status     string          `json:"status"`
	RepoHint   string          `json:"repo_hint"`
	Version    string          `json:"version"`
	ConfigJSON json.RawMessage `json:"config_json"`
	CreatedAt  time.Time       `json:"created_at"`
	UpdatedAt  time.Time       `json:"updated_at"`
}

type CreateInput struct {
	Name       string          `json:"name"`
	Role       string          `json:"role"`
	Runtime    string          `json:"runtime"`
	Status     string          `json:"status"`
	RepoHint   string          `json:"repo_hint"`
	Version    string          `json:"version"`
	ConfigJSON json.RawMessage `json:"config_json"`
}

type UpdateInput struct {
	Name       *string          `json:"name,omitempty"`
	Role       *string          `json:"role,omitempty"`
	Runtime    *string          `json:"runtime,omitempty"`
	Status     *string          `json:"status,omitempty"`
	RepoHint   *string          `json:"repo_hint,omitempty"`
	Version    *string          `json:"version,omitempty"`
	ConfigJSON *json.RawMessage `json:"config_json,omitempty"`
}

type Repository interface {
	List(context.Context, store.DBTX) ([]Bot, error)
	Get(context.Context, store.DBTX, string) (Bot, error)
	Create(context.Context, store.DBTX, CreateInput) (Bot, error)
	Update(context.Context, store.DBTX, Bot) (Bot, error)
}

type Service interface {
	List(context.Context) ([]Bot, error)
	Get(context.Context, string) (Bot, error)
	Create(context.Context, CreateInput, string) (Bot, error)
	Update(context.Context, string, UpdateInput, string) (Bot, error)
}
