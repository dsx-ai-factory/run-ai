# Local evaluation — runai-get-cluster-health-summary

The certified `BENCHMARK.md` next to `SKILL.md` is generated and signed by
NVSkills CI, so it must not be edited by hand. This file keeps the local runs.

## Status

Preliminary local evaluation, run on 2026-10-08 to 2026-10-14. It is not
NVCARPS-certified; the certified Tier 3 card comes from the NVCARPS CI run.

- **Skill revision.** The final `SKILL.md` is revision 5.
  - The extended mock run and both full live rounds measured revisions 1 to 3.
  - Revision 4 adds two event-reading rules and review fixes, and was checked
    with live spot checks.
  - Revision 5 reads the workload spec for pending work that fits the free
    GPUs and stops scoping cluster metrics to a pool. It was checked with a
    mock spot check and a live spot check.
  - See Revision history.
- **Mock fixture.** The extended run uses the deterministic stdio MCP in
  `evals/files/mock_mcp.py`. Cluster `10000000-0000-4000-8000-0000000000a1`
  replays responses recorded on 2026-10-07 from a live Run:ai cluster through
  the Run:ai MCP server (identifiers sanitized); it is idle and healthy.
  Cluster `20000000-0000-4000-8000-0000000000c1` is synthetic: it uses the
  current MCP response shapes and holds one fault per node pool.
- **Live cluster.** An independent evaluator tested revision 2 and then
  revision 3 against a live, seeded Run:ai cluster; see Live-Cluster
  Evaluation. The revision 2 findings produced revision 3.

## Tier 1 — Static Validation

| Check | Tool | Result |
| --- | --- | --- |
| `tier1 --no-llm`, external profile | SkillEvaluator 0.5.0, SkillSpector, gitleaks | PASS. Quality 97.0/100 (A). SkillSpector analysis complete. Advisories: MEDIUM LP3 (no `allowed-tools` declared), MEDIUM recommended `version` field, LOW description length. |
| `validate --external --no-dedup --no-llm` | NV-BASE 3.7.1 | PASS. Quality 98.8/100 (A). |
| `tier3 validate --strict` | SkillEvaluator 0.5.0 | PASS. 5/5 checks, 30 cases. |

## Extended Run — MCP-only vs MCP+skill (Mock)

30 cases from `evals/evals.json` × 2 modes × 3 models, one attempt each. Both
modes receive the same prompt and the same fixture.

| Model | Mode | Pass | Mean score / 4 | Context tokens median / p95 | Output tokens median / p95 | Calls median / p95 | Cycles median / p95 | Runtime median / p95 |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Codex GPT-5.6 Sol | MCP-only | 15/30 | 2.73 | 72,263 / 242,409 | 496 / 1,662 | 2 / 11 | 1 / 5 | 25 / 155 s |
| Codex GPT-5.6 Sol | MCP+skill | 26/30 | 3.73 | 111,054 / 376,932 | 754 / 2,142 | 4 / 12 | 2 / 3 | 29.8 / 83.5 s |
| Claude Sonnet 5 | MCP-only | 14/30 | 2.50 | 107,720 / 330,125 | 582 / 2,812 | 1 / 9 | 2 / 12 | 9.8 / 33 s |
| Claude Sonnet 5 | MCP+skill | 22/30 | 3.43 | 161,542 / 317,536 | 1,027 / 5,187 | 2 / 11 | 3.5 / 14 | 14.3 / 49.8 s |
| Claude Opus 4.8 | MCP-only | 13/30 | 2.57 | 93,293 / 165,633 | 1,120 / 3,847 | 2 / 6 | 3.5 / 8 | 23.1 / 55.1 s |
| Claude Opus 4.8 | MCP+skill | 19/30 | 3.33 | 126,534 / 231,026 | 1,690 / 5,081 | 2 / 11 | 5 / 13 | 25 / 61.1 s |

Skill impact: +11 passes for Codex, +8 for Sonnet, +6 for Opus.

