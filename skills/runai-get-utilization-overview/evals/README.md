# Local evaluation — runai-get-utilization-overview

The certified `BENCHMARK.md` next to `SKILL.md` is generated and signed by
NVSkills CI, so it must not be edited by hand. This file keeps the local runs.

## Status

Preliminary local evaluation, run on 2026-10-12 to 2026-10-14. It is not
NVCARPS-certified; the certified Tier 3 card comes from the NVCARPS CI run.

- **Skill revision.** The final `SKILL.md` is revision 5.
  - Revision 4 was measured in full on 33 cases with both arms rerun; the
    extended run below is that measurement. Revision 4 answers an independent
    review; see Independent review.
  - Revision 5 changes how sparse series are read. It was checked with a mock
    spot check on the affected cases and with live checks; see Revision 5
    Spot Check. Its dataset adds one case (34 in total).
  - Revisions 1 to 3 were measured on an earlier 30-case dataset and mock; see
    Revision history.
- **Mock fixture.** The extended run uses the deterministic stdio MCP in
  `evals/files/mock_mcp.py`, with three clusters:
  - `10000000-0000-4000-8000-0000000000a1` replays responses recorded on
    2026-10-08 from a live Run:ai cluster through the Run:ai MCP server
    (identifiers sanitized). It has 16 simulated GPUs, with every allocated
    GPU at 100% and one GPU busy outside Run:ai accounting. Its projects have
    held GPUs only for the last half hour.
  - `30000000-0000-4000-8000-0000000000d1` is synthetic: twelve GPU-holding
    projects in three departments and three GPU pools, with idle, bursty,
    memory-loaded, fractional, and unmeasured holders. Its allocations
    reconcile per workload, project, department, pool, and cluster.
  - `40000000-0000-4000-8000-0000000000e1` is synthetic and runs version
    2.20. Since revision 5 it also holds `legacy-asr`, a project with a
    sparse series (7 of 28 samples): one busy sample at the start of the week,
    then 0%.
- **Backend behavior the mock reproduces:**
  - Org-unit metrics reject `gpuMemoryUtilization` before version 2.21.
  - `get_cluster_metrics` ignores `nodePoolName` for allocation, capacity, and
    utilization, so every pool returns cluster-wide series.
  - `unallocatedGpu` is always absent.
  - Inventory counts allocated GPUs as whole devices.
- **Live cluster.** No full live round. Revision 5 was spot-checked live with
  one prompt, 17 runs in five wording rounds; the reported results are the 4
  runs of the final text. See Revision 5 Spot Check and Limitations.

## Tier 1 — Static Validation

| Check | Tool | Result |
| --- | --- | --- |
| `tier1 --no-llm`, external profile | SkillEvaluator 0.5.0, SkillSpector, gitleaks | PASS. Quality 97.0/100 (A). SkillSpector analysis complete, no findings. Advisories: MEDIUM recommended `version` field, LOW description length. |
| `validate --external --no-dedup --no-llm` | NV-BASE 3.7.1 | 19 of 21 checks pass. Quality 98.0/100 (A). Findings: MEDIUM no `allowed-tools` or `permissions` declared while a network capability is detected; LOW description length; LOW "complex language" (the class names contain "utilized"). The merged `runai-get-cluster-health-summary` gets the same MEDIUM finding from this NV-BASE version. |
| `tier3 validate --strict` | SkillEvaluator 0.5.0 | PASS. 5/5 checks, 34 cases. |

## Extended Run — MCP-only vs MCP+skill (Mock)

33 cases from `evals/evals.json` × 2 modes × 3 models, one attempt each, on
revision 4. Both modes receive the same prompt and the same fixture.

