# workloads-and-policies

`workloads-and-policies` is a **shared Go library** (module
`github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies`) that holds the
domain logic for turning Run:ai control-plane workloads into their Kubernetes CRD
representation and for evaluating them against policies. It is owned by the
**Viper Team**. It is not a deployable service and builds no container image —
instead it is imported by the control-plane services that actually run workload
and policy logic: `workloads`, `workloads-manager`, `assets-service`,
`catalog-service`, `audit-service` and `policy-service`.

Because it is consumed as a library, its CI job only lints and runs the test
suite; there is intentionally nothing to "build".

## Contents

- [Overview](#overview)
- [Repository layout](#repository-layout)
- [Building and running](#building-and-running)
- [Testing](#testing)
- [Configuration](#configuration)

## Overview

The module provides:

- **composer_v2** — the "Composer" that converts a control-plane workload into
  its CRD representation and performs field-level validations so invalid
  workloads fail in the control plane rather than later in the cluster.
  Validations/transforms are implemented as `Compose*` functions on
  `FieldsComposer` plus pluggable `setter` functions (see
  [`composer/AGENTS.md`](composer/AGENTS.md) for concepts, patterns, and the
  list of concrete composers).
- **policy_compliance** — evaluation of workloads against policy defaults and
  rules: computing effective fields, applying defaults in order, and reporting
  compliance errors.
- **gvk_compliance** — GroupVersionKind-level policy/action validation, rule
  resolution, selector dependencies and set-merge logic.
- **topology** — network-topology handling for workloads (segments, annotations,
  network-topology derivation).
- **auth** — HTTP auth middleware for the consuming services: header parsing,
  RBAC, permitted-scope checks (with test mocks).
- **clients** — thin Go clients (and mocks) for the sibling control-plane
  services this logic depends on: `authorization`, `clusterservice`, `orgunit`,
  `policies`.
- **workloadutils** — shared workload helpers (constants, LWS availability,
  serving configuration, timeout synchronization) and a small `api` error type.
- **itemized** — itemized info/numbering helpers used when reporting per-item
  results.

## Repository layout

```text
.
├── composer_v2/         # Workload → CRD composer + field validations (see its README/AGENTS.md)
├── policy_compliance/   # Policy defaults/rules evaluation, effective fields, compliance errors
├── gvk_compliance/      # GVK-level policy action validation, rule/selector resolution, set-merge
├── topology/            # Network-topology segments/annotations for workloads
├── auth/                # Auth middleware: headers, RBAC, permitted scopes (+ mocks)
├── clients/             # Go clients + mocks for authorization / clusterservice / orgunit / policies
├── workloadutils/       # Shared workload helpers + api/ error type
├── itemized/            # Itemized info / numbering helpers
├── third_party/         # Vendored third-party module (kserve) with its own go.mod
├── go.mod / go.sum      # Module: github.com/dsx-ai-factory/run-ai/control-plane/workloads-and-policies
└── Makefile
```

## Building and running

Built through the [`Makefile`](Makefile), which includes
`../../build/makefile/control-plane.mk` and `index.mk`. This is a library, so the
CI contract is lint + test only.

| Target | Description |
|--------|-------------|
| `make build` | **No-op** — nothing is compiled into an artifact; the module exists to be imported, and its CI only runs tests |
| `make validate` | **No-op** (overrides the imported target; prints "Validation step skipped.") |
| `make lint` | `lint-go` — run golangci-lint in the Go builder container |
| `make purify` | `purify-go` — `go fmt` / `go vet` / `go mod tidy` |

There is **no image**, so this module is not part of `local-dev.sh`'s
`CP_GO_SERVICES` build loop and has no `runai-local/...` tag. Consuming services
pull it in via their `go.mod` and build it as part of their own binaries.

## Testing

| Target | Description |
|--------|-------------|
| `make test` | `test-ginkgo` — run the [Ginkgo](https://onsi.github.io/ginkgo/) test suites across the module (`BUILD_IN_PATH=./...`) |
| `make coverage` | `test-go-acc` — run coverage via go-acc, excluding `pkg/clients` and `pkg/kube_client` (`GO_ACC_IGNORE`) |

Tests live next to the code (`*_test.go` + `suite_test.go` per package) and rely
on the in-package mocks under `clients/` and `auth/` rather than live backing
services.

## Configuration

This module is a library and has **no service configuration** of its own — no
entry point, no env vars, no config file. Runtime configuration is supplied by
the services that import it. The clients under `clients/` take their connection
details from the consuming service.
