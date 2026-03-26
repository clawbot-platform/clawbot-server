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

- downstream repos can consume this control plane rather than embedding platform state management locally.
- `clawbot-trust-lab` and `clawmem` are useful examples of those consumers, but they are not required by the runtime or API surface here.
- ZeroClaw and OmniRoute remain integrated external components, not reimplemented internals.