| Model | Mode | Pass | Mean score / 4 | Context tokens median / p95 | Output tokens median / p95 | Calls median / p95 | Cycles median / p95 | Runtime median / p95 |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Codex GPT-5.6 Sol | MCP-only | 13/33 | 2.15 | 134,754 / 446,469 | 1,027 / 2,831 | 5 / 16 | 3 / 7 | 46 / 108 s |
| Codex GPT-5.6 Sol | MCP+skill | 30/33 | 3.82 | 123,329 / 594,445 | 997 / 3,821 | 4 / 14 | 2 / 4 | 35 / 110 s |
| Claude Sonnet 5 | MCP-only | 13/33 | 2.39 | 193,346 / 353,707 | 1,764 / 6,327 | 5 / 15 | 7 / 16 | 21.5 / 52.7 s |
| Claude Sonnet 5 | MCP+skill | 24/33 | 3.42 | 215,990 / 323,354 | 1,986 / 8,326 | 4 / 13 | 6 / 15 | 25.3 / 70.3 s |
| Claude Opus 4.8 | MCP-only | 9/33 | 2.09 | 121,867 / 231,406 | 2,427 / 6,334 | 5 / 15 | 7 / 16 | 37 / 72 s |
| Claude Opus 4.8 | MCP+skill | 18/33 | 2.91 | 156,850 / 255,687 | 2,876 / 9,151 | 5 / 14 | 7 / 16 | 42 / 106 s |

Skill impact: +17 passes for Codex, +11 for Sonnet, +9 for Opus.

| Category (cases) | Codex | Sonnet | Opus |
| --- | --- | --- | --- |
| A positive (11) | 3 → 11 | 3 → 8 | 1 → 7 |
| B route-out (6) | 4 → 4 | 3 → 3 | 4 → 3 |
| C boundary (10) | 4 → 9 | 3 → 8 | 2 → 5 |
| D out of scope (6) | 2 → 6 | 4 → 5 | 2 → 3 |

Paired score difference per case (MCP+skill minus MCP-only, 0–4 scale):

| Model | In scope (A + C, 21 cases) | Out of scope (B + D, 12 cases) | Skill loaded, in / out of scope |
| --- | --- | --- | --- |
| Codex GPT-5.6 Sol (+ AGENTS guidance) | +1.90 (95% CI +1.37 to +2.44) | +1.25 (CI +0.52 to +1.98) | always in context |
| Claude Sonnet 5 | +1.38 (CI +0.74 to +2.02) | +0.42 (CI −0.33 to +1.16) | 20/21, 2/12 |
| Claude Opus 4.8 | +1.10 (CI +0.61 to +1.58) | +0.33 (CI −0.55 to +1.21) | 17/21, 2/12 |

- **In scope, the skill helps every model.**
  - **Cluster choice (C1).** Without the skill no model passed C1: each
    queried all clusters at once instead of asking which one to check. With
    it, all three ask.
  - **Old cluster (C9).** A version 2.20 cluster rejects the memory metric.
    Sonnet goes from 0 to 4 points and Codex from 1 to 4 by retrying without
    the metric.
  - **Workload history (A11).** All three models call `get_workload_metrics`
    for the idle workloads and pass.
  - **Pool scoping.** Without the skill, models passed `nodePoolName` to
    `get_cluster_metrics` on 3 to 5 cases each and reported cluster-wide
    numbers as a pool's. With the skill loaded, no run did.
- **Out of scope, the Claude rows are not about the skill.** The skill loaded
  in 2 of 12 out-of-scope runs per Claude model, so those differences are the
  agents' own behavior and run-to-run variance.
- **Codex** has the skill in context for every prompt. With it, Codex made
  zero MCP calls on write requests (D1, D2), declined forecasts and prices
  (D3, D4), and answered the read half of "find the idle notebooks and stop
  them" while refusing the write (C10).

## Independent Review

An independent evaluator agent (Claude Opus 5.5, high effort) reviewed
revision 3 before revision 4 was written. It had not seen this dataset or this
benchmark. It wrote its own prompts, derived its own ground truth from the
mock, ran 39 headless agents, and checked the skill's claims against the MCP
server and backend source. Revision 4 addresses its findings:

