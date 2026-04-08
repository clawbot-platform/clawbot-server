# API

`clawbot-server` exposes a versioned control-plane API under `/api/v1`.

This surface now includes a first-class run execution contract for deterministic, llm, and dual-mode workflows plus week-run cycle orchestration, artifact manifests, comparison objects, and model profile registration.

## System endpoints

- `GET /healthz`
- `GET /readyz`
- `GET /version`

## Dashboard summary

- `GET /api/v1/dashboard/summary`

## Operations console

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

## Runs and RunSpec

- `GET /api/v1/runs`
- `POST /api/v1/runs`
- `GET /api/v1/runs/{id}`
- `PATCH /api/v1/runs/{id}`
- `POST /api/v1/runs/{id}/start`

RunSpec-compatible fields on create/update include:

- `run_type`: `replay_run | agent_run | week_run`
- `execution_mode`: `deterministic | llm | dual`
- `execution_ring`: `ring_0 | ring_1 | ring_2 | ring_3`
- `repo`, `domain`
- `dataset_refs`
- `prompt_pack_version`
- `rule_pack_version`
- `model_profile`
- `guardrail_profile`
- `memory_namespace`
- `requested_by`
- `started_at`, `finished_at`
- `status`
- `guardrail_status`
- `artifact_bundle_refs`
- `review_metadata_json`
- `notes`

### Example run create request

```json
{
  "name": "ach-week-1",
  "description": "NACHA 2026 dual-mode showcase run",
  "run_type": "week_run",
  "execution_mode": "dual",
  "status": "pending",
  "repo": "ach-trust-lab",
  "domain": "ach",
  "dataset_refs": ["data/samples/sample_ach_events.json"],
  "prompt_pack_version": "ach-week/v1",
  "rule_pack_version": "detectors/2026.1",
  "model_profile": "ach-default",
  "guardrail_profile": "ach-guardian-default",
  "memory_namespace": {
    "repo_namespace": "ach-trust-lab",
    "run_namespace": "weekrun-2026-06-demo"
  },
  "review_metadata_json": {
    "approval_required": true
  },
  "notes": "Deterministic replay remains authoritative"
}
```

## Artifacts

- `GET /api/v1/runs/{id}/artifacts`
- `POST /api/v1/runs/{id}/artifacts`

Artifact request fields:

- `cycle_id` (optional)
- `artifact_type`
- `uri`
- `content_type`
- `version`
- `checksum`
- `metadata_json`

## Cycles (week-run orchestration)

- `POST /api/v1/runs/{id}/cycles`
- `GET /api/v1/runs/{id}/cycles/{cycleID}`
- `PATCH /api/v1/runs/{id}/cycles/{cycleID}`
- `POST /api/v1/runs/{id}/cycles/{cycleID}/execute`

Cycle fields include:

- `cycle_key` (`day-1` through `day-7` recommended)
- `focus`
- `objective`
- `detector_pack`
- `execution_ring`
- `summary_ref`
- `carry_forward_summary_ref`
- `status`
- `memory_snapshot_ref` (response field)

Supported cycle statuses:

- `pending`
- `running`
- `review_pending`
- `approved`
- `rejected`
- `completed`
- `failed`
- `cancelled`
- `guardrail_deferred`
- `failed_runtime`
- `failed_policy`
- `overridden`
- `deferred`

## Dual-mode comparison

- `GET /api/v1/runs/{id}/comparison`
- `POST /api/v1/runs/{id}/comparison`

Comparison payload supports:

- `deterministic_summary`
- `llm_summary`
- `guardrail_summary`
- `deltas`
- `review_status`
- `reviewer_notes`
- `final_disposition`
- `final_output`

## Runtime execution behavior

`POST /api/v1/runs/{id}/start` is the execution path for `agent_run` (and can run `week_run` when `cycle_id` is supplied).  
`POST /api/v1/runs/{id}/cycles/{cycleID}/execute` is the cycle-scoped execution path for `week_run`.

Execution behavior by mode:

