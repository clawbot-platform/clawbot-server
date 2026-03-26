# Ports And Services

This document reflects the current **core vs optional** service split.

## Core services

| Service          | Port   | Purpose                                           |
|------------------|--------|---------------------------------------------------|
| PostgreSQL       | `5432` | Shared relational storage with `pgvector` enabled |
| Redis            | `6379` | Cache and short-lived coordination data           |
| NATS             | `4222` | Messaging backbone                                |
| NATS monitor     | `8222` | Local monitoring endpoint                         |
| `clawbot-server` | `8080` | Control-plane API and operations console          |

## Optional services

| Service       | Port    | Purpose                                   |
|---------------|---------|-------------------------------------------|
| MinIO API     | `9000`  | S3-compatible object storage              |
| MinIO console | `9001`  | Local object storage admin UI             |
| Prometheus    | `9090`  | Metrics collection and readiness endpoint |
| Grafana       | `3001`  | Dashboards and observability UI           |
| OmniRoute     | `20128` | Shared model gateway                      |
| ZeroClaw      | `3000`  | Shared runtime substrate gateway          |

## Service responsibilities

### Core services

- PostgreSQL stores durable control-plane state.
- Redis supports cache and short-lived coordination behavior.
- NATS provides eventing and asynchronous messaging support.
- `clawbot-server` runs against the core stack and persists control-plane metadata in PostgreSQL.

### Optional services

- OmniRoute is an optional model ingress for local development traffic and extended runtime experiments.
- ZeroClaw is an optional runtime substrate and can be configured to call OmniRoute rather than a provider-specific endpoint.
- Prometheus is an optional metrics backend for observability.
- Grafana is an optional dashboard layer for observability and management views.
- MinIO is an optional object storage layer for artifacts, reports, exports, and future replay bundles.

## Core vs optional deployment guidance

### Default/core path

Use only:

- PostgreSQL
- Redis
- NATS
- `clawbot-server`

This is the recommended path for:

- local development
- smoke testing
- lean deployments
- DRQ Version 1 dry runs

### Optional path

Enable optional services only when your deployment actually needs:

- dashboards
- metrics
- artifact storage
- runtime/model-gateway integrations

## DRQ Version 1 note

DRQ Version 1 uses only the **core services**:

- PostgreSQL
- Redis
- NATS
- `clawbot-server`

It does **not** require:

- MinIO
- Prometheus
- Grafana
- OmniRoute
- ZeroClaw
