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
* [Migrate GPU Fractions](migrate-gpu-fractions.md), if your tenant clusters use GPU fractions

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

## Upgrade the Control Plane

### Before You Upgrade

Before upgrading the NVIDIA Run:ai control plane, review the [System Requirements](../system-requirements/README.md) and complete the applicable steps in [Preparations](../installation/install-using-helm/preparations.md).

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
 --set controller.service.nodePorts.https=32443 \
 --set-string controller.config.ssl-redirect-port=32443
```

<Note>

**Note**

For ingresses with TLS, HAProxy redirects HTTP requests to the port set in `controller.config.ssl-redirect-port`. The default is 8443, the controller's container port, which clients cannot reach through a NodePort Service. Set it to the HTTPS port that clients connect to. Use the HTTPS NodePort (32443 in this example) when clients reach the nodes directly, or 443 when an external load balancer forwards port 443 to the HTTPS NodePort. For more information, refer to [ssl-redirect-port](https://www.haproxy.com/documentation/kubernetes-ingress/community/configuration-reference/configmap/) in the HAProxy Kubernetes Ingress Controller documentation.

</Note>

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
