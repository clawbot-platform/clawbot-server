gii# clawbot-server

[![ci](https://github.com/clawbot-platform/clawbot-server/actions/workflows/ci.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/ci.yml)
[![quality](https://github.com/clawbot-platform/clawbot-server/actions/workflows/quality.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/quality.yml)
[![security](https://github.com/clawbot-platform/clawbot-server/actions/workflows/security.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/security.yml)
[![docker-compose-validate](https://github.com/clawbot-platform/clawbot-server/actions/workflows/docker-compose-validate.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/docker-compose-validate.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=clawbot-platform_clawbot-server&metric=alert_status&token=a62881a65b052737ef2b8b6c8a7ccf13f3e3764f)](https://sonarcloud.io/summary/new_code?id=clawbot-platform_clawbot-server)

![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)
![Docker Compose](https://img.shields.io/badge/Docker_Compose-Local_Stack-2496ED?logo=docker)
![Postgres](https://img.shields.io/badge/Postgres-pgvector-4169E1?logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?logo=redis)
![NATS](https://img.shields.io/badge/NATS-Eventing-27AAE1)
![Prometheus](https://img.shields.io/badge/Prometheus-Metrics-E6522C?logo=prometheus)
![Grafana](https://img.shields.io/badge/Grafana-Dashboards-F46800?logo=grafana)

`clawbot-server` is a reusable Go-first platform foundation for projects that need a local infrastructure stack, a small control plane, and a clean runtime substrate without baking product-specific domain logic into the platform layer.

It started as part of the broader `clawbot-platform` organization, but it is not tied to `clawbot-trust-lab`. Trust Lab is one consumer example, not a required dependency or the only intended use case.

## What this repository is for

Use `clawbot-server` when you want a boring, inspectable platform base that provides:

- a repeatable Docker Compose lab stack
- a small HTTP control plane for generic runs, bots, policies, and audit events
- observability and storage primitives that downstream projects can reuse
- clear CI, security, and coverage automation
- integration points for ZeroClaw and OmniRoute without reimplementing either system

This makes the repo suitable for:

- internal agent platforms
- evaluation harnesses
- runtime-adjacent developer environments
- domain-specific control planes that need shared infrastructure but keep business logic in their own repos

## What this repository does not do

`clawbot-server` intentionally does not own:

- business or simulation logic from downstream verticals
- fraud, risk, replay, or Red Queen domain behavior
- memory-engine internals
- bespoke replacements for ZeroClaw runtime features
- bespoke replacements for OmniRoute routing features

Those remain downstream concerns. This repo stays generic on purpose.

## Quick start

Bring up the shared stack:

```bash
cp .env.example .env
make up
```

Apply the embedded control-plane schema and run the service:

```bash
make migrate-up
make run-server
```

Validate the local stack:

```bash
make smoke
make coverage
```

Stop everything:

```bash
make down
```

## How to use this in any project

1. Start the foundation stack and control plane from this repo.
2. Point your own project at the shared platform services it needs: Postgres, Redis, NATS, MinIO, OmniRoute, or ZeroClaw.
3. Use the `/api/v1` control-plane endpoints for generic run, bot, policy, and audit lifecycle data if they fit your workflow.
4. Keep domain models, evaluation logic, and vertical-specific orchestration in your own repository.

That separation is deliberate. `clawbot-server` should be reusable whether the downstream project is Trust Lab, another control-validation system, or a different agentic application entirely.

## Local validation

```bash
go test ./...
go vet ./...
golangci-lint run ./...
make coverage
```

Optional local security tooling when installed:

```bash
make security
```

SonarCloud is configured for CI with:

- organization: `clawbot-platform`
- project key: `clawbot-platform_clawbot-server`
- project page: [SonarCloud overview](https://sonarcloud.io/project/overview?id=clawbot-platform_clawbot-server)

## Repo layout

- `cmd/clawbot-server/` contains the control-plane service entrypoint.
- `cmd/stack-smoke/` contains the Go reachability checker used by scripts and CI.
- `internal/app/` wires config, DB, router, and graceful shutdown.
- `internal/http/` contains versioned handlers, routes, and middleware.
- `internal/platform/` contains generic platform services for runs, bots, policies, scheduler, audit, and common store helpers.
- `internal/db/` contains embedded migrations for the control-plane schema.
- `deploy/compose/` contains the local foundation stack definitions.
- `deploy/docker/` contains Docker-related assets needed by the stack.
- `configs/` contains Prometheus, Grafana, ZeroClaw, and environment templates.
- `docs/` contains contributor-facing platform documentation.
- `scripts/` contains thin operational helpers for CI and local workflows.

## Additional docs

- [Foundation](./docs/foundation.md)
- [Architecture](./docs/architecture.md)
- [API](./docs/api.md)
- [Phase 1 control plane](./docs/phase-1-control-plane.md)
- [Ports and services](./docs/ports-and-services.md)
- [Security](./docs/security.md)
- [Development](./docs/development.md)