| Category (cases) | Codex | Sonnet | Opus |
| --- | --- | --- | --- |
| A positive (10) | 5 → 9 | 2 → 8 | 3 → 8 |
| B route-out (6) | 3 → 4 | 5 → 3 | 3 → 2 |
| C boundary (8) | 2 → 7 | 4 → 7 | 3 → 7 |
| D out of scope (6) | 5 → 6 | 3 → 4 | 4 → 2 |

The gains are on the core questions: per-pool verdicts (A1), stuck vs full
(A3, A5, C2), the one-line blocker (A9), and asking for a cluster instead of
guessing (C1). For the Claude models, the B and D differences come from runs
in which the skill did not load (expected for those prompts), so they reflect
the agents' own behavior and run-to-run variance, not the skill.

## Live-Cluster Evaluation (Independent)

An evaluator agent that had not seen this dataset, its expected answers, or
this benchmark wrote its own 16 questions (9 in scope, 7 out of scope) before
reading the skill body. It seeded a live Run:ai cluster with test workloads: a
full GPU pool, jobs pending for capacity, one job pinned to a node type that
does not exist, and one job failing on purpose. It built ground truth with
read-only `kubectl`, `runai` CLI, and MCP calls, ran 16 questions × 2 arms ×
3 models × 2 attempts (192 runs) per revision through the real Run:ai MCP
server (read-only tools), and graded every answer blind. Snapshots before,
during, and after each round showed no cluster state change.

### Revision 3 (2026-10-11)

The worker node had restarted shortly before this round. The pool was full:
14 of 16 GPUs allocated to Run:ai workloads and 1 held by fractional pods the
MCP server does not list. Two jobs were pending for capacity. Workload events
still showed webhook and device admission errors from the restart that no
longer applied.

| Model | Pass, MCP-only → MCP+skill | In-scope pass | Out-of-scope pass |
| --- | --- | --- | --- |
| Claude Sonnet 5 | 23/32 → 27/32 | 14/18 → 18/18 | 9/14 → 9/14 |
| Claude Opus 4.8 | 23/32 → 29/32 | 11/18 → 17/18 | 12/14 → 12/14 |
| Codex GPT-5.6 Sol (+ AGENTS guidance) | 23/32 → 29/32 | 14/18 → 18/18 | 9/14 → 11/14 |

- **Paired difference (skill minus MCP-only, n = 96):** +0.35 overall (95% CI
  +0.19 to +0.52); +0.63 in scope (CI +0.41 to +0.87); 0.00 out of scope
  (CI −0.21 to +0.24).
- **Stuck vs full:** without the skill, agents concluded the scheduler was
  broken 3 times, each from the stale restart events; with the skill, never.
- **Revision 2 regressions:**
  - Route-outs fixed. Codex with the skill in context now answers
    out-of-scope read questions instead of refusing them (9/12 pass, 0/12 in
    revision 2).
  - Fragmentation wording fixed. No answer with the skill loaded claimed GPUs
    were fragmented across nodes.
  - Triggers improved. Sonnet loaded the skill on every GPU-count and
    saturation run, Opus on half. The pending backlog appeared in 5 of 6
    skill-arm saturation answers (2 of 6 in revision 2).
  - No write attempts with the skill in context.
  - `includeDetails`: the fallback recovered in all 11 runs where the server
    rejected the argument.
- **Judge cross-check:** an LLM judge agreed with the evaluator's blind grades
  on 88% of pass/fail decisions and showed the same direction (in scope 2.93
  without the skill, 3.44 with it).

### Revision 4 Spot Check (2026-10-11)

Five prompts, MCP+skill only, Claude Sonnet 5, against the same cluster
later the same day. By then 15 of 16 GPUs were allocated and the 16th was held
by fractional pods. The newest scheduler event of the job pinned to a missing
node type named only missing GPUs; its affinity mismatch was in an older
event.

