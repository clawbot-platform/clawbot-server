# Report Template
## Clawbot Trust Lab — Benchmark Round Report

**Report ID:** `RPT-YYYY-NN`  
**Round ID:** `ROUND-YYYY-NN`  
**Report Date:** `YYYY-MM-DD`  
**Prepared By:** `[name or team]`  
**Cadence:** `Weekly | Bi-weekly`  
**Report Status:** `Draft | Final`

---

## 1. Executive Summary

### Overall Assessment
`[Brief summary of how this round went.]`

### Key Takeaways
- `[Takeaway 1]`
- `[Takeaway 2]`
- `[Takeaway 3]`

### Headline Outcome
Choose one:
- `Robustness improved`
- `Mixed results`
- `Regression observed`
- `Adversary win / new blind spot discovered`

### Recommendation
Choose one:
- `Continue current cadence`
- `Promote to weekly reporting`
- `Remain bi-weekly`
- `Pause and stabilize benchmark pipeline`
- `Immediate control update required`

---

## 2. Round Metadata

| Field | Value |
|---|---|
| Round ID | `ROUND-YYYY-NN` |
| Detector Version | `detector-vX.Y.Z` |
| Policy Version | `policy-vX.Y.Z` |
| Stable Suite Version | `stable-suite-vX.Y.Z` |
| Living Suite Version | `living-suite-vX.Y.Z` |
| Scenario Pack Version | `scenario-pack-vX.Y.Z` |
| Memory Archive Snapshot | `archive-snapshot-vX.Y.Z` |
| Model Lane Profile | `local-only | local-first-with-fallback | mixed | cloud-challenge` |
| Run Start | `timestamp` |
| Run End | `timestamp` |
| Owner | `[team or person]` |
| Git Commit / Release Ref | `[ref]` |

### Local vs Cloud Usage Disclosure
- **Primary lane:** `[local-only / mixed / cloud-challenge]`
- **Local models used:** `[list]`
- **Cloud models used:** `[list or none]`
- **Did any material result depend on cloud-only capability?** `Yes | No`
- **Notes:** `[short note]`

---

## 3. What Changed Since Last Round

### Detector Changes
- `[change 1]`
- `[change 2]`

### Policy Changes
- `[change 1]`
- `[change 2]`

### Scenario Changes
- `[change 1]`
- `[change 2]`

### Living Adversary Changes
- `[new attack family / mutation / stronger tactic]`

### Replay Archive Changes
- `[promoted cases]`
- `[retired or summarized low-value artifacts if any]`

### Infrastructure / Runtime Changes
- `[ZeroClaw / OmniRoute / serving / deployment changes]`

---

## 4. Benchmark Execution Summary

### Run Size
| Metric | Value |
|---|---|
| Total episodes | `[n]` |
| Total transactions | `[n]` |
| Human-only transactions | `[n]` |
| Agent-assisted transactions | `[n]` |
| Fully delegated transactions | `[n]` |
| Fraud / adversarial transactions | `[n]` |
| Review-triggered cases | `[n]` |

### Actor Mix
- **Buyer agents:** `[n]`
- **Merchant agents:** `[n]`
- **Fraud agents:** `[n]`
- **Compliance agents:** `[n]`
- **Reviewer agents:** `[n]`

### Notes on Benchmark Validity
- `[Were all required inputs versioned?]`
- `[Any execution anomalies?]`
- `[Any caveats affecting trustworthiness of this round?]`

---

## 5. Stable Benchmark Suite Results

### Summary
`[Describe whether prior gains held and whether regressions were observed.]`

### Stable Suite Metrics

| Metric | Previous Round | Current Round | Delta |
|---|---:|---:|---:|
| Precision | `[ ]` | `[ ]` | `[ ]` |
| Recall | `[ ]` | `[ ]` | `[ ]` |
| False Positive Rate | `[ ]` | `[ ]` | `[ ]` |
| False Negative Rate | `[ ]` | `[ ]` | `[ ]` |
| Replay Pass Rate | `[ ]` | `[ ]` | `[ ]` |
| Estimated Fraud Loss Missed | `[ ]` | `[ ]` | `[ ]` |

### Regression Findings
- `[None]` or
- `[Describe any regression and affected attack family/control]`

### Confidence Assessment
Choose one:
- `High confidence`
- `Medium confidence`
- `Low confidence`

Reason:
`[Short explanation]`

---

## 6. Living Adversarial Suite Results

### Summary
`[Describe how the detector performed against evolving attacks.]`

### Living Suite Metrics

| Metric | Previous Round | Current Round | Delta |
|---|---:|---:|---:|
| Precision | `[ ]` | `[ ]` | `[ ]` |
| Recall | `[ ]` | `[ ]` | `[ ]` |
| False Positive Rate | `[ ]` | `[ ]` | `[ ]` |
| False Negative Rate | `[ ]` | `[ ]` | `[ ]` |
| New Successful Evasions | `[ ]` | `[ ]` | `[ ]` |
| Performance Drop Under Mutation | `[ ]` | `[ ]` | `[ ]` |

### Notable Adversary Wins
- `[attack name] — [why it succeeded]`
- `[attack name] — [why it succeeded]`

### Notable Defender Successes
- `[attack family now blocked or reduced]`
- `[control improvement outcome]`

---

