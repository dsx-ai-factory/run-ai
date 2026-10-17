# 002 · Run: the hotspot benchmark, baseline vs our handler, one set of numbers

## TL;DR (the whole thing in 3 lines)

- One workload that **creates a hotspot on purpose**: every request shares a long prefix, so the prefix scorer (weight 3 in the default profile) pulls them all to one pod.
- Two runs with only the EPP changed: **A · default profile on `main`** and **C · our adaptive profile handler**. Same pods, same model, same load, same seed.
- The claim to prove: C **lowers TTFT p99 and raises served tokens/s per GPU** on this workload, while keeping the prefix-cache hit rate close to A. If the numbers do not say that, there is nothing to submit.

Status: run plan, written 2026-10-12. Run A can be executed today. Run C needs the prototype.

## The hypothesis, written down before running anything

| | Statement |
|---|---|
| **H1 · latency** | At concurrency ≥ 32 on the shared-prefix workload, C has lower TTFT p99 than A. |
| **H2 · efficiency** | At the same concurrency, C serves more output tokens/s per GPU than A (the three cold GPUs do work). |
| **H3 · cost of H1/H2** | C's prefix-cache hit rate stays within 20 points of A at the concurrency where A is not yet hot (8), and is allowed to drop where A is saturated. |
| **H4 · stability** | C switches profile fewer than 3 times per minute at steady load (hysteresis works). |
| Kill criterion | If H1 and H2 both fail at every concurrency, the discrete handler adds nothing over the default profile on this workload. |

## The runs

```text
                 ┌────────────────────────────────────────────────────────────┐
  same setup     │  4 × vLLM pods · 1 model · prefix caching on (doc 001)      │
  (doc 001)      └────────────────────────────┬───────────────────────────────┘
                                              │
                          ┌───────────────────┴───────────────────┐
                          ▼                                       ▼
            ┌──────────────────────────┐            ┌──────────────────────────┐
            │ A · default profile       │            │ C · adaptive handler      │
            │ queue 2 · KV 2 · prefix 3 │            │ cache / balanced / load   │
            │ release image v0.11.0     │            │ rules on cv and P,        │
            │ ✅ runs today             │            │ enter/exit bounds         │
            └──────────────┬───────────┘            │ 🔜 needs prototype        │
                           │                        └──────────────┬───────────┘
                           └────────────────┬───────────────────────┘
                                            ▼
                  same hotspot workload · same concurrency sweep · 3 repetitions each
                                            ▼
                  TTFT p50 / p99 · tokens/s per GPU · queue spread · hit rate · switches (C)
```


## 1. The workload

**Goal:** make the prefix scorer concentrate traffic. All requests carry the same 1024-token prefix. Under the default profile the EPP keeps sending them to the pod that already holds the prefix; its queue grows while three pods idle. That is the hotspot.

GuideLLM is the load tool. Deploy it as a long-running pod and `exec` in.

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: guidellm
  namespace: epp-lab
spec:
  containers:
  - name: guidellm
    image: ghcr.io/vllm-project/guidellm:v0.8.1      # check: latest release at writing time
    command: ["/bin/bash", "-c", "sleep infinity"]
    resources:
      requests: { cpu: "2", memory: 4Gi }
    volumeMounts:
    - { mountPath: /data, name: results }
  volumes:
  - name: results
    persistentVolumeClaim: { claimName: guidellm-results }
---
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: guidellm-results
  namespace: epp-lab
spec:
  accessModes: [ReadWriteOnce]
  resources: { requests: { storage: 5Gi } }
```

```bash
kubectl apply -f guidellm.yaml
kubectl exec -it guidellm -- bash
```

One sweep:

```bash
GATEWAY=http://$(kubectl get gateway/inference-gateway -o jsonpath='{.status.addresses[0].value}'):80
MODEL=Qwen/Qwen3-8B

