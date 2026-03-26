# API

`clawbot-server` exposes a versioned control-plane API under `/api/v1`.

The API is intentionally generic. It exists to manage platform metadata and audit-friendly lifecycle records that downstream projects can reuse.

## System endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /version`

## Dashboard summary

- `GET /api/v1/dashboard/summary`

Returns aggregate counts for runs, bots, policies, and audit events.

## Operations console

The operations console is a generic platform surface for service health, scheduler state, and recent operator-relevant activity.

Read endpoints:

- `GET /api/v1/ops/overview`
- `GET /api/v1/ops/services`
- `GET /api/v1/ops/services/{id}`
- `GET /api/v1/ops/schedulers`
- `GET /api/v1/ops/schedulers/{id}`
- `GET /api/v1/ops/events`

Safe write endpoints:

- `POST /api/v1/ops/services/{id}/maintenance`
- `POST /api/v1/ops/services/{id}/resume`
- `POST /api/v1/ops/schedulers/{id}/pause`
- `POST /api/v1/ops/schedulers/{id}/resume`
- `POST /api/v1/ops/schedulers/{id}/run-once`

Example overview response:

```json
{
  "data": {
    "status": "degraded",
    "services_total": 3,
    "services_healthy": 2,
    "services_degraded": 1,
    "services_down": 0,
    "services_maintenance": 0,
    "schedulers_active": 2,
    "schedulers_paused": 1,
    "recent_failures": 2,
    "last_updated_at": "2026-03-26T15:00:00Z"
  }
}
```

Example service response:

```json
{
  "data": {
    "id": "clawbot-server",
    "name": "clawbot-server",
    "service_type": "control-plane",
    "status": "healthy",
    "version": "dev",
    "uptime_seconds": 11520,
    "last_heartbeat_at": "2026-03-26T14:59:45Z",
    "maintenance_mode": false,
    "last_error": "",
    "dependency_status": {
      "postgres": "healthy",
      "redis": "healthy",
      "nats": "healthy"
    }
  }
}
```

Example scheduler response:

```json
{
  "data": {
    "id": "control-plane-sync",
    "name": "Control-plane sync",
    "enabled": true,
    "interval_seconds": 300,
    "last_run_at": "2026-03-26T14:58:00Z",
    "next_run_at": "2026-03-26T15:03:00Z",
    "last_result": "ok",
    "last_duration_ms": 140,
    "last_error": ""
  }
}
```

Example recent activity response:

```json
{
  "data": [
    {
      "id": "evt-003",
      "time": "2026-03-26T14:55:00Z",
      "source": "downstream-app",
      "event_type": "service.degraded",
      "severity": "warn",
      "message": "downstream-app reported delayed heartbeats and entered a degraded state."
    }
  ]
}
```

## Runs

- `GET /api/v1/runs`
- `POST /api/v1/runs`
- `GET /api/v1/runs/{id}`
- `PATCH /api/v1/runs/{id}`

Example create request:

```json
{
  "name": "platform-baseline",
  "description": "Reusable control-plane run scaffold",
  "status": "pending",
  "scenario_type": "placeholder",
  "metadata_json": {
    "owner": "platform"
  }
}
```

## Bots

- `GET /api/v1/bots`
- `POST /api/v1/bots`
- `GET /api/v1/bots/{id}`
- `PATCH /api/v1/bots/{id}`

Example create request:

```json
{
  "name": "shared-runtime-operator",
  "role": "review",
  "runtime": "zeroclaw",
  "status": "active",
  "repo_hint": "example-consumer-repo",
  "version": "v1",
  "config_json": {
    "provider": "omniroute"
  }
}
```

## Policies

- `GET /api/v1/policies`
- `POST /api/v1/policies`
- `GET /api/v1/policies/{id}`
- `PATCH /api/v1/policies/{id}`

Example create request:

```json
{
  "name": "default-safety-policy",
  "category": "safety",
  "version": "v1",
  "enabled": true,
  "description": "Generic control-plane policy example",
  "rules_json": {
    "mode": "placeholder"
  }
}
```

## Response shape

Successful responses return:

```json
{
  "data": {}
}
```

Errors return:

```json
{
  "error": {
    "code": "bad_request",
    "message": "name is required"
  }
}
```

## Status notes

- Run statuses are scaffolded as `pending`, `scheduled`, `running`, `completed`, `failed`, `cancelled`.
- Bot statuses are scaffolded as `active`, `inactive`, `deprecated`.
- Policy behavior is intentionally generic; `enabled` is the main operational field.
- Operations console service statuses are `healthy`, `degraded`, `down`, `maintenance`.
- Sample names in this document are illustrative only. The API is reusable by projects inside or outside the Clawbot organization.
