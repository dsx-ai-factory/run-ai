# Control Plane System Requirements

The NVIDIA Run:ai control plane is a Kubernetes application. It has specific system and Kubernetes environment requirements that must be met before installation. These requirements ensure that the control plane services can be deployed successfully.

This section describes the minimum hardware, supported Kubernetes/OpenShift versions, and environment prerequisites required for the NVIDIA Run:ai control plane installation. Environment prerequisites include critical infrastructure components such as a properly configured Fully Qualified Domain Name (FQDN) with DNS resolution, TLS certificate configuration and ingress readiness.

## Installer Machine

The machine running the installation script (typically the Kubernetes master) must have:

* At least 50GB of free space
* Docker installed
* [Helm](https://helm.sh/) 3.14 or later

<Info>

**Note**

If you are installing an air-gapped version of NVIDIA Run:ai, the NVIDIA Run:ai [software artifacts](preparations.md#software-artifacts) include the Helm binary.

</Info>

## Hardware Requirements

The following hardware requirements are for the NVIDIA Run:ai control plane system nodes. By default, all control plane services run on all available nodes in the control plane app cluster.

### Architecture

* **x86** - Supported for Kubernetes and OpenShift
* **ARM** - Supported for Kubernetes and OpenShift

### NVIDIA Run:ai Control Plane - System Nodes

This configuration is the minimum requirement you need to install and use NVIDIA Run:ai control plane:

| Component  | Required Capacity |
| ---------- | ----------------- |
| CPU        | 10 cores          |
| Memory     | 12GB              |
| Disk space | 110GB             |

<Info>

**Note**

To designate nodes to NVIDIA Run:ai system services, follow the instructions as described in [System nodes](../../../installation-and-configuration/configuration/node-roles.md#system-nodes).

</Info>

If NVIDIA Run:ai control plane is planned to be installed on the same Kubernetes cluster as the NVIDIA Run:ai cluster, make sure the cluster [Hardware requirements](system-requirements.md#hardware-requirements) are considered in addition to the NVIDIA Run:ai control plane hardware requirements.

## Software Requirements

The following software requirements must be fulfilled.

### Operating System

* Any **Linux** operating system supported by both Kubernetes and NVIDIA GPU Operator
* Internal tests are being performed on **Ubuntu 22.04** and **CoreOS** for OpenShift.

### Network Time Protocol

Nodes are required to be synchronized by time using NTP (Network Time Protocol) for proper system functionality.

### Kubernetes Distribution

NVIDIA Run:ai control plane requires Kubernetes. The following Kubernetes distributions are supported:

* Vanilla Kubernetes
* OpenShift Container Platform (OCP)
* Elastic Kubernetes Engine (EKS)
* Google Kubernetes Engine (GKE)
* Azure Kubernetes Service (AKS)
* Oracle Kubernetes Engine (OKE)
* Rancher Kubernetes Engine 2 (RKE2)

<Info>

**Note**

The latest release of the NVIDIA Run:ai control plane supports **Kubernetes 1.34 to 1.36** and **OpenShift 4.18 to 4.21**.

</Info>

For the full list of NVIDIA Run:ai compatible distributions and per-version Kubernetes and OpenShift compatibility, see [Support matrix](../support-matrix.md#nvidia-run-ai-compatible-distributions).

### NVIDIA Run:ai Namespace

The NVIDIA Run:ai control plane uses a namespace or project (OpenShift) called `runai-backend`. Use the following to create the namespace/project:

<Tabs>

<Tab title="Kubernetes">

```bash
kubectl create namespace runai-backend
```

</Tab>

<Tab title="OpenShift">

```bash
oc new-project runai-backend
```

</Tab>

</Tabs>

### Default Storage Class

<Info>

**Note**

Default storage class applies to Kubernetes only.

</Info>

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

* OpenShift and RKE2 come with a pre-installed ingress controller.
* Make sure that a default ingress controller, `global.ingress.ingressClass` is set. For more details, see [Advanced control plane configurations](../../../installation-and-configuration/configuration/control-plane-config.md).

There are multiple ways to install and configure an ingress controller. The following example demonstrates how to install and configure the [HAProxy Kubernetes Ingress Controller](https://www.haproxy.com/documentation/kubernetes-ingress/) using [helm](https://helm.sh/).

<Accordion title="Vanilla Kubernetes">

**Vanilla Kubernetes**

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
 --create-namespace \
 --namespace haproxy-controller \
 --set controller.kind=DaemonSet \
 --set controller.service.type=NodePort
```

</Accordion>

<Accordion title="Managed Kubernetes (EKS, GKE, AKS)">

**Managed Kubernetes (EKS, GKE, AKS)**

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
 --create-namespace \
 --namespace haproxy-controller \
 --set controller.service.type=LoadBalancer
```

</Accordion>

<Accordion title="Oracle Kubernetes Engine (OKE)">

**Oracle Kubernetes Engine (OKE)**

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
 --create-namespace \
 --namespace haproxy-controller \
 --set controller.kind=DaemonSet \
 --set controller.service.type=LoadBalancer \
 --set controller.service.externalTrafficPolicy=Local \
 --set controller.service.annotations."oci-network-load-balancer\.oraclecloud\.com/is-preserve-source"="True" \
 --set controller.service.annotations."oci-network-load-balancer\.oraclecloud\.com/security-list-management-mode"=All \
 --set controller.service.annotations."oci\.oraclecloud\.com/load-balancer-type"=nlb
```

</Accordion>

## Fully Qualified Domain Name (FQDN)

<Info>

**Note**

Fully Qualified Domain Name applies to Kubernetes only.

</Info>

NVIDIA Run:ai services rely on a Fully Qualified Domain Name (FQDN) to route traffic to the control plane and enable communication between the control plane and connected clusters. You must configure a domain name for control plane access (for example, `runai.mycorp.local`). This cannot be an IP address. The FQDN must be resolvable within the organization's private network.

## TLS Certificate

### Kubernetes

You must have a TLS certificate that is associated with the FQDN for HTTPS access. Create a [Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/) named `runai-backend-tls` in the `runai-backend` namespace and include the path to the TLS `--cert` and its corresponding private `--key` by running the following:

* Replace `/path/to/fullchain.pem` with the actual path to your TLS certificate.
* Replace `/path/to/private.pem` with the actual path to your private key.

```bash
kubectl create secret tls runai-backend-tls -n runai-backend \
  --cert /path/to/fullchain.pem \
  --key /path/to/private.pem
```

### OpenShift

NVIDIA Run:ai uses the OpenShift default Ingress router for serving. The TLS certificate configured for this router must be issued by a trusted CA. For more details, see the OpenShift documentation on [configuring certificates](https://docs.redhat.com/en/documentation/openshift_container_platform/4.18/html/security_and_compliance/configuring-certificates#replacing-default-ingress).

## Local Certificate Authority (CA)

A local certificate authority serves as the root certificate for organizations that cannot use **publicly trusted certificate authority**. Follow the steps below to configure the local certificate authority.

In air-gapped environments, you **must** configure and install the local CA's public key in the Kubernetes cluster. This is required for the installation to succeed:

1. Add the public key to the `runai-backend` namespace:

<Tabs>

<Tab title="Kubernetes">

```bash
kubectl -n runai-backend create secret generic runai-ca-cert \
    --from-file=runai-ca.pem=<ca_bundle_path>
```

</Tab>

<Tab title="OpenShift">

```bash
oc -n runai-backend create secret generic runai-ca-cert \
    --from-file=runai-ca.pem=<ca_bundle_path>
```

</Tab>

</Tabs>

2. When installing the control plane, make sure the following flag is added to the helm command `--set global.customCA.enabled=true`. See [Install control plane](install-control-plane.md).

## External Postgres Database (Optional)

The NVIDIA Run:ai control plane installation includes a default PostgreSQL database. However, you may opt to use an existing PostgreSQL database if you have specific requirements or preferences as detailed in [External Postgres database configuration](preparations.md#external-postgres-database-optional). Note that only PostgreSQL version 16 is supported.
