# Local evaluation — runai-audit-assigned-quota

The certified `BENCHMARK.md` next to `SKILL.md` is generated and signed by
NVSkills CI, so it must not be edited by hand. This file keeps the local runs.

Evaluated with **NVSkills-Eval (NV-BASE / NV-ACES)** under the `external` publication
profile — three tiers: static validation, deduplication, and live agent evaluation
(with-skill vs baseline).

## Tier 1 / 2 — static + dedup (certified)
`nv-base validate --external` → **PASS (exit 0)**.
- All 9 validators pass: security scan, PII, license, schema/governance, unicode,
  code integrity, script lint.
- **Quality Score 89/100 (Grade B)** across correctness, discoverability,
  effectiveness, efficiency.

## Tier 3 — live agent evaluation (deterministic mock)

Each admin question is answered by an agent that calls the Run:ai MCP tools, served by a
deterministic stdio **mock** (`evals/files/mock_mcp.py`) replaying a frozen testbed — so the
with-skill and baseline arms see byte-identical tool data and every run is 100% reproducible,
offline, and network-free.

**Evaluation matrix.** 30 intents spanning four categories, each in **3 meaning-preserving
paraphrases** (90 prompt samples), run in **2 modes** — Arm A: MCP tools only (baseline);
Arm B: MCP + skill — across **3 models** = **540 runs**.

| Category | Intents | What it tests |
|---|---|---|
| Positive | A1–A10 | True starvation / over-subscription, department roll-ups, reclaim headroom |
| False-positive / Trap | B1–B6 | Self-inflicted non-preemptible pending, cross-pool reclaim, "who's hogging X" when the cause is over-subscription |
| Boundary | C1–C6 | Over-borrower with no victim, idle-below-quota, structural full-pool, device-count caveat |
| Out-of-scope / Routing | D1–D8 | Utilization, cluster health, failed-workload, mutation requests, unrelated |

**How a run is graded.** Each question carries a frozen **MUST** / **MUST-NOT** assertion list
written before the run. A run is *correct* when every MUST group matches and no MUST-NOT group
matches affirmatively (matches inside a negated clause don't count — "there is **no** borrower
to reclaim from" must not read as recommending a reclaim). Correctness judges the **answer
only**; tool-call counts are reported separately as a cost metric. An earlier revision also
required `mcp_calls ≤ 3` on out-of-scope questions — that threshold was calibrated on one
model's tool-call appetite and made cross-model comparison invalid (Opus investigates far more
eagerly, so it scored as "wrong" while answering correctly). Because every run's full answer is
persisted, all three models were re-graded under the corrected rule **without re-running**.

**Golden truth (frozen testbed):** `node-pool-2` is over-subscribed (Σ deserved 42 > 32
capacity); `hpc-fill` (16/24) and `starved-svc` (12/16) are genuinely starved; `quota-hog`
(0/2) pending is self-inflicted (non-preemptible over quota) — **no** project over-borrows
on node-pool-2 (no reclaim target; the cause is structural). On `default`, `research-lab`
(24/4) over-borrows but has no same-pool victim (reclaimable-when-demanded); `underused`
(4/12) is idle headroom.

### Overall baseline table

**Three models, 540 runs.**

| Metric | Opus<br>MCP only | **Opus<br>+ skill** | Sonnet<br>MCP only | **Sonnet<br>+ skill** | Codex<br>MCP only | **Codex<br>+ skill** |
|---|---:|---:|---:|---:|---:|---:|
| Correctness — in-scope (A/B/C) | 91% | **98%** | 79% | **91%** | 52% | **88%** |
| Correctness — routing (D) | 71% | 75% | 67% | 79% | 50% | 50% |
| Correctness — overall | 86% | **92%** | 76% | **88%** | 51% | **78%** |
| False-positive rate (any MUST-NOT hit) | 1% | 3% | 0% | 1% | 0% | 3% |
| MCP calls — median / p95 | 5 / 11 | 7 / 10 | 2 / 5 | 6 / 10 | 2 / 5 | 3 / 9 |
| Output tokens — median / p95 | 1,899 / 4,176 | 2,712 / 4,261 | 1,074 / 3,157 | 3,354 / 5,099 | 294 / 633 | 368 / 771 |

**Lift scales inversely with baseline strength — the skill is model-*amplifying*.**

