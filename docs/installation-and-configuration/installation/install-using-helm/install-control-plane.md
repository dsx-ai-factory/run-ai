# Install the Control Plane

In this section you will install the NVIDIA Run:ai control plane on your Kubernetes cluster using Helm. The control plane provides the central management layer for NVIDIA Run:ai handling multi-cluster management, resource and access management as well as workload submission and monitoring.

This procedure includes:

* Adding the NVIDIA Run:ai Helm repository from NGC - without credentials for Open Source, or with an NGC Personal Key for Support-Entitled Release installations
* Configuring key settings such as domain name, ingress, and administrator credentials
* Deploying the control plane into the `runai-backend` namespace

By completing this process, the NVIDIA Run:ai control plane will be running in your cluster and accessible via the configured domain.

## Before You Start

Before installing the NVIDIA Run:ai control plane, complete the entire [System Requirements](../../system-requirements/README.md) section and [Preparations](preparations.md).

## Permissions

As part of the installation, you will be required to install the NVIDIA Run:ai control plane [Helm charts](https://helm.sh/). The Helm charts require Kubernetes administrator permissions. You can review the exact objects that are created by the charts using the `--dry-run` on both helm charts.

## Installation

<Note>

**Note**

* To customize the installation based on your environment, see [Advanced control plane configurations](../../configuration/control-plane-config.md).
* PostgreSQL and Keycloakx are installed with default usernames and passwords. To change the default credentials, see [Third-Party Services](../../system-requirements/third-party-services.md).

</Note>

### Open Source and Support-Entitled Release

The installation steps below cover both Open Source and Support-Entitled Release, using the same control plane Helm chart. Support-Entitled Release installations authenticate to the Helm registry (`helm.ngc.nvidia.com/nvidia/runai`) with an [NGC Personal Key](preparations.md#registry-credentials) and add `--set supportEntitled=true` to the `helm upgrade` command.

Setting `--set supportEntitled=true` also enables telemetry - usage metrics the control plane periodically sends to NVIDIA, including cluster utilization, allocation, adoption, and workload submission types. Sharing an NGC [Service Key](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#generating-a-service-api-key) lets NVIDIA associate this data with your account, so support can help you faster; without one, data is still sent anonymously.

Setting `--set supportEntitled=true` also switches the UI branding and support contact link shown in the control plane to the NVIDIA enterprise variant. For air-gapped Support-Entitled installations, this flag is still required, but authentication uses your internal registry credentials instead of an NGC Personal Key.

<Warning>

**TODO**

Confirm and document the actual Helm value/flag for supplying the Service Key. Per Jira (RUN-35049 epic; RUN-39433 HL design: "provide the nca id or a token in the values file for authentication, should be stored as a secret"; RUN-41574: "call-home chart - pass customer-provided SAK Secret to diagnostics service TelemetryToken" - Done), this is implemented, but the field lives in the external `runai-telemetry` chart (see `telemetryService` dependency in `control-plane-chart/Chart.yaml`, pulled from `runai.jfrog.io/artifactory/api/helm/diagnostics`), which isn't in the `runai` monorepo clone. The LL design doc is on `nvidia.atlassian.net` (RUN-41063), a different Atlassian site than our Jira access covers.

</Warning>

| Type | Registry Authentication | `supportEntitled` | Telemetry |
| ---- | ------------------------ | ------------------ | --------- |
| Open Source | Not required | Unset (`false`) | Not sent |
| Support-Entitled Release | NGC Personal Key | `true` | Sent by default; opt out with `--set telemetryService.enabled=false` |

### Kubernetes

<Accordion title="Connected">

**Connected**

The same `helm upgrade` command installs the control plane for both Open Source and Support-Entitled Release. Update the values as described below; Support-Entitled Release installations additionally authenticate `helm repo add` with an [NGC Personal Key](preparations.md#registry-credentials) (`helm.ngc.nvidia.com/nvidia/runai`) and add `--set supportEntitled=true` to the `helm upgrade` command - see the examples below.

* Replace `global.domain=<DOMAIN>` with the fully qualified domain name (FQDN) obtained in the [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns) section.
* The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class: `--set global.ingress.ingressClass=<ingress class>`
* If you are using a local certificate authority, add `--set global.customCA.enabled=true` to the Helm command as described in the [Control Plane Local CA](preparations.md#control-plane-local-ca) section.
* Support-Entitled Release installations must reference the `runai-reg-creds` secret created in [Software Artifacts](preparations.md#software-artifacts): add `--set global.imagePullSecrets[0].name=runai-reg-creds`.
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator's email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * Minimum Length: Passwords must be at least 8 characters long.
  * Digits: Must contain at least 1 numeric digit (0-9).
  * Lowercase Characters: Must contain at least 1 lowercase letter (a-z).
  * Uppercase Characters: Must contain at least 1 uppercase letter (A-Z).
  * Special Characters: Must contain at least 1 special character (e.g., !, @, #, $).

Open Source:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=<DOMAIN> \
  --set global.ingress.ingressClass=haproxy \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>"
```

Support-Entitled, using your NGC Personal Key:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password=<NGC_PERSONAL_KEY>
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=<DOMAIN> \
  --set global.ingress.ingressClass=haproxy \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
  --set supportEntitled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds
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
  --set tenantsManager.config.adminPassword="myPassw0rd\!" \
  --set supportEntitled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds
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

The same `helm upgrade` command installs the control plane for both Open Source and Support-Entitled Release. Update the values as described below; Support-Entitled Release installations additionally add `--set supportEntitled=true` to the `helm upgrade` command. The `control-plane-airgap-values.yaml` file is written by the `airgap load` step in [Preparations](preparations.md#software-artifacts):

* Replace `control-plane-<VERSION>.tgz` with the full filename of the control plane Helm chart (e.g., `control-plane-3.0.2.tgz`), located in the `charts` folder of the working directory you specified to `airgap load`.
* Replace `<VALUES_DIR>` with the `valuesDir` path you set in `airgap-load.yaml` during [Preparations](preparations.md#software-artifacts) - this is where `control-plane-airgap-values.yaml` was written.

* Replace `global.domain=<DOMAIN>` with the fully qualified domain name (FQDN) obtained in the [FQDN and DNS](../../system-requirements/networking-setup.md#fqdn-and-dns) section.
* Set `global.customCA.enabled=true` as described in the [Control Plane Local CA](preparations.md#control-plane-local-ca) section.
* The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class: `--set global.ingress.ingressClass=<ingress class>`
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator's email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * Minimum Length: Passwords must be at least 8 characters long.
  * Digits: Must contain at least 1 numeric digit (0-9).
  * Lowercase Characters: Must contain at least 1 lowercase letter (a-z).
  * Uppercase Characters: Must contain at least 1 uppercase letter (A-Z).
  * Special Characters: Must contain at least 1 special character (e.g., !, @, #, $).

Open Source:

```bash
helm upgrade -i runai-backend ./charts/control-plane-<VERSION>.tgz \
    --set global.domain=<DOMAIN> \
    --set global.customCA.enabled=true \
    --set global.ingress.ingressClass=haproxy \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    -n runai-backend -f <VALUES_DIR>/control-plane-airgap-values.yaml
```

Support-Entitled Release:

```bash
helm upgrade -i runai-backend ./charts/control-plane-<VERSION>.tgz \
    --set global.domain=<DOMAIN> \
    --set global.customCA.enabled=true \
    --set global.ingress.ingressClass=haproxy \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    --set supportEntitled=true \
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

The same `helm upgrade` command installs the control plane for both Open Source and Support-Entitled Release. Update the values as described below; Support-Entitled Release installations additionally authenticate `helm repo add` with an [NGC Personal Key](preparations.md#registry-credentials) (`helm.ngc.nvidia.com/nvidia/runai`) and add `--set supportEntitled=true` to the `helm upgrade` command - see the examples below.

* Replace the `<OPENSHIFT-CLUSTER-DOMAIN>` with the domain configured for the OpenShift cluster. To determine the OpenShift cluster domain, run `oc get routes -A`.
* If you are using a local certificate authority, add `--set global.customCA.enabled=true` to the Helm command as described in the [Control Plane Local CA](preparations.md#control-plane-local-ca) section.
* Support-Entitled Release installations must reference the `runai-reg-creds` secret created in [Software Artifacts](preparations.md#software-artifacts): add `--set global.imagePullSecrets[0].name=runai-reg-creds`.
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator's email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * Minimum Length: Passwords must be at least 8 characters long.
  * Digits: Must contain at least 1 numeric digit (0-9).
  * Lowercase Characters: Must contain at least 1 lowercase letter (a-z).
  * Uppercase Characters: Must contain at least 1 uppercase letter (A-Z).
  * Special Characters: Must contain at least 1 special character (e.g., !, @, #, $).

Open Source:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
  --set global.config.kubernetesDistribution=openshift \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>"
```

Support-Entitled, using your NGC Personal Key:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password=<NGC_PERSONAL_KEY>
helm repo update
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
  --set global.config.kubernetesDistribution=openshift \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
  --set supportEntitled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds
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
  --set tenantsManager.config.adminPassword="myPassw0rd\!" \
  --set supportEntitled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds
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

The same `helm upgrade` command installs the control plane for both Open Source and Support-Entitled Release. Update the values as described below; Support-Entitled Release installations additionally add `--set supportEntitled=true` to the `helm upgrade` command. The `control-plane-airgap-values.yaml` file is written by the `airgap load` step in [Preparations](preparations.md#software-artifacts):

* Replace `control-plane-<VERSION>.tgz` with the full filename of the control plane Helm chart (e.g., `control-plane-3.0.2.tgz`), located in the `charts` folder of the working directory you specified to `airgap load`.
* Replace `<VALUES_DIR>` with the `valuesDir` path you set in `airgap-load.yaml` during [Preparations](preparations.md#software-artifacts) - this is where `control-plane-airgap-values.yaml` was written.

* Replace `<OPENSHIFT-CLUSTER-DOMAIN>` with the domain configured for the OpenShift cluster. To determine the OpenShift cluster domain, run `oc get routes -A`.
* Set `global.customCA.enabled=true` as described in the [Control Plane Local CA](preparations.md#control-plane-local-ca) section.
* Set `tenantsManager.config.adminUsername=<ADMIN_EMAIL>` to the administrator's email address.
* Set `tenantsManager.config.adminPassword=<ADMIN_PASSWORD>` to the initial administrator password. The password must meet the following requirements:
  * Minimum Length: Passwords must be at least 8 characters long.
  * Digits: Must contain at least 1 numeric digit (0-9).
  * Lowercase Characters: Must contain at least 1 lowercase letter (a-z).
  * Uppercase Characters: Must contain at least 1 uppercase letter (A-Z).
  * Special Characters: Must contain at least 1 special character (e.g., !, @, #, $).

Open Source:

```bash
helm upgrade -i runai-backend ./charts/control-plane-<VERSION>.tgz -n runai-backend \
    --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
    --set global.config.kubernetesDistribution=openshift \
    --set global.customCA.enabled=true \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    -f <VALUES_DIR>/control-plane-airgap-values.yaml
```

Support-Entitled Release:

```bash
helm upgrade -i runai-backend ./charts/control-plane-<VERSION>.tgz -n runai-backend \
    --set global.domain=runai.apps.<OPENSHIFT-CLUSTER-DOMAIN> \
    --set global.config.kubernetesDistribution=openshift \
    --set global.customCA.enabled=true \
    --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
    --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
    --set supportEntitled=true \
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
    -f <VALUES_DIR>/control-plane-airgap-values.yaml
```

<Note>

**Note**

Use the dry-run flag `--dry-run=client` to gain an understanding of what is being installed before the actual installation.

</Note>

</Accordion>

## Enable Telemetry Service (Support-Entitled)

This section applies to Connected installations only. Air-gapped environments have no outbound connectivity to NVIDIA, so telemetry is not applicable.

The control plane can export authenticated telemetry to NVIDIA using an NGC Service API Key (SAK).

<Note>

**Note**

Enabling telemetry requires setting `supportEntitled=true`. You must also provide the SAK using one of the two methods below.

</Note>

### Scenario 1: Pass the SAK Value Directly

Pass the SAK as a plaintext Helm value. The chart creates and manages the Kubernetes Secret for you.

```bash
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=<DOMAIN> \
  --set global.ingress.ingressClass=haproxy \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
  --set supportEntitled=true \
  --set telemetryService.telemetryAuth.token=<SAK_TOKEN>
```

<Warning>

**Warning**

The SAK value is passed as a `--set` flag, which may be visible in shell history and process listings.

</Warning>

### Scenario 2: Reference a Pre-Created Secret

Create the Secret containing the SAK yourself before installing:

```bash
kubectl create secret generic runai-telemetry-sak \
  --namespace runai-backend \
  --from-literal=TELEMETRY_TOKEN=<SAK_TOKEN>
```

Then reference it by name at install time - the chart does not create or manage this Secret. Always set the secret key name explicitly, even if it matches the default (`TELEMETRY_TOKEN`):

```bash
helm upgrade -i runai-backend -n runai-backend runai/control-plane \
  --set global.domain=<DOMAIN> \
  --set global.ingress.ingressClass=haproxy \
  --set tenantsManager.config.adminUsername=<ADMIN_EMAIL> \
  --set tenantsManager.config.adminPassword="<ADMIN_PASSWORD>" \
  --set supportEntitled=true \
  --set telemetryService.telemetryAuth.existingSecret=runai-telemetry-sak \
  --set telemetryService.telemetryAuth.secretKey=TELEMETRY_TOKEN
```

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

With the control plane installed, continue to [Install the Cluster](helm-install.md). Before installing the cluster, review [Prerequisites and Features](prerequisites-and-features.md) for any additional setup the NVIDIA Run:ai cluster requires.
