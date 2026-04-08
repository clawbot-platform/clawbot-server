# CLAWBOT_SERVER_GOVERNANCE_BACKLOG.md

## Purpose

This backlog translates governance-oriented design patterns into concrete implementation work for `clawbot-server`.

The focus is to evolve `clawbot-server` from a capable control plane into a governed execution plane for:

- `ach-trust-lab`
- future domain workers
- deterministic and non-deterministic DRQ execution
- reviewable, auditable, policy-controlled agent workflows

This backlog assumes the current platform direction:

- deterministic replay remains authoritative
- Granite provides reasoning
- Guardian provides risk / compliance gating
- `clawbot-server` owns execution control
- `clawmem` owns scoped continuity
- domain repos should not re-implement orchestration locally

---

## Governance goals

`clawbot-server` should become responsible for:

- deterministic policy checks before execution
- execution-ring enforcement
- reviewer-driven approvals and overrides
- auditability of run and cycle decisions
- artifact provenance
- runtime fallback handling when model or guardrail phases degrade
- policy-as-data rather than policy hidden inside code branches

---

## Priority 0 — immediate backlog

These are the highest-value items because they convert the current runtime slice into a governed control plane.

### 0.1 Add a policy decision point before execution

Add a policy evaluation step before these actions:

- create run
- create cycle
- start run
- execute cycle
- attach artifact
- finalize review

Policy input should include:

- run type
- execution mode
- execution ring
- model profile
- guardrail enabled / disabled
- repo namespace
- run namespace
- requested_by / actor
- environment
- external endpoint targets
- prompt pack version
- rule pack version

Policy output should include:

- decision: allow / deny
- policy bundle id
- policy bundle version
- reason code
- conditions applied
- fallback instruction if applicable

#### Deliverables
- policy evaluation service
- policy decision schema
- policy decision persistence
- unit tests
- API docs

---

### 0.2 Add execution rings

Add an `execution_ring` field to runs and cycles.

Initial ring model:

- `ring_0`: metadata only, no external effects
- `ring_1`: deterministic replay + internal persistence only
- `ring_2`: local inference allowed
- `ring_3`: external tools / network calls allowed by explicit policy

#### Rules
- `deterministic` should be allowed in `ring_1`
- `llm` should require at least `ring_2`
- `dual` should require at least `ring_2`
- external tool calls should require `ring_3`

#### Deliverables
- schema changes
- handler support
- policy validation
- docs update
- tests

---

### 0.3 Add reviewer action endpoints

Promote review from a status string into a first-class governance workflow.

Add endpoints like:

- `POST /api/v1/runs/{runID}/review/approve`
- `POST /api/v1/runs/{runID}/review/reject`
- `POST /api/v1/runs/{runID}/review/override`
- `POST /api/v1/runs/{runID}/review/defer`

Each action should record:

- reviewer id
- reviewer type
- timestamp
- prior status
- new status
- rationale
- optional linked policy decision id

#### Deliverables
- review action schema
- endpoints
- persistence
- audit events
- tests

---

### 0.4 Add guardrail fallback policy

Guardrail execution should not blindly fail entire runs.

Introduce configurable fallback behavior:

- `inline_guardrail_required`
- `inline_guardrail_optional`
- `async_guardrail_allowed`
- `guardrail_disabled_for_validation`

If inline Guardian exceeds latency budget or fails:

- persist status such as `review_pending`
- attach `guardrail_deferred` marker
- persist primary artifacts
- preserve comparison object if available
- do not silently discard evidence

#### Deliverables
- fallback policy logic
- new status / reason codes
- docs
- tests

---

### 0.5 Add append-only audit events with hash chaining

Create a tamper-evident audit chain for control-plane decisions.

Each audit event should include:

- event id
- prior event hash
- current event hash
- actor id
- actor type
- action type
- target run id
- target cycle id
- target artifact id
- policy decision id
- timestamp
- event payload summary

#### Deliverables
- audit event schema
- hash-chain logic
- persistence
- retrieval endpoints if useful
- tests

---