| Model | Baseline (in-scope) | With skill | Lift |
|---|---:|---:|---:|
| Codex `gpt-5.4-mini` | 52% | 88% | **+36 pp** |
| Claude `sonnet-4-6` | 79% | 91% | **+12 pp** |
| Claude `opus-5` | 91% | 98% | **+7 pp** |

Every model converges to **88–98%** with the skill, from baselines spanning 52–91% — a
39-point spread collapses to 10. The knowledge lives in the skill, not in the model; the
stronger models were simply better at inferring it unaided. Opus + skill is effectively
saturated on in-scope work: **100% on positive, 100% on boundary, 94% on traps**.

*Opus = `claude-opus-5`, Sonnet = `claude-sonnet-4-6`, both via `claude -p` headless. Codex =
`azure/openai/gpt-5.4-mini` (GPT-5 series) via `codex exec` against the NVIDIA inference
endpoint. n = 90 samples/arm/model, same deterministic mock. Codex answers are ~10× more
concise (≈300 vs ≈3,000 output tokens)
but send far larger uncached input context per turn (≈52–77 k fresh tokens vs Claude's
≈9–16 k fresh + prompt cache).*

### Per-category accuracy and false positives

Accuracy (FP) per category, n = 30 positive / 18 trap / 18 boundary / 24 routing per arm.

| Category | Mode | Opus | Sonnet | Codex |
|---|---|---:|---:|---:|
| Positive | MCP only | 93% (0%) | 80% (0%) | 50% (0%) |
| Positive | **+ skill** | **100%** (0%) | **97%** (0%) | **97%** (0%) |
| Trap | MCP only | 83% (0%) | 67% (0%) | 39% (0%) |
| Trap | **+ skill** | **94%** (0%) | **78%** (0%) | **83%** (0%) |
| Boundary | MCP only | 94% (0%) | 89% (0%) | 67% (0%) |
| Boundary | **+ skill** | **100%** (0%) | **94%** (0%) | 78% (17%) |
| Routing | MCP only | 71% (4%) | 67% (0%) | 50% (0%) |
| Routing | **+ skill** | 75% (12%) | 79% (0%) | 50% (0%) |

**What the skill buys.** On in-scope questions the fixed audit procedure lifts accuracy in
every category for every model — and crucially holds **false positives at 0% on positive and
trap categories across all three models**: none of them invents a borrower or recommends the
unsafe reclaim, even on adversarial trap phrasings. All three converge on the positive
category at **97–100%**, from baselines of 50–93%.

**What it costs.** Two honest weak spots, both concentrated outside the core:
- **Routing** is the weakest category for every model — the skill sometimes runs the audit on
  utilization or rebalance prompts instead of handing off. Opus also carries the highest
  routing FP (12%), a consequence of investigating eagerly before declining.
- **Codex boundary FP (17%)** is the one model-specific defect: on a structurally full pool it
  falls back to "there is a borrower" (see C5 below). Opus and Sonnet handle it at 0%.

### Wording stability (skill mode, 3 paraphrases per intent)

Does the same question, reworded, produce the same verdict?

| Model | Stable Pass (3/3) | Wording-sensitive (1–2/3) | Stable Fail (0/3) |
|---|---:|---:|---:|
| **Opus** | **26** | 3 — B3, D5, D8 | 1 — D7 |
| **Sonnet** | 23 | 7 — A1, B1, B6, C1, D5, D7, D8 | **0** |
| **Codex** | 22 | 2 — A5, C4 | 6 — B3, C5, D5, D6, D7, D8 |

Stability tracks model strength on the in-scope categories: **Opus is 3/3 on every A, B
(except B3), and C intent** — all 10 positives, 5 of 6 traps, all 6 boundaries. Sonnet has zero
outright failures. Codex's 6 stable-fails are concentrated in routing (D5–D8) plus the two
known weak spots (**B3** cross-pool reasoning, **C5** full-pool-as-borrowing), confirming that
the smaller model is the one that needs the distinctions stated most explicitly.

