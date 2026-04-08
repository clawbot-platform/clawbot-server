# GUARDRAIL_FALLBACK_RULES.md

## Purpose

This document defines fallback rules for guardrail execution in `clawbot-server`.

The goal is to make guardrail behavior:

- explicit
- policy-driven
- auditable
- operationally resilient

Guardrail failures or slowdowns must not silently appear as successful guarded execution.

---

## Background

The current platform direction supports:

- deterministic replay
- Granite primary reasoning
- optional Granite Guardian guardrails
- dual-mode comparison flows

Runtime validation showed that:

- primary Granite execution is fast enough for inline use
- Guardian can be made fast enough when using compact payloads and non-thinking mode
- guardrail timeouts must still be handled as governed operational events

---

## Design goals

Guardrail fallback must:

1. avoid silent safety degradation
2. preserve useful artifacts where safe
3. support reviewer-driven adjudication
4. distinguish between:
   - guardrail not required
   - guardrail required and passed
   - guardrail required but deferred
   - guardrail required and failed
5. support deterministic and dual evidence preservation

---

## Guardrail policy modes

## `inline_guardrail_required`
### Meaning
Guardrail must complete before the run/cycle can be considered normally complete.

### If guardrail succeeds
- proceed normally
- persist guardrail artifact or summary

### If guardrail fails or times out
- do not mark result as fully complete
- set status to `guardrail_deferred` or `review_pending`
- preserve primary artifacts
- record guardrail failure reason

---

## `inline_guardrail_optional`
### Meaning
Guardrail is preferred but not required for initial result persistence.

### If guardrail succeeds
- persist guardrail output
- proceed normally

### If guardrail fails or times out
- preserve primary artifacts
- mark run/cycle as `review_pending`
- set `guardrail_present=false`
- attach reason code

---

## `async_guardrail_allowed`
### Meaning
Primary output may complete first; guardrail may run later.

### If primary succeeds and guardrail is deferred
- persist primary artifacts
- mark status `guardrail_deferred`
- create follow-up task or reviewer action requirement
- prevent silent promotion to approved state

---

## `guardrail_disabled_for_validation`
### Meaning
Guardrails are intentionally disabled for a controlled validation scenario.

### Required behavior
- persist explicit marker that guardrails were disabled
- do not imply guarded execution
- comparison and artifacts should reflect no guardrail summary

---

## Fallback outcomes

## 1. `guardrail_passed`
### Meaning
Guardrail completed and did not flag the output.

### Result
- guardrail summary present
- run/cycle can proceed to normal review state or completion depending on policy

---

## 2. `guardrail_flagged`
### Meaning
Guardrail completed and flagged the output.

### Result
- run/cycle enters `review_pending` or `rejected` depending on policy
- reason recorded
- reviewer attention required

---

## 3. `guardrail_timeout`
### Meaning
Guardrail did not complete within budget.

### Result
- primary artifacts preserved
- run/cycle status should become:
  - `guardrail_deferred`, or
  - `review_pending`
- explicit reason recorded

---

## 4. `guardrail_unavailable`
### Meaning
Guardrail endpoint or runtime was unavailable.

### Result
- same as timeout unless stricter policy says fail
- preserve primary evidence
- make degradation explicit

---

## 5. `guardrail_disabled`
### Meaning
Guardrail was intentionally not used for this run/cycle.

### Result
- explicit policy marker
- no guardrail summary expected
- reviewer should know this was not governed by inline guardrails

---

## Required status behavior

### Allowed statuses after guardrail issues
- `review_pending`
- `guardrail_deferred`
- `failed_runtime`
- `rejected`

### Not allowed
- silently returning `completed` or equivalent guarded success when guardrail did not actually run

---

## Required metadata

Whenever guardrail fallback occurs, persist:

- fallback mode
- reason code
- timeout or runtime error summary
- whether primary output exists
- whether deterministic output exists
- whether comparison exists
- whether reviewer action is required
- whether async follow-up is pending

---

## Artifact behavior

### Primary artifacts
Should still be preserved when safe and policy allows:
- deterministic output
- llm output
- comparison output

### Guardrail artifact
Should reflect one of:
- present and successful
- present and flagged
- absent because deferred
- absent because disabled
- absent because unavailable

---

## Comparison behavior

In `dual` mode:

### If guardrail succeeds
Comparison should include:
- deterministic summary
- llm summary
- guardrail summary
- `guardrail_present=true`

### If guardrail does not succeed
Comparison should still exist if deterministic and llm completed, but should include:
- empty or absent guardrail summary
- `guardrail_present=false`
- explicit status / delta marker

---

## Reviewer workflow implications

### Guardrail deferred
Must require review or follow-up.

### Guardrail disabled
Must be visible to reviewer.

### Guardrail flagged
Should block silent approval.

### Guardrail timeout with primary artifacts present
Should not discard evidence; should route to review state.

---

## Recommended reason codes

Suggested structured reason codes:

- `guardrail_timeout`
- `guardrail_unavailable`
- `guardrail_disabled_policy`
- `guardrail_disabled_validation`
- `guardrail_flagged_output`
- `guardrail_payload_invalid`
- `guardrail_parse_failure`

---

## Recommended runtime budgets

Initial practical local-validation guidance:

- `INFERENCE_TIMEOUT=120s`
- `GUARDRAIL_TIMEOUT=30s`
- `HELPER_TIMEOUT=30s`

Guardrail payloads should remain compact.
Guardrail requests should use non-thinking mode where supported.

---

## Operational rules

### Rule 1
Guardrail fallback must be logged as an audit event.

### Rule 2
Fallback outcome must be visible in run/cycle state.

### Rule 3
Artifacts must remain reviewable even when guardrails fail, unless policy forbids artifact persistence.

### Rule 4
A reviewer or approver should be able to see whether guardrails:
- passed
- flagged
- timed out
- were disabled

### Rule 5
Guardrail fallback must not silently downgrade governance guarantees.

---

## Recommended policy decisions

### Allow
- deterministic run with no guardrail requirement
- llm or dual with deferred guardrail in validation environment if policy allows

### Deny
- promote to approved without review when guardrail was required but absent
- mark run as fully guarded when guardrail was disabled or timed out

---

## Definition of success

Guardrail fallback behavior is acceptable when:

- failure modes are explicit
- primary evidence is preserved where appropriate
- review workflow is triggered when needed
- no silent governance downgrade occurs
- operators can distinguish guardrail-disabled from guardrail-passed outcomes