## Priority 1 — next backlog

These items deepen governance and traceability.

### 1.1 Add actor and worker identity fields

Enrich run, cycle, artifact, and comparison objects with:

- actor_id
- actor_type
- worker_id
- host_id
- policy_bundle_version
- model_profile_id
- inference_provider
- execution_ring

#### Goal
Every important action should be attributable.

---

### 1.2 Add artifact provenance enrichment

For every artifact, persist:

- produced_by_run_id
- produced_by_cycle_id
- produced_by_worker_id
- policy_decision_id
- model_profile_id
- execution_mode
- review_status_at_creation
- artifact_hash

#### Goal
Artifacts should be auditable, reviewable, and export-safe.

---

### 1.3 Add external action allowlists

Create allowlist support for:

- inference endpoints
- memory endpoints
- artifact sinks
- other outbound HTTP targets
- future tool/action targets

Tie allowlists to:

- execution ring
- policy bundle
- environment

---

### 1.4 Add governance-oriented run statuses

Expand run/cycle statuses to support operational governance, for example:

- `pending`
- `running`
- `completed`
- `review_pending`
- `guardrail_deferred`
- `policy_denied`
- `failed_runtime`
- `failed_policy`
- `approved`
- `rejected`
- `overridden`

---

### 1.5 Add per-phase runtime budgets

Define explicit budgets for:

- deterministic phase
- primary model phase
- guardrail phase
- helper phase
- memory fetch
- artifact registration

#### Goal
Operational governance should include SRE-style budgets and circuit-breaking.

---

## Priority 2 — later backlog

These are important, but can follow once the core governance surface is in place.

### 2.1 Move policy rules into versioned bundles

Introduce policy bundles as versioned config artifacts.

Suggested structure:

- execution mode rules
- ring rules
- model allowlists
- guardrail requirements
- reviewer requirements
- retention expectations
- export rules

Start with YAML if needed.

---

### 2.2 Add sandbox metadata hooks

Prepare for future stronger isolation by adding metadata fields like:

- runtime_class
- container_image
- network_policy_profile
- isolation_profile
- capability_profile

This enables future hardening without changing run semantics later.

---

### 2.3 Add governance dashboard / read model

Expose a small governance-focused read surface showing:

- runs by execution mode
- runs by review status
- policy denials
- reviewer overrides
- guardrail deferred count
- average phase timings
- artifacts by type

---

### 2.4 Add signed or hash-linked artifact manifests

Move beyond simple metadata to:

- run manifest hashes
- cycle manifest hashes
- artifact list hashes
- export bundle hashes

---

## Suggested implementation order

### Phase A
- policy decision point
- execution rings
- reviewer endpoints
- guardrail fallback rules

### Phase B
- hash-chained audit events
- identity/provenance enrichment
- runtime budgets
- allowlists

### Phase C
- policy bundles
- dashboard
- artifact manifest hashing
- sandbox metadata hooks

---

## Recommended ADRs

Add or expand ADRs for:

1. deterministic replay remains authoritative
2. governance decisions happen before action execution
3. execution rings model
4. guardrail fallback and deferred review
5. hash-chained audit model

---

## Recommended threat model topics

Document threats including:

- unauthorized run creation
- unauthorized mode escalation (`deterministic` -> `dual`)
- unapproved model profile usage
- external endpoint misuse
- prompt pack tampering
- review bypass
- artifact tampering
- memory poisoning via bad notes or context
- degraded guardrail path causing silent unsafe approvals

---

## Definition of done

`clawbot-server` governance backlog is meaningfully complete when:

- execution is policy-gated
- execution rings are enforced
- reviewer actions are first-class
- audit entries are append-only and hash-linked
- fallback policies exist for slow/failed guardrails
- artifacts carry provenance
- status transitions are reviewable and explainable

---

## Open questions

- Should reviewer actions operate at run level, cycle level, or both?
- Should policy bundles be YAML first or go straight to a policy engine?
- Should deferred guardrails block export or just mark export as provisional?
- How much provenance belongs in artifact metadata vs audit log only?
