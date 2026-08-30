# Install Using Base Command Manager (BCM)

This section explains the steps required to install the NVIDIA Run:ai cluster on a DGX Kubernetes Cluster using NVIDIA [Base Command Manager (BCM)](https://docs.nvidia.com/base-command-manager/index.html).

## NVIDIA Run:ai Installer

The NVIDIA Run:ai installer is a wizard that simplifies the deployment of the NVIDIA Run:ai cluster on DGX. The NVIDIA Run:ai installer is installed via the BCM cluster wizard when the cluster is created.

<Callout intent="info">
For custom deployment options, check the [Install using Helm](../installation-and-configuration/installation/install-using-helm/helm-install.md).
</Callout>

## System and Network Requirements

Before installing the NVIDIA Run:ai cluster on a DGX system using BCM, ensure that your [system requirements](../installation-and-configuration/installation/install-using-helm/system-requirements.md) and [network requirements](../installation-and-configuration/installation/install-using-helm/network-requirements.md) meets the necessary prerequisites.

The BCM cluster wizard deploys essential [software requirements](../installation-and-configuration/installation/install-using-helm/system-requirements.md#software-requirements), such as the [Kubernetes Ingress Controller](../installation-and-configuration/installation/install-using-helm/system-requirements.md#kubernetes-ingress-controller), [NVIDIA GPU Operator](../installation-and-configuration/installation/install-using-helm/system-requirements.md#nvidia-gpu-operator), and [Prometheus](../installation-and-configuration/installation/install-using-helm/system-requirements.md#prometheus), as part of the NVIDIA Run:ai Installer deployment. Additional optional software requirements for [Distributed training](../installation-and-configuration/installation/install-using-helm/system-requirements.md#distributed-training) and [Inference](../installation-and-configuration/installation/install-using-helm/system-requirements.md#inference) requires manual setup.

## Tenant Name

Your tenant name is predefined and supplied by NVIDIA Run:ai. Each customer is provided with a unique, dedicated URL in the format `<tenant-name>.run.ai` which includes the required tenant name.

## Application Secret Key

An application secret key is required to connect the cluster to the NVIDIA Run:ai platform. In order to get the application secret key, a new cluster must be added.

1. Follow the [Adding a new cluster](../installation-and-configuration/installation/install-using-helm/helm-install.md) setup instructions. **Do not follow the Installation instructions**.
2. Once cluster instructions are displayed, find the `controlPlane.clientSecret` flag in the displayed Helm command, copy and save its value.

## TLS Certificate

A TLS private and public keys for the cluster’s [Fully Qualified Domain Name (FQDN)](../installation-and-configuration/installation/install-using-helm/system-requirements.md#fully-qualified-domain-name-fqdn) are required for HTTP access to the cluster

<Callout intent="info">
TLS Certificate must be trusted. Self-signed certificates are not supported.
</Callout>

## Installation

Follow these instructions to install using BCM.

### Installing a Cluster

The cluster installer is available via the locally installed BCM landing page,

1. Go to the locally installed BCM landing page, Select the NVIDIA Run:ai tile or access directly to `http://<BCM-CLUSTER-IP>:30080/runai-installer` (HTTP only)
2. Click **VERIFY** in order to check System Requirements are met.
3. After verification completed successfully, click **CONTINUE**.
4. Enter the cluster information and click **CONTINUE**.
5. The NVIDIA Run:ai installation will start and should be complete within a few minutes
6. Once a message of **NVIDIA Run:ai was installed successfully!** is displayed, Click on **START USING NVIDIA Run:ai** to launch the login page of the tenant in a new browser tab.

## Troubleshooting

If you encounter an issue with the installation, try the troubleshooting scenario below.

### NVIDIA Run:ai Installer

The NVIDIA Run:ai installer is a pod in Kubernetes. The pod is responsible for the installation preparation and prerequisite gathering phase. If there is an error during the prerequisite verification process, run the following command to print the logs:

```bash
kubectl get pods -n runai | grep 'cluster-installer' #Find the cluster installer pod's name
kubectl logs <POD-NAME> -n runai #Print the cluster installer pod logs
```

### Installation

If the NVIDIA Run:ai cluster installation failed, check the installation logs to identify the issue. Run the following script to print the installation logs:

Contact NVIDIA Run:ai Support to obtain the `get-installation-logs.sh` script.

### Cluster Status

If the NVIDIA Run:ai cluster installation is complete but the cluster status did not change to **Connected**, check the cluster [troubleshooting scenarios](../infrastructure-setup/procedures/clusters.md#troubleshooting-scenarios)
