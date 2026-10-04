---
---

# Preparations

This page prepares your environment for installing the NVIDIA Run:ai control plane and a tenant cluster: setting up your installer machine, creating the namespaces and TLS/CA secrets each install needs, and retrieving software artifacts and registry access.

## System Requirements

Complete the [System Requirements](../../system-requirements/README.md) section before starting. It covers:

* [Infrastructure Requirements](../../system-requirements/infrastructure-requirements.md) - Hardware, operating system, Kubernetes distribution, storage, and permissions required for the control plane's and each tenant cluster's Kubernetes environment.
* [Networking Setup](../../system-requirements/networking-setup.md) - Load balancer and ingress requirements, the control plane's wildcard FQDN and TLS certificate, and each tenant cluster's own FQDNs and TLS certificates. The exact requirements depend on your routing configuration and the workload types each tenant plans to support.
* [Third-Party Services](../../system-requirements/third-party-services.md) - The third-party services deployed as part of the installation.
* [Metrics Store Requirements](../../system-requirements/metrics-store-requirements/README.md) - The external metrics store the control plane requires.
* [External Access to Critical Services](../../system-requirements/external-access-to-critical-services.md) - The inbound and outbound connectivity required to reach those services and NVIDIA registries.

Your [deployment model](../../system-requirements/deployment-model.md) is always multi-cluster: control plane setup is completed once, and cluster setup is repeated for each tenant cluster.

## General Setup

This section covers what you set up once, before installing anything: your installer machine and the registry credentials needed to pull NVIDIA Run:ai images. None of it depends on whether you are installing the control plane or a tenant cluster.

### Installer Machine

The machine you run Helm from needs:

* At least 50GB of free space
* Docker installed
* [Helm](https://helm.sh/) 3.14 or later

<Note>

**Note**

Helm 4 defaults to [server-side apply](https://helm.sh/docs/overview/#server-side-apply) when installing a new chart release, which can conflict with resources managed by the NVIDIA Run:ai operator. Append `--server-side=false` to your `helm upgrade` command. NVIDIA Run:ai clusters originally installed with Helm 3.x are unaffected.

</Note>

### Registry Credentials

NVIDIA Run:ai Support-Entitled Release installations require an NGC Personal Key to authenticate to the NVIDIA Run:ai container registry (`nvcr.io`). Generate a Personal Key from the [NGC Personal Keys](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#ngc-api-keys) page.

Personal Keys begin with `nvapi-`. NGC authenticates using the literal username `$oauthtoken` with the key as the password. The Personal Key is used during installation, to pull the NVIDIA Run:ai container images. Use this key to create the `runai-reg-creds` secret described in [Software Artifacts](#software-artifacts) below.

For connected installs, NVIDIA Run:ai's telemetry for Support-Entitled Release installations is not associated with your organization by default. To associate it with your NCA ID, generate an [NGC Service API Key](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#generating-a-service-api-key). In the **Service** field, select both **NVIDIA Run:ai Self-Hosted** and **NGC Catalog**. The Personal Key is used during installation. The Service API Key is used at runtime, after installation, by the NVIDIA Run:ai control plane to report telemetry. Pass the resulting key as `telemetryService.telemetryAuth.token` during installation, or reference a secret you create yourself beforehand via `telemetryService.telemetryAuth.existingSecret`/`secretKey`. Create the `runai-backend` namespace first, as described in [Control Plane Namespace](#control-plane-namespace), then create the secret:

```bash
kubectl create secret generic <SECRET_NAME> \
  --namespace runai-backend \
  --from-literal=<KEY_NAME>=<SERVICE_API_KEY>
```

## Control Plane Setup

This section covers what you set up for the NVIDIA Run:ai control plane: its namespace and TLS certificate. A multi-tenant deployment has only one control plane, always on its own dedicated Kubernetes cluster. See [Deployment Model](../../system-requirements/deployment-model.md).

### Control Plane Namespace

The NVIDIA Run:ai control plane uses a namespace called `runai-backend`. Use the following to create the namespace:

```bash
kubectl create namespace runai-backend
```

### Control Plane Certificate

This certificate secures the NVIDIA Run:ai control plane's wildcard FQDN you prepared in [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns). Before creating the secret below, see [TLS Certificates](../../system-requirements/networking-setup.md#tls-certificates) to confirm which certificate you need and why.

You must have a **wildcard** TLS certificate that is associated with the control plane's FQDN for HTTPS access, covering every tenant subdomain. Create a [Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/) named `runai-backend-tls` in the `runai-backend` namespace and include the path to the TLS `--cert` and its corresponding private `--key` by running the following:

* Replace `/path/to/fullchain.pem` with the actual path to your wildcard TLS certificate.
* Replace `/path/to/private.pem` with the actual path to your private key.

```bash
kubectl create secret tls runai-backend-tls -n runai-backend \
  --cert /path/to/fullchain.pem \
  --key /path/to/private.pem
```

## Cluster Setup

This section covers what you set up for a tenant's NVIDIA Run:ai cluster: its namespace, TLS certificates, and host-based routing. Set it up on that tenant's own Kubernetes cluster. Repeat this section once per tenant cluster.

### Cluster Namespace

The NVIDIA Run:ai cluster must be installed in a namespace or project (OpenShift) called `runai`. Use the following to create the namespace/project:

<Tabs>

<Tab title="Kubernetes">

```bash
kubectl create ns runai
```

</Tab>

<Tab title="OpenShift">

```bash
oc new-project runai
```

</Tab>

</Tabs>

### Pod Security Admission

NVIDIA Run:ai supports `restricted` policy for [Pod Security Admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/) (PSA) on OpenShift only. Other Kubernetes distributions are only supported with `privileged` policy.

For NVIDIA Run:ai on OpenShift to run with PSA `restricted` policy:

*   Label the `runai` namespace as described in [Pod Security Admission](https://kubernetes.io/docs/concepts/security/pod-security-admission/) with the following labels:

    ```bash
    pod-security.kubernetes.io/audit=privileged
    pod-security.kubernetes.io/enforce=privileged
    pod-security.kubernetes.io/warn=privileged
    ```
* The workloads submitted through NVIDIA Run:ai should comply with the restrictions of PSA restricted policy. This can be enforced using Policies.

### Cluster TLS Certificates

These certificates secure the tenant cluster's domain, workspaces, training, distributed inference, and standard inference FQDNs you prepared in [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns). Before creating the secrets below, see [TLS Certificates](../../system-requirements/networking-setup.md#tls-certificates) to confirm which certificates you need and why. The workspaces, training, and distributed inference wildcard certificate is only needed for the default host-based routing - see [Host-Based Routing (Ingress Resource)](#host-based-routing-ingress-resource) below for the matching Ingress setup. If you are using path-based routing instead, skip that certificate.

#### Cluster Domain Certificates (Single-Domain)

*   **Kubernetes** - To enable secure communication between the NVIDIA Run:ai control plane and the cluster, configure a TLS certificate associated with the cluster's main domain (e.g. `runai.tenant-a-corp.com`). This certificate should be stored as a secret named `runai-cluster-domain-tls-secret` in the `runai` namespace.

    * Replace `/path/to/fullchain.pem` with the actual path to your TLS certificate.
    * Replace `/path/to/private.pem` with the actual path to your private key.

    ```bash
    kubectl create secret tls runai-cluster-domain-tls-secret -n runai \
      --cert /path/to/fullchain.pem \
      --key /path/to/private.pem
    ```
* **OpenShift** - NVIDIA Run:ai uses the OpenShift default Ingress router for serving. The TLS certificate configured for this router must be issued by a trusted CA. For more details, see the OpenShift documentation on [configuring certificates](https://docs.redhat.com/en/documentation/openshift_container_platform/4.18/html/security_and_compliance/configuring-certificates#replacing-default-ingress).

#### Workspaces, Training, and Distributed Inference Wildcard Certificate

<Note>

**Note**

For path-based routing, this certificate is not required.

</Note>

* **Kubernetes** - To allow secure access to workspace, training, and distributed inference workloads via subdomains, configure a wildcard TLS certificate that matches the cluster domain (e.g. `*.runai.tenant-a-corp.com`). This certificate should be stored as a secret named `runai-cluster-domain-star-tls-secret` in the `runai` namespace.
  * Replace `/path/to/fullchain.pem` with the actual path to your TLS certificate.
  *   Replace `/path/to/private.pem` with the actual path to your private key.

      ```bash
      kubectl create secret tls runai-cluster-domain-star-tls-secret -n runai \
        --cert /path/to/fullchain.pem \
        --key /path/to/private.pem
      ```
* **OpenShift** - A wildcard TLS certificate for Workspace and Training workloads is not required. OpenShift Routes handle TLS termination for workspace and training endpoints using the platform's built-in routing and certificate management.

#### Standard Inference Wildcard Certificate

<Note>

**Note**

The inference wildcard certificate is only required for standard inference (Knative, single-node) workloads exposed externally via wildcard FQDN. Other inference workloads (for example, distributed inference using serving frameworks such as LWS, Dynamo, KServe, or others) do not require this certificate. For cluster-local standard inference, skip this section.

</Note>

*   **Kubernetes** - To securely expose inference services over HTTPS, configure a wildcard TLS certificate for the inference domain (e.g. `*.runai-inference.tenant-a-corp.com`). This certificate should be stored as a secret named `runai-cluster-inference-tls-secret` in the `knative-serving` namespace.

    * Replace `/path/to/fullchain.pem` with the actual path to your TLS certificate.
    * Replace `/path/to/private.pem` with the actual path to your private key.

    ```bash
    kubectl create secret tls runai-cluster-inference-tls-secret -n knative-serving \
        --cert /path/to/fullchain.pem \
        --key /path/to/private.pem
    ```
* **OpenShift** - A wildcard TLS certificate for Inference workloads is not required. OpenShift Routes handle TLS termination for inference endpoints using the platform's built-in routing and certificate management.

### Host-Based Routing (Ingress Resource)

<Note>

**Note**

* The following steps are required for Kubernetes only. For OpenShift, no additional configuration is required.
* NVIDIA Run:ai also supports path-based routing. If you prefer to use it instead of the default host-based routing, disable host-based routing by setting `clusterConfig.global.subdomainSupport: false` during the Helm installation. See [Advanced Cluster Configurations](../../../../installation-and-configuration/configuration/cluster-config.md).
* If you choose path-based routing, skip the steps below.

</Note>

Host-based routing binds together the configured domains (FQDN), TLS certificates, and ingress rules to expose workloads externally. With NVIDIA Run:ai host-based routing, workloads are exposed using subdomains, so each workload is assigned its own URL. For example:

```bash
https://<project>-<workload>.<CLUSTER_URL>
```

Host-based routing relies on:

* The FQDN structure defined earlier
* The TLS certificates configured for those domains

1. Ensure that:
   * Wildcard DNS records are configured (see [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns))
   * TLS certificates are created and stored as Kubernetes secrets (see [Cluster TLS Certificates](#cluster-tls-certificates) above)
2.  Create the ingress resource

    ```yaml
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: runai-cluster-domain-star-ingress
      namespace: runai
    spec:
      rules:
      - host: '*.<CLUSTER_URL>'
      tls:
      - hosts:
        - '*.<CLUSTER_URL>'
        secretName: runai-cluster-domain-star-tls-secret
    ```
3.  Run the following:

    ```bash
    kubectl apply -f <filename>
    ```

## Software Artifacts

This section describes how to prepare the software artifacts required for installing the NVIDIA Run:ai [control plane](install-control-plane.md) and [onboarding a tenant](create-a-tenant-and-cluster.md).

### Kubernetes

In connected environments, Kubernetes pulls NVIDIA Run:ai container images directly from a remote registry at runtime.

<Note>

**Note**

This section applies only to NVIDIA Run:ai Support-Entitled Release installations. NVIDIA Run:ai Open Source images are served publicly from `nvcr.io` with no authentication, so no registry secret is needed.

</Note>

For Support-Entitled Release, you must create a Kubernetes secret for registry access. NVIDIA Run:ai container images are hosted in the NVIDIA NGC container registry under `nvidia/runai-oss`. Using the NGC Personal Key from [Registry Credentials](#registry-credentials) above, create the required secret:

```bash
kubectl create secret docker-registry runai-reg-creds \
--docker-server=https://nvcr.io \
--docker-username='$oauthtoken' \
--docker-password=<NGC_API_KEY> \
--docker-email=<EMAIL> \
--namespace=runai-backend
```

<Note>

**Note**

Creating this secret alone does not configure the control plane to use it - you must also add `--set global.imagePullSecrets[0].name=runai-reg-creds` to the `helm upgrade` command. See [Control Plane Installation](install-control-plane.md#installation).

</Note>

### OpenShift

In connected environments, OpenShift pulls NVIDIA Run:ai container images directly from a remote registry at runtime.

<Note>

**Note**

This section applies only to NVIDIA Run:ai Support-Entitled Release installations. NVIDIA Run:ai Open Source images are served publicly from `nvcr.io` with no authentication, so no registry secret is needed.

</Note>

For Support-Entitled Release, you must create a Kubernetes secret for registry access. NVIDIA Run:ai container images are hosted in the NVIDIA NGC container registry under `nvidia/runai-oss`. Using the NGC Personal Key from [Registry Credentials](#registry-credentials) above, create the required secret:

```bash
oc create secret docker-registry runai-reg-creds \
--docker-server=https://nvcr.io \
--docker-username='$oauthtoken' \
--docker-password=<NGC_API_KEY> \
--docker-email=<EMAIL> \
--namespace=runai-backend
```

<Note>

**Note**

Creating this secret alone does not configure the control plane to use it - you must also add `--set global.imagePullSecrets[0].name=runai-reg-creds` to the `helm upgrade` command. See [Control Plane Installation](install-control-plane.md#installation).

</Note>

## Next Steps

* [Control Plane Installation](install-control-plane.md) - Install the control plane.
* [API Access Setup](../guided-installation/api-access-setup.md) - Configure programmatic API access using service account credentials.
* [Prerequisites and Features](prerequisites-and-features.md) - Install the GPU Operator, Prometheus, and any workload-specific features (such as distributed training or inference) that tenant's cluster needs.
* [Create a Tenant and Cluster](create-a-tenant-and-cluster.md) - Create a tenant, provision and register its cluster, and install that cluster.
