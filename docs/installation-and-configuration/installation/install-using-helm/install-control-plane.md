# Install the Control Plane

This page describes how to install the NVIDIA Run:ai control plane on your Kubernetes cluster using Helm. The control plane provides the central management layer for NVIDIA Run:ai handling multi-cluster management, resource and access management as well as workload submission and monitoring.

This procedure includes:

* Configuring key settings such as domain name, ingress, and administrator credentials
* Deploying the control plane into the `runai-backend` namespace

By completing this process, the NVIDIA Run:ai control plane will be running in your cluster and accessible via the configured domain.

## Before You Start

Before installing the NVIDIA Run:ai control plane, complete [Preparations](preparations.md).

## Permissions

As part of the installation, you will be required to install the NVIDIA Run:ai control plane [Helm charts](https://helm.sh/). The Helm charts require Kubernetes administrator permissions. You can review the exact objects that are created by the charts using the `--dry-run` on both helm charts.

## Installation

<Note>

**Note**

* To customize the installation based on your environment, see [Advanced control plane configurations](../../configuration/control-plane-config.md).
* PostgreSQL and Keycloakx are installed with default usernames and passwords. To change the default credentials, see [Changing PostgreSQL Passwords](../../configuration/control-plane-config.md#changing-postgresql-passwords) and [Changing the Keycloak Admin Password](../../configuration/control-plane-config.md#changing-the-keycloak-admin-password) in Advanced Control Plane Configurations.

</Note>

### Kubernetes

<Accordion title="Connected">

**Connected**

Use the Helm command for the release type you are installing and update the values as follows.

* Replace `global.domain=<DOMAIN>` with the fully qualified domain name (FQDN) obtained in [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns).
* The recommended ingress controller is HAProxy. If you are using a different ingress controller, set `--set global.ingress.ingressClass=<INGRESS_CLASS>` to the ingress class configured for your environment.
* If you are using a local certificate authority, add `--set global.customCA.enabled=true` as described in [Control Plane Local CA](preparations.md#control-plane-local-ca).
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * **Minimum Length**: At least 8 characters long.
  * **Digits**: At least 1 numeric digit (0-9).
  * **Lowercase Characters**: At least 1 lowercase letter (a-z).
  * **Uppercase Characters**: At least 1 uppercase letter (A-Z).
  * **Special Characters**: At least 1 special character (e.g., !, @, #, $).
* Save the image pull secrets, the external PostgreSQL connection, and the hardened images and settings of the control plane third-party services, in a values file and add `-f <GOV_READY_VALUES_FILE>` to the `helm upgrade` command. See [Gov-Ready Third-Party Services](../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
* Add `--set global.fipsMode=only` to the `helm upgrade` command to enable FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. <Badge intent="info" minimal outlined>Gov-ready</Badge>

**Open Source**

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=<DOMAIN> \
  --set global.ingress.ingressClass=haproxy \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>"
```

**Support-Entitled**

In addition to the values above, update the Helm commands as follows:

* Add `--set supportEntitled=true` to the `helm upgrade` command.
* Add `--set telemetryService.enabled=true` to the `helm upgrade` command to enable telemetry.
* Add `--set global.imagePullSecrets[0].name=runai-reg-creds` to the `helm upgrade` command to reference the secret created in [Software Artifacts](preparations.md#software-artifacts).
* Optional: associate NVIDIA Run:ai's telemetry with your NCA ID by adding one of the following to the `helm upgrade` command. See [Registry Credentials](preparations.md#registry-credentials) for how to generate the NGC Service API Key. By default, telemetry is not associated with your organization.
  - `--set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>` - Passes the Service API Key directly. Visible in shell history and process listings.
  - `--set telemetryService.telemetryAuth.existingSecret=<SECRET_NAME> --set telemetryService.telemetryAuth.secretKey=<KEY_NAME>` - References a secret you create yourself beforehand - the chart does not create or manage it.

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=<DOMAIN> \
  --set global.ingress.ingressClass=haproxy \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
  --set supportEntitled=true \
  --set telemetryService.enabled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds \
  --set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>
```

For example:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=runai.mycorp.local \
  --set global.ingress.ingressClass=haproxy \
  --set tenantsManager.config.adminUsername=email@local.com \
  --set tenantsManager.config.adminPassword="myPassw0rd\!" \
  --set supportEntitled=true \
  --set telemetryService.enabled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds \
  --set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>
```

<Note>

**Note**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Note>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

<Warning>

**Prerequisite**

If your internal registry requires authentication, you must create the `runai-reg-creds` imagePullSecret before proceeding. See [Software Artifacts](preparations.md#software-artifacts) in Preparations.

</Warning>

Use the Helm command for the release type you are installing and update the values as follows. The `control-plane-airgap-values.yaml` file is written by the `airgap load` step in [Preparations](preparations.md#software-artifacts).

* Replace `control-plane-<VERSION>.tgz` with the full filename of the control plane Helm chart (e.g., `control-plane-3.0.2.tgz`), located in the `charts` folder of the working directory you specified to `airgap load`.
* Replace `<VALUES_DIR>` with the `valuesDir` path you set in `airgap-load.yaml` during [Preparations](preparations.md#software-artifacts) - this is where `control-plane-airgap-values.yaml` was written.
* Replace `global.domain=<DOMAIN>` with the fully qualified domain name (FQDN) obtained in [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns).
* Set `global.customCA.enabled=true` as described in [Control Plane Local CA](preparations.md#control-plane-local-ca).
* The recommended ingress controller is HAProxy. If you are using a different ingress controller, set `--set global.ingress.ingressClass=<INGRESS_CLASS>` to the ingress class configured for your environment.
* If your internal registry requires authentication, add `--set global.imagePullSecrets[0].name=runai-reg-creds` to reference the secret created in [Software Artifacts](preparations.md#software-artifacts).
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * **Minimum Length**: At least 8 characters long.
  * **Digits**: At least 1 numeric digit (0-9).
  * **Lowercase Characters**: At least 1 lowercase letter (a-z).
  * **Uppercase Characters**: At least 1 uppercase letter (A-Z).
  * **Special Characters**: At least 1 special character (e.g., !, @, #, $).
* Save the external PostgreSQL connection and the Keycloak and NATS settings of the control plane third-party services in a values file, and add `-f <GOV_READY_VALUES_FILE>` to the `helm upgrade` command, next to the `-f` of the generated values file. Do not set image values, because `airgap load` generates them. See [Gov-Ready Third-Party Services](../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
* Add `--set global.fipsMode=only` to the `helm upgrade` command to enable FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. <Badge intent="info" minimal outlined>Gov-ready</Badge>

**Open Source**

```bash
helm upgrade -i runai-backend ./charts/control-plane-<VERSION>.tgz \
    --set global.domain=<DOMAIN> \
    --set global.customCA.enabled=true \
    --set global.ingress.ingressClass=haproxy \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    -n runai-backend -f <VALUES_DIR>/control-plane-airgap-values.yaml
```

**Support-Entitled**

In addition to the values above, add `--set supportEntitled=true` and `--set telemetryService.enabled=true` to the `helm upgrade` command:

```bash
helm upgrade -i runai-backend ./charts/control-plane-<VERSION>.tgz \
    --set global.domain=<DOMAIN> \
    --set global.customCA.enabled=true \
    --set global.ingress.ingressClass=haproxy \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    --set supportEntitled=true \
    --set telemetryService.enabled=true \
    -n runai-backend -f <VALUES_DIR>/control-plane-airgap-values.yaml
```

For example:

```bash
helm upgrade -i runai-backend ./charts/control-plane-3.0.2.tgz \
    --set global.domain=runai.mycorp.local \
    --set global.customCA.enabled=true \
    --set global.ingress.ingressClass=haproxy \
    --set tenantsManager.config.adminUsername=email@local.com \
    --set tenantsManager.config.adminPassword="myPassw0rd\!" \
    --set supportEntitled=true \
    --set telemetryService.enabled=true \
    -n runai-backend -f <VALUES_DIR>/control-plane-airgap-values.yaml
```

<Note>

**Note**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Note>

</Accordion>

### OpenShift

<Accordion title="Connected">

**Connected**

Use the Helm command for the release type you are installing and update the values as follows.

* Replace the `<OPENSHIFT-CLUSTER-DOMAIN>` with the domain configured for the OpenShift cluster. To determine the OpenShift cluster domain, run `oc get routes -A`.
* If you are using a local certificate authority, add `--set global.customCA.enabled=true` as described in [Control Plane Local CA](preparations.md#control-plane-local-ca).
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * **Minimum Length**: At least 8 characters long.
  * **Digits**: At least 1 numeric digit (0-9).
  * **Lowercase Characters**: At least 1 lowercase letter (a-z).
  * **Uppercase Characters**: At least 1 uppercase letter (A-Z).
  * **Special Characters**: At least 1 special character (e.g., !, @, #, $).
* Save the image pull secrets, the external PostgreSQL connection, and the hardened images and settings of the control plane third-party services, in a values file and add `-f <GOV_READY_VALUES_FILE>` to the `helm upgrade` command. See [Gov-Ready Third-Party Services](../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
* Add `--set global.fipsMode=only` to the `helm upgrade` command to enable FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. <Badge intent="info" minimal outlined>Gov-ready</Badge>

**Open Source**

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
  --set global.config.kubernetesDistribution=openshift \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>"
```

**Support-Entitled**

In addition to the values above, update the Helm commands as follows:

* Add `--set supportEntitled=true` to the `helm upgrade` command.
* Add `--set telemetryService.enabled=true` to the `helm upgrade` command to enable telemetry.
* Add `--set global.imagePullSecrets[0].name=runai-reg-creds` to the `helm upgrade` command to reference the secret created in [Software Artifacts](preparations.md#software-artifacts).
* Optional: associate NVIDIA Run:ai's telemetry with your NCA ID by adding one of the following to the `helm upgrade` command. See [Registry Credentials](preparations.md#registry-credentials) for how to generate the NGC Service API Key. By default, telemetry is not associated with your organization.
  - `--set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>` - Passes the Service API Key directly. Visible in shell history and process listings.
  - `--set telemetryService.telemetryAuth.existingSecret=<SECRET_NAME> --set telemetryService.telemetryAuth.secretKey=<KEY_NAME>` - References a secret you create yourself beforehand - the chart does not create or manage it.

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
  --set global.config.kubernetesDistribution=openshift \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
  --set supportEntitled=true \
  --set telemetryService.enabled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds \
  --set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>
```

For example:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=runai.apps.mycorp.local \
  --set global.config.kubernetesDistribution=openshift \
  --set tenantsManager.config.adminUsername=email@local.com \
  --set tenantsManager.config.adminPassword="myPassw0rd\!" \
  --set supportEntitled=true \
  --set telemetryService.enabled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds \
  --set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>
```

<Note>

**Note**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Note>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

<Warning>

**Prerequisite**

If your internal registry requires authentication, you must create the `runai-reg-creds` imagePullSecret before proceeding. See [Software Artifacts](preparations.md#software-artifacts) in Preparations.

</Warning>

Use the Helm command for the release type you are installing and update the values as follows. The `control-plane-airgap-values.yaml` file is written by the `airgap load` step in [Preparations](preparations.md#software-artifacts).

* Replace `control-plane-<VERSION>.tgz` with the full filename of the control plane Helm chart (e.g., `control-plane-3.0.2.tgz`), located in the `charts` folder of the working directory you specified to `airgap load`.
* Replace `<VALUES_DIR>` with the `valuesDir` path you set in `airgap-load.yaml` during [Preparations](preparations.md#software-artifacts) - this is where `control-plane-airgap-values.yaml` was written.
* Replace `<OPENSHIFT-CLUSTER-DOMAIN>` with the domain configured for the OpenShift cluster. To determine the OpenShift cluster domain, run `oc get routes -A`.
* Set `global.customCA.enabled=true` as described in [Control Plane Local CA](preparations.md#control-plane-local-ca).
* If your internal registry requires authentication, add `--set global.imagePullSecrets[0].name=runai-reg-creds` to reference the secret created in [Software Artifacts](preparations.md#software-artifacts).
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * **Minimum Length**: At least 8 characters long.
  * **Digits**: At least 1 numeric digit (0-9).
  * **Lowercase Characters**: At least 1 lowercase letter (a-z).
  * **Uppercase Characters**: At least 1 uppercase letter (A-Z).
  * **Special Characters**: At least 1 special character (e.g., !, @, #, $).
* Save the external PostgreSQL connection and the Keycloak and NATS settings of the control plane third-party services in a values file, and add `-f <GOV_READY_VALUES_FILE>` to the `helm upgrade` command, next to the `-f` of the generated values file. Do not set image values, because `airgap load` generates them. See [Gov-Ready Third-Party Services](../../system-requirements/gov-ready-third-party-services.md) for the values. <Badge intent="info" minimal outlined>Gov-ready</Badge>
* Add `--set global.fipsMode=only` to the `helm upgrade` command to enable FIPS mode, so that the NVIDIA Run:ai services use only FIPS-validated cryptography. <Badge intent="info" minimal outlined>Gov-ready</Badge>

**Open Source**

```bash
helm upgrade -i runai-backend ./charts/control-plane-<VERSION>.tgz -n runai-backend \
    --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
    --set global.config.kubernetesDistribution=openshift \
    --set global.customCA.enabled=true \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    -f <VALUES_DIR>/control-plane-airgap-values.yaml
```

**Support-Entitled**

In addition to the values above, add `--set supportEntitled=true` and `--set telemetryService.enabled=true` to the `helm upgrade` command:

```bash
helm upgrade -i runai-backend ./charts/control-plane-<VERSION>.tgz -n runai-backend \
    --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
    --set global.config.kubernetesDistribution=openshift \
    --set global.customCA.enabled=true \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    --set supportEntitled=true \
    --set telemetryService.enabled=true \
    -f <VALUES_DIR>/control-plane-airgap-values.yaml
```

For example:

```bash
helm upgrade -i runai-backend ./charts/control-plane-3.0.2.tgz -n runai-backend \
    --set global.domain=runai.apps.mycorp.local \
    --set global.config.kubernetesDistribution=openshift \
    --set global.customCA.enabled=true \
    --set tenantsManager.config.adminUsername=email@local.com \
    --set tenantsManager.config.adminPassword="myPassw0rd\!" \
    --set supportEntitled=true \
    --set telemetryService.enabled=true \
    -f <VALUES_DIR>/control-plane-airgap-values.yaml
```

<Note>

**Note**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Note>

</Accordion>

## Connect to NVIDIA Run:ai User Interface

<Note>

**Note**

After installing the NVIDIA Run:ai control plane, it may take a few minutes for the UI to become accessible (up to 15 minutes). Accessing it too early may result in a "server cannot be reached" error.

</Note>

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

## Next Steps

With the control plane installed, continue to [Install the Cluster](helm-install.md). Before installing the cluster, make sure you have completed the required setup in [Prerequisites and Features](prerequisites-and-features.md).
