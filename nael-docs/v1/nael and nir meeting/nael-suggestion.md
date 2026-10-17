# Nael suggestion

```text
REQUEST PATH (Adaptive profile selection for each request — proposed)

Incoming ───►┌───────────────────────────────────────────┐
Request      │ ✅ EXISTS · Locate and screen candidates   │
             └─────────────────────┬─────────────────────┘
                                   │ Pass candidate snapshots
             .─────────────────────▼─────────────────────.
            ( ✅ EXISTS · Candidate metrics on request     )
            ( WaitingQueueSize and UpdateTime [ref:1]       )
             '─────────────────────┬─────────────────────'
                                   │ Read in-memory queues and timestamps
             ┌─────────────────────▼─────────────────────┐
             │ 🆕 NEW · Adaptive handler: Produce [ref:1] │  meanQueue = Σ queue[i] / N
             │ Validate data and save request summary    │  valid: N > 0, recorded metrics,
             └─────────────────────┬─────────────────────┘         queue[i] >= 0,
                                   │ Publish complete summary     0 <= age[i] <= maxAge
             ┌─────────────────────▼─────────────────────┐
             │ ✅ EXISTS · Admission and scheduling      │
             └─────────────────────┬─────────────────────┘
                                   │ If admitted, call Pick
             ┌─────────────────────▼─────────────────────┐
             │ 🆕 NEW · Same handler: Pick [ref:2]        │  First matching rule wins:
             │ Recheck freshness; store one choice       │  meanQueue >= 20 → load-oriented
             └─────────────────────┬─────────────────────┘  meanQueue < 5   → cache-oriented
                                   │ Run selected profile   Otherwise / invalid → balanced
             ┌─────────────────────▼─────────────────────┐
             │ ✅ EXISTS · Profile execution [ref:3]     │  score(p) = Σ w[profile,j] × s[j,p]
             │ Filter → score → max-score picker         │  Pick pod with highest total.
             │ Uses the chosen profile's FIXED weights   │
             └─────────────────────┬─────────────────────┘
                                   │ Pass profile result
             ┌─────────────────────▼─────────────────────┐
             │ 🆕 NEW · Same handler: ProcessResults     │  Exactly one valid result for
             │ Return selected profile as primary       │  the stored profile is required.
             └─────────────────────┬─────────────────────┘
                                   └──────────────────────────► Selected Pod
```

One proposed plugin prepares the request's mean queue and chooses a profile; the existing scheduler executes that profile. The metrics cylinder represents in-memory snapshots, not a new database. Other producers are omitted for clarity.

**Legend:** ✅ EXISTS = existing mechanisms described by the proposal; 🆕 NEW = proposed additions, not implemented; 🔜 PLANNED = later work. Boxes are components or execution stages. All three adaptive-handler boxes are hooks on the **same instance**, not separate services.

**Pseudocode**

```text
1. Locate and screen the pods that could serve the request.
2. Check every candidate's queue value and metric timestamp, then calculate the mean.
3. Save one complete summary on the request and continue through admission.
4. If admitted, read the summary once and recheck its freshness.
5. Select the first matching profile, or the default when no rule matches or data is invalid.
6. Store the choice so later metric publication cannot change it.
7. Run that profile's filters, scorers, and picker to choose a pod.
8. Return the selected profile's result; a failed selection produces an error, without trying another profile.
```

## TL;DR (the whole thing in 3 lines)

