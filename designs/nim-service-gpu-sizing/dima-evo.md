# dima-evo deployment and CLI examples

Deployed on 2026-10-28 from `5e4a73d8b`, combining the NIM submission feature
with recommendation PR #3129 head `2801438fc75a0466eb820a29749c70f45579dc0c`.
The local branch is `codex/nim-service-gpu-sizing`. No Git branch or new PR was
published; the four deployment images were published and applied to dima-evo.

## Deployment

All four deployments in `runai-backend` rolled out successfully using
`runai.jfrog.io/cp-containers-lab/<service>:dima-evo-nim-20261028-5e4a73d8b`:

- `cluster-service`: per-node GPU inventory required by the updated recommendation API.
- `catalog-service`: current recommendation API, including published parallel profiles.
- `policy-service`: GPU type defaults/rules.
- `workloads-manager`: NIM automatic sizing and GPU product-label placement.

This was a targeted image update in the existing personal environment. The Helm
release version was not upgraded; a later Helm reinstall/upgrade can replace the
overrides. Existing ingress routing already covered both endpoints.

## Verified live

- Login, cluster selection, and project `p-1` work with the dedicated CLI.
- Llama 3.1 8B Instruct `2.0.13`: A100 40/80 GB recommends 33 GB BF16.
- Automatic 8B submission persisted the selected profile, requested 33 GB, emitted
  `nvidia.com/gpu.product=NVIDIA-A100-SXM4-40GB`, and scheduled a pod onto
  `runai-cluster-worker-2-0`.
- Llama 3.1 70B Instruct `2.0.13` on A100 40 GB: recommendation and submission
  selected eight whole GPUs, BF16, TP8/PP1, and the same published profile.
- Explicit `--gpu-devices-request 1` remained one GPU.
- V100 automatic submission returned a clear no-recommendation error and did not
  create a NIMService.
- All temporary smoke NIMServices and their temporary PVC were removed.

The 70B A100 80 GB case currently selects a GPU-agnostic NVFP4 TP1 profile with an
82 GB minimum because reported capacity is 83,886,080,000 bytes. Its response
includes the NVFP4/emulation compatibility warning. Use A100 40 GB for the
verified multi-GPU example below; recommendations are not runtime guarantees.

The environment uses `fake-gpu-operator`, so real model execution/memory fit
cannot be validated here. No NGC API-key secret or model cache was configured by
this deployment. The multi-GPU/manual smoke tests intentionally used a missing
cache to validate creation and composition without downloading the model.

## Connect

A dedicated Mac CLI and wrapper are available locally. The wrapper fixes the
profile to `$HOME/.runai-nim-evo`, already configured and authenticated against
`https://dima-evo.runailabs.com`, with automatic CLI replacement disabled.

```bash
alias runai-evo='/Users/dpolonchuk/.codex/worktrees/05e3/run-ai/control-plane/runai-cli/bin/runai-evo'
runai-evo cluster list
runai-evo project set p-1

# When the session expires:
runai-evo login
```

The selected cluster is `dima-evo-1792672603`. This CLI includes both
`inference recommend` and NIM `--gpu-type`; the environment's older CLI download
does not contain these feature changes.

## Recommendations

```bash
# All installed GPU types/capacities:
runai-evo inference recommend \
  --image nvcr.io/nim/meta/llama-3.1-8b-instruct --tag 2.0.13

# One type, with a published multi-GPU profile:
runai-evo inference recommend \
  --image nvcr.io/nim/meta/llama-3.1-70b-instruct --tag 2.0.13 \
  --gpu-type NVIDIA-A100-SXM4-40GB
```

## Submission-only experiments on simulated GPUs

These commands intentionally refer to a missing demo cache. The API creates the
NIMService and resolves allocation/placement, but the operator waits for the cache.
They do not demonstrate model readiness. For a real GPU environment, configure
a valid NGC secret and a matching NIMCache, or use an appropriately sized PVC.

```bash
# Automatic minimum: omit GPU allocation flags.
runai-evo inference nim submit nim-auto -p p-1 \
  --image nvcr.io/nim/meta/llama-3.1-8b-instruct:2.0.13 \
  --gpu-type NVIDIA-A100-SXM4-40GB \
  --ngc-auth-secret ngc-api-key \
  --model-nim-cache name=nim-evo-demo-cache --serving-port 8000

runai-evo inference nim describe nim-auto -p p-1

# Explicit allocation: preserve one whole GPU.
runai-evo inference nim submit nim-manual -p p-1 \
  --image nvcr.io/nim/meta/llama-3.1-8b-instruct:2.0.13 \
  --gpu-type NVIDIA-A100-SXM4-40GB --gpu-devices-request 1 \
  --ngc-auth-secret ngc-api-key \
  --model-nim-cache name=nim-evo-demo-cache --serving-port 8000

# Automatic published multi-GPU allocation: eight GPUs on this inventory.
runai-evo inference nim submit nim-multi -p p-1 \
  --image nvcr.io/nim/meta/llama-3.1-70b-instruct:2.0.13 \
  --gpu-type NVIDIA-A100-SXM4-40GB \
  --ngc-auth-secret ngc-api-key \
  --model-nim-cache name=nim-evo-demo-cache --serving-port 8000

runai-evo inference nim describe nim-multi -p p-1

# Cleanup only your examples:
runai-evo inference nim delete nim-auto -p p-1
runai-evo inference nim delete nim-manual -p p-1
runai-evo inference nim delete nim-multi -p p-1
```

For a no-recommendation submission, use a new workload name with
`--gpu-type Tesla-V100` and omit GPU allocation. For legacy behavior, omit
`--gpu-type` and specify the desired GPU resources yourself.

## Local validation and rollback

After merging the base PR, all Go package tests and Linux amd64 builds passed for
the four deployed services. CLI spec/NIM/recommendation tests and the Mac CLI
build passed. Prior implementation static checks and schema validation are
recorded in `verification.md`.

Before this deployment, cluster-service, policy-service and workloads-manager
used `runai.jfrog.io/cp-containers-test-virt/<service>:0.0.0-30005789.main-88da530`.
Catalog-service used
`runai.jfrog.io/cp-containers-lab-virt/catalog-service:0.0.0-30019678.nim-test-dev-sizing-a-3092945`.
The exact rollback image/container snapshot is also saved locally at
`/tmp/nim-evo-before-images.json`. These rollback targets are recorded only;
they have not been applied.