| Prompt | Revision 3 (1 run) | Revision 4 (2 runs) | Revision 4 after review fixes |
| --- | --- | --- | --- |
| Health summary | Pinned job classed `CAPACITY`; restart-time device error offered as a possible bad GPU | Pinned job `SPEC`; device error called an earlier event (2/2) | Same (4/4) |
| Free GPUs, saturated? | Stranded GPU blamed on an unhealthy device | Pinned job `SPEC`; no GPU fault claimed (2/2) | Same in 3/4; one run classed the pinned job `CAPACITY` and cited the old device error as a possible bad GPU |
| GPUs offline or failing? | Correct, no stale-event claim | Correct, 4 calls (2/2) | Correct, 3–6 calls (3/3) |
| Team GPU waste (out of scope) | Skill not loaded; answered read-only | Same (2/2) | Same (3/3) |
| Cordon the GPU node (write) | Skill not loaded; no write, no MCP call | Same (2/2) | No write (3/3); one run made 2 read-only calls to find the node |

A first revision 4 draft kept the any-event rule only below the
classification table; one "failing GPUs" run then classed the pinned job
`CAPACITY`. Moving the rule into the `SPEC` row fixed it in both later runs.

The review fixes first rewrote the `SPEC` row to separate old and current
events. That draft classed the pinned job `CAPACITY` in 3 of 5 "free GPUs"
runs, while the unchanged row got 4 of 4 in a control run at the same time.
The final text keeps the row and states the precedence in the history rule.
The first two rows of the last column use the final text; the other three
used a draft that differs only in the `SPEC` row, which those prompts do not
reach.

### Revision 5 Spot Check (2026-10-14)

**What the revision 4 answer got wrong.** On 2026-10-14, after a node
restart, a revision 4 health answer on the same cluster classed the pinned
job `CAPACITY`:

- It guessed at a GPU held outside Run:ai accounting.
- The real cause is in the job's spec: it pins `nodeType: gpu-h200`, and the
  pool's only GPU node is a V100.
- The workload history shows the affinity mismatch on 10-08, 10-11 and
  10-12. Since 10-13 every scheduling event names only the GPU shortage, so
  none of the events `get_workload_status` returns showed the pin.

Revision 5 adds one `get_workload_spec` call when a not-enough-resources
workload asks for no more GPUs than are free. The workload is classed `SPEC`
only when the spec pins placement and an event proves the pin matches no
node. When the recent status events no longer show that, one
`get_workload_history` call reads the older events. A pin without such an
event stays `CAPACITY`, with the pin named as an unconfirmed cause: node
labels are not exposed, and the pinned nodes may simply be full.

A first draft classed any pin as `SPEC`. Review pointed out that this
misclassifies a job pinned to nodes that exist but are full, so the history
proof was added. The results below are for the final text.

**Mock.** The mock now mirrors these behaviors:

- The pinned job's status events name only the GPU shortage. Its spec holds
  the pin, and its history holds the earlier affinity mismatch.
- `get_cluster_metrics` returns cluster-wide series for any `nodePoolName`,
  as the real backend does. Its description still advertises pool scoping, as
  the real server's does.

The four cases about the pinned job (A1, A3, A8, C3), both arms, three
models:

| Model | MCP-only pass | MCP+skill pass | Pinned job classed `SPEC` (skill) |
| --- | ---: | ---: | --- |
| Codex GPT-5.6 Sol (+ AGENTS guidance) | 1/4 | 4/4 | 4/4, spec and history calls |
| Claude Sonnet 5 | 1/4 | 4/4 | 4/4, one spec and one history call each |
| Claude Opus 4.8 | 0/4 | 3/4 | 4/4, one spec and one history call each |

- **Codex A1:** the first run hit the 600 s harness timeout and was rerun.
- **Opus A8:** it missed an unrelated hand-off.
- **Untested path:** the mock has no pinned job without a mismatch event, so
  the "pin stays `CAPACITY`" branch is not exercised.
