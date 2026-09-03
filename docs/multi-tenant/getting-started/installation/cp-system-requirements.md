# Control Plane System Requirements

The NVIDIA Run:ai control plane is a Kubernetes-based application that centrally manages workloads, users, scheduling, and cluster integrations across multiple tenants.

In a multi-tenant deployment, the control plane is installed once, in a dedicated Kubernetes cluster and configured for multi-tenancy. This section outlines the hardware and software system requirements needed to deploy and operate the control plane in a multi-tenant environment.

## Installer Machine

The machine running the installation script (typically the Kubernetes master) must have:

* At least 50GB of free space
* Docker installed
* [Helm](https://helm.sh/) 3.14 or later

<Info>

**Note**

Helm 4 defaults to [server-side apply](https://helm.sh/docs/overview/#server-side-apply) when installing a new chart release, which can conflict with resources managed by the NVIDIA Run:ai operator. Append `--server-side=false` to your `helm upgrade` command. NVIDIA Run:ai clusters originally installed with Helm 3.x are unaffected.

</Info>

## Hardware Requirements

The following hardware requirements are for the control plane system nodes. By default, all NVIDIA Run:ai control plane services run on all available nodes.

### Architecture

**x86** and **ARM** architectures are supported for Kubernetes.

### NVIDIA Run:ai Control Plane - System Nodes

This configuration is the minimum requirement you need to install and use NVIDIA Run:ai control plane:

| Component  | Required Capacity |
| ---------- | ----------------- |
| CPU        | 10 cores          |
| Memory     | 12GB              |
| Disk space | 110GB             |

<Info>

**Note**

* To designate nodes to NVIDIA Run:ai system services, follow the instructions as described in [System nodes](../../infrastructure-setup/advanced-setup/node-roles.md#system-nodes).
* If you are using **Grafana Mimir** for monitoring, we recommend using the **Microservices mode** to properly size your environment. Refer to the following guide for capacity planning, [Planning Grafana Mimir capacity](https://grafana.com/docs/mimir/latest/manage/run-production-environment/planning-capacity/#microservices-mode).

</Info>

## Software Requirements

The following software requirements must be fulfilled.

### Operating System

* Any **Linux** operating system supported by both Kubernetes and NVIDIA GPU Operator
* Internal tests are being performed on **Ubuntu 22.04.**

### Network Time Protocol

Nodes are required to be synchronized by time using NTP (Network Time Protocol) for proper system functionality.

### Kubernetes Distribution

NVIDIA Run:ai control plane requires Kubernetes. The following Kubernetes distributions are supported:

* Vanilla Kubernetes
* NVIDIA Base Command Manager (BCM)
* Elastic Kubernetes Engine (EKS)
* Google Kubernetes Engine (GKE)
* Azure Kubernetes Service (AKS)
* Oracle Kubernetes Engine (OKE)
* Rancher Kubernetes Engine 2 (RKE2)

See the following Kubernetes version support matrix for the latest NVIDIA Run:ai releases:

| NVIDIA Run:ai version | Supported Kubernetes versions |
| --------------------- | ----------------------------- |
| v2.25 (latest)        | 1.34 to 1.36                  |
| v2.25                 | 1.33 to 1.35                  |
| v2.24                 | 1.33 to 1.35                  |
| v2.23                 | 1.31 to 1.34                  |
| v2.22                 | 1.31 to 1.33                  |

For information on supported versions of managed Kubernetes, it's important to consult the release notes provided by your Kubernetes service provider. There, you can confirm the specific version of the underlying Kubernetes platform supported by the provider, ensuring compatibility with NVIDIA Run:ai. For an up-to-date end-of-life statement see [Kubernetes Release History](https://kubernetes.io/releases/) or [OpenShift Container Platform Life Cycle Policy](https://access.redhat.com/support/policy/updates/openshift).

### NVIDIA Run:ai Namespace

The NVIDIA Run:ai control plane uses a namespace `runai-backend`. Use the following to create the namespace:

```bash
kubectl create namespace runai-backend
```

### Default Storage Class

The NVIDIA Run:ai control plane requires a **default storage class** to create persistent volume claims for NVIDIA Run:ai storage. The storage class, as per Kubernetes standards, controls the reclaim behavior, whether the NVIDIA Run:ai persistent data is saved or deleted when the NVIDIA Run:ai control plane is deleted.

<Info>

**Note**

For a simple (non-production) storage class example see [Kubernetes Local Storage Class](https://kubernetes.io/docs/concepts/storage/storage-classes/#local). The storage class will set the directory `/opt/local-path-provisioner` to be used across all nodes as the path for provisioning persistent volumes. Then set the new storage class as default:

```bash
kubectl patch storageclass local-path -p '{"metadata": {"annotations":{"storageclass.kubernetes.io/is-default-class":"true"}}}'
```

</Info>

### Kubernetes Load Balancer

In Kubernetes, services of type `LoadBalancer` are used to expose applications outside the cluster through a single, stable IP address, providing a consistent entry point for external traffic. In managed cloud environments this capability is built-in, while in self-hosted and on-premise deployments it must be provided explicitly.

MetalLB fulfills this role by allocating external IP addresses from a predefined pool and advertising them on the external network, enabling access to services running inside the cluster.

In NVIDIA Run:ai, this is required to support north-south traffic, including access to the NVIDIA Run:ai control plane, APIs, UI, inference endpoints, and externally exposed development workspaces and training workloads.

1. Reserve a range of IP addresses (recommended a full 32 subnet) for example: `172.20.10.0-172.20.10.255`
2.  Install MetalLB:

    ```bash
    helm repo add metallb https://metallb.github.io/metallb
    helm repo update
    helm install metallb metallb/metallb --version 0.15.3 --namespace metallb-system --create-namespace
    ```
3.  Create a YAML file named `metalLB-config.yaml` and replace `<IPADDRESS-RANGE-START>-<IPADDRESS-RANGE-END>` with the reserved range of IP addresses:

    ```yaml
    apiVersion: metallb.io/v1beta1
    kind: IPAddressPool
    metadata:
      name: runai-ip-pool
      namespace: metallb-system
    spec:
      addresses:
      - <IPADDRESS-RANGE-START>-<IPADDRESS-RANGE-END> # for example 172.20.10.0-172.20.10.255
    ---
    apiVersion: metallb.io/v1beta1
    kind: L2Advertisement
    metadata:
      name: runai-l2-advertisement
      namespace: metallb-system
    spec:
      ipAddressPools:
      - runai-ip-pool
    ```
4.  Apply the YAML:

    ```bash
    kubectl apply -f metalLB-config.yaml
    ```

### Kubernetes Ingress Controller

The NVIDIA Run:ai control plane requires [Kubernetes Ingress Controller](https://kubernetes.io/docs/concepts/services-networking/ingress-controllers/) to be installed on the Kubernetes cluster.

* RKE2 comes with a pre-installed ingress controller.
* Make sure that a default ingress controller, `global.ingress.ingressClass` is set. For more details, see [Advanced control plane configurations](../../infrastructure-setup/advanced-setup/control-plane-config.md).

There are multiple ways to install and configure an ingress controller. The following example demonstrates how to install and configure the [HAProxy Kubernetes Ingress Controller](https://www.haproxy.com/documentation/kubernetes-ingress/) using [helm](https://helm.sh/).

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
--create-namespace \
--namespace haproxy-controller \
--set controller.kind=DaemonSet \
--set controller.service.type=NodePort
```

## Domain and Certificate Requirements

To install the NVIDIA Run:ai control plane in a NVIDIA Run:ai multi-tenant deployment, configure a wildcard DNS and wildcard certificate that allow secure access across tenant environments.

### Wildcard DNS Record

In order to expose the NVIDIA Run:ai platform under a unified domain, configure a wildcard DNS record (e.g., `*.runai.hostorg.com`) that resolves to the cluster's load balancer IP address.

### Wildcard TLS Certificate

You must provide a TLS certificate that matches your wildcard DNS domain (e.g., `*.runai.hostorg.com`). This certificate is used to secure HTTPS access to tenant-facing endpoints, ensuring that each tenant receives a secure URL when accessing NVIDIA Run:ai services.

Create a [Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/) named `runai-backend-tls` in the `runai-backend` namespace and include the path to the TLS `--cert` and its corresponding private `--key` by running the following:

```bash
kubectl create secret tls runai-backend-tls -n runai-backend \
  --cert /path/to/fullchain.pem  \ # Replace /path/to/fullchain.pem with the actual path to your TLS certificate 
  --key /path/to/private.pem # Replace /path/to/private.pem with the actual 
```

## External Postgres Database (Optional)

The NVIDIA Run:ai control plane installation includes a default PostgreSQL database. However, you may opt to use an existing PostgreSQL database if you have specific requirements or preferences as detailed in [External Postgres database configuration.](preparations.md#external-postgres-database-optional) Note that only PostgreSQL version 16 is supported.
