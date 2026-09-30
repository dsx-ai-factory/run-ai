# Distributed Training with PyTorch

This tutorial demonstrates how to run a distributed training workload with PyTorch on the NVIDIA Run:ai platform. Distributed training enables you to scale model training across multiple GPUs and nodes, improving performance and reducing training time. The example configurations provided here can be adapted to fit your own models, datasets, and training workflows.

<Note>

**Note**

* Before running the example, verify the supported[ CUDA version of PyTorch](https://github.com/pytorch/pytorch/blob/main/RELEASE.md#release-compatibility-matrix) is compatible with your target [GPU hardware](https://developer.nvidia.com/cuda/gpus).
* While the walkthrough uses PyTorch as the framework, the same principles apply to other distributed training [frameworks](../../installation-and-configuration/installation/guided-installation/prerequisites-and-features.md#distributed-training) supported by NVIDIA Run:ai.

</Note>

In this tutorial, you will learn how to:

* Build a custom PyTorch container image
* Create a user access key for API integrations with NVIDIA Run:ai
* Create a persistent data source for storing checkpoints
* Submit and run a distributed training workload through the NVIDIA Run:ai user interface, CLI, or API
* Monitor workload progress and retrieve generated checkpoints

## Prerequisites

* Before you start, make sure you have created a [project](../../platform-management/aiinitiatives/organization/projects.md) or have one created for you.
* To run a distributed PyTorch training workload, you must first build a custom Docker container. This allows you to package the required code into a container that can be run and shared for future workloads.
  * The Docker runtime must be installed on a local machine with the **same CPU architecture** as the target hosts. For example, if the hosts use AMD64 CPUs, build the container on an AMD64 machine. If the hosts use ARM CPUs, build it on an ARM-based machine. Follow the [Docker Engine installation guide](https://docs.docker.com/engine/install/) to install Docker locally.
  * To store and pull the image from your own private Docker registry, create your own Docker registry credential as detailed in [Step 4](distributed-pytorch.md#step-4-creating-a-user-credential).

## Step 1: Creating a Custom Docker Container

1.  On your local machine where Docker is installed, create and navigate to a directory to save the Dockerfile, such as `pytorch-distributed`:

    ```bash
    $ mkdir -p pytorch-distributed
    $ cd pytorch-distributed
    ```
2.  In the new directory, open a new file named `run.sh` and copy the following contents to the file. This is a very simple script that uses `torchrun` to launch a distributed training workload and copies the generated checkpoint to the `/checkpoints` directory inside the container so it can be used again later:

    ```bash
    #!/bin/bash

    torchrun multinode.py --batch_size 32 100 25
    # Arguments: batch size = 32, epochs = 100, checkpoint interval = 25
    cp snapshot.pt /checkpoints
    ```
3. Save and close the file.
4.  Open another new file named `Dockerfile` and copy the following contents to the file. This Dockerfile uses the 24.07 PyTorch container hosted on NGC as a base, clones the official PyTorch examples repository inside the container, and copies the `run.sh` file created previously into the container.

    ```dockerfile
    FROM nvcr.io/nvidia/pytorch:24.07-py3

    WORKDIR /runai-distributed
    RUN git clone https://github.com/pytorch/examples

    WORKDIR /runai-distributed/examples/distributed/ddp-tutorial-series
    COPY run.sh .
    ```
5. Save and close the file.
6.  Once both files have been saved locally, build a container with the following command, replacing \<your-reg> with the name of your private or public registry. This will build the custom container locally:

    ```bash
    docker build -t <your-reg>/pytorch-ddp-example:24.07-py3 .
    ```

<Note>

**Note**

If you are building the container on a machine with a different CPU architecture than the target cluster (for example, AMD64), use the following command:

```bash
docker buildx build --platform linux/amd64 -t <your-image>
```

</Note>

7.  Once the build has finished, push the image to your private or public registry with:

    ```bash
    docker push <your-reg>/pytorch-ddp-example:24.07-py3
    ```

The custom container will be available in your registry and can be used immediately for workloads.

## Step 2: Logging In

<Tabs>

<Tab title="UI">

Browse to the provided NVIDIA Run:ai user interface and log in with your credentials.

</Tab>

<Tab title="API">

To use the API, you will need to obtain a token as shown in [Creating a user access key](distributed-pytorch.md#step-3-creating-a-user-access-key).

</Tab>

<Tab title="CLI v2">

Run the below --help command to obtain the login options and log in according to your setup:

```sh
runai login --help
```

</Tab>

</Tabs>

## Step 3: Creating a User Access Key

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

## Step 4: Creating a User Credential

<Note>

**Note**

* Creating Docker registry user credentials is supported via UI only.
* Pulling images from a private Docker registry using a user credential is supported in the UI (Flexible submission) and API.

</Note>

User credentials allow users to securely store private authentication secrets, which are accessible only to the user who created them. See [User credentials](../../settings/user-settings/user-credentials.md) for more details.

In the NVIDIA Run:ai user interface:

1. Click the user avatar at the top right corner, then select **Settings**
2. Click **+CREDENTIAL** and select **Docker registry** from the dropdown
3. Enter a **name** for the credential. The name must be unique.
4. Optional: Provide a description of the credential
5. Enter the **username**, **password**, and **Docker registry URL**
6. Click **CREATE CREDENTIAL**

## Step 5: Creating a PVC Data Source

To make it easier to reuse code and checkpoints in future workloads, create a data source in the form of a Persistent Volume Claim (PVC). The PVC can be mounted to workloads and will persist after the workload completes, allowing any data it contains to be reused.

<Note>

**Note**

The first time a workload is launched using a new PVC, it will take longer to start as the storage gets provisioned only once the first claim to the PVC is made.

</Note>

<Tabs>

<Tab title="UI">

1. To create a PVC, go to Workload manager → Data sources.
2. Click +**NEW DATA SOURCE** and select **PVC** from the dropdown menu.
3. Within the new form, set the [scope](../../workloads-in-nvidia-run-ai/workload-assets/overview.md#asset-scope).

<Note>

**Important**

PVC data sources created at the cluster or department level do not replicate data across projects or namespaces. Each project or namespace will be provisioned as a separate PVC replica with different underlying PVs; therefore, the data in each PVC is not replicated.

</Note>

4. Enter a **name** for the data source. The name must be unique.
5. For the data options, select **New PVC** and the **storage class** that suits your needs:
   * To allow all nodes to read and write from/to the PVC, select **Read-write by many nodes** for the **access mode**.
   * Enter `10 TB` for the **claim size** to ensure there is capacity for future workloads.
   * Select **Filesystem** (default) as the volume mode. The volume will be mounted as a filesystem, enabling the usage of directories and files.
   * Set the **Container path** to `/checkpoints` which is where the PVC will be mounted inside containers.
6. Click **CREATE DATA SOURCE**

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the following parameters:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 3](distributed-pytorch.md#step-3-creating-a-user-access-key).

For all other parameters within the JSON body, refer to the [PVC](api:main-api:POST/api/v1/asset/datasource/pvc) API.

After creating the data source, wait for the PVC to be provisioned. Use the [List PVC asset](api:main-api:GET/api/v1/asset/datasource/pvc) API to retrieve the claim name. This claim name is the exact value that will be used for the `<pvc-claim-name>` when submitting the workload.

```shellscript
curl -L 'https://<COMPANY-URL>/api/v1/asset/datasource/pvc' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "meta": {
      "name": "<pvc-name>",
      "scope": "<scope>"
    },
    "spec": {
      "path": "/checkpoints",
      "existingPvc": false,
      "claimInfo": {
        "size": "10TB",
        "storageClass": "<my-storage-class>",
        "accessModes": {
          "readWriteMany": true
        },
        "volumeMode": "Filesystem"
        }
     }
  }'
```

</Tab>

</Tabs>

## Step 6: Creating the Workload

<Note>

**Note**

Flexible workload submission is enabled by default. If unavailable, contact your administrator to enable it under **General settings** → Workloads → Flexible workload submission.

</Note>

### How the Configuration Works

* Distributed workload configuration: Workers & master - Selected to enable coordination between pods during multi-node distributed training. This configuration is required for workloads that perform collective communication operations, such as `all_reduce`, to synchronize gradients across processes. In this example, the master pod participates in training rather than acting as a coordination-only process.
* Number of workers: 1 - Configured to create two pods (1 master and 1 worker) running across two nodes. Since the master pod participates in training, the number of workers is set to one fewer than the total number of nodes.
* GPU devices per pod: 1 - Each pod requests a single GPU, matching the example cluster setup where each node provides one GPU.
* Runtime command `torchrun` - Used to launch the distributed training job. `torchrun` starts one process per allocated GPU and assigns each process a unique rank. The PyTorch training operator, managed by NVIDIA Run:ai, automatically sets environment variables such as `RANK`, `LOCAL_RANK`, and `WORLD_SIZE` based on the total number of GPUs allocated to the workload.
* Training parameters - The training script runs with a batch size of 32 for 100 total epochs and saves a checkpoint every 25 epochs. These values are chosen to demonstrate periodic checkpointing during a long-running training job.
* PVC mounted at `/checkpoints` - A PVC is mounted at `/checkpoints` to persist training checkpoints beyond the lifetime of the workload. This enables reuse of checkpoints in future runs and continuation of training with modified hyperparameters.

### Submitting the Workload

<Tabs>

<Tab title="UI - Flexible">

1. To create the training workload, go to Workload manager → Workloads.
2. Click **+NEW WORKLOAD** and select **Training** from the dropdown menu**.**
3. Within the new training form, select the **cluster** and **project**.
4. Set the training workload **architecture** to **Distributed** workload**.** This runs multiple processes that can span across different nodes. Distributed workloads are supported only in environments where distributed training is enabled.
5. Select **Start from scratch** to launch a new workload quickly.
6. Set the framework for the distributed workload to **PyTorch**. If PyTorch isn’t available, see [Distributed training prerequisites](../../installation-and-configuration/installation/guided-installation/prerequisites-and-features.md#distributed-training) for details on enabling.
7. Set the **distributed workload configuration**, which defines how distributed training workloads are divided across multiple machines or processes. For this example, select **Workers & master**, which enables coordination between pods during distributed training.
   * Choose **Workers & master** or **Workers only** based on training requirements and infrastructure.
   * A **master pod** is typically required for multi-node training workloads that must coordinate across nodes, such as workloads that perform `all_reduce` operations.
   * The master pod can either participate in training like a worker or act as a lightweight coordination process.
   * If coordination is not required, select **Workers only**.
8. Enter a **name** for the distributed training workload. If the name already exists in the project, you will be requested to submit a different name.
9. Click **CONTINUE**
10. Under **Environment**, set the **Image URL** by entering the image tag specified during the container build in the [Creating a Custom Docker Container](distributed-pytorch.md#step-1-creating-a-custom-docker-container) section above (`<your-reg>/pytorch-ddp-example:24.07-py3`).
11. Set the **image pull policy**:
    * Set the **condition for pulling the image**. It is recommended to select **Pull the image only if it's not already present on the host**. If you are pushing new containers to your registry with the same tag, select **Always pull the image from the registry** to check if there are updates to the image.
    * Set the **credentials for pulling the image** from your private Docker registry. Select **My credentials** as the source and **Docker registry** as the type, then choose the credential created in [Step 4](distributed-pytorch.md#step-4-creating-a-user-credential) from the **Credential name** dropdown.
12. The **Runtime settings** define the working directory and command executed when the container starts. This example launches the `multinode.py` script using `torchrun`, which runs a multi-process application where each process has a unique rank. The PyTorch training operator coordinates with `torchrun` to automatically set environment variables such as `RANK`, `LOCAL_RANK`, and `WORLD_SIZE` based on the total number of GPUs allocated to the workload.
    * Click **+COMMAND & ARGUMENTS**
    * In the **Command** field, enter `bash run.sh`. This will run the script on all allocated GPUs with a batch size of 32 for 100 total epochs and save a checkpoint every 25 epochs. The final checkpoint will be saved to the `/checkpoints` directory.
    * Set the **container working directory** to `/runai-distributed/examples/distributed/ddp-tutorial-series`. This directory contains the training scripts copied into the container during the image build and is the path the pod opens to when it starts.
13. Set the **GPU devices per pod** and **number of workers**:
    * Set **GPU devices per pod** to **1**. In this example, the workload runs on **two nodes**, each with **one GPU**.
    * Set **1 worker**, resulting in **two pods** (1 master + 1 worker) running across **two nodes**. Since the master participates in training, specify one fewer worker than the total number of nodes.
14. In the **Data sources** form, click the **load** icon. A side pane appears, displaying a list of available data sources. Select the PVC that was created in the [Step 5](distributed-pytorch.md#step-5-creating-a-pvc-data-source).
15. Click **CONTINUE**
16. Ensure the **Allow different setup for the master** toggle is **disabled**. Although the master pod can be configured differently from worker pods, this example uses the same configuration for both.
17. Click **CREATE TRAINING**

</Tab>

<Tab title="UI - Original">

1. To create the training workload, go to Workload manager → Workloads.
2. Click **+NEW WORKLOAD** and select **Training** from the dropdown menu**.**
3. Within the new training form, select the **cluster** and **project**.
4. Set the training workload **architecture** to **Distributed** workload**.** This runs multiple processes that can span across different nodes. Distributed workloads are supported only in environments where distributed training is enabled.
5. Set the framework for the distributed workload to **PyTorch**. If PyTorch isn’t available, see [Distributed training prerequisites](../../installation-and-configuration/installation/guided-installation/prerequisites-and-features.md#distributed-training) for details on enabling.
6. Set the **distributed workload configuration**, which defines how distributed training workloads are divided across multiple machines or processes. For this example, select **Workers & master**, which enables coordination between pods during distributed training.
   * Choose **Workers & master** or **Workers only** based on training requirements and infrastructure.
   * A **master pod** is typically required for multi-node training workloads that must coordinate across nodes, such as workloads that perform `all_reduce` operations.
   * The master pod can either participate in training like a worker or act as a lightweight coordination process.
   * If coordination is not required, select **Workers only**.
7. Select **Start from scratch** to launch a new workload quickly.
8. Enter a **name** for the distributed training workload. If the name already exists in the project, you will be requested to submit a different name.
9. Click **CONTINUE**
10. Click **+ NEW ENVIRONMENT**

    * Enter a **name**
    * Under **Image URL**, enter the image tag specified during the container build in the [Creating a Custom Docker Container](distributed-pytorch.md#step-1-creating-a-custom-docker-container) section above,(`<your-reg>/pytorch-ddp-example:24.07-py3`).
    * Set the **image pull policy.** It is recommended to select **Pull the image only if it's not already present on the host**. If you are pushing new containers to your registry with the same tag, select **Always pull the image from the registry** to check if there are updates to the image.
    * The **Runtime settings** define the working directory and command executed when the container starts. This example launches the `multinode.py` script using `torchrun`, which runs a multi-process application where each process has a unique rank. The PyTorch training operator coordinates with `torchrun` to automatically set environment variables such as `RANK`, `LOCAL_RANK`, and `WORLD_SIZE` based on the total number of GPUs allocated:
      * Click **+COMMAND & ARGUMENTS**
      * In the **Command** field, enter `bash run.sh`. This will run the script on all allocated GPUs with a batch size of 32 for 100 total epochs and save a checkpoint every 25 epochs. The final checkpoint will be saved to the `/checkpoints` directory.
      * Set the **container working directory** to `/runai-distributed/examples/distributed/ddp-tutorial-series`.\
        This directory contains the training scripts copied into the container during the image build and is the path the pod opens to when it starts.
    * Click **CREATE ENVIRONMENT**

    The newly created environment will be selected automatically
11. Set the **number of workers** to **1**, resulting in **two pods** (1 master + 1 worker) running across **two nodes**. Since the master participates in training, specify one fewer worker than the total number of nodes.
12. Select the **‘one-gpu’** compute resource for your workload:
    *   If ‘one-gpu’ is not displayed in the gallery, follow the below steps:

        * Click **+NEW COMPUTE RESOURCE**
        * Enter one-gpu as the **name** for the compute resource. The name must be unique
        * Set **GPU devices** per pod - 1
        * Click **CREATE COMPUTE RESOURCE**

        The newly created compute resource will be selected automatically
13. In the **Data sources** form, select the PVC that was created in [Step 5](distributed-pytorch.md#step-5-creating-a-pvc-data-source).
14. Click **CONTINUE**
15. Ensure the **Allow different setup for the master** toggle is **disabled**. Although the master pod can be configured differently from worker pods, this example uses the same configuration for both.
16. Click **CREATE TRAINING**

</Tab>

<Tab title="API">

Copy the following example request and update the parameters as needed. For more details, see [Distributed](api:main-api:POST/api/v1/workloads/distributed) API:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 3](distributed-pytorch.md#step-3-creating-a-user-access-key).
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](api:main-api:GET/api/v1/org-unit/projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](api:main-api:GET/api/v1/clusters) API.
* `<dockerreg-name>` - The name of the user credential created in [Step 4](distributed-pytorch.md#step-4-creating-a-user-credential). Replace this with the name of the credential preceded by the system prefix `dockerreg-`.
* `<pvc-claim-name>` - The claim name associated with the PVC created in [Step 5](distributed-pytorch.md#step-5-creating-a-pvc-data-source).

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/distributed' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{  
    "name": "<workload-name>",  
    "projectId": "<PROJECT-ID>",
    "clusterId": "<CLUSTER-UUID>",
    "masterSpecSameAsWorker": true,  
    "spec": {
        "command": "bash run.sh",  
        "compute": { 
            "gpuDevicesRequest": 1
        },
        "image": "<your-reg>/pytorch-ddp-example:24.07-py3",
        "imagePullPolicy": "IfNotPresent",
        "imagePullSecrets": [
          {
            "name": "<dockerreg-name>",
            "userCredential": true,
           }
        ],
        "numWorkers": 1, 
        "distributedFramework": "PyTorch",
        "storage": {
            "pvc": [
                {
                    "claimName": "<pvc-claim-name>",
                    "dataSharing": false,
                    "existingPvc": true,
                    "path": "/checkpoints",
                    "readOnly": false
                }
            ]
        },
        "workingDir": "/runai-distributed/examples/distributed/ddp-tutorial-series" 
    } 
}
```

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. Make sure to update the below with the name of your project and workload. For more details, see [CLI reference](../../reference/cli/runai/runai_training_pytorch_submit.md):

```bash
runai project set "project-name"
runai training pytorch submit <workload-name> \
-i <your-reg>/pytorch-ddp-example:24.07-py3 --workers 1 -g 1 \
--existing-pvc claimname=<pvc-claim-name>,path=/checkpoints \
--working-dir /runai-distributed/examples/distributed/ddp-tutorial-series \
--command -- bash run.sh
```

</Tab>

</Tabs>

## Step 7: Monitoring

After the training workload is created, it is added to the [Workloads](../../workloads-in-nvidia-run-ai/workloads.md) table, where it can be viewed, monitored, and managed throughout its lifecycle.

## Step 8: Getting the Checkpoint

At the end of the `run.sh` script, the latest generated checkpoint is copied to the PVC attached to the workload. Any workload that mounts this same PVC can load the checkpoint from `/checkpoints/snapshot.pt`. The PVC can also be used to persist additional data written to the specified filesystem path. This enables checkpoint reuse across workloads, allowing long-running training jobs to resume from a saved state or continue training with different hyperparameters.

## Step 9: Cleaning up the Environment

After the workload finishes, it can be deleted to free up resources for other workloads. To reclaim the disk space used by the PVC, delete the PVC once it is no longer needed.