guidellm benchmark \
  --target "$GATEWAY" \
  --model "$MODEL" \
  --processor "$MODEL" \
  --rate-type concurrent \
  --rate 8,16,32,64,96 \
  --max-seconds 120 \
  --data "prefix_tokens=1024,prompt_tokens=128,output_tokens=128" \
  --output-path /data/run-A-rep1.json
```

| Argument | Why this value |
|---|---|
| `--rate-type concurrent --rate 8,…,96` | Concurrency sweep. With `max-num-seq 256` per pod the only way a queue forms below 96 is a hotspot, which is the point. |
| `prefix_tokens=1024` | The shared prefix. Long enough that a cache hit is worth a lot to the prefix scorer. Documented in GuideLLM `docs/en/guides/datasets.md` (synthetic data). |
| `prompt_tokens=128, output_tokens=128` | Short unique tail and short generation, so the experiment is about routing, not decode time. |
| `--max-seconds 120` | Long enough to reach steady state at each level. |

Run each sweep **3 times** per run type and report the median. Name files `run-<A|C>-rep<n>.json`.

## 2. Switching between A and C

| | Command (from the `llm-d-router` checkout, doc 001) |
|---|---|
| **A** | `helm upgrade $POOL ./config/charts/llm-d-router-gateway --reuse-values --set router.epp.image.registry=ghcr.io/llm-d --set router.epp.image.tag=v0.11.0 --set router.epp.pluginsConfigFile=default-plugins.yaml` |
| **C** | doc 001 step 5: our image + `--set router.epp.pluginsConfigFile=adaptive-profiles.yaml` |

After every switch: `kubectl rollout status deploy/$POOL-epp`, then **wait 60 s** so the EPP has scraped every pod a few times and the vLLM prefix caches start from a comparable state, then start the sweep.

Cache policy: do **not** restart the vLLM pods between repetitions. Warm caches are the realistic case. Write this in the results log.

## 3. What to record

### From GuideLLM (per run file)

| Metric | Where | Shows |
|---|---|---|
| TTFT p50 / p99 per concurrency | `benchmarks[].metrics.time_to_first_token_ms` | H1 |
| Output tokens/s per concurrency | `benchmarks[].metrics.output_tokens_per_second` | H2 (divide by 4 for per GPU) |
| Request latency p99 | `benchmarks[].metrics.request_latency` | Tail behavior |
| Errors / timeouts | error totals | Must be 0 for a run to count |

### From Prometheus (PromQL over each run's window)

| Name | Query | Shows |
|---|---|---|
| Queue per pod | `vllm:num_requests_waiting{namespace="epp-lab"}` | The hotspot: one line up, three flat |
| **Queue spread** | `max(vllm:num_requests_waiting{namespace="epp-lab"}) - min(vllm:num_requests_waiting{namespace="epp-lab"})` | Large in A, small in C |
| KV usage per pod | `vllm:gpu_cache_usage_perc{namespace="epp-lab"}` | Whether the hot pod is also memory-hot |
| Prefix hit rate per pod (vLLM side) | `rate(vllm:prefix_cache_hits_total[30s]) / rate(vllm:prefix_cache_queries_total[30s])` | H3. **check** names for vLLM v0.21.0 |
| Prefix prediction (EPP side) | `llm_d_router_endpoint_picker_prefix_predicted_cached_tokens`, `…_prefix_prompt_tokens` | What the EPP believed about cache reuse when it chose |
| Pool averages (EPP) | `inference_pool_average_queue_size`, `inference_pool_average_kv_cache_utilization` | Sanity. **check** these names still exist on `main` |
| Our decisions (C) | `adaptive_profile_selections_total{profile,rule,reason}` 🔜 | H4: selections and regime switches |

Export each query for the run window:

```bash
PROM=http://localhost:9090   # kubectl port-forward -n monitoring svc/kube-prometheus-stack-prometheus 9090
START=<unix>; END=<unix>
curl -sG "$PROM/api/v1/query_range" \
  --data-urlencode 'query=max(vllm:num_requests_waiting{namespace="epp-lab"}) - min(vllm:num_requests_waiting{namespace="epp-lab"})' \
  --data-urlencode "start=$START" --data-urlencode "end=$END" --data-urlencode "step=5" \
  > spread-run-A-rep1.json
