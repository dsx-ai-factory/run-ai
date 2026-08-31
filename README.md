# Run:ai

Run:ai is the AI workload orchestration platform for Kubernetes. This is the Run:ai monorepo — it contains the control plane services, the in-cluster components (the Run:ai Kubernetes operator and controllers), shared libraries and clients, the installer, and the local development environment.

## Overview

The platform is split into a **control plane** (multi-tenant management services and APIs) and a **cluster** side (Kubernetes controllers and agents that run in each managed cluster). The two communicate over authenticated HTTP: the cluster reports state up and applies desired state down.

Included components:

| Component | Path | Description |
|-----------|------|-------------|
| **Control Plane** | [`control-plane/`](control-plane) | Management services, APIs, and the control-plane Helm chart |
| **Run:ai Cluster** | [`runai-cluster/`](runai-cluster) | In-cluster operator, controllers, and the cluster Helm chart |
| **Installer** | [`runai-installer/`](runai-installer) | Installer and provisioner tooling |
| **Clients & shared packages** | [`runai-clients/`](runai-clients), [`runai-common-packages/`](runai-common-packages), [`runai-application-packages/`](runai-application-packages) | Generated API clients and shared Go libraries |
| **E2E** | [`e2e/`](e2e) | End-to-end test suites |
| **MCP server** | [`runai-mcp-server/`](runai-mcp-server) | Model Context Protocol server |
| **Local environment** | [`local-dev.sh`](local-dev.sh), [`local-dev.md`](local-dev.md) | One-command local Kubernetes (Kind) development environment |

## Installation

To deploy Run:ai on a real cluster, follow the official self-hosted installation guide:

- **Self-hosted install guide:** https://run-ai-docs.nvidia.com/self-hosted

For a full local environment on your own machine (control plane + cluster on Kind), see [Usage](#usage) below.

## Usage

### Local development environment

The fastest way to run the whole stack locally is the [`local-dev.sh`](local-dev.sh) script, which builds the services and brings up a self-contained [Kind](https://kind.sigs.k8s.io/) cluster with a local image registry. See [`local-dev.md`](local-dev.md) for the full guide.

```bash
./local-dev.sh                       # full build + cluster + deploy
./local-dev.sh --skip-build --skip-cluster   # re-deploy only (fastest iteration)
./local-dev.sh --teardown            # tear the environment down
```

Prerequisites (Docker/Podman, `kind`, `kubectl`, `helm`, Go, Node) and all flags are documented in [`local-dev.md`](local-dev.md).

### Building individual services

Each service directory has a `Makefile` with standard targets (`make build`, `make test`, `make lint`, `make validate`) and a `README.md` describing that service. See the per-service docs for details.

## Contributing

Contributions are welcome. Please read [`CONTRIBUTING.md`](CONTRIBUTING.md) for the development workflow, branch and PR policy, and CI expectations, and [`CODE_OF_CONDUCT.md`](CODE_OF_CONDUCT.md) for community guidelines.

## Security & Support

- **Reporting a vulnerability:** see [`SECURITY.md`](SECURITY.md) — do not open public issues for security reports.
- **Supported versions and getting help:** see [`SUPPORT.md`](SUPPORT.md).

## License

Licensed under the Apache License, Version 2.0 — see [`LICENSE`](LICENSE). Third-party attributions are listed in [`NOTICE`](NOTICE) and in the per-component `NOTICE` files.

### Nested and third-party licenses

Some subdirectories contain vendored third-party code or bundled assets under their own license. Each is listed below with its license and the path to that license file:

| Subdirectory | License | License file |
|--------------|---------|--------------|
| `skills/` | Apache-2.0 | [`skills/LICENSE`](skills/LICENSE) |
| `runai-cluster/cluster/goofys/` | Apache-2.0 | [`runai-cluster/cluster/goofys/LICENSE.md`](runai-cluster/cluster/goofys/LICENSE.md) (attributions in the adjacent `NOTICE`) |
| `control-plane/assets-service/third_party/go_acc/` | Apache-2.0 | [`.../go_acc/LICENSE`](control-plane/assets-service/third_party/go_acc/LICENSE) |
| `control-plane/workloads-and-policies/third_party/kserve/` | Apache-2.0 | [`.../kserve/LICENSE`](control-plane/workloads-and-policies/third_party/kserve/LICENSE) |
| `runai-application-packages/jsonutils/conjungo/` | MIT | [`.../conjungo/LICENSE`](runai-application-packages/jsonutils/conjungo/LICENSE) |
| `runai-application-packages/jsonutils/gojsonq/` | MIT | [`.../gojsonq/LICENSE.md`](runai-application-packages/jsonutils/gojsonq/LICENSE.md) |
| `runai-application-packages/validation/` | MIT | [`.../validation/LICENSE`](runai-application-packages/validation/LICENSE) |
| `runai-cluster/cluster/workload-controller/third_party/validation/` | MIT | [`.../third_party/validation/LICENSE`](runai-cluster/cluster/workload-controller/third_party/validation/LICENSE) |
| `control-plane/frontend/public/static-cli/fonts/` | Apache-2.0 | [`.../fonts/LICENSE.txt`](control-plane/frontend/public/static-cli/fonts/LICENSE.txt) |

The following subdirectories ship an aggregated third-party **dependency** license report (not a license for their own code):

| Subdirectory | Report |
|--------------|--------|
| `runai-cluster/cluster/init-ca/licenses/` | [`THIRD-PARTY-LICENSES.md`](runai-cluster/cluster/init-ca/licenses/THIRD-PARTY-LICENSES.md) (generated by `go-licenses`) |
| `runai-cluster/cluster/karta-migrate/licenses/` | placeholder (`dummy.txt`) — to be populated with a generated dependency report |
