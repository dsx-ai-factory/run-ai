# 001 · Setup: real GPUs, real model, llm-d Router EPP from `main`, Gateway mode

## TL;DR (the whole thing in 3 lines)

- One Kubernetes cluster with **4 GPUs**, **4 vLLM pods** serving one real model with prefix caching on, **kgateway** as the Inference Gateway, and the **llm-d Router EPP** from the released `main` image behind it.
- Prometheus scrapes the EPP and the vLLM pods. That is where the numbers come from.
- Everything copies manifests and chart values that exist in `llm-d/llm-d-router` today; only GPU count, model, and (for run C) the EPP image change.

Status: setup guide, written 2026-10-12 against `llm-d/llm-d-router` `main` (`c0795deb0`, release `v0.11.0`). Not yet executed end to end. Items marked **check** must be confirmed at run time.

## Why llm-d Router and not GAIE

The Endpoint Picker code moved. `kubernetes-sigs/gateway-api-inference-extension` `main` now holds only the `InferencePool` CRD and a lightweight conformance EPP. The real EPP, the plugin framework, the scorers and the profile handlers live in `llm-d/llm-d-router` (renamed from `llm-d-inference-scheduler`). A benchmark that upstream will accept has to run there.

## What you end up with

```text
 ┌──────────────────────────────┐
 │ 📈 GuideLLM (doc 002)        │   pod inside the cluster
 └──────────────┬───────────────┘
                │ HTTP POST /v1/completions
 ╔══════════════▼═══════════════════════════════════════════════════════════╗
 ║ Kubernetes cluster · namespace epp-lab                                    ║
 ║                                                                           ║
 ║  ┌──────────────────┐  ext-proc gRPC   ┌───────────────────────────────┐  ║
 ║  │ ✅ Gateway        │────────────────►│ ✅ EPP  (run A)                │  ║
 ║  │ kgateway          │◄────────────────│  ghcr.io/llm-d/               │  ║
 ║  │ class: kgateway   │  chosen pod IP  │  llm-d-router-endpoint-picker │  ║
 ║  │ + HTTPRoute       │                 │  :v0.11.0                      │  ║
 ║  └────────┬─────────┘                  │ 🆕 EPP  (run C) our build      │  ║
 ║           │ forward request            └───────────────┬───────────────┘  ║
 ║           ▼                                            │ scrape /metrics   ║
 ║  ┌─────────────────────────────────────────────────────▼────────────────┐ ║
 ║  │ ✅ InferencePool "vllm-model" · 4 × vLLM pods · 1 GPU each            │ ║
 ║  │   real model, --enable-prefix-caching, --max-num-seq 256             │ ║
 ║  │   vllm:num_requests_waiting / running / gpu_cache_usage_perc         │ ║
 ║  └──────────────────────────────────────────────────────────────────────┘ ║
 ║  ┌───────────────────────┐                                                ║
 ║  │ 📊 Prometheus         │  EPP ServiceMonitor + vLLM PodMonitor           ║
 ║  └───────────────────────┘                                                ║
 ╚═══════════════════════════════════════════════════════════════════════════╝
```

✅ = exists today (released image, chart, manifests) · 🆕 = the only custom piece, our EPP build for run C.

## Prerequisites

| Need | Minimum | Notes |
|---|---|---|
| Kubernetes cluster | 1.30+, one node with **4 GPUs** | GPU Operator or device plugin installed; `nvidia.com/gpu` allocatable. 8 GPUs lets you match the repo's 8-replica manifest later. |
| GPU type | A100 40 GB / H100 / L40S | 8B model in bf16 with room for KV cache. L4 (24 GB): use a 3B model or fp8. |
| Model | `Qwen/Qwen3-8B` (no gate) | The repo manifest uses `Qwen/Qwen3-32B`; that needs 80 GB per pod. |
| Container registry | reachable from the cluster | Only for run C's image. Run A uses the public release image. |
| Local tools | `kubectl`, `helm`, `docker`, `git`, `jq` | Present on this machine. |
| Cluster access | `kc add --session` | Per the personal rules. |

## Sizing

