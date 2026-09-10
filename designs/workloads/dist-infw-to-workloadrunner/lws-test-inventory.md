# LWS test inventory

This is the baseline for the future typed `DistributedInferenceSpecSpec → LeaderWorkerSetTarget` composer tests.

## TL;DR

- The native LWS builder registers **81 executions / 80 unique cases**: 62 successful cases and 18 error cases. `path-rewrite-nginx-http` is registered twice.
- It uses **80 input fixtures** and **62 expected-output goldens**.
- Port the LWS-shape cases; adapt networking/MNNVL to `WorkloadV2Configuration`; keep controller, status and installer tests out of the composer suite.

## Source of truth

| Current suite | Coverage | Reuse for new composer |
|---|---:|---|
| [`builders_test.go`](../../runai-cluster/cluster/inference-workload-controller/internal/builders/builders_test.go) | 80 unique LWS builder cases | Primary behavior inventory |
| [`testing/inputs/lws`](../../runai-cluster/cluster/inference-workload-controller/internal/builders/testing/inputs/lws) | 80 inputs | Translate inputs to typed API fixtures |
| [`testing/expected/lws`](../../runai-cluster/cluster/inference-workload-controller/internal/builders/testing/expected/lws) | 62 goldens | Compare only the LWS portion owned by the new composer |
| [`distributed_builder_test.go`](../../runai-cluster/cluster/inference-workload-controller/internal/builders/distributed_builder_test.go) | 7 preview/build parity tests | Reuse UID naming and complex leader/worker ideas |
| [`workload_composer_dinf_test.go`](../../control-plane/workloads-and-policies/composer_v2/runai_workloads/workload_composer_dinf_test.go) | 54 native CP composer cases | Reuse typed-input and validation scenarios |

## All native LWS builder cases

Legend: **Port** = same behavior belongs in the new composer. **Adapt** = preserve intent but change expected ownership/output. **Do not port** = not a composer responsibility.

### Topology and metadata — 4

| Current case | New composer |
|---|---|
| `basic` | Port |
| `no_spec` | Adapt to typed required-field validation |
| `no_worker_spec` | Port: copy leader into required LWS worker template |
| `zero_workers_and_replicas` | Port |

### Networking — 13

| Current case | New composer |
|---|---|
| `serving_port_only_port` | Adapt: LWS port + `configuration.networking` |
| `serving_port_expose_externally_true` | Adapt: external access instruction |
| `serving_port_expose_externally_false` | Adapt: internal Service only |
| `serving_port_with_custom_url` | Adapt: external host instruction |
| `serving_port_public_auth` | Port public behavior |
| `serving_port_auth_external` | Adapt to internal networking config + plugin auth proxy + external access |
| `serving_port_auth_internal` | Adapt to internal networking config + plugin auth proxy |
| `serving_port_auth_users` | Adapt to internal networking authorization with users |
| `serving_port_auth_groups` | Adapt to internal networking authorization with groups |
| `serving_port_all_fields` | Adapt to LWS + configuration golden |
| `path-rewrite-nginx-http` | Do not port: ingress implementation belongs to networking plugin |
| `path-rewrite-haproxy-http` | Do not port: ingress implementation belongs to networking plugin |
| `path-rewrite-no-path` | Do not port: ingress implementation belongs to networking plugin |

Note: `path-rewrite-nginx-http` appears twice in the current `toTest` table; count it once in the new suite.

### Lifecycle — 5

| Current case | New composer |
|---|---|
| `startup_policy_leadercreated` | Port |
| `startup_policy_leaderready` | Port |
| `restart_policy_recreate` | Port |
| `restart_policy_none` | Port |
| `restart_policy_entities` | Port |

### Pod, compute, security and placement — 25

| Current cases | New composer |
|---|---|
| `container_resources`, `container_resources_multi_frac` | Port using shared compute composers |
| `container_environment` | Port using shared environment composer |
| `container_security_context`, `privileged`, `host_ipc_network`, `run_as_non_root`, `read_only_root_filesystem`, `seccomp_profile_type` | Port using shared security composers |
| `container_probes`, `container_probes_with_serving_port` | Port |
| `annotations_labels`, `backend_workload_counter_annotation` | Port user metadata; adapt backend-owned annotations |
| `priority_class`, `preemptibility` | Port to leader and worker |
| `tolerations`, `pod_affinity`, `node_pools`, `node_type`, `node_affinity`, `node_affinity_with_node_pools` | Port |
| `image_pull_policy` | Port |
| `create_home_dir` | Port |
| `working_dir`, `stdin_tty` | Port |

### Storage and credentials — 7

| Current case | New composer |
|---|---|
| `pvc_existing` | Port to LWS volume and mount |
| `pvc_new` | Adapt to LWS mount + internal `configuration.pvcs` for the storage plugin |
| `pvc_mix` | Adapt both existing and new claims in the same target |
| `configmap_volume` | Port |
| `secret_volume` | Port |
| `emptydir_volume` | Port |
| `image_pull_secrets` | Adapt to shared credentials ownership |

