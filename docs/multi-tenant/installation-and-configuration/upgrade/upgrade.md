# Upgrade

## Before Upgrade

Before proceeding with the upgrade, it's crucial to apply the specific prerequisites associated with your current version of NVIDIA Run:ai and every version in between up to the version you are upgrading to.

To ensure a smooth and supported upgrade process:

* **Align control plane and cluster versions** - For best results, upgrade the control plane and cluster components to the same NVIDIA Run:ai version during the same maintenance window. Keeping versions aligned helps avoid unexpected behavior caused by version mismatches and ensures full compatibility across platform components.
* **Upgrade order** - When performing an upgrade:
  * Upgrade the control plane Helm chart first
  * Upgrade the cluster Helm chart only after the control plane upgrade completes successfully

## Helm

NVIDIA Run:ai requires [Helm](https://helm.sh/) 3.14 or later. Before you continue, validate your installed helm client version. To install or upgrade Helm, see [Installing Helm](https://helm.sh/docs/intro/install/).

<Info>

**Note**

Helm 4 defaults to [server-side apply](https://helm.sh/docs/overview/#server-side-apply) when installing a new chart release, which can conflict with resources managed by the NVIDIA Run:ai operator. Append `--server-side=false` to your `helm upgrade` command. NVIDIA Run:ai clusters originally installed with Helm 3.x are unaffected.

</Info>

## Software Files

NVIDIA Run:ai artifacts are available on both NVIDIA NGC and JFrog. Deployments that were originally installed using JFrog can also be upgraded using NGC. As JFrog support is deprecated, upgrading via NGC is the recommended approach, provided that you have an NGC API key.

Use the tab that matches your environment:

* **NGC (Recommended)** - To upgrade using NGC, complete the [Preparations](../installation/install-using-helm/preparations.md) section first and make sure you have NGC API key.
* **JFrog** - Existing customers may chose to continue upgrading from JFrog or to upgrade from NGC.

<Tabs>

<Tab title="NGC">

Before upgrading, complete the steps in the [Preparations](../installation/install-using-helm/preparations.md) section to set up your image pull secret.

Run the following commands to add the NVIDIA Run:ai Helm repository and browse the available versions:

```bash
helm repo add runai https://helm.ngc.nvidia.com/nvidia/runai --force-update \
  --username='$oauthtoken' \
  --password=<NGC_API_KEY>
helm repo update
helm search repo -l runai/control-plane
```

</Tab>

<Tab title="JFrog">

Run the following commands to add the NVIDIA Run:ai Helm repository and browse the available versions:

```bash
helm repo add runai-backend https://runai.jfrog.io/artifactory/cp-charts-prod
helm repo update
helm search repo -l runai-backend
```

</Tab>

</Tabs>

## Upgrade Control Plane

### System and Network Requirements

Before upgrading the NVIDIA Run:ai control plane, validate that the latest [system requirements](../system-requirements/infrastructure-requirements.md) and [networking setup](../system-requirements/networking-setup.md) are met, as they can change from time to time.

### Upgrade

To upgrade run the following:

```bash
helm get values runai-backend -n runai-backend > runai_control_plane_values.yaml
helm upgrade runai-backend -n runai-backend runai/control-plane --version "<VERSION>" -f runai_control_plane_values.yaml --reset-then-reuse-values
```

<Info>

**Note**

To upgrade to a specific version, modify the `--version` flag by specifying the desired `<VERSION>`. You can find all available versions by using the `helm search repo runai/control-plane --versions` command.

</Info>

## Migrate from NGINX to HAProxy Kubernetes Ingress Controller

<Info>

**Note**

This section applies to Kubernetes only. OpenShift includes a pre-installed ingress controller by default and does not require this migration.

</Info>

NVIDIA Run:ai recommends using [HAProxy Kubernetes Ingress Controller](https://www.haproxy.com/documentation/kubernetes-ingress/) as the ingress controller. This change aligns with the announced retirement of the upstream NGINX Ingress Controller project.

Clusters upgraded from earlier versions typically already have NGINX installed. After upgrading to v2.25, follow the steps below to migrate ingress traffic from NGINX Ingress Controller to HAProxy Kubernetes Ingress Controller.

<Info>

**Migration Resources**

You can utilize additional resources to streamline your migration. Visit the [HAProxy Migration Center](https://www.haproxy.com/landing/ingress-nginx-retirement) to access a dedicated migration assistant tool and a technical webinar on moving from NGINX to HAProxy.

</Info>

### Check the Service Type of the Existing Ingress Controller

Before installing the HAProxy ingress controller, identify which ingress controller is currently in use. If your cluster already has an ingress controller installed, verify how it is exposed to avoid port address conflicts.

```bash
kubectl get svc -n <nginx-namespace>
```

If the existing ingress controller uses **NodePort**, note the HTTP/HTTPS NodePort values to ensure HAProxy is configured with non-overlapping ports.

<Info>

**Note**

If your setup differs from the examples above, adjust the configuration accordingly. When using external LoadBalancer on top of Ingress with service type NodePort, you may need to update external resources to route traffic to HAProxy’s configured NodePort values.

</Info>

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
  --version "~2.25.0" \
  --reuse-values \
  --set-string global.ingress.ingressClass=haproxy
```