The recurring wording-sensitive intents across models are the hardest trap (**B6** —
"research-lab is at 600%, reclaim now", where the correct answer is "no same-pool victim,
reclaimable-when-demanded, not a violation to act on now") and the routing prompts
(**D5** rebalance-request, **D7** delete-workloads).

### Token / call / runtime (Sonnet, median / p95, per run)

| Metric | MCP only | + skill | Ratio |
|---|---:|---:|---:|
| MCP API calls | 2 / 5 | 6 / 10 | ~3× |
| Distinct tools | 2 / 4 | 4 / 5 | — |
| Sequential tool-request cycles | 4 / 8 | 9 / 14 | ~2.3× |
| Output tokens | 1,074 / 3,157 | 3,354 / 5,099 | ~3.1× |
| Answer tokens | 244 | 429 | ~1.8× |

### 30-question matrix (per-intent pass rate, x/3 paraphrases)

| ID | Category | Opus<br>only | Opus<br>+skill | Sonnet<br>only | Sonnet<br>+skill | Codex<br>only | Codex<br>+skill |
|---|---|:---:|:---:|:---:|:---:|:---:|:---:|
| A1 | positive | 3/3 | 3/3 | 2/3 | 2/3 | 0/3 | 3/3 |
| A2 | positive | 3/3 | 3/3 | 2/3 | 3/3 | 2/3 | 3/3 |
| A3 | positive | 2/3 | 3/3 | 1/3 | 3/3 | 0/3 | 3/3 |
| A4 | positive | 3/3 | 3/3 | 3/3 | 3/3 | 2/3 | 3/3 |
| A5 | positive | 2/3 | 3/3 | 1/3 | 3/3 | 1/3 | 2/3 |
| A6 | positive | 3/3 | 3/3 | 3/3 | 3/3 | 0/3 | 3/3 |
| A7 | positive | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 |
| A8 | positive | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 |
| A9 | positive | 3/3 | 3/3 | 3/3 | 3/3 | 1/3 | 3/3 |
| A10 | positive | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 |
| B1 | trap | 3/3 | 3/3 | 2/3 | 1/3 | 3/3 | 3/3 |
| B2 | trap | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 |
| B3 | trap | 2/3 | 2/3 | 3/3 | 3/3 | 0/3 | 0/3 |
| B4 | trap | 2/3 | 3/3 | 1/3 | 3/3 | 0/3 | 3/3 |
| B5 | trap | 3/3 | 3/3 | 3/3 | 3/3 | 1/3 | 3/3 |
| B6 | trap | 2/3 | 3/3 | 0/3 | 1/3 | 0/3 | 3/3 |
| C1 | boundary | 3/3 | 3/3 | 2/3 | 2/3 | 1/3 | 3/3 |
| C2 | boundary | 3/3 | 3/3 | 3/3 | 3/3 | 2/3 | 3/3 |
| C3 | boundary | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 |
| C4 | boundary | 2/3 | 3/3 | 2/3 | 3/3 | 3/3 | 2/3 |
| C5 | boundary | 3/3 | 3/3 | 3/3 | 3/3 | 0/3 | 0/3 |
| C6 | boundary | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 |
| D1 | routing | 2/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 |
| D2 | routing | 0/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 |
| D3 | routing | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 | 3/3 |
| D4 | routing | 3/3 | 3/3 | 2/3 | 3/3 | 3/3 | 3/3 |
| D5 | routing | 3/3 | 2/3 | 1/3 | 1/3 | 0/3 | 0/3 |
| D6 | routing | 3/3 | 3/3 | 2/3 | 3/3 | 0/3 | 0/3 |
| D7 | routing | 2/3 | 0/3 | 0/3 | 1/3 | 0/3 | 0/3 |
| D8 | routing | 1/3 | 1/3 | 2/3 | 2/3 | 0/3 | 0/3 |

Reading across a row shows how much of the answer the skill supplies rather than the model.
**A3, A5, B4, B6** are the clearest cases: weak or mixed baselines across all three models,
3/3 with the skill on Opus and Sonnet. **B3** (cross-pool) and **C5** (full pool) are the two
intents where Codex stays at 0/3 even with the skill — the distinctions that need stating more
explicitly. The D-block is where the remaining headroom is, for every model.

### Per-question tool-call dynamics (Sonnet, median of 3 paraphrases)

MCP API calls, distinct tools, and sequential tool-request cycles for **every** question in
both modes — the same user question is sent to both arms.

| ID | Category | Mode | MCP calls | Distinct tools | Cycles | Output tok |
|---|---|---|---:|---:|---:|---:|
| A1 | positive | MCP only | 3 | 3 | 4 | 2,545 |
| A1 | positive | + skill | 7 | 5 | 11 | 4,916 |
| A2 | positive | MCP only | 2 | 2 | 3 | 1,247 |
| A2 | positive | + skill | 10 | 5 | 13 | 4,384 |
| A3 | positive | MCP only | 3 | 3 | 4 | 1,679 |
| A3 | positive | + skill | 7 | 5 | 11 | 3,804 |
| A4 | positive | MCP only | 3 | 3 | 4 | 2,410 |
| A4 | positive | + skill | 10 | 5 | 13 | 4,701 |
| A5 | positive | MCP only | 2 | 2 | 3 | 1,345 |
| A5 | positive | + skill | 7 | 5 | 11 | 3,275 |
| A6 | positive | MCP only | 1 | 1 | 2 | 649 |
| A6 | positive | + skill | 8 | 5 | 13 | 4,177 |
| A7 | positive | MCP only | 1 | 1 | 2 | 695 |
| A7 | positive | + skill | 8 | 5 | 13 | 4,632 |
| A8 | positive | MCP only | 1 | 1 | 2 | 585 |
| A8 | positive | + skill | 7 | 4 | 10 | 4,512 |
| A9 | positive | MCP only | 3 | 3 | 5 | 1,925 |
| A9 | positive | + skill | 8 | 5 | 12 | 4,451 |
| A10 | positive | MCP only | 3 | 3 | 4 | 699 |
| A10 | positive | + skill | 4 | 3 | 7 | 1,989 |
| B1 | trap | MCP only | 3 | 3 | 4 | 934 |
| B1 | trap | + skill | 0 | 0 | 1 | 469 |
| B2 | trap | MCP only | 5 | 4 | 9 | 3,672 |
| B2 | trap | + skill | 7 | 5 | 10 | 4,769 |
| B3 | trap | MCP only | 2 | 1 | 4 | 1,280 |
| B3 | trap | + skill | 7 | 4 | 9 | 3,540 |
| B4 | trap | MCP only | 2 | 2 | 3 | 1,403 |
| B4 | trap | + skill | 6 | 4 | 9 | 3,766 |
| B5 | trap | MCP only | 2 | 1 | 5 | 1,090 |
| B5 | trap | + skill | 3 | 3 | 8 | 3,764 |
| B6 | trap | MCP only | 2 | 2 | 7 | 1,300 |
| B6 | trap | + skill | 3 | 3 | 10 | 2,409 |
| C1 | boundary | MCP only | 2 | 2 | 3 | 1,126 |
| C1 | boundary | + skill | 4 | 4 | 7 | 2,843 |
| C2 | boundary | MCP only | 3 | 3 | 4 | 2,143 |
| C2 | boundary | + skill | 7 | 5 | 10 | 3,795 |
| C3 | boundary | MCP only | 1 | 1 | 2 | 624 |
| C3 | boundary | + skill | 8 | 5 | 12 | 4,798 |
| C4 | boundary | MCP only | 0 | 0 | 1 | 663 |
| C4 | boundary | + skill | 3 | 3 | 5 | 2,024 |
| C5 | boundary | MCP only | 0 | 0 | 0 | 646 |
| C5 | boundary | + skill | 9 | 5 | 13 | 4,098 |
| C6 | boundary | MCP only | 0 | 0 | 2 | 572 |
| C6 | boundary | + skill | 0 | 0 | 2 | 840 |
| D1 | negative | MCP only | 2 | 2 | 4 | 1,316 |
| D1 | negative | + skill | 3 | 3 | 8 | 2,132 |
| D2 | negative | MCP only | 2 | 2 | 4 | 1,498 |
| D2 | negative | + skill | 6 | 2 | 9 | 2,984 |
| D3 | negative | MCP only | 3 | 3 | 4 | 734 |
| D3 | negative | + skill | 3 | 3 | 4 | 999 |
| D4 | negative | MCP only | 0 | 0 | 3 | 640 |
| D4 | negative | + skill | 0 | 0 | 3 | 570 |
| D5 | negative | MCP only | 0 | 0 | 1 | 200 |
| D5 | negative | + skill | 8 | 5 | 12 | 4,574 |
| D6 | negative | MCP only | 2 | 2 | 7 | 1,593 |
| D6 | negative | + skill | 3 | 2 | 5 | 2,370 |
| D7 | negative | MCP only | 2 | 1 | 8 | 1,347 |
| D7 | negative | + skill | 0 | 0 | 6 | 887 |
| D8 | negative | MCP only | 0 | 0 | 0 | 36 |
| D8 | negative | + skill | 0 | 0 | 0 | 28 |

**How much MCP investigation equals one skill request.** On the deep positive audits (A1–A9)
the baseline stops at **1–3 calls / 1–3 tools / 2–5 cycles**, while the skill spends
**7–10 calls / 4–5 tools / 10–13 cycles** to reach the same question's answer — roughly
**3–8× the investigation** for the same user prompt. The gap is not uniform: on narrow
questions (A10, C6) the two arms converge, and on out-of-scope prompts (D4, D8) both
correctly do nothing. The over-trigger is visible as D5's `0 → 8` calls.

**Full answers.** Every run's complete answer text is persisted alongside its metrics, so any
verdict in this document can be re-read or re-graded offline without re-running the models.

## Cross-model notes

All three models run the identical 30 × 3 × 2 matrix against the same mock. Claude models load
the skill on demand via the Skill tool; **Codex has no skill mechanism**, so the instructions are
supplied as `AGENTS.md` in the working directory — always in context rather than invoked. This
asymmetry is worth noting when comparing Codex's routing numbers, since it cannot choose *not*
to load the skill.

Model-specific findings, each verified by reading the stored answers rather than trusting the
grader:

- **Codex · C5 (boundary), FP 3/3** — misdiagnoses a structurally-full pool as *"a borrowing
  problem"* with *"a clear borrower"* — the exact false positive the skill exists to prevent.
  Opus and Sonnet both handle C5 at 0% FP. This is the single genuinely model-specific defect,
  and it motivates stating the structural-vs-borrowing distinction more explicitly.
- **Codex · B3 (cross-pool trap), 0/3** — safely *refuses* the reclaim (FP 0) but omits the
  explicit "different pool" reasoning the golden requires: a completeness miss, not an unsafe
  answer. Worth reading as a grading-strictness result, not a safety failure.
- **Opus · routing FP 12%** — the highest routing FP of the three, a by-product of
  investigating eagerly (median 5 MCP calls, p95 11, even in baseline mode) before declining.
  It answers correctly but spends calls doing so.
- **Opus · in-scope saturation** — 100% positive, 100% boundary, 94% trap with the skill; the
  remaining headroom is entirely in routing discipline, not in audit correctness.

## Read-only Kubernetes fallback

| Question | Answer |
|---|---|
| **What is defined in the skill?** | **Nothing** — the audit is pure-MCP by design. `SKILL.md` and `references/` contain no Kubernetes, `kubectl`, or cluster-level path; all evidence comes from the five read-only Run:ai MCP tools. |
| **What does the environment support?** | The deterministic eval environment exposes **only** the mock MCP server. No kubeconfig, no cluster access is provided or required. |
| **What is the defined failure mode?** | If the MCP quota source is unavailable the skill **stops and says so** — it retries once on a transient error and never fabricates quota numbers. |
| **What remains untested?** | Behaviour against a **real** unavailable-MCP cluster (the mock always answers). Covered by assertion, not by a live outage test. |

A behaviour case (`boundary-no-k8s-fallback`) asserts this directly: asked to "check quota with
`kubectl` instead", the skill must decline the Kubernetes path, state that it is MCP-only, and
refuse to fabricate numbers.

## Actions this benchmark surfaces
- **Trigger tightening** — narrow the skill description so utilization (D2) and
  rebalance-request (D5) prompts route away instead of loading the audit; this is the only
  measured regression, and it affects both models.
- **B6 wording** — strengthen the "over-quota with no same-pool victim → reclaimable-when-
  demanded, not an urgent violation" branch so it survives paraphrase.
- **C5 on weaker models** — the "full pool = structural, not a borrower" distinction should be
  made explicit enough that a smaller model (codex) does not fall back to "there is a borrower".

## Reproduce
```shell
# Tier 1/2 (static + dedup):
nv-base validate skills/runai-audit-assigned-quota --external

# Tier 3 (this benchmark): 540 runs against the in-repo deterministic mock, driven headless.
#   Arm A (baseline): agent + MCP mock only.   Arm B (skill): same, with the skill loaded.
#   Claude: claude -p <prompt> --mcp-config <mock> --strict-mcp-config --allowedTools <5 quota tools>
#   Codex : codex exec <prompt> -c mcp_servers.runai.args=["…/evals/files/mock_mcp.py"] \
#           --dangerously-bypass-approvals-and-sandbox   (NVIDIA key; enterprise login set aside)
# Each answer is graded against a frozen golden verdict (per-intent MUST / MUST-NOT assertions).
```
The mock is `evals/files/mock_mcp.py`; no cluster or network is required. NV-BASE/NVCARPS
certification (signing + certified benchmark) is a separate official CI pass.
