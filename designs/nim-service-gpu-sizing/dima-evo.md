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

## GPU type completion and case-insensitive CLI lookup

The local feature CLI now completes NIM submit `--gpu-type` from the selected
cluster's paginated node inventory. It returns sorted, unique, exact product
labels and matches typed prefixes without regard to case. No static GPU map is
used. Standard shell completion must be enabled for Tab to invoke it.

NIM submission resolves a case-insensitive inventory match to its exact label
before sending the API request, preserving explicit allocations. Exact matches
win; ambiguous case-only matches fail with a request to use the exact label.
Unknown names retain existing API/manual scheduling behavior. Inventory errors
are returned rather than silently submitting a possibly incorrect selector.
Omitting the GPU type skips this inventory lookup. This is CLI-only behavior;
direct API callers must still supply the exact label.

Enable completion for the existing zsh alias in the current terminal:

```bash
autoload -Uz compinit && compinit
source <(runai-evo completion zsh)
compdef _runai runai-evo
zstyle ':completion:*:*:*runai*:*' matcher 'm:{a-zA-Z}={A-Za-z}'
```

The local wrapper sets `RUNAI_CLI_CONFIG_PATH` so normal execution and Cobra's
completion process use the same authenticated profile. A live mixed-case
submission resolved `nViDiA-a100-sxm4-40gb` to
`nvidia.com/gpu.product=NVIDIA-A100-SXM4-40GB` and requested 33 GB. The temporary
service was deleted after verification.

## Review follow-up (2026-10-28)

- CLI version: 2.124.0; generated NIM submit reference includes the GPU-type flag.
- Real inventory clients are rebuilt per call, preserving current credentials and
  domain. Mock injection remains supported and regression tested.
- Explicit zero GPU requests retain CPU-only semantics; they do not trigger
  recommendations. Omitting GPU allocation is distinct from requesting zero.
- With a GPU type, limit-only requests (including zero plus a limit) now fail
  before creation instead of silently losing their limits.
- With a GPU type, a memory request with more than one GPU now fails explicitly.
  The current memory annotation represents one device and previously lost the
  count. Use a whole-GPU count or device count plus memory fraction instead.
- Policy integration tests cover default allocation precedence, locked defaults,
  rejection of 33 GB against a 20 GB maximum, and rejection of eight devices
  against a four-device maximum. Rejections produce no queued workload.
- Positive placement checks covered all three simulated GPU types, auto sizing
  on both A100 capacities, an eight-GPU recommendation, and combined flags.
  All pods were assigned to nodes with the requested product label. The original
  count-plus-memory probe exposed the unsupported-count behavior above; it is
  not counted as a successful two-device allocation.

The cluster still uses fake GPUs. Pod node assignment and requested resource
composition are verified; model execution/readiness is not established.

The final review build is deployed as
`runai.jfrog.io/cp-containers-lab/workloads-manager:dima-evo-nim-review-20261028-v2`
(digest `sha256:6696c7ebead35dfd5f3c4e016bef66ccbe8376d3197ef409c4f17e940ab6b84b`).
Rollout completed successfully. Live checks confirmed rejection of limit-only
and multi-device memory requests and acceptance of explicit zero. Actual zsh
Tab keystrokes completed both `tesla-` and `nvidia-a100-sxm4-4` to the exact
inventory labels with the per-command matcher setting above. Evidence and
commands are recorded in `review-followup-results.json`.
