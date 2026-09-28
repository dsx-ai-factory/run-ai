# Upgrade

This page walks through the NVIDIA Run:ai upgrade procedure: preparing software artifacts, upgrading the control plane, upgrading the cluster, and verifying the result.

<Note>

**Note**

Upgrades are performed using Manual Helm, regardless of which path you used to install NVIDIA Run:ai. Guided Installation does not support upgrades.

</Note>

## Before Upgrade

NVIDIA Run:ai supports a maximum gap of four versions per upgrade; attempting to skip more than four versions in a single upgrade may fail. If your target version is more than four versions ahead of your current version, perform multiple upgrades. Each upgrade can span a maximum of four versions, and each upgrade must complete successfully before you begin the next.

<Note>

**Note**

For Support-Entitled Releases, upgrades are officially supported and tested only for versions within the current Supported Versions.

</Note>

To ensure a smooth and supported upgrade process:

* **Align control plane and cluster versions** - For best results, upgrade the control plane and cluster components to the same NVIDIA Run:ai version during the same maintenance window. Keeping versions aligned helps avoid unexpected behavior caused by version mismatches and ensures full compatibility across platform components.
* **Upgrade order** - When performing an upgrade:
  * Upgrade the control plane Helm chart first
  * Upgrade the cluster Helm chart only after the control plane upgrade completes successfully

## Kubernetes Compatibility

Before upgrading NVIDIA Run:ai, verify that your Kubernetes distribution version is compatible with the target NVIDIA Run:ai version. If your current Kubernetes distribution release falls outside the supported range, upgrade your distribution first.

| NVIDIA Run:ai version | Supported Kubernetes versions |
| ---------------------- | ------------------------------ |
| 3.0 (latest)            | 1.35 to 1.37                   |
| 2.26                    | 1.34 to 1.36                   |
| 2.25                    | 1.33 to 1.35                   |
| 2.24                    | 1.33 to 1.35                   |

