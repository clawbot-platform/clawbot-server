# Architecture

## Repository role

`clawbot-server` is the reusable control-plane repository in `clawbot-platform` and now supports compliance-oriented orchestration contracts for trust-lab workloads.

## Layered planes

The current architecture separates concerns into four planes:

- deterministic evidence plane: authoritative replay scoring and durable metadata
- reasoning plane: model-backed synthesis and recommendation artifacts
- guardrail/review plane: compliance gating and adjudication records
- memory context plane: scoped context through clawmem namespaces

## Control-plane entities

The server persists and serves first-class contracts for:

- RunSpec (`run_type`, `execution_mode`, model/rule/prompt/memory scope metadata)
- cycle orchestration (`day-1` through `day-7` style lifecycle units)
- artifact registry manifests
- model profiles (primary, guardrail, helper routes)
- dual-mode comparisons and reviewer disposition metadata

## Boundary decisions

- deterministic replay is authoritative for scored outcomes and regressions
- LLM outputs are first-class but reviewable synthesis artifacts
- memory improves continuity but is not the source of truth for final scored evidence
- domain-specific ACH logic remains in downstream workers (for example, `ach-trust-lab`)

## Deployment model

The service remains Go-first and Postgres-backed, with optional dependency wiring for:

- clawmem via `CLAWMEM_BASE_URL`
- remote inference (for ai-precision/Ollama-style hosts) via `INFERENCE_BASE_URL`

This keeps the control plane reusable while still enabling deterministic + non-deterministic coexistence for compliance-oriented development programs.
