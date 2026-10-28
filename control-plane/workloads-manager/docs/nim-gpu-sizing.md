# NIM Service GPU type and automatic sizing

NIM Service submission accepts `spec.gpuType` through
`POST /api/v2/workloads/nim-services`. The CLI exposes it as
`runai inference nim submit --gpu-type`.

Use the exact `nvidia.com/gpu.product` node label, for example
`NVIDIA-H100-80GB-HBM3`. The control plane copies this value into the NIMService
CR's `spec.nodeSelector`. There is no PCI translation in this submission path.
Placement assumes each node has one GPU model. A selector does not reserve
capacity or guarantee that the workload can immediately be scheduled.

| Submission | Behavior |
| --- | --- |
| No GPU type | Existing resource/default behavior. No recommendation call. |
| GPU type and explicit GPU resources | Preserve resources and restrict placement. |
| GPU type and policy-default GPU resources | Preserve policy defaults and restrict placement. |
| GPU type without GPU resources | Request a recommendation for the project's cluster and exact image tag, apply it, then validate policy. |
| No applicable recommendation | HTTP 400 before workload creation, with a reason. |
| Recommendation dependency fails | Return an error, preserving the dependency's HTTP status. |

CPU resources and `gpuComputeMode` alone do not count as a GPU allocation.
Explicit zero values or GPU limits are not replaced by automatic sizing; normal
resource validation still applies. An image with an explicit tag is required for
automatic sizing; digest-only or combined tag/digest references are unsupported.

The recommendation service owns profile eligibility, minimum memory, the 90%
sharing cutoff, and multi-GPU profile selection. The submit path uses the returned
byte quantity or device count without adding headroom. It accepts validated
multi-GPU recommendations but refuses a count marked as an arithmetic estimate.
Automatic multi-node sizing is not supported; specify resources explicitly.

Automatic sizing pins `NIM_MODEL_PROFILE` and, when using a NIMCache, its profile
to the selected profile. A conflicting explicit profile is a 400; specify GPU
resources explicitly to retain another profile. The referenced NIMCache must
actually contain that profile; submission does not populate the cache. Runtime
readiness remains a cluster-side check.

When multiple capacities share a GPU product label, the recommendations must
agree on profile and allocation across all matching capacities. Otherwise,
submission returns 400 and asks for explicit resources. A product label alone
cannot select a capacity variant.

Automatic choices are persisted with the submitted spec, so unrelated updates
reuse the resolved resources/profile rather than invoking sizing again. Changes
to image, GPU type, or resources during updates are explicit configuration edits;
automatic sizing is a creation-time operation.

The catalog service endpoint is configurable with `catalog_service_url`, default
`http://runai-backend-catalog-service:8080`. Calls carry the request context and
have a 60-second client timeout. This feature adds no recommendation cache.

NGC's published minimum is used for test/dev sizing. Rajan clarified that it
includes runtime overhead and baseline KV cache; exact settings await written
confirmation. The 90% cutoff is a sharing policy, not OOM protection. These
recommendations do not establish production or runtime guarantees.
