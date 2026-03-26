```md
# Deploying Clawbot Server
## Single-Document Deployment and Operations Guide

This document is the primary deployment guide for `clawbot-server`.

It is intended for teams that want to:
- deploy `clawbot-server`
- run the local foundation stack it depends on
- verify the control-plane service is healthy
- access the Operations Console UI
- avoid jumping across multiple repository documents

This guide reflects the current repository deployment model:
- source checkout
- Docker Compose foundation stack
- local or internal-only deployment
- `clawbot-server` run as the control-plane service

---

## What `clawbot-server` is

`clawbot-server` is a reusable server and control-plane foundation for Clawbot-based systems.

It provides:
- a Go-first HTTP control-plane service
- embedded database migrations
- operational endpoints
- a small embedded Operations Console UI
- a local foundation stack for development, smoke checks, and platform validation

It is intended to be reusable beyond any single domain application.

---

## What this guide deploys

This guide deploys:

- PostgreSQL + pgvector
- Redis
- NATS
- MinIO
- Prometheus
- Grafana
- OmniRoute
- ZeroClaw
- `clawbot-server`

Important distinction:

- the **foundation stack** includes multiple supporting services for local validation and observability
- the **core `clawbot-server` process** is the reusable control-plane service
- not every listed supporting service is necessarily a hard runtime dependency for every future consumer of `clawbot-server`

---

## Deployment model

Current deployment model:
- clone the repository
- configure `.env`
- start the foundation stack with Docker Compose
- run `clawbot-server` locally against that stack
- verify health, version, and Operations Console endpoints

This is the current practical path for internal, lab, and homelab deployment.

---

## Prerequisites

You need:

- Git
- Docker / Docker Compose
- Go toolchain
- `curl`

Optional but useful:
- `golangci-lint`
- `gosec`
- `govulncheck`

---

## 1. Clone the repository

```bash
git clone https://github.com/clawbot-platform/clawbot-server.git
cd clawbot-server
```

---

## 2. Create `.env`

Copy the example file:

```bash
cp .env.example .env
```

Then edit `.env`.

### Required values to replace

At minimum, replace all placeholder secret values:

- `POSTGRES_PASSWORD`
- `MINIO_ROOT_PASSWORD`
- `GRAFANA_ADMIN_PASSWORD`
- `ZEROCLAW_API_KEY`
- `ZEROCLAW_GATEWAY_TOKEN`

Also make sure `DATABASE_URL` uses the same PostgreSQL password as `POSTGRES_PASSWORD`.

Example pattern:

```env
POSTGRES_PASSWORD=replace_with_real_value
DATABASE_URL=postgres://clawbot:replace_with_real_value@127.0.0.1:5432/clawbot?sslmode=disable
```

### Values that usually do not need changing

Most default values can stay as-is unless:
- you have port conflicts
- you want non-default local bindings
- you are integrating into a different internal environment

These usually can remain unchanged:
- host bindings
- ports
- image names
- database name/user
- Redis settings
- NATS settings
- Prometheus/Grafana ports
- OmniRoute defaults
- ZeroClaw defaults
- smoke timeout
- server log level and shutdown timeout

---

## 3. Validate environment configuration

Run:

```bash
make check-env
```

Expected outcome:
- the environment file passes validation
- there are no missing required variables

---

## 4. Start the foundation stack

Bring up the foundation stack:

```bash
make up
```

Then validate it with the built-in smoke checks:

```bash
make smoke
```

Expected outcome:
- Docker containers start successfully
- smoke checks pass for the foundation services

---

## 5. Run `clawbot-server`

Start the control-plane service:

```bash
SERVER_ADDRESS=127.0.0.1:8081 make run-server
```

Notes:
- this runs the server locally against the foundation stack
- the server uses embedded database migrations
- version metadata is injected by the `Makefile`

Expected outcome:
- the server starts successfully
- no migration or database errors appear
- the service listens on `127.0.0.1:8081`

---

## 6. Verify health and version

Open another terminal and run:

```bash
curl http://127.0.0.1:8081/healthz
curl http://127.0.0.1:8081/readyz
curl http://127.0.0.1:8081/version
```

Expected outcome:
- `/healthz` returns success
- `/readyz` returns success
- `/version` returns real build metadata instead of fallback placeholders

Example:

```json
{
  "version": "v1.0.0",
  "commit": "abc1234",
  "build_date": "2026-03-26T14:35:00Z"
}
```

---

## 7. Access the Operations Console UI

Open the embedded Operations Console in a browser:

```text
http://127.0.0.1:8081/ops
```

Useful pages:

```text
http://127.0.0.1:8081/ops
http://127.0.0.1:8081/ops/services
http://127.0.0.1:8081/ops/schedulers
http://127.0.0.1:8081/ops/events
```

The Operations Console is intended to provide:
- overview/status
- services/Clawbots
- schedulers/jobs
- recent activity
- safe maintenance and scheduler actions

---

## 8. Validate the operations API

Run these from another terminal:

### Overview

```bash
curl http://127.0.0.1:8081/api/v1/ops/overview
```

### Services

```bash
curl http://127.0.0.1:8081/api/v1/ops/services
```

### Schedulers

```bash
curl http://127.0.0.1:8081/api/v1/ops/schedulers
```

### Events

```bash
curl http://127.0.0.1:8081/api/v1/ops/events
```

Expected outcome:
- overview data is returned
- service list is returned
- scheduler list is returned
- recent events are returned

---

## 9. Test safe maintenance actions

Use the Operations Console UI or the API.

### Put a service into maintenance mode

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/services/<service-id>/maintenance
```