- **Pool scoping:** no skill-arm run passed `nodePoolName` to
  `get_cluster_metrics`. The MCP-only runs did on 1 to 3 cases per model.
- **A7:** the per-pool utilization case was checked on the first draft (5/5
  on every model with the skill). It is unaffected by the history change.

**Live**, against the same cluster with "Is the cluster healthy? Give me a
verdict per node pool." (2 Sonnet 5 runs, 1 Opus 4.8 run; read-only MCP
server without write tools):

- Every run loaded the skill and made one spec and one history call for the
  pinned job.
- Every run classed it `SPEC`, citing the history's affinity-mismatch events
  (10-08 to 10-12).
- Every run gave the cluster `DEGRADED`, with 2 `CAPACITY` and 1 `SPEC` in
  the default pool.
- No run scoped `get_cluster_metrics` to a pool.

### Revision 2 (2026-10-08)

The single GPU pool reported 4 free GPUs, but fractional pods outside Run:ai
accounting held one of them, so two 4-GPU jobs stayed pending for capacity.

| Model | Pass, MCP-only → MCP+skill | In-scope pass | Out-of-scope pass |
| --- | --- | --- | --- |
| Claude Sonnet 5 | 22/32 → 27/32 | 13/18 → 17/18 | 9/14 → 10/14 |
| Claude Opus 4.8 | 22/32 → 26/32 | 11/18 → 16/18 | 11/14 → 10/14 |
| Codex GPT-5.6 Sol (+ AGENTS guidance) | 18/32 → 20/32 | 9/18 → 18/18 | 9/14 → 2/14 |

- **Paired difference:** +0.34 overall (CI +0.09 to +0.59); +0.82 in scope
  (CI +0.54 to +1.09); −0.26 out of scope.
- **Stuck vs full:** without the skill, agents concluded "the scheduler is
  stuck" 5 times, sometimes recommending a restart or delete-and-resubmit; with
  the skill, never.
- **Regressions found, fixed in revision 3:** Codex, with the skill always in
  context, refused out-of-scope read questions because the owning skills were
  not installed; "fragmented across nodes" was repeated on a pool with one GPU
  node; Claude models under-triggered on GPU-count and saturation questions;
  one agent tried `kubectl cordon` (denied by permissions); an older MCP server
  build rejected `includeDetails`.

The two rounds used different cluster states, so absolute scores are not
comparable across revisions; the paired difference within each round is.

## CI-Style Run — SkillEvaluator Tier 3 (Incomplete)

`skillevaluator tier3` with agents `claude-code` (`azure/anthropic/claude-sonnet-5`)
and `codex` (`azure/openai/gpt-5.6-sol`) through NVIDIA inference, Docker,
`--n-attempts 1`, `--n-concurrent 4`, on revision 1 of the skill.

- **Wall-clock:** 61.7 minutes for 120 trials on one workstation.
- **Outcome: INCOMPLETE.** SkillEvaluator invalidates a run unless every trial
  scores. Claude Code scored 55/60: four agent installs failed on network
  errors and one trial hit the 300 s agent timeout. Codex scored 0/60: the
  in-container Codex did not reach the NVIDIA endpoint. The tool publishes no
  lift for an incomplete run.
- **Indicative only.** The per-dimension means below are computed from the
  per-trial `reward.json` files of the 25 cases that scored in both Claude Code
  arms, not by the tool:

| Dimension | Metric | Without skill | With skill | Lift |
| --- | --- | ---: | ---: | ---: |
| Security | `security` | 1.00 | 1.00 | +0.00 |
| Correctness | `accuracy` | 0.65 | 0.79 | +0.14 |
| Discoverability | `skill_execution` | 0.60 | 0.82 | +0.22 |
| Effectiveness | ½ `goal_accuracy` + ½ `behavior_check` | 0.64 | 0.76 | +0.12 |
| Efficiency | `skill_efficiency` | 0.60 | 0.81 | +0.21 |

## Quality Dimensions

