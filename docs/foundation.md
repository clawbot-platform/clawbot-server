# Phase 0 Foundation

## What Phase 0 includes

Phase 0 establishes one repeatable shared home-lab foundation stack for `clawbot-platform`. It focuses on shared infrastructure and platform bootstrap only.

Included services:

- ZeroClaw
- OmniRoute
- PostgreSQL with `pgvector`
- Redis
- NATS
- MinIO
- Prometheus
- Grafana

## What is intentionally not included

This phase does not add:

- trust-lab business logic
- simulation logic
- fraud or risk scoring logic
- Red Queen logic
- `clawmem` internals
- custom model routing beyond OmniRoute integration
- custom agent runtime features beyond ZeroClaw integration

## Startup

```bash
cp .env.example .env
make up
make smoke
```

Stop the stack:

```bash
make down
```

Run the Phase 1 control-plane service after the foundation is up:

```bash
make migrate-up
make run-server
```

## Environment variables

The root `.env.example` contains the runtime defaults used by Docker Compose and the smoke checker.

Required secret replacements:

- `POSTGRES_PASSWORD`
- `MINIO_ROOT_PASSWORD`
- `GRAFANA_ADMIN_PASSWORD`
- `ZEROCLAW_API_KEY`
- `ZEROCLAW_GATEWAY_TOKEN`

Optional local binding variables:

- `POSTGRES_HOST`
- `REDIS_HOST`
- `NATS_HOST`
- `MINIO_HOST`
- `PROMETHEUS_HOST`
- `GRAFANA_HOST`
- `OMNIROUTE_HOST`
- `ZEROCLAW_HOST`

By default, these host bindings are set to `127.0.0.1` so Phase 0 stays local to the workstation running Docker Compose.

## Ports

- All published ports bind to `127.0.0.1` by default.
- `5432` PostgreSQL
- `6379` Redis
- `4222` NATS client
- `8222` NATS monitoring
- `9000` MinIO API
- `9001` MinIO console
- `9090` Prometheus
- `3001` Grafana
- `20128` OmniRoute
- `3000` ZeroClaw

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

## Future integration points

- `clawbot-trust-lab` will consume this stack later as the primary platform foundation.
- `clawmem` will join later as a separate service rather than being implemented here.
- other verticals such as watchlist review can reuse the same shared foundation.
