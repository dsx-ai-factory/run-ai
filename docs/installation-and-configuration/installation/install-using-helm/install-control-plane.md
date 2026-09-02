# Install the Control Plane

In this section you will install the NVIDIA Run:ai control plane on your Kubernetes cluster using Helm. The control plane provides the central management layer for NVIDIA Run:ai handling multi-cluster management, resource and access management as well as workload submission and monitoring.

This procedure includes:

* Adding the NVIDIA Run:ai Helm repository from NGC
* Configuring key settings such as domain name, ingress, and administrator credentials
* Deploying the control plane into the `runai-backend` namespace

By completing this process, the NVIDIA Run:ai control plane will be running in your cluster and accessible via the configured domain.

## System and Network Requirements
Before installing the NVIDIA Run:ai control plane, validate that the [system requirements](cp-system-requirements.md) and [network requirements](network-requirements.md) are met. For air-gapped environments, make sure you have the [software artifacts](preparations.md#software-artifacts) prepared.

## Permissions

As part of the installation, you will be required to install the NVIDIA Run:ai control plane [Helm charts](https://helm.sh/). The Helm charts require Kubernetes administrator permissions. You can review the exact objects that are created by the charts using the `--dry-run` on both helm charts.

## Installation

<Info>

**Note**

* To customize the installation based on your environment, see [Advanced control plane configurations](../../post-installation-configuration/control-plane-config.md).
* PostgreSQL and Keycloakx are installed with default usernames and passwords. To change the default credentials, see [Additional third-party configurations](../../post-installation-configuration/control-plane-config.md#additional-third-party-configurations).

</Info>

### Kubernetes

<Accordion title="Connected">

**Connected**

Run the following command and update the values as described below:

* Replace `<NGC_API_KEY>` with your NGC API key.
* Replace `global.domain=<DOMAIN>` with the fully qualified domain name (FQDN) obtained in the [Fully qualified domain name](cp-system-requirements.md#fully-qualified-domain-name-fqdn) section.
* The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class: `--set global.ingress.ingressClass=<ingress class>`
* If you are using a local certificate authority, add `--set global.customCA.enabled=true` to the Helm command as described in the [Local certificate authority](cp-system-requirements.md#local-certificate-authority) section.
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator's email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * Minimum Length: Passwords must be at least 8 characters long.
  * Digits: Must contain at least 1 numeric digit (0-9).
  * Lowercase Characters: Must contain at least 1 lowercase letter (a-z).
  * Uppercase Characters: Must contain at least 1 uppercase letter (A-Z).
  * Special Characters: Must contain at least 1 special character (e.g., !, @, #, $).

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password=<NGC_API_KEY>
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=<DOMAIN> \
  --set global.ingress.ingressClass=haproxy \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>"
```

For example:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password='example_password_l0v_1e@tghlfkjfncj_dfjkljf'
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=runai.mycorp.local \
  --set global.ingress.ingressClass=haproxy \
  --set tenantsManager.config.adminUsername=email@local.com \
  --set tenantsManager.config.adminPassword="myPassw0rd\!"
```

<Info>

**Note**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Info>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

Run the following command and update the values as described below. The `custom-env.yaml` file is created during the [preparations](preparations.md) step:

<Warning>

**Prerequisite**

If your internal registry requires authentication, you must create the `runai-reg-creds` imagePullSecret before proceeding. See [Private Docker Registry](preparations.md#private-docker-registry) in Preparations.

</Warning>

* Replace `control-plane-<VERSION>.tgz` with the full filename of the control plane Helm chart (e.g., `control-plane-2.24.58.tgz`), located in the `chart` folder of the extracted software artifacts.
* Replace `global.domain=<DOMAIN>` with the fully qualified domain name (FQDN) obtained in the [Fully qualified domain name](cp-system-requirements.md#fully-qualified-domain-name-fqdn) section.
* Set `global.customCA.enabled=true` as described in the [Local certificate authority](cp-system-requirements.md#local-certificate-authority) section.
* The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class: `--set global.ingress.ingressClass=<ingress class>`
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator's email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * Minimum Length: Passwords must be at least 8 characters long.
  * Digits: Must contain at least 1 numeric digit (0-9).
  * Lowercase Characters: Must contain at least 1 lowercase letter (a-z).
  * Uppercase Characters: Must contain at least 1 uppercase letter (A-Z).
  * Special Characters: Must contain at least 1 special character (e.g., !, @, #, $).

Run the following command from the root of the extracted software artifacts directory:

```bash
helm upgrade -i runai-backend ./chart/control-plane-<VERSION>.tgz \
    --set global.domain=<DOMAIN> \
    --set global.customCA.enabled=true \
    --set global.ingress.ingressClass=haproxy \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    -n runai-backend -f custom-env.yaml
```

For example:

```shellscript
helm upgrade -i runai-backend ./chart/control-plane-2.25.10.tgz \
    --set global.domain=runai.mycorp.local \
    --set global.customCA.enabled=true \
    --set global.ingress.ingressClass=haproxy \
    --set tenantsManager.config.adminUsername=email@local.com \
    --set tenantsManager.config.adminPassword="myPassw0rd\!" \
    -n runai-backend -f custom-env.yaml
```

<Info>

**Note**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Info>

</Accordion>

### OpenShift

<Accordion title="Connected">

**Connected**

Run the following command and update the values as described below:

* Replace `<NGC_API_KEY>` with your NGC API key.
* Replace the `<OPENSHIFT-CLUSTER-DOMAIN>` with the domain configured for the OpenShift cluster. To determine the OpenShift cluster domain, run `oc get routes -A`.
* If you are using a local certificate authority, add `--set global.customCA.enabled=true` to the Helm command as described in the [Local certificate authority](cp-system-requirements.md#local-certificate-authority) section.
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator's email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * Minimum Length: Passwords must be at least 8 characters long.
  * Digits: Must contain at least 1 numeric digit (0-9).
  * Lowercase Characters: Must contain at least 1 lowercase letter (a-z).
  * Uppercase Characters: Must contain at least 1 uppercase letter (A-Z).
  * Special Characters: Must contain at least 1 special character (e.g., !, @, #, $).

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password=<NGC_API_KEY>
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
  --set global.config.kubernetesDistribution=openshift \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>"
```

For example:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password='example_password_l0v_1e@tghlfkjfncj_dfjkljf'
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=runai.apps.mycorp.local \
  --set global.config.kubernetesDistribution=openshift \
  --set tenantsManager.config.adminUsername=email@local.com \
  --set tenantsManager.config.adminPassword="myPassw0rd\!"
```

<Info>

**Note**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Info>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

Run the following command and update the values as described below. The `custom-env.yaml` file is created during the [preparations](preparations.md) step:

<Warning>

**Prerequisite**

If your internal registry requires authentication, you must create the `runai-reg-creds` imagePullSecret before proceeding. See [Private Docker Registry](preparations.md#private-docker-registry) in Preparations.

</Warning>

* Replace `control-plane-<VERSION>.tgz` with the full filename of the control plane Helm chart (e.g., `control-plane-2.24.49.tgz`), located in the `chart` folder of the extracted software artifacts.
* Replace `<OPENSHIFT-CLUSTER-DOMAIN>` with the domain configured for the OpenShift cluster. To determine the OpenShift cluster domain, run `oc get routes -A`.
* Set `global.customCA.enabled=true` as described in the [Local certificate authority](cp-system-requirements.md#local-certificate-authority) section.
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator's email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * Minimum Length: Passwords must be at least 8 characters long.
  * Digits: Must contain at least 1 numeric digit (0-9).
  * Lowercase Characters: Must contain at least 1 lowercase letter (a-z).
  * Uppercase Characters: Must contain at least 1 uppercase letter (A-Z).
  * Special Characters: Must contain at least 1 special character (e.g., !, @, #, $).

Run the following command from the root of the extracted software artifacts directory:

```bash
helm upgrade -i runai-backend ./chart/control-plane-<VERSION>.tgz -n runai-backend \
    --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
    --set global.config.kubernetesDistribution=openshift \
    --set global.customCA.enabled=true \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    -f custom-env.yaml
```

For example:

```bash
helm upgrade -i runai-backend ./chart/control-plane-2.24.49.tgz -n runai-backend \
    --set global.domain=runai.apps.mycorp.local \
    --set global.config.kubernetesDistribution=openshift \
    --set global.customCA.enabled=true \
    --set tenantsManager.config.adminUsername=email@local.com \
    --set tenantsManager.config.adminPassword="myPassw0rd\!" \
    -f custom-env.yaml
```

<Info>

**Note**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Info>

</Accordion>

## Connect to NVIDIA Run:ai User Interface

<Info>

**Note**

After installing the NVIDIA Run:ai control plane, it may take a few minutes for the UI to become accessible (up to 15 minutes). Accessing it too early may result in a "server cannot be reached" error.

</Info>

1. Open your browser and go to:

<Tabs>

<Tab title="Kubernetes">

`https://runai.<DOMAIN>.local`

</Tab>

<Tab title="OpenShift">

`https://runai.apps.<OpenShift-DOMAIN>`

</Tab>

</Tabs>

2. Log in using the administrator credentials provided during the installation. It is recommended to change the password after the first login.
