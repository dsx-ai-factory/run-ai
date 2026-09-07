# NVIDIA Run:ai

**NVIDIA Run:ai is an open-source AI workload and GPU orchestration platform** for managing shared AI infrastructure on Kubernetes.

NVIDIA Run:ai provides a consistent way to run development, training, and inference workloads across Kubernetes environments, whether deployed on-premises, in the cloud, or across hybrid infrastructure. It combines GPU scheduling, resource management, workload orchestration, governance, and observability to help organizations share GPU resources efficiently across teams and clusters.

This monorepo contains the NVIDIA Run:ai control plane and cluster services, installer, CLI, MCP server, documentation, shared client packages, and the local development environment.

## Architecture

<p align="center">
  <img src="docs/assets/system-overview.png" alt="NVIDIA Run:ai system overview" width="900" />
</p>

**[Architecture Overview →](docs/developer/architecture.md)**

NVIDIA Run:ai is made up of two components, both installed on a Kubernetes cluster:

- NVIDIA Run:ai control plane - Provides centralized resource management, handles workload submission, and delivers cluster monitoring and analytics. The control plane is the management layer; it connects to one or more NVIDIA Run:ai clusters and gives administrators and users a single place to operate across them.
- NVIDIA Run:ai cluster - Set of cluster services that provide enhanced scheduling and workload management on each Kubernetes cluster where workloads run. The cluster component extends Kubernetes native scheduling with Run:ai’s GPU-aware capabilities.

At a high level:

- Users and external systems interact with NVIDIA Run:ai through the web UI, CLI, or REST API.
- The NVIDIA Run:ai control plane manages workloads, organizational resources, policies, and configuration across connected Kubernetes clusters.
- NVIDIA Run:ai cluster services synchronize configuration and workload requests from the control plane to the appropriate cluster.
- NVIDIA Run:ai applies scheduling and resource policies to determine where and how workloads run on available GPU and compute resources.
- Workload and cluster status is synchronized back to the NVIDIA Run:ai control plane, providing centralized visibility across connected clusters.

### Core Capabilities

| Capability                        | What it does                                                                                                                                |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| **AI-aware GPU scheduling**       | Schedules GPU workloads using fair-share, preemption, bin-packing, priorities, and topology-aware placement.                                |
| **Dynamic GPU allocation**        | Shares GPU resources across workloads with fractional GPU allocation and dynamic use of available capacity.                                 |
| **Resource quotas and policies**  | Allocates GPU resources across departments and projects and enforces organizational resource policies.                                      |
| **Multi-cluster management**      | Centrally manages workloads, resources, and configuration across multiple Kubernetes clusters.                                              |
| **AI workload management**        | Manages workspaces, training, distributed training, and inference, with extensible support for additional Kubernetes workloads.             |
| **Access control and governance** | Controls access to NVIDIA Run:ai resources with role-based permissions and enforces organizational policies across workloads and resources. |
| **Monitoring and observability**  | Provides visibility into clusters, nodes, GPUs, workloads, resource utilization, and platform activity.                                     |
| **Multiple interfaces**           | Supports NVIDIA Run:ai operations through the web UI, CLI, and REST API.                                                                     |

## Repository Map