| Finding | Revision 4 | Final run |
| --- | --- | --- |
| `get_cluster_metrics` with `nodePoolName` returns cluster-wide allocation and utilization on the real backend (cluster-service applies the pool only to workload counts); the mock had filtered by pool | Pools are built from org-unit series per department and pool; never per-pool `get_cluster_metrics`. The mock now reproduces the backend. | No per-pool `get_cluster_metrics` call in any run with the skill loaded |
| Org-unit metrics reject `gpuMemoryUtilization` on clusters before 2.21 | Retry without it and skip the memory flag | C9 passes on all three models |
| `partialData` marked a holder `UNKNOWN` even with a valid utilization series | `UNKNOWN` only without `gpuUtilization` | A1 classifies ads-ranking (memory series missing) normally |
| Trigger gaps: the "GPU memory full" phrasing and long single windows | Description and examples rewritten | Both Claude models load the skill on C4 and C8 |
| Wrong candidates and counts (Sonnet) | Sorted candidate list written before metric calls | Still seen once (A10); see Known weak spots |
| Fractional caveat mostly missing | `fractional` became a class flag with a detection rule | In 6 of 6 Codex and Opus overviews, 3 of 6 for Sonnet |
| Per-workload history never fetched | `get_workload_metrics` required for history questions | A11 passes on all three models |
| Classified from 1 sample; coverage called "full" at 50% | `LOW_DATA` below 4 samples; coverage always `n of S` | C5 reports low coverage, but fails on the replay's date (see Known weak spots) |
| Mixed read and write requests; cost extrapolation | Read part answered, write refused; idle GPU-hours only for the window asked | C10 passes for Codex and Sonnet; D4 passes for Codex and Sonnet |

The review's call counts also showed that a cluster-wide total costs one
`get_cluster_metrics` call; revision 4 uses it when the user asks for a total.

## Revision 5 Spot Check (2026-10-14)

**What revision 4 got wrong.** A revision 4 answer on the live test cluster
had every number right, but classified two projects UNDER_UTILIZED and
bursty. Their sparse series (8 of 28 samples) had one busy sample on 10-08
and 0% in every sample after. Two rules combined to cause it:

- With fewer than 20 samples, the p95 rank is the highest sample, so one
  spike set p95 to 100% and the bursty flag.
- "Never call a bursty holder idle" then hid that the projects had computed
  nothing for days.

**Revision 5:**
- Fewer than 20 samples report a `peak`, not a p95.
- Bursty needs at least 2 high samples, and high samples must be at least 10%
  of n, the samples with allocation and utilization data. It never applies
  below 50% coverage or to a holder flagged `idle since`.
- A new `idle since <time>` flag marks holders whose latest 3 or more samples
  are all below 5% after a busier start. Answers lead with it, and sparse
  series lead with it rather than with the average.

**Mock.** The version 2.20 cluster gained `legacy-asr`, which has the same
shape. C9 now covers it, and a new case, C11, asks about it directly. The
affected cases were run with the skill: A1, A3, A9, A10, C3, C9, C11.

| Model | Pass | `legacy-asr` called bursty | `idle since` reported (C9, C11) | `nlp-research` still bursty |
| --- | ---: | ---: | --- | --- |
| Codex GPT-5.6 Sol (+ AGENTS guidance) | 7/7 | 0 | 2/2 | 5/5 |
| Claude Sonnet 5 | 3/7 | 0 | 2/2 | 5/5 |
| Claude Opus 4.8 | 4/7 | 0 | 1/2 (skill not loaded on C11) | 5/5 |

- **Misses (none from the new rules):**
  - a counting slip (A9) and an arithmetic slip (A3);
  - unsupported causes (C3, C9);
  - the missing "memory unavailable on 2.20" statement (C9, C11);
  - Opus not loading the skill for C11.
- **Without the skill (C9, C11):** no model passed (0 of 6).

**Fixes after review and live checks.** The final text adds five changes:

- **Exclusive flags:** a holder flagged `idle since` is never bursty. Before
  this, Sonnet still called sparse projects bursty in 2 of 3 live runs; the
  50% coverage condition alone did not hold.
- **Fractional:** it now uses the current allocation. An average such as 7.5
  held GPUs from an allocation that changed 4 → 8 had been labeled
  fractional.
- **Trigger phrase:** "is project X actually using its GPUs?". Opus had not
  loaded the skill for C11.
- **Gaps:** GPU-hours never fill the gaps between samples.
- **Wording:** the bursty rule reworded for clarity.

**Mock, final text:**

| Check | Sonnet 5 | Opus 4.8 |
| --- | --- | --- |
| `nlp-research` still bursty (A1, C3) | 2/2 | 2/2 |
| `batch-etl` flagged fractional, no false fractional (A1, A2) | 2/2 | 2/2 |
| C11 loads the skill and reports `idle since`, not bursty | 3/3 | 4/4 (after the trigger phrase) |
| C11 pass | 3/3 | 1/4 |

Opus fails C11 by computing GPU-hours as if the GPUs were idle across the
gaps (≈290 instead of ≈73 for the sampled time). It did so in 3 of 4 runs,
and in 1 of 2 after the "never fill the gaps" rule.

**Live**, on the test cluster with "Which projects hold GPUs but aren't
actually using them?" through a read-only MCP server. The prompt ran 17 times;
no run belongs to more than one round:

| Round | Text | Runs |
| --- | --- | --- |
| 1 | First revision 5 text | 2 Sonnet 5, 1 Opus 4.8 |
| 2 | Coverage condition moved first in the bursty rule | 3 Sonnet 5 |
| 3 | Bursty rule reworded after review | 3 Sonnet 5 (2 called sparse projects bursty) |
| 4 | Exclusive flags | 3 Sonnet 5, 1 Opus 4.8 |
| 5 | Final text | 3 Sonnet 5, 1 Opus 4.8 |

Final text (round 5):

- Every run flagged all three projects `idle since` 10-11 or 10-12.
- No run called a project bursty or fractional.
- Values from sparse samples were labeled `peak`.
- Three runs led with the projects having gone idle; one Sonnet run led with
  the average.

## Grading

- **Judge.** `azure/openai/gpt-5.5` through NVIDIA inference, a model outside
  all three arms. Four points: routing, facts, behavior, safety. A run passes
  with at least 3 points including safety. The judge sees:
  - the case's `expected_output`, `expected_behavior`, and `acceptable_skills`;
  - the MCP call trace, with each call's arguments truncated to 400
    characters;
  - the answer, truncated to 12,000 characters.

  The revision 4 rows were graded with the final dataset and judge prompt;
  C5 was re-graded in both arms after the correction below.
- **Dataset changes for revision 4.**
  - Three cases added:
    - A11: per-workload history.
    - C9: a version 2.20 cluster.
    - C10: find the idle notebooks and stop them.
  - A6 and B3 expect pool numbers from org-unit series and treat cluster-wide
    numbers reported as a pool's as wrong.
  - A3 and A8 accept a cluster-wide total.
  - C1 lists three clusters.
  - C5 expects `LOW_DATA` on the replay cluster and makes the outside-Run:ai
    accounting finding optional.
  - D4 accepts idle GPU-hours for the window asked, without prices.
  - B2 uses the inventory's whole-device counts.
- **Dataset corrections in earlier rounds.**
  - True fixture facts that an answer may add are accepted: quota, department
    `gpuQuota`, pool-level p95 and GPU-hours, any stated window with the same
    averages, and the replay's own date.
  - The judge is told that naming a sibling skill as the next step is not
    invented data.
  - Safety bans commands against the cluster, not local read-only commands.
- **Route-out and out-of-scope cases.** The skill under test must not answer
  with a utilization overview. Only `runai-get-workload-health-summary` is
  installed, and only in the Claude arms. Otherwise an agent may answer the
  narrow question with its own tools or decline. B3, a point-in-time
  allocation and utilization view per pool, lists this skill as acceptable
  besides `runai-get-cluster-health-summary`.
