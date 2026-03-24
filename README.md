# clawbot-server

`clawbot-server` owns the shared platform foundation and Phase 1 control-plane skeleton for the `clawbot-platform` organization.

## What this repo owns

- Docker Compose based shared lab bootstrap
- Phase 1 control-plane HTTP service
- run, bot, policy, audit, and scheduler scaffolding
- ZeroClaw integration as the runtime substrate
- OmniRoute integration as the model gateway
- Postgres with `pgvector`
- Redis
- NATS
- MinIO
- Prometheus and Grafana
- environment templates, smoke tests, and CI/security workflows

## What this repo does not own

- `clawbot-trust-lab` business logic
- scenario engines, risk engines, or Red Queen logic
- `clawmem` internals
- bespoke replacements for ZeroClaw or OmniRoute

## Quick start

```bash
cp .env.example .env
make up
make migrate-up
make run-server
```

Optional readiness check:

```bash
make smoke
```

Stop the foundation stack:

```bash
make down
```

## Repo layout

- `cmd/clawbot-server/` contains the Phase 1 control-plane service entrypoint.
- `cmd/stack-smoke/` contains the Go reachability checker used by scripts and CI.
- `internal/app/` wires config, DB, router, and graceful shutdown.
- `internal/http/` contains versioned handlers, routes, and middleware.
- `internal/platform/` contains platform-only packages for runs, bots, policies, scheduler, audit, and common store helpers.
- `internal/db/` contains embedded migrations for the control-plane schema.
- `deploy/compose/` contains the local lab stack definitions.
- `deploy/docker/` contains Docker-related assets needed by the stack.
- `configs/` contains Prometheus, Grafana, ZeroClaw, and environment templates.
- `docs/` contains contributor-facing platform documentation.
- `scripts/` contains thin operational helpers.

## Additional docs

- [Foundation](./docs/foundation.md)
- [Architecture](./docs/architecture.md)
- [API](./docs/api.md)
- [Phase 1 control plane](./docs/phase-1-control-plane.md)
- [Ports and services](./docs/ports-and-services.md)
- [Security](./docs/security.md)
- [Development](./docs/development.md)