### Resume a service

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/services/<service-id>/resume
```

### Pause a scheduler

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/schedulers/<scheduler-id>/pause
```

### Resume a scheduler

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/schedulers/<scheduler-id>/resume
```

### Run a scheduler once

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/schedulers/<scheduler-id>/run-once
```

Then re-check state:

```bash
curl http://127.0.0.1:8081/api/v1/ops/services/<service-id>
curl http://127.0.0.1:8081/api/v1/ops/schedulers/<scheduler-id>
curl http://127.0.0.1:8081/api/v1/ops/events
```

---

## 10. Local quality checks

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

This is useful before pushing changes or preparing a release.

---

## 11. Common startup issue: PostgreSQL role does not exist

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
make clean
docker compose down -v
make up
make smoke
SERVER_ADDRESS=127.0.0.1:8081 make run-server
```

### Alternative fix
If you need to preserve the volume, manually create the role/database in PostgreSQL.

---

## 12. Common startup issue: `/version` shows `dev` / `unknown`

If:

```bash
curl http://127.0.0.1:8081/version
```

returns:

```json
{"version":"dev","commit":"unknown","build_date":"unknown"}
```

then build metadata is not being injected during startup.

The current repository `Makefile` should handle this. If it does not, verify the local `Makefile` includes `-ldflags` injection for:
- version
- commit
- build date

---

## 13. Common issue: Operations Console template/render failures

If the UI returns a 500 error or template panic:
- confirm all templates parse successfully
- confirm `base.gohtml` renders only known named templates
- confirm the server no longer injects raw HTML with `template.HTML(...)`

This is especially relevant after recent UI/security hardening.

---

## 14. Stop the stack

To stop the local stack and remove volumes:

```bash
make clean
```

To stop only the foreground server process:

```bash
Ctrl+C
```

---

## 15. Recommended deployment flow

Use this sequence:

1. `cp .env.example .env`
2. replace required secrets
3. `make check-env`
4. `make up`
5. `make smoke`
6. `SERVER_ADDRESS=127.0.0.1:8081 make run-server`
7. verify `/healthz`, `/readyz`, `/version`
8. open `/ops`
9. validate ops APIs
10. run maintenance/scheduler actions
11. run quality checks if needed

---

## 16. What a successful deployment looks like

A deployment is in good shape if:

- foundation stack starts cleanly
- smoke checks pass
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

It covers:
- environment setup
- foundation stack startup
- service execution
- health verification
- Operations Console access
- basic API and action validation
- common troubleshooting

Use this as the primary onboarding document for deploying and validating `clawbot-server`.
