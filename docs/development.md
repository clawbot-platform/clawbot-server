# Development

## Local workflow

The current local development model is:

- **core infrastructure in Docker**
- **`clawbot-server` running on the host**

That means:
- Docker Compose starts the foundation services
- the control-plane binary is run locally with the repository `Makefile`

## Core workflow

### 1. Copy `.env.example` to `.env`

```bash
cp .env.example .env
```

### 2. Validate the core environment

```bash
./scripts/check-env.sh .env
```

### 3. Start the core stack

```bash
make up
```

This starts the repository’s core Docker Compose stack:

- PostgreSQL
- Redis
- NATS

### 4. Apply migrations

```bash
make migrate-up
```

### 5. Run the control-plane service

```bash
make run-server
```

### 6. Verify readiness

```bash
make smoke
```

### 7. Inspect services

```bash
make ps
make logs
```

## Optional services workflow

Optional services are **not** started by default.

They are provided through `deploy/compose/docker-compose.optional.yml`.

### Validate optional environment values

```bash
VALIDATE_OPTIONAL_STACK=1 ./scripts/check-env.sh .env
```

### Start core + optional stack

```bash
docker compose --env-file .env \
  -f deploy/compose/docker-compose.yml \
  -f deploy/compose/docker-compose.override.yml \
  -f deploy/compose/docker-compose.optional.yml \
  up -d
```

Use this only when you need:

- MinIO
- OmniRoute
- ZeroClaw
- Prometheus
- Grafana

## Make targets

### Foundation and runtime

- `make up`
- `make down`
- `make restart`
- `make ps`
- `make logs`
- `make smoke`
- `make run-server`
- `make migrate-up`
- `make migrate-down`
- `make clean`

### Quality and validation

- `make lint`
- `make test`
- `make security`
- `make compose-validate`

## Notes on `make up`

`make up` is intended to start the **core** foundation stack only.

Optional services should be enabled explicitly with the optional Compose file rather than being part of the default development path.

## Recommended development behavior

### Use the core stack by default

Most local development should use only:

- PostgreSQL
- Redis
- NATS
- `clawbot-server`

This keeps the repo’s day-to-day workflow lean and aligned with the current documented deployment path.

### Enable optional services only when needed

Use the optional stack when you are working on:

- observability
- dashboards
- artifact storage
- runtime/model-gateway integrations

## Contributor guidance

- Keep the repo Go-first.
- Put compose files under `deploy/compose`.
- Keep the core vs optional split explicit.
- Put versioned configs under `configs`.
- Keep docs explicit and recruiter-friendly.
- Keep downstream business logic outside this repository.
- Treat downstream verticals as consumer examples, not design constraints.

## DRQ Version 1 note

For DRQ Version 1 dry runs and benchmark validation, use the **core stack only** unless there is a specific need for additional observability or artifact storage.