- `deterministic`
  - builds deterministic summary output
  - persists deterministic artifact metadata
  - marks execution as completed
- `llm`
  - fetches scoped memory context first
  - loads model profile
  - routes inference by provider:
    - `local_ollama`: direct Ollama HTTP `POST /api/chat` with `stream=false`
    - `gateway` (or other providers): `POST /api/v1/inference/execute`
  - applies per-phase timeout controls (primary / guardrail / helper)
  - persists LLM output (and guardrail report when present)
  - marks execution as `review_pending` or `guardrail_deferred` based on guardrail outcome
- `dual`
  - executes deterministic and llm paths in one execution
  - uses compact dual payload mode when enabled (`ENABLE_COMPACT_DUAL_PAYLOAD=true`)
  - persists both output artifacts
  - upserts a comparison object (`deterministic_summary`, `llm_summary`, `guardrail_summary`, `deltas`)
  - marks execution as `review_pending`

Guardrail outcome statuses surfaced in run/cycle metadata:

- `guardrail_passed`
- `guardrail_flagged`
- `guardrail_timeout`
- `guardrail_unavailable`
- `guardrail_disabled`

Execution request payload:

```json
{
  "cycle_id": "optional-cycle-id-for-week-run-start",
  "agent_namespace": "daily-summary",
  "prompt": "analyze this cycle",
  "system_prompt": "optional system prompt",
  "input_json": {
    "context": "additional domain input"
  },
  "memory_note": "optional explicit carry-forward note"
}
```

During execution, clawmem integration behavior is:

- fetch scoped context before execution (repo/run/cycle/agent namespace)
- persist scoped notes after execution
- attach returned memory snapshot reference to run and cycle metadata when provided

## Model profiles

- `POST /api/v1/model-profiles`
- `GET /api/v1/model-profiles/{idOrName}`

`ach-default` is seeded in migrations with:

- `provider`: `local_ollama`
- `primary_model`: `ibm/granite3.3:8b`
- `guardrail_model`: `ibm/granite3.3-guardian:8b`
- `helper_model`: `granite4:3b`

Provider behavior is explicit:

- `local_ollama`:
  - uses direct Ollama HTTP APIs
  - does **not** call `/api/v1/inference/execute`
  - defaults to model profile `base_url` (or falls back to `INFERENCE_BASE_URL`)
  - guardrail requests force `think:false`, `stream:false`, and compact payloads
  - can disable guardrails for local validation with `LOCAL_OLLAMA_DISABLE_GUARDRAILS=true`
- `gateway` (or other non-Ollama providers):
  - uses control-plane inference gateway path `/api/v1/inference/execute`

Optional timeout/environment controls:

- `GUARDRAIL_TIMEOUT` (duration)
- `HELPER_TIMEOUT` (duration)

## Reviewer actions

Run-level reviewer endpoints:

- `POST /api/v1/runs/{id}/approve`
- `POST /api/v1/runs/{id}/reject`
- `POST /api/v1/runs/{id}/override`
- `POST /api/v1/runs/{id}/defer`

Request payload:

```json
{
  "reviewer_id": "reviewer-123",
  "reviewer_type": "human",
  "rationale": "manual compliance decision",
  "cycle_id": "optional-cycle-id",
  "policy_decision_id": "optional-policy-decision-id"
}
```

## Governance controls

The control plane now evaluates a policy decision point before:

- run create
- cycle create
- run/cycle execution
- artifact attach
- reviewer actions

Policy decisions and governance audit events are persisted with hash-chained sequencing metadata.

## Control-plane dependencies

- `GET /api/v1/control-plane/dependencies`

Returns current readiness details for:

- postgres
- clawmem integration endpoint
- inference endpoint

## Bots

- `GET /api/v1/bots`
- `POST /api/v1/bots`
- `GET /api/v1/bots/{id}`
- `PATCH /api/v1/bots/{id}`

## Policies

- `GET /api/v1/policies`
- `POST /api/v1/policies`
- `GET /api/v1/policies/{id}`
- `PATCH /api/v1/policies/{id}`

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
