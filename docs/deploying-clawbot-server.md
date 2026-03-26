# Deploying Clawbot Server
## Single-Document Core Stack Deployment and Operations Guide

This document is the primary deployment guide for `clawbot-server`.

It is intended for teams that want to:
- run the core foundation services `clawbot-server` depends on
- start the control-plane service cleanly
- verify that the control-plane service is healthy
- access the Operations Console UI
- avoid jumping across multiple repository documents

This guide reflects the current supported repository workflow:

- Docker Compose for the core foundation services
- optional Docker Compose for additional services
- `clawbot-server` running locally on the host
- local or internal-only deployment

---

## What `clawbot-server` is

`clawbot-server` is a reusable server and control-plane foundation for Clawbot-based systems.

It provides:
- a Go-first HTTP control-plane service
- embedded database migrations
- operational endpoints
- a small embedded Operations Console UI
- a foundation for managing multiple ClawBots and related services

It is intended to be reusable beyond any single domain application.

---

## What this guide deploys

This guide deploys the **core `clawbot-server` stack**:

- PostgreSQL + pgvector
- Redis
- NATS
- `clawbot-server`

Optional services such as:
- MinIO
- OmniRoute
- ZeroClaw
- Prometheus
- Grafana

are **not required** for the current lean deployment path and are intentionally excluded from the default Version 1 DRQ deployment.

---

## Deployment model

Current recommended deployment model:
- configure `.env`
- validate the core environment
- run the core foundation services with Docker Compose
- optionally enable optional services only when needed
- run `clawbot-server` locally on the host
- verify health, version, and Operations Console endpoints

This is the recommended practical path for internal, homelab, and incumbent lab deployments.

---

## Core vs Optional Services

`clawbot-server` is designed to work as a reusable control-plane foundation for multiple Clawbot-based systems.

Not every service in the broader foundation stack is required for every deployment.

### Recommended service matrix

| Service        | Role                        | Recommendation for generic `clawbot-server` | Required for DRQ Version 1? | Why                                                                                                     |
|----------------|-----------------------------|---------------------------------------------|-----------------------------|---------------------------------------------------------------------------------------------------------|
| **Postgres**   | durable control-plane state | **Core**                                    | **Yes**                     | source of truth for runs, scheduler/control-plane state, and persistent platform data                   |
| **Redis**      | cache / coordination        | **Core**                                    | **Yes**                     | useful for short-lived coordination, cached state, and runtime support                                  |
| **NATS**       | event bus / async signaling | **Core**                                    | **Yes**                     | strong fit for multi-Clawbot orchestration and decoupled control-plane events                           |
| **Prometheus** | metrics collection          | **Optional production-ops layer**           | **No**                      | useful for observability, but not required to run DRQ Version 1                                         |
| **Grafana**    | dashboards / visualization  | **Optional production-ops layer**           | **No**                      | useful for management and operational dashboards, but not required for DRQ Version 1                    |
| **MinIO**      | artifact / object storage   | **Optional artifact layer**                 | **No**                      | useful later for reports, audit exports, and replay bundles, but not required for current DRQ Version 1 |
| **OmniRoute**  | model gateway               | **Optional runtime integration layer**      | **No**                      | useful for extended model-routing workflows, but not required for the lean control-plane stack          |
| **ZeroClaw**   | runtime substrate           | **Optional runtime integration layer**      | **No**                      | useful for broader agent-runtime experiments, but not required for DRQ Version 1                        |

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

---

## Prerequisites

You need:

- Docker
- Docker Compose
- Go toolchain
- `curl`

Optional but useful:
- `jq`

This guide uses Docker for the foundation services and the local Go toolchain for the `clawbot-server` process.

---

## 1. Prepare the repository environment file

From the repository root:

```bash
cp .env.example .env
```

### Core vs optional environment validation

Validate the **core stack**:

```bash
./scripts/check-env.sh .env
```

Validate **core + optional stack** only if you plan to use the optional compose file:

```bash
VALIDATE_OPTIONAL_STACK=1 ./scripts/check-env.sh .env
```

### What the environment file means

The `.env.example` file is aligned to the Compose split:

- **Core/default**
  - Postgres
  - Redis
  - NATS

- **Optional**
  - MinIO
  - OmniRoute
  - ZeroClaw
  - Prometheus
  - Grafana

Optional secrets and values are only required when enabling the optional stack.

---

## 2. Start the core stack

Use the repository Compose files directly.

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  up -d
```

Check container status:

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  ps
```

Expected outcome:
- core containers are running
- Postgres, Redis, and NATS become healthy

---

## 3. Start optional services only when needed

Optional services are provided by a separate Compose file.

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  -f deploy/compose/docker-compose.optional.yml \
  up -d
