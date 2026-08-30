# Install the Cluster

In this section you will install the NVIDIA Run:ai cluster on your Kubernetes environment using Helm. The cluster extends Kubernetes with NVIDIA Run:ai orchestration capabilities - scheduling and workload management - and connects to the previously installed [control plane](install-control-plane.md) for centralized management.

Once the control plane is installed and you access the NVIDIA Run:ai UI for the first time, an onboarding wizard opens automatically. The wizard guides you through the cluster setup and generates a Helm installation command. Follow the instructions below to modify and run the command based on your artifact source and environment.

This procedure includes:

* Adding the NVIDIA Run:ai Helm repository from NGC
* Installing the NVIDIA Run:ai cluster into the `runai` namespace
* Registering the NVIDIA Run:ai cluster with the NVIDIA Run:ai control plane using the provided connection details

By completing this process, the NVIDIA Run:ai cluster will be connected to the NVIDIA Run:ai control plane and ready to run training, inference, and other workloads.

## System and Network Requirements <a href="#system-and-network-requirements" id="system-and-network-requirements"></a>

Before installing the NVIDIA Run:ai cluster, validate that the [system requirements](system-requirements.md) and [network requirements](network-requirements.md) are met. For air-gapped environments, make sure you have the [software artifacts](preparations.md#software-artifacts) prepared.

Once all the requirements are met, it is highly recommended to use the NVIDIA Run:ai cluster preinstall diagnostics tool to:

* Test the below requirements in addition to failure points related to Kubernetes, NVIDIA, storage, and networking
* Look at additional components installed and analyze their relevance to a successful installation

For more information, see [preinstall diagnostics](https://github.com/run-ai/preinstall-diagnostics). To run the preinstall diagnostics tool, [download](https://runai.jfrog.io/ui/native/pd-cli-prod/preinstall-diagnostics-cli/) the latest version, and run:

<Tabs>

<Tab title="Connected">

```bash
chmod +x ./preinstall-diagnostics-<platform> && \ 
./preinstall-diagnostics-<platform> \
  --domain ${CONTROL_PLANE_FQDN} \
  --cluster-domain ${CLUSTER_FQDN} \
#if the diagnostics image is hosted in a private registry
  --image-pull-secret ${IMAGE_PULL_SECRET_NAME} \
  --image ${PRIVATE_REGISTRY_IMAGE_URL}    
```

</Tab>

<Tab title="Air-gapped">

In an air-gapped deployment, the diagnostics image is saved, pushed, and pulled manually from the organization's registry.

```bash
#Save the image locally
docker save --output preinstall-diagnostics.tar gcr.io/run-ai-lab/preinstall-diagnostics:${VERSION}
#Load the image to the organization's registry
docker load --input preinstall-diagnostics.tar
docker tag gcr.io/run-ai-lab/preinstall-diagnostics:${VERSION} ${CLIENT_IMAGE_AND_TAG} 
docker push ${CLIENT_IMAGE_AND_TAG}
```

Run the binary with the `--image` parameter to modify the diagnostics image to be used:

```bash
chmod +x ./preinstall-diagnostics-darwin-arm64 && \
./preinstall-diagnostics-darwin-arm64 \
  --domain ${CONTROL_PLANE_FQDN} \
  --cluster-domain ${CLUSTER_FQDN} \
  --image-pull-secret ${IMAGE_PULL_SECRET_NAME} \
  --image ${PRIVATE_REGISTRY_IMAGE_URL}    
```

</Tab>

</Tabs>

## Helm

NVIDIA Run:ai requires [Helm](https://helm.sh/) 3.14 or later. To install Helm, see [Installing Helm](https://helm.sh/docs/intro/install/). If you are installing an air-gapped version of NVIDIA Run:ai, the NVIDIA Run:ai tar file contains the [helm binary](preparations.md#software-artifacts).

## Permissions <a href="#permissions" id="permissions"></a>

Using a Kubernetes user with the `cluster-admin` role to ensure a successful installation is recommended. For more information, see [Using RBAC authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

## Installation

<Info>

**Note**

* To customize the installation based on your environment, see [Advanced cluster configurations](../../post-installation-configuration/cluster-config.md).
* You can store the `clientSecret` as a Kubernetes secret within the cluster instead of using plain text. You can then configure the installation to use it by setting the `controlPlane.existingSecret` and `controlPlane.secretKeys.clientSecret` parameters as described in [Advanced cluster configurations](../../post-installation-configuration/cluster-config.md).

</Info>

### Kubernetes

<Accordion title="Connected">

**Connected**

When adding a cluster for the first time, the onboarding wizard opens automatically when you log in to the NVIDIA Run:ai platform. You cannot perform other actions in the platform until the cluster is created.

1. Enter a unique name for your cluster.
2. Set the cluster location. Choose where the NVIDIA Run:ai cluster will be installed:
   * Same as the control plane - Install the NVIDIA Run:ai cluster on the same Kubernetes cluster as the NVIDIA Run:ai control plane.
   *   Remote control plane - Install the NVIDIA Run:ai cluster on a different Kubernetes cluster than the NVIDIA Run:ai control plane.

       <Info><p><strong>Note</strong></p><p>The selected location must align with the <a href="system-requirements.md">system</a> requirements you prepared earlier. The NVIDIA Run:ai cluster system requirements differ depending on whether the NVIDIA Run:ai cluster is installed on the same Kubernetes cluster as the NVIDIA Run:ai control plane or on a separate one.</p></Info>
3. If you selected Remote control plane, enter the Cluster URL. For more information, see [Fully Qualified Domain Name](system-requirements.md#fully-qualified-domain-name-fqdn) requirement.

**Install NVIDIA Run:ai on Your Cluster**

In the next section, the NVIDIA Run:ai cluster installation steps will be presented.

1. Before installing the NVIDIA Run:ai cluster, ensure that all required [system](system-requirements.md) and [network](network-requirements.md) requirements are met.
2. The NVIDIA Run:ai platform displays the Helm installation command in the cluster wizard.
3. Modify the UI-generated command as follows:
   * Replace `<NGC_API_KEY>` with your NGC API key.
   * If you are using a local certificate authority, add `--set global.customCA.enabled=true` to the Helm command as described in the [Local certificate authority](cp-system-requirements.md#local-certificate-authority) section.
   * The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class to match the ingress controller configured during the [control plane installation](install-control-plane.md#connected).

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password=<NGC_API_KEY>
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
    --set controlPlane.url=... \
    --set controlPlane.clientSecret=... \
    --set cluster.uid=... \
    --set cluster.url=... --version="<VERSION>" --create-namespace \
    --set clusterConfig.global.ingress.ingressClass=haproxy
```

The wizard displays **Waiting for cluster to connect** while the cluster is being installed and connected to the control plane. Once the installation completes successfully and the cluster establishes communication with the control plane, the wizard updates to **Cluster connected**. After completing the wizard flow, the cluster is added to the [Clusters](../../../infrastructure-setup/procedures/clusters.md#clusters-table) table.

<Info>

**Tip**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Info>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

<Warning>

**Prerequisite**

If your internal registry requires authentication, you must create the `runai-reg-creds` imagePullSecret before proceeding. See [Private Docker Registry](preparations.md#private-docker-registry) in Preparations.

</Warning>

When adding a cluster for the first time, the onboarding wizard opens automatically when you log in to the NVIDIA Run:ai platform. You cannot perform other actions in the platform until the cluster is created.

1. Enter a unique name for your cluster.
2. Set the cluster location. Choose where the NVIDIA Run:ai cluster will be installed:
   * Same as the control plane - Install the NVIDIA Run:ai cluster on the same Kubernetes cluster as the NVIDIA Run:ai control plane.
   *   Remote control plane - Install the NVIDIA Run:ai cluster on a different Kubernetes cluster than the NVIDIA Run:ai control plane.

       <Info><p><strong>Note</strong></p><p>The selected location must align with the <a href="system-requirements.md">system</a> requirements you prepared earlier. The NVIDIA Run:ai cluster system requirements differ depending on whether the NVIDIA Run:ai cluster is installed on the same Kubernetes cluster as the NVIDIA Run:ai control plane or on a separate one.</p></Info>
3. If you selected Remote control plane, enter the Cluster URL. For more information, see [Fully Qualified Domain Name](system-requirements.md#fully-qualified-domain-name-fqdn) requirement.

**Install NVIDIA Run:ai on Your Cluster**

In the next section, the NVIDIA Run:ai cluster installation steps will be presented.

1. Before installing the NVIDIA Run:ai cluster, ensure that all required [system](system-requirements.md) and [network](network-requirements.md) requirements are met.
2.  The NVIDIA Run:ai platform displays the Helm installation command in the cluster wizard.

    **— Do not run the command exactly as shown in the UI —**
3. Update the UI-generated Helm command as follows (see example command below) and use the [pre-provided installation file](preparations.md#air-gapped) instead of using helm repositories:
   * Do not add the helm repository, `helm repo add`, and do not run `helm repo update`.
   * Instead, edit the `helm upgrade` command:
     * Replace `runai/runai-cluster` with `./chart/runai-cluster-<VERSION>.tgz`, where `<VERSION>` is the full version number (e.g., `runai-cluster-2.24.58.tgz`). This file is located in the `chart` folder of the extracted software artifacts.
     * Add `--set global.image.registry=<DOCKER REGISTRY ADDRESS>` where `<DOCKER_REGISTRY_ADDRESS>` is the Docker registry address configured in the [Preparations](preparations.md#air-gapped) section.
     * Add `--set clusterConfig.prometheus.spec.baseImage=<DOCKER REGISTRY ADDRESS>/<FULL_IMAGE_PATH>`. The registry address should point to the location where the Prometheus image is hosted.
     * Add `--set global.customCA.enabled=true` as described in the [Local certificate authority](system-requirements.md#local-certificate-authority) section.
     * The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class to match the ingress controller configured during the [control plane installation](install-control-plane.md#connected).
     * Keep the remaining `--set` values exactly as generated by the UI.

Run the following command from the root of the extracted software artifacts directory:

```bash
helm upgrade -i runai-cluster ./chart/runai-cluster-<VERSION>.tgz \
    --set controlPlane.url=... \
    --set controlPlane.clientSecret=... \
    --set cluster.uid=... \
    --set cluster.url=... --create-namespace \
    --set global.image.registry=registry.mycompany.local \
    --set clusterConfig.prometheus.spec.baseImage=registry.mycompany.local/prometheus/prometheus \
    --set global.customCA.enabled=true \
    --set clusterConfig.global.ingress.ingressClass=haproxy
```

The wizard displays **Waiting for cluster to connect** while the cluster is being installed and connected to the control plane. Once the installation completes successfully and the cluster establishes communication with the control plane, the wizard updates to **Cluster connected**. After completing the wizard flow, the cluster is added to the [Clusters](../../../infrastructure-setup/procedures/clusters.md#clusters-table) table.

<Info>

**Tip**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Info>

</Accordion>

### OpenShift

<Accordion title="Connected">

**Connected**

When adding a cluster for the first time, the onboarding wizard opens automatically when you log in to the NVIDIA Run:ai platform. You cannot perform other actions in the platform until the cluster is created.

1. Enter a unique name for your cluster.
2. Set the cluster location. Choose where the NVIDIA Run:ai cluster will be installed:
   * Same as the control plane - Install the NVIDIA Run:ai cluster on the same Kubernetes cluster as the NVIDIA Run:ai control plane.
   *   Remote control plane - Install the NVIDIA Run:ai cluster on a different Kubernetes cluster than the NVIDIA Run:ai control plane.

       <Info><p><strong>Note</strong></p><p>The selected location must align with the <a href="system-requirements.md">system</a> requirements you prepared earlier. The NVIDIA Run:ai cluster system requirements differ depending on whether the NVIDIA Run:ai cluster is installed on the same Kubernetes cluster as the NVIDIA Run:ai control plane or on a separate one.</p></Info>
3. If you selected Remote control plane, enter the Cluster URL. For more information, see [Fully Qualified Domain Name](system-requirements.md#fully-qualified-domain-name-fqdn) requirement.

**Install NVIDIA Run:ai on Your Cluster**

In the next section, the NVIDIA Run:ai cluster installation steps will be presented.

1. Before installing the NVIDIA Run:ai cluster, ensure that all required [system](system-requirements.md) and [network](network-requirements.md) requirements are met.
2. The NVIDIA Run:ai platform displays the Helm installation command in the cluster wizard.
3. Modify the UI-generated command as follows:
   * Replace `<NGC_API_KEY>` with your NGC API key.
   * If you are using a local certificate authority, add `--set global.customCA.enabled=true` to the Helm command as described in the [Local certificate authority](cp-system-requirements.md#local-certificate-authority) section.

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password=<NGC_API_KEY>
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
  --set controlPlane.url=... \
  --set controlPlane.clientSecret=... \
  --set cluster.uid=... \
  --set cluster.url=... --version="<VERSION>" --create-namespace
```

The wizard displays **Waiting for cluster to connect** while the cluster is being installed and connected to the control plane. Once the installation completes successfully and the cluster establishes communication with the control plane, the wizard updates to **Cluster connected**. After completing the wizard flow, the cluster is added to the [Clusters](../../../infrastructure-setup/procedures/clusters.md#clusters-table) table.

<Info>

**Tip**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Info>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

<Warning>

**Prerequisite**

If your internal registry requires authentication, you must create the `runai-reg-creds` imagePullSecret before proceeding. See [Private Docker Registry](preparations.md#private-docker-registry) in Preparations.

</Warning>

When adding a cluster for the first time, the onboarding wizard opens automatically when you log in to the NVIDIA Run:ai platform. You cannot perform other actions in the platform until the cluster is created.

1. Enter a unique name for your cluster.
2. Set the cluster location. Choose where the NVIDIA Run:ai cluster will be installed:
   * Same as the control plane - Install the NVIDIA Run:ai cluster on the same Kubernetes cluster as the NVIDIA Run:ai control plane.
   *   Remote control plane - Install the NVIDIA Run:ai cluster on a different Kubernetes cluster than the NVIDIA Run:ai control plane.

       <Info><p><strong>Note</strong></p><p>The selected location must align with the <a href="system-requirements.md">system</a> requirements you prepared earlier. The NVIDIA Run:ai cluster system requirements differ depending on whether the NVIDIA Run:ai cluster is installed on the same Kubernetes cluster as the NVIDIA Run:ai control plane or on a separate one.</p></Info>
3. If you selected Remote control plane, enter the Cluster URL. For more information, see [Fully Qualified Domain Name](system-requirements.md#fully-qualified-domain-name-fqdn) requirement.

**Install NVIDIA Run:ai on Your Cluster**

In the next section, the NVIDIA Run:ai cluster installation steps will be presented.

1. Before installing the NVIDIA Run:ai cluster, ensure that all required [system](system-requirements.md) and [network](network-requirements.md) requirements are met.
2.  The NVIDIA Run:ai platform displays the Helm installation command in the cluster wizard.

    **— Do not run the command exactly as shown in the UI —**
3.  Update the UI-generated Helm command as follows (see example command below) and use the [pre-provided installation file](preparations.md#air-gapped-1) instead of using helm repositories:

    * Do not add the helm repository and do not run `helm repo update`.
    * Instead, edit the `helm upgrade` command:
      * Replace `runai/runai-cluster` with `./chart/runai-cluster-<VERSION>.tgz`, where `<VERSION>` is the full version number (e.g., `runai-cluster-2.24.58.tgz`). This file is located in the `chart` folder of the extracted software artifacts.
      * Add `--set global.image.registry=<DOCKER REGISTRY ADDRESS>` where `<DOCKER_REGISTRY_ADDRESS>` is the Docker registry address configured in the [Preparations](preparations.md#air-gapped) section.
      * Add `--set clusterConfig.prometheus.spec.baseImage=<DOCKER REGISTRY ADDRESS>/<FULL_IMAGE_PATH>`. The registry address should point to the location where the Prometheus image is hosted.
      * Add `--set global.customCA.enabled=true` as described in the [Local certificate authority](system-requirements.md#local-certificate-authority) section.
      * Keep the remaining `--set` values exactly as generated by the UI.

    Run the following command from the root of the extracted software artifacts directory:

    ```bash
    helm upgrade -i runai-cluster ./chart/runai-cluster-<VERSION>.tgz \
        --set controlPlane.url=... \
        --set controlPlane.clientSecret=... \
        --set cluster.uid=... \
        --set cluster.url=... --create-namespace \
        --set global.image.registry=registry.mycompany.local \
        --set clusterConfig.prometheus.spec.baseImage=registry.mycompany.local/prometheus/prometheus \
        --set global.customCA.enabled=true
    ```

The wizard displays **Waiting for cluster to connect** while the cluster is being installed and connected to the control plane. Once the installation completes successfully and the cluster establishes communication with the control plane, the wizard updates to **Cluster connected**. After completing the wizard flow, the cluster is added to the [Clusters](../../../infrastructure-setup/procedures/clusters.md#clusters-table) table.

<Info>

**Tip**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Info>

</Accordion>

## Troubleshooting <a href="#troubleshooting" id="troubleshooting"></a>

If you encounter an issue with the installation, try the troubleshooting scenario below.

### Installation <a href="#installation_1" id="installation_1"></a>

If the NVIDIA Run:ai cluster installation failed, check the installation logs to identify the issue. Run the following script to print the installation logs:

[get-installation-logs.sh](../../../assets/get-installation-logs.sh)

### Cluster Status <a href="#cluster-status" id="cluster-status"></a>

If the NVIDIA Run:ai cluster installation completed, but the cluster status did not change its status to **Connected**, refer to the cluster [Troubleshooting scenarios](../../../infrastructure-setup/procedures/clusters.md#troubleshooting-scenarios) section.

## Next Steps

Once the cluster is installed and connected, the NVIDIA Run:ai UI guides you through optional post-installation configurations. These steps are optional but recommended for a production setup:

* **SSO (Single Sign-On)** - Configure SSO to allow users to log in with your organization's identity provider. See [SSO](../../../infrastructure-setup/authentication/sso/saml.md) for setup instructions using SAML or OpenID Connect.
* **Email server** - Configure an SMTP server to enable email notifications for workload events and system alerts. See [Email notifications](../../../settings/general-settings/notifications.md#email-notifications) for configuration details.
* **Create your first research team** - Set up [projects](../../../platform-management/aiinitiatives/organization/projects.md) and [permissions](../../../infrastructure-setup/authentication/accessrules.md) to organize your AI practitioners and allocate GPU resources.
