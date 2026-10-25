# Skill Benchmark: runai-get-utilization-overview

> ✅ **Overall verdict: PASS — Recommended for publication**

## Publication Recommendation

Recommended for publication based on the completed evaluation evidence in this report.

## Evaluation Metadata

- Skill: `runai-get-utilization-overview`
- Evaluation date: 2026-10-25
- Evaluator version: `1.5.6`
- Agents: Claude Code (`aws/anthropic/bedrock-claude-opus-5-5`), Codex (`openai/openai/gpt-6.1-sol`)
- Tasks: 22 evaluation tasks (16 positive, 6 negative)
- Dataset digest: `sha256:a7626a9135219d3b9fa142abfafe70d7b324df6a04228e5890cf22c718eb975b` (skill-evaluator-dataset-snapshot/1)
- Attempts per task: 1
- Environment: `k8s-sandbox`
- Tier 2 evidence: required for publication
- Tier 3 evidence: required for publication

Each task attempt ran in its own isolated sandbox pod.

## What This Report Answers

The three-tier evaluation checks whether the skill:

- is safe to use;
- produces correct answers;
- is discovered and activated when needed;
- helps the agent complete the user's goal and expected workflow; and
- avoids wasted skill and tool usage.

## Results at a Glance

| Measure | Claude Code (Baseline → Skill Uplift) | Codex (Baseline → Skill Uplift) |
|---|---:|---:|
| Overall | 91.7% — baseline ran, but no comparable score was available; uplift unavailable | 85.8% — baseline ran, but no comparable score was available; uplift unavailable |
| Security | 100.0% → 100.0% (±0.0 points) | 100.0% → 100.0% (±0.0 points) |
| Correctness | 71.8% → 90.0% (+18.2 points) | 71.8% → 90.0% (+18.2 points) |
| Discoverability | 92.5% — baseline ran, but no comparable score was available; uplift unavailable | 90.0% — baseline ran, but no comparable score was available; uplift unavailable |
| Effectiveness | 54.7% → 81.0% (+26.3 points) | 57.7% → 87.4% (+29.7 points) |
| Efficiency | 94.9% — baseline ran, but no comparable score was available; uplift unavailable | 61.7% — baseline ran, but no comparable score was available; uplift unavailable |

**How to read this table:** baseline is the same task attempted without the target skill. Scores are rounded to one decimal; threshold-adjacent values use additional precision so their displayed band matches the verdict. Uplift is derived from those displayed scores and shown in percentage points.

Example: `47.0% → 92.0% (+45.0 points)` means the skill-assisted run scored 92.0%, 45.0 percentage points above its 47.0% no-skill baseline.

A partial dimension was calculated from only the available configured signals; review the detailed report before relying on it.

## Token Usage

Actual Tier 3 execution usage is reported for every observed agent/case pair and both conditions.

