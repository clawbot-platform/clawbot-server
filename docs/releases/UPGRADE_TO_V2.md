# Upgrade to clawbot-server v2

## Who should use this

Use this guide when moving from pre-semver image tags (for example `drq-v1-*`) to the `v2` release stream.

## What changed for operators

- release tags now follow semver (`v2.0.0`) with a major stream alias (`v2`)
- Docker publishing is tag-driven (`v*`) and release-gated by `go test ./...`
- no control-plane contract break was introduced for `/api/v1` endpoints validated in the ACH integration

## Upgrade steps

1. Pull the v2 image:

```bash
docker pull ghcr.io/clawbot-platform/clawbot-server:v2.0.0
```

2. Confirm environment is still valid:

- `DATABASE_URL`
- `REDIS_URL`
- `NATS_URL`
- `INFERENCE_BASE_URL`
- optional `CLAWMEM_BASE_URL` when using scoped continuity integration

3. Restart service on `v2.0.0` (or `v2`) and run smoke checks:

```bash
curl -s http://127.0.0.1:8080/healthz | jq
curl -s http://127.0.0.1:8080/readyz | jq
curl -s http://127.0.0.1:8080/version | jq
```

4. Validate governed execution path with a known run/cycle workflow:

- create run
- create cycle
- start run
- fetch artifacts and comparison
- verify reviewer actions persist

## Rollback

Rollback is image-tag based. Redeploy the previous known-good immutable tag (`sha-...`) and repeat health/version checks.
