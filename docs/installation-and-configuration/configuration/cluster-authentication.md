# Cluster Authentication

To allow users to securely submit workloads using `kubectl`, you must configure the Kubernetes API server to authenticate users via the NVIDIA Run:ai identity provider. This is done by adding OpenID Connect (OIDC) flags to the Kubernetes API server configuration on each cluster.&#x20;

Each predefined role is mapped to Kubernetes ClusterRoles. When users authenticate to the cluster with their NVIDIA Run:ai identity, they receive the permissions of these ClusterRoles when they use `kubectl`. To see what each role grants in Kubernetes, refer to [Role to ClusterRole Mapping](../../platform-management/authentication/role-clusterrole-mapping.md).

### Retrieve Required OIDC Flags

1. Go to **General settings**
2. Navigate to **Cluster authentication**

```yaml
  containers:
  - command:
    ...
    - --oidc-client-id=runai
    - --oidc-issuer-url=https://<HOST>/auth/realms/runai
    - --oidc-username-prefix=-
```

* `--oidc-client-id` - A client id that all tokens must be issued for.
* `--oidc-issuer-url` - The URL of the NVIDIA Run:ai identity provider
* `--oidc-username-prefix` - Prefix prepended to username claims to prevent clashes with existing names (e.g., `-user@example.com`).

<Note>

**Note**

These flags must be configured in the API server startup parameters for each cluster in your environment.

</Note>

## Kubernetes Distribution-Specific Configuration

<Note>

**Note**

* Azure Kubernetes Service (AKS) is not supported.
* For other Kubernetes distributions, refer to specific instructions in the documentation.

</Note>

<Accordion title="Vanilla Kubernetes">

**Vanilla Kubernetes**