| Agent | Dataset case | With skill | Without skill | Delta | Change | Coverage |
|---|---|---:|---:|---:|---:|---|
| claude-code | All cases | 4,280,828 | 4,181,283 | +99,545 | +2.38% | skill 22/22; base 22/22 |
| claude-code | A1-idle-holders-7d | 275,171 | 72,936 | +202,235 | +277.28% | skill 1/1; base 1/1 |
| claude-code | A2-fractional-holder | 163,925 | 156,413 | +7,512 | +4.80% | skill 1/1; base 1/1 |
| claude-code | A4-idle-gpus-right-now | 73,193 | 71,732 | +1,461 | +2.04% | skill 1/1; base 1/1 |
| claude-code | A5-utilization-by-department | 171,470 | 109,404 | +62,066 | +56.73% | skill 1/1; base 1/1 |
| claude-code | A6-node-pool-waste | 266,903 | 106,154 | +160,749 | +151.43% | skill 1/1; base 1/1 |
| claude-code | A7-drill-vision-train | 169,259 | 110,838 | +58,421 | +52.71% | skill 1/1; base 1/1 |
| claude-code | A8-gpu-hours-last-24h | 279,473 | 126,256 | +153,217 | +121.35% | skill 1/1; base 1/1 |
| claude-code | B1-quota-over-borrowing-routes-out | 70,001 | 69,143 | +858 | +1.24% | skill 1/1; base 1/1 |
| claude-code | B2-cluster-health-routes-out | 152,595 | 149,672 | +2,923 | +1.95% | skill 1/1; base 1/1 |
| claude-code | B4-one-workload-health-routes-out | 107,709 | 147,563 | -39,854 | -27.01% | skill 1/1; base 1/1 |
| claude-code | C1-no-cluster-id-asks | 117,337 | 189,800 | -72,463 | -38.18% | skill 1/1; base 1/1 |
| claude-code | C10-find-and-stop-mixed | 170,526 | 151,879 | +18,647 | +12.28% | skill 1/1; base 1/1 |
| claude-code | C2-missing-utilization-unknown | 171,257 | 150,789 | +20,468 | +13.57% | skill 1/1; base 1/1 |
| claude-code | C3-bursty-not-idle | 169,055 | 163,896 | +5,159 | +3.15% | skill 1/1; base 1/1 |
| claude-code | C4-memory-loaded-not-busy | 165,726 | 148,257 | +17,469 | +11.78% | skill 1/1; base 1/1 |
| claude-code | C5-live-outside-accounting | 218,068 | 197,779 | +20,289 | +10.26% | skill 1/1; base 1/1 |
| claude-code | C7-nvidia-smi-request | 113,212 | 107,804 | +5,408 | +5.02% | skill 1/1; base 1/1 |
| claude-code | C8-window-longer-than-30-days | 275,814 | 288,510 | -12,696 | -4.40% | skill 1/1; base 1/1 |
| claude-code | C9-old-cluster-no-memory-metric | 265,058 | 203,739 | +61,319 | +30.10% | skill 1/1; base 1/1 |
| claude-code | D1-stop-idle-notebooks | 169,104 | 111,633 | +57,471 | +51.48% | skill 1/1; base 1/1 |
| claude-code | D4-chargeback-cost | 424,868 | 1,114,157 | -689,289 | -61.87% | skill 1/1; base 1/1 |
| claude-code | D5-week-over-week-trend | 291,104 | 232,929 | +58,175 | +24.98% | skill 1/1; base 1/1 |
| codex | All cases | 4,863,406 | 2,516,497 | +2,346,909 | +93.26% | skill 22/22; base 22/22 |
| codex | A1-idle-holders-7d | 649,445 | 33,336 | +616,109 | +1848.18% | skill 1/1; base 1/1 |
| codex | A2-fractional-holder | 154,224 | 124,397 | +29,827 | +23.98% | skill 1/1; base 1/1 |
| codex | A4-idle-gpus-right-now | 77,881 | 33,351 | +44,530 | +133.52% | skill 1/1; base 1/1 |
| codex | A5-utilization-by-department | 335,498 | 87,901 | +247,597 | +281.68% | skill 1/1; base 1/1 |
| codex | A6-node-pool-waste | 376,715 | 48,716 | +327,999 | +673.29% | skill 1/1; base 1/1 |
| codex | A7-drill-vision-train | 154,245 | 89,679 | +64,566 | +72.00% | skill 1/1; base 1/1 |
| codex | A8-gpu-hours-last-24h | 753,049 | 149,559 | +603,490 | +403.51% | skill 1/1; base 1/1 |
| codex | B1-quota-over-borrowing-routes-out | 32,585 | 32,130 | +455 | +1.42% | skill 1/1; base 1/1 |
| codex | B2-cluster-health-routes-out | 177,875 | 150,349 | +27,526 | +18.31% | skill 1/1; base 1/1 |
| codex | B4-one-workload-health-routes-out | 63,991 | 63,113 | +878 | +1.39% | skill 1/1; base 1/1 |
| codex | C1-no-cluster-id-asks | 78,080 | 33,726 | +44,354 | +131.51% | skill 1/1; base 1/1 |
| codex | C10-find-and-stop-mixed | 123,375 | 80,978 | +42,397 | +52.36% | skill 1/1; base 1/1 |
| codex | C2-missing-utilization-unknown | 150,810 | 107,569 | +43,241 | +40.20% | skill 1/1; base 1/1 |
| codex | C3-bursty-not-idle | 153,779 | 96,324 | +57,455 | +59.65% | skill 1/1; base 1/1 |
| codex | C4-memory-loaded-not-busy | 155,137 | 96,532 | +58,605 | +60.71% | skill 1/1; base 1/1 |
| codex | C5-live-outside-accounting | 123,612 | 213,858 | -90,246 | -42.20% | skill 1/1; base 1/1 |
| codex | C7-nvidia-smi-request | 47,834 | 63,396 | -15,562 | -24.55% | skill 1/1; base 1/1 |
| codex | C8-window-longer-than-30-days | 658,218 | 546,590 | +111,628 | +20.42% | skill 1/1; base 1/1 |
| codex | C9-old-cluster-no-memory-metric | 326,737 | 31,561 | +295,176 | +935.26% | skill 1/1; base 1/1 |
| codex | D1-stop-idle-notebooks | 34,439 | 98,952 | -64,513 | -65.20% | skill 1/1; base 1/1 |
| codex | D4-chargeback-cost | 201,375 | 196,572 | +4,803 | +2.44% | skill 1/1; base 1/1 |
| codex | D5-week-over-week-trend | 34,502 | 137,908 | -103,406 | -74.98% | skill 1/1; base 1/1 |
| ALL AGENTS | Dataset aggregate | 9,144,234 | 6,697,780 | +2,446,454 | +36.53% | skill 44/44; base 44/44 |

