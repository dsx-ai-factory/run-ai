# Local evaluation — runai-diagnose-failed-workload

The certified `BENCHMARK.md` next to `SKILL.md` is generated and signed by
NVSkills CI, so it must not be edited by hand. This file keeps the local runs.

Preliminary local evaluation under the NV-BASE `external` profile. These
results are not NVCARPS-certified.

Static validation was refreshed on 2026-07-27 with NV-BASE `3.7.1` and Astra
Skill Eval `0.8.5`. The model-backed attempts below were run on 2026-07-26
with Docker client/server `27.4.1`/`27.4.0`.

## Tier 1 — Static Validation

`nv-base validate --external --no-dedup --no-llm` passed with exit 0 in
0.40 seconds:

- Overall validation passed; quality score 96.2/100 (A).
- Completed checks reported no CRITICAL, HIGH, or MEDIUM findings and four LOW
  advisories.
- PII, licensing, code-integrity, and Unicode checks passed. The optional
  SkillSpector security scan did not run because SkillSpector is not installed.
- Astra evaluation-structure validation passed 5/5 checks.

## Tier 2 / 3 — Model-Backed Evaluation

| Run | Exit | Time | Result |
| --- | ---: | ---: | --- |
| Full validation | 1 | 0.42 s | Tier 1 passed; Tier 2 embedding did not complete |
| Claude Code, Docker | 1 | 2.78 s | No trials started |
| Codex, Docker | 1 | 2.85 s | No trials started |

Full validation stopped at NV-BASE Tier 2 because `NVIDIA_INFERENCE_KEY` was
not configured. The agent runs stopped because neither `NVIDIA_INFERENCE_KEY`
nor `NVIDIA_API_KEY` was configured. The attempted commands took 6.05 seconds;
the runtime of the current 104-trial matrix is unknown.

The matrix contains 26 cases × 2 arms × 2 agents. Both arms use the same
network-free MCP mock and receive identical fixture responses.

## Evaluation Coverage

The dataset contains 12 trigger/routing cases and 14 deterministic MCP workflow
cases.

| Coverage | Cases |
| --- | --- |
| Failure categories | 8/8, including `UNKNOWN` |
| Routing boundaries | Scheduling-only Pending and preempted |
| Recurrence | Recurring OOM |
| Name resolution | Terminal and active CrashLoopBackOff name/project lookup |
| Active failure | Running workload with confirmed CrashLoopBackOff |
| Startup failure | Pending workload with a direct image, init, or mount failure |
| MCP evidence | Status, submitted spec, pods, history/events, and metrics |

## Quality Dimensions

| Dimension | Result | Evidence |
| --- | --- | --- |
| Security | Partial (static) | PII passed; optional SkillSpector scan was not run |
| Correctness | 100/100 (static) | No static correctness findings; agent accuracy not measured |
| Discoverability | 95/100 (static) | Positive and negative trigger cases are defined; trigger accuracy not measured |
| Effectiveness | Not measured | Baseline and with-skill trials did not start |
| Efficiency | 100/100 (static) | `SKILL.md` has 3,499 tokens; agent usage not measured |

Trigger accuracy, false-trigger rate, category accuracy, behavior completion,
skill lift, agent tokens, MCP calls, and per-trial runtime remain unmeasured.

## Limitations

- MCP responses are sanitized synthetic fixtures, not recordings from a seeded
  live cluster.
- The optional SkillSpector security scan was not run locally.
- Model-backed results require a successful Claude Code and Codex run.
- NVCARPS certification, signing, skill-card generation, and publication remain
  deferred.

## Reproduce

NV-BASE Tier 2 requires `NVIDIA_INFERENCE_KEY`. Astra agent evaluation accepts
`NVIDIA_INFERENCE_KEY` or `NVIDIA_API_KEY`. See the NVCARPS onboarding guide,
and do not commit either key.

```shell
nv-base validate skills/runai-diagnose-failed-workload \
  --external --no-dedup --no-llm

nv-base validate skills/runai-diagnose-failed-workload \
  --external --no-dedup --agent-eval \
  --agents claude-code --env-mode docker
```

Repeat the agent command with `--agents codex`. Replace unmeasured values only
with results from successful runs.
