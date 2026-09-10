# Security Policy

NVIDIA takes the security of our software products and services seriously, including the source code in this repository.

If you believe you have found a security vulnerability, **please do not open a public GitHub issue**. Use one of the private channels below.

## Supported Versions

Security fixes are only issued for currently supported releases. See [Support](SUPPORT.md) for the current list of supported minor versions. End-of-life branches do not receive security updates.

## Reporting a Vulnerability

We accept private reports through the channels below. Any of them is acceptable; use whichever you prefer.

### 1. GitHub Security Advisory (Recommended for Repository-Specific Issues)

Open a private advisory from the **Security** tab of this repository, then **Report a vulnerability**. The advisory is only visible to repository maintainers and the reporter.

This is the preferred channel for vulnerabilities specific to the code in this repository.

### 2. NVIDIA PSIRT (for Cross-Product Issues or if You Prefer Email)

For vulnerabilities that may affect multiple NVIDIA products, or if you prefer to use NVIDIA's coordinated disclosure process:

- Vulnerability Disclosure Program: <https://www.nvidia.com/en-us/security/>
- Web: [NVIDIA Security Vulnerability Submission Form](https://www.nvidia.com/object/submit-security-vulnerability.html)
- Email: <psirt@nvidia.com>
  - PGP key: <https://www.nvidia.com/en-us/security/pgp-key>

Please include in your report:

- Product name and the version, branch, or commit SHA that contains the vulnerability
- Vulnerability class (for example: remote code execution, denial of service, privilege escalation, information disclosure)
- Reproduction steps
- Proof-of-concept code or exploit, if available
- Your assessment of the potential impact

## What to Expect

- **Acknowledgement:** within 3 business days of your report
- **Initial assessment:** within 10 business days
- **Status updates:** at least every 14 days while the issue is open

We will coordinate a public disclosure timeline with you once a fix is available.

## Acknowledgement

NVIDIA does not currently run a paid bug bounty program for this repository. We do publicly acknowledge externally reported security issues that are addressed under our [coordinated disclosure policy](https://www.nvidia.com/en-us/security/psirt-policies/), where the reporter wishes to be named.

## Security Architecture and Context

Run:ai is an AI workload orchestration platform for Kubernetes. This repository contains the full stack: a multi-tenant **control plane** of HTTP services, the **in-cluster** operator and controllers deployed into each managed Kubernetes cluster, shared Go libraries, generated API clients, the installer, and a CLI.

This software operates primarily at the **service** and **Kubernetes controller** level, with additional CLI and library components. Its primary security responsibilities are:

- **Tenant isolation** — keeping each tenant's projects, workloads, assets, and metrics separate in a shared control plane.
- **Authentication and authorization** — establishing caller identity and enforcing a scoped role-based access model across the platform API.
- **Custody of tenant-supplied credentials** — container registry logins, object-storage keys, API keys, Git credentials, and generic secrets that users register as reusable assets and that the platform injects into their workloads.
- **Mediating privileged Kubernetes operations** — translating platform-level API requests into pods, jobs, role bindings, and namespace changes in managed clusters.

**Repository Exposure Classification:** Public.
Basis: `origin` remote is github.com; the project is Apache-2.0 licensed with public contribution and support documentation. This document is written to public-safe detail.

**Service Exposure Classification:** External / Regulated (high confidence).
Basis: commercially released, customer-deployed enterprise platform with a published self-hosted installation guide and support obligations; multi-tenant; handles customer credentials, user identities, and workload data.

### Trust Boundaries

The main boundaries an attacker would have to cross, and what is trusted on each side:

1. **Internet → control-plane ingress.** TLS is terminated at the cluster ingress (ingress-nginx, HAProxy, or an OpenShift Route). Behind it, an internal router performs a forward-auth call to a dedicated authentication component, which validates an OIDC bearer token issued by Keycloak (RSA signature, issuer-to-realm resolution, expiry, non-empty subject and tenant). On success, identity is attached to the upstream request as a fixed allowlist of internal headers.

2. **Edge → control-plane services.** Individual services do **not** re-validate the bearer token. They read identity from the headers injected by the edge and enforce authorization per route against the central authorization service, which returns the set of scopes (system, tenant, cluster, department, project) the caller may act within. This makes the edge the sole point at which identity is established.

3. **Control plane ↔ managed cluster.** In-cluster components authenticate outbound to the control plane using per-cluster OIDC client credentials, provisioned at install time from a bootstrap credential and stored as Kubernetes Secrets. State flows both ways: the cluster reports inventory and workload status up, and applies desired state down.

4. **Platform API → Kubernetes.** User-supplied workload specifications become real Kubernetes objects. Mutating admission webhooks rewrite pod specs (scheduler name, resources, environment, affinity, volumes), and controllers build pods whose images, commands, arguments, and sidecars derive from user input. An in-cluster API additionally proxies interactive `logs`, `exec`, `attach`, and `port-forward` sessions into workload pods over WebSockets.

5. **Workload → platform.** Containers submitted by end users are untrusted code. Isolation between a workload and both the platform and other tenants' workloads is delegated to Kubernetes namespaces, RBAC, and the container runtime.

### Threat Model

The following are the primary security concerns for this project, including auxiliary and support paths. They are ordered roughly by combined severity and likelihood.

1. **Cross-tenant data exposure through incomplete query scoping.** Tenant isolation in the control-plane databases is enforced per query — each data-access method is responsible for constraining results to the caller's tenant and permitted scopes. There is no database-level row-level security policy, ORM-wide tenant callback, or shared base model that enforces a tenant predicate as a backstop. A single data-access path that omits or incorrectly derives the tenant or scope filter yields a cross-tenant read or write. The scope-resolution helpers in the authorization client library are correspondingly high-value review targets.

2. **Identity spoofing by bypassing the authenticating edge.** Because downstream services trust the identity headers injected by the edge rather than re-validating the token, any actor able to reach a control-plane service directly — a compromised pod, a misconfigured network policy, a route registered without the authentication middleware, or an ingress path that reaches a service without traversing the forward-auth check — can present arbitrary tenant, subject, and group identity. The blast radius is full impersonation across tenants, so the completeness of the middleware attachment and of pod-to-pod network restrictions is a critical control.

3. **Escalation from a workload specification into unintended cluster execution.** Users control substantial portions of the pod specifications the platform materializes: container images, commands and arguments, environment, mounted assets, and certain sidecars. Several supporting paths compose shell command lines or auxiliary container configuration from user-influenced values, and the object-storage mounting sidecar runs with elevated container privileges in order to perform FUSE mounts. Weaknesses in the parsing, quoting, or validation of these fields, or in the admission webhooks that rewrite them, can let a tenant influence execution beyond their intended workload or obtain more privilege on the node than the workload should carry.

4. **Compromise of stored tenant credentials.** The platform is a deliberate concentration point for third-party secrets: registry credentials, object-storage access keys, Git credentials, API keys, generic secrets, per-tenant identity-provider client secrets, and messaging-integration credentials. Credential assets are encrypted at the application layer with per-tenant asymmetric keys whose private halves are held by the same service in Kubernetes Secrets, without an external key-management service or envelope encryption; field-level symmetric encryption exists in the shared database layer but is applied to only a small number of models, leaving some integration credentials at rest in plaintext columns. Database compromise, Secret read access, or an authorization flaw on the asset APIs therefore exposes credentials to systems outside the platform.

5. **Interactive session proxying as a route to workload and cluster access.** The in-cluster API exposes internet-reachable endpoints that proxy `logs`, `exec`, `attach`, and `port-forward` into pods. Authorization for these is decided from a bearer token verified against the control plane's key set, combined with Kubernetes access review. Errors in audience or issuer binding, in the scope check, or in the WebSocket upgrade path translate directly into shell access inside another tenant's workload. Origin checking on the WebSocket upgrade is intentionally disabled to support the split control-plane/cluster domains, which places the full weight of the control on token validation.

6. **Theft or misuse of cluster registration credentials.** Cluster-to-control-plane authentication rests on OIDC client secrets distributed at install time and persisted in Kubernetes Secrets on the managed cluster, and the installation command itself carries a bootstrap credential. No automated rotation path exists. An attacker who obtains these can impersonate a managed cluster to the control plane — reporting false inventory and status, or receiving desired state intended for that cluster. Several client components can additionally be configured to skip TLS certificate verification, and in some cases do so by mutating the process-wide default HTTP transport, which widens the effect of that setting beyond the call site that set it.

7. **Availability and supply-chain exposure in the delivery path.** No application-level rate limiting or request throttling is present on the platform API, including on authentication and on expensive list and query endpoints, so protection against abuse is delegated entirely to infrastructure. Separately, the build and installation tooling fetches components over the network with inconsistent verification: most runtime images derive from pinned Red Hat UBI bases, but some development and auxiliary images use floating tags, install tooling from unpinned sources, or download third-party binaries without checksum or signature verification, and installer steps apply manifests from remote URLs. Compromise of one of those upstreams reaches the built artifacts.

### Critical Security Assumptions

The following are conditions this software assumes are already satisfied by another layer. They are not enforced by the code in this repository, and they define the boundary of what a deployment must provide.

- **The authenticating edge is the only path to control-plane services.** Services assume every request has traversed the ingress and forward-auth check, and that the internal identity headers are therefore trustworthy. Deployments must prevent direct pod-to-pod access to control-plane services from untrusted sources; the platform does not independently verify caller identity at each service.

- **The identity provider is trusted and correctly configured.** No password hashing or token issuance occurs in this repository. User authentication, credential storage, session and refresh-token policy, and token signing are wholly delegated to Keycloak. Compromise or misconfiguration of that identity provider — including its realm keys and per-tenant identity-provider federation settings — defeats platform authentication entirely.

- **The in-cluster network is trusted for backend traffic.** TLS is terminated at the ingress and internal service-to-service traffic is not mutually authenticated; there is no client-certificate verification anywhere in the stack. The default database SSL mode is disabled, and TLS is not wired through for all supported message-bus transports. Deployments that require encrypted internal traffic must supply it through the platform or service mesh and set the corresponding configuration explicitly.

- **Kubernetes enforces workload and tenant isolation at runtime.** The platform maps its project and department model onto namespaces, RBAC role bindings, and scheduler queues, and then relies on the Kubernetes control plane, the container runtime, and the node kernel to enforce that separation. It does not provide sandboxing of its own, and it assumes that node-level controls contain a workload that escapes its container — a particularly relevant assumption given that GPU workloads share device-level resources.

- **Cluster operators and the values supplied at install time are trusted.** Helm values, generated secrets, and the operator custom resources are treated as authoritative configuration. Several security-relevant behaviors — TLS verification, proxy settings, admission webhook scope — are configurable, and a deployment that sets them permissively is assumed to have done so deliberately. Anyone with write access to platform configuration or to the platform's namespace is effectively a platform administrator.

- **Infrastructure provides denial-of-service protection.** The application layer does not rate limit, throttle, or apply request quotas. Protection of authentication endpoints and expensive query endpoints against volumetric or resource-exhaustion abuse is assumed to come from the ingress controller, an API gateway, or an upstream CDN or WAF.

- **Audit logging is best-effort, not a control.** Audit records are written asynchronously and a failure to record an event does not fail the underlying request. Audit data is retained for a bounded window by default. Deployments requiring guaranteed non-repudiation must ship these records to an external system with its own durability guarantees.

- **The build and dependency supply chain is verified outside this repository.** Base images, Go modules, JavaScript packages, and tooling downloaded during build and installation are assumed to be vetted by the surrounding CI, registry, and scanning infrastructure. This repository does not itself verify signatures for every artifact it consumes.
