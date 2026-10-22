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
- The enabled-by-default decision has a TODO in the SDK to revisit before merge.
