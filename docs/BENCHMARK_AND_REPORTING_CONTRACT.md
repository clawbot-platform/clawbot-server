# Benchmark & Reporting Contract
## Clawbot Trust Lab — Digital Red Queen for Agentic Commerce Fraud

## 1. Purpose

Clawbot Trust Lab is a **continuous adversarial evaluation program** for agentic commerce fraud.

Its purpose is to:

- prove that **Digital Red Queen (DRQ)** dynamics are real in mixed human + agent commerce,
- continuously discover new detector blind spots,
- preserve gains through replay, regression, and control hardening,
- show incremental robustness improvement over time,
- generate regular benchmark reports that demonstrate progress in a durable and reviewable way.

This effort is **not** a one-time demo or static fraud POC.

It is intended to become a **living benchmark and continuous improvement program**.

---

## 2. Scope

This contract governs how the trust lab is evaluated, reported, and evolved across benchmark rounds.

It defines:

- what a benchmark round is,
- which artifacts must be versioned,
- which metrics must be reported,
- how new attacks are promoted into replay,
- how stable and living benchmark suites coexist,
- when weekly vs bi-weekly reporting is appropriate,
- what success means for this program.

---

## 3. Program Statement

The official program statement is:

> Clawbot Trust Lab is a continuous adversarial fraud-evaluation lab for agentic commerce.
>
> The goal is to demonstrate that Digital Red Queen dynamics are real in this domain and that fraud detection can improve incrementally through repeated cycles of adversarial mutation, replay, detector hardening, and policy updates.
>
> Early progress will be measured bi-weekly. Weekly reporting becomes appropriate only after the evaluation loop is stable and attack generation has enough depth.
>
> Success is not defined by monotonic improvement every round, but by the lab’s ability to discover new blind spots, preserve gains against prior attacks, and improve robustness over time.

---

## 4. Non-Goals

This program does **not** promise:

- monotonic improvement every round,
- linear week-over-week gains,
- permanent advantage of the defender over the adversary,
- production-grade fraud guarantees,
- a frozen benchmark that never changes.

The adversary is expected to win some rounds.
That is part of validating that the benchmark is alive.

---

## 5. Benchmark Model

The lab operates with **two benchmark layers**.

### 5.1 Stable Benchmark Suite

The Stable Benchmark Suite is the frozen regression layer.

It is used to answer:

- Did we preserve prior gains?
- Did we regress on previously solved attacks?
- Did a recent control change break earlier defenses?

Characteristics:

- versioned,
- replayable,
- slow-changing,
- used for regression and release gating.

### 5.2 Living Adversarial Suite

The Living Adversarial Suite is the evolving challenge layer.

It is used to answer:

- Are new evasions emerging?
- Is the benchmark still meaningful?
- Can current controls handle adaptive fraud agents?

Characteristics:

- continuously updated,
- includes newly mutated attacks,
- includes newly discovered blind spots,
- changes more frequently than the stable suite.

### 5.3 Why Both Are Required

If only the stable suite exists, the benchmark becomes stale.

If only the living suite exists, comparisons become noisy and hard to trust.

Therefore:

- the **stable suite** proves we did not regress,
- the **living suite** proves the program is still alive.

---

## 6. Definition of a Benchmark Round

A benchmark round is valid only if all the following are versioned and recorded:

- detector version,
- policy version,
- stable benchmark suite version,
- living adversarial suite version,
- scenario pack version,
- model lane configuration,
- run window,
- output report.

A round should be treated as the minimum comparable unit of progress.

---

## 7. Minimum Round Inputs

Each benchmark round must declare:

- `round_id`
- `detector_version`
- `policy_version`
- `stable_suite_version`
- `living_suite_version`
- `scenario_pack_version`
- `memory_archive_snapshot`
- `model_lane_profile`
- `run_start_ts`
- `run_end_ts`
- `owner`
- `notes`

### 7.1 Model Lane Profile

Each round must explicitly state which inference lanes were used:

- local-only,
- local-first with cloud fallback,
- mixed local + cloud,
- cloud challenge round.

This ensures benchmark claims remain honest and reproducible.

---

