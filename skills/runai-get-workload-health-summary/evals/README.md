# Local evaluation — runai-get-workload-health-summary

The certified `BENCHMARK.md` next to `SKILL.md` is generated and signed by
NVSkills CI, so it must not be edited by hand. This file keeps the local runs.

## Status

The three-model pilot is complete: Codex GPT-5.6 Sol, Claude Sonnet 5, and
Claude Opus 4.8 each ran 30 MCP-only and 30 MCP+skill trials. This benchmark
uses deterministic MCP replay because the supplied live `tal1` service account
authenticates but sees zero Run:ai projects and workloads and cannot create a
project.

The replay data is an API-faithful synthetic fixture maintained against the
current MCP schemas. It is not the ticket's v2 live recording from a seeded
cluster; that recording remains pending access to a suitable seeded project.

The results below are the baseline that identified the routing, drill-down,
and overview-evidence problems. `SKILL.md` and the acceptance expectations were
updated afterward. Rerun the 30 cases on all three models before treating these
numbers as measurements of the revised skill.

## Results

| Model | Mode | Pass | Mean score / 4 | Context tokens median / p95 | Output tokens median / p95 | Calls median / p95 | Cycles median / p95 | Runtime median / p95 |
| --- | --- | ---: | ---: | ---: | ---: | ---: | ---: | ---: |
| Codex GPT-5.6 Sol | MCP-only | 12/30 | 2.20 | 12,116 / 78,509 | 144 / 1,137 | 0 / 7 | 0 / 3 | 9.3 / 35.2 s |
| Codex GPT-5.6 Sol | MCP+skill | 29/30 | 3.80 | 83,385 / 157,759 | 833 / 1,511 | 1.5 / 10 | 1.5 / 5 | 27.1 / 40.0 s |
| Claude Sonnet 5 | MCP-only | 10/30 | 2.07 | 20,728 / 32,826 | 283 / 928 | 1 / 3 | 1 / 2 | 7.1 / 12.6 s |
| Claude Sonnet 5 | MCP+skill | 20/30 | 2.97 | 26,909 / 101,500 | 513 / 2,542 | 1 / 7.5 | 1 / 3 | 8.7 / 27.7 s |
| Claude Opus 4.8 | MCP-only | 13/30 | 2.23 | 9,631 / 26,922 | 737 / 1,878 | 1.5 / 5.5 | 1 / 4 | 13.4 / 29.3 s |
| Claude Opus 4.8 | MCP+skill | 25/30 | 3.50 | 30,529 / 70,209 | 1,175 / 3,717 | 2 / 9 | 1.5 / 3 | 19.5 / 50.5 s |

In this baseline, skill impact was +17 passes for Codex, +10 for Sonnet, and
+12 for Opus. Codex had no score regressions; Sonnet had one minor regression
on the restart-sort boundary; Opus had none. The degraded-idle root-cause
hand-off failed for all three models; the revised skill now routes that request
before MCP, but the improvement has not yet been remeasured.

## Quality Dimensions

| Dimension | Result |
| --- | --- |
| Security | Pass: read-only fixture, mutation refusals, no stored credentials, static security checks pass |
| Correctness | Pre-revision baseline only: MCP+skill Codex 29/30, Sonnet 20/30, Opus 25/30 |
| Discoverability | Positive, negative, and cross-skill boundaries exercised on all three models |
| Effectiveness | Pass-rate gains: Codex +56.7 pp, Sonnet +33.3 pp, Opus +40.0 pp |
| Efficiency | Tokens, cache, calls, cycles, runtime, answer size, and Claude cost measured |

## Evidence

- Eval dataset: `evals/evals.json`
- Deterministic synthetic MCP: `evals/files/mock_mcp.py`
- MCP configuration: `evals/environment/mcp_servers.toml`

Detailed local captures, live-cluster notes, comparison utilities, and review
presentations are intentionally excluded from the published skill package.