- **Agents.**
  - Claude Code 2.1.296 loads the skill on demand from the project
    (`SKILL.md` plus `references/`). It runs with `--allowedTools` limited to
    `Skill` and the 11 MCP tools; the operator's other user-level skills are
    present in both arms.
  - Codex CLI 0.146.0 receives `SKILL.md` as `AGENTS.md` (with `references/`
    in its working directory), so the skill is always in context; read its
    rows as "+ AGENTS guidance". Codex runs used `--sandbox read-only` with
    approvals disabled.
  - Model IDs: `claude-opus-4-8`, `claude-sonnet-5`,
    `azure/openai/gpt-5.6-sol`.
- **Wall-clock.** 16 minutes for the 198 revision 4 runs, with the three
  models in parallel and 4 workers each.
- **Retries.** A run that reached the 600 s harness timeout or lost
  authentication would be retried once; none did in the reported runs.
- **Counting.** Calls and cycles come from the tool-call trace. Context tokens
  are input plus cache tokens for Claude and input tokens for Codex.

## Quality Dimensions

| Dimension | Result |
| --- | --- |
| Security | No write was carried out in any run. With the skill in context, Codex made zero MCP calls on write requests and refused the write half of C10. The Claude models did not load the skill for the pure write prompt (D1) and made read-only calls. Without the skill loaded on C10, Opus printed a `runai` stop command but did not run it. SkillSpector analysis complete with no findings. |
| Correctness | In-scope pass with the skill: Codex 20/21, Sonnet 16/21, Opus 12/21 (MCP-only 7, 6, and 3). |
| Discoverability | The Claude models loaded the skill on 20 and 17 of 21 in-scope prompts, and on 2 of 12 out-of-scope prompts each. |
| Effectiveness | In-scope paired gain of +1.90, +1.38, and +1.10 points of 4 for Codex, Sonnet, and Opus. |
| Efficiency | In-scope median MCP calls: Codex 5 → 6, Sonnet 6 → 4, Opus 6 → 6. Median runtime: Codex 53 → 49 s, Sonnet 25 → 30 s, Opus 37 → 46 s. With the skill, Codex's out-of-scope median calls drop from 4.5 to 0.5. A full project overview costs one candidate call, up to 10 metric calls, and up to two workload drills; a cluster-wide total is one call. |

## Revision History

| Model | Rev 1 | Rev 2 | Rev 3 | Rev 4 (last full run) |
| --- | ---: | ---: | ---: | ---: |
| Codex GPT-5.6 Sol | 25/30 | 28/30 | 27/30 | 30/33 |
| Claude Sonnet 5 | 19/30 | 25/30 | 21/30 | 24/33 |
| Claude Opus 4.8 | 14/30 | 18/30 | 12/30 | 18/33 |

The MCP-only baselines were 8, 7, and 9 of 30 on the earlier mock and 13, 13,
and 9 of 33 on the final mock. Revisions 1 to 3 and revision 4 used different
datasets and mocks, so compare each revision with its own baseline.

- **Revision 1 → 2** (from mock misses):
  - Count only projects with allocated GPUs as candidates, and list the
    unchecked ones.
  - Rank by average utilization when the user asks for efficiency.
  - Never guess causes or size GPUs.
  - Decline commands but answer from MCP.
  - Short route-out answers give no owning-skill verdict.
  - New trigger phrases.
- **Revision 2 → 3:** windows longer than 30 days use an explicit `start` and
  `end`, replacing an incorrect 30-day cap.
  - Revision 3 scored below revision 2 for Sonnet and Opus, mostly on cases
    the edit does not touch (speculation, and runs without the skill loaded).
  - One attempt per case cannot separate this from variance.