## 7. Fraud Effectiveness Metrics

| Metric | Previous Round | Current Round | Delta | Notes |
|---|---:|---:|---:|---|
| Precision | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Recall | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| False Positive Rate | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| False Negative Rate | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Estimated Fraud Loss Caught | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Estimated Fraud Loss Missed | `[ ]` | `[ ]` | `[ ]` | `[ ]` |

### Interpretation
`[Explain the most important changes.]`

---

## 8. Agentic Trust Metrics

| Metric | Previous Round | Current Round | Delta | Notes |
|---|---:|---:|---:|---|
| Mandate Violation Detection Rate | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Provenance-Gap Detection Rate | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Legitimate-Agent False Positive Rate | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Step-Up Precision | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Freeze-Agent Precision | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Policy Breach Catch Rate | `[ ]` | `[ ]` | `[ ]` | `[ ]` |

### Interpretation
`[Explain where trust-native signals helped or failed.]`

---

## 9. Red Queen Robustness Metrics

| Metric | Previous Round | Current Round | Delta | Notes |
|---|---:|---:|---:|---|
| New Successful Evasions | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Replay Pass Rate | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Performance Drop Under Mutated Attacks | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Rounds to Recover from New Evasion | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Robustness Trend Score | `[ ]` | `[ ]` | `[ ]` | `[ ]` |

### Interpretation
`[Summarize whether the defender is becoming more robust over time.]`

---

## 10. Operations Metrics

| Metric | Previous Round | Current Round | Delta | Notes |
|---|---:|---:|---:|---|
| Review Queue Size | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Analyst Time per Case | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Explanation Usefulness | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Compute Cost per Round | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Local Model Usage % | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Cloud Model Usage % | `[ ]` | `[ ]` | `[ ]` | `[ ]` |

### Interpretation
`[Explain operational cost and review burden movement.]`

---

## 11. Program Health

| Metric | Previous Round | Current Round | Delta | Notes |
|---|---:|---:|---:|---|
| Attacks Promoted to Replay | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Archive Growth | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Detector Changes Shipped | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Policy Updates Shipped | `[ ]` | `[ ]` | `[ ]` | `[ ]` |
| Unresolved Blind Spots | `[ ]` | `[ ]` | `[ ]` | `[ ]` |

### Interpretation
`[Explain whether the benchmark program itself is healthy.]`

---

## 12. New Evasions Discovered This Round

### Evasion 1
- **Name:** `[short name]`
- **Attack family:** `[family]`
- **Affected control:** `[control/policy/detector component]`
- **Scenario conditions:** `[short note]`
- **Why it worked:** `[short explanation]`
- **Impact level:** `Low | Medium | High | Critical`
- **Promote to replay archive:** `Yes | No`
- **Recommended next action:** `[short action]`

### Evasion 2
- **Name:** `[short name]`
- **Attack family:** `[family]`
- **Affected control:** `[control/policy/detector component]`
- **Scenario conditions:** `[short note]`
- **Why it worked:** `[short explanation]`
- **Impact level:** `Low | Medium | High | Critical`
- **Promote to replay archive:** `Yes | No`
- **Recommended next action:** `[short action]`

_Add or remove sections as needed._

---

## 13. Regressions Caught This Round

### Regression 1
- **Name:** `[short name]`
- **Previously passing control:** `[control]`
- **Current failure mode:** `[short description]`
- **Suspected cause:** `[short description]`
- **Severity:** `Low | Medium | High | Critical`
- **Immediate mitigation required:** `Yes | No`

_Add more as needed._

---

## 14. Recommended Changes Before Next Round

### Detector Changes
- `[recommended change]`
- `[recommended change]`

### Policy Changes
- `[recommended change]`
- `[recommended change]`

### Scenario Changes
- `[recommended change]`
- `[recommended change]`

### Memory / Replay Archive Changes
- `[promote or pin cases]`
- `[archive hygiene / summarize low-value traces]`

### Reporting / Program Changes
- `[cadence adjustment if needed]`
- `[metric quality or automation changes]`

---

## 15. Round Decision

Choose one:

- `Proceed to next round with current cadence`
- `Proceed, but remain bi-weekly`
- `Promote to weekly cadence`
- `Hold next round until regressions are fixed`
- `Repeat round due to execution quality concerns`

### Rationale
`[Short explanation]`

---

## 16. Open Issues

- `[issue 1]`
- `[issue 2]`
- `[issue 3]`

---

## 17. Appendix A — Version References

| Component | Version / Ref |
|---|---|
| Detector | `[ ]` |
| Policies | `[ ]` |
| Stable Suite | `[ ]` |
| Living Suite | `[ ]` |
| Scenario Pack | `[ ]` |
| Archive Snapshot | `[ ]` |
| ZeroClaw | `[ ]` |
| OmniRoute | `[ ]` |
| clawmem | `[ ]` |
| clawbot-server | `[ ]` |
| clawbot-trust-lab | `[ ]` |

---

## 18. Appendix B — Benchmark Notes

### Caveats
`[List anything that may affect interpretation.]`

### Data Quality Notes
`[List benchmark/data limitations.]`

### Reporting Integrity Statement
Choose one:
- `This report is complete and trusted`
- `This report is usable with caveats`
- `This report is incomplete and should not be used for trend claims`

### Notes
`[Short explanation]`
