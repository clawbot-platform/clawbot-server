# EXECUTION_RINGS.md

## Purpose

This document defines the execution ring model for governed agent execution.

Execution rings are capability boundaries enforced by `clawbot-server` before or during execution.

They are intended to:

- reduce accidental privilege escalation
- make run capabilities explicit
- support policy-driven execution control
- separate safe deterministic operations from higher-risk agent actions

---

## Design goals

The execution ring model should be:

- simple enough to enforce immediately
- strict enough to prevent mode escalation
- extensible for future tool and sandbox integration

---

## Ring definitions

## `ring_0` — metadata only

### Purpose
Used for control-plane actions that do not perform meaningful execution.

### Allowed actions
- create run metadata
- create cycle metadata
- register policy-denied events
- inspect configuration and status
- attach non-executive notes

### Not allowed
- deterministic replay
- inference
- external HTTP calls
- memory snapshot generation that implies execution output
- artifact generation representing completed computation

### Typical use
- drafts
- denied requests
- metadata preparation

---

## `ring_1` — deterministic internal execution

### Purpose
Used for deterministic replay and internal-only execution with no model dependency.

### Allowed actions
- deterministic replay
- internal metric computation
- internal artifact registration
- internal memory writes
- run/cycle state transitions
- comparison object partial creation where deterministic only

### Not allowed
- local inference
- external tool calls
- arbitrary outbound network calls

### Typical use
- replay baselines
- regression testing
- authoritative scoring path

---

## `ring_2` — local inference allowed

### Purpose
Used for execution that can call approved local inference providers.

### Allowed actions
- all `ring_1` actions
- local Ollama inference
- Granite primary reasoning
- Guardian guardrail calls
- helper-model calls
- dual mode execution
- comparison generation

### Not allowed
- arbitrary external APIs
- external SaaS tools unless explicitly allowed by higher ring/policy

### Typical use
- llm mode
- dual mode
- governed local AI execution on homelab

---

## `ring_3` — controlled external action ring

### Purpose
Used for execution that can call external services or tools under explicit policy.

### Allowed actions
- all `ring_2` actions
- approved outbound HTTP calls
- approved remote tools
- approved external artifact sinks
- future managed gateways or SaaS integrations

### Not allowed
- unapproved endpoints
- unrestricted tool execution
- silent privilege escalation

### Typical use
- future enterprise connectors
- internet-enabled workflows
- controlled external retrieval or governance services

---

## Mode-to-ring mapping

### `deterministic`
Minimum required ring:
- `ring_1`

### `llm`
Minimum required ring:
- `ring_2`

### `dual`
Minimum required ring:
- `ring_2`

### future tool/external mode
Minimum required ring:
- `ring_3`

---

## Run-type guidance

### `replay_run`
Typical ring:
- `ring_1`

### `agent_run`
Typical ring:
- `ring_2`

### `week_run`
Typical ring:
- `ring_1` for deterministic-only cycles
- `ring_2` for llm / dual cycles

---

## Enforcement rules

### Rule 1
Execution ring must be evaluated by policy before run creation or start.

### Rule 2
A run may not escalate to a higher ring during execution without an explicit approved override.

### Rule 3
The ring must be persisted on the run and cycle records.

### Rule 4
Artifacts should capture which ring produced them.

### Rule 5
Reviewer overrides that change effective ring must be auditable.

---

## Policy expectations by ring

## `ring_0`
- metadata only
- no execution artifacts
- no model usage
- no memory-derived operational context

## `ring_1`
- deterministic artifacts allowed
- memory writes allowed
- memory reads allowed
- no model usage
- no external endpoints

## `ring_2`
- approved local inference endpoints only
- model profile allowlist required
- guardrail behavior governed by policy
- dual comparison artifacts allowed

## `ring_3`
- endpoint allowlist required
- stronger provenance and audit expectations
- reviewer approval may be mandatory depending on policy bundle

---

## Recommended metadata additions

Runs and cycles should store:

- `execution_ring`
- `policy_bundle_id`
- `policy_bundle_version`
- `actor_id`
- `worker_id`
- `host_id`

Artifacts should store:
- `execution_ring`
- `model_profile_id`
- `policy_decision_id`

---

## Interaction with sandboxing

Execution rings are a governance abstraction.
They are not a full sandbox.

Future implementation may map rings onto:
- container isolation profiles
- network policies
- capability profiles
- runtime classes

For now, the ring is a policy-enforced capability boundary.

---

## Example policies

### Example 1
`week_run` + `deterministic` + `ring_1`:
- allowed

### Example 2
`week_run` + `dual` + `ring_1`:
- denied

### Example 3
`agent_run` + `llm` + `ring_2` + approved local model profile:
- allowed

### Example 4
`agent_run` + external SaaS inference + `ring_2`:
- denied

### Example 5
`dual` + `ring_2` + local guardrails deferred:
- allowed if policy bundle permits deferred guardrails

---

## Definition of success

Execution rings are correctly implemented when:

- mode/ring mismatches are denied by policy
- ring escalation is not silent
- artifacts reveal which ring produced them
- review workflows can reason about execution privilege
- future external integrations have a clear security boundary