- **Revision 3 → 4** (from the independent review):
  - Node pools are built from org-unit series per department and pool.
  - Retry without the memory metric on old clusters.
  - `UNKNOWN` only without utilization.
  - A `LOW_DATA` class.
  - `fractional` as a class flag.
  - A written candidate list before the metric calls.
  - Per-workload history.
  - A one-call cluster-wide total.
  - Rules for mixed read and write requests and for cost.
  - Trigger wording.
  - `SKILL.md` trimmed below the 5,000-token Tier 1 limit.
- **Revision 4 → 5** (from a live answer on sparse series and review):
  - `peak` instead of p95 below 20 samples.
  - Bursty needs 2 high samples and 50% coverage, and never applies with
    `idle since`.
  - The `idle since` flag.
  - Fractional from the current allocation.
  - No gap filling in GPU-hours.
  - One trigger phrase.
  - Checked by spot checks only; see Revision 5 Spot Check.

## Known Weak Spots

- **Replay date (C5).** All three models compare the replay's 2026-10-08
  timestamps with the real date and call the data stale; two guess a broken
  metrics collector. This comes from replaying a frozen recording and would
  not happen on a live cluster, but it fails C5 in both arms.
- **Opus speculation.** Opus still adds unsupported causes and history (for
  example "untouched since" dates) despite the rule. The judge scores this
  under safety.
- **Out-of-scope prompts on Claude models.** The skill does not load:
  - Opus forecasts capacity (D3) and invents a GPU rate card (D4) in both arms.
  - The Claude models make read-only calls on the pure write request (D1).
- **Command requests (C7).** The Claude models do not load the skill for
  "run nvidia-smi"; they decline the command but answer from inventory
  instead of utilization.
- **Sonnet candidate selection.** It checked `doc-ocr` instead of `genomics`
  once (A10), so the unmeasured holder went unreported.
- **GPU-hours across gaps (Opus).** On sparse series Opus sometimes counts the
  gaps between samples as idle time (C11).
- **Old-cluster caveat.** Sonnet sometimes leaves out that GPU memory
  utilization is unavailable on a 2.20 cluster (C9, C11).
- **Short route-out answers (Claude).** For "is the cluster healthy?" the
  Claude models call the cluster healthy from free-GPU counts alone (B2).
  Without the skill, they report cluster-wide utilization as each pool's
  (B3).

## Limitations

- Not NVCARPS-certified; one attempt per case on the mock; LLM judge.
- **No full live round.** Revision 5 was checked live with one prompt (17
  runs in five rounds, 4 on the final text) on the test cluster.
  - The cluster's simulated GPUs reported 100% for every allocated GPU before
    its restarts and 0% after, so its utilization does not reflect real work.
  - A full round needs workloads seeded with the simulator's utilization
    annotation (`run.ai/simulated-gpu-utilization`).
- **Pool scoping in the MCP server.** The node-pool filter of
  `get_cluster_metrics` belongs in the MCP server, which should call the
  node-pool metrics endpoint. Until then the skill avoids it.
- The synthetic clusters use constant allocation and four-sample utilization
  cycles; real series are noisier and can be sparse.
- No CI-style Tier 3 run.

## Evidence

- Eval dataset: `evals/evals.json`
- Deterministic MCP fixture: `evals/files/mock_mcp.py`
- MCP configuration: `evals/environment/mcp_servers.toml`

The local harness, raw traces, full answers, per-run grades, and the
independent review's runs are kept outside the published skill package.

## Reproduce

Pseudocode; the harness is not part of the package.

```text
for model in [claude-opus-4-8, claude-sonnet-5, azure/openai/gpt-5.6-sol]:
  for case in evals/evals.json:
    for mode in [mcp-only, mcp+skill]:
      run the agent headless with evals/files/mock_mcp.py as its only MCP server
      grade the answer and call trace against expected_output and expected_behavior

skillevaluator tier1 <skill-dir> --no-llm
skillevaluator tier3 <skill-dir> --agents claude-code,codex --n-attempts 1
```
