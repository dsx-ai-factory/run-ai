# Upgrade NVIDIA Run:ai

This page walks through the NVIDIA Run:ai upgrade procedure: preparing software artifacts, upgrading the control plane, upgrading the cluster, and verifying the result.

<Note>

**Note**

Upgrades are performed using Manual Helm, regardless of which path you used to install NVIDIA Run:ai. Guided Installation does not support upgrades.

</Note>

## Before You Begin

Complete the following before starting the upgrade:

* [Plan the Upgrade](plan-the-upgrade.md)
* [Back Up Before Upgrading](backup-before-upgrading.md)
* [Migrate GPU Fractions](migrate-gpu-fractions.md), if your cluster uses GPU fractions

## Helm

NVIDIA Run:ai requires [Helm](https://helm.sh/) 3.14 or later. Before you continue, validate your installed Helm client version. To install or upgrade Helm, see [Installing Helm](https://helm.sh/docs/intro/install/).

<Note>

**Note**

Helm 4 defaults to [server-side apply](https://helm.sh/docs/overview/#server-side-apply) when installing a new chart release, which can conflict with resources managed by the NVIDIA Run:ai operator. Append `--server-side=false` to your `helm upgrade` command. NVIDIA Run:ai clusters originally installed with Helm 3.x are unaffected.

</Note>

## Prepare Software Artifacts

<Note>

**Note**

Following a deprecation that started in v2.25, JFrog is no longer supported as a software artifact source. NGC is the only supported path for upgrading NVIDIA Run:ai.

</Note>

<Accordion title="Connected (NGC)">

**Connected (NGC)**

For connected environments, prepare the artifacts according to the release type you are upgrading to.

**Open Source**

Open Source artifacts are publicly available from NGC and do not require authentication.

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm search repo -l runai/control-plane
```

**Support-Entitled**

Before upgrading, make sure an image pull secret is already configured, as described in [Software Artifacts](../installation/install-using-helm/preparations.md#software-artifacts).

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai-oss --force-update
helm repo update
helm search repo -l runai/control-plane
```

<Note>

**Note**

If you are upgrading from a JFrog-based installation, prepare the NGC-based `runai-reg-creds` secret as described in [Software Artifacts](../installation/install-using-helm/preparations.md#software-artifacts).

</Note>

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

The artifact preparation process is the same for Open Source and Support-Entitled environments.

Before upgrading:

* Set `runaiVersion` in `airgap-pack.yaml` to the NVIDIA Run:ai version you are upgrading to.
* Complete the `airgap pack` and `airgap load` procedures described in [Software Artifacts](../installation/install-using-helm/preparations.md#software-artifacts).

Make sure the required artifacts are available in the air-gapped environment before upgrading the control plane and cluster.

</Accordion>

## Upgrade the Control Plane

### Before You Upgrade

Before upgrading the NVIDIA Run:ai control plane, review the [System Requirements](../system-requirements/README.md) and complete the applicable steps in [Preparations](../installation/install-using-helm/preparations.md).

### Upgrade

<Accordion title="Connected (NGC)">

**Connected (NGC)**

Run the following command for the release type you are upgrading to and update the values as described below.

**Open Source**

Replace `<VERSION>` with the NVIDIA Run:ai version you are upgrading to.

```bash
helm get values runai-backend -n runai-backend > runai_control_plane_values.yaml
helm upgrade runai-backend -n runai-backend runai/control-plane \
  --version "<VERSION>" \
  -f runai_control_plane_values.yaml \
  --reset-then-reuse-values
```

**Support-Entitled**

* Replace `<VERSION>` with the NVIDIA Run:ai version you are upgrading to.
* Set `supportEntitled=true` to configure the installation as Support-Entitled.
* Set `telemetryService.enabled=true` to enable NVIDIA Run:ai's telemetry.
* Set `global.imagePullSecrets[0].name=runai-reg-creds` to use the NGC image pull secret prepared in [Software Artifacts](#prepare-software-artifacts).
* Optional: associate NVIDIA Run:ai's telemetry with your NCA ID by adding one of the following to the `helm upgrade` command. See [Registry Credentials](../installation/install-using-helm/preparations.md#registry-credentials) for how to generate the NGC Service API Key. By default, telemetry is not associated with your organization.
  - `--set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>` - Passes the Service API Key directly. Visible in shell history and process listings.
  - `--set telemetryService.telemetryAuth.existingSecret=<SECRET_NAME> --set telemetryService.telemetryAuth.secretKey=<KEY_NAME>` - References a secret you create yourself beforehand - the chart does not create or manage it.

<Note>

**Note**

These flags are new in v3.0, so add them explicitly when upgrading from v2.26 or earlier. `--reset-then-reuse-values` keeps the rest of your current configuration, such as your domain and tenant administrator settings. When you upgrade from v3.0 or later, it also keeps these flags, so you do not need to pass them again.

</Note>

```bash
helm get values runai-backend -n runai-backend > runai_control_plane_values.yaml
helm upgrade runai-backend -n runai-backend runai/control-plane \
  --version "<VERSION>" \
  -f runai_control_plane_values.yaml \
  --reset-then-reuse-values \
  --set supportEntitled=true \
  --set telemetryService.enabled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds \
  --set telemetryService.telemetryAuth.token=<SERVICE_API_KEY>
```

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

<Warning>

**Prerequisite**

If your internal registry requires authentication, you must create the `runai-reg-creds` imagePullSecret before proceeding. See [Software Artifacts](#prepare-software-artifacts) in Preparations.

</Warning>

Run the following command for the release type you are upgrading to and update the values as described below.

* Replace `<VERSION>` with the NVIDIA Run:ai version you are upgrading to.
* Replace `<VALUES_DIR>` with the `valuesDir` path configured in `airgap-load.yaml`. The `control-plane-airgap-values.yaml` file is generated by `airgap load`.
* If your internal registry requires authentication, add `--set global.imagePullSecrets[0].name=runai-reg-creds` to reference the secret created in [Software Artifacts](#prepare-software-artifacts).

**Open Source**

```bash
helm upgrade runai-backend ./charts/control-plane-<VERSION>.tgz -n runai-backend \
  -f <VALUES_DIR>/control-plane-airgap-values.yaml \
  --reset-then-reuse-values
```

**Support-Entitled**

Add `--set supportEntitled=true` and `--set telemetryService.enabled=true` to the `helm upgrade` command:

<Note>

**Note**

These flags are new in v3.0, so add them explicitly when upgrading from v2.26 or earlier. When you upgrade from v3.0 or later, `--reset-then-reuse-values` keeps them, so you do not need to pass them again.

</Note>

```bash
helm upgrade runai-backend ./charts/control-plane-<VERSION>.tgz -n runai-backend \
  -f <VALUES_DIR>/control-plane-airgap-values.yaml \
  --reset-then-reuse-values \
  --set supportEntitled=true \
  --set telemetryService.enabled=true
```

</Accordion>

## Upgrade the Cluster

### Before You Upgrade

Before upgrading the NVIDIA Run:ai cluster, review the [System Requirements](../system-requirements/README.md) and complete the applicable steps in [Preparations](../installation/install-using-helm/preparations.md) and [Prerequisites and Features](../installation/install-using-helm/prerequisites-and-features.md).

<Note>

**Note**

It is highly recommended to upgrade the Kubernetes version together with the NVIDIA Run:ai cluster version, to ensure compatibility with the latest supported version of your Kubernetes or OpenShift distribution. See [Kubernetes and OpenShift Compatibility](plan-the-upgrade.md#kubernetes-and-openshift-compatibility).

</Note>

### Setup

1. In the NVIDIA Run:ai UI, go to Resources -> Clusters
2. Select the cluster you want to upgrade
3. Click **INSTALLATION INSTRUCTIONS**
4. Optional: Select the NVIDIA Run:ai cluster version (latest, by default)
5. Under **Settings**, review the cluster location (**Same as control plane** or **Remote control plane**) and Cluster URL
6. Click **CONTINUE**

### Run the Upgrade Command

The command below is the same for both Open Source and Support-Entitled Release upgrades.

<Note>

**Note**

The `bootstrapToken` in the generated Helm command is a short-lived token that expires 2 hours after generation. If you do not complete the installation within that window, return to the cluster wizard in the NVIDIA Run:ai UI to generate a new command before running it.

</Note>

<Accordion title="Connected (NGC)">

**Connected (NGC)**

1. The NVIDIA Run:ai platform displays the Helm upgrade command in the cluster wizard.
2.  Modify the UI-generated command as follows:

    * If you are using a local certificate authority, add `--set global.customCA.enabled=true` to the Helm command as described in the [Cluster Local CA](../installation/install-using-helm/preparations.md#cluster-local-ca) section.
    * The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class to match the ingress controller configured during the [control plane installation](../installation/install-using-helm/install-control-plane.md#kubernetes).
    * Keep the remaining `--set` values exactly as generated by the UI.

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
3. Click **DONE**

</Accordion>

<Accordion title="Air-gapped">

**Air-gapped**

1. The NVIDIA Run:ai platform displays the Helm upgrade command in the cluster wizard.\
   **Do not run the command exactly as shown in the UI.**
2.  Update the UI-generated Helm command as follows:

    * Do not add the Helm repository - skip `helm repo add` and `helm repo update`.
    * Replace `runai/runai-cluster` with `./charts/runai-cluster-<VERSION>.tgz`, the chart written by `airgap load`.
    * Add `-f <VALUES_DIR>/cluster-airgap-values.yaml`. Replace `<VALUES_DIR>` with the `valuesDir` path you set in `airgap-load.yaml`.
    * Add `--set global.customCA.enabled=true` as described in the [Cluster Local CA](../installation/install-using-helm/preparations.md#cluster-local-ca) section.
    * The recommended ingress controller is HAProxy. If you are using a different ingress controller, update the ingress class to match the ingress controller configured during the [control plane installation](../installation/install-using-helm/install-control-plane.md#kubernetes).
    * Keep the remaining `--set` values exactly as generated by the UI.

    ```bash
    helm upgrade -i runai-cluster ./charts/runai-cluster-<VERSION>.tgz \
      --set controlPlane.url=... \
      --set controlPlane.bootstrapToken=... \
      --set cluster.uid=... \
      --set cluster.url=... --create-namespace \
      --set global.customCA.enabled=true \
      --set clusterConfig.global.ingress.ingressClass=haproxy \
      -f <VALUES_DIR>/cluster-airgap-values.yaml
    ```
3. Click **DONE**

</Accordion>

Once installation is complete, validate the cluster is **Connected** and listed with the new cluster version. Once you have done this, the cluster is upgraded to the latest version. If the cluster does not appear as expected, see the [cluster troubleshooting scenarios](../operations/clusters.md#troubleshooting-scenarios).

## Migrate from NGINX to HAProxy Kubernetes Ingress Controller

<Note>

**Note**

This section applies to Kubernetes only. OpenShift includes a pre-installed ingress controller by default and does not require this migration.

</Note>

NVIDIA Run:ai recommends using [HAProxy Kubernetes Ingress Controller](https://www.haproxy.com/documentation/kubernetes-ingress/) as the ingress controller. This change aligns with the announced retirement of the upstream NGINX Ingress Controller project.

Clusters upgraded from earlier versions typically already have NGINX installed. After upgrading to `<VERSION>`, follow the steps below to migrate ingress traffic from NGINX Ingress Controller to HAProxy Kubernetes Ingress Controller.

<Note>

**Migration Resources**

You can utilize additional resources to streamline your migration. Visit the [HAProxy Migration Center](https://www.haproxy.com/landing/ingress-nginx-retirement) to access a dedicated migration assistant tool and a technical webinar on moving from NGINX to HAProxy.

</Note>

### Check the Service Type of the Existing Ingress Controller

Before installing the HAProxy Kubernetes Ingress Controller, identify which ingress controller is currently in use. If your cluster already has an ingress controller installed, verify how it is exposed to avoid port address conflicts.

```bash
kubectl get svc -n <nginx-namespace>
```

* If the existing ingress controller uses **NodePort**, note the HTTP/HTTPS NodePort values to ensure HAProxy is configured with non-overlapping ports.
* If the existing ingress controller uses **LoadBalancer**, no additional action is required.

When running more than one ingress controller in the same cluster, port conflicts are relevant only for NodePort-based setups. LoadBalancer-based controllers automatically receive separate external IP addresses.

<Note>

**Note**

If your setup differs from the examples above, adjust the configuration accordingly. When using external LoadBalancer on top of Ingress with service type NodePort, you may need to update external resources to route traffic to HAProxy’s configured NodePort values.

</Note>

### Install and Configure HAProxy Kubernetes Ingress Controller

Ingress controllers can be installed and configured in different ways depending on your Kubernetes distribution and how you expose services (for example, NodePort vs. LoadBalancer).

The sections below provide environment-specific Helm installation examples. Select the option that matches your deployment environment.

<Note>

**Note**

OpenShift and RKE2 include a pre-installed ingress controller by default.

</Note>

<Accordion title="Vanilla Kubernetes">

**Vanilla Kubernetes**

If your cluster already has an ingress controller installed (for example, NGINX) and it is exposed via NodePort, configure HAProxy to use different NodePort values so both controllers can run simultaneously.

**Ensure the selected NodePort values do not overlap with ports already used by the existing ingress controller.**

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
 --create-namespace \
 --namespace haproxy-controller \
 --set controller.ingressClassResource.enabled=true \
 --set controller.service.type=NodePort \
 --set controller.service.nodePorts.http=32080 \
 --set controller.service.nodePorts.https=32443 \
 --set-string controller.config.ssl-redirect-port=32443
```

<Note>

**Note**

For ingresses with TLS, HAProxy redirects HTTP requests to the port set in `controller.config.ssl-redirect-port`. The default is 8443, the controller's container port, which clients cannot reach through a NodePort Service. Set it to the HTTPS port that clients connect to. Use the HTTPS NodePort (32443 in this example) when clients reach the nodes directly, or 443 when an external load balancer forwards port 443 to the HTTPS NodePort. For more information, refer to [ssl-redirect-port](https://www.haproxy.com/documentation/kubernetes-ingress/community/configuration-reference/configmap/) in the HAProxy Kubernetes Ingress Controller documentation.

</Note>

</Accordion>

<Accordion title="Managed Kubernetes (EKS, GKE, AKS)">

**Managed Kubernetes (EKS, GKE, AKS)**

When using a LoadBalancer, each ingress controller automatically receives its own external IP address from the cloud provider. This allows multiple ingress controllers to run in the same cluster without additional configuration.

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
 --create-namespace \
 --namespace haproxy-controller \
 --set controller.service.type=LoadBalancer
```

</Accordion>

<Accordion title="Oracle Kubernetes Engine (OKE)">

**Oracle Kubernetes Engine (OKE)**

When using a LoadBalancer, each ingress controller automatically receives its own external IP address from the cloud provider. This allows multiple ingress controllers to run in the same cluster without additional configuration.

```bash
helm repo add haproxytech https://haproxytech.github.io/helm-charts
helm repo update
helm install haproxy-kubernetes-ingress haproxytech/kubernetes-ingress \
 --create-namespace \
 --namespace haproxy-controller \
 --set controller.kind=DaemonSet \
 --set controller.service.type=LoadBalancer \
 --set controller.service.externalTrafficPolicy=Local \
 --set controller.service.annotations."oci-network-load-balancer\.oraclecloud\.com/is-preserve-source"="True" \
 --set controller.service.annotations."oci-network-load-balancer\.oraclecloud\.com/security-list-management-mode"=All \
 --set controller.service.annotations."oci\.oraclecloud\.com/load-balancer-type"=nlb
```

</Accordion>

### Verify HAProxy Ingress

After installing the HAProxy Kubernetes Ingress Controller, verify that HAProxy ingresses are reachable before switching NVIDIA Run:ai components to use it. You can do this by deploying a simple hello-world application.

To run the test, identify the IP address that should reach the cluster’s nodes in your environment.

1.  Create a local `haproxy-test.yml` file:

    ```yaml
    apiVersion: apps/v1
    kind: Deployment
    metadata:
      name: hello
    spec:
      replicas: 1
      selector:
        matchLabels:
          app: hello
      template:
        metadata:
          labels:
            app: hello
        spec:
          containers:
          - name: hello
            image: hashicorp/http-echo:1.0
            args:
              - "-text=hello from haproxy-ingress"
            ports:
              - containerPort: 5678
    ---
    apiVersion: v1
    kind: Service
    metadata:
      name: hello
    spec:
      selector:
        app: hello
      ports:
      - port: 80
        targetPort: 5678
    ---
    apiVersion: networking.k8s.io/v1
    kind: Ingress
    metadata:
      name: hello
    spec:
      ingressClassName: haproxy
      rules:
      - http:
          paths:
          - path: /
            pathType: Prefix
            backend:
              service:
                name: hello
                port:
                  number: 80
    ```
2.  Run the following command:

    ```yaml
    kubectl apply -f haproxy-test.yml
    ```

Once the application is deployed, access the cluster’s IP address in a browser. If the page displays **“hello from haproxy-ingress”**, HAProxy is functioning correctly and you can proceed with upgrading NVIDIA Run:ai.

### Upgrade the Control Plane

Run the following Helm command to update the NVIDIA Run:ai control plane to use HAProxy instead of NGINX.

```bash
helm upgrade runai-backend -n runai-backend runai/control-plane \
  --version "<VERSION>" \
  --reuse-values \
  --set-string global.ingress.ingressClass=haproxy
```

### Upgrade the Cluster

#### Setup

1. In the NVIDIA Run:ai UI, go to **Resources -> Clusters**
2. Select the cluster you want to upgrade
3. Click **INSTALLATION INSTRUCTIONS**
4. Optional: Select the NVIDIA Run:ai cluster version (latest, by default)
5. Under **Settings**, review the cluster location (**Same as control plane** or **Remote control plane**) and Cluster URL
6. Click **CONTINUE**

#### Installation Instructions

1. Follow the installation instructions. Run the Helm commands provided on your Kubernetes cluster.
2.  If not present, add the following flag to the helm install command:

    ```bash
    --set clusterConfig.global.ingress.ingressClass=haproxy
    ```
3. Click **DONE**
4. Once installation is complete, validate the cluster is **Connected** and listed with the new cluster version (see the [cluster troubleshooting scenarios](../operations/clusters.md#troubleshooting-scenarios)). Once you have done this, the cluster is upgraded and the workloads in this cluster will now use HAProxy instead of NGINX.

## Troubleshooting

If you encounter an issue with the cluster upgrade, use the troubleshooting scenarios below.

### Installation Fails

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

### Cluster Status

If the NVIDIA Run:ai cluster upgrade completes, but the cluster status does not show as **Connected**, refer to [Troubleshooting scenarios](../operations/clusters.md#troubleshooting-scenarios).
