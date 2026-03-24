# Executive Report Template
## Clawbot Trust Lab — Benchmark Round Summary

**Report ID:** `RPT-YYYY-NN`  
**Round ID:** `ROUND-YYYY-NN`  
**Report Date:** `YYYY-MM-DD`  
**Cadence:** `Weekly | Bi-weekly`  
**Prepared By:** `[name or team]`  
**Audience:** `Leadership | Stakeholders | Engineering + Product`

---

## 1. Executive Summary

**Overall outcome:**  
`[One short paragraph summarizing whether robustness improved, remained mixed, or regressed.]`

**Headline decision:**  
Choose one:
- `Robustness improved`
- `Mixed results`
- `Regression observed`
- `New adversary blind spot discovered`

---

## 2. What Changed This Round

- `[detector/control change]`
- `[policy or trust-rule change]`
- `[new adversary mutation or scenario change]`
- `[important infrastructure/model-lane change]`

---

## 3. Key Results

| Area                     | Previous Round | Current Round | Direction |
|--------------------------|---------------:|--------------:|-----------|
| Fraud Precision          |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |
| Fraud Recall             |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |
| False Positive Rate      |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |
| Replay Pass Rate         |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |
| New Successful Evasions  |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |
| Legitimate-Agent FP Rate |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |

**Result summary:**  
`[2–4 sentences explaining the most important movement.]`

---

## 4. Stable vs Living Benchmark View

### Stable Suite
**Status:** `Improved | Held | Regressed`

`[One short paragraph explaining whether previously solved attacks stayed solved.]`

### Living Adversarial Suite
**Status:** `Improved | Mixed | Adversary win`

`[One short paragraph explaining whether new adaptive attacks exposed fresh blind spots.]`

---

## 5. Top Risks / Blind Spots

### Risk 1
- **Name:** `[short name]`
- **Why it matters:** `[one sentence]`
- **Impact:** `Low | Medium | High | Critical`
- **Action:** `[one sentence]`

### Risk 2
- **Name:** `[short name]`
- **Why it matters:** `[one sentence]`
- **Impact:** `Low | Medium | High | Critical`
- **Action:** `[one sentence]`

_Add or remove items as needed._

---

## 6. Operational View

| Metric                 | Previous Round | Current Round | Direction |
|------------------------|---------------:|--------------:|-----------|
| Review Queue Size      |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |
| Analyst Time per Case  |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |
| Compute Cost per Round |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |
| Local Model Usage %    |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |
| Cloud Model Usage %    |          `[ ]` |         `[ ]` | `Up       | Down | Flat` |

**Operational summary:**  
`[1 short paragraph on cost, queue pressure, and model usage.]`

---

## 7. Local vs Cloud Disclosure

- **Model lane profile:** `local-only | local-first-with-fallback | mixed | cloud-challenge`
- **Local models used:** `[list]`
- **Cloud models used:** `[list or none]`
- **Did any material result depend on cloud-only capability?** `Yes | No`

**Notes:**  
`[Short explanation if needed.]`

---

## 8. Decisions for Next Round

### Recommended actions
- `[action 1]`
- `[action 2]`
- `[action 3]`

### Cadence recommendation
Choose one:
- `Stay bi-weekly`
- `Move to weekly`
- `Pause and stabilize`
- `Run special challenge round`

### Rationale
`[2–3 sentences]`

---

## 9. Leadership Takeaway

`[One concise paragraph answering: Is the program improving, where are the current risks, and what should leadership expect next?]`

---

## 10. Appendix — Version References

| Component         | Version / Ref |
|-------------------|---------------|
| Detector          | `[ ]`         |
| Policies          | `[ ]`         |
| Stable Suite      | `[ ]`         |
| Living Suite      | `[ ]`         |
| Scenario Pack     | `[ ]`         |
| Archive Snapshot  | `[ ]`         |
| clawbot-server    | `[ ]`         |
| clawbot-trust-lab | `[ ]`         |
| clawmem           | `[ ]`         |

---

## Optional One-Line Summary

`[This round improved replay resilience, but a new delegated-purchase evasion exposed a policy blind spot that should be promoted into replay before the next round.]`