```

This enables:

- MinIO
- OmniRoute
- ZeroClaw
- Prometheus
- Grafana

These are not required for the lean `clawbot-server` deployment path.

---

## 4. Run `clawbot-server`

Run the control-plane service locally against the core stack:

```bash
make migrate-up
make run-server
```

Expected outcome:
- embedded migrations apply cleanly
- the service starts successfully
- the service listens on `127.0.0.1:8080` unless overridden in `.env`

---

## 5. Verify health and version

Open another terminal and run:

```bash
curl http://127.0.0.1:8080/healthz
curl http://127.0.0.1:8080/readyz
curl http://127.0.0.1:8080/version
```

Expected outcome:
- `/healthz` returns success
- `/readyz` returns success
- `/version` returns build metadata if ldflags are configured

---

## 6. Access the Operations Console UI

Open the embedded Operations Console in a browser:

```text
http://127.0.0.1:8080/ops
```

Useful pages:

```text
http://127.0.0.1:8080/ops
http://127.0.0.1:8080/ops/services
http://127.0.0.1:8080/ops/schedulers
http://127.0.0.1:8080/ops/events
```

---

## 7. Validate the operations API

Run these from another terminal:

```bash
curl http://127.0.0.1:8080/api/v1/ops/overview
curl http://127.0.0.1:8080/api/v1/ops/services
curl http://127.0.0.1:8080/api/v1/ops/schedulers
curl http://127.0.0.1:8080/api/v1/ops/events
```

Expected outcome:
- overview data is returned
- service list is returned
- scheduler list is returned
- recent events are returned

---

## 8. Test safe maintenance actions

Use the Operations Console UI or the API.

### Put a service into maintenance mode

```bash
curl -X POST http://127.0.0.1:8080/api/v1/ops/services/<service-id>/maintenance
```

### Resume a service

```bash
curl -X POST http://127.0.0.1:8080/api/v1/ops/services/<service-id>/resume
```

### Pause a scheduler

```bash
curl -X POST http://127.0.0.1:8080/api/v1/ops/schedulers/<scheduler-id>/pause
```

### Resume a scheduler

```bash
curl -X POST http://127.0.0.1:8080/api/v1/ops/schedulers/<scheduler-id>/resume
```

### Run a scheduler once

```bash
curl -X POST http://127.0.0.1:8080/api/v1/ops/schedulers/<scheduler-id>/run-once
```

Then re-check state:

```bash
curl http://127.0.0.1:8080/api/v1/ops/services/<service-id>
curl http://127.0.0.1:8080/api/v1/ops/schedulers/<scheduler-id>
curl http://127.0.0.1:8080/api/v1/ops/events
```

---

## 9. Local quality checks

Run the local quality suite:

```bash
go test ./...
go vet ./...
golangci-lint run ./...
gosec ./...
govulncheck ./...
```

Generate coverage:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Optional HTML coverage:

```bash
go tool cover -html=coverage.out -o coverage.html
```

---

## 10. Common startup issue: PostgreSQL role does not exist

If you see an error like:

```text
FATAL: role "clawbot" does not exist
```

the most likely cause is:
- the PostgreSQL volume was initialized earlier with different credentials
- changing `.env` later did not recreate the `clawbot` role automatically

### Fastest local fix

If you do not need to preserve the existing database volume:

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  down -v

docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  up -d
```

### Alternative fix

If you need to preserve the volume, manually create the role/database in PostgreSQL.

---

## 11. Common startup issue: `/version` shows `dev` / `unknown`

If:

```bash
curl http://127.0.0.1:8080/version
```

returns fallback values, then build metadata is not being injected during startup.

The current repository `Makefile` should handle this. If it does not, verify the local `Makefile` includes `-ldflags` injection for:
- version
- commit
- build date

---

## 12. Common issue: Operations Console template/render failures

If the UI returns a 500 error or template panic:
- confirm all templates parse successfully
- confirm `base.gohtml` renders only known named templates
- confirm the server no longer injects raw HTML with `template.HTML(...)`

This is especially relevant after recent UI/security hardening.

---

## 13. Stop the stack

Stop the core stack:

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  down
```

Stop the core stack and remove volumes:

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  down -v
```

Stop core + optional stack:

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  -f deploy/compose/docker-compose.optional.yml \
  down
```

To stop only the foreground server process:

```bash
Ctrl+C
```

---

## 14. Recommended deployment flow

Use this sequence:

1. `cp .env.example .env`
2. `./scripts/check-env.sh .env`
3. start the core stack with the repo Compose files
4. optionally enable the optional stack only when needed
5. `make migrate-up`
6. `make run-server`
7. verify `/healthz`, `/readyz`, `/version`
8. open `/ops`
9. validate ops APIs
10. run maintenance/scheduler actions

---

## 15. What a successful deployment looks like

A deployment is in good shape if:

- core stack starts cleanly
- Postgres, Redis, and NATS are healthy
- `clawbot-server` starts without migration errors
- health/readiness endpoints pass
- version endpoint shows real metadata
- Operations Console pages load correctly
- maintenance and scheduler actions work
- ops API responses match UI state
- local quality checks pass

---

## Summary

This guide is the single deployment and operations entry point for `clawbot-server`.

It reflects the repository-native deployment model:

- `docker-compose.yml` = core only
- `docker-compose.override.yml` = local dev tweaks for core only
- `docker-compose.optional.yml` = optional services

Use the core stack by default.
Enable optional services only when your deployment actually needs them.