1. Locate the Kubernetes API server configuration file. For vanilla Kubernetes, the configuration file is typically located at: `/etc/kubernetes/manifests/kube-apiserver.yaml`.
2. Edit the file. Under the `command` section, add the [required OIDC flags](cluster-authentication.md#retrieve-required-oidc-values).
3.  Verify that the changes have been applied. After saving the file, the API server should automatically restart since it is managed as a static pod. Confirm that the `kube-apiserver-<master-node-name>` pod in the `kube-system` namespace has restarted and is running with the new configuration. You can run the following command to check the pod status:

    ```bash
    kubectl get pods -n kube-system kube-apiserver-<master-node-name> -o yaml
    ```

</Accordion>

<Accordion title="OpenShift Container Platform (OCP)">

**OpenShift Container Platform (OCP)**

No additional configuration is required.

</Accordion>

<Accordion title="Rancher Kubernetes Engine 2 (RKE2)">

**Rancher Kubernetes Engine 2 (RKE2)**

If you are using the [RKE2 Quickstart](https://docs.rke2.io/install/quickstart/):

1. Edit `/etc/rancher/rke2/config.yaml`.
2.  Add the [required OIDC flags](cluster-authentication.md#retrieve-required-oidc-values) under `kube-apiserver-arg`, using the format shown below:

    ```yaml
    kube-apiserver-arg:
    - "oidc-client-id=runai" # 
    ...
    ```

If you are using Rancher UI:

1. Add the required flags during the cluster provisioning process.
2. Navigate to: Cluster Management > Create, select RKE2, and choose your platform.
3. In the Cluster Configuration screen, go to: Advanced > Additional API Server Args.
4. Add the [required OIDC flags](cluster-authentication.md#retrieve-required-oidc-values) as `<key>=<value>` (e.g. `oidc-username-prefix=-`).

</Accordion>

<Accordion title="Google Kubernetes Engine (GKE)">

**Google Kubernetes Engine (GKE)**

To configure researcher authentication on GKE, use **Anthos Identity Service** and apply the appropriate OIDC configuration.

1.  Install [Anthos identity service](https://cloud.google.com/kubernetes-engine/docs/how-to/oidc#enable-oidc) by running:

    ```bash
    gcloud container clusters update <gke-cluster-name> \
        --enable-identity-service --project=<gcp-project-name> --zone=<gcp-zone-name>
    ```

2) Install the [yq](https://github.com/mikefarah/yq) utility.
3)  Configure the OIDC provider for username-password authentication. Make sure to use the [required OIDC flags](cluster-authentication.md#retrieve-required-oidc-values):&#x20;

    ```bash
    kubectl get clientconfig default -n kube-public -o yaml > login-config.yaml
    yq -i e ".spec +={\"authentication\":[{\"name\":\"oidc\",\"oidc\":{\"clientID\":\"runai\",\"issuerURI\":\"$OIDC_ISSUER_URL\",\"kubectlRedirectURI\":\"http://localhost:8000/callback\",\"userClaim\":\"sub\",\"userPrefix\":\"-\"}}]}" login-config.yaml
    kubectl apply -f login-config.yaml
    ```

4.  Or, configure the OIDC provider for single-sign-on. Make sure to use the [required OIDC flags](cluster-authentication.md#retrieve-required-oidc-values):&#x20;

    ```bash
    kubectl get clientconfig default -n kube-public -o yaml > login-config.yaml
    yq -i e ".spec +={\"authentication\":[{\"name\":\"oidc\",\"oidc\":{\"clientID\":\"runai\",\"issuerURI\":\"$OIDC_ISSUER_URL\",\"groupsClaim\":\"groups\",\"kubectlRedirectURI\":\"http://localhost:8000/callback\",\"userClaim\":\"sub\",\"userPrefix\":\"-\"}}]}" login-config.yaml
    kubectl apply -f login-config.yaml
    ```

5)  Update the `runaiconfig` with the Anthos Identity Service endpoint. First, get the external IP of the `gke-oidc-envoy` service:

    ```bash
    kubectl get svc -n anthos-identity-service
    NAME               TYPE           CLUSTER-IP    EXTERNAL-IP     PORT(S)              AGE
    gke-oidc-envoy     LoadBalancer   10.37.3.111   39.201.319.10   443:31545/TCP        12h
    ```

6.  Then, patch the `runaiconfig` to use this endpoint. Replace the below with the actual IP address of the `gke-oidc-envoy` service:

    ```bash
    kubectl -n runai patch runaiconfig runai -p '{"spec": {"researcher-service": 
    {"args": {"gkeOidcEnvoyHost": "35.236.229.19"}}}}'  --type="merge"
    ```

</Accordion>

<Accordion title="Elastic Kubernetes Engine (EKS)">

**Elastic Kubernetes Engine (EKS)**

1. In the AWS Console, under EKS, find your cluster.

2) Go to `Configuration` and then to `Authentication`.

3. Associate a new `identity provider`. Use the [required OIDC flags](cluster-authentication.md#retrieve-required-oidc-values).

The process can take up to 30 minutes.

</Accordion>

<Accordion title="NVIDIA Base Command Manager (BCM)">

**NVIDIA Base Command Manager (BCM)**

While it is possible to edit the manifest files directly on the control plane nodes (as described in the Vanilla Kubernetes instructions), the BCM preferred method is to make changes through the kubeadm configmap. This ensures that kubeadm can always regenerate the manifest files correctly, for example, when adding new control plane nodes or upgrading Kubernetes.

<Warning>

**Warning**

If you previously modified `/etc/kubernetes/manifests/kube-apiserver.yaml` directly on the nodes, those changes may be overwritten the next time `cm-kubeadm-manage` is used or when Kubernetes is upgraded. Use the configmap method below to ensure your OIDC flags persist.

</Warning>

1.  Edit the kubeadm configmap:

    ```bash
    kubectl edit configmap -n kube-system kubeadm-config
    ```

2.  Add the [required OIDC flags](cluster-authentication.md#retrieve-required-oidc-values) under the `apiServer.extraArgs` block:

    ```yaml
    extraArgs:
    - name: oidc-client-id
      value: "runai"
    - name: oidc-issuer-url
      value: "https://<HOST>/auth/realms/runai"
    - name: oidc-username-prefix
      value: "-"
    ```

    A full example of the relevant configmap section:

    ```yaml
    apiVersion: v1
    data:
      ClusterConfiguration: |
        apiServer:
          certSANs:
          - <your-cluster-hostname>
          - master
          - localhost
          extraArgs:
          - name: oidc-client-id
            value: "runai"
          - name: oidc-issuer-url
            value: "https://<HOST>/auth/realms/runai"
          - name: oidc-username-prefix
            value: "-"
        apiVersion: kubeadm.k8s.io/v1beta4
    ```

3.  Propagate the configmap changes to the control plane nodes:

    ```bash
    /cm/local/apps/cmd/scripts/cm-kubeadm-manage --kube-cluster default update_configmap
    ```

4.  For each control plane node, regenerate the API server manifest:

    ```bash
    /cm/local/apps/cmd/scripts/cm-kubeadm-manage --kube-cluster default update_apiserver <node-name>
    ```

    Validate the changes by confirming the API server pod is using the OIDC flags:

    ```bash
    kubectl describe pod -n kube-system kube-apiserver-<node-name> | grep oidc
    ```

    Expected output:

    ```
    --oidc-client-id=runai
    --oidc-issuer-url=https://<HOST>/auth/realms/runai
    --oidc-username-prefix=-
    ```

    Once confirmed, proceed to the next control plane node and repeat until all nodes are updated.

</Accordion>
