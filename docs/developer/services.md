# NVIDIA Run:ai Service Reference

For an overview of how these services fit together, see the [Architecture and Components](architecture.md).

## Control Plane Services

The control plane is a set of independently deployable microservices. Services communicate internally over a pub/sub event bus (NATS or Redis Streams), grouped around five functional domains:

- **Identity and access** - handles authentication, JWT validation, identity provider integration (OIDC, SAML, OCP), and RBAC enforcement across all services: `gateway-hub`, `identity-manager`, `authorization`, `keycloak-extensions`
- **Organizational structure** - manages the hierarchy of tenants, departments, and projects that scopes resource quotas and access rules: `org-unit-service`, `tenants-manager`
- **Workload lifecycle** - handles workload submission, policy enforcement, asset management (environments, credentials, compute resources), persistent storage, and the model/image catalog: `workloads-manager`, `assets-service`, `policy-service`, `datavolumes-service`, `catalog-service`
- **Observability** - collects and serves GPU and workload metrics, audit logs, and event-driven notifications across channels: `metrics-service`, `audit-service`, `notifications-service`, `notifications-proxy`
- **Platform delivery** - serves the web UI, CLI binaries, and the cluster-facing APIs that track node and Kubernetes object state: `frontend`, `bff-service`, `cli-exposer`, `cluster-service`, `k8s-objects-tracker`

