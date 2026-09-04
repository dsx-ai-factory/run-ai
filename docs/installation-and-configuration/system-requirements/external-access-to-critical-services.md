# External Access to Critical Services

NVIDIA Run:ai requires certain network connectivity and access. This page outlines the network endpoints and protocols that must be reachable from your NVIDIA Run:ai control plane and cluster nodes to support installation, artifact retrieval, and ongoing platform communication.

Meeting these network requirements ensures that:

* Users and API clients can reach the NVIDIA Run:ai control plane and cluster
* The NVIDIA Run:ai cluster can sync with and report metrics to the control plane
* The control plane and cluster nodes can pull the images required for installation

<img src="../../assets/firewall-registry-access.svg" alt="Network access map: inbound access from users and API clients, cluster sync and metrics between the control plane and cluster in multi-cluster deployments, internal node-to-node traffic, and outbound access to the NVIDIA Run:ai, third-party service, and prerequisite and feature image registries" />

| Category | Requirement | Details |
|---|---|---|
| Inbound | Ports open for user and API access | See [Inbound Rules](#inbound-rules) |
| Outbound | Control plane and cluster sync | See [Control Plane and Cluster Sync](#control-plane-and-cluster-sync) |
| Outbound | NVIDIA Run:ai image registry access | See [NVIDIA Run:ai Registry](#nvidia-run-ai-registry) |
| Outbound | Third-party service registry access | See [Third-Party Service Registries](#third-party-service-registries) |
| Outbound | Prerequisite and feature component registry access | See [Prerequisite and Feature Registries](#prerequisite-and-feature-registries) |
| Internal | Node-to-node connectivity | See [Internal Network](#internal-network) |

<Info>

**Note**

Ensure the inbound and outbound rules below are correctly applied to your firewall before proceeding with installation.

</Info>

## Inbound Rules

To allow your organization's NVIDIA Run:ai users to interact with the cluster using the [NVIDIA Run:ai Command-line interface](../../../reference/cli/runai/), or access specific UI features, certain inbound ports need to be open:

| Name                        | Description       | Source  | Destination                | Port |
| --------------------------- | ----------------- | ------- | --------------------------- | ---- |
| NVIDIA Run:ai control plane | HTTPS entry point | 0.0.0.0 | NVIDIA Run:ai system nodes | 443  |
| NVIDIA Run:ai cluster       | HTTPS entry point | 0.0.0.0 | NVIDIA Run:ai system nodes | 443  |

## Outbound Rules

### Control Plane and Cluster Sync

<Info>

**Note**

These rules apply only to [multi-cluster deployments](deployment-model.md), where the NVIDIA Run:ai cluster is installed on a separate Kubernetes cluster from the control plane. In a same-cluster deployment, the control plane and cluster run on the same Kubernetes cluster and communicate internally.

</Info>

| Name          | Description                                                                      | Source                             | Destination                      | Port |
| ------------- | --------------------------------------------------------------------------------- | ----------------------------------- | --------------------------------- | ---- |
| Cluster sync  | Sync NVIDIA Run:ai cluster with NVIDIA Run:ai control plane                       | NVIDIA Run:ai cluster system nodes | NVIDIA Run:ai control plane FQDN | 443  |
| Metric store  | Push NVIDIA Run:ai cluster metrics to NVIDIA Run:ai control plane's metric store | NVIDIA Run:ai cluster system nodes | NVIDIA Run:ai control plane FQDN | 443  |

### NVIDIA Run:ai Registry

| Name                       | Description                                                 | Source                | Destination | Port |
| -------------------------- | ------------------------------------------------------------ | ---------------------- | ----------- | ---- |
| NVIDIA Run:ai NGC Registry | Pull NVIDIA Run:ai images and Helm chart for installation   | All Kubernetes nodes   | nvcr.io     | 443  |

### Third-Party Service Registries

NVIDIA Run:ai deploys the following third-party services as part of installation. The registry each is pulled from depends on your installation type - for OSS and NVIDIA Supported installations, images come from Docker Hub, Quay, or the Kubernetes registry; for Gov-ready installations, images come from Chainguard. For the full registry model, see [Image Registry](../third-party-services.md#image-registry) in Third-Party Services.

| Name          | Description                        | Source                              | Destination      | Destination (Gov-ready) | Port |
| ------------- | ------------------------------------ | ------------------------------------ | ------------------ | -------------------------- | ---- |
| Keycloak      | Identity federation and SSO service | NVIDIA Run:ai control plane system nodes | docker.io / quay.io | Chainguard | 443  |
| PostgreSQL    | Centralized platform database       | NVIDIA Run:ai control plane system nodes | docker.io / quay.io | Chainguard | 443  |
| Traefik       | API gateway                         | NVIDIA Run:ai control plane system nodes | docker.io / quay.io | Chainguard | 443  |
| NATS          | Message broker                      | NVIDIA Run:ai control plane system nodes | docker.io / quay.io | Chainguard | 443  |
| Thanos        | Long-term metrics storage           | NVIDIA Run:ai control plane system nodes | docker.io / quay.io | Chainguard | 443  |
| Valkey        | Workload cache                      | NVIDIA Run:ai cluster system nodes  | docker.io / quay.io | Chainguard | 443  |
| oauth2-proxy  | Workload access authentication      | NVIDIA Run:ai cluster system nodes  | docker.io / quay.io | Chainguard | 443  |
| git-sync      | Git data source sync                | NVIDIA Run:ai cluster system nodes  | registry.k8s.io      | Chainguard | 443  |

### Prerequisite and Feature Registries

The NVIDIA Run:ai installation has [software requirements](../install-using-helm/system-requirements.md) that require additional components to be installed on the cluster. These prerequisite components are pulled from the following registries:

| Name                        | Description                                     | Source                | Destination | Port |
| ---------------------------- | ------------------------------------------------ | ---------------------- | ----------- | ---- |
| Ingress Controller Registry  | HAProxy ingress controller image repository (registry source TBD) | All Kubernetes nodes | docker.io   | 443  |
| Google Container Registry    | GPU Operator (registry source TBD) and Knative image repository | All Kubernetes nodes   | gcr.io      | 443  |
| Red Hat Container Registry   | Prometheus Operator image repository             | All Kubernetes nodes   | quay.io     | 443  |
| Docker Hub Registry          | Training Operator image repository               | All Kubernetes nodes   | docker.io   | 443  |

## Internal Network

Ensure that all Kubernetes nodes can communicate with each other across all necessary ports. Kubernetes assumes full interconnectivity between nodes, so you must configure your network to allow this seamless communication. Specific port requirements may vary depending on your network setup.

## IPv6 Considerations

For IPv6-only environments, `nvcr.io` only has an IPv4 DNS record, so clients on IPv6-only networks cannot resolve it and image pulls will fail. Two options:

* **Configure NAT64/DNS64** - Translates between IPv6 and IPv4 so the cluster reaches the registry transparently.
* **Deploy an internal mirror registry** - Use Harbor, Artifactory, or a similar registry over IPv6, configured to pull from `nvcr.io` over IPv4. Point the cluster at the mirror through the container runtime config and NVIDIA Run:ai Helm image-registry overrides. Choose this option for air-gapped or strictly controlled networks.

`gcr.io`, `quay.io`, and `docker.io` are reachable over IPv6 directly.