## 8. Promotion Rules for New Attacks

A newly observed tactic must be promoted into the replay archive when one or more of the following is true:

- it causes a material fraud miss,
- it bypasses an intended trust or policy control,
- it reveals a new blind spot,
- it materially increases analyst workload,
- it produces a false-positive pattern on legitimate agents,
- it exposes a new class of mandate/provenance weakness.

Once promoted, the attack should be:

- versioned,
- described,
- linked to its lineage if mutated from prior tactics,
- added to the archive,
- considered for stable-suite inclusion after review.

---

## 9. Replay Archive Policy

The replay archive exists to preserve learning across rounds.

### 9.1 Archive Contents

Each replayable attack entry should include:

- attack family,
- tactic name,
- lineage parent if any,
- scenario assumptions,
- required actor roles,
- detector version beaten,
- expected outcome,
- actual outcome,
- payload or scenario reference,
- explanation of why it mattered.

### 9.2 Retention Policy

The following should never be allowed to decay out of the archive:

- attacks that beat the detector,
- attacks that broke policy assumptions,
- high-value adversarial variants,
- gold regression cases.

Low-value raw traces may decay or summarize over time, but promoted replay cases should remain durable.

---

## 10. Reporting Cadence

### 10.1 Default Starting Cadence

The starting cadence is:

- **bi-weekly**

This is the default until the lab is stable enough for weekly reporting.

### 10.2 Criteria for Weekly Reporting

Weekly reporting becomes appropriate only when all the following are true:

- scenario generation is stable,
- replay promotion is mostly automated,
- metrics are trusted and not excessively noisy,
- benchmark execution is repeatable,
- reporting generation is operationally lightweight,
- living adversary depth is sufficient to justify weekly signal.

### 10.3 Allowed Exceptions

If adversary evolution is temporarily weak or benchmark execution quality is low, cadence should remain bi-weekly even after prior weekly reporting.

Trustworthiness of reporting is more important than frequency.

---

## 11. Required Metric Categories

Every report must include metrics from all five categories below.

### 11.1 Fraud Effectiveness

- precision
- recall
- false positive rate
- false negative rate
- estimated fraud loss caught
- estimated fraud loss missed

### 11.2 Agentic Trust Effectiveness

- mandate violation detection rate
- provenance-gap detection rate
- legitimate-agent false positive rate
- step-up precision
- freeze-agent precision
- policy breach catch rate

### 11.3 Red Queen Robustness

- new successful evasions discovered this round
- replay pass rate
- performance drop under mutated attacks
- rounds to recover after a new evasion
- robustness trend across recent rounds

### 11.4 Operational Metrics

- review queue size
- analyst time per case
- explanation usefulness
- compute cost per round
- local vs cloud model usage

### 11.5 Program Health

- number of attacks promoted into replay
- archive growth
- detector changes shipped
- policies updated
- unresolved blind spots carried forward

---

## 12. Standard Report Structure

Each benchmark report should follow this structure.

### 12.1 Executive Summary

Concise summary of:

- whether this round improved robustness,
- whether regressions occurred,
- whether new major evasions were found.

### 12.2 What Changed Since Last Round

List:

- detector updates,
- policy updates,
- model lane changes,
- scenario pack changes,
- replay archive changes.

### 12.3 Stable Suite Results

Explain:

- pass/fail movement,
- regressions,
- retained gains,
- release confidence.

### 12.4 Living Suite Results

Explain:

- new evasions,
- adversary wins,
- defender weaknesses,
- notable changes in false positives or analyst burden.

### 12.5 New Evasions Discovered

For each meaningful evasion:

- short name,
- affected control,
- why it succeeded,
- whether it is promoted to replay,
- recommended response.

### 12.6 Operational Summary

Include:

- analyst queue movement,
- review burden,
- compute cost,
- local/cloud lane usage.

### 12.7 Recommendations

List:

- detector updates,
- policy changes,
- reporting cadence changes if needed,
- whether the round is strong enough to promote new stable cases.

---

## 13. Success Criteria

Success is not defined as “the defender always wins.”

Success means the program can reliably do the following:

- discover new blind spots,
- preserve gains against prior attacks,
- improve robustness over longer windows,
- avoid unacceptable collapse in legitimate-agent handling,
- produce trustworthy reports at a sustainable cadence.

### 13.1 Short-Term Success

- DRQ behavior is demonstrated,
- static controls are shown to degrade,
- replay improves future performance,
- meaningful reports can be generated.

### 13.2 Mid-Term Success

- stable suite pass rate trends upward,
- recovery after new evasions becomes faster,
- false positives on legitimate agents remain controlled,
- the archive becomes a durable benchmark asset.

### 13.3 Long-Term Success

- the trust lab becomes a reusable internal benchmark,
- new fraud and trust controls can be validated against it,
- the benchmark remains alive without losing comparability.

---

## 14. Failure Signals

The program should be treated as unhealthy if any of the following persist:

- no meaningful new evasions are discovered for too long,
- the stable suite is not maintained,
- reports become manual or untrustworthy,
- metrics are selectively presented,
- attack promotion rules are inconsistently applied,
- adversary evolution becomes superficial,
- local/cloud usage is not disclosed clearly.

---

## 15. Local vs Cloud Model Policy

The benchmark should be run with a **local-first, hybrid-capable** model strategy.

### 15.1 Default Policy

Default mode:

- local-first inference,
- cloud fallback disabled unless explicitly needed,
- all model traffic routed through the model gateway.

### 15.2 Allowed Cloud Usage

Cloud usage is allowed for:

- challenge rounds,
- sanity-check judging,
- occasional stronger adversarial rounds,
- fallback when local capacity is insufficient.

### 15.3 Reporting Requirement

Every report must disclose:

- whether the round was local-only or mixed,
- whether cloud challengers were used,
- whether any important result depended on cloud-only capability.

This keeps the value proposition honest and protects reproducibility.

---

## 16. Roles and Ownership

### 16.1 `clawbot-server`

Owns:

- round metadata,
- scheduling and run APIs,
- report templates and report APIs,
- audit trail and release/report history.

### 16.2 `clawbot-trust-lab`

Owns:

- benchmark execution,
- scenario packs,
- living adversarial suites,
- stable benchmark execution,
- detector evaluation,
- Red Queen mutation and replay workflows.

### 16.3 `clawmem`

Owns:

- replay archive storage,
- promoted attack memory,
- long-term benchmark memory,
- durable retrieval of prior attack and defense cases.

---

## 17. Versioning Policy

The following must be versioned independently:

- detector
- policies
- stable suite
- living suite
- scenario pack
- memory archive snapshot
- report schema

A round should never rely on ambiguous or ad hoc versions.

---

## 18. Readiness Gates

### 18.1 Gate for “Round is Valid”

A round is valid only if:

- inputs are versioned,
- outputs are captured,
- metrics are complete,
- local/cloud usage is disclosed,
- notable failures are documented.

### 18.2 Gate for “Promote Attack to Replay”

Promotion requires:

- documented impact,
- reproducibility,
- concise explanation,
- archive metadata,
- clear linkage to affected control or blind spot.

### 18.3 Gate for “Ready for Weekly Reporting”

Weekly reporting requires:

- 3 consecutive clean rounds,
- repeatable execution,
- stable archive operations,
- trustworthy metrics,
- low manual reporting burden.

---

## 19. Review Process

Each report should be reviewed for:

- metric completeness,
- honesty of interpretation,
- correctness of version references,
- clarity of local/cloud usage,
- whether new blind spots were handled properly.

If a report is incomplete or misleading, the round should be marked incomplete rather than published as if healthy.

---

## 20. Contract Change Policy

This contract may evolve, but changes must be documented.

Any change to:

- success criteria,
- required metrics,
- cadence rules,
- promotion rules,
- stable vs living benchmark policy

must be recorded in version history.

---

## 21. Final Principle

This benchmark exists to remain alive.

If the lab stops discovering meaningful weaknesses, stops preserving prior gains, or stops producing honest comparable reports, then it is no longer functioning as a Digital Red Queen program.

The benchmark should therefore be treated as a **continuous adversarial trust program**, not a one-time fraud demo.
