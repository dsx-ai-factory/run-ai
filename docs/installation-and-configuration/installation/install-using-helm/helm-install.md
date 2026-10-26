# Install the Cluster

This page describes how to install the NVIDIA Run:ai cluster on your Kubernetes environment using Helm. The cluster extends Kubernetes with NVIDIA Run:ai orchestration capabilities - scheduling and workload management - and connects to the previously installed [control plane](install-control-plane.md) for centralized management.

Once the control plane is installed and you access the NVIDIA Run:ai UI for the first time, an onboarding wizard opens automatically. The wizard guides you through the cluster setup and generates a Helm installation command. Follow the instructions below to modify and run the command based on your artifact source and environment.

This procedure includes:

* Installing the NVIDIA Run:ai cluster into the `runai` namespace
* Registering the NVIDIA Run:ai cluster with the NVIDIA Run:ai control plane using the provided connection details

By completing this process, the NVIDIA Run:ai cluster will be connected to the NVIDIA Run:ai control plane and ready to run training, inference, and other workloads.

## Before You Start

Before installing the NVIDIA Run:ai cluster, complete [Preparations](preparations.md) and [Prerequisites and Features](prerequisites-and-features.md), and make sure you have [installed the control plane](install-control-plane.md).

## Helm

NVIDIA Run:ai requires [Helm](https://helm.sh/) 3.14 or later. To install Helm, see [Installing Helm](https://helm.sh/docs/intro/install/). If you are installing an air-gapped version of NVIDIA Run:ai, the NVIDIA Run:ai tar file contains the [helm binary](preparations.md#software-artifacts).

<Note>

**Note**

Helm 4 defaults to [server-side apply](https://helm.sh/docs/overview/#server-side-apply) when installing a new chart release, which can conflict with resources managed by the NVIDIA Run:ai operator. Append `--server-side=false` to your `helm upgrade` command. NVIDIA Run:ai clusters originally installed with Helm 3.x are unaffected.

</Note>

## Permissions <a href="#permissions" id="permissions"></a>

Using a Kubernetes user with the `cluster-admin` role to ensure a successful installation is recommended. For more information, see [Using RBAC authorization](https://kubernetes.io/docs/reference/access-authn-authz/rbac/).

## Installation

<Note>

**Note**

* To customize the installation based on your environment, see [Advanced cluster configurations](../../configuration/cluster-config.md).
* You can store the `clientSecret` as a Kubernetes secret within the cluster instead of using plain text. You can then configure the installation to use it by setting the `controlPlane.existingSecret` and `controlPlane.secretKeys.clientSecret` parameters as described in [Advanced cluster configurations](../../configuration/cluster-config.md).
* The `bootstrapToken` in the generated Helm command is a short-lived token that expires 2 hours after generation. If you do not complete the installation within that window, return to the cluster wizard in the NVIDIA Run:ai UI to generate a new command before running it.

</Note>

### Kubernetes

<Accordion title="Connected">

**Connected**

When adding a cluster for the first time, the onboarding wizard opens automatically when you log in to the NVIDIA Run:ai platform. You cannot perform other actions in the platform until the cluster is created. For additional clusters, use the Clusters form instead. See [Clusters](../../operations/clusters.md).

1. Enter a unique name for your cluster.
2. Set the cluster location. Choose where the NVIDIA Run:ai cluster will be installed:
   * Same as the control plane - Install the NVIDIA Run:ai cluster on the same Kubernetes cluster as the NVIDIA Run:ai control plane.
   *   Remote control plane - Install the NVIDIA Run:ai cluster on a different Kubernetes cluster than the NVIDIA Run:ai control plane.

       <Note><p><strong>Note</strong></p><p>This selection depends on your <a href="../../system-requirements/deployment-model.md">deployment model</a> - the NVIDIA Run:ai cluster system requirements differ depending on whether the NVIDIA Run:ai cluster is installed on the same Kubernetes cluster as the NVIDIA Run:ai control plane (same-cluster deployment) or on a separate one (multi-cluster deployment).</p></Note>
3. If you selected Remote control plane, enter the Cluster URL. For more information, see [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns) requirement.

**Install NVIDIA Run:ai on Your Cluster**

In the next section, the NVIDIA Run:ai cluster installation steps will be presented.

1. Before installing the NVIDIA Run:ai cluster, see [Preparations](preparations.md) and follow the instructions there.
2. The NVIDIA Run:ai platform displays the Helm installation command in the cluster wizard.
3. Modify the UI-generated command as follows:
   * If you are using a local certificate authority, add `--set global.customCA.enabled=true` to the Helm command as described in the [Cluster Local CA](preparations.md#cluster-local-ca) section.
   * The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class to match the ingress controller configured during the [control plane installation](install-control-plane.md#kubernetes).
   * Add `-f <GOV_READY_VALUES_FILE>` to the Helm command. This file holds the Gov-ready values: the image pull secrets and the hardened images of the cluster third-party services. See [Gov-Ready Third-Party Services](../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
   * Add `--set global.fipsMode=only` to the Helm command to enable FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. <Badge intent="info" minimal outlined>Gov-ready</Badge>

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
    --set controlPlane.url=... \
    --set controlPlane.bootstrapToken=... \
    --set cluster.uid=... \
    --set cluster.url=... --version="<VERSION>" --create-namespace \
    --set clusterConfig.global.ingress.ingressClass=haproxy
```

The wizard displays **Waiting for cluster to connect** while the cluster is being installed and connected to the control plane. Once the installation completes successfully and the cluster establishes communication with the control plane, the wizard updates to **Cluster connected**. After completing the wizard flow, the cluster is added to the [Clusters](../../operations/clusters.md#clusters-table) table.

<Note>

**Tip**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Note>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

When adding a cluster for the first time, the onboarding wizard opens automatically when you log in to the NVIDIA Run:ai platform. You cannot perform other actions in the platform until the cluster is created. For additional clusters, use the Clusters form instead. See [Clusters](../../operations/clusters.md).

1. Enter a unique name for your cluster.
2. Set the cluster location. Choose where the NVIDIA Run:ai cluster will be installed:
   * Same as the control plane - Install the NVIDIA Run:ai cluster on the same Kubernetes cluster as the NVIDIA Run:ai control plane.
   *   Remote control plane - Install the NVIDIA Run:ai cluster on a different Kubernetes cluster than the NVIDIA Run:ai control plane.

       <Note><p><strong>Note</strong></p><p>This selection depends on your <a href="../../system-requirements/deployment-model.md">deployment model</a> - the NVIDIA Run:ai cluster system requirements differ depending on whether the NVIDIA Run:ai cluster is installed on the same Kubernetes cluster as the NVIDIA Run:ai control plane (same-cluster deployment) or on a separate one (multi-cluster deployment).</p></Note>
3. If you selected Remote control plane, enter the Cluster URL. For more information, see [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns) requirement.

**Install NVIDIA Run:ai on Your Cluster**

In the next section, the NVIDIA Run:ai cluster installation steps will be presented.

1. Before installing the NVIDIA Run:ai cluster, see [Preparations](preparations.md) and follow the instructions there.
2.  The NVIDIA Run:ai platform displays the Helm installation command in the cluster wizard.

    **- Do not run the command exactly as shown in the UI -**
3. Update the UI-generated Helm command as follows (see example command below) and use the `cluster-airgap-values.yaml` file written by the `airgap load` step in [Preparations](preparations.md#software-artifacts), instead of using helm repositories:
   * Do not add the helm repository, `helm repo add`, and do not run `helm repo update`.
   * Instead, edit the `helm upgrade` command:
     * Replace `runai/runai-cluster` with `./charts/runai-cluster-<VERSION>.tgz`, where `<VERSION>` is the full version number. This file is located in the `charts` folder of the working directory you specified to `airgap load`.
     * Add `-f <VALUES_DIR>/cluster-airgap-values.yaml`. Replace `<VALUES_DIR>` with the `valuesDir` path you set in `airgap-load.yaml` during [Preparations](preparations.md#software-artifacts) - this is where `cluster-airgap-values.yaml` was written.
     * If you are using a local certificate authority, add `--set global.customCA.enabled=true` as described in the [Cluster Local CA](preparations.md#cluster-local-ca) section.
     * The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class to match the ingress controller configured during the [control plane installation](install-control-plane.md#kubernetes).
     * No Gov-ready values are needed for the cluster third-party services. Only their images change, and the generated values file already sets them. <Badge intent="info" minimal outlined>Gov-ready</Badge>
     * Add `--set global.fipsMode=only` to the Helm command to enable FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. <Badge intent="info" minimal outlined>Gov-ready</Badge>
     * Keep the remaining `--set` values exactly as generated by the UI.

```bash
helm upgrade -i runai-cluster ./charts/runai-cluster-<VERSION>.tgz \
    --set controlPlane.url=... \
    --set controlPlane.bootstrapToken=... \
    --set cluster.uid=... \
    --set cluster.url=... --create-namespace \
    --set clusterConfig.global.ingress.ingressClass=haproxy \
    -f <VALUES_DIR>/cluster-airgap-values.yaml
```

The wizard displays **Waiting for cluster to connect** while the cluster is being installed and connected to the control plane. Once the installation completes successfully and the cluster establishes communication with the control plane, the wizard updates to **Cluster connected**. After completing the wizard flow, the cluster is added to the [Clusters](../../operations/clusters.md#clusters-table) table.

<Note>

**Tip**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Note>

</Accordion>

### OpenShift

<Accordion title="Connected">

**Connected**

When adding a cluster for the first time, the onboarding wizard opens automatically when you log in to the NVIDIA Run:ai platform. You cannot perform other actions in the platform until the cluster is created. For additional clusters, use the Clusters form instead. See [Clusters](../../operations/clusters.md).

1. Enter a unique name for your cluster.
2. Set the cluster location. Choose where the NVIDIA Run:ai cluster will be installed:
   * Same as the control plane - Install the NVIDIA Run:ai cluster on the same Kubernetes cluster as the NVIDIA Run:ai control plane.
   *   Remote control plane - Install the NVIDIA Run:ai cluster on a different Kubernetes cluster than the NVIDIA Run:ai control plane.

       <Note><p><strong>Note</strong></p><p>This selection depends on your <a href="../../system-requirements/deployment-model.md">deployment model</a> - the NVIDIA Run:ai cluster system requirements differ depending on whether the NVIDIA Run:ai cluster is installed on the same Kubernetes cluster as the NVIDIA Run:ai control plane (same-cluster deployment) or on a separate one (multi-cluster deployment).</p></Note>
3. If you selected Remote control plane, enter the Cluster URL. For more information, see [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns) requirement.

**Install NVIDIA Run:ai on Your Cluster**

In the next section, the NVIDIA Run:ai cluster installation steps will be presented.

1. Before installing the NVIDIA Run:ai cluster, see [Preparations](preparations.md) and follow the instructions there.
2. The NVIDIA Run:ai platform displays the Helm installation command in the cluster wizard.
3. Modify the UI-generated command as follows:
   * If you are using a local certificate authority, add `--set global.customCA.enabled=true` to the Helm command as described in the [Cluster Local CA](preparations.md#cluster-local-ca) section.
   * Add `-f <GOV_READY_VALUES_FILE>` to the Helm command. This file holds the Gov-ready values: the image pull secrets and the hardened images of the cluster third-party services. See [Gov-Ready Third-Party Services](../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
   * Add `--set global.fipsMode=only` to the Helm command to enable FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. <Badge intent="info" minimal outlined>Gov-ready</Badge>

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-cluster runai/runai-cluster -n runai \
  --set controlPlane.url=... \
  --set controlPlane.bootstrapToken=... \
  --set cluster.uid=... \
  --set cluster.url=... --version="<VERSION>" --create-namespace
```

The wizard displays **Waiting for cluster to connect** while the cluster is being installed and connected to the control plane. Once the installation completes successfully and the cluster establishes communication with the control plane, the wizard updates to **Cluster connected**. After completing the wizard flow, the cluster is added to the [Clusters](../../operations/clusters.md#clusters-table) table.

<Note>

**Tip**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Note>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

When adding a cluster for the first time, the onboarding wizard opens automatically when you log in to the NVIDIA Run:ai platform. You cannot perform other actions in the platform until the cluster is created. For additional clusters, use the Clusters form instead. See [Clusters](../../operations/clusters.md).

1. Enter a unique name for your cluster.
2. Set the cluster location. Choose where the NVIDIA Run:ai cluster will be installed:
   * Same as the control plane - Install the NVIDIA Run:ai cluster on the same Kubernetes cluster as the NVIDIA Run:ai control plane.
   *   Remote control plane - Install the NVIDIA Run:ai cluster on a different Kubernetes cluster than the NVIDIA Run:ai control plane.

       <Note><p><strong>Note</strong></p><p>This selection depends on your <a href="../../system-requirements/deployment-model.md">deployment model</a> - the NVIDIA Run:ai cluster system requirements differ depending on whether the NVIDIA Run:ai cluster is installed on the same Kubernetes cluster as the NVIDIA Run:ai control plane (same-cluster deployment) or on a separate one (multi-cluster deployment).</p></Note>
3. If you selected Remote control plane, enter the Cluster URL. For more information, see [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns) requirement.

**Install NVIDIA Run:ai on Your Cluster**

In the next section, the NVIDIA Run:ai cluster installation steps will be presented.

1. Before installing the NVIDIA Run:ai cluster, see [Preparations](preparations.md) and follow the instructions there.
2.  The NVIDIA Run:ai platform displays the Helm installation command in the cluster wizard.

    **- Do not run the command exactly as shown in the UI -**
3.  Update the UI-generated Helm command as follows (see example command below) and use the `cluster-airgap-values.yaml` file written by the `airgap load` step in [Preparations](preparations.md#software-artifacts), instead of using helm repositories:

    * Do not add the helm repository, `helm repo add`, and do not run `helm repo update`.
    * Instead, edit the `helm upgrade` command:
      * Replace `runai/runai-cluster` with `./charts/runai-cluster-<VERSION>.tgz`, where `<VERSION>` is the full version number. This file is located in the `charts` folder of the working directory you specified to `airgap load`.
      * Add `-f <VALUES_DIR>/cluster-airgap-values.yaml`. Replace `<VALUES_DIR>` with the `valuesDir` path you set in `airgap-load.yaml` during [Preparations](preparations.md#software-artifacts) - this is where `cluster-airgap-values.yaml` was written.
      * If you are using a local certificate authority, add `--set global.customCA.enabled=true` as described in the [Cluster Local CA](preparations.md#cluster-local-ca) section.
      * No Gov-ready values are needed for the cluster third-party services. Only their images change, and the generated values file already sets them. <Badge intent="info" minimal outlined>Gov-ready</Badge>
      * Add `--set global.fipsMode=only` to the Helm command to enable FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. <Badge intent="info" minimal outlined>Gov-ready</Badge>
      * Keep the remaining `--set` values exactly as generated by the UI.

    ```bash
    helm upgrade -i runai-cluster ./charts/runai-cluster-<VERSION>.tgz \
        --set controlPlane.url=... \
        --set controlPlane.bootstrapToken=... \
        --set cluster.uid=... \
        --set cluster.url=... --create-namespace \
        -f <VALUES_DIR>/cluster-airgap-values.yaml
    ```

The wizard displays **Waiting for cluster to connect** while the cluster is being installed and connected to the control plane. Once the installation completes successfully and the cluster establishes communication with the control plane, the wizard updates to **Cluster connected**. After completing the wizard flow, the cluster is added to the [Clusters](../../operations/clusters.md#clusters-table) table.

<Note>

**Tip**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Note>

</Accordion>

## Troubleshooting <a href="#troubleshooting" id="troubleshooting"></a>

If you encounter an issue with the installation, try the troubleshooting scenario below.

### Installation <a href="#installation_1" id="installation_1"></a>

If the NVIDIA Run:ai cluster installation failed, check the installation logs to identify the issue. Run the following script to print the installation logs:

```bash
#!/bin/bash

# Print logs of helm hooks if installation failed
hook_pods=$(kubectl get pod -n runai -l run.ai/hook 2> /dev/null | wc -l)
if [ $hook_pods -ne 0 ]
then
  kubectl logs -n runai -l run.ai/hook
fi

runaiconfig="$(kubectl get runaiconfig runai -n runai 2> /dev/null; echo $?)"
if [[ $runaiconfig -eq 0 ]]
then
  kubectl logs -n runai -l name=runai | grep ERROR
fi
```

### Cluster Status <a href="#cluster-status" id="cluster-status"></a>

If the NVIDIA Run:ai cluster installation completed, but the cluster status did not change its status to **Connected**, refer to the cluster [Troubleshooting scenarios](../../operations/clusters.md#troubleshooting-scenarios) section.

## Next Steps

Once the cluster is installed and connected, the NVIDIA Run:ai UI guides you through optional post-installation configurations. These steps are optional but recommended for a production setup:

* **SSO (Single Sign-On)** - Configure SSO to allow users to log in with your organization's identity provider. See [SSO](../../../platform-management/authentication/sso/saml.md) for setup instructions using SAML or OpenID Connect.
* **Email server** - Configure an SMTP server to enable email notifications for workload events and system alerts. See [Email notifications](../../../settings/general-settings/notifications.md#email-notifications) for configuration details.
* **Create your first research team** - Set up [projects](../../../platform-management/aiinitiatives/organization/projects.md) and [permissions](../../../platform-management/authentication/accessrules.md) to organize your AI practitioners and allocate GPU resources.