### MNNVL — 8

| Current case | New composer |
|---|---|
| `mnnvl_basic` | Adapt: emit `configuration.mnnvl=Required` |
| `mnnvl_none` | Adapt: emit `configuration.mnnvl=None` |
| `mnnvl_leader_gpu_only` | Adapt configuration; plugin owns claims/affinity |
| `mnnvl_worker_gpu_only` | Adapt configuration; plugin owns claims/affinity |
| `mnnvl_multiple_replicas` | Adapt configuration; plugin owns ComputeDomain |
| `mnnvl_gpu_memory_and_fraction` | Adapt configuration; shared compute composer owns resource metadata |
| `mnnvl_pod_affinity` | Adapt configuration; MNNVL plugin owns injected affinity |
| `mnnvl_elastic_enabled` | Do not port unless the new V2 contract exposes this option |

## All 18 native negative cases

| Current case | Expected new-composer treatment |
|---|---|
| `serving_port_missing_port` | Validation error |
| `serving_port_multiple_ports` | Not applicable to singular typed `servingPort`; cover DTO projection instead |
| `serving_port_auth_public_with_users` | Validation error |
| `serving_port_auth_public_with_groups` | Validation error |
| `serving_port_auth_authenticated_with_users` | Validation error: allowlists are invalid for `authenticatedUsers` |
| `serving_port_auth_authenticated_with_groups` | Validation error: allowlists are invalid for `authenticatedUsers` |
| `serving_port_auth_authorized_users_with_groups` | Validation error |
| `serving_port_auth_authorized_groups_with_users` | Validation error |
| `serving_port_auth_authorized_users_empty` | Validation error |
| `serving_port_auth_authorized_groups_empty` | Validation error |
| `serving_port_auth_authorized_users_both` | Validation error |
| `serving_port_auth_authorized_groups_both` | Validation error |
| `serving_port_auth_authorized_users_or_groups_invalid` | Not in the current typed DistInf enum; reject during validation |
| `startup_policy_invalid` | Validation error |
| `restart_policy_invalid` | Validation error |
| `pvc_ephemeral` | Validation error; DistInf does not support ephemeral PVCs |
| `mnnvl_required_without_gpus` | Workloads-manager validation; not an LWS composer error |
| `mnnvl_no_replicas` | Workloads-manager/plugin validation; not an LWS composer error |

## Other current LWS suites

These tests remain valuable regression coverage but should not be copied into the new composer package.

| Suite | Tests | Responsibility |
|---|---:|---|
| [`distributed_builder_test.go`](../../runai-cluster/cluster/inference-workload-controller/internal/builders/distributed_builder_test.go) | 7 | Native `GetPreview`/`Build` parity, UID naming, compound resources |
| [`controller_test.go`](../../runai-cluster/cluster/inference-workload-controller/internal/controller/distributedinference/controller_test.go) | 11 | Native controller reconciliation and LWS upsert |
| [`distributed_resources_test.go`](../../runai-cluster/cluster/inference-workload-controller/internal/resources/distributed_resources_test.go) | 12 | Native resource discovery and missing-resource handling |
| [`lws_status_test.go`](../../runai-cluster/cluster/inference-workload-controller/internal/status/distributed/lws_status_test.go) | 45 | Pod/group aggregation into LWS status |
| [`distributed_status_test.go`](../../runai-cluster/cluster/inference-workload-controller/internal/status/distributed/distributed_status_test.go) | 48 | Native DistInf status, URLs and degraded states |
| [`components/lws_test.go`](../../runai-installer/installer/internal/prereq/components/lws_test.go) | 5 | Installer LWS component metadata |
| [`steps/lws_test.go`](../../runai-installer/installer/internal/prereq/steps/lws_test.go) | 7 | LWS CRD/install detection |
| [`policy-service LWS fixtures`](../../control-plane/policy-service/pkg/system_mapping/testing/inputs/leaderworkerset_x_k8s_io_LeaderWorkerSet_v1) | 4 goldens | Generic policy mapping at workload/template/pod/container layers |

## Minimum new composer suite

Start with these goldens before expanding to full parity:

1. Basic leader-only; verify worker-template fallback.
2. Multiple workers; verify `size = workers + 1`.
3. Multiple replicas and deterministic UID-suffixed LWS name.
4. Different leader/worker image, command, compute, security and placement.
5. Startup and restart policy variants.
6. Internal and external serving; verify leader-only Service selector.
7. MNNVL `Required` and `None`; verify configuration only.
8. Missing/`Preferred` MNNVL; verify it maps to `None` for native parity.
9. Existing and new PVCs; verify LWS mounts and internal storage-plugin configuration.
10. Update recomposition produces the same LWS shape as create.

## 🔧 If you only do one thing

Use the 80 native inputs as the behavior checklist, but create fresh typed inputs and V2 goldens rather than copying native CR fixtures directly.
