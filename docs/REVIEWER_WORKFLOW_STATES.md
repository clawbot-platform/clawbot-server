# REVIEWER_WORKFLOW_STATES.md

## Purpose

This document defines the reviewer workflow states for governed run and cycle evaluation.

The workflow is intended to support:

- deterministic-only results
- llm results
- dual comparison results
- guardrail-enabled and guardrail-deferred outcomes
- explicit reviewer decisions with rationale

This workflow should be implemented primarily in `clawbot-server`, with references preserved in artifacts and optionally in `clawmem`.

---

## Design goals

The reviewer workflow must be:

- explicit
- auditable
- attributable
- resistant to silent overrides
- compatible with deterministic and non-deterministic execution modes

---

## State model

## 1. `pending`
### Meaning
The run or cycle has been created but not yet executed.

### Entry conditions
- run created
- cycle created
- execution not started

### Exit conditions
- move to `running`
- move to `policy_denied`

---

## 2. `policy_denied`
### Meaning
Execution was blocked before runtime due to governance policy.

### Entry conditions
- policy decision point returned deny

### Exit conditions
- none unless a new approved run is created
- may be superseded by explicit approved override in a new review event

### Required metadata
- policy decision id
- deny reason
- actor id
- timestamp

---

## 3. `running`
### Meaning
Execution has started.

### Entry conditions
- run started
- cycle execution in progress

### Exit conditions
- `completed`
- `review_pending`
- `guardrail_deferred`
- `failed_runtime`
- `failed_policy`

---

## 4. `completed`
### Meaning
Execution completed and no review is required for the current governance profile.

### Intended use
Primarily for:
- deterministic baseline runs
- internal non-review-gated operations

### Exit conditions
- optional transition to `review_pending` if manually escalated

---

## 5. `review_pending`
### Meaning
Execution completed, but reviewer approval is required before the result is treated as accepted.

### Typical causes
- dual mode completed
- llm mode completed
- policy requires human signoff
- run is awaiting adjudication

### Exit conditions
- `approved`
- `rejected`
- `overridden`
- `deferred`

---

## 6. `guardrail_deferred`
### Meaning
Primary execution completed, but guardrail evaluation did not complete inline.

### Typical causes
- guardrail timeout
- guardrail runtime unavailable
- policy allows deferred guardrail evaluation

### Exit conditions
- `review_pending`
- `approved`
- `rejected`
- `failed_runtime`

### Required metadata
- guardrail reason
- whether primary output exists
- whether comparison object exists
- whether deferred guardrail follow-up is required

---

## 7. `failed_runtime`
### Meaning
Execution failed due to runtime or infrastructure issues.

### Examples
- inference timeout
- persistence failure
- artifact write failure
- memory fetch failure

### Exit conditions
- new rerun
- reviewer override only if policy allows

---

## 8. `failed_policy`
### Meaning
Execution started but later hit a policy-enforced failure condition.

### Examples
- policy violation discovered during execution
- disallowed endpoint usage detected
- execution ring breach

### Exit conditions
- reviewer override
- rerun under corrected policy

---

## 9. `approved`
### Meaning
A reviewer accepted the run or cycle outcome.

### Required metadata
- reviewer id
- reviewer type
- timestamp
- rationale
- review event id

### Exit conditions
- `overridden` only with explicit subsequent reviewer event

---

## 10. `rejected`
### Meaning
A reviewer explicitly rejected the outcome.

### Required metadata
- reviewer id
- timestamp
- rationale
- recommended next action if applicable

### Exit conditions
- rerun
- `overridden` only with explicit authorized reviewer action

---

## 11. `overridden`
### Meaning
A reviewer with sufficient authority overrode a prior result or review state.

### Typical use
- accept a previously rejected result
- reject a previously approved result
- mark a policy-denied outcome as allowed under an exception
- accept output despite deferred guardrails

### Required metadata
- overriding reviewer id
- authority level
- prior status
- new status / disposition
- rationale
- linked policy exception if applicable

---

## 12. `deferred`
### Meaning
A reviewer intentionally postponed final adjudication.

### Typical use
- awaiting more evidence
- waiting for async guardrail result
- pending domain-owner review

### Required metadata
- reviewer id
- defer reason
- expected follow-up action
- optional due date

---

## Review actions

## `approve`
Use when:
- output is acceptable
- artifacts are complete enough
- comparison or deterministic evidence is sufficient
- policy conditions are satisfied

## `reject`
Use when:
- output is materially flawed
- evidence is insufficient
- comparison reveals unacceptable divergence
- policy conditions are not met

## `override`
Use when:
- explicit exception handling is justified
- higher-authority reviewer needs to change prior disposition
- documented business/governance exception is allowed

## `defer`
Use when:
- more information is required
- guardrails are deferred
- async follow-up is pending

---

## Recommended status transitions

### Core transitions
- `pending` -> `running`
- `pending` -> `policy_denied`
- `running` -> `completed`
- `running` -> `review_pending`
- `running` -> `guardrail_deferred`
- `running` -> `failed_runtime`
- `running` -> `failed_policy`

### Review transitions
- `review_pending` -> `approved`
- `review_pending` -> `rejected`
- `review_pending` -> `overridden`
- `review_pending` -> `deferred`

### Deferred transitions
- `guardrail_deferred` -> `review_pending`
- `guardrail_deferred` -> `approved`
- `guardrail_deferred` -> `rejected`
- `deferred` -> `review_pending`
- `deferred` -> `approved`
- `deferred` -> `rejected`

### Restricted transitions
- `approved` should not silently revert without an explicit override event
- `rejected` should not silently revert without an explicit override event
- `policy_denied` should not become `running` without a new or overridden execution event

---

## Review scope

The workflow should support both:

- **run-level review**
- **cycle-level review**

### Recommended rule
- cycle-level review for week-run details
- run-level review for final acceptance / rejection / export readiness

---

## Metadata required for every reviewer action

- reviewer id
- reviewer type
- timestamp
- prior status
- new status
- rationale
- target run id
- target cycle id if applicable
- policy decision id if relevant
- artifact ids or comparison id if relevant

---

## Reviewer roles

Suggested initial roles:

- `operator`
- `analyst`
- `reviewer`
- `approver`
- `admin`

### Governance recommendation
Only `reviewer`, `approver`, or `admin` should be able to finalize approval states.

Only `approver` or `admin` should be allowed to use `override`.

---

## Artifact and comparison implications

### Approved
Artifacts may be treated as accepted evidence.

### Rejected
Artifacts remain historical evidence but are not accepted output.

### Review pending / deferred
Artifacts are provisional and should be labeled accordingly.

### Override
Artifacts should retain original review history plus override event history.

---

## `clawmem` implications

Reviewer workflow should optionally write scoped memory entries for:
- reviewer_note
- policy_exception
- deferred_guardrail_followup
- accepted_risk

This keeps continuity visible across later cycles.

---

## Recommended API actions

Suggested endpoint families:

- `POST /api/v1/runs/{runID}/review/approve`
- `POST /api/v1/runs/{runID}/review/reject`
- `POST /api/v1/runs/{runID}/review/override`
- `POST /api/v1/runs/{runID}/review/defer`

Optional cycle-level variants:
- `POST /api/v1/runs/{runID}/cycles/{cycleID}/review/...`

---

## Definition of workflow success

The reviewer workflow is acceptable when:

- no meaningful review state change happens silently
- every approval/rejection/override/defer action is attributable
- provisional states are distinguishable from accepted states
- deferred guardrails do not masquerade as completed governance
- run and cycle histories remain auditable
