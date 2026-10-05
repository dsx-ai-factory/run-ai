# Inference endpoint controller

## TL;DR

Viper-owned controller that turns an `InferenceEndpoint` into an HTTPRoute,
InferencePool, endpoint-picker Deployment, and optional model rewrites.
The Run:ai operator installs it on the workload cluster, with configurable
image, replicas, resources, affinity, logging, and metrics.

## Installation

The controller is enabled by default for this integration. **Before merging,
revisit that default** in the SDK configuration: clusters without the prerequisites
will report the operand as not ready.

Install Gateway API v1 CRDs, InferencePool v1, and the GIE v1alpha2
InferenceModelRewrite and InferenceObjective CRDs on the workload cluster.
The pinned test manifests are under `test/crds/`. Configure a compatible Gateway
implementation and an existing Gateway before creating an endpoint. These upstream
components are prerequisites; the Run:ai chart installs the Run:ai-owned
InferenceEndpoint CRD through its CRD subchart and CRD upgrader.

Helm passes controller settings through `runai-operator.config`:

```yaml
runai-operator:
  config:
    inference-endpoint-controller:
      enabled: true
      replicas: 2
      image:
        registry: nvcr.io/nvidia/runai
        name: inference-endpoint-controller
        tag: <runai-release-tag>
```

The same controller block is available directly under `RunaiConfig.spec`.
Setting `enabled: false` removes the operator-managed controller resources.
Existing InferenceEndpoints and their children are retained.

Use `config/samples/inference_v1alpha1_inferenceendpoint.yaml` for the baseline
API, `_tuned.yaml` for scorer weights, and `_profiles.yaml` for the experimental
llm-d runtime profile. The API group remains `inference.platform.run.ai/v1alpha1`.
The controller manages routing resources; it does not create serving workloads.

## Build and validation

| Command | Purpose |
|---|---|
| `make build` | Shared monorepo build, including amd64 and arm64 binaries |
| `make test` | Containerized renderer and controller/envtest suites |
| `make test-dev` | The same suites with local envtest binaries |
| `make lint` | Shared cluster Go lint configuration |
| `make validate` | Regenerate and check API, CRD, and Helm RBAC artifacts |
| `make test-e2e` | Create a disposable Kind cluster and run the smoke suite |
| `make generate manifests copy-crds` | Update deepcopy, CRD, RBAC, and packaged manifests |
| `make run` | Run against the selected kubeconfig |

For host-side Linux image builds use `OS=linux ARCH=amd64`, as in `local-dev.sh`.
The normal artifact metadata drives PR/release builds and controller image inventory.
The scoped GitHub workflow adds the Kind smoke test.

`test/testdata/golden/` contains the hand-authored rendering fixtures.
`test/acceptance/` and `hack/e2e.sh` retain the standalone lab acceptance checks;
run them only against a prepared test cluster. `hack/e2e.sh` requires Python 3 with
PyYAML (`python3 -m pip install PyYAML`) to compare embedded YAML by value.
Product E2E scenarios belong in the separate `runai-e2e-internal` repository.

## Current limits

EPP configuration changes replace its single Pod using `Recreate`; this does not
provide EPP high availability. Controller replicas use leader election independently.
The experimental llm-d profile retains its pinned-image compatibility restrictions.
EPP images and external Gateway components require their own registry/mirror setup;
changing the controller image registry does not rewrite endpoint-picker images.

Imported from `NaelAsbi123/inference-networking-controller` at
`253d8fb2ef3f495e4806d13de5d4f6dea631d35f`; research and presentation files remain there.
