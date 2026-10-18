# Security Best Practices

This guide provides actionable best practices for administrators to securely configure, operate, and manage NVIDIA Run:ai environments. Each section highlights both platform-native features and mapped Kubernetes security practices to maintain robust protection for workloads and resources.

| Security Area                             | Best Practice                                                                              |
| ----------------------------------------- | ------------------------------------------------------------------------------------------ |
| Access control (RBAC)                     | Enforce least privilege, segment roles by scope, audit regularly                           |
| Authentication and sessions management    | Use SSO, token-based authentication, strong passwords, limit idle time                     |
| Workload policies                         | Require non-root, set UID/GID, block overrides, use trusted images                         |
| Namespace and resource management         | Require namespace approval, limit secret propagation, apply quotas                         |
| Tools and serving endpoint access control | Control who can access tools and endpoints; restrict network exposure                      |
| Maintenance and compliance                | Follow secure install guides, perform vulnerability scans, maintain data-privacy alignment |

## Access Control (RBAC)

NVIDIA Run:ai uses Role‑Based Access Control to define what each user, group, or service account can do, and where. Roles are assigned within a scope, such as a project, department, or cluster, and permissions cover actions like viewing, creating, editing, or deleting entities. Unlike Kubernetes RBAC, NVIDIA Run:ai’s RBAC works across multiple clusters, giving you a single place to manage access rules. See [Role Based Access Control (RBAC)](../../../platform-management/authentication/overview.md#role-based-access-control-rbac-in-nvidia-run-ai) for more details.

### Best Practices

* Assign the minimum required permissions to users, groups and service accounts.
* Segment duties using organizational scopes to restrict roles to specific projects or departments.
* Regularly audit access rules and remove unnecessary privileges, especially admin-level roles.

### Kubernetes Connection

NVIDIA Run:ai predefined roles are automatically mapped to Kubernetes cluster roles (also predefined by NVIDIA Run:ai). This means administrators do not need to manually configure role mappings.

These cluster roles define permissions for the entities NVIDIA Run:ai manages and displays (such as workloads) and also apply to users who access cluster data directly through Kubernetes tools (for example, `kubectl`).

## Authentication and Session Management

NVIDIA Run:ai supports several authentication methods to control platform access. You can use single sign-on (SSO) for unified enterprise logins, traditional username/password accounts if SSO isn’t an option, and API secret keys for automated application access. Authentication is mandatory for all interfaces, including the UI, CLI, and APIs, ensuring only verified users or applications can interact with your environment.

Administrators can also configure session timeout. This refers to the period of inactivity before a user is automatically logged out. Once the timeout is reached, the session ends and re‑authentication is required, helping protect against risks from unattended or abandoned sessions. See [Authentication and authorization](../../../platform-management/authentication/overview.md) for more details.

### Best Practices

* Integrate corporate SSO for centralized identity management.
* Enforce strong password policies for local accounts.
* Set appropriate session timeout values to minimize idle session risk.
* Prefer SSO to eliminate password management within NVIDIA Run:ai.

### Kubernetes Connection

Configure the Kubernetes API server to validate tokens via NVIDIA Run:ai’s identity service, ensuring unified authentication across the platform. For more information, see [Cluster authentication](../cluster-authentication.md).

## Workload Policies: Enforcing Security at Submission

Workload policies allow administrators to define and enforce how AI workloads are submitted and controlled across projects and teams. With these policies, you can set clear rules and defaults for workload parameters such as which resources can be requested, required security settings, and which defaults should apply. Policies are enforced whether workloads are submitted via the UI, CLI, API or Kubernetes YAML, and can be scoped to specific projects, departments, or clusters for fine-grained control. See [Policies and rules](../../../platform-management/policies/policies-and-rules.md) for more details.

### Best Practices

* Enforce containers to run as non-root by default. Define policies that set constraints and defaults for workload submissions, such as requiring non-root users or specifying minimum UID/GID. Example security fields in policies:
  * `security.runAsNonRoot: true`
  * `security.runAsUid: 1000`
  * Restrict `runAsUid` with `canEdit: false` to prevent users from overriding.
* Require explicit user/group IDs for all workload containers.
* Impose data source and resource usage limits through policies.
  * Control `hostPath` volumes. A `hostPath` volume mounts a directory from the node into the workload container. Workloads can add `hostPath` volumes unless a policy restricts them. To prevent workloads from adding them, set `storage.hostPath.instances.canAdd: false`.
  * When a project requires `hostPath` volumes, define them in the policy and set `readOnly: true` where possible. Use `locked` to keep the defined volumes in place, and `canAdd: false` to prevent additional ones.
  * Combine `hostPath` controls with `security.runAsNonRoot: true` and a locked `security.runAsUid` to control the user identity that accesses mounted paths.
  * See [Policy YAML examples](../../../platform-management/policies/policy-yaml-examples.md) and [Policy YAML reference](../../../platform-management/policies/policy-yaml-reference.md).
* Use policy rules to prevent users from submitting non-compliant workloads.
* Apply policies by organizational scope for nuanced control within departments or projects.

### Kubernetes Connection

Map these policies to `PodSecurityContext` settings in Kubernetes, and enforce them with Pod Security Admission or Kyverno for stricter compliance.

## Managing Namespace and Resource Creation

NVIDIA Run:ai offers flexible controls for how namespaces and resources are created and managed within your clusters. When a new project is set up, you can choose whether Kubernetes namespaces are created automatically, and whether users are auto-assigned to those projects. There are also options to manage how secrets are propagated across namespaces and to enable or disable resource limit enforcement using Kubernetes LimitRange objects. See [Advanced cluster configurations](../cluster-config.md) for more details.

### Best Practices

* Require admin approval for namespace creation to avoid sprawl.
* Limit secret propagation to essential cases only.
* Use Kubernetes `LimitRanges` and `ResourceQuotas` alongside NVIDIA Run:ai policies for layered resource control.
* Regularly audit and remove unused namespaces, secrets, and workloads.

## Tools and Serving Endpoint Access Control

NVIDIA Run:ai provides flexible options to control access to tools and serving endpoints. Access can be defined during workload submission or updated later, ensuring that only the intended users or groups can interact with the resource.

When configuring an endpoint or tool, users can select from the following access levels:

* **Public** - Everyone within the network can access with no authentication (serving endpoints).
* **All authenticated users** - Access is granted to anyone in the organization who can log in (NVIDIA Run:ai or SSO).
* **Specific groups** - Access is restricted to members of designated identity provider groups.
* **Specific users** - Access is restricted to individual users by email or username.

By default, network exposure is restricted, and access must be explicitly granted. Model endpoints automatically inherit RBAC and workload policy controls, ensuring consistent enforcement of role- and scope-based permissions across the platform. Administrators can also limit who can deploy, view, or manage endpoints, and should open network access only when required.

### Best Practices

* Define explicit roles for model management/use.
* Restrict endpoint access to authorized users, groups and applications.
* Monitor and audit endpoint access logs.

### Kubernetes Connection

Use Kubernetes `NetworkPolicies` to limit inter-pod and external traffic to model-serving pods. Pair with NVIDIA Run:ai RBAC for end-to-end control.

## Secure Installation and Maintenance

A secure deployment is the foundation on which all other controls rest, and NVIDIA Run:ai’s installation procedures are built to align with organizational policies such as OpenShift Security Context Constraints (SCC). See [Advanced cluster configurations](../cluster-config.md) for more details.

* Deploy NVIDIA Run:ai cluster following secure installation guides (including IT compliance mandates such as SCC for OpenShift).
* Run regular security scans and patch/update NVIDIA Run:ai deployments promptly when vulnerabilities are reported.
* Regularly review and update all security policies, both at the NVIDIA Run:ai and Kubernetes levels, to adapt to evolving risks.

## Compliance and Data Privacy

NVIDIA Run:ai keeps all models, logs, and user data entirely within your infrastructure, satisfying strict data security needs.

* Maintain full control over the environment, including deployment and day-2 operations such as upgrades, monitoring, backup, and metadata restore.
* Encrypt secrets and sensitive resources; control secret propagation.
* Document and audit data flows for regulatory alignment.
