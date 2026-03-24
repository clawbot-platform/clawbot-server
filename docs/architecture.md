# Architecture

## Repository role

`clawbot-server` is the shared platform foundation repository for the `clawbot-platform` organization.

Current organization shape:

- `clawbot-server`: shared infrastructure bootstrap and control-plane-adjacent scaffolding
- `clawbot-trust-lab`: trust-lab and adversarial simulation vertical
- `clawmem`: future reusable memory subsystem

## Phase 0 and Phase 1 topology

The current local topology is intentionally simple:

- ZeroClaw acts as the runtime substrate
- OmniRoute is the single model gateway
- stateful platform primitives are provided by Postgres, Redis, NATS, and MinIO
- Prometheus and Grafana provide the initial observability layer
- `clawbot-server` adds a small Go control-plane service on top of the shared foundation

## Boundary decisions

- domain logic stays out of this repo
- OmniRoute remains responsible for model routing
- ZeroClaw remains responsible for runtime behavior
- this repo owns infrastructure bootstrap and control-plane scaffolding, not vertical-specific execution logic

## Next handoff points

- `clawbot-trust-lab` can later consume Postgres, Redis, NATS, MinIO, OmniRoute, and ZeroClaw from this shared base
- `clawbot-trust-lab` can later integrate with the control-plane API for run, bot, and policy management
- `clawmem` can later join as a separate service with its own lifecycle and storage choices
- future platform work can add exporters, dashboards, and deployment variants without changing these ownership boundaries
