# Upstream issue draft

Target repository: `llm-d/llm-d-router` (the EPP code moved here from GAIE; GAIE `main` keeps only the InferencePool API)
Template: feature request. Cross-link GAIE #1992, where the alternative was first discussed.
Status: **draft, not posted**. Review with Nir first. Everything below the line is the issue body.

Suggested title:

> Discrete scheduling-profile selection on normalized imbalance and pressure signals (the "select groups of scorers" alternative from GAIE #1992)

---

<!-- Please only use this template for submitting enhancement requests -->

**What would you like to be added**:

A `ProfileHandler` plugin that selects one of several operator-defined scheduling profiles **per request**, based on two normalized load signals computed over the request's candidate endpoints:

| Signal | Formula | Answers | Source metrics (already on `datalayer.Metrics`, `pkg/epp/framework/interface/datalayer/metrics.go`) |
|---|---|---|---|
| `imbalance` (`cv`) | `min(1, σ / mean)` of `WaitingQueueSize` across candidates; `0` when `mean < minMeanQueue` | How **uneven** is the load? | `WaitingQueueSize` |
| `pressure` (`P`) | `Σ waiting / Σ (waiting + running)` across candidates | How **much** of the work is waiting? | `WaitingQueueSize`, `RunningRequestsSize` |

Rules are ordered, first match wins, and each rule carries an **enter** and an **exit** threshold so the selection does not flip on scrape jitter. A mandatory default profile is used when no rule matches or candidate metrics are missing or stale.

Proposed configuration (names and values illustrative):

```yaml
apiVersion: llm-d.ai/v1
kind: EndpointPickerConfig
plugins:
- type: adaptive-profile-handler
  parameters:
    defaultProfile: balanced
    maxMetricAge: 1s
    minMeanQueue: 3        # gate: cv is 0 below this mean queue
    smoothing: 0.1         # α for x_s = α·x_s + (1−α)·x ; 0 disables
    rules:
    - name: hotspot        # uneven → spread the load
      when: { signal: imbalance, operator: gte, enter: 0.5, exit: 0.35 }
      profile: load-oriented
    - name: idle           # little waiting → exploit prefix cache
      when: { signal: pressure, operator: lt, enter: 0.2, exit: 0.3 }
      profile: cache-oriented
- type: decode-filter
- type: prefix-cache-scorer
- type: kv-cache-utilization-scorer
- type: queue-scorer
- type: max-score-picker
schedulingProfiles:
- name: cache-oriented
  plugins: [{pluginRef: decode-filter}, {pluginRef: prefix-cache-scorer, weight: 4}, {pluginRef: kv-cache-utilization-scorer, weight: 1}, {pluginRef: queue-scorer, weight: 1}, {pluginRef: max-score-picker}]
- name: balanced
  plugins: [{pluginRef: decode-filter}, {pluginRef: prefix-cache-scorer, weight: 2}, {pluginRef: kv-cache-utilization-scorer, weight: 1}, {pluginRef: queue-scorer, weight: 1}, {pluginRef: max-score-picker}]
- name: load-oriented
  plugins: [{pluginRef: decode-filter}, {pluginRef: prefix-cache-scorer, weight: 1}, {pluginRef: kv-cache-utilization-scorer, weight: 1}, {pluginRef: queue-scorer, weight: 4}, {pluginRef: max-score-picker}]
```

How it fits the existing extension points:

```text
✅ Locate and screen candidates
        │  per-endpoint snapshots: WaitingQueueSize · RunningRequestsSize · UpdateTime
        ▼
🆕 DataProducer.Produce    one pass over candidates → {mean, cv, P, count, expiresAt} on the request
        │
        ▼
✅ Admission
        │
        ▼
🆕 ProfileHandler.Pick     regime check (enter/exit) → first matching rule → store choice   ◄──► 🆕 shared regime state
        │                                                                                     (plugin instance, mutex)
        ▼
✅ Profile execution       filters → weighted scorers → picker, fixed weights per profile
        │
        ▼
🆕 ProcessResults          mark the selected profile primary; error if it produced no result
```

- One plugin instance implements both `DataProducer` and `ProfileHandler`; it declares `WaitingQueueSizeKey` and `RunningRequestsSizeKey` as data dependencies exactly like `queue-scorer`.
- No new CRD. Rules live in plugin parameters of `EndpointPickerConfig`.
- Metrics are read from the in-memory snapshots via `GetMetrics()`. No scraping or network calls in `Produce` or `Pick`.
- Selection is logged once per request as `rule <name> entered|held|exited at <signal>=<value>`, and counted as `adaptive_profile_selections_total{profile, rule, reason}`.

**Why is this needed**:

Today the default profile (`queue-scorer` 2, `kv-cache-utilization-scorer` 2, `prefix-cache-scorer` 3 in `routerlib/templates/_config.yaml`) has fixed weights. Scores react to live metrics, weights do not, so under a hotspot the prefix-cache scorer keeps concentrating traffic on the hot endpoint. #1992 asks for adaptation and lists two directions: adjust each scorer's weight continuously (the approach in the `adaptive-routing` / `adaptive-with-pressure` branches, with prerequisites #2119 and #2207 already merged), or **select among groups of distributing vs concentrating scorers**. This issue proposes the second direction, built on the same signals as the first.

Why these signals and not mean queue depth (the obvious first choice):

| Candidate queues | mean | cv | Decision on mean | Decision on cv |
|---|---|---|---|---|
| `10, 10, 10` | 10 | 0.00 | balanced | balanced ✔ |
| `0, 0, 30` | 10 | 1.00 | balanced ✘ misses the hot pod | load-oriented ✔ |
| `18, 20, 22` | 20 | 0.10 | load-oriented ✘ even load, queue scorer cannot help | balanced ✔ |
| `5, 20, 35` | 20 | 0.75 | load-oriented | load-oriented ✔ |

1. **The mean cannot see unevenness**, and unevenness is the only reason to trade affinity for distribution. The coefficient of variation can.
2. **A raw count is not portable.** "20 waiting" is ~8% of capacity for an 8B model on one GPU and ~60% for a 70B model on eight. `cv` and `P` are ratios in `[0, 1]`, so one threshold works across models and hardware.
3. **A single threshold flips on jitter.** With `cv ≥ 0.5` alone, a sequence `0.48, 0.52, 0.47, 0.55, 0.40` switches profile four times and wastes the shared prefix cache. With `enter 0.5 / exit 0.35` the same sequence enters once and exits once.

Why discrete selection in addition to continuous weight adaptation:

- **Operator control and explainability.** Operators author the profiles; the system only chooses among them. "Request X ran `load-oriented` because rule `hotspot` entered at cv 0.52" is easier to reason about and bound than a derived weight vector.
- **Request scope.** Pool-wide averages can mislead after filtering: queues `0, 0, 30` have pool mean 10, but if screening leaves only the first two pods the candidate signal is 0. Computing over candidates keeps the decision consistent with what the profile will actually score.
- **Same vocabulary.** `P` is identical to the pressure detector's statistic; `cv` is the imbalance detector's statistic scoped to candidates. If selection later moves into the configurator (see open question 1), the rules and profiles carry over and only the signal source changes.

**Relation to existing work**:

- GAIE #1992 — dynamic adaptation of scorers/weights; this is the "select groups of scorers" alternative discussed there.
- GAIE #2119 (merged) — scorer preference categories (affinity vs distribution).
- GAIE #2207 (merged) — float scorer weights.
- `nirrozenbaum/gateway-api-inference-extension` branch `adaptive-with-pressure` — imbalance and pressure detectors plus a background configurator, never merged; this proposal reuses its statistics and smoothing constant.
- Existing profile handlers on `main`: `single`, `headerprofile`, `disagg`, `dataparallel` under `pkg/epp/framework/plugins/scheduling/profilehandler/`; the proposed `adaptive` handler sits beside them.

**Open questions**:

1. **Placement.** The loader allows one `ProfileHandler`, so a dedicated handler cannot coexist with `disagg-profile-handler`. Alternative: a `mode: discrete` in the adaptive configurator that picks one of N predefined weight sets and applies it to every profile. That is pool-scoped and runs in the background, but keeps the handler seat free. Which do maintainers prefer?
2. **Normalization of cv.** Plain `σ / mean` capped at 1, or the detector's `(σ / mean) / √(N−1)`? Only the threshold scale changes.
3. **Metric coverage.** Require fresh metrics on all candidates (strict, falls back often during rollouts) or on ≥ 80% (lenient, matches `PodsWithFreshMetrics` behavior in the pool metrics logger)?
4. **Regime state scope.** One regime per plugin instance (one pool per EPP) or per candidate set?

**Validation plan**:

Prototype on a branch of `main`, deploy with the `llm-d-router-gateway` chart on 4 GPUs with a real model and prefix caching, and run a shared-prefix hotspot workload (GuideLLM, `prefix_tokens=1024`, concurrency 8–96) against the default profile and against the adaptive handler. Report TTFT p50/p99, output tokens/s per GPU, per-pod queue spread, prefix-cache hit rate, and profile switches per minute. A PR follows only if the handler lowers TTFT p99 and raises tokens/s per GPU on that workload.

**Status**: design only. No implementation, no benchmark results yet. All thresholds, weights and the gate value above are illustrative.
