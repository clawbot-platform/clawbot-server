# Ports And Services

| Service | Port | Purpose |
| --- | --- | --- |
| PostgreSQL | `5432` | Shared relational storage with `pgvector` enabled |
| Redis | `6379` | Cache and short-lived coordination data |
| NATS | `4222` | Messaging backbone |
| NATS monitor | `8222` | Local monitoring endpoint |
| MinIO API | `9000` | S3-compatible object storage |
| MinIO console | `9001` | Local object storage admin UI |
| Prometheus | `9090` | Metrics collection and readiness endpoint |
| Grafana | `3001` | Dashboards and observability UI |
| OmniRoute | `20128` | Shared model gateway |
| ZeroClaw | `3000` | Shared runtime substrate gateway |
| clawbot-server | `8080` | Control-plane API |

## Service responsibilities

- OmniRoute is the intended model ingress for local development traffic.
- ZeroClaw is configured to call OmniRoute rather than a provider-specific endpoint.
- Prometheus scrapes the baseline metrics surface exposed by the stack.
- Grafana is pre-provisioned with a Prometheus datasource and a starter dashboard.
- `clawbot-server` runs against the foundation stack and persists control-plane metadata in Postgres.