| Component                       | Path                                                                                                                                                                                                                                                                                  | Description                                                                                                       |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| **Control Plane**               | [`control-plane/`](https://github.com/run-ai/runai/tree/master/control-plane)                                                                                                                                                                                                         | APIs and services that manage workloads, policies, identity, organizational structure, and platform configuration |
| **NVIDIA Run:ai Cluster**       | [`runai-cluster/`](https://github.com/run-ai/runai/tree/master/runai-cluster)                                                                                                                                                                                                         | Controllers, admission webhooks, and services deployed into each managed Kubernetes cluster                       |
| **Installer**                   | [`runai-installer/`](https://github.com/run-ai/runai/tree/master/runai-installer)                                                                                                                                                                                                     | CLI tool and provisioner service for deploying NVIDIA Run:ai on Kubernetes                                        |
| **MCP Server**                  | [`runai-mcp-server/`](https://github.com/run-ai/runai/tree/master/runai-mcp-server)                                                                                                                                                                                                   | Model Context Protocol server that exposes NVIDIA Run:ai operations as tools for AI assistants                    |
| **Clients and shared packages** | [`runai-clients/`](https://github.com/run-ai/runai/tree/master/runai-clients), [`runai-common-packages/`](https://github.com/run-ai/runai/tree/master/runai-common-packages), [`runai-application-packages/`](https://github.com/run-ai/runai/tree/master/runai-application-packages) | Generated API clients and shared Go libraries                                                                     |
| **E2E**                         | [`e2e/`](https://github.com/run-ai/runai/tree/master/e2e)                                                                                                                                                                                                                             | End-to-end test suites                                                                                            |
| **Local environment**           | [`local-dev.sh`](https://github.com/run-ai/runai/blob/master/local-dev.sh), [`local-dev.md`](https://github.com/run-ai/runai/blob/master/local-dev.md)                                                                                                                                | One-command local Kubernetes (Kind) development environment                                                       |

For a detailed breakdown of individual services and components, see the [Service Reference](docs/developer/services.md).

## Usage

### Local development environment

The [`local-dev.sh`](https://github.com/run-ai/runai/blob/master/local-dev.sh) script builds the NVIDIA Run:ai control plane and cluster services from source and deploys them to a self-contained [Kind](https://kind.sigs.k8s.io/) cluster with a local image registry.

```bash
./local-dev.sh                                  # full build + cluster + deploy
./local-dev.sh --skip-build --skip-cluster      # re-deploy only
./local-dev.sh --teardown                       # tear the environment down
```

See [`local-dev.md`](https://github.com/run-ai/runai/blob/master/local-dev.md) for prerequisites (Docker/Podman, `kind`, `kubectl`, `helm`, Go, Node) and all available flags.

### Building individual services

Each service directory has a Makefile with standard targets (`make build`, `make test`, `make lint`, `make validate`) and a README describing how to build, test, and work with that service. See the [Service Reference](docs/developer/services.md) for a breakdown of individual services and components.

## Installation

NVIDIA Run:ai can be deployed on Kubernetes or OpenShift in connected or air-gapped environments. To understand the full setup lifecycle, see the [Installation and Configuration Overview](https://nvidia-preview-01a02a68-bbb9-71d8-b4ab-3579d895216b.docs.buildwithfern.com/run-ai/self-hosted/documentation/installation-and-configuration/overview).

**Before you start:**

- **[Support Matrix](https://nvidia-preview-01a02a68-bbb9-71d8-b4ab-3579d895216b.docs.buildwithfern.com/run-ai/self-hosted/documentation/installation-and-configuration/support-matrix)** - Kubernetes/OpenShift versions, operating systems, GPU and network operators, and third-party services NVIDIA Run:ai supports.
- **[System Requirements](https://nvidia-preview-01a02a68-bbb9-71d8-b4ab-3579d895216b.docs.buildwithfern.com/run-ai/self-hosted/documentation/installation-and-configuration/system-requirements)** - Infrastructure, networking, and access you need to prepare before installing.

Choose an installation method based on how you want to manage the deployment:

- **[Guided installation](https://nvidia-preview-01a02a68-bbb9-71d8-b4ab-3579d895216b.docs.buildwithfern.com/run-ai/self-hosted/documentation/installation-and-configuration/installation/guided-installation-recommended) (recommended)** - Uses the `runai-installer` CLI with a configuration YAML file. The installer performs preflight checks, installs required dependencies, and deploys the NVIDIA Run:ai control plane and cluster.

- **[Manual installation](https://nvidia-preview-01a05288-a018-735b-ad78-e678676b8b99.docs.buildwithfern.com/run-ai/self-hosted/documentation/installation-and-configuration/installation/manual-helm/control-plane-system-requirements)** - Deploys the NVIDIA Run:ai Helm charts directly. Use this path if you manage deployments through your own tooling, such as Argo CD, or require granular control over Helm values.

To build and run NVIDIA Run:ai locally for development, see [Usage](#usage).

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
