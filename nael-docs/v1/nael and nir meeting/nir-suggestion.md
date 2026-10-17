# Nir suggestion

```text
REQUEST PATH (Zero Added Latency — design goal, not a measured guarantee)

             ┌────────────────────────────────────┐
Incoming ───►│ ✅ EXISTS · EPP Scheduler [ref:4]   │───► Selected Pod
Request      │ Uses dynamic in-memory weights     │     score(p) = Σ w[j] × s[j,p]
             └─────────────────▲──────────────────┘     Choose the highest score
                               │                        when using max-score picker.
             Updates Scorer Weights Periodically
                               │
             ┌─────────────────┴──────────────────┐
             │ 🆕 NEW · Adaptive Configurator     │  g = 1 / (1 + exp(-k × (I-s0)))
             │ [ref:3]                            │  cap = Dmax - P × (Dmax-Cmax)
             │ Updates each configured profile    │  D = Dmin + g × (cap-Dmin)
             └────────▲────────────────▲──────────┘  A = 1-D
                      │                │             w'[j] = budget × w[j] / Σcategory w
            Imbalance │                │ Queued Fraction
            Signal I  │                │ Signal P
           ┌──────────┴────────┐  ┌────┴──────────────────┐
           │ 🆕 NEW · Imbalance│  │ 🆕 NEW · Pressure     │
           │ Detector [ref:1]  │  │ Detector [ref:2]      │
           └──────────▲────────┘  └──────────▲───────────┘
           rawI = min(CVn(KV),       rawP = Σwaiting / Σ(waiting+running)
                      CVn(assigned)) P = 0.1 × previousP + 0.9 × rawP
           I = 0.1 × previousI
               + 0.9 × rawI
                      │                      │
                      └───────────┬──────────┘
                                  │ Reads Pool Metrics
                       .──────────┴───────────.
                      ( ✅ EXISTS · Endpoint   )
                      ( Metrics Store          )
                       '──────────────────────'
```

Background detectors measure the pool; the configurator retunes scorer weights while requests use the scheduler's current weights. The math beside each box describes the inspected experimental implementation.

**Legend:** ✅ EXISTS = baseline machinery; 🆕 NEW = added in Nir's experimental branch, not a claim of upstream merge; 🔜 PLANNED = future work. Boxes are components; the cylinder is stored metrics. The upward background flow follows the meeting diagram supplied by Nael.

**Pseudocode**

```text
1. Periodically read KV usage and waiting/running counts for the pool's endpoints.
2. Measure how uneven the load is and what fraction of requests are waiting.
3. Smooth both signals using their previous values.
4. Calculate how much weight to give distribution and affinity scorers.
5. Update scorer weights while preserving proportions within each category.
6. For each request, score candidate pods using the current weights and run the picker.
```

## TL;DR (the whole thing in 3 lines)