```

## 4. How to present the result

One table, medians of 3 repetitions:

| Concurrency | A TTFT p99 | C TTFT p99 | Δ | A tok/s per GPU | C tok/s per GPU | Δ | A hit rate | C hit rate | A spread | C spread | C switches/min |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 8 | | | | | | | | | | | |
| 16 | | | | | | | | | | | |
| 32 | | | | | | | | | | | |
| 64 | | | | | | | | | | | |
| 96 | | | | | | | | | | | |

Three charts, A and C overlaid:

1. **Queue per pod over time at concurrency 64.** The picture: A shows one line climbing, C shows four lines close together.
2. **TTFT p99 and tokens/s per GPU vs concurrency.** H1 and H2 on one page.
3. **Prefix hit rate vs concurrency.** H3, the honest counterweight: what distribution costs.

Say exactly what the table supports: *on a shared-prefix hotspot workload with 4 pods, the adaptive handler reduced TTFT p99 by X% and increased tokens/s per GPU by Y% at concurrency ≥ 32, at the cost of Z points of prefix hit rate, with N profile switches per minute.* If X or Y is not positive, say that.

## 5. Pitfalls that invalidate a run

| Pitfall | Symptom | Fix |
|---|---|---|
| Prefix caching off or prefix too short | No hotspot in A; A and C look identical | Check vLLM prefix-hit counters increase; `prefix_tokens ≥ 1024` |
| `max-num-seq` too high for the load | Queues stay 0, pressure stays 0 | Lower to 64 or raise concurrency |
| Changing two things at once | Cannot attribute the difference | Only image + plugins config differ between A and C |
| One repetition | Jitter dominates | 3 repetitions, median, show spread |
| EPP or caches still warming | First 30 s of a run look different | 60 s settle time after every switch |
| Model download during a run | TTFT spikes unrelated to routing | `rollout status` first; HF cache on a PVC |
| C tuned on the same data it is judged on | Thresholds overfit | Pick `enter/exit` and the gate from a **different** run (concurrency 48), then freeze them |

## 6. Results log

One directory per campaign: `nael-docs/v2/results/<YYYY-MM-DD>-hotspot/` with the GuideLLM JSONs, the PromQL exports, and a `README.md` holding the table from section 4 plus:

```text
Date · cluster · GPU type × count · model · vLLM version · EPP image (v0.11.0 / adaptive-<sha>)
Run A plugins config (default-plugins.yaml, weights) · Run C plugins config (full YAML)
Exact guidellm command · repetitions · cache policy (warm) · settle time
```

## Code references

| Ref | Location | Used for |
|---|---|---|
| 1 | `llm-d-router` `config/charts/routerlib/templates/_config.yaml` | Run A profile: queue 2 · KV 2 · prefix 3 |
| 2 | `llm-d-router` `pkg/epp/metrics/metrics.go` | EPP `llm_d_router_endpoint_picker_prefix_*` metric names |
| 3 | `llm-d-router` `config/manifests/benchmark/benchmark.yaml` | Alternative load tool (ShareGPT sweep); no shared-prefix knob, not used |
| 4 | GuideLLM `docs/en/guides/datasets.md` | `prefix_tokens` synthetic data option |
| 5 | `nael-docs/v2/epp-upstream_issue-template.md` | Run C plugins config and `adaptive_profile_selections_total` |

## 🔧 If you only do one thing

Run A once at concurrency 64 and watch `vllm:num_requests_waiting` per pod. If one pod's queue does not climb while the others stay flat, fix the workload before anything else.

## 🧠 Remember it as

Write the hypothesis first, make the hotspot on purpose, change one switch, watch four queues.
