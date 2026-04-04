# ADR 0001: Deterministic Authority with LLM and Dual Modes

## Status

Accepted - April 4, 2026

## Context

`ach-trust-lab` moved from local harness orchestration to a control-plane-backed architecture through `clawbot-server`. The target operating model needed to preserve reproducible fraud measurement while enabling model-backed reasoning and compliance gating.

## Decision

### 1. Deterministic replay remains authoritative

Deterministic replay is the source of truth for scored outcomes, regressions, and acceptance gates. This keeps baseline evidence reproducible and auditable.

### 2. `llm` and `dual` are first-class execution modes

- `llm` mode stores guarded reasoning outputs as versioned artifacts.
- `dual` mode runs deterministic and LLM-oriented paths together and persists a comparison object containing deterministic summary, LLM summary, deltas, review status, and final disposition.

### 3. clawmem is scoped context, not scored truth

Memory is integrated via explicit namespaces:

- repo namespace
- run namespace
- cycle namespace
- agent namespace

The control plane allows writing scoped notes and snapshot references, but final scored evidence remains in run artifacts and comparison records.

### 4. Granite / Guardian / helper routing is model-profile driven

The control plane stores model profiles with provider, endpoint, and model tags for:

- primary reasoning model
- guardrail model
- helper model

The seeded default profile maps to:

- `provider`: `local_ollama`
- `primary_model`: `ibm/granite3.3:8b`
- `guardrail_model`: `ibm/granite3.3-guardian:8b`
- `helper_model`: `granite4:3b`

Endpoint routing is configured by environment (`INFERENCE_BASE_URL`) to support remote ai-precision-style hosts without hardcoding host-specific IPs.

## Consequences

- deterministic evidence remains defensible for compliance and regression control
- model-backed reasoning is available without eroding replay authority
- week-run cycles gain auditable lineage across artifacts, comparisons, and memory snapshot references
- downstream ACH logic can evolve independently while using a stable control-plane contract
