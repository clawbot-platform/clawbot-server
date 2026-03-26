# Architecture

## Repository role

`clawbot-server` is the shared platform foundation repository for the `clawbot-platform` organization.

Current organization shape:

- `clawbot-server`: shared infrastructure bootstrap and control-plane-adjacent scaffolding
- `clawbot-trust-lab`: one downstream consumer example focused on adversarial trust testing
- `clawmem`: one downstream memory-service example

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
- downstream repos named in this document are examples of consumers, not hard dependencies of the platform layer

## Next handoff points

- any downstream project can consume Postgres, Redis, NATS, MinIO, OmniRoute, and ZeroClaw from this shared base
- any downstream project can integrate with the control-plane API for run, bot, policy, and audit management
- Trust Lab and clawmem remain useful examples of how to consume the platform without coupling the platform to their internals
- future platform work can add exporters, dashboards, and deployment variants without changing these ownership boundaries
