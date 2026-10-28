# NIM Service GPU placement and automatic sizing

Base: recommendation PR #3129, f6999bdcf14c8fe6ac3a422299cc1ccaeabcd51b.
Implementation is authorized; deployment waits for the user's environment selection.

## Contract

Only NIM Service gains optional `gpuType` / `--gpu-type`. The value is the exact
`nvidia.com/gpu.product` label. No PCI translation or scheduler changes are needed.
Without a type, retain existing behavior. With an explicit or policy-default GPU
allocation, retain that allocation. With a type and no allocation, request sizing
for the project's cluster and exact image tag, then validate the resolved spec
against policy before creating anything. CPU resources and compute mode alone do
not count as a GPU allocation. Missing recommendations return HTTP 400.

Use the recommendation API's allocation and selected runtime profile together.
Do not silently promote estimated multi-GPU arithmetic to validated sizing.
For multiple capacities under one product label, automatic submission must have a
single allocation/profile valid across the matching results; reject ambiguity
instead of relying on inventory order. Explicit allocations remain available.
Keep resolved automatic values in the persisted spec so updates do not resize
the workload from changing inventory or metadata. Preserve policy precedence.

## Implementation checklist

- [x] Add regression tests for placement and CLI omission/manual allocation.
- [x] Add `gpuType` to fielder sources and regenerate NIM schemas, policy schemas,
      server/client models, and frontend API models using repository generators.
- [x] Compose validated GPU product label into NIM `spec.nodeSelector` using
      the existing FieldsComposer/setter conventions.
- [x] Add catalog client adapter and sizing resolution before policy validation;
      test no type, manual/default resources, memory/count results, failures,
      ambiguous capacities, profile consistency, and persistence/update behavior.
- [x] Add CLI submit flag and describe output; preserve existing allocation flags.
- [x] Run focused suites, relevant wider checks, generation validation, and an
      independent final review. Document completed and unverified checks.

Relevant paths: `control-plane/fielder/config.yaml`, shared OpenAPI schemas,
`control-plane/workloads-and-policies/composer/composer_v2/nim_workloads`,
`control-plane/workloads-manager/pkg/services`,
`control-plane/workloads-manager/pkg/clients`, and NIM CLI submit/spec/describe.

Recommendation policy (including 90% cutoff), metadata caching, and validated
multi-GPU profile selection remain owned by the base recommendation PR.

## Execution record

- Implemented optional GPU type, direct product-label placement, automatic sizing
  after policy defaults, and persisted allocation/profile. No changes to the
  recommendation selection algorithm or cache.
- Ruling: preserve allocation and runtime profile on updates; sizing runs only
  during creation. This avoids silently resizing an existing service.
- Ruling: require agreement across capacity variants sharing a product label;
  a node selector alone cannot distinguish those variants.
- Ruling: consume validated multi-GPU counts from the recommendation contract;
  refuse arithmetic estimates. The base PR owns validated profile selection.
- Review found and fixed numeric policy round-trip preservation and malformed
  HTTP 200 dependency responses. Both have regression coverage.
- Schema generation used repository Makefile generator targets and local Go
  formatting/standard scaffold cleanup. Fielder validation passes.
- Full local workloads-manager package suite passes, including isolated-database
  submission/persistence tests. NIM composer and CLI spec/NIM suites pass.
- Go vet passes for services and the new catalog adapter. Diff whitespace check
  passes. Optional module tidy was stopped while downloading unrelated transitive
  test dependencies; compilation and tests resolve the new local dependency.
- No live environment/API deployment test has been run. Deployment is explicitly
  deferred until the user selects an environment. No push or new PR publication.