- Adapt **weights in background cycles** using pool-level signals; the request does not run the detector calculations.
- Imbalance shifts the weight budget toward distribution; a configured pressure cap can limit that shift.
- Core code exists in [Nir's in-progress branch](https://github.com/nirrozenbaum/gateway-api-inference-extension/tree/adaptive-with-pressure); no runtime or latency tests were run for these meeting notes.

## What the symbols mean

| Symbol | Meaning |
|---|---|
| `assigned` | Waiting plus running requests on each endpoint |
| `CVn(x)` | `min(1, (standard deviation(x) / mean(x)) / sqrt(N-1))`; returns 0 for a nonpositive mean. The implementation uses Gonum's `stat.MeanVariance`, with sample variance. |
| `I`, `P` | Smoothed imbalance and pressure signals |
| `D`, `A` | Total distribution and affinity weight budgets; `D + A = 1` |
| `Dmin`, `Dmax` | Minimum distribution budget and its unconstrained maximum |
| `Cmax` | Conservative distribution maximum used when pressure is high |
| `s0`, `k` | Sigmoid midpoint and steepness |
| `w[j]`, `s[j,p]` | Scorer j's original weight and its score for pod p |

The imbalance code normalizes each endpoint's assigned requests by the pool total before calculating their coefficient of variation. That scaling does not change CV. It takes the **minimum** of request dispersion and KV-usage dispersion, so both must indicate imbalance for a positive raw imbalance signal.

Both detectors emit raw zero below **5 total assigned requests**; the imbalance detector also emits zero with at most one endpoint. A raw zero still decays a previously nonzero smoothed signal rather than immediately clearing it. Constructor signals start at zero. These examples assume valid, nonnegative metrics and at least two endpoints.

## Defaults and the pressure caveat

| Setting | Inspected runner default |
|---|---:|
| Enable adaptive configurator | false |
| `Dmin` | 0.2 |
| `Dmax` | 1.0 |
| `Cmax` | 1.0 |
| `s0` | 0.6 |
| `k` | 6 |

With defaults, `cap = 1 - P × (1-1) = 1`: **pressure is measured but does not change the weights**. A lower `Cmax` activates that constraint. The later example uses `Cmax = 0.6` as an explicit illustrative override, not as a reported benchmark setting.

The constructor starts at `D = 0.2`, `A = 0.8`. After a configurator cycle with `I = 0`, the sigmoid produces `D ≈ 0.2213`, not exactly 0.2.

## Example 1 Balanced endpoints preserve affinity

Assume three endpoints, with each detector's previous signal zero:

| Endpoint | Waiting | Running | Assigned | KV utilization |
|---|---:|---:|---:|---:|
| Pod A | 0 | 10 | 10 | 0.5 |
| Pod B | 0 | 10 | 10 | 0.5 |
| Pod C | 0 | 10 | 10 | 0.5 |

Both vectors are uniform, so `rawI = 0`, and no requests are queued, so `rawP = 0`. Smoothed `I = 0`, `P = 0`.

```text
g = 1 / (1 + exp(-6 × (0-0.6))) = 0.026597
D = 0.2 + 0.026597 × (1-0.2)    = 0.221278
A = 1-D                       = 0.778722
```

**Result:** roughly 22.13% of the budget goes to distribution and 77.87% to affinity after the update.

## Example 2 A hotspot increases distribution

Assume each previous signal is zero and the configurator reads both detectors after their first update:

| Endpoint | Waiting | Running | Assigned | KV utilization |
|---|---:|---:|---:|---:|
| Pod A | 0 | 0 | 0 | 0 |
| Pod B | 0 | 0 | 0 | 0 |
| Pod C | 20 | 10 | 30 | 0.9 |

Both normalized dispersion values cap at 1: `rawI = min(1,1) = 1`. The queued fraction is `rawP = 20/30 = 2/3`.

```text
I = 0.1 × 0 + 0.9 × 1   = 0.9
P = 0.1 × 0 + 0.9 × 2/3 = 0.6
g = 1 / (1 + exp(-6 × (0.9-0.6))) = 0.858149
cap = 1                 (default Cmax equals Dmax)
D = 0.2 + 0.858149 × 0.8 = 0.886519
A = 1-D                 = 0.113481
```

**Result:** roughly 88.65% distribution and 11.35% affinity. This is an illustrative snapshot; independent background loops can observe different update timing.

**A useful boundary case:** keep the request counts above but give every endpoint KV utilization `0.5`. KV dispersion becomes zero, so `rawI = min(request dispersion, 0) = 0`. With previous signals zero and default pressure settings, the distribution budget remains approximately 0.2213 despite the request hotspot.

## Example 3 Pressure can limit the distribution increase

Use Example 2's `I = 0.9`, `P = 0.6`, but explicitly set **`Cmax = 0.6`**:

```text
cap = 1 - 0.6 × (1-0.6)     = 0.76
D = 0.2 + 0.858149 × (0.76-0.2) = 0.680563
A = 1-D                     = 0.319437
```

| Same measured signals | Distribution | Affinity |
|---|---:|---:|
| Default `Cmax = 1` | 0.8865 | 0.1135 |
| Override `Cmax = 0.6` | 0.6806 | 0.3194 |

**Result:** the configured pressure cap preserves more affinity even though the imbalance signal is unchanged.

## Example 4 From category budgets to a selected pod

Assume one affinity scorer, prefix, with original weight 2, and two distribution scorers, KV and queue, with original weights 1 each. This is an illustrative profile, not a recovered benchmark configuration.

| Scenario | Prefix weight | KV weight | Queue weight |
|---|---:|---:|---:|
| Example 1 | 0.778722 | 0.110639 | 0.110639 |
| Example 2 | 0.113481 | 0.443260 | 0.443260 |
| Example 3 | 0.319437 | 0.340282 | 0.340282 |

For a separate illustrative request, hold plugin scores fixed to isolate the effect of weights:

| Eligible pod | Prefix score | KV score | Queue score |
|---|---:|---:|---:|
| Cache-favored pod | 1 | 0 | 0 |
| Distribution-favored pod | 0 | 1 | 1 |

The first pod's total is `A`; the second's is `D`. With max-score selection, Example 1 chooses the cache-favored pod, while Examples 2 and 3 choose the distribution-favored pod. Real plugin scores must be computed for the actual request and candidates; these are not outputs derived from the earlier metric tables.

## What exists and what remains open

| Status | Scope |
|---|---|
| Implemented in experimental branch | Imbalance and pressure detectors, smoothing, periodic weight updates, metrics, and runner wiring |
| Merged prerequisites | [Scorer categories #2119](https://github.com/kubernetes-sigs/gateway-api-inference-extension/pull/2119) and [float weights #2207](https://github.com/kubernetes-sigs/gateway-api-inference-extension/pull/2207) |
| 🔜 PLANNED | Burst detection and automatic max-score/weighted-random picker switching, according to the supplied document |
| Not established by this inspection | End-to-end latency impact, benchmark reproduction, full upstream integration, and the exact Dynamo continuation |

Configuration validation remains a TODO in the inspected constructor. Explicit timestamp checks and dwell/hysteresis controls were not found in the detector/configurator functions inspected. The document reports initial benchmarks, but raw results were not supplied.

## Code references

Snapshot: 5 October 2026. Source links pin revision `6c0f8d8496b56d68c30369520f3c4477adf8cd42`.

| Ref | Source |
|---|---|
| 1 | [ImbalanceDetector](https://github.com/nirrozenbaum/gateway-api-inference-extension/blob/6c0f8d8496b56d68c30369520f3c4477adf8cd42/pkg/epp/asyncdetector/imbalance_detector.go) |
| 2 | [PressureDetector](https://github.com/nirrozenbaum/gateway-api-inference-extension/blob/6c0f8d8496b56d68c30369520f3c4477adf8cd42/pkg/epp/asyncdetector/pressure_detector.go) |
| 3 | [AdaptiveConfigurator](https://github.com/nirrozenbaum/gateway-api-inference-extension/blob/6c0f8d8496b56d68c30369520f3c4477adf8cd42/pkg/epp/scheduling/adaptive_configurator.go) |
| 4 | [Runner and defaults](https://github.com/nirrozenbaum/gateway-api-inference-extension/blob/6c0f8d8496b56d68c30369520f3c4477adf8cd42/cmd/epp/runner/runner.go) |
| 5 | [Signal constants](https://github.com/nirrozenbaum/gateway-api-inference-extension/blob/6c0f8d8496b56d68c30369520f3c4477adf8cd42/pkg/epp/asyncdetector/consts.go) |

Full background: [Nir already working in progress](<../Nir already working in progress.md>). [Original design document](https://docs.google.com/document/d/1NFfIz-BdUUOJrfErD2pfCizgncsWOYgFak8Z_HaP7e4/edit), authored by Nir Rozenbaum and Maroon Ayoub, with Kellen Swain as contributor; read through the supplied text.

## 🔧 If you only do one thing

Ask Nir which branch revision and parameter values produced his benchmark results, especially whether the pressure cap was enabled.

## 🧠 Remember it as

Background measurements periodically retune the weights used by incoming requests.
