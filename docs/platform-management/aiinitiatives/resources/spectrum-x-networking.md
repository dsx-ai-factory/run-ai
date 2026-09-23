# Enabling Spectrum-X Networking for NVIDIA Run:ai Workloads

NVIDIA Spectrum-X is an AI-optimized Ethernet networking platform designed to deliver high throughput and predictable performance for large-scale, multi-node GPU workloads.

Leveraging Spectrum-X with NVIDIA Run:ai extends these benefits into day-to-day AI operations. NVIDIA Run:ai streamlines workload submission and management, enabling workloads to be scheduled in a way that takes advantage of Spectrum-X–enabled infrastructure and improves scale-out efficiency. At the same time, administrators can define policies and apply best practices that direct eligible workloads to the appropriate network configuration, helping maintain consistent operations aligned with network capabilities designed for AI at scale. See [NVIDIA Spectrum-X Ethernet Networking Platform](https://www.nvidia.com/en-eu/networking/spectrumx/) for more details.

## Prerequisites

Before using Spectrum-X with NVIDIA Run:ai, ensure the following components are installed and configured:

* NVIDIA Network Operator version 26.1.0. See the [NVIDIA Network Operator](../../../installation-and-configuration/installation/install-using-helm/prerequisites-and-features.md#nvidia-network-operator) section for installation instructions.
* NVIDIA Spectrum-X Operator version 2.1, installed via the Network Operator.

## Submitting a Workload with Spectrum-X

To leverage Spectrum-X networking with NVIDIA Run:ai, the workload must be configured to use the Spectrum-X network and request the required networking capabilities and resources.

For [NVIDIA Run:ai native workloads](../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md) and [supported workload types](../../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md), you can configure these settings across all interfaces - UI, API, CLI, or MCP. Supported workload types submitted [via YAML](../../../workloads-in-nvidia-run-ai/submit-via-yaml.md) require the settings to be included in the workload YAML manifest.

### Required Configuration

* **Add the required Linux capabilities** - Configure the workload container to include the `IPC_LOCK` Linux capability. This provides the necessary networking and system permissions required for Spectrum-X, without granting full root privileges.
* **Request extended resources** - Request the appropriate NVIDIA extended resource and quantity according to your Spectrum-X configuration. The resource name is taken from the `spec.resourceName` field of the `OVSNetwork`. Use this value as the resource key when requesting resources for the workload (for example, `nvidia.com/sriov_resource`).
* **Attach the Spectrum-X network(s)** - Add the required Kubernetes annotation, `k8s.v1.cni.cncf.io/networks`, to attach one or more Spectrum-X networks to the workload. Each network name in the annotation must match the `metadata.name` of an existing `OVSNetwork` resource.

### Configuring Native Workloads in the UI

In the Workload manager, click **+ NEW WORKLOAD** and select the native workload type you want to create from the dropdown menu. In the workload creation form, configure the required settings as follows:

1. Under **Environment**, expand **Security**. Under **Select additional Linux capabilities for the container**, open the **Capabilities** dropdown and select `IPC_LOCK`.
2. Under **Compute resources**, click **+EXTENDED RESOURCE**. In **Resource**, add the Spectrum-X resource name from `spec.resourceName` in the `OVSNetwork`, such as `nvidia.com/sriov_resource`. In **Quantity**, enter the required number of resources.
3. Under **General**, add an annotation. In **Name**, enter `k8s.v1.cni.cncf.io/networks`. In **Value**, enter the `OVSNetwork` name.

### Configuring Supported Workload Types via YAML

In the Workload manager, click **+ NEW WORKLOAD** and select **Via YAML** from the dropdown menu. In the YAML form, use the following example to see where to define the required annotation, Linux capability, and extended resource in the workload manifest.

```yaml
spec:
  template:
    metadata:
      annotations:
        k8s.v1.cni.cncf.io/networks: <network-name>
    spec:
      containers:
        - name: example-container
          securityContext:
            capabilities:
              add:
                - IPC_LOCK
          resources:
            requests:
              nvidia.com/sriov_resource: 1
            limits:
              nvidia.com/sriov_resource: 1
```

## Applying Spectrum-X Requirements with Workload Policies

### Native Workload Policies

Administrators can use NVIDIA Run:ai [workload policies](../../policies/native-workload-policies.md) to ensure that workloads intended to run on Spectrum-X infrastructure are configured with the required capabilities, resources, and network attachments. For full policy structure and supported fields, see [Policy YAML Reference](../../policies/policy-yaml-reference.md).

**Example workload policy:**

```yaml
defaults:
  annotations:
    instances:
      - name: k8s.v1.cni.cncf.io/networks
        value: <network-name>
  security:
    capabilities:
      - IPC_LOCK
  compute:
    extendedResources:
      instances:
        - resource: nvidia.com/sriov_resource
          quantity: "1"
rules:
  security:
    capabilities:
      canEdit: false
```

### Supported Workload Type Policies

Policies for supported workload types can also be used to enforce Spectrum-X requirements, provided the required workload fields are available in the workload type mapping. Configure the policy using the Rule builder or YAML editor for the selected workload type. Refer to [Supported workload types policies](../../policies/supported-workload-type-policies.md) for details on creating rules and selecting mapped fields.