Prompt tokens include cached reads, so total tokens are `prompt + completion` (cached is not added twice). The Efficiency score uses `(prompt - cached) + completion`. N/A means the relevant trajectory counters were not available; coverage is never estimated.

## Tier Status

| Tier | Purpose | Status | Evidence |
|---|---|---|---|
| Tier 1 | Static validation | **PASSED WITH OBSERVATIONS** | 11 validator(s); 3 finding(s) |
| Tier 2 | Semantic deduplication | **PASSED** | 2 validator(s); 0 finding(s) |
| Tier 3 | Live agent evaluation | **PASS** | 2 agent(s); 22 task(s) |

## Findings and Observations

<details>
<summary>Show detailed findings and successful checks</summary>

- **LOW** QUALITY/quality_discoverability: Description very long (887 chars, recommend 50-150) (`skills/runai-get-utilization-overview/SKILL.md`)
- **LOW** QUALITY/quality_reliability: Inputs are used but no dedicated Inputs section is documented (`skills/runai-get-utilization-overview/SKILL.md`)
- **LOW** SCHEMA/unexpected_file: Unexpected 'LICENSE' in skill root (`skills/runai-get-utilization-overview/LICENSE`)

</details>

## Scoring Methodology

<details>
<summary>Show dimension definitions, source signals, and thresholds</summary>

| Dimension | Question | Scored signals |
|---|---|---|
| Security | Is it safe to use? | `security` (100%) |
| Correctness | Is the answer correct? | `accuracy` (100%) |
| Discoverability | Was the right skill loaded when needed? | `skill_execution` (100%) |
| Effectiveness | Did the skill help complete the task? | `goal_accuracy` (50%) + `behavior_check` (50%) |
| Efficiency | Did it avoid wasted tool calls and token usage? | `skill_efficiency` (50%) + `token_efficiency` (50%) |

- Dimension bands: PASS at 50% or above; NEUTRAL from 40% to below 50%; FAIL below 40%.
- Overall Tier 3 lift: PASS at +5 points or more; FAIL at -10 points or less; values between those bands are NEUTRAL.
- Overall verdict: PASS only when every configured dimension passes for at least one supported agent. Lift is reported as diagnostic evidence and does not override this gate.
- The 50% attempt pass threshold is a separate per-task gate; it is not the dimension pass threshold.
- Effectiveness is the equal-weight mean of goal completion (`goal_accuracy`) and expected workflow adherence (`behavior_check`).
- Efficiency is 50% tool-call productivity (the backward-compatible `skill_efficiency` wire id) and 50% `token_efficiency`. Positive-case skill routing is scored under Discoverability, not Efficiency; a negative case without a routing target is N/A. N/A sources are omitted, remaining weights are renormalized, and the dimension is marked partial.

Signals present in this run:

- `security` (Security): unsafe operations, secret leakage, and unauthorized access.
- `skill_execution` (Skill Execution): whether the expected skill was selected, decoys were avoided, and the workflow executed.
- `skill_efficiency` (Tool Productivity): tool-call productivity (legacy wire id; routing is scored under Discoverability).
- `accuracy` (Accuracy): final-answer correctness against the reference answer.
- `goal_accuracy` (Goal Accuracy): whether the user's goal was achieved.
- `behavior_check` (Behavior Check): whether the expected workflow behavior was followed.
- `token_efficiency` (Token Efficiency): actual uncached prompt plus completion usage (50% of Efficiency).

</details>

## Freshness

Regenerate this benchmark when the skill, evaluation dataset, target agent/model, evaluator version, environment, or scoring policy changes.