| Dimension | Result |
| --- | --- |
| Security | No write was carried out in any run. For cordon requests, Claude models (with and without the skill available) and Codex without the skill tried shell commands; the harness denied Claude's `Bash` calls and ran Codex in a read-only sandbox. Codex with the skill made no shell attempts. Live (revision 3): no write attempt with the skill in context. SkillSpector analysis complete with one MEDIUM advisory. |
| Correctness | Mock, MCP+skill pass: Codex 26/30, Sonnet 22/30, Opus 19/30 (MCP-only 15, 14, 13). Live (revision 3), in-scope pass: 18/18, 18/18, 17/18 (MCP-only 14, 14, 11). |
| Discoverability | Positive, boundary, route-out, and out-of-scope cases on all three models; see Known weak spots for trigger gaps. |
| Effectiveness | Mock pass-rate gains: Codex +37 pp, Sonnet +27 pp, Opus +20 pp. Live (revision 3) paired gain +0.63 of 4 in scope, 0.00 out of scope. |
| Efficiency | Mock: with the skill, median MCP calls stay at 2–4 and median runtime rises by 2–5 s; the extra calls are the pending-workload sample and pool inventory. Live (revision 3), in-scope mean calls per run: Sonnet 4.3 → 7.4, Opus 5.3 → 6.7, Codex 6.3 → 9.2. |

## Grading

- **Judge.** `azure/openai/gpt-5.5` through NVIDIA inference, a model outside
  all three arms. Four points: routing, facts, behavior, safety. A run passes
  with at least 3 points including safety. The judge sees the case's
  `expected_output` and `expected_behavior`, the MCP call trace, and the answer.
  All records from all revisions were graded in one pass with the final
  dataset, so the tables compare like with like.
- **Dataset corrections during evaluation.** Route-out cases (B) changed to
  allow a short read-only answer when the owning skill is not installed;
  B4, D1, and D2 were aligned with the skill's rules; A7 no longer required a
  wait-time statement the prompt never asked for. After review, B3 and B5
  expect a hand-off when the owning skill is installed, and the judge is told
  which sibling skills each arm has; B3 to B5 were re-graded for every
  revision. Of the 15 grades that changed, most were on runs that loaded no
  skill and moved in both directions, which is judge variance. Every arm was
  re-graded after each change.
- **Route-out and out-of-scope cases.** The skill under test must not answer
  with a health summary. Only `runai-get-workload-health-summary` is installed,
  and only in the Claude arms, so there B3 and B4 expect a hand-off to it.
  Otherwise an agent may answer the narrow question with its own tools or
  decline.
- **Agents.** Claude Code loads the skill on demand from the project; the
  operator's other user-level skills, including
  `runai-get-workload-health-summary`, are present in both arms. Codex receives
  `SKILL.md` as `AGENTS.md`, so it is always in context; read its rows as
  "+ AGENTS guidance", not on-demand loading. Codex runs used
  `--sandbox read-only` with approvals disabled.
- **Retries.** A run that reached the 600 s harness timeout was retried once
  (three Codex runs in revision 2; none in revision 3).
- **Counting.** Calls and cycles come from the tool-call trace. Context tokens
  are input plus cache tokens for Claude and input tokens for Codex.

## Revision History

| Model | MCP+skill pass, revision 1 | Revision 2 | Revision 3 |
| --- | ---: | ---: | ---: |
| Codex GPT-5.6 Sol | 24/30 | 26/30 | 26/30 |
| Claude Sonnet 5 | 22/30 | 23/30 | 22/30 |
| Claude Opus 4.8 | 20/30 | 20/30 | 19/30 |

- **Revision 1 → 2** (from mock misses): trigger phrases for node health,
  failing GPUs, and `kubectl` requests; per-pool utilization calls; mandatory
  point-in-time and Ready-node-GPU statements; the failures line in pool-scoped
  answers; every stuck workload ID in the hand-off; narrow questions answered
  first.
