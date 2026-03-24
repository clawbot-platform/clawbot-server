# Phase 1 Control Plane

## What this phase adds

Phase 1 adds a Go-first control-plane skeleton on top of the Phase 0 foundation:

- a local `clawbot-server` HTTP service
- versioned `/api/v1` endpoints
- Postgres-backed persistence for runs, bots, policies, and audit events
- a scheduler placeholder that records intent instead of executing scenarios
- dashboard-summary backend support

## What still does not belong here

This phase still does not add:

- trust-lab scenario execution
- risk or fraud engines
- Red Queen logic
- `clawmem` internals
- custom ZeroClaw runtime behavior
- custom OmniRoute routing logic

## Package boundaries

- `internal/platform/runs`, `bots`, and `policies` hold platform resource types and services.
- `internal/platform/audit` persists structured control-plane events.
- `internal/platform/scheduler` records execution intent only.
- `internal/platform/store` holds common Postgres transaction and dashboard helpers.
- `internal/http` owns API delivery.
- `internal/db` owns embedded migrations.

## Relationship to other repos

- `clawbot-trust-lab` will later consume this control plane rather than embedding platform state management locally.
- `clawmem` remains a separate future subsystem and is not implemented here.
- ZeroClaw and OmniRoute remain integrated external components, not reimplemented internals.
