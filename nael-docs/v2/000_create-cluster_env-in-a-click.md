# 000 · Create the benchmark cluster with env-in-a-click

## TL;DR (the whole thing in 3 lines)

- One EKS cluster, one node with 4 real GPUs, no Run:AI installed. That is all docs 001 and 002 need.
- Six wizard steps, about 3 minutes of clicking, then 30–45 minutes of provisioning.
- Name: `nasbi-run-43533` · team: `viper` · ticket: RUN-43533.

URL: https://argocd-infra-main-02.runailabs.com/env-in-a-click
The page loads slowly (SSO redirects). Wait until the form appears inside the page.

```text
 Step 1 General ─► Step 2 Kubernetes ─► Step 3 Run:AI ─► Step 4 GPU ─► Step 5 Tools ─► Step 6 Summary ─► ArgoCD "Create"
```

## Step 1 · General Settings

| Field | Value | Note |
|---|---|---|
| Environment Name | `nasbi-run-43533` | `<whoami>-<ticket lowercased>`; becomes the cluster and ArgoCD app name |
| Team Name | `viper` | dropdown |

→ Continue

## Step 2 · Kubernetes Settings

| Field | Value | Note |
|---|---|---|
| Distribution | **EKS** | options are at the top of the step; scroll up if hidden |
| Region | `us-east-1` (default) | g5 instances available there |
| Disk | 100 GB (default) | enough for the model cache |
| Worker nodes | if a GPU node type is selectable here: **1 × g5.12xlarge** (4 × A10G 24 GB). If only CPU workers appear, leave default; the GPU node comes from step 4 | |

→ Continue

## Step 3 · Run:AI Settings

| Field | Value | Note |
|---|---|---|
| Installation type | **None** | we benchmark the EPP only |
| Version fields | ignore | greyed out with None |

→ Continue

If **None** cannot be combined with real GPUs, choose **Self-Hosted** with defaults and note it; the benchmark still works, Run:AI just sits beside it.

## Step 4 · GPU Settings

| Field | Value | Note |
|---|---|---|
| GPU type | **Real** | not Fake |
| GPUs per node | **4** | |
| GPU card / instance, if offered | **A10G** (g5.12xlarge) or **A100 40 GB** (p4d) | A10G is cheaper and fits Qwen3-8B in bf16 |
| Memory / other fields | leave default | the Fake-GPU fields (Tesla-V100, 12000 MB) do not apply to Real |

→ Continue

## Step 5 · Third-Party Tools

| Tool | Setting |
|---|---|
| Ingress Controller (nginx) | leave **on** (default, harmless) |
| KWOK, Kubeflow, Training / MPI / NIM / Ray operators, NFS, Knative, Harbor, ArgoCD | all **off** |

→ Continue

## Step 6 · Summary & Deploy

1. Check: name `nasbi-run-43533`, team viper, EKS us-east-1, Run:AI None, GPU Real × 4, no extra tools.
2. Click **Deploy Environment** (the button, not the description text).
3. ArgoCD opens a "Create application" dialog on the main page: click the blue **Create**.

## After deploy

- ArgoCD shows the app `nasbi-run-43533`. Wait for **Healthy / Synced**, typically 30–45 minutes for EKS with a GPU node.
- If it stalls on the GPU node, it is usually instance capacity in the region; retry in `us-east-2` or `us-west-2`.

## Hand-off checklist (what happens next, from doc 001)

```text
1. Connect:  kc add --session   (cluster nasbi-run-43533)
2. Verify:   kubectl get nodes -o custom-columns=NAME:.metadata.name,GPU:.status.allocatable.nvidia\.com/gpu
             expect one node with 4; read the card type from the node labels
3. Run 001:  CRDs → kgateway → 4 × vLLM (Qwen3-8B) → EPP v0.11.0 → Prometheus → smoke test
4. Run 002:  run A, 3 sweeps, export GuideLLM JSON + PromQL
```

## 🔧 If you only do one thing

Get Run:AI = **None** and GPU = **Real × 4** right. Everything else is a default.

## 🧠 Remember it as

One EKS node, four real GPUs, nothing else installed.