| GPUs | Pods | Model | Why |
|---|---|---|---|
| **4** (minimum) | 4 | 8B bf16 | 1 hot pod vs 3 cold is visible; CV over 4 queues is meaningful |
| 8 | 8 | 8B bf16 | Repo manifest shape, closer to the reported numbers |
| 8 × 80 GB | 8 | 32B | Repo manifest as-is |

## Steps

### 0. Namespace, variables

```bash
export NS=epp-lab
export MODEL=Qwen/Qwen3-8B
export POOL=vllm-model                         # InferencePool name and pod label
export EPP_RELEASE=v0.11.0                     # run A image tag, verified pullable
export REGISTRY=<your-registry>/epp-lab        # run C image only

kubectl create namespace $NS
kubectl config set-context --current --namespace=$NS
# only if the model is gated:
kubectl create secret generic hf-token --from-literal=token=<your-hf-token>
```

### 1. CRDs: Gateway API + Inference Extension

```bash
git clone https://github.com/llm-d/llm-d-router.git && cd llm-d-router
git checkout c0795deb0                         # main at writing time; or: git checkout v0.11.0
kubectl apply -k deploy/components/crds-gateway-api
kubectl apply -k deploy/components/crds-gie
```

The two kustomizations pin the Gateway API and InferencePool CRD versions the router was built against (GIE `v1.6.2`).

### 2. Gateway provider: kgateway

```bash
KGTW_VERSION=v2.4.6                            # check: latest non-prerelease at writing time
helm upgrade -i --create-namespace -n kgateway-system --version $KGTW_VERSION \
  kgateway-crds oci://cr.kgateway.dev/kgateway-dev/charts/kgateway-crds
helm upgrade -i -n kgateway-system --version $KGTW_VERSION \
  kgateway oci://cr.kgateway.dev/kgateway-dev/charts/kgateway \
  --set inferenceExtension.enabled=true        # check: key name in the v2.4.x chart; may be on by default
```

