
# TESTING-UI.md
## clawbot-server Operations Console

This document provides a practical UI testing flow for the embedded `clawbot-server` Operations Console.

It is intended for:
- local developer validation
- homelab/operator testing
- pre-PR UI smoke checks
- validating the generic operations console after recent changes

---

## Purpose

The Operations Console should provide a small, generic UI for:

- platform health
- service status
- scheduler status
- recent activity
- safe maintenance actions

This testing guide focuses on validating:
- page rendering
- navigation
- backend-to-UI consistency
- safe action behavior
- error handling

---

## Prerequisites

Make sure the following are available locally:

- Docker / container runtime
- Go toolchain
- `curl`
- a browser

Run commands from the `clawbot-server` repository root unless stated otherwise.

---

## 1. Start the local stack

Start the supporting stack and run the server:

```bash
make up
make smoke
SERVER_ADDRESS=127.0.0.1:8081 make run-server
```

Expected outcome:
- local stack comes up successfully
- smoke checks pass
- `clawbot-server` is listening on `127.0.0.1:8081`

If startup fails because of database role or migration issues, fix those first before testing the UI.

---

## 2. Basic health checks

In another terminal, verify the service is healthy:

```bash
curl http://127.0.0.1:8081/healthz
curl http://127.0.0.1:8081/readyz
curl http://127.0.0.1:8081/version
```

Expected outcome:
- health returns success
- readiness returns success
- version returns build metadata

---

## 3. Open the UI in a browser

Open the Operations Console home page:

```text
http://127.0.0.1:8081/ops
```

Also test these pages directly:

```text
http://127.0.0.1:8081/ops
http://127.0.0.1:8081/ops/services
http://127.0.0.1:8081/ops/schedulers
http://127.0.0.1:8081/ops/events
```

If detail pages are linked from the UI, click into them from the browser.

---

## 4. Overview page checks

On `/ops`, confirm:

- page loads with no 500 error
- the page title/header renders
- navigation sidebar is visible
- summary cards or overview metrics render
- status values are visible and readable
- styles load correctly
- no raw template markers are visible
- no “unknown page” fallback appears unexpectedly

Expected outcome:
- overview page renders cleanly and looks like an operations dashboard, not a broken template

---

## 5. Services page checks

Open:

```text
http://127.0.0.1:8081/ops/services
```

Confirm:

- services list/table renders
- each row/card shows expected fields such as:
  - name
  - status
  - version
  - maintenance mode
  - last heartbeat or equivalent
- status pills render correctly
- no broken links or empty placeholders appear unexpectedly

If detail links exist:
- click through to at least one service detail page

Expected outcome:
- services page clearly reflects platform service state

---

## 6. Service detail page checks

For at least one service, confirm:

- detail page loads successfully
- service name and status are visible
- maintenance mode state is visible
- dependency health and/or recent error information is visible if supported
- page layout is readable
- navigation back to services or overview works

Expected outcome:
- service detail page exposes operationally useful information

---

## 7. Schedulers page checks

Open:

```text
http://127.0.0.1:8081/ops/schedulers
```

Confirm:

- scheduler list renders
- each scheduler shows fields such as:
  - enabled state
  - interval
  - last run
  - next run
  - last result
- scheduler action buttons render:
  - pause
  - resume
  - run once

Expected outcome:
- schedulers page clearly shows current scheduler state

---

## 8. Events page checks

Open:

```text
http://127.0.0.1:8081/ops/events
```

Confirm:

- recent events/activity render correctly
- timestamps are readable
- event messages are understandable
- empty state looks intentional if there are no events
- page does not break when event list is short or empty

Expected outcome:
- recent activity is visible and useful for operators

---

## 9. Test maintenance actions in the UI

From the browser UI, if controls are present:

### Put a service into maintenance mode
- click the maintenance action for a service

Confirm:
- UI action succeeds
- service status updates visibly
- maintenance mode is reflected in the page

### Resume a service
- click the resume action

Confirm:
- maintenance mode clears
- state updates correctly in the UI

Expected outcome:
- maintenance actions are safe and visible

---

## 10. Test scheduler actions in the UI

From the browser UI:

### Pause a scheduler
- click pause

Confirm:
- enabled state changes
- scheduler appears paused

### Resume a scheduler
- click resume

Confirm:
- scheduler becomes active again

### Run a scheduler once
- click run once

Confirm:
- last run metadata updates or the action is visible in recent activity

Expected outcome:
- scheduler actions work and result in visible state changes

---

## 11. Verify UI actions through the API

While the UI is open, confirm the backend state matches the UI state.

### Overview

```bash
curl http://127.0.0.1:8081/api/v1/ops/overview
```

### Services

```bash
curl http://127.0.0.1:8081/api/v1/ops/services
```

### Schedulers

```bash
curl http://127.0.0.1:8081/api/v1/ops/schedulers
```

### Events

```bash
curl http://127.0.0.1:8081/api/v1/ops/events
```

Expected outcome:
- API responses match what the UI shows
- maintenance mode or scheduler state changes are reflected consistently
- events include the actions just performed

---

## 12. Negative-path browser checks

Try to validate the UI behaves sensibly when data is missing or invalid.

### Invalid detail page
If routes support direct IDs, try:
- an invalid service detail URL
- an invalid scheduler detail URL

Expected outcome:
- UI shows a safe error or not-found response
- no broken template or server panic occurs

### Empty state
If you can simulate an empty state:
- no services
- no events
- no schedulers

Expected outcome:
- page remains readable
- empty state is intentional and not broken

---

## 13. Browser console / network checks

Open browser developer tools and confirm:

- no JavaScript errors if any browser-side behavior exists
- no repeated failing network requests
- UI actions map to expected HTTP requests
- no mixed-content or asset-loading issues

Expected outcome:
- the console is quiet or only shows harmless development noise

---

## 14. Automated validation alongside UI testing

Run the backend quality suite as part of UI validation:

```bash
go test ./...
go vet ./...
golangci-lint run ./...
gosec ./...
govulncheck ./...
```

Optional coverage report:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

Expected outcome:
- UI changes did not break backend quality or tests

---

## 15. Recommended quick UI smoke flow

Use this sequence for fast validation:

1. Start the stack:
   ```bash
   make up
   make smoke
   SERVER_ADDRESS=127.0.0.1:8081 make run-server
   ```

2. Open:
  - `/ops`
  - `/ops/services`
  - `/ops/schedulers`
  - `/ops/events`

3. Perform:
  - one maintenance action
  - one scheduler action

4. Confirm state via API:
   ```bash
   curl http://127.0.0.1:8081/api/v1/ops/services
   curl http://127.0.0.1:8081/api/v1/ops/schedulers
   curl http://127.0.0.1:8081/api/v1/ops/events
   ```

5. Run backend quality checks

6. Stop the stack:
   ```bash
   make clean
   ```

---

## 16. What a good result looks like

A UI change is in good shape if:

- all Operations Console pages load
- layout/styles render correctly
- no template/render panic occurs
- overview/services/schedulers/events pages are usable
- maintenance mode actions work
- scheduler actions work
- API state matches UI state
- invalid states fail safely
- backend quality checks remain green

---

## Notes

This testing guide assumes the Operations Console is:
- generic
- internal
- operational
- separate from domain-specific UIs like Clawbot Trust Lab

Use this document together with:
- `TESTING.md`
- CI workflows
- SonarCloud quality gate results
```