| Service | Description |
|---------|-------------|
| [assets-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/assets-service) | Manages reusable workload assets (environments, compute resources, credentials, data sources); exposes a REST API and enforces policy and authorization constraints when assets are created or referenced. |
| [audit-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/audit-service) | Ingests user and system action events from a pub/sub bus, persists them in a database, and serves audit log query results via REST; purges records when a tenant is deleted. |
| [authorization](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/authorization) | Enforces RBAC across the platform by managing roles, permission sets, and access rules scoped to org units and clusters; exposes an authorization API consumed by every other control-plane service. |
| [bff-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/bff-service) | Backend-for-Frontend gateway that consolidates UI-facing reads and writes for tenant-scoped settings (branding, notifications, workload list defaults) and exposes an unauthenticated boot-config endpoint the browser needs before login. |
| [catalog-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/catalog-service) | Manages catalogs of container image registries (including NGC), Helm chart repositories, and Hugging Face model entries; proxies image and chart discovery queries to upstream sources. |
| [cli-exposer](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/cli-exposer) | Serves and distributes versioned NVIDIA Run:ai CLI binaries to end users and automation. |
| [cluster-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/cluster-service) | Maintains control-plane state for cluster-scoped resources, currently node inventory; syncs node statuses from connected clusters and serves them via REST. |
| [datavolumes-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/datavolumes-service) | Manages data volumes (persistent storage backed by Kubernetes PVCs) that users can create and attach to workloads; emits pub/sub events on volume lifecycle changes. |
| [frontend](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/frontend) | Vue 3 single-page application that serves as the browser-based UI for the NVIDIA Run:ai control plane. |
| [gateway-hub](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/gateway-hub) | API gateway sidecar (alongside Traefik) that validates Bearer JWTs, resolves tenant IDs from request subdomains, and guards the Keycloak OAuth endpoint against open-redirect attacks. |
| [identity-manager](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/identity-manager) | Manages per-tenant Keycloak realms, external identity providers (OIDC, SAML, OCP), OAuth2 clients, service accounts, local users, and token issuance and revocation. |
| [k8s-objects-tracker](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/k8s-objects-tracker) | Watches Kubernetes objects (PVCs and Secrets) in connected clusters and maintains an up-to-date control-plane representation, making cluster-local objects queryable from the control plane. |
| [keycloak-extensions](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/keycloak-extensions) | Sidecar image that copies custom Keycloak Java SPI providers, a Run:ai login and email theme, and the OpenTelemetry Java agent into a running Keycloak instance at startup. |
| [metrics-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/metrics-service) | Queries Prometheus metric stores for GPU and workload metrics, executes PromQL queries per tenant and cluster, persists scheduled reports to S3 and PostgreSQL, and serves results via REST. |
| [notifications-proxy](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/notifications-proxy) | Enforces authorization on incoming notification requests and forwards authorized events to notifications-service; cleans up state on tenant deletion. |
| [notifications-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/notifications-service) | Ingests platform events from HTTP, SMTP, Kubernetes, and Alertmanager; routes and renders them via configurable rules and templates; dispatches to email and Slack. |
| [org-unit-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/org-unit-service) | Manages the organizational unit hierarchy (departments and projects), persists quota and resource configuration per org unit, and syncs resource state to clusters. |
| [policy-service](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/policy-service) | Stores workload policies per scope (tenant/department/project), computes effective policy by merging the hierarchy, validates workload specs for compliance, and syncs effective policies to cluster agents. |
| [tenants-manager](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/tenants-manager) | Manages tenant lifecycle (creation, deletion, configuration) in multi-tenant deployments; delegates identity management to identity-manager and publishes tenant lifecycle events. |
| [workloads-manager](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/workloads-manager) | Provides the control-plane REST API for workload submission and lifecycle management; enforces policy compliance, generates Kubernetes manifests, and publishes workload events to cluster agents. |
| [runai-cli](https://github.com/dsx-ai-factory/run-ai/tree/main/control-plane/runai-cli) | CLI v2 for researchers and operators to submit and manage workloads, configure projects, and query cluster state; distributed via cli-exposer. |

---

## Cluster Components

Source: [`dsx-ai-factory/run-ai/runai-cluster/cluster/`](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster)

Each managed Kubernetes cluster runs a set of controllers and webhooks installed and managed by the `go-operator`:

- **Data sync** - the two-channel bridge between cluster and control plane: `agent` pulls desired state down and reconciles it into cluster resources; `cluster-sync` watches live Kubernetes objects and pushes their state back up
- **Workload lifecycle** - translates Run:ai workload CRDs into Kubernetes resources and manages their full lifecycle (creation, status, teardown): `anyworkload-controller`
- **Scheduling enforcement** - intercepts pod admission and applies runtime time-limit policies: `admission-controller`, `runtime-rules`
- **RBAC and project isolation** - keeps cluster-side namespaces, queues, and role bindings in sync with the control plane's access rules and project definitions: `accessrule-controller`, `shared-objects-controller`

| Component | Description |
|-----------|-------------|
| [agent](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/agent) | Pulls desired state for all resource kinds (projects, workloads, access rules, node pools, credentials) from the control plane and reconciles each into the cluster via pluggable resource handlers; also runs a TLS validating webhook for project and secret-PVC admission. |
| [admission-controller](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/admission-controller) | Mutating and validating admission webhook server that intercepts pod creation and modification requests to inject NVIDIA Run:ai scheduling attributes (labels, annotations, resource patches) before pods are admitted. |
| [cluster-api](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/cluster-api) | Cluster-side HTTP REST API server, secured with OIDC/JWT authentication, that exposes programmatic access to cluster workloads and streaming endpoints for the Run:ai UI and CLI. |
| [cluster-sync](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/cluster-sync) | Runs reconcilers for all cluster resource kinds (nodes, pods, projects, workloads, PVCs, policies) that watch Kubernetes objects and push live state to the Run:ai control plane over authenticated HTTP. |
| [external-workload-integrator](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/external-workload-integrator) | Watches third-party (non-Run:ai) workload types and creates an ExternalWorkload CRD for each, which cluster-sync then propagates to the control plane. |
| [go-operator](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/go-operator) | Kubernetes operator that owns the RunaiConfig CRD and manages the full lifecycle (deploy, configure, update) of all NVIDIA Run:ai cluster components. |
| [init-ca](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/init-ca) | Bootstrap service that generates the TLS CA certificate for Run:ai cluster webhooks, injects the CA bundle into Kubernetes webhook configurations, and rotates it every 30 days. |
| [metrics-exporter](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/metrics-exporter) | Watches cluster resources (Pods, Queues, Nodes) for state changes and exports scheduling and utilization metrics in Prometheus format. |
| [oauth2-proxy](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/oauth2-proxy) | Packages the upstream oauth2-proxy with a custom Run:ai error page; provides OAuth2 authentication proxying for interactive workspaces such as Jupyter. |
| [shared-objects-controller](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/shared-objects-controller) | Replicates shared objects (primarily PVCs from DataVolume CRDs) into each project or department namespace that references them. |
| [status-updater](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/status-updater) | Watches pods and aggregates their state to compute and patch the status of Run:ai workload objects. |
| [accessrule-controller](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/accessrule-controller) | Reconciles AccessRule CRDs to native Kubernetes RoleBinding objects, ensuring each access rule is reflected as a RoleBinding in the corresponding project namespace. |
| [advice-restart-controller](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/advice-restart-controller) | Watches Heimdall Advice CRs and restarts distributed workloads (Kubeflow training operators) based on scheduling recommendations, enforcing configurable per-workload restart limits. |
| [anyworkload-controller](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/anyworkload-controller) | Thin binary that wires Run:ai-specific plugins (networking, credentials) and generic dromos plugins (storage) into the dromos WorkloadRunner controller. |
| [application-controller](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/application-controller) | Manages Application CRs in two modes: tracks existing deployed resources and reflects their state into Application status, or owns the full install/upgrade/uninstall lifecycle of a Helm chart. |
| [runtime-rules](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/runtime-rules) | Enforces time-based runtime policies on cluster workloads, terminating jobs that exceed configured idle or time limits for interactive and training workload types. |
| [workload-exporter](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-cluster/cluster/workload-exporter) | Exports Prometheus metrics for Run:ai workloads at the workload level (replaces the legacy pod-group metrics-exporter as of v2.17), covering GPU allocation, GPU memory, CPU, and memory per pod. |

---

## Installer

Source: [`dsx-ai-factory/run-ai/runai-installer/`](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-installer)

| Component | Description |
|-----------|-------------|
| [installer](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-installer/installer) | CLI tool that reads a YAML config file and orchestrates a full NVIDIA Run:ai installation on Kubernetes; installs prerequisites (GPU Operator, Ingress, Prometheus, Network Operator) and optional features (Knative, Kubeflow, LWS, NIM Operator, Dynamo), then deploys the control-plane and cluster Helm charts. Supports dry-run, preflight-only, and airgap modes. |

---

## MCP Server

Source: [`dsx-ai-factory/run-ai/runai-mcp-server/`](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-mcp-server)

| Component | Description |
|-----------|-------------|
| [runai-mcp-server](https://github.com/dsx-ai-factory/run-ai/tree/main/runai-mcp-server) | MCP (Model Context Protocol) server that exposes NVIDIA Run:ai control-plane operations as tools for AI assistants (Claude Code, Cursor, etc.); supports stdio transport (authenticates via OAuth2 client credentials) and HTTP transport (delegates auth to the MCP client via Bearer token). |