Gateway resource (what the router's own `kubernetes-kgateway` dev overlay creates):

```yaml
apiVersion: gateway.networking.k8s.io/v1
kind: Gateway
metadata:
  name: inference-gateway
  namespace: epp-lab
spec:
  gatewayClassName: kgateway
  listeners:
  - name: default
    port: 80
    protocol: HTTP
```

```bash
kubectl apply -f gateway.yaml
kubectl wait gateway/inference-gateway --for=condition=Programmed --timeout=120s
```

Istio is the router's default dev provider (`deploy/environments/dev/base-kind-istio`); kgateway is the documented alternative overlay. Either works with the same chart.

### 3. vLLM pods (real model)

Adapted from `config/manifests/vllm/gpu-prefix-cache-deployment.yaml` on `main`: 4 replicas instead of 8, 8B model instead of 32B, `max-num-seq` lowered so queues form, label kept so the pool selects the pods.

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: vllm-model
spec:
  replicas: 4
  selector:
    matchLabels:
      app: vllm-model
  template:
    metadata:
      labels:
        app: vllm-model
        llm-d.ai/engine-type: vllm
    spec:
      containers:
      - name: vllm
        image: vllm/vllm-openai:v0.21.0          # as in the repo manifest
        command: ["python3", "-m", "vllm.entrypoints.openai.api_server"]
        args:
        - --model
        - Qwen/Qwen3-8B                          # $MODEL
        - --enable-prefix-caching
        - --tensor-parallel-size
        - "1"
        - --port
        - "8000"
        - --max-num-seq
        - "256"                                  # repo: 1024; lower so queues form under the sweep
        - --max-model-len
        - "16000"
        env:
        - { name: VLLM_USE_V1, value: "1" }
        - { name: PORT, value: "8000" }
        # only if gated:
        # - name: HUGGING_FACE_HUB_TOKEN
        #   valueFrom: { secretKeyRef: { name: hf-token, key: token } }
        ports:
        - { containerPort: 8000, name: http, protocol: TCP }
        readinessProbe:
          httpGet: { path: /health, port: http }
          periodSeconds: 2
        startupProbe:
          httpGet: { path: /health, port: http }
          failureThreshold: 600
          periodSeconds: 2
        lifecycle:
          preStop: { sleep: { seconds: 30 } }
        resources:
          limits: { nvidia.com/gpu: 1 }
          requests: { nvidia.com/gpu: 1 }
        volumeMounts:
        - { mountPath: /dev/shm, name: shm }
        - { mountPath: /root/.cache/huggingface, name: hf-cache }
      terminationGracePeriodSeconds: 130
      volumes:
      - name: shm
        emptyDir: { medium: Memory }
      - name: hf-cache
        emptyDir: {}                              # PVC if you will redeploy often
```

```bash
kubectl apply -f vllm-model.yaml
kubectl rollout status deploy/vllm-model --timeout=25m   # pull + model download
```

| Flag | Effect on the experiment |
|---|---|
| `--enable-prefix-caching` | Without it the prefix scorer never sees a hit, no hotspot forms, nothing to test. |
| `--max-num-seq 256` | Requests above this per pod queue. Lower = pressure appears at lower load. |
| 1 GPU × 4 pods | Four independent queues, so imbalance is a real quantity. |

### 4. EPP + InferencePool: the `llm-d-router-gateway` chart

Run A uses the released image. No build.

```bash
helm install $POOL ./config/charts/llm-d-router-gateway \
  --set router.modelServers.matchLabels.app=$POOL \
  --set router.epp.image.registry=ghcr.io/llm-d \
  --set router.epp.image.repository=llm-d-router-endpoint-picker \
  --set router.epp.image.tag=$EPP_RELEASE \
  --set router.epp.resources.requests.cpu=2 \
  --set router.epp.resources.requests.memory=2Gi \
  --set provider.name=none \
  --set httpRoute.create=true \
  --set httpRoute.inferenceGatewayName=inference-gateway \
  --set router.monitoring.prometheus.enabled=true    # check: value path in routerlib (monitoring.prometheus.enabled)
kubectl rollout status deploy/$POOL-epp              # check: deployment name from the chart NOTES
```

`provider.name=none` is the right value for kgateway (the chart's `istio` and `gke` providers add provider-specific objects). `httpRoute.create=true` creates the HTTPRoute from `inference-gateway` to the InferencePool. The chart's default CPU request is 8 cores, hence the override.

**What the default EPP config is** (from `routerlib/templates/_config.yaml`, this is run A):

```yaml
schedulingProfiles:
- name: default
  plugins:
  - pluginRef: queue-scorer
    weight: 2
  - pluginRef: kv-cache-utilization-scorer
    weight: 2
  - pluginRef: prefix-cache-scorer
    weight: 3
```

One profile, fixed weights, `single-profile-handler` injected by default. This is the baseline the whole experiment compares against.

### 5. Run C image (our prototype, later)

```bash
git checkout -b adaptive-profile-handler            # on top of main
# add pkg/epp/framework/plugins/scheduling/profilehandler/adaptive/ and register it
IMAGE_REGISTRY=$REGISTRY make image-build-epp image-push-epp
helm upgrade $POOL ./config/charts/llm-d-router-gateway --reuse-values \
  --set router.epp.image.registry=$REGISTRY \
  --set router.epp.image.tag=adaptive-<sha> \
  --set router.epp.pluginsConfigFile=adaptive-profiles.yaml \
  --set-file router.epp.pluginsCustomConfig.adaptive-profiles\\.yaml=adaptive-profiles.yaml
```

`adaptive-profiles.yaml` is the configuration from the issue draft, with the three profiles built as variations of the default weights above.

### 6. Prometheus

```bash
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm upgrade -i kube-prometheus-stack prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace \
  --set prometheus.prometheusSpec.serviceMonitorSelectorNilUsesHelmValues=false \
  --set prometheus.prometheusSpec.podMonitorSelectorNilUsesHelmValues=false
```

The chart's ServiceMonitor (step 4) covers the EPP. For the vLLM pods, the router's `deploy/components/monitoring/vllm-sim-pod-monitor.yaml` is the pattern; ours points at the `http` port:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: PodMonitor
metadata:
  name: vllm-model
  namespace: epp-lab
spec:
  selector:
    matchLabels: { app: vllm-model }
  podMetricsEndpoints:
  - port: http
    path: /metrics
    interval: 5s
```

### 7. Smoke test

```bash
IP=$(kubectl get gateway/inference-gateway -o jsonpath='{.status.addresses[0].value}')
curl -s ${IP}:80/v1/completions -H 'Content-Type: application/json' -d "{
  \"model\": \"$MODEL\", \"prompt\": \"Write as if you were a critic: San Francisco\", \"max_tokens\": 32
}" | jq .
kubectl logs deploy/$POOL-epp --tail=30 | grep -i -E 'scor|profile|selected|endpoint'
```

## Verify before moving to doc 002

| Check | Command | Expect |
|---|---|---|
| 4 vLLM pods ready | `kubectl get pods -l app=vllm-model -o wide` | 4 × `1/1 Running`, one GPU each |
| InferencePool accepted | `kubectl get inferencepool $POOL -o yaml \| yq .status` | `Accepted=True`, `ResolvedRefs=True` |
| HTTPRoute attached | `kubectl get httproute -o yaml \| yq '.items[].status'` | parent `inference-gateway`, `Accepted=True` |
| Gateway programmed | `kubectl get gateway inference-gateway` | `PROGRAMMED=True`, address set |
| EPP sees 4 endpoints | EPP `/metrics` → `inference_pool_ready_pods` (**check** name on main) or logs | 4 |
| Prometheus targets | UI → Status → Targets | EPP ServiceMonitor and vLLM PodMonitor `UP` |
| Prefix caching live | two identical prompts, then `curl <pod-ip>:8000/metrics \| grep prefix_cache` | hit counter increases (**check** exact vLLM metric names for v0.21.0) |

## Switching EPP behavior (used by doc 002)

| Run | What changes | How |
|---|---|---|
| **A · static** | default profile, release image | step 4 as written |
| **C · discrete (ours)** | our image + `adaptive-profiles.yaml` | step 5 |

## Teardown

```bash
helm uninstall $POOL
kubectl delete -f vllm-model.yaml -f gateway.yaml
helm uninstall kgateway kgateway-crds -n kgateway-system
helm uninstall kube-prometheus-stack -n monitoring
kubectl delete -k deploy/components/crds-gie; kubectl delete -k deploy/components/crds-gateway-api
kubectl delete namespace $NS
```

## Code references

All paths are in `llm-d/llm-d-router` at `main` (`c0795deb0`, 2026-10-08) unless noted.

| Ref | Location | Used for |
|---|---|---|
| 1 | `deploy/components/crds-gateway-api/`, `deploy/components/crds-gie/` | CRD install (step 1) |
| 2 | `deploy/environments/dev/kubernetes-kgateway/patch-gateways.yaml` | Gateway with `gatewayClassName: kgateway` (step 2) |
| 3 | `config/manifests/vllm/gpu-prefix-cache-deployment.yaml` | Base for the vLLM deployment (step 3) |
| 4 | `config/charts/llm-d-router-gateway/values.yaml`, `config/charts/routerlib/values.yaml` | `router.epp.image.*`, `router.epp.env`, `pluginsConfigFile`, `pluginsCustomConfig`, `httpRoute.*`, `provider.*` |
| 5 | `config/charts/routerlib/templates/_config.yaml` | The default profile: queue 2 · KV 2 · prefix 3 |
| 6 | `deploy/components/monitoring/epp-service-monitor.yaml`, `vllm-sim-pod-monitor.yaml` | Monitoring patterns (step 6) |
| 7 | `Makefile:455-491` | `image-build-epp`, `image-push-epp`; `IMAGE_REGISTRY`, `EPP_IMAGE_NAME` |
| 8 | `pkg/epp/framework/plugins/scheduling/profilehandler/{single,headerprofile,disagg,dataparallel}` | Where our `adaptive` handler goes |
| 9 | `pkg/epp/framework/interface/datalayer/metrics.go:35-46` | `RunningRequestsSize`, `WaitingQueueSize`, `KVCacheUsagePercent`, `UpdateTime` |
| 10 | `ghcr.io/llm-d/llm-d-router-endpoint-picker:v0.11.0` | Run A image, manifest verified pullable on 2026-10-12 |

## 🔧 If you only do one thing

Get step 7's curl to return a completion through kgateway with the `v0.11.0` EPP. Everything else is tuning.

## 🧠 Remember it as

Four real GPUs, four real queues, the EPP everyone already runs, one Prometheus watching all five.
