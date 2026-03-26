# Foundation Stack

## What this stack includes

The repository ships one repeatable local platform stack built around:

- ZeroClaw
- OmniRoute
- PostgreSQL with `pgvector`
- Redis
- NATS
- MinIO
- Prometheus
- Grafana

The goal is to provide a shared base for control-plane and runtime-adjacent development without embedding vertical behavior in the platform repo.

## What is intentionally not included

This stack does not provide:

- vertical business logic
- simulation engines
- risk or fraud scoring
- replay or benchmark logic
- memory-service internals
- custom routing beyond OmniRoute integration
- custom runtime behavior beyond ZeroClaw integration

## Startup

```bash
cp .env.example .env
make up
make smoke
```

Run the control-plane service after the stack is ready:

```bash
make migrate-up
make run-server
```

Stop the stack:

```bash
make down
```

## Environment variables

The root `.env.example` contains the defaults used by Docker Compose and the smoke checker.

Replace these placeholders before using shared or remote environments:

- `POSTGRES_PASSWORD`
- `MINIO_ROOT_PASSWORD`
- `GRAFANA_ADMIN_PASSWORD`
- `ZEROCLAW_API_KEY`
- `ZEROCLAW_GATEWAY_TOKEN`

Common binding variables:

- `POSTGRES_HOST`
- `REDIS_HOST`
- `NATS_HOST`
- `MINIO_HOST`
- `PROMETHEUS_HOST`
- `GRAFANA_HOST`
- `OMNIROUTE_HOST`
- `ZEROCLAW_HOST`

By default, the stack binds to `127.0.0.1` so it stays local to the machine running Docker Compose.

## Volumes

Named volumes preserve local state for:

- `postgres_data`
- `redis_data`
- `nats_data`
- `minio_data`
- `omniroute_data`
- `zeroclaw_workspace`
- `prometheus_data`
- `grafana_data`

## Reuse guidance

Downstream projects can consume this stack as-is, pick only the services they need, or run the control plane beside their own application logic. The foundation is intentionally generic.
