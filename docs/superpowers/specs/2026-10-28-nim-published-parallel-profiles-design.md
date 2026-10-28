# Published NIM parallel-profile sizing

Approved direction: select fitting published profiles, retaining exact image/tag,
precision, backend, profile ID and explicit TP/PP. Remove arithmetic count
estimates: a scalar TP1 memory floor cannot establish another layout's memory.
No experimental estimator is implemented because exact-runtime support and a
calibrated memory model are unavailable.

## Selection contract

- Preserve the fitting TP1/PP1 path and its configurable 90% sharing cutoff.
- Otherwise consider explicit positive TP/PP layouts. Count is TP × PP, without
  power-of-two rounding; reject overflow and inconsistent `max_gpu_count`.
- Filter for matching hardware, positive finite minimum, per-device memory fit,
  and enough installed matching GPUs on one node before ranking.
- Rank surviving parallel profiles by direct PCI, tested PCI, GPU-agnostic tier;
  then count, published minimum, profile ID. Tested membership is evidence, not
  an exclusive allowlist. Existing identity and compatibility warnings remain.
- Parallel profiles allocate whole GPUs. Return TP and PP with the source
  profile. `estimated=false` describes published layout/allocation arithmetic,
  not model startup, free capacity, quota, driver or interconnect compatibility.
- Absent metadata/minimum, unknown per-node count/capacity, no applicable fit or
  a layout exceeding each node yields an explained `no_recommendation`.
- Keep dependency errors, authorization, exact-tag cache, pagination, explicit
  type filtering and independent null-type results.

## Inventory and API

Extend the existing nodepool GPU type/capacity grouping with
`maxDevicesPerNode = MAX(gpu_count)`. Preserve tenant/scope predicates and pool
identity. Catalog merges duplicate type/capacity entries by maximum count,
never sum, and validates impossible counts. Missing counts stay unknown.
This adds no new discovery endpoint or direct Kubernetes access.

Return `tensorParallelSize` and `pipelineParallelSize` on recommended results.
Generate all affected server/client/schema artifacts using repository tooling.
Older cluster-service versions lacking per-node counts return no recommendation
until upgraded; no counts are inferred from pool totals.

## Assumptions and non-goals

Published minima are publisher-defined. KV cache, concurrency/context and
speculative decoding assumptions may vary. Use the number unchanged, add no
headroom, and do not describe it as a universal runtime guarantee.
No model-name estimator, unsupported layout synthesis, multi-node placement,
free-resource scheduling or automatic deployment is added.

## Approaches

The accepted approach uses published layouts plus conservative no-recommendation
outcomes. Rescaling a TP1 floor or rounding its quotient was rejected because
neither proves a supported layout or sufficient per-device memory. A runtime-
specific experimental estimator remains a separate future capability.

## Verification

Exercise maximum-per-node aggregation/merge and access boundaries; exact TP/PP
selection including non-power-of-two/PP layouts; metadata absence and rejected
baselines; one-node limits, mixed capacities, deterministic ranking, unchanged
cutoff boundaries and cached metadata with fresh inventory. Run catalog/cluster
tests, generation/validation, lint, builds and available API validation. Update
documentation and existing PR diagrams. No merge or deployment is authorized.
