# Network Requirements

The following network requirements are for the NVIDIA Run:ai components installation and usage.

## External Access

Set out below are the domains to whitelist and ports to open for installation, upgrade, and usage of the application and its management.

<Info>

**Note**

Ensure the inbound and outbound rules are correctly applied to your firewall.

</Info>

### Inbound Rules

To allow your organization’s NVIDIA Run:ai users to interact with the cluster using the NVIDIA Run:ai Command-line interface, or access specific UI features, certain inbound ports need to be open:

| Name                        | Description      | Source  | Destination                | Port |
| --------------------------- | ---------------- | ------- | -------------------------- | ---- |
| NVIDIA Run:ai control plane | HTTPS entrypoint | 0.0.0.0 | NVIDIA Run:ai system nodes | 443  |
| NVIDIA Run:ai cluster       | HTTPS entrypoint | 0.0.0.0 | NVIDIA Run:ai system nodes | 443  |

### Outbound Rules

<Info>

**Note**

* Outbound rules are applied to the NVIDIA Run:ai cluster component only. In case the NVIDIA Run:ai cluster is installed together with the NVIDIA Run:ai control plane, the NVIDIA Run:ai cluster FQDN refers to the NVIDIA Run:ai control plane FQDN.
* **For IPv6-only environments** - `runai.jfrog.io` and `nvcr.io` only have IPv4 DNS records, so clients on IPv6-only networks cannot resolve them and image pulls will fail. Two options:
  * **Configure NAT64/DNS64** - Translates between IPv6 and IPv4 so the cluster reaches these registries transparently.
  * **Deploy an internal mirror registry** - Use Harbor, Artifactory, or a similar registry over IPv6, configured to pull from `runai.jfrog.io` and `nvcr.io` over IPv4. Point the cluster at the mirror through the container runtime config and NVIDIA Run:ai Helm image-registry overrides. Choose this option for air-gapped or strictly controlled networks.
* `gcr.io`, `quay.io`, and `docker.io` are reachable over IPv6 directly.

</Info>

For the NVIDIA Run:ai cluster installation and usage, certain **outbound** ports must be open:

| Name                       | Description                                                                      | Source                                   | Destination                      | Port |
| -------------------------- | -------------------------------------------------------------------------------- | ---------------------------------------- | -------------------------------- | ---- |
| Cluster sync               | Sync NVIDIA Run:ai cluster with NVIDIA Run:ai control plane                      | NVIDIA Run:ai cluster system nodes       | NVIDIA Run:ai control plane FQDN | 443  |
| Metric store               | Push NVIDIA Run:ai cluster metrics to NVIDIA Run:ai control plane's metric store | NVIDIA Run:ai cluster system nodes       | NVIDIA Run:ai control plane FQDN | 443  |
| NVIDIA Run:ai NGC Registry | Pull NVIDIA Run:ai images and Helm chart for installation                        | All kubernetes nodes                     | nvcr.io                          | 443  |
| NVIDIA NGC                 | Browse NGC catalog                                                               | NVIDIA Run:ai control plane system nodes | api.ngc.nvidia.com               | 443  |
| Hugging Face               | Browse Hugging Face models                                                       | NVIDIA Run:ai control plane system nodes | huggingface.co                   | 443  |

The NVIDIA Run:ai installation has [software requirements](../../organization-life-cycle/organization-onboarding/system-requirements.md) that require additional components to be installed on the cluster. This article includes simple installation examples which can be used optionally and require the following cluster outbound ports to be open:

| Name                       | Description                                | Source               | Destination | Port |
| -------------------------- | ------------------------------------------ | -------------------- | ----------- | ---- |
| Kubernetes Registry        | Ingress HAProxy image repository           | All kubernetes nodes | docker.io   | 443  |
| Google Container Registry  | GPU Operator, and Knative image repository | All kubernetes nodes | gcr.io      | 443  |
| Red Hat Container Registry | Prometheus Operator image repository       | All kubernetes nodes | quay.io     | 443  |
| Docker Hub Registry        | Training Operator image repository         | All kubernetes nodes | docker.io   | 443  |

## Internal Network

Ensure that all Kubernetes nodes can communicate with each other across all necessary ports. Kubernetes assumes full interconnectivity between nodes, so you must configure your network to allow this seamless communication. Specific port requirements may vary depending on your network setup.
