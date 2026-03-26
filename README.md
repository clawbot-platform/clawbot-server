# clawbot-server

[![ci](https://github.com/clawbot-platform/clawbot-server/actions/workflows/ci.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/ci.yml)
[![quality](https://github.com/clawbot-platform/clawbot-server/actions/workflows/quality.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/quality.yml)
[![security](https://github.com/clawbot-platform/clawbot-server/actions/workflows/security.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/security.yml)
[![docker-compose-validate](https://github.com/clawbot-platform/clawbot-server/actions/workflows/docker-compose-validate.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/docker-compose-validate.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=clawbot-platform_clawbot-server&metric=alert_status&token=a62881a65b052737ef2b8b6c8a7ccf13f3e3764f)](https://sonarcloud.io/summary/new_code?id=clawbot-platform_clawbot-server)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)
![Docker Compose](https://img.shields.io/badge/Docker_Compose-Local_Stack-2496ED?logo=docker)
![Postgres](https://img.shields.io/badge/Postgres-pgvector-4169E1?logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?logo=redis)
![NATS](https://img.shields.io/badge/NATS-Messaging-27AAE1)
![Prometheus](https://img.shields.io/badge/Prometheus-Metrics-E6522C?logo=prometheus)
![Grafana](https://img.shields.io/badge/Grafana-Dashboards-F46800?logo=grafana)
![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)

`clawbot-server` is a reusable Go-first platform foundation for teams that need a local runtime stack, a small control plane, and a clean operational base without baking product-specific domain logic into the platform layer.

It belongs to the broader `clawbot-platform` organization, but it is not tied to any single downstream project or evaluation effort.

## What this repository provides

`clawbot-server` is the reusable server and control-plane foundation for Clawbot-based systems.

It provides:
- a Go-first HTTP control-plane service
- embedded database migrations
- runtime APIs and operational endpoints
- a minimal operations console for platform operators
- local development and validation workflows
- a foundation stack that can be reused by projects beyond Clawbot Trust Lab

## Local foundation stack

For local development, smoke testing, and platform validation, this repository can bring up a broader supporting stack that may include:

- Postgres
- Redis
- NATS
- MinIO
- Prometheus
- Grafana
- OmniRoute
- ZeroClaw

These services should be understood as **supporting platform components for local stack bring-up, validation, and observability**, not as a claim that every one of them is a hard runtime dependency of the core `clawbot-server` process in every deployment.

In practice:

- **Postgres** is part of the core server story because the service uses embedded migrations and database-backed runtime behavior.
- **Redis** and **NATS** are part of the broader platform stack and may support messaging, caching, or future integrations depending on the deployment.
- **Prometheus** and **Grafana** are observability components used for monitoring and local platform visibility.
- **MinIO**, **OmniRoute**, and **ZeroClaw** are part of the extended local foundation and integration story, not mandatory requirements for every consumer of `clawbot-server`.

## Deployment interpretation

When evaluating this repository, use the following distinction:

- **Core service**: the `clawbot-server` binary and the dependencies required for its current control-plane responsibilities
- **Foundation stack**: the broader local platform services used for development, smoke tests, integration validation, and observability

This distinction is important because `clawbot-server` is intended to remain a **reusable foundation** for many projects, not something tied only to one domain-specific stack.

This makes the repo suitable for:

- agent platform foundations
- internal automation backplanes
- evaluation harnesses
- control-plane services for vertical applications
- local-first integration environments for downstream systems

## Clawbot Operations Console v1

`clawbot-server` now includes a small generic operations surface for platform operators.

It is intentionally narrow. It answers:

- Is the platform healthy?
- Which services or Clawbots are healthy, degraded, down, or in maintenance?
- Which schedulers are active?
- What failed recently?
- Can an operator safely pause, resume, or trigger one run of a scheduler?

The console is available in two forms:

- JSON APIs under `/api/v1/ops/*`
- a server-rendered operator UI under `/ops`

The console is generic. It is meant to monitor `clawbot-server`, sibling services such as `clawmem`, and any future app or worker that can be surfaced through the same status model.

## What this repository does not provide

`clawbot-server` intentionally does not own:

- vertical business logic
- vertical-specific orchestration
- fraud or risk engines
- replay or benchmark logic
- memory-engine internals
- custom replacements for ZeroClaw runtime features
- custom replacements for OmniRoute routing features

That separation is deliberate. The server stays generic so other projects can reuse it without inheriting assumptions from any single downstream consumer.

## Quick start

Bring up the shared stack:

```bash
cp .env.example .env
make up
```

Apply embedded schema changes and run the service:

```bash
make migrate-up
make run-server
```

Validate the stack:

```bash
make smoke
make coverage
```

Stop everything:

```bash
make down
```

Open the operations console in a browser at `http://127.0.0.1:8080/ops`.

Validate it from the terminal:

```bash
curl http://127.0.0.1:8080/api/v1/ops/overview
curl http://127.0.0.1:8080/api/v1/ops/services
curl http://127.0.0.1:8080/api/v1/ops/schedulers
curl http://127.0.0.1:8080/api/v1/ops/events
```

## How to use this in any project

1. Start the infrastructure stack from this repo.
2. Point your own project at the services it needs: Postgres, Redis, NATS, MinIO, OmniRoute, ZeroClaw, or the HTTP control plane.
3. Use the generic `/api/v1` endpoints for run, bot, policy, and audit lifecycle data when they fit your workflow.
4. Keep domain models, product rules, and vertical orchestration in your own repository.

Typical consumers could include:

- a domain-specific evaluation harness
- an internal automation system
- an operations control plane
- a non-Clawbot service that needs a local-first foundation and lightweight platform API
- a small operations console for shared platform health and safe scheduler controls

## Quality and validation

Local validation:

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

CI includes:

- formatting checks
- `go vet`
- `golangci-lint`
- unit tests with coverage
- `gosec`
- `govulncheck`
- secret scanning
- filesystem scanning
- SonarCloud analysis and quality gate enforcement

SonarCloud:

- organization: `clawbot-platform`
- project key: `clawbot-platform_clawbot-server`
- project page: [SonarCloud overview](https://sonarcloud.io/project/overview?id=clawbot-platform_clawbot-server)

## Repository layout

- `cmd/clawbot-server/` contains the service entrypoint.
- `cmd/stack-smoke/` contains the Go reachability checker used by scripts and CI.
- `internal/app/` wires config, database access, router setup, and graceful shutdown.
- `internal/http/` contains handlers, routes, and middleware.
- `internal/platform/` contains generic platform services for runs, bots, policies, scheduler intent, audit, and common store helpers.
- `internal/db/` contains embedded migrations.
- `deploy/compose/` contains the local foundation stack definitions.
- `deploy/docker/` contains Docker-related assets used by the stack.
- `configs/` contains versioned configuration templates and provisioning.
- `docs/` contains contributor-facing platform documentation.
- `scripts/` contains operational helpers for CI and local workflows.

## Deployment Guide

For a single-document deployment and operations walkthrough, see:

- [Deploying Clawbot Server](docs/deploying-clawbot-server.md)

This guide covers:
- `.env` setup
- foundation stack startup
- running `clawbot-server`
- health and version checks
- Operations Console access
- ops API validation
- common troubleshooting

## Documentation

- [Architecture](./docs/architecture.md)
- [Foundation stack](./docs/foundation.md)
- [Control plane](./docs/control-plane.md)
- [API](./docs/api.md)
- [Ports and services](./docs/ports-and-services.md)
- [Security](./docs/security.md)
- [Development](./docs/development.md)

## License

This repository is licensed under the Apache License 2.0. See [LICENSE](./LICENSE).

The repo uses a top-level license file instead of per-file source headers to stay consistent with the existing code style and avoid noisy boilerplate.
