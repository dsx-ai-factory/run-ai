# workloads-and-policies — Build Context

**Type:** (unset in `artifact-metadata.json`) — a shared Go library, no image, no binary
**Team:** Viper Team (`artifact-metadata.json`)
**Module:** `github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies` (Go 1.26.3)

Shared Go library holding the workload→CRD composer (`composer_v2`), policy/GVK
compliance, topology, auth middleware and service clients. Imported by
`workloads`, `workloads-manager`, `assets-service`, `catalog-service`,
`audit-service`, `policy-service`. CI only lints and tests it.

## Make targets

Targets come from the service `Makefile` plus shared includes
(`../../build/makefile/control-plane.mk`, `index.mk`, `golang.mk`).

| Target | What it does |
|--------|-------------|
| `build` | No-op — library, nothing compiled into an artifact ("CI just runs tests") |
| `validate` | No-op — overrides the imported target (prints "Validation step skipped.") |
| `lint` | `lint-go` — golangci-lint in the Go builder container |
| `purify` | `purify-go` — `go fmt` / `go vet` / `go mod tidy` |
| `test` | `test-ginkgo` — Ginkgo suites over `BUILD_IN_PATH=./...` |
| `coverage` | `test-go-acc` — coverage via go-acc, `GO_ACC_IGNORE=pkg/clients,pkg/kube_client` |

## Key build notes

- **No container image, no binary.** `artifact_type` is empty; not in
  `local-dev.sh`'s `CP_GO_SERVICES`; no `runai-local/...` tag. It is compiled as
  part of the services that import it.
- **Layout:** `composer_v2/` (workload→CRD composer + field validations; read its
  `README.md` and `AGENTS.md` before editing), `policy_compliance/`,
  `gvk_compliance/`, `topology/`, `auth/` (middleware + mocks), `clients/`
  (authorization/clusterservice/orgunit/policies clients + mocks),
  `workloadutils/`, `itemized/`, `third_party/kserve` (vendored, own `go.mod`).
- **Tests** are Ginkgo-based, colocated with the code (`suite_test.go` per
  package) and use in-package mocks — no live backing services needed.
- **No config:** the library has no entry point / env vars / config file;
  configuration comes from the consuming service.

## composer_v2 conventions (from AGENTS.md)

When touching `composer_v2`: use existing `fc.*` copy helpers and setters; keep
`Compose*` generic and put workload-type-specific logic in setters; follow
`iw_tw_workloads/` as the reference; wrap compose calls in `runner.Batch(...)`.

## local-dev.sh iteration

Not built or deployed by `local-dev.sh`. Lint/test directly:
```bash
cd control-plane/workloads-and-policies && make test && make lint
```
General CP iteration flag:
```bash
./local-dev.sh --skip-build-rc --skip-cluster
```
