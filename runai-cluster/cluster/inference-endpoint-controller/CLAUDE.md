# Inference endpoint controller — build context

Viper Team (`@viper`); GitHub owner `@dsx-ai-factory/workloads`.
Independent Go module and container artifact under `runai-cluster/cluster`.

- `api/v1alpha1`: InferenceEndpoint API; generate deepcopy and CRDs after edits.
- `internal/controller`: reconciliation and status; Ginkgo/Gomega envtest suites.
- `internal/render`: embedded YAML templates; golden fixtures in `test/testdata/golden`.
- `config`: standalone test deployment and generated RBAC/CRD.
- Product deployment is owned by the go-operator operand, not standalone Kustomize.
- SDK settings live in `../sdk/apis/run/v1/inference-endpoint-controller`.
- Run `make generate manifests copy-crds` after API/RBAC changes. Never hand-edit generated output.
- Run `make build test lint validate`; `make test-dev` provides local envtest.
- Keep embedded templates in the Go binary. Docker packages the built binary and NOTICE.
- No external cluster is needed for unit/envtest tests. `make test-e2e` uses a disposable Kind cluster.
- The controller is opt-in (`enabled: false` by default) because the Gateway API and GIE CRDs are prerequisites.

## Make targets

| Target | What it does |
|--------|-------------|
| build | `build-go`: amd64 and arm64 binaries into `bin/` |
| build-go-amd | amd64 binary (reads `OS=`/`ARCH=`) → `bin/inference-endpoint-controller-amd64` |
| build-go-arm | arm64 binary (sets `GOOS=linux GOARCH=arm64` itself) → `bin/inference-endpoint-controller-arm64` |
| test | `make test-dev` inside the shared Go container |
| test-dev | `setup-envtest`, then `go test ./... -coverprofile cover.out -timeout 10m` on the host |
| lint | `lint-go` |
| format | `format-go` |
| validate | `generate manifests copy-crds`, then `git diff --exit-code` on `api`, `config`, the two packaged CRD copies, and the chart ClusterRole |
| generate | controller-gen deepcopy code |
| manifests | controller-gen CRD into `config/crd/bases` and RBAC (`manager-role`) |
| copy-crds | `manifests`, copy the CRD to `../runai-chart/charts/crds/crds/` and `../runai-crd-upgrader/crds/`, run `hack/generate-helm-rbac.sh` |
| controller-gen, kustomize, setup-envtest | install the pinned tool into `bin/` |
| run | `go run ./cmd/main.go` against the selected kubeconfig |
| docker-build | host-architecture binary and `docker build -t $(IMG)`; used by the Kind smoke suite |
| build-installer | `kustomize build config/default` → `dist/install.yaml` |
| install, uninstall | apply or delete `config/crd` on the selected cluster |
| deploy, undeploy | apply or delete `config/default` (standalone test deployment) with image `$(IMG)` |
| test-e2e | `hack/test-e2e.sh`: create a Kind cluster, run `go test -tags=e2e ./test/e2e/`, delete the cluster |

Includes: `../../../build/makefile/runai-cluster.mk`, `index.mk`.

## Local build (local-dev.sh)

runai-cluster Go service (`RC_GO_SERVICES` → `build_go_service`). Image tag `0.0.0-dev`, the tag the
runai-cluster chart values reference. `local-dev.sh` builds amd64 only:

```bash
make build-go-amd OS=linux ARCH=amd64 DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=amd64 -t runai-local/inference-endpoint-controller:0.0.0-dev .
```

arm64 is a manual path with the same tag; do not pass `OS=`/`ARCH=` to `build-go-arm`:

```bash
make build-go-arm DOCKERHUB_PROXY= GITHUB_TOKEN=
docker build --build-arg TARGETARCH=arm64 -t runai-local/inference-endpoint-controller:0.0.0-dev .
```

Iterate with `./local-dev.sh --skip-build-cp --skip-cluster` from the repository root.
