# NIM test/dev sizing flows

These diagrams describe the implemented API and CLI behavior. They document the current flow; they do not validate runtime compatibility or model startup.

Exact image/tag metadata is read through the existing CP cache with a configurable 15-minute default expiry. Cache misses/failures fetch NGC directly; only successful nonempty profile lists are cached. Missing data and errors are not cached. Inventory and allocations remain fresh on every request. See the [service configuration and cache policy](../../README.md#nim-testdev-right-sizing).

## Full feature flow

Request inputs, authorization, paginated inventory, exact-tag metadata, profile selection, allocation and response boundaries.

![NIM sizing feature flow](nim-sizing-feature-flow.png)

[Editable SVG](nim-sizing-feature-flow.svg)

## API decisions and responses

Covers omitted/null and explicit `gpuType`, empty inventory, absent types, missing capacity or metadata, direct/tested/GPU-agnostic classification, empty or missing tested-device lists, and mixed result statuses. A missing tested list does not exclude a direct or GPU-agnostic profile. `no_recommendation` is a valid HTTP 200 result with a reason.

![API decisions and responses](nim-sizing-api-decisions.png)

[Editable SVG](nim-sizing-api-decisions.svg)

## Allocation policy

Shows the single configurable sharing cutoff (default 90%), exact boundary examples, and the explicitly estimated count above device capacity. The examples use a decimal 80 GB device; they are arithmetic illustrations, not model-specific NGC measurements.

![Allocation policy and boundaries](nim-sizing-allocation-policy.png)

[Editable SVG](nim-sizing-allocation-policy.svg)

## Share and maintain

- [Three-page PDF](feature-flow-diagrams.pdf)
- [Structured diagram layout](layout.json)
- PNG files are suitable for presentations; SVG files preserve editable vector text and connections.

Update the diagrams when the corresponding implementation changes. Source references:

- [API request and outcomes](../../swagger/source/resources/RightSizing.yaml) and [per-hardware response schema](../../swagger/source/schemas/rightSizing/RightSizingResult.yaml)
- [Inventory selection, profile matching and ranking](../../pkg/services/rightsizing/service.go)
- [Allocation calculation](../../pkg/services/rightsizing/allocation.go)
- [PCI identity resolution](../../pkg/clients/gpuinventory/pci.go) and [NVIDIA snapshot provenance](../../pkg/clients/gpuinventory/data/README.md)
- [Exact-tag NGC metadata and dependency handling](../../pkg/clients/ngcmetadata/client.go)
- [Endpoint authorization](../../pkg/router/router.go)

Published minima include runtime overhead and baseline KV cache according to the upstream clarification. Written confirmation of the exact settings remains pending. The cutoff controls sharing; estimated GPU counts do not establish supported tensor parallelism, device availability or schedulability.
