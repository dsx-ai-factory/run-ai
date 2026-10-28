# Published NIM parallel profiles implementation plan

> Execute inline using superpowers:executing-plans; follow test-driven development.

**Goal:** Recommend the smallest preferred published layout that fits one node,
and remove unsupported arithmetic estimates.

**Architecture:** Add per-node count evidence to the existing authorized
nodepool inventory. Select published TP × PP layouts in catalog-service, return
their layout, preserve cache/single-device sharing behavior and regenerate APIs.

**Tech stack:** Go, GORM/PostgreSQL, existing OpenAPI generators, shared CP cache.

**Spec:** ../specs/2026-10-28-nim-published-parallel-profiles-design.md

## Global constraints

Preserve unrelated work. No PCI guessing or tested-only allowlist. Explicit
positive TP/PP and finite memory are required. No rescaled/power-of-two estimate.
Counts must fit one node; unknown counts/capacity produce no recommendation.
Keep existing authorization, pagination, dependency errors and 90% policy.
Generate contracts; never hand-edit generated code. No deployment/merge.

## Review focus

- Pool totals across nodes must never establish per-node feasibility.
- Different device-memory variants must retain their own per-node count.
- Published PP/non-power-of-two layouts must return their exact shape.
- No profile fit or absent metadata must not synthesize a count.
- Warm metadata cache must still see changed inventory and layout feasibility.

## Task 1: Per-node inventory evidence

**Files:** cluster-service schema `NodepoolGpuTypeDetails.yaml`,
`pkg/database/dao/node_dao.go`, `pkg/services/api_nodepools_gpu_types_test.go`;
catalog `pkg/clients/gpuinventory/inventory.go` and tests.

**Produces:** optional `maxDevicesPerNode` in inventory API and
`Hardware.MaxDevicesPerNode int64` (zero means unknown).

- [x] Add failing behavioral tests: nodes 8+4 yield maximum 8, not total 12;
  capacity variants and cross-cluster/scope isolation remain distinct.
- [x] Add schema source field; regenerate cluster server/Go/TS clients/specs.
- [x] Implement MAX aggregation and inventory merging by maximum. Reject negative,
  zero-explicit or greater-than-total counts; preserve omitted as unknown.
- [x] Run focused tests and verify failures become green.

## Task 2: Published parallel-profile selection

**Files:** catalog rightsizing service/tests; NGC metadata client model/tests;
rightSizing result/recommendation schema sources; generated API clients/specs.

**Consumes:** Task 1 count. **Produces:** exact published layout recommendations.

- [x] Add failing tests for TP2/4/8, TP3/PP2 exact count, fewest GPUs within a tier,
  targeting priority after fit, missing/overflowing layout, `max_gpu_count`,
  memory failure, unknown counts, one-node limits and no arithmetic fallback.
- [x] Add response TP/PP fields and regenerate catalog server/Go/TS clients/specs.
- [x] Implement explicit-layout validation and ranked selection; return exact
  whole-GPU count for parallel profiles and preserve TP1 cutoff behavior.
- [x] Update fixtures for per-node evidence, replace estimate assertions with
  no-recommendation assertions and strengthen cache/inventory regression.
- [x] Run full affected suites and targeted race/static analysis checks.

## Task 3: Review, documentation and publication

**Files:** catalog README, flow diagram source/assets and API documentation;
local verification record and concrete ticket-correction draft.

- [x] Update published count/assumptions descriptions and all three diagrams.
- [x] Run generation/up-to-date checks, catalog/cluster lint, required builds,
  CLI consumer tests and available Dima validation; record blockers precisely.
- [x] Request one fresh review of changed behavior and resolve valid findings.
- [x] Commit only scoped changes and publish to the existing PR. Update its
  description/image links and verify the remote head. Do not modify Jira until
  explicitly authorized; prepare its wording for review.
