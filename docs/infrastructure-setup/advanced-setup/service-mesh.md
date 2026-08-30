# Service Mesh

NVIDIA Run:ai supports service mesh implementations. When a service mesh is deployed with sidecar injection, specific configurations must be applied to ensure compatibility with NVIDIA Run:ai. This document outlines the required changes for the NVIDIA Run:ai control plane and cluster.

## Control Plane Configuration

By default, NVIDIA Run:ai prevents Istio from injecting sidecar containers into system jobs in the control plane. For other service mesh solutions, users must manually add annotations during installation.

To disable sidecar injection in the NVIDIA Run:ai control plane, modify the Helm values file by adding the required pod labels to the following components. See [Advanced control plane configurations](../../installation-and-configuration/post-installation-configuration/control-plane-config.md) for more details.

Example for [Open Service Mesh](https://release-v0-11.docs.openservicemesh.io/docs/guides/app_onboarding/sidecar_injection/#explicitly-disabling-automatic-sidecar-injection-on-pods):

```yaml
authorizationMigrator:
  podLabels:
    openservicemesh.io/sidecar-injection: disabled
clusterMigrator:
  podLabels:
    openservicemesh.io/sidecar-injection: disabled
identityProviderReconciler:
  podLabels:
    openservicemesh.io/sidecar-injection: disabled
keepPVC:
  podLabels:
    openservicemesh.io/sidecar-injection: disabled
orgUnitsMigrator:
  podLabels:
    openservicemesh.io/sidecar-injection: disabled
```

## Cluster Configuration

### Installation Phase

Sidecar containers injected by some service mesh solutions can prevent NVIDIA Run:ai installation hooks from completing. To avoid this, modify the Helm installation command to include the required labels or annotations:

```bash
helm upgrade -i ... 
--set global.additionalJobLabels.A=B --set global.additionalJobAnnotations.A=B
```

Example for [Istio Service Mesh](https://istio.io/latest/docs/setup/additional-setup/sidecar-injection/#controlling-the-injection-policy):

```bash
helm upgrade -i ... 
--set-json global.additionalJobLabels='{"sidecar.istio.io/inject":false}'
```

### Workloads

To prevent sidecar injection in workloads created at runtime (such as training workloads), update the `runaiconfig` resource. See [Advanced cluster configurations](../../installation-and-configuration/post-installation-configuration/cluster-config.md) for more details:

```yaml
spec:
  workload-controller:
    additionalPodLabels:
      sidecar.istio.io/inject: false
```
