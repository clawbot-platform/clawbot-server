# Foundation Stack

## What this stack includes

This repository now uses a **core vs optional** foundation model.

### Core stack

The default/core stack is:

- PostgreSQL with `pgvector`
- Redis
- NATS

This is the recommended baseline for:

- local development
- smoke testing
- control-plane validation
- lean DRQ Version 1 dry runs
- downstream projects that only need the core control-plane foundation

### Optional stack

Optional services are available through a separate Compose file:

- MinIO
- OmniRoute
- ZeroClaw
- Prometheus
- Grafana

These services are useful for:

- object and artifact storage
- extended runtime integrations
- model gateway experimentation
- observability
- dashboards

They are **not required** for the default/core deployment path.

## What is intentionally not included

This stack does not provide:

- vertical business logic
- simulation engines
- risk or fraud scoring
- replay or benchmark logic
- memory-service internals
- custom routing beyond optional OmniRoute integration
- custom runtime behavior beyond optional ZeroClaw integration

## Core vs optional Compose files

The repository uses three Compose files:

- `deploy/compose/docker-compose.yml`
  - core services only
- `deploy/compose/docker-compose.override.yml`
  - local development tweaks for the core stack only
- `deploy/compose/docker-compose.optional.yml`
  - optional services

## Startup

### Prepare the environment file

```bash
cp .env.example .env
```

### Validate the core environment

```bash
./scripts/check-env.sh .env
```

### Validate the optional stack only when needed

```bash
VALIDATE_OPTIONAL_STACK=1 ./scripts/check-env.sh .env
```

### Start the core stack

You can use the repository Make target:

```bash
make up
```

Or use Docker Compose directly:

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  up -d
```

### Start optional services only when needed

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  -f deploy/compose/docker-compose.optional.yml \
  up -d
```

### Run the control-plane service after the stack is ready

```bash
make migrate-up
make run-server
```

### Validate readiness

```bash
make smoke
```

### Stop the core stack

```bash
make down
```

Or directly:

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  down
```

### Stop core + optional stack

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  -f deploy/compose/docker-compose.optional.yml \
  down
```

## Environment variables

The root `.env.example` is aligned with the core vs optional split.

### Core/default variables

These are used for the normal lean deployment path:

- `COMPOSE_PROJECT_NAME`
- `POSTGRES_*`
- `REDIS_*`
- `NATS_*`
- `STACK_SMOKE_TIMEOUT`
- `APP_ENV`
- `SERVER_ADDRESS`
- `LOG_LEVEL`
- `AUTO_MIGRATE`
- `SHUTDOWN_TIMEOUT`
- `DATABASE_URL`

### Optional variables

These are only required when enabling `docker-compose.optional.yml`:

- `MINIO_*`
- `PROMETHEUS_*`
- `GRAFANA_*`
- `OMNIROUTE_*`
- `ZEROCLAW_*`

Replace placeholder secrets before using optional services in shared or remote environments.

By default, the stack binds to `127.0.0.1` so it stays local to the machine running Docker Compose.

## Volumes

Named volumes preserve local state for the current stack split.

### Core volumes

- `postgres_data`
- `redis_data`
- `nats_data`

### Optional volumes

- `minio_data`
- `omniroute_data`
- `zeroclaw_workspace`
- `prometheus_data`
- `grafana_data`

## Reuse guidance

Downstream projects can:

- consume the core stack as-is
- enable only the optional services they actually need
- run the control plane beside their own application logic
- keep vertical business logic outside this repository

The foundation remains intentionally generic.

## DRQ Version 1 note

DRQ Version 1 should use the **core stack only**:

- PostgreSQL
- Redis
- NATS
- `clawbot-server`

It does **not** require:

- MinIO
- OmniRoute
- ZeroClaw
- Prometheus
- Grafana

This keeps the Version 1 deployment lean, reduces container count, and avoids unnecessary moving parts during dry runs and benchmark validation.
