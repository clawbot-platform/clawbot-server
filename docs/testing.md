
# TESTING.md
## clawbot-server

This document provides a practical test flow for validating recent changes in `clawbot-server`, including:

- local quality checks
- coverage generation
- local stack startup
- health checks
- operations API validation
- maintenance and scheduler action validation

---

## Prerequisites

Make sure the following are available locally:

- Go toolchain
- Docker / container runtime
- `golangci-lint`
- `gosec`
- `govulncheck`
- `curl`

Run all commands from the `clawbot-server` repository root unless stated otherwise.

---

## 1. Fast local quality checks

Run the full local quality suite:

```bash
go test ./...
go vet ./...
golangci-lint run ./...
gosec ./...
govulncheck ./...
```

Expected outcome:
- all tests pass
- no vet errors
- no linter failures
- no blocking security or vulnerability issues

---

## 2. Coverage check

Generate Go coverage:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -func=coverage.out
```

Optional HTML coverage report:

```bash
go tool cover -html=coverage.out -o coverage.html
```

Then open `coverage.html` in your browser.

Expected outcome:
- coverage report is generated successfully
- recently added code paths have meaningful test coverage

---

## 3. Start the local stack

Bring up the supporting stack and start the server:

```bash
make up
make smoke
SERVER_ADDRESS=127.0.0.1:8081 make run-server
```

Expected outcome:
- local dependencies start successfully
- smoke checks pass
- `clawbot-server` starts on `127.0.0.1:8081`

---

## 4. Basic health checks

In a separate terminal, validate service health:

```bash
curl http://127.0.0.1:8081/healthz
curl http://127.0.0.1:8081/readyz
curl http://127.0.0.1:8081/version
```

Expected outcome:
- health returns success
- readiness returns success
- version endpoint returns build/version metadata

---

## 5. Test the operations API

### Overview

```bash
curl http://127.0.0.1:8081/api/v1/ops/overview
```

Validate:
- overall platform status is returned
- high-level service/scheduler information is present

### Services list

```bash
curl http://127.0.0.1:8081/api/v1/ops/services
```

Validate:
- service list is returned
- each service has fields such as:
  - id
  - name
  - status
  - version
  - maintenance mode
  - last heartbeat or equivalent

### Service detail

Replace `<service-id>` with one of the service IDs returned above.

```bash
curl http://127.0.0.1:8081/api/v1/ops/services/<service-id>
```

Validate:
- detailed service status is returned
- dependency / status / error details are visible

### Schedulers list

```bash
curl http://127.0.0.1:8081/api/v1/ops/schedulers
```

Validate:
- scheduler list is returned
- each scheduler exposes:
  - id
  - enabled state
  - interval
  - last run
  - next run
  - last result

### Scheduler detail

Replace `<scheduler-id>` with one of the scheduler IDs returned above.

```bash
curl http://127.0.0.1:8081/api/v1/ops/schedulers/<scheduler-id>
```

Validate:
- scheduler detail is returned
- scheduling state and recent execution information are visible

### Recent events

```bash
curl http://127.0.0.1:8081/api/v1/ops/events
```

Validate:
- recent events/activity are returned
- event list is readable and relevant

---

## 6. Test safe write actions

### Put service into maintenance mode

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/services/<service-id>/maintenance
```


Validate:
- request succeeds
- maintenance mode is enabled for the service

### Resume service from maintenance mode

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/services/<service-id>/resume
```

Validate:
- request succeeds
- maintenance mode is cleared

### Pause scheduler

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/schedulers/<scheduler-id>/pause
```

Validate:
- scheduler is paused
- enabled state changes accordingly

### Resume scheduler

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/schedulers/<scheduler-id>/resume
```

Validate:
- scheduler resumes correctly
- next run is visible again if applicable

### Run scheduler once

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/schedulers/<scheduler-id>/run-once
```

Validate:
- request succeeds
- recent run history or last-run metadata updates

---

## 7. Re-check state after write actions

After running maintenance or scheduler actions, re-query the resources:

```bash
curl http://127.0.0.1:8081/api/v1/ops/services/<service-id>
curl http://127.0.0.1:8081/api/v1/ops/schedulers/<scheduler-id>
curl http://127.0.0.1:8081/api/v1/ops/events
```

Validate:
- state changes are persisted
- events reflect the operator actions performed

---

## 8. Negative-path checks

These are useful for validating handler behavior.

### Invalid service ID

```bash
curl http://127.0.0.1:8081/api/v1/ops/services/does-not-exist
```

### Invalid scheduler ID

```bash
curl http://127.0.0.1:8081/api/v1/ops/schedulers/does-not-exist
```

### Invalid maintenance action target

```bash
curl -X POST http://127.0.0.1:8081/api/v1/ops/services/does-not-exist/maintenance
```

Expected outcome:
- API returns appropriate error response
- invalid requests do not crash the server

---

## 9. CI-equivalent validation

This is the closest local equivalent to what CI should enforce:

```bash
go test ./... -coverprofile=coverage.out
go vet ./...
golangci-lint run ./...
gosec ./...
govulncheck ./...
```

If SonarCloud coverage is configured, make sure `coverage.out` is generated successfully before pushing.

---



## 10. Shutdown

When finished, stop the local stack:

```bash
make clean
```

If `run-server` is running in the foreground, stop it with:

```bash
Ctrl+C
```

---

## Recommended test flow

For quick validation after a change:

1. `go test ./...`
2. `go vet ./...`
3. `golangci-lint run ./...`
4. `go test ./... -coverprofile=coverage.out`
5. `make up`
6. `make smoke`
7. `SERVER_ADDRESS=127.0.0.1:8081 make run-server`
8. Run the ops API checks with `curl`
9. Run maintenance/scheduler actions
10. Re-check state and events
11. `make clean`

---

## What a good result looks like

A change is in good shape if:

- all backend quality checks pass
- coverage report is generated cleanly
- the server starts without issue
- health/readiness/version endpoints work
- ops overview/services/schedulers/events endpoints work
- maintenance mode and scheduler actions work
- invalid requests fail safely
- no unexpected crashes or state corruption occur

---

## Notes

This document is intended for:
- local developer validation
- homelab/operator testing
- pre-PR or pre-release smoke validation

For CI and SonarCloud validation, use this together with the repository workflows and quality gate checks.

