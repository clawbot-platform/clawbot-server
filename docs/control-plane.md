# Control Plane

## What the service provides

`clawbot-server` now provides a production-style control-plane contract for downstream trust-lab workloads:

- versioned `/api/v1` endpoints
- Postgres-backed persistence for runs, cycles, artifacts, model profiles, comparisons, bots, policies, and audit events
- execution contracts for `deterministic`, `llm`, and `dual` modes
- run types for `replay_run`, `agent_run`, and `week_run`
- week-run cycle lifecycle and review-state transitions
- artifact manifest indexing and retrieval
- dual-mode comparison storage for deterministic vs model-backed outputs
- policy decision point enforcement and persisted policy decisions
- execution rings (`ring_0` to `ring_3`) across runs/cycles
- reviewer action endpoints with auditable transitions
- hash-chained governance audit events
- clawmem namespace integration points (repo -> run -> cycle -> agent)
- configurable remote inference client wiring for ai-precision-style model hosts
- executable run paths for `agent_run` and cycle-scoped `week_run`

## Authoritative boundary

Deterministic replay remains the authoritative measurement plane. Model-backed outputs are stored as reviewable evidence and recommendations, not as silent replacements for deterministic scoring.

## Execution flow

When a run execution endpoint is invoked, the control plane:

1. validates run type and execution mode
2. resolves cycle context for week runs
3. fetches clawmem scoped context (`repo -> run -> cycle -> agent`)
4. executes deterministic and/or llm paths based on mode
5. persists output artifacts and (for dual mode) comparison records
6. persists scoped memory notes and snapshot references
7. updates run/cycle status (`completed` for deterministic-only; `review_pending` for llm/dual)

### Inference provider routing

- `provider=local_ollama` uses direct Ollama HTTP `POST /api/chat` with `stream=false`
- non-Ollama providers (for example `provider=gateway`) use `/api/v1/inference/execute`
- `base_url` from the model profile is honored per execution request, with server-level fallback to `INFERENCE_BASE_URL`
- per-phase timeouts are supported for primary, guardrail, and helper calls
- compact dual payload mode can be toggled with `ENABLE_COMPACT_DUAL_PAYLOAD`
- local validation can disable guardrails for Ollama with `LOCAL_OLLAMA_DISABLE_GUARDRAILS=true`
- local Guardian guardrail calls force `think:false`, `stream:false`, and use compact payloads
- ACH default Granite stack:
  - `ibm/granite3.3:8b` (primary)
  - `ibm/granite3.3-guardian:8b` (guardrail)
  - `granite4:3b` (helper)

## What does not belong here

The control plane does not embed ACH business logic directly. It does not own:

- NACHA policy semantics implementation
- detector engineering details
- replay engine internals
- final production decisioning pipelines

Those responsibilities stay in downstream domain workers such as `ach-trust-lab`.

## Package boundaries

- `internal/platform/runs` owns RunSpec contracts, cycle orchestration models, artifact registry records, model profiles, comparisons, and integration adapters.
- `internal/platform/runs` also owns native governance controls:
  - policy decision evaluation
  - execution ring enforcement
  - reviewer actions
  - guardrail fallback state mapping
  - governance hash-chain event persistence
- `internal/platform/audit` persists structured control-plane events.
- `internal/platform/scheduler` records scheduling intent.
- `internal/platform/store` holds common Postgres helpers and dashboard queries.
- `internal/http` owns API delivery.
- `internal/db` owns embedded migrations.