- Choose a predefined scheduling profile **for each request**, based on mean waiting queue over that request's candidates.
- Profile weights remain fixed; the selected profile changes how the existing scorers influence pod selection.
- This is a draft proposal in [PR #1701](https://github.com/dsx-ai-factory/run-ai/pull/1701), with no adaptive plugin implementation or performance results.

## Selection math and profile configuration

The example uses `maxMetricAge = 1 second`, `defaultProfile = balanced`, and ordered rules: high queue first, then low queue. These are illustrative settings, not benchmark-tuned values.

| Condition at first Pick | Profile | Prefix weight | KV weight | Queue weight |
|---|---|---:|---:|---:|
| Valid mean queue >= 20 | load-oriented | 1 | 1 | 4 |
| Valid mean queue < 5 | cache-oriented | 4 | 1 | 1 |
| Valid mean queue from 5 to below 20 | balanced | 2 | 1 | 1 |
| Missing or invalid summary | balanced | 2 | 1 | 1 |

`meanQueue = (q1 + q2 + ... + qN) / N`, with no rounding before rule evaluation.

`age[i] = current time - recorded update time[i]`. Every candidate must have recorded, nonnegative queue data with `0 <= age[i] <= 1 second`; otherwise the entire summary is unavailable. Pick repeats the freshness check using the recorded timestamps. Zero candidates also means unavailable data. A recorded zero queue is valid; a missing queue is not zero.

All alternative profiles must share endpoint eligibility and inference stage, and the summarized candidates must satisfy those eligibility requirements. The MVP assumes homogeneous workers able to serve the full request.

## Example 1 Low load selects cache-oriented

Candidate queues are `0, 2, 4`, with fresh metrics:

```text
meanQueue = (0 + 2 + 4) / 3 = 2
High-queue rule: 2 >= 20? No.
Low-queue rule: 2 < 5? Yes.
Selected profile: cache-oriented.
Weights: prefix 4, KV 1, queue 1.
```

**Result:** prefix reuse gets the strongest weight for this request. The profile still filters and scores candidates; selecting this profile does not itself select a pod.

## Example 2 High load selects load-oriented

Candidate queues are `10, 20, 30`, with fresh metrics:

```text
meanQueue = (10 + 20 + 30) / 3 = 20
High-queue rule: 20 >= 20? Yes.
Selected profile: load-oriented.
Weights: prefix 1, KV 1, queue 4.
```

**Result:** the queue scorer has the strongest weight. No weights are rewritten; this request runs a different predefined profile.

## Example 3 From profile selection to a selected pod

Use these illustrative normalized plugin scores for two eligible pods. Higher scores are better. Hold the scores fixed to isolate the weight calculation; they are not measured outputs from Examples 1 and 2.

| Pod | Prefix score | KV score | Queue score |
|---|---:|---:|---:|
| Pod A | 1.0 | 0.2 | 0.0 |
| Pod B | 0.0 | 0.8 | 1.0 |

| Selected profile | Pod A weighted total | Pod B weighted total | Max-score result |
|---|---|---|---|
| cache-oriented | `4×1 + 1×0.2 + 1×0 = 4.2` | `4×0 + 1×0.8 + 1×1 = 1.8` | Pod A |
| balanced | `2×1 + 1×0.2 + 1×0 = 2.2` | `2×0 + 1×0.8 + 1×1 = 1.8` | Pod A |
| load-oriented | `1×1 + 1×0.2 + 4×0 = 1.2` | `1×0 + 1×0.8 + 4×1 = 4.8` | Pod B |

**Result:** the handler selects the profile, then the profile's weights determine which pod wins from the plugin scores.

## Example 4 The request candidate set matters

The pool has three pods with queues `0, 0, 30`.

| Request population | Calculation | Selection |
|---|---|---|
| All three pods are candidates | `(0+0+30)/3 = 10` | balanced |
| Screening leaves only the first two pods | `(0+0)/2 = 0` | cache-oriented |

The second row assumes the screened candidates and all alternative profiles satisfy the shared eligibility contract. The handler uses the candidates passed to Produce, not a separate pool-wide average.

**Limitation:** `0, 0, 30` and `10, 10, 10` both have mean 10. This proposal's mean alone does not distinguish their load distribution.

## Example 5 Boundaries and fallback

| Situation | Calculation or validation | Outcome |
|---|---|---|
| Queues `4, 5, 6`, all fresh | Mean 5; neither rule matches | balanced |
| Queues `19, 20, 21`, all fresh | Mean 20; high rule matches | load-oriented |
| One candidate has no recorded metrics | Full coverage fails | balanced |
| Queues `0, 2, 4`; one sample is 1.2 seconds old | Older than the 1-second limit | balanced |
| A sample is 0.9 seconds old during Produce, 1.1 seconds old during Pick | Freshness recheck fails | balanced |
| A timestamp is in the future, or a queue is negative | Summary invalid | balanced |
| Summary arrives after first Pick selected balanced | The request already owns a decision | Keep balanced |
| Selected profile cannot produce a valid endpoint | Profile execution fails | Scheduling error; no alternative-profile retry |

The default profile remains subject to admission, endpoint availability, and its own filters. Fallback selects a policy; it does not guarantee successful routing.

## Scope and implementation status

| Area | Proposal |
|---|---|
| New code | One plugin implementing DataProducer and ProfileHandler, plus registration, data dependencies, and parameter validation |
| Data access | Existing in-memory snapshots; no scraping or network calls inside Produce/Pick |
| State | Immutable summary and stored choice on each request; shared instance holds configuration |
| Configuration | Rules and profiles loaded at startup; no new CRD required by this design |
| Cost | Adds candidate aggregation and rule evaluation per request; latency overhead has not been measured |
| Outside MVP | Hot reload, automatic weight tuning, arbitrary metric expressions, shared active-profile state, hysteresis, and composition with another profile handler |
| Current status | Draft only; examples specify intended behavior and do not establish runtime correctness or performance |

Because there is no hysteresis, requests arriving around a threshold can choose different profiles. Thresholds and scorer weights require a workload and benchmarks before any benefit can be claimed.

## Code references

These references point to the proposal, since the adaptive handler does not exist yet. The proposal's baseline inspection was of local upstream code at `8a2f37d3` on 29 September 2026; this meeting note is dated 5 October 2026.

| Ref | Proposal section |
|---|---|
| 1 | [Reusing the data layer](../2026-09-28-adaptive-profile-handler-v2.html#7-preparing-load-signals) |
| 2 | [Selection contract](../2026-09-28-adaptive-profile-handler-v2.html#6-proposed-selection-contract) |
| 3 | [Profiles and example YAML](../2026-09-28-adaptive-profile-handler-v2.html#8-complete-proposed-yaml-example) |
| 4 | [Proposed flow](../2026-09-28-adaptive-profile-handler-v2.html#3-proposed-flow) |

Full proposal: [Adaptive Profile Handler v2](../2026-09-28-adaptive-profile-handler-v2.html). Review: [PR #1701](https://github.com/dsx-ai-factory/run-ai/pull/1701), proposal revision `81ec806ae6a3cf396f8591cad1c9729a0a811982`.

## 🔧 If you only do one thing

Choose a workload and measure whether selecting among profiles improves results over a fixed combined profile before implementing the adaptive handler.

## 🧠 Remember it as

The request chooses a fixed policy; that policy chooses the pod.