For information on supported versions of managed Kubernetes, consult the release notes provided by your Kubernetes service provider. For an up-to-date end-of-life statement see [Kubernetes Release History](https://kubernetes.io/releases/).

## Helm

NVIDIA Run:ai requires [Helm](https://helm.sh/) 3.14 or later. Before you continue, validate your installed helm client version. To install or upgrade Helm, see [Installing Helm](https://helm.sh/docs/intro/install/).

<Note>

**Note**

Helm 4 defaults to [server-side apply](https://helm.sh/docs/overview/#server-side-apply) when installing a new chart release, which can conflict with resources managed by the NVIDIA Run:ai operator. Append `--server-side=false` to your `helm upgrade` command. NVIDIA Run:ai clusters originally installed with Helm 3.x are unaffected.

</Note>

## Prepare Software Artifacts

<Note>

**Note**

Following a deprecation that started in v2.25, JFrog is no longer supported as a software artifact source. NGC is the only supported path for upgrading NVIDIA Run:ai.

</Note>

Prepare the artifacts according to the release type you are upgrading to.

**Open Source**

Open Source artifacts are publicly available from NGC and do not require authentication.

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update
helm repo update
helm search repo -l runai/control-plane
```

**Support-Entitled**

Support-Entitled artifacts require authentication to NGC. Before upgrading, make sure that:

* You have an NGC Personal Key.
* An image pull secret is already configured using the same NGC Personal Key, as described in [Software Artifacts](installation/install-using-helm/preparations.md#software-artifacts).

Add the NVIDIA Run:ai Helm repository using your NGC Personal Key:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password=<NGC_API_KEY>
helm repo update
helm search repo -l runai/control-plane
```

<Note>

**Note**

If you are upgrading from a JFrog-based installation, prepare the NGC-based `runai-reg-creds` secret as described in [Software Artifacts](installation/install-using-helm/preparations.md#software-artifacts).

</Note>

## Upgrade the Control Plane

### Before You Upgrade

Before upgrading the NVIDIA Run:ai control plane, review the [System Requirements](system-requirements/README.md) and complete the applicable steps in [Preparations](installation/install-using-helm/preparations.md).

### Upgrade

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
* Set `global.imagePullSecrets[0].name=runai-reg-creds` to use the NGC image pull secret prepared in [Software Artifacts](#prepare-software-artifacts).

```bash
helm get values runai-backend -n runai-backend > runai_control_plane_values.yaml
helm upgrade runai-backend -n runai-backend runai/control-plane \
  --version "<VERSION>" \
  -f runai_control_plane_values.yaml \
  --reset-then-reuse-values \
  --set supportEntitled=true \
  --set global.imagePullSecrets[0].name=runai-reg-creds
```

### Enable Telemetry Service (Support-Entitled)

NVIDIA Run:ai includes telemetry for Support-Entitled deployments, reporting initial metrics including platform usage and adoption signals, GPU utilization and allocation, and cluster health signals to support proactive troubleshooting and continued product improvement. By default, telemetry is not associated with your organization. You can opt in to identified telemetry by providing an [NGC Service API Key](https://docs.nvidia.com/ngc/latest/ngc-catalog-user-guide.html#generating-a-service-api-key) during installation, which associates the telemetry with your NCA ID.

<Accordion title="Pass the SAK Value Directly">

**Pass the SAK Value Directly**

Pass the SAK as a plaintext Helm value. The chart creates and manages the Kubernetes Secret for you.

```bash
helm upgrade runai-backend -n runai-backend runai/control-plane \
  --version "<VERSION>" \
  -f runai_control_plane_values.yaml \
  --reset-then-reuse-values \
  --set supportEntitled=true \
  --set telemetryService.telemetryAuth.token=<SAK_TOKEN>
```

<Warning>

**Warning**

The SAK value is passed as a `--set` flag, which may be visible in shell history and process listings.

</Warning>

</Accordion>

<Accordion title="Reference a Pre-Created Secret">

**Reference a Pre-Created Secret**

1. Create the Secret containing the SAK before upgrading:

```bash
kubectl create secret generic runai-telemetry-sak \
  --namespace runai-backend \
  --from-literal=TELEMETRY_TOKEN=<SAK_TOKEN>
```

2. Reference the Secret in the upgrade command. The chart does not create or manage it.

3. Set both the Secret name and key explicitly:

```bash
helm upgrade runai-backend -n runai-backend runai/control-plane \
  --version "<VERSION>" \
  -f runai_control_plane_values.yaml \
  --reset-then-reuse-values \
  --set supportEntitled=true \
  --set telemetryService.telemetryAuth.existingSecret=runai-telemetry-sak \
  --set telemetryService.telemetryAuth.secretKey=TELEMETRY_TOKEN
```

</Accordion>

## Migrate from NGINX to HAProxy Kubernetes Ingress Controller

<Note>

**Note**

This section applies to Kubernetes only. OpenShift includes a pre-installed ingress controller by default and does not require this migration.

</Note>

NVIDIA Run:ai recommends using [HAProxy Kubernetes Ingress Controller](https://www.haproxy.com/documentation/kubernetes-ingress/) as the ingress controller. This change aligns with the announced retirement of the upstream NGINX Ingress Controller project.

Clusters upgraded from earlier versions typically already have NGINX installed. After upgrading to `<VERSION>`, follow the steps below to migrate ingress traffic from NGINX Ingress Controller to HAProxy Kubernetes Ingress Controller.

<Note>

**Note**

You can utilize additional resources to streamline your migration. Visit the [HAProxy Migration Center](https://www.haproxy.com/landing/ingress-nginx-retirement) to access a dedicated migration assistant tool and a technical webinar on moving from NGINX to HAProxy.

</Note>

### Check the Service Type of the Existing Ingress Controller

Before installing the HAProxy ingress controller, identify which ingress controller is currently in use. If your cluster already has an ingress controller installed, verify how it is exposed to avoid port address conflicts.

```bash
kubectl get svc -n <nginx-namespace>
```

If the existing ingress controller uses **NodePort**, note the HTTP/HTTPS NodePort values to ensure HAProxy is configured with non-overlapping ports.

<Note>

**Note**

If your setup differs from the examples above, adjust the configuration accordingly. When using external LoadBalancer on top of Ingress with service type NodePort, you may need to update external resources to route traffic to HAProxy’s configured NodePort values.

</Note>

### Install and Configure HAProxy Kubernetes Ingress Controller

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
 --set controller.service.nodePorts.https=32443
```

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
    kubectl apply -f ha-proxy-test.yml
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
