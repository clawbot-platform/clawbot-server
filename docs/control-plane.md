# Control Plane

## What the service provides

The `clawbot-server` HTTP service adds a small reusable control plane on top of the foundation stack.

It provides:

- versioned `/api/v1` endpoints
- Postgres-backed persistence for runs, bots, policies, and audit events
- dashboard-summary backend support
- scheduler intent recording for downstream integrations

## What does not belong here

The control plane does not add:

- downstream scenario execution
- business- or vertical-specific workflows
- fraud or risk engines
- replay orchestration
- memory-service internals
- custom ZeroClaw runtime behavior
- custom OmniRoute routing logic

## Package boundaries

- `internal/platform/runs`, `bots`, and `policies` hold generic platform resource types and services.
- `internal/platform/audit` persists structured control-plane events.
- `internal/platform/scheduler` records scheduling intent only.
- `internal/platform/store` holds common Postgres helpers and dashboard queries.
- `internal/http` owns API delivery.
- `internal/db` owns embedded migrations.

## Relationship to downstream projects

Downstream applications can consume this control plane instead of embedding their own platform-state management. Consumer repositories are examples of that model, not hard dependencies of the runtime or API surface here.
