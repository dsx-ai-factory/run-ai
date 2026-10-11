# Gov-Ready Installation

A Gov-ready installation follows the standard NVIDIA Run:ai installation procedures. You can install using [Guided Installation](guided-installation/README.md) (recommended) or [Manual Helm](install-using-helm/preparations.md), on Kubernetes or OpenShift, in connected or air-gapped environments. The standard system requirements, control plane and cluster installation flow, and post-installation steps all apply, with the Gov-ready additions described on this page.

Gov-ready adds the following requirements:

- **Third-party services** - Use FIPS-validated hardened images and an external PostgreSQL instance.
- **GPU Operator and Network Operator** - Install the Gov-ready versions of both operators according to their respective documentation, [GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/install-gpu-operator-gov-ready.html) and [Network Operator](https://docs.nvidia.com/networking/display/kubernetes2670/install-network-operator-gov-ready.html).
- **FIPS mode** - Enable FIPS mode when installing the control plane and the cluster.

Each installation procedure identifies where Gov-ready-specific configuration is required, so follow the procedure for your installation path and apply the Gov-ready steps where indicated. For an overview of Gov-ready and what it provides, refer to [Gov-Ready](../overview.md#gov-ready).

## System Requirements

Review the [system requirements](../system-requirements/README.md) for your deployment model, infrastructure, and networking setup before you begin.

## Third-Party Services

Gov-ready uses an external PostgreSQL instance and FIPS-validated hardened images for the third-party services. The hardened images are Docker Hardened Images (DHI), which you pull from Docker Hub (`docker.io`) from repositories in your Docker Hub organization, using your Docker Hub credentials. For the image registry settings, registry credentials, and Helm values each service requires, refer to [Gov-Ready Third-Party Services](../system-requirements/gov-ready-third-party-services.md).

Pass the Gov-ready values to the installation as custom Helm values:

- **Guided Installation** - Add the values under `controlPlane.customHelmValues` and `cluster.customHelmValues` in the installer configuration.
- **Manual Helm** - Save the values in a values file and pass the file with `-f` when installing the control plane and cluster. In air-gapped environments, pass it in addition to the values files that `airgap load` generates.

The hardened images are provided differently in connected and air-gapped environments:

- **Connected** - Create the registry secret for the hardened images before installation, and set the image values.
- **Air-gapped** - Add the hardened images and your Docker Hub credentials to the pack configuration when preparing the air-gapped bundle. You do not create a Kubernetes secret or set the image values. The load step generates them. Set only the other values, such as the external PostgreSQL connection.

## Prerequisites and Features

Install the prerequisites and features as described in the installation procedure, except for the GPU Operator and Network Operator. Both provide Gov-ready versions, so install each according to its Gov-ready documentation, including any required preparation and Helm values. Refer to the Gov-ready documentation of the [NVIDIA GPU Operator](https://docs.nvidia.com/datacenter/cloud-native/gpu-operator/latest/install-gpu-operator-gov-ready.html) and the [NVIDIA Network Operator](https://docs.nvidia.com/networking/display/kubernetes2670/install-network-operator-gov-ready.html).

- **Guided Installation** - Pass the required operator values under the operator's `customHelmValues`. Refer to [Prerequisites and Features](guided-installation/prerequisites-and-features.md). In air-gapped environments, install the operators before running the installer.
- **Manual Helm** - Install the operators with the required values. Refer to [Prerequisites and Features](install-using-helm/prerequisites-and-features.md).

## Install the Control Plane and the Cluster

In addition to the third-party service values, enable FIPS mode when installing both the control plane and the cluster by setting `global.fipsMode: only` in the control plane and cluster configuration.

- **Guided Installation** - Add `global.fipsMode: only` under both `controlPlane.customHelmValues` and `cluster.customHelmValues` in the installer configuration.
- **Manual Helm** - Add `--set global.fipsMode=only` to the `helm upgrade` commands for the control plane and the cluster.

## Next Steps

Continue with the installation procedure for your selected installation method:

- [Guided Installation](guided-installation/README.md) - Recommended for most installations.
- [Manual Helm](install-using-helm/preparations.md) - For organizations that manage releases through their own tooling.
