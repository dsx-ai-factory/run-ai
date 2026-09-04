# User Identity in Containers

The identity of the user inside a container determines its access to various resources. For example, network file systems often rely on this identity to control access to mounted volumes. As a result, propagating the correct user identity into a container is crucial for both functionality and security.

By default, containers in both Docker and Kubernetes run as the `root` user. This means any process inside the container has full administrative privileges, capable of modifying system files, installing packages, or changing configurations.

While this level of access provides researchers with maximum flexibility, it conflicts with modern enterprise security practices. If the container’s root identity is propagated to external systems (e.g., network-attached storage), it can result in elevated permissions outside the container, increasing the risk of security breaches.

## NVIDIA Run:ai Controls for User Identity and Privileges

NVIDIA Run:ai allows you to enhance security and enforce organizational policies by:

* Controlling root access and privilege escalation within containers
* Propagating the user identity to align with enterprise access policies

## Root Access and Privilege Escalation

NVIDIA Run:ai supports security-related workload configurations to control user permissions and restrict privilege escalation. These options are available via the API and CLI during workload creation:

* `runAsNonRoot` / `--run-as-user` - Force the container to run as non-root user.
* `allowPrivilegeEscalation` / `--allow-privilege-escalation` - Allow the container to use `setuid` binaries to escalate privileges, even when running as a non-root user. This setting can increase security risk and should be disabled if elevated privileges are not required.&#x20;
* `privileged` - Grant full host access to the container, bypassing isolation and security restrictions. By system policy, this parameter is always set to `false` to prevent containers from gaining privileged access. Only an administrator can override this restriction by explicitly creating a policy to override the system policy. This ensures that privileged access is granted only through a deliberate, audited action. See [System policies](../../../platform-management/policies/policies-and-rules.md#system-policies) for more details.

Administrators can enforce secure defaults across the environment using [Policies](../../../platform-management/policies/policy-yaml-reference.md), ensuring consistent workload behavior aligned with organizational security practices.&#x20;

<Info>

**Note**

If both `privileged` and `allowPrivilegeEscalation` are set to `true`, the `allowPrivilegeEscalation` setting becomes redundant. Enabling `privileged` gives the container full host access, which already includes all privilege escalation capabilities.

</Info>

## Passing User Identity

### Passing User Identity from Identity Provider

A best practice is to store the **User Identifier (UID)** and **Group Identifier (GID)** in the organization's directory. NVIDIA Run:ai allows you to pass these values to the container and use them as the container identity. To perform this, you must set up [single sign-on](../../authentication/overview.md) and perform the steps for UID/GID integration.

### Passing User Identity via UI

It is possible to explicitly pass user identity when creating an [environment](../../../workloads-in-nvidia-run-ai/workload-assets/environments.md) or submitting a [workload](../../../workloads-in-nvidia-run-ai/workloads.md):&#x20;

* **From the image** - Use the UID/GID defined in the container image.
* **From the IdP token** - Use identity attributes provided by the SSO identity provider (available only in SSO-enabled installations).
* **Custom** - Manually set the **User ID (UID)**, **Group ID (GID)** and **supplementary groups** that can run commands in the container.

Administrators can enforce secure defaults across the environment using [Policies](../../../platform-management/policies/policy-yaml-reference.md), ensuring consistent workload behavior aligned with organizational security practices.&#x20;

<Info>

**Note**

It is also possible to set the above using the API or CLI.

</Info>

## Using OpenShift or Gatekeeper to Provide Cluster Level Controls

In OpenShift, Security Context Constraints (SCCs) manage pod-level security, including root access. By default, containers are assigned a random non-root UID, and flags such as `--run-as-user` and `--allow-privilege-escalation` are disabled.

On non-OpenShift Kubernetes clusters, similar enforcement can be achieved using tools like [Gatekeeper](https://open-policy-agent.github.io/gatekeeper/website/docs/), which applies system-level policies to restrict containers from running as root.

## Enabling UID and GID on OpenShift

By default, OpenShift restricts setting specific user and group IDs (UIDs/GIDs) in workloads through its SCCs. To allow NVIDIA Run:ai workloads to run with explicitly defined UIDs and GIDs, a cluster administrator must modify the relevant SCCs.

To enable UID and GID assignment:

1.  Edit the `runai-user-job` SCC:

    ```bash
    oc edit scc runai-user-job
    ```
2.  Edit the `runai-jupyter-notebook` SCC (only required if using Jupyter environments):

    ```bash
    oc edit scc runai-jupyter-notebook
    ```
3.  In both SCC definitions, ensure the following sections are configured:

    ```yaml
    runAsUser:
      type: RunAsAny
    supplementalGroups:
      type: RunAsAny
    ```

These settings allow NVIDIA Run:ai to pass specific UID and GID values into the container, enabling compatibility with identity-aware file systems and enterprise access controls.

## Creating a Temporary Home Directory

When containers run as a specific user, the user must have a home directory defined within the image. Otherwise, starting a shell session will fail due to the absence of a home directory.

Since pre-creating a home directory for every possible user is impractical, NVIDIA Run:ai offers the `createHomeDir` / `--create-home-dir` option. When enabled, this flag creates a temporary home directory for the user inside the container at runtime. By default, the directory is created at `/home/<username>`.

<Info>

**Note**

* This home directory is temporary and exists only for the duration of the container's lifecycle. Any data saved in this location will be lost when the container exits.
* By default, this flag is set to `true` when `--run-as-user` is enabled, and `false` otherwise.

</Info>
