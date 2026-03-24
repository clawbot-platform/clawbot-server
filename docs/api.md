# API

Phase 1 exposes a versioned control-plane API under `/api/v1`.

## System endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /version`

## Dashboard scaffold

- `GET /api/v1/dashboard/summary`

Returns aggregate counts for runs, bots, policies, and audit events. This is a backend-ready shell endpoint, not a UI implementation.

## Runs

- `GET /api/v1/runs`
- `POST /api/v1/runs`
- `GET /api/v1/runs/{id}`
- `PATCH /api/v1/runs/{id}`

Example create request:

```json
{
  "name": "trust-lab-baseline",
  "description": "Platform-created run scaffold",
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
  "name": "trust-analyst",
  "role": "review",
  "runtime": "zeroclaw",
  "status": "active",
  "repo_hint": "clawbot-trust-lab",
  "version": "phase-1",
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
  "description": "Generic control-plane placeholder policy",
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
- Policy behavior is intentionally generic in Phase 1; `enabled` is the main operational field.
