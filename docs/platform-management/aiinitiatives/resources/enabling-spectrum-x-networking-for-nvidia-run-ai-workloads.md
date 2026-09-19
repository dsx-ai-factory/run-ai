# Enabling Spectrum-X Networking for NVIDIA Run:ai Workloads

NVIDIA Spectrum-X is an AI-optimized Ethernet networking platform designed to deliver high throughput and predictable performance for large-scale, multi-node GPU workloads.

Leveraging Spectrum-X with NVIDIA Run:ai extends these benefits into day-to-day AI operations. NVIDIA Run:ai streamlines workload submission and management, enabling workloads to be scheduled in a way that takes advantage of Spectrum-X–enabled infrastructure and improves scale-out efficiency. At the same time, administrators can define policies and apply best practices that direct eligible workloads to the appropriate network configuration, helping maintain consistent operations aligned with network capabilities designed for AI at scale. See [NVIDIA Spectrum-X Ethernet Networking Platform](https://www.nvidia.com/en-eu/networking/spectrumx/) for more details.

## Prerequisites

Before using Spectrum-X with NVIDIA Run:ai, ensure the following components are installed and configured:

* NVIDIA Network Operator version 26.1.0. See the [NVIDIA Network Operator](../../../installation-and-configuration/installation/install-using-helm/prerequisites-and-features.md#nvidia-network-operator) section for installation instructions.
* NVIDIA Spectrum-X Operator version 2.1, installed via the Network Operator.

## Submitting a Workload with Spectrum-X

To leverage Spectrum-X networking with NVIDIA Run:ai, the workload must be configured to use the Spectrum-X network and request the required networking capabilities and resources.

These settings can be applied when submitting [NVIDIA Run:ai native workloads](../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md) or workloads submitted [via YAML](../../../workloads-in-nvidia-run-ai/submit-via-yaml.md).

### Required Configuration

* **Add the required Linux capabilities** - Configure the workload container to include the `IPC_LOCK` Linux capability. This provides the necessary networking and system permissions required for Spectrum-X, without granting full root privileges:
* **Request extended resources** - Request the appropriate NVIDIA extended resource and quantity according to your Spectrum-X configuration. The resource name is taken from the `spec.resourceName` field of the `OVSNetwork`. Use this value as the resource key when requesting resources for the workload (for example, `nvidia.com/sriov_resource`).
* **Attach the Spectrum-X network(s)** - Add the required Kubernetes annotation, `k8s.v1.cni.cncf.io/networks`, to attach one or more Spectrum-X networks to the workload. Each network name in the annotation must match the `metadata.name` of an existing `OVSNetwork` resource.

### YAML Example

The following example shows only the relevant sections of a workload manifest. It illustrates where to define the annotation, capabilities, and extended resource.

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

## Enforcing Spectrum-X Requirements Using Workload Policies

Administrators can use NVIDIA Run:ai [workload policies](../../policies/native-workload-policies.md) to ensure that workloads intended to run on Spectrum-X infrastructure are configured with the required capabilities, resources, and network attachments. For full policy structure and supported fields, see [Policy YAML Reference](../../policies/policy-yaml-reference.md).

<Info>

**Note**

Workload policies are supported for [NVIDIA Run:ai native workloads](../../../workloads-in-nvidia-run-ai/workload-types/native-workloads.md) only.

</Info>

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
