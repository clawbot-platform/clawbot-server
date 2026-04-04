# clawbot-server

[![ci](https://github.com/clawbot-platform/clawbot-server/actions/workflows/ci.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/ci.yml)
[![quality](https://github.com/clawbot-platform/clawbot-server/actions/workflows/quality.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/quality.yml)
[![security](https://github.com/clawbot-platform/clawbot-server/actions/workflows/security.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/security.yml)
[![docker-compose-validate](https://github.com/clawbot-platform/clawbot-server/actions/workflows/docker-compose-validate.yml/badge.svg)](https://github.com/clawbot-platform/clawbot-server/actions/workflows/docker-compose-validate.yml)
[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=clawbot-platform_clawbot-server&metric=alert_status&token=a62881a65b052737ef2b8b6c8a7ccf13f3e3764f)](https://sonarcloud.io/summary/new_code?id=clawbot-platform_clawbot-server)
![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go)
![Docker Compose](https://img.shields.io/badge/Docker_Compose-Core_Stack-2496ED?logo=docker)
![Postgres](https://img.shields.io/badge/Postgres-pgvector-4169E1?logo=postgresql)
![Redis](https://img.shields.io/badge/Redis-Cache-DC382D?logo=redis)
![NATS](https://img.shields.io/badge/NATS-Messaging-27AAE1)
![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)

`clawbot-server` is a reusable Go-first control-plane foundation for Clawbot-based systems.

It is designed to provide:
- a small HTTP control plane
- embedded database migrations
- operational APIs and endpoints
- a minimal operations console
- a reusable platform base that is not tied to any one downstream project

It belongs to the broader `clawbot-platform` organization, but it is not coupled to a single vertical such as DRQ or Trust Lab.

## GHCR images

`clawbot-server` images are published to GHCR from GitHub Actions, not from developer laptops.

- image: `ghcr.io/clawbot-platform/clawbot-server`
- immutable tag pattern: `sha-<12-char-sha>`
- operational tag examples:
  - `drq-v1-baseline-20260329`
  - `drq-v1-tuned-20260401`

Runtime hosts should pull published images instead of building locally. They do not need Go, npm, or other development tooling just to deploy or run the control plane.

Publish from GitHub Actions with the `publish-image` workflow and a `release_tag` input such as:

- `drq-v1-baseline-20260329`
- `drq-v1-tuned-20260401`

If a stale package already exists in GHCR from an older manual or CLI push and is not linked to this repository, fix that in GitHub Packages before relying on the new workflow:

- connect the package to the repository, or
- delete the stale package and republish from Actions, or
- publish once to a temporary new image name if cleanup must be staged

Avoid PAT-based publishing workarounds. The repo workflow uses the repository `GITHUB_TOKEN`.

## What this repository provides

`clawbot-server` provides:

- a Go-first HTTP control-plane service
- embedded database migrations
- runtime APIs and operational endpoints
- a minimal operations console for platform operators
- local development and validation workflows
- a reusable foundation stack for projects beyond Clawbot Trust Lab

## ACH Trust Lab Control-Plane Upgrade (April 2026)

This repository now includes a production-style control-plane contract used by the redesigned `ach-trust-lab` program:

- first-class RunSpec fields for execution mode, run type, model/guardrail profiles, prompt/rule packs, and memory scope metadata
- run orchestration modes: `deterministic`, `llm`, and `dual`
- run types: `replay_run`, `agent_run`, and `week_run`
- week-run cycle entities with lifecycle status transitions and carry-forward references
- artifact registry endpoints for replay, agent, summary, guardrail, and bundle references
- dual-mode comparison objects with review/adjudication metadata
- model profile registration and retrieval (including seeded `ach-default`)
- clawmem integration adapters for scoped namespace note persistence
- configurable inference adapter wiring for ai-precision-style remote model hosts
  - `provider=local_ollama` talks directly to Ollama (`/api/chat`, `stream=false`)
  - `provider=gateway` (or other non-Ollama providers) uses `/api/v1/inference/execute`
  - per-phase timeout controls are supported for primary / guardrail / helper calls
  - compact dual payload mode and local Ollama guardrail-disable are controlled by env flags

Deterministic replay remains the authoritative measurement path. LLM and dual outputs are persisted as reviewable, versioned artifacts alongside deterministic evidence.

## Local foundation stack

This repository uses a **core vs optional** Docker Compose split.

### Core stack

The default/core stack is:

- Postgres
- Redis
- NATS

This is the recommended foundation for:
- local development
- smoke testing
- control-plane validation
- DRQ Version 1 dry runs

### Optional stack

Optional services are available through a separate Compose file:

- MinIO
- OmniRoute
- ZeroClaw
- Prometheus
- Grafana

These services are useful for:
- artifact storage
- extended runtime integrations
- observability
- dashboarding

They are **not required** for the lean `clawbot-server` deployment path.

### Current local execution model

The default local workflow is:

- **infrastructure in Docker**
- **`clawbot-server` on the host**

That means:
- Docker Compose starts the core foundation services
- `make migrate-up` and `make run-server` run the control-plane binary locally against that stack

This is the current supported repo-native workflow.

## Quick start

### 1. Prepare the environment file

```bash
cp .env.example .env
```

Validate the **core** environment:

```bash
./scripts/check-env.sh .env
```

Validate **core + optional** only if you plan to use the optional compose file:

```bash
VALIDATE_OPTIONAL_STACK=1 ./scripts/check-env.sh .env
```

### 2. Start the core stack

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  up -d
```

### 3. Start the optional stack only when needed

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  -f deploy/compose/docker-compose.optional.yml \
  up -d
```

### 4. Run the control-plane service locally

```bash
make migrate-up
make run-server
```

### 5. Validate the stack

```bash
make smoke
make coverage
```

Open the operations console in a browser at:

```text
http://127.0.0.1:8080/ops
```

Validate it from the terminal:

```bash
curl http://127.0.0.1:8080/api/v1/ops/overview
curl http://127.0.0.1:8080/api/v1/ops/services
curl http://127.0.0.1:8080/api/v1/ops/schedulers
curl http://127.0.0.1:8080/api/v1/ops/events
```

### 6. Stop the stack

Stop core stack:

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  down
```

Stop core + optional stack:

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  -f deploy/compose/docker-compose.optional.yml \
  down
```

## Core vs Optional Services

`clawbot-server` is designed to work as a reusable control-plane foundation for multiple Clawbot-based systems.

Not every service in the broader foundation stack is required for every deployment.

### Recommended service matrix

| Service        | Role                        | Recommendation for generic `clawbot-server` | Why                                                                                                           |
|----------------|-----------------------------|---------------------------------------------|---------------------------------------------------------------------------------------------------------------|
| **Postgres**   | durable control-plane state | **Core**                                    | source of truth for runs, scheduler/control-plane state, and persistent platform data                         |
| **Redis**      | cache / coordination        | **Core**                                    | useful for short-lived coordination, cached state, and runtime support                                        |
| **NATS**       | event bus / async signaling | **Core**                                    | strong fit for multi-Clawbot orchestration and decoupled control-plane events                                 |
| **Prometheus** | metrics collection          | **Optional production-ops layer**           | useful for observability, but not required for a lean deployment                                              |
| **Grafana**    | dashboards / visualization  | **Optional production-ops layer**           | useful for management and operational dashboards, but not required for a lean deployment                      |
| **MinIO**      | artifact / object storage   | **Optional artifact layer**                 | useful later for reports, audit exports, and replay bundles, but not required for the current lean deployment |
| **OmniRoute**  | model gateway               | **Optional runtime integration layer**      | useful for extended model-routing workflows, but not required for the lean control-plane stack                |
| **ZeroClaw**   | runtime substrate           | **Optional runtime integration layer**      | useful for broader agent-runtime experiments, but not required for the lean control-plane stack               |

### DRQ Version 1 note

**Version 1 of DRQ uses only the core `clawbot-server` foundation.**

That means DRQ Version 1 requires:

- Postgres
- Redis
- NATS
- `clawbot-server`

It does **not** require:

- MinIO
- Prometheus
- Grafana
- OmniRoute
- ZeroClaw

### Recommended deployment guidance

For the **24-hour dry run** and **1-week DRQ Version 1 run**, keep the stack lean and enable only the core services unless you explicitly want additional observability, artifact storage, or runtime integrations.

This helps:
- reduce container count
- simplify deployment
- reduce moving parts during benchmark validation
- keep Version 1 aligned with its current supported scope

## How to use this in other projects

`clawbot-server` can be used as the reusable control-plane foundation for projects beyond DRQ.

A downstream project typically needs to:
1. run the core stack
2. run `clawbot-server`
3. point its own services or workers at the control plane
4. enable optional services only when the project actually needs them

Use:
- the core stack for durable control-plane state and coordination
- the optional stack only for projects that benefit from artifact storage, runtime integrations, or observability

## Deployment Guide

For a single-document deployment and operations walkthrough, see:

- [Deploying Clawbot Server](docs/deploying-clawbot-server.md)

This guide covers:
- `.env` setup for the core vs optional split
- environment validation with `check-env.sh`
- foundation stack startup using the repo Compose files
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
- [Deployment guide](./docs/deploying-clawbot-server.md)

## License

This repository is licensed under the Apache License 2.0. See [LICENSE](./LICENSE).

The repo uses a top-level license file instead of per-file source headers to stay consistent with the existing code style and avoid noisy boilerplate.
