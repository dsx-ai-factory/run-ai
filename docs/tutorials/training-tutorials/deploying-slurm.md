# Deploying a Slurm Cluster on NVIDIA Run:ai with Slinky

This tutorial demonstrates how to deploy a Slurm cluster on an NVIDIA Run:ai-managed Kubernetes cluster using Slinky's `slurm-operator`. You install the Slurm control plane and compute nodes as Kubernetes workloads, then confirm NVIDIA Run:ai recognizes and schedules them. The workflow also covers GPU scheduling through Slurm GRES.

In this tutorial, you will learn how to:

* Log in to NVIDIA Run:ai and create an API access token
* Install Slinky and deploy a Slurm cluster in an NVIDIA Run:ai project namespace
* Confirm that NVIDIA Run:ai recognizes the Slurm control plane and compute workloads
* Add a GPU-enabled Slurm NodeSet and configure Slurm GRES scheduling
* Verify GPU scheduling from the Slurm login pod

## What This Covers

[Slinky](https://github.com/SlinkyProject) is an open source toolkit for running Slurm on Kubernetes. Its `slurm-operator` runs the Slurm controller (`slurmctld`) and compute daemons (`slurmd`) as Kubernetes pods represented by custom resources.

This is a standard Kubernetes installation. It uses Helm charts and custom resource definitions (CRDs).

This tutorial does not cover `slurm-bridge`, which enables Slurm to schedule native Kubernetes workloads.

## Prerequisites

Before you begin, make sure you have:

* A [project](../../platform-management/aiinitiatives/organization/projects.md) and its Kubernetes namespace. This tutorial uses the `slurm` namespace. Creating a project with `--namespace slurm` does not create the namespace. If it is missing, the project is `NotReady` with `Can't find namespace for project` until you create it with `kubectl create ns slurm`.
* `kubectl` and Helm access to the cluster.
* GPU-enabled nodes, if you plan to test GPU scheduling. Size the project GPU quota for the NodeSet request. NVIDIA Run:ai enforces this quota for NodeSets. An insufficient quota leaves the GPU NodeSet in `Pending`.

## Step 1: Logging In

<Tabs>

<Tab title="UI">

Browse to the provided NVIDIA Run:ai user interface and log in with your credentials.

</Tab>

<Tab title="API">

To use the API, obtain a token as shown in [Creating a User Access Key](deploying-slurm.md#step-2-creating-a-user-access-key).

</Tab>

<Tab title="CLI v2">

Run the following `--help` command to review login options and log in according to your setup:

```sh
runai login --help
```

</Tab>

</Tabs>

## Step 2: Creating a User Access Key

<Note>

**Note**

This step is only required if you intend to follow the API steps in this tutorial.

</Note>

Access keys are used for API integrations with NVIDIA Run:ai. An access key contains a client ID and a client secret. With the client credentials, you can obtain a token and use it within subsequent API calls.

In the NVIDIA Run:ai user interface:

1. Click the user avatar at the top right corner, then select **Settings**
2. Click **+ACCESS KEY**
3. Enter the access key's **name** and click **CREATE**
4. Copy the **Client ID** and **Client secret** and store securely
5. Click **DONE**

To request an API access token, use the client credentials to get a token to access NVIDIA Run:ai using the [Tokens](api:main-api:POST/api/v2/token) API. For example:

```bash
curl -X POST \
  # Replace <COMPANY_URL> below with:
  # For self-hosted use the NVIDIA Run:ai user interface URL.
  'https://<COMPANY_URL>/api/v2/token' \
  --header 'Accept: */*' \
  --header 'Content-Type: application/x-www-form-urlencoded' \
  --data-urlencode 'grant_type=client_credentials' \
  --data-urlencode 'client_id=<CLIENT ID>' \
  --data-urlencode 'client_secret=<CLIENT SECRET>'
```

## Step 3: Installing Slinky and the Slurm Cluster

<Note>

**Note**

Check the [Slinky installation documentation](https://github.com/SlinkyProject/slurm-operator/blob/main/docs/installation.md) for newer chart versions. Pin chart versions with `--version` for reproducible deployments. The commands below pull latest by default.

</Note>

1.  Install cert-manager:

    ```bash
    helm install cert-manager oci://quay.io/jetstack/charts/cert-manager \
      --namespace cert-manager --create-namespace \
      --set crds.enabled=true
    ```
2.  Install the `slurm-operator` CRDs, then install the operator:

    ```bash
    helm install slurm-operator-crds oci://ghcr.io/slinkyproject/charts/slurm-operator-crds

    helm install slurm-operator oci://ghcr.io/slinkyproject/charts/slurm-operator \
      --namespace=slinky --create-namespace
    ```
3.  Verify that the operator is running:

    ```bash
    kubectl --namespace=slinky get pods \
      --selector='app.kubernetes.io/instance=slurm-operator'
    ```
4.  Install the Slurm cluster in the project's namespace. This example enables the default worker NodeSet, partition, and SSH login access:

    ```bash
    helm install slurm oci://ghcr.io/slinkyproject/charts/slurm \
      --set-json 'nodesets={"slinky":{}}' \
      --set 'partitions.all.enabled=true' \
      --set 'loginsets.slinky.enabled=true' \
      --set-file "loginsets.slinky.rootSshAuthorizedKeys=${HOME}/.ssh/id_ed25519.pub" \
      --namespace=slurm
    ```

The `nodesets` and `partitions.all.enabled` settings create the default worker NodeSet and partition. Without them, the chart does not create compute nodes.

`rootSshAuthorizedKeys` grants root SSH access to the login pod. Review your organization's access policy before enabling it. Use the path to your own public key where needed.

## Step 4: Viewing Slurm Workloads in NVIDIA Run:ai

After installation, NVIDIA Run:ai shows the controller, login service, REST API, and worker NodeSets in the `slurm` project. Slinky NodeSets are a natively supported NVIDIA Run:ai [workload type](../../workloads-in-nvidia-run-ai/workload-types/supported-workload-types.md). No workload-type registration is required. NVIDIA Run:ai schedules the NodeSet pod through a pod group and enforces the project's GPU quota.

The following example uses the CLI:

```bash
runai workload list --project slurm
```

The output includes workloads similar to the following:

```
Workload             Type         Framework   Status      Project  GPU Alloc.
slurm-controller     StatefulSet  Runai       Running     slurm    0.00
slurm-login-slinky   Deployment   Runai       Running     slurm    0.00
slurm-restapi        Deployment   Runai       Running     slurm    0.00
slurm-worker-slinky  NodeSet      Runai       Running     slurm    0.00
```

You can also view these workloads in **Workload manager → Workloads**. Filter the view by the `slurm` project.

## Step 5: Adding a GPU-Enabled Worker NodeSet

The default Slinky NodeSet has no GPU request. Submit a GPU-enabled NodeSet through the NVIDIA Run:ai user interface, API, or CLI.

Adjust the `slurmd.resources`, `nodeSelector`, and GPU count for your cluster. The `logfile.image` is required for the log-shipping init container. The workload appears as a `NodeSet` with its requested GPU allocation. After a few seconds, it also appears in **Workload manager → Workloads** alongside the default NodeSet.

### Submitting the Workload

<Tabs>

<Tab title="UI">

1. To create a workload, go to Workload manager → Workloads.
2. Click **+ NEW WORKLOAD** and select **Via YAML** from the dropdown.
3. In the YAML submission form, select the cluster where the workload will run.
4.  Upload or paste your YAML manifest. To upload a file, click **UPLOAD YAML FILE** and choose your YAML. To **paste the YAML**, insert it directly into the editor:

    ```yaml
    apiVersion: slinky.slurm.net/v1beta1
    kind: NodeSet
    metadata:
      name: slurm-worker-gpu
      namespace: slurm
    spec:
      controllerRef:
        name: slurm
      replicas: 1
      scalingMode: StatefulSet
      logfile:
        image: docker.io/library/alpine:latest
      slurmd:
        image: ghcr.io/slinkyproject/slurmd:26.05-ubuntu26.04
        resources:
          limits:
            nvidia.com/gpu: "4"
      template:
        spec:
          nodeSelector:
            kubernetes.io/os: linux
    ```
5. Select a **project**. If the `namespace` is not defined in the YAML, select a **project** from the submission form.
6. Click **CREATE WORKLOAD**.

</Tab>

<Tab title="API">

Copy the following command to your terminal. Update these parameters as needed. For more details, see the [Workloads V2](api:main-api:POST/api/v2/workloads) API.

* `<COMPANY-URL>` - The URL of the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 2](deploying-slurm.md#step-2-creating-a-user-access-key).
* `<PROJECT-ID>` - The project ID for the `slurm` project. Retrieve it with the [Get Projects](api:main-api:GET/api/v1/org-unit/projects) API.

```bash
curl -L 'https://<COMPANY-URL>/api/v2/workloads' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "metadata": {
      "name": "slurm-worker-gpu",
      "projectId": "<PROJECT-ID>"
    },
    "manifest": {
      "apiVersion": "slinky.slurm.net/v1beta1",
      "kind": "NodeSet",
      "metadata": {
        "name": "slurm-worker-gpu",
        "namespace": "slurm"
      },
      "spec": {
        "controllerRef": {
          "name": "slurm"
        },
        "replicas": 1,
        "scalingMode": "StatefulSet",
        "logfile": {
          "image": "docker.io/library/alpine:latest"
        },
        "slurmd": {
          "image": "ghcr.io/slinkyproject/slurmd:26.05-ubuntu26.04",
          "resources": {
            "limits": {
              "nvidia.com/gpu": "4"
            }
          }
        },
        "template": {
          "spec": {
            "nodeSelector": {
              "kubernetes.io/os": "linux"
            }
          }
        }
      }
    }
  }'
```

</Tab>

<Tab title="CLI v2">

Follow the below steps. For more details, see the [CLI reference](../../reference/cli/runai/runai_workload_submit.md).

1.  Save the following manifest as `nodeset-gpu.yaml`:

    ```yaml
    apiVersion: slinky.slurm.net/v1beta1
    kind: NodeSet
    metadata:
      name: slurm-worker-gpu
      namespace: slurm
    spec:
      controllerRef:
        name: slurm
      replicas: 1
      scalingMode: StatefulSet
      logfile:
        image: docker.io/library/alpine:latest
      slurmd:
        image: ghcr.io/slinkyproject/slurmd:26.05-ubuntu26.04
        resources:
          limits:
            nvidia.com/gpu: "4"
      template:
        spec:
          nodeSelector:
            kubernetes.io/os: linux
    ```
2.  Submit the workload:

    ```bash
    runai workload submit --file nodeset-gpu.yaml --project slurm
    ```

</Tab>

</Tabs>

### Configuring Slurm GRES

To make GPUs visible to Slurm as GRES, configure `gres.conf` with NVML auto-detection. This lets `sinfo` and `scontrol` report `Gres=gpu:4`, and lets jobs request GPUs with `srun --gres=gpu:N`. `AutoDetect=nvidia` uses NVML to query the NVIDIA drivers on the node.

In Slinky's configless mode, `configFileRefs` mounts the ConfigMap into the controller's `/etc/slurm` directory. The controller serves files in that directory to nodes through the configless protocol. `slurmd` fetches `gres.conf` over the network and caches it locally. It does not read a file mounted into the NodeSet container, so mounting the file there has no effect. Deliver the file through the controller instead:

```bash
kubectl create configmap slurm-gres-conf -n slurm \
  --from-literal=gres.conf='AutoDetect=nvidia'

kubectl patch controller slurm -n slurm --type=merge -p \
  '{"spec":{"configFileRefs":[{"name":"slurm-gres-conf"}],"extraConf":"PartitionName=all Nodes=ALL Default=YES MaxTime=UNLIMITED\nGresTypes=gpu"}}'
```

If the GPU worker pod already exists, force Slurm to detect its GRES again:

```bash
kubectl exec slurm-controller-0 -n slurm -c slurmctld -- \
  scontrol delete nodename=slurm-worker-gpu-0
kubectl delete pod slurm-worker-gpu-0 -n slurm --grace-period=0 --force
```

Apply `configFileRefs` before the GPU NodeSet starts to avoid this recovery step. Dynamic nodes cache GRES during their first registration. A plain pod restart re-registers the same cached empty GRES value.

## Step 6: Verifying GPU Scheduling from Slurm

1.  Connect to the Slurm login pod with the SSH key configured in Step 3. Check the available nodes:

    ```bash
    sinfo -o "%20N %10P %8T %5c %8m %25G"
    ```

    The output should show a GPU GRES entry for the GPU worker and none for the default worker:

    ```
    NODELIST             PARTITION  STATE    CPUS  MEMORY   GRES
    slinky-0              all*       idle     8     31383    (null)
    slurm-worker-gpu-0    all*       idle     8     31719    gpu:nvidia_a10g:1(S:0)
    ```
2.  Confirm that Slurm registered the GPU resource on the GPU worker:

    ```bash
    scontrol show node slurm-worker-gpu-0 | grep Gres
    ```

    ```
    Gres=gpu:nvidia_a10g:1(S:0)
    ```
3.  Request a GPU from the GPU-enabled worker:

    ```bash
    srun --gres=gpu:1 --nodelist=slurm-worker-gpu-0 \
      nvidia-smi --query-gpu=name,uuid --format=csv
    ```

    The command returns the GPU name and UUID. For example:

    ```bash
    name, uuid
    NVIDIA A10G, <GPU-UUID>
    ```
4.  Confirm that Slurm rejects a GPU request against a non-GPU node:

    ```bash
    srun --gres=gpu:1 --nodelist=slinky-0 hostname
    ```

    ```
    srun: error: Unable to allocate resources: Requested node configuration is not available
    ```

## Scope and Limitations

This deployment runs Slurm's control plane and compute nodes as Kubernetes workloads. NVIDIA Run:ai monitors the NodeSet pods at the Kubernetes layer.

* Node assignment is static. Nodes are assigned to Slurm or Kubernetes workloads during deployment or configuration. This deployment does not move nodes dynamically between the pools based on demand.
* NVIDIA Run:ai does not manage the Slurm control plane. It runs and monitors NodeSet pods at the Kubernetes layer. Submit jobs and manage scheduling, partitions, and accounts through Slurm tools such as `sinfo`, `sbatch`, and `sacctmgr`.
* Configure user access through your existing identity stack, not NVIDIA Run:ai or Slinky. The chain is OIDC, LDAP or Active Directory, SSSD on the Slurm login pod, and Slurm accounting. Slinky supports an SSSD sidecar on the login pod. Your IT team supplies and configures the directory, SSO, and SSSD components.