- **Revision 2 → 3** (from the live-cluster evaluation): short read-only
  answers on route-outs when the owning skill is not installed; fallback when
  `includeDetails` is rejected; fragmentation wording tied to the number of GPU
  nodes, with the utilization-above-allocation clue on single-node pools;
  triggers for GPU counts and saturation, and the pending backlog in every
  saturation answer; a 7-day capacity trend instead of a flat history refusal;
  no `kubectl` or CLI commands, even with user approval.
- **Revision 3 → 4** (from the revision 3 live round and a spot check): a node
  selector or affinity mismatch in any event makes a pending workload `SPEC`,
  because a full pool hides it behind a GPU-shortage message; admission,
  device, and webhook errors older than the newest scheduling event are
  history, not a current GPU or scheduler fault. Review fixes in the same
  revision: such old errors never make a workload `SPEC`; an `UNCLASSIFIED`
  class gives every sampled workload a class; `nodePoolName` is no longer sent
  to `list_node_pools`, which rejects it.
- **Revision 4 → 5** (from a live answer on 2026-10-14):
  - A not-enough-resources workload whose request fits the free GPUs gets one
    `get_workload_spec` call. A `nodeType` or affinity pin makes it `SPEC`
    when an event, from the status or one `get_workload_history` call, shows
    the pin matches no node; otherwise it stays `CAPACITY`.
  - `get_cluster_metrics` is no longer scoped to a pool, because the current
    server ignores `nodePoolName` for allocation and utilization. Utilization
    per pool is routed to `runai-get-utilization-overview`.

On the mock, the three revisions are within about 2 passes of each other per
model, which is within run-to-run variance for one attempt per case. The
revision 3 changes target failures the live cluster exposed and the mock
exercises little.

## Known Weak Spots

- **C8, explicit `kubectl` request.** Both Claude models follow the user's
  instruction and never load the skill (score 0 of 4 in every revision).
  Codex, with the skill in context, declines `kubectl`.
- **D2 and D3, write requests.** The skill does not load for these, as
  intended, and Claude models then try or offer `kubectl`. This is agent
  behavior outside the skill.
- **B6, GPU errors on one node.** The Claude models answer from general
  knowledge without loading the skill; Codex routes correctly but adds
  unsupported hardware details from a web search.
- **Opus.** It tends to add speculative causes, which the judge scores under
  safety.
- **History questions (live).** Claude models did not load the skill for a
  "last week" question, and only 1 of 6 skill-arm runs used the 7-day capacity
  trend.
- **Stale events (live).** In the revision 3 round, answers in both arms read
  old device admission errors from a node restart as a current GPU fault.
  Revision 4 treats them as history; this is verified by spot checks only,
  not by a full round, and 1 of 8 final-text runs on the two pending-sample
  prompts still misread them.
- **`includeDetails` (live).** The MCP server build on the live cluster
  rejects the argument; the skill's fallback recovers, at the cost of one
  failed call.
- **Failure correlation (live).** Sonnet sometimes loads this skill for
  "several jobs failed, is it a node issue?" and answers with a health verdict
  instead of examining the failures.

## Limitations

- Not NVCARPS-certified; one attempt per case on the mock; LLM judge.
- Revision 4 was checked with 5 live prompts on one model, not re-run on the
  mock or in a full live round. Revision 5 was checked on the affected mock
  cases and with 3 live runs of one prompt; the full mock run and the live
  rounds measured earlier revisions on the earlier mock.
- Mock fault scenarios are synthetic; only the healthy cluster is a live
  recording. The live cluster has one GPU node with simulated GPUs, so
  multi-node fragmentation and real utilization were not exercised live.
- The CI-style Tier 3 run is incomplete and ran revision 1.

## Evidence

- Eval dataset: `evals/evals.json`
- Deterministic MCP fixture: `evals/files/mock_mcp.py`
- MCP configuration: `evals/environment/mcp_servers.toml`

The local harness, raw traces, full answers, per-trial grades, and the
live-cluster evaluation report are kept outside the published skill package.

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
