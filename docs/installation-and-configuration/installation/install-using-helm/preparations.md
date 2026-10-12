# Preparations

This page prepares your environment for installing the NVIDIA Run:ai control plane and cluster: setting up your installer machine, creating the namespaces and TLS/CA secrets each install needs, and retrieving software artifacts and registry access.

## System Requirements

Complete the [System Requirements](../../system-requirements/README.md) section before starting. It covers:

* [Infrastructure Requirements](../../system-requirements/infrastructure-requirements.md) - Hardware, operating system, Kubernetes distribution, storage, and permissions required in your Kubernetes environment.
* [Networking Setup](../../system-requirements/networking-setup.md) - Load balancer and ingress requirements, FQDNs, TLS certificates and private keys, and a local CA chain if your certificates are privately signed. The exact requirements depend on your deployment model, routing configuration, and the workload types you plan to support.
* [Third-Party Services](../../system-requirements/third-party-services.md) - The third-party services deployed as part of the installation.
* [External Access to Critical Services](../../system-requirements/external-access-to-critical-services.md) - The inbound and outbound connectivity required to reach those services and NVIDIA registries.

<Note>

**Gov-ready**

For a Gov-ready installation, also review [Gov-Ready Third-Party Services](../../system-requirements/gov-ready-third-party-services.md) and the [Gov-ready outbound access](../../system-requirements/external-access-to-critical-services.md#gov-ready) to Docker Hub, which replace the standard third-party services images.

</Note>

Your [deployment model](../../system-requirements/deployment-model.md) determines how the installation steps are applied. In a same-cluster deployment, the control plane and NVIDIA Run:ai cluster are installed on the same Kubernetes cluster. In a multi-cluster deployment, control plane setup is completed once, and cluster setup is repeated for each NVIDIA Run:ai cluster.

## General Setup

This section covers what you set up once, before installing anything: your installer machine and the registry credentials needed to pull NVIDIA Run:ai images. None of it depends on your deployment model or on whether you are installing the control plane, the cluster, or both.

### Installer Machine

The machine you run Helm from needs:

* At least 50GB of free space
* Docker installed
* [Helm](https://helm.sh/) 3.14 or later

<Note>

**Note**

If you are installing air-gapped, the machine running `airgap pack` and the machine running `airgap load` each need at least 20GB of free space.

</Note>

<Note>

**Note**

Helm 4 defaults to [server-side apply](https://helm.sh/docs/overview/#server-side-apply) when installing a new chart release, which can conflict with resources managed by the NVIDIA Run:ai operator. Append `--server-side=false` to your `helm upgrade` command. NVIDIA Run:ai clusters originally installed with Helm 3.x are unaffected.

</Note>

### Registry Credentials

NVIDIA Run:ai Support-Entitled Release installations require an NGC Personal Key to authenticate to the NVIDIA Run:ai container registry (`nvcr.io`), for connected installs. Generate a Personal Key from the [NGC Personal Keys](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#ngc-api-keys) page.

Personal Keys begin with `nvapi-`. NGC authenticates using the literal username `$oauthtoken` with the key as the password. The Personal Key is used during installation, to pull the NVIDIA Run:ai container images. Use this key to create the `runai-reg-creds` secret described in [Software Artifacts](#software-artifacts) below.

For connected installs, NVIDIA Run:ai's telemetry for Support-Entitled Release installations is not associated with your organization by default. To associate it with your NCA ID, generate an [NGC Service API Key](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#generating-a-service-api-key). In the **Service** field, select both **NVIDIA Run:ai Self-Hosted** and **NGC Catalog**. The Personal Key is used during installation. The Service API Key is used at runtime, after installation, by the NVIDIA Run:ai control plane to report telemetry. Pass the resulting key as `telemetryService.telemetryAuth.token` during installation, or reference a secret you create yourself beforehand via `telemetryService.telemetryAuth.existingSecret`/`secretKey`. Create the `runai-backend` namespace first, as described in [Control Plane Namespace](#control-plane-namespace), then create the secret:

```bash
kubectl create secret generic <SECRET_NAME> \
  --namespace runai-backend \
  --from-literal=<KEY_NAME>=<SERVICE_API_KEY>
```

## Control Plane Setup

This section covers what you set up for the NVIDIA Run:ai control plane: its namespace and TLS certificate and local CA secrets. A deployment has only one control plane, whether same-cluster or multi-cluster. Set it up on the control plane's own Kubernetes cluster - the same Kubernetes cluster as the first NVIDIA Run:ai cluster in a same-cluster deployment, or a dedicated Kubernetes cluster in a multi-cluster deployment. See [Deployment Model](../../system-requirements/deployment-model.md).

### Control Plane Namespace

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

### Control Plane Certificate

This certificate secures the NVIDIA Run:ai control plane FQDN you prepared in [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns). Before creating the secret below, see [TLS Certificates](../../system-requirements/networking-setup.md#tls-certificates) to confirm which certificate you need and why.

*   **Kubernetes** - You must have a TLS certificate that is associated with the FQDN for HTTPS access. Create a [Kubernetes Secret](https://kubernetes.io/docs/concepts/configuration/secret/) named `runai-backend-tls` in the `runai-backend` namespace and include the path to the TLS `--cert` and its corresponding private `--key` by running the following:

    * Replace `/path/to/fullchain.pem` with the actual path to your TLS certificate.
    * Replace `/path/to/private.pem` with the actual path to your private key.

    ```bash
    kubectl create secret tls runai-backend-tls -n runai-backend \
      --cert /path/to/fullchain.pem \
      --key /path/to/private.pem
    ```
* **OpenShift** - NVIDIA Run:ai uses the OpenShift default Ingress router for serving. The TLS certificate configured for this router must be issued by a trusted CA. For more details, see the OpenShift documentation on [configuring certificates](https://docs.redhat.com/en/documentation/openshift_container_platform/4.18/html/security_and_compliance/configuring-certificates#replacing-default-ingress).

### Control Plane Local CA

If your TLS certificates are signed by a private or local CA rather than a publicly trusted one, you need to provide that CA's certificate chain as part of installation. See [Local Certificate Authority (CA)](../../system-requirements/networking-setup.md#local-certificate-authority-ca) for details.

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

## Cluster Setup

This section covers what you set up for each NVIDIA Run:ai cluster: its namespace, TLS certificates, host-based routing, and local CA secrets. Set it up on the Kubernetes cluster where the NVIDIA Run:ai cluster will run - the same Kubernetes cluster as the control plane in a same-cluster deployment, or a separate Kubernetes cluster per NVIDIA Run:ai cluster in a multi-cluster deployment. In a multi-cluster deployment, repeat this section once per cluster. See [Deployment Model](../../system-requirements/deployment-model.md).

<Note>

**Note**

If you are deploying the NVIDIA Run:ai cluster on the same Kubernetes cluster as the NVIDIA Run:ai control plane, the control plane's FQDN, TLS certificate, and CA secrets already satisfy the cluster's requirements - you do not need to repeat them.

</Note>

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
* The workloads submitted through NVIDIA Run:ai should comply with the restrictions of PSA restricted policy. This can be enforced using [Policies](../../../platform-management/policies/native-workload-policies.md).

### Cluster TLS Certificates

These certificates secure the NVIDIA Run:ai cluster, workspaces, training, distributed inference, and standard inference FQDNs you prepared in [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns). Before creating the secrets below, see [TLS Certificates](../../system-requirements/networking-setup.md#tls-certificates) to confirm which certificates you need and why. The workspaces, training, and distributed inference wildcard certificate is only needed for the default host-based routing - see [Host-Based Routing (Ingress Resource)](#host-based-routing-ingress-resource) below for the matching Ingress setup. If you are using path-based routing instead, skip that certificate.

#### Cluster Domain Certificates (Single-Domain)

*   **Kubernetes** - To enable secure communication between the NVIDIA Run:ai control plane and the cluster, configure a TLS certificate associated with the cluster's main domain (e.g. `runai.mycorp.local`). This certificate should be stored as a secret named `runai-cluster-domain-tls-secret` in the `runai` namespace.

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

* **Kubernetes** - To allow secure access to workspace, training, and distributed inference workloads via subdomains, configure a wildcard TLS certificate that matches the cluster domain (e.g. `*.runai.mycorp.local`). This certificate should be stored as a secret named `runai-cluster-domain-star-tls-secret` in the `runai` namespace.
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

*   **Kubernetes** - To securely expose inference services over HTTPS, configure a wildcard TLS certificate for the inference domain (e.g. `*.runai-inference.mycorp.local`). This certificate should be stored as a secret named `runai-cluster-inference-tls-secret` in the `knative-serving` namespace.

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
* NVIDIA Run:ai also supports path-based routing. If you prefer to use it instead of the default host-based routing, disable host-based routing by setting `clusterConfig.global.subdomainSupport: false` during the Helm installation. See [Advanced cluster configurations](../../configuration/cluster-config.md).
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
   * TLS certificates are created and stored as Kubernetes secrets (see [Cluster TLS Certificates](preparations.md#cluster-tls-certificates) above)
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

### Cluster Local CA

If your TLS certificates are signed by a private or local CA rather than a publicly trusted one, you need to provide that CA's certificate chain as part of installation. See [Local Certificate Authority (CA)](../../system-requirements/networking-setup.md#local-certificate-authority-ca) for details.

1\. Add the public key to the required namespace:

<Tabs>

<Tab title="Kubernetes">

```bash
kubectl -n runai create secret generic runai-ca-cert \
    --from-file=runai-ca.pem=<ca_bundle_path>
kubectl label secret runai-ca-cert -n runai run.ai/cluster-wide=true run.ai/name=runai-ca-cert --overwrite
```

</Tab>

<Tab title="OpenShift">

```bash
oc -n runai create secret generic runai-ca-cert \
    --from-file=runai-ca.pem=<ca_bundle_path>
oc -n openshift-monitoring create secret generic runai-ca-cert \
    --from-file=runai-ca.pem=<ca_bundle_path>
oc label secret runai-ca-cert -n runai run.ai/cluster-wide=true run.ai/name=runai-ca-cert --overwrite
```

</Tab>

</Tabs>

2. When installing the cluster, make sure the following flag is added to the helm command `--set global.customCA.enabled=true`. See [Install the cluster](helm-install.md#installation).

<Note>

**Note**

For Git and S3 data source integrations, NVIDIA Run:ai supports the following options:

* Use the same custom CA defined during cluster installation by setting:\
  `--set global.customCAGit.enabled=true` or\
  `--set global.customCAS3.enabled=true`.
* Use a different CA certificate specifically for Git or S3 by enabling the setting and providing a custom secret name:\
  `--set global.customCAGit.enabled=true`\
  `--set global.customCAGit.secret.name=<git-ca-cert>` or\
  `--set global.customCAS3.enabled=true`\
  `--set global.customCAS3.secret.name=<s3-ca-cert>`

For more details, see [Advanced cluster configurations](../../configuration/cluster-config.md).

</Note>

## Software Artifacts

This section describes how to prepare the software artifacts required for installing the NVIDIA Run:ai [control plane](install-control-plane.md) and [cluster](helm-install.md). The steps depend on:

* Platform distribution - Kubernetes or OpenShift
* On-premise environment type - connected or air-gapped (i.e. not connected)

### Kubernetes

<Accordion title="Connected">

**Connected**

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

Creating this secret alone does not configure the control plane to use it - you must also add `--set global.imagePullSecrets[0].name=runai-reg-creds` to the `helm upgrade` command. See [Install the Control Plane](install-control-plane.md#installation).

</Note>

<Note>

**Gov-ready**

A Gov-ready installation also pulls the hardened third-party images from Docker Hub, which requires a separate Docker Hub access token and image pull secret in the `runai-backend` and `runai` namespaces. Refer to [Docker Hub Credentials](../../system-requirements/gov-ready-third-party-services.md#docker-hub-credentials) for how to create them and reference them in the Helm values.

</Note>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

In air-gapped environments, the cluster cannot pull container images from external registries during runtime. Instead, use the `runai-installer` binary's `airgap pack` command to build a bundle containing every image your release needs on a connected machine, transfer it into your air-gapped network, then use its `airgap load` command to push those images to your internal registry and generate the Helm values that point to it.

**Pack the Bundle (Connected Machine)**

Packing happens on a separate connected machine - any machine with outbound internet access.

On that connected machine, download the `runai-installer` binary and use its `airgap pack` command to build a single tarball containing every image your release needs - NVIDIA Run:ai platform images and third-party service images (PostgreSQL, Keycloak, and others) alike. See [Third-Party Services](../../system-requirements/third-party-services.md#image-registry).

1. Download the `runai-installer` binary for your operating system and CPU architecture, on each machine you run the installer from. Replace `<VERSION>` with the NVIDIA Run:ai version you are installing, for example `3.0.17`:

   ```bash
   # Linux, AMD64 (x86_64)
   curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-linux-amd64
   chmod +x runai-installer-linux-amd64

   # Linux, ARM64
   curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-linux-arm64
   chmod +x runai-installer-linux-arm64

   # macOS, AMD64 (Intel)
   curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-darwin-amd64
   chmod +x runai-installer-darwin-amd64

   # macOS, ARM64 (Apple Silicon)
   curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-darwin-arm64
   chmod +x runai-installer-darwin-arm64
   ```

2. Create a pack configuration file (`airgap-pack.yaml`) and run `airgap pack` against it:

   - `settings` - Required. Your target NVIDIA Run:ai version, architecture, and output path.
   - `imageOverrides` - Points specific third-party service images (for example, Keycloak) at your own source instead of the default one. Each entry also accepts `group`, required only when `name` exists in more than one image group - see [`imageOverrides`](../guided-installation/configuration-reference.md#pack-an-air-gapped-bundle) in the Configuration Reference.
   - `registryAuth` - Authenticates to your own registry hosting an `imageOverrides` source.
   - Third-party services - Add an entry to `imageOverrides` for each hardened third-party image, and an entry to `registryAuth` for `docker.io` with your Docker Hub credentials. See [Gov-Ready Third-Party Services](../../system-requirements/gov-ready-third-party-services.md) for the full example. <Badge intent="info" minimal outlined>Gov-ready</Badge>

   ```bash
   cat > airgap-pack.yaml <<'EOF'
   settings:
     runaiVersion: <RUNAI_VERSION>
     architecture: amd64
     output: runai-airgap-<RUNAI_VERSION>.tar.gz

   imageOverrides:
     - name: keycloak
       source: <your-registry>/keycloak:<tag>
     # Gov-ready: add an entry for each hardened image (keycloak, postgres, traefik, nats, thanos, valkey, git-sync)

   registryAuth:
     - registry: <your-registry-host>
       username: <username>
       passwordPath: <path-to-registry-password>
     # Gov-ready: add an entry for docker.io with your Docker Hub credentials
   EOF

   ./runai-installer airgap pack \
     --config airgap-pack.yaml
   ```

3. Transfer the resulting tarball into the air-gapped network.

**Load the Bundle (Air-Gapped Network)**

Loading happens once the tarball is in the air-gapped network, on any machine there with access to both the tarball and your internal mirror registry.

1. Extract the `runai-installer` binary from the transferred `runai-airgap-<RUNAI_VERSION>.tar.gz` bundle.
2. Create a load configuration file (`airgap-load.yaml`) and run `airgap load` against it:

```bash
cat > airgap-load.yaml <<'EOF'
settings:
  bundlePath: runai-airgap-<RUNAI_VERSION>.tar.gz
  workingDir: ./airgap-work
  valuesDir: ./airgap-values

registry:
  url: <REGISTRY_URL>
  username: <REGISTRY_USERNAME>
  passwordPath: <REGISTRY_PASSWORD_PATH>
EOF

./runai-installer airgap load \
  --config airgap-load.yaml
```

This extracts the bundle, pushes every image to `<REGISTRY_URL>`, and writes `control-plane-airgap-values.yaml` and `cluster-airgap-values.yaml` into `valuesDir`.

Pass these as `-f <VALUES_DIR>/control-plane-airgap-values.yaml` and `-f <VALUES_DIR>/cluster-airgap-values.yaml` to the `helm upgrade` commands in [Install the Control Plane](install-control-plane.md#installation) and [Install the Cluster](helm-install.md#installation).

`<VALUES_DIR>` is the `valuesDir` path you set in `airgap-load.yaml` above.

**Private Docker Registry**

To access the organization's docker registry, set the registry's credentials (imagePullSecret).

Create the secret named `runai-reg-creds` in the `runai-backend` namespace based on your existing credentials. The configuration will be copied over to the `runai` namespace at cluster install. For more information, see [Pull an Image from a Private Registry](https://kubernetes.io/docs/tasks/configure-pod-container/pull-image-private-registry/).

</Accordion>

### OpenShift

<Accordion title="Connected">

**Connected**

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

Creating this secret alone does not configure the control plane to use it - you must also add `--set global.imagePullSecrets[0].name=runai-reg-creds` to the `helm upgrade` command. See [Install the Control Plane](install-control-plane.md#installation).

</Note>

<Note>

**Gov-ready**

A Gov-ready installation also pulls the hardened third-party images from Docker Hub, which requires a separate Docker Hub access token and image pull secret in the `runai-backend` and `runai` namespaces. Refer to [Docker Hub Credentials](../../system-requirements/gov-ready-third-party-services.md#docker-hub-credentials) for how to create them and reference them in the Helm values.

</Note>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

In air-gapped environments, the cluster cannot pull container images from external registries during runtime. Instead, use the `runai-installer` binary's `airgap pack` command to build a bundle containing every image your release needs on a connected machine, transfer it into your air-gapped network, then use its `airgap load` command to push those images to your internal registry and generate the Helm values that point to it.

**Pack the Bundle (Connected Machine)**

Packing happens on a separate connected machine - any machine with outbound internet access.

On that connected machine, download the `runai-installer` binary and use its `airgap pack` command to build a single tarball containing every image your release needs - NVIDIA Run:ai platform images and third-party service images (PostgreSQL, Keycloak, and others) alike. See [Third-Party Services](../../system-requirements/third-party-services.md#image-registry).

1. Download the `runai-installer` binary for your operating system and CPU architecture, on each machine you run the installer from. Replace `<VERSION>` with the NVIDIA Run:ai version you are installing, for example `3.0.17`:

   ```bash
   # Linux, AMD64 (x86_64)
   curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-linux-amd64
   chmod +x runai-installer-linux-amd64

   # Linux, ARM64
   curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-linux-arm64
   chmod +x runai-installer-linux-arm64

   # macOS, AMD64 (Intel)
   curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-darwin-amd64
   chmod +x runai-installer-darwin-amd64

   # macOS, ARM64 (Apple Silicon)
   curl -LO https://github.com/dsx-ai-factory/run-ai/releases/download/<VERSION>/runai-installer-darwin-arm64
   chmod +x runai-installer-darwin-arm64
   ```

2. Create a pack configuration file (`airgap-pack.yaml`) and run `airgap pack` against it:

   - `settings` - Required. Your target NVIDIA Run:ai version, architecture, and output path.
   - `imageOverrides` - Points specific third-party service images (for example, Keycloak) at your own source instead of the default one. Each entry also accepts `group`, required only when `name` exists in more than one image group - see [`imageOverrides`](../guided-installation/configuration-reference.md#pack-an-air-gapped-bundle) in the Configuration Reference.
   - `registryAuth` - Authenticates to your own registry hosting an `imageOverrides` source.
   - Third-party services - Add an entry to `imageOverrides` for each hardened third-party image, and an entry to `registryAuth` for `docker.io` with your Docker Hub credentials. See [Gov-Ready Third-Party Services](../../system-requirements/gov-ready-third-party-services.md) for the full example. <Badge intent="info" minimal outlined>Gov-ready</Badge>

   ```bash
   cat > airgap-pack.yaml <<'EOF'
   settings:
     runaiVersion: <RUNAI_VERSION>
     architecture: amd64
     output: runai-airgap-<RUNAI_VERSION>.tar.gz

   imageOverrides:
     - name: keycloak
       source: <your-registry>/keycloak:<tag>
     # Gov-ready: add an entry for each hardened image (keycloak, postgres, traefik, nats, thanos, valkey, git-sync)

   registryAuth:
     - registry: <your-registry-host>
       username: <username>
       passwordPath: <path-to-registry-password>
     # Gov-ready: add an entry for docker.io with your Docker Hub credentials
   EOF

   ./runai-installer airgap pack \
     --config airgap-pack.yaml
   ```

3. Transfer the resulting tarball into the air-gapped network.

**Load the Bundle (Air-Gapped Network)**

Loading happens once the tarball is in the air-gapped network, on any machine there with access to both the tarball and your internal mirror registry.

1. Extract the `runai-installer` binary from the transferred `runai-airgap-<RUNAI_VERSION>.tar.gz` bundle.
2. Create a load configuration file (`airgap-load.yaml`) and run `airgap load` against it:

```bash
cat > airgap-load.yaml <<'EOF'
settings:
  bundlePath: runai-airgap-<RUNAI_VERSION>.tar.gz
  workingDir: ./airgap-work
  valuesDir: ./airgap-values

registry:
  url: <REGISTRY_URL>
  username: <REGISTRY_USERNAME>
  passwordPath: <REGISTRY_PASSWORD_PATH>
EOF

./runai-installer airgap load \
  --config airgap-load.yaml
```

This extracts the bundle, pushes every image to `<REGISTRY_URL>`, and writes `control-plane-airgap-values.yaml` and `cluster-airgap-values.yaml` into `valuesDir`.

Pass these as `-f <VALUES_DIR>/control-plane-airgap-values.yaml` and `-f <VALUES_DIR>/cluster-airgap-values.yaml` to the `helm upgrade` commands in [Install the Control Plane](install-control-plane.md#installation) and [Install the Cluster](helm-install.md#installation).

`<VALUES_DIR>` is the `valuesDir` path you set in `airgap-load.yaml` above.

**Private Docker Registry**

To access the organization's docker registry, set the registry's credentials (imagePullSecret).

Create the secret named `runai-reg-creds` in the `runai-backend` namespace based on your existing credentials. The configuration will be copied over to the `runai` namespace at cluster install. For more information, see [Allowing pods to reference images from other secured registries](https://docs.openshift.com/container-platform/latest/openshift_images/managing_images/using-image-pull-secrets.html#images-allow-pods-to-reference-images-from-secure-registries_using-image-pull-secrets).

</Accordion>

## Next Steps

With your environment prepared, continue to [Prerequisites and Features](prerequisites-and-features.md) to install the GPU Operator, Prometheus, and any workload-specific features (such as distributed training or inference) the NVIDIA Run:ai cluster needs.
