# Validating GPU Cluster Communication with NCCL Tests

This tutorial demonstrates how to run NCCL `all_reduce_perf_mpi` benchmarks inside PyTorch containers launched as NVIDIA Run:ai training workloads. A single workflow covers single-node, single-rack, and multi-rack validation, so you can verify GPU-to-GPU communication on the cluster before running training, inference, or other production workloads.

<Info>

**Note**

* The example uses the `nvcr.io/nvidia/pytorch:26.01-py3` container, which ships with the NCCL test binaries pre-built under `/usr/local/bin/`. Adjust the tag to match the [CUDA and driver versions](https://docs.nvidia.com/deeplearning/frameworks/pytorch-release-notes/index.html) supported by your cluster.
* The example is configured for GB200 systems. For other hardware, verify compatibility with your target [GPU hardware](https://developer.nvidia.com/cuda/gpus).
* While the walkthrough uses the NVIDIA PyTorch container, you can adapt this workflow to other container images and hardware configurations.

</Info>

In this tutorial, you will learn how to:

* Submit single-node, single-rack, and multi-rack training workloads through the user interface, API, or CLI
* Run an `all_reduce` NCCL benchmark interactively from inside each workload
* Verify pod placement across nodes and racks
* Use additional NCCL collective tests to characterize specific communication patterns

## What NCCL Tests Validate

NCCL (NVIDIA Collective Communications Library) benchmarks are the fastest health check for a GPU cluster. A short [`nccl-tests`](https://github.com/NVIDIA/nccl-tests) run validates:

* **Driver and library install** - CUDA, NCCL, and the MPI launcher work end-to-end inside the container.
* **GPU visibility** - All requested GPUs are passed through to the pod.
* **Intra-node fabric** - NVLink and NVSwitch bandwidth on a single node.
* **Inter-node fabric** - InfiniBand or RoCE bandwidth between nodes, including rack-to-rack hops.
* **Scheduling and topology** - Pods land where the fabric expects.

A failing or low-bandwidth NCCL run is much cheaper to diagnose than burning GPU-hours on a misconfigured stack.

## Configurations Covered

The tutorial covers three tests, scaling from a single node up to multiple racks:

| # | Test        | Workers × GPUs per worker | Validates             |
| - | ----------- | ------------------------- | --------------------- |
| 1 | Single node | 1 × 4 = 4 GPUs            | Intra-node NVLink     |
| 2 | Single rack | 2 × 4 = 8 GPUs            | Inter-node, same rack |
| 3 | Multi-rack  | 3 × 4 = 12 GPUs           | Cross-rack fabric     |

## Prerequisites

Before you start, make sure the following requirements are met:

* Your administrator has:
  * Created a [project](../../platform-management/aiinitiatives/organization/projects.md) with a sufficient GPU quota for the benchmark workloads.
* You have:
  * A container image that bundles NCCL and the [`nccl-tests`](https://github.com/NVIDIA/nccl-tests) binaries. The example uses `nvcr.io/nvidia/pytorch:26.01-py3`, which ships `/usr/local/bin/all_reduce_perf_mpi`.
  * For multi-node tests, an image that supports MPI launch between pods (SSH keys or site PMIx setup).
  * Optional: `kubectl` access for inspecting Kubernetes cluster node labels (including rack-level information, if configured).

<Info>

**Note**

The example image is publicly pullable from NGC without authentication. If you use a container that requires authentication, your administrator must configure a Docker registry [credential](../../workloads-in-nvidia-run-ai/workload-assets/credentials.md#docker-registry) so the cluster can pull the image.

</Info>

## Step 1: Logging In

<Tabs>

<Tab title="UI">

Browse to the provided NVIDIA Run:ai user interface and log in with your credentials.

</Tab>

<Tab title="API">

To use the API, you will need to obtain a token as shown in [Creating a user access key](nccl-tests.md#step-2-creating-a-user-access-key).

</Tab>

<Tab title="CLI v2">

Run the below `--help` command to obtain the login options and log in according to your setup:

```sh
runai login --help
```

</Tab>

</Tabs>

## Step 2: Creating a User Access Key

<Info>

**Note**

This step is only required if you intend to follow the API steps in this tutorial.

</Info>

Access keys are used for API integrations with NVIDIA Run:ai. An access key contains a client ID and a client secret. With the client credentials, you can obtain a token and use it within subsequent API calls.

In the NVIDIA Run:ai user interface:

1. Click the user avatar at the top right corner, then select **Settings**
2. Click **+ACCESS KEY**
3. Enter the access key's **name** and click **CREATE**
4. Copy the **Client ID** and **Client secret** and store securely
5. Click **DONE**

To request an API access token, use the client credentials to get a token to access NVIDIA Run:ai using the [Tokens](https://run-ai-docs.nvidia.com/api/2.26/authentication-and-authorization/tokens) API. For example:

```bash
curl -X POST \ 
  # Replace <COMPANY_URL> below with:
  # For self-hosted use the NVIDIA Run:ai user interface URL.
  'https://<COMPANY_URL>/api/v1/token' \ 
  --header 'Accept: */*' \ 
  --header 'Content-Type: application/json' \ 
  --data-raw '{ 
  "grantType":"client_credentials", 
  "clientId":"<CLIENT ID>", 
  "clientSecret" : "<CLIENT SECRET>" 
}'
```

## Step 3: Identifying Available Nodes and Racks

Before submitting any workloads, identify which nodes are available and which racks they belong to. This is required to design tests that exercise intra-rack and inter-rack communication paths.

1.  SSH to a Kubernetes node and load the cluster module. The node name and module path below are examples; substitute the values for your environment:

    ```bash
    ssh t06-p1-k8s-arm-01
    module load kubernetes/k8s-user/
    ```
2.  List nodes with `kubectl` to see node names and labels. Node names typically encode the rack identifier (for example, `s03` and `s04` indicate different racks):

    ```bash
    kubectl get nodes -o wide
    ```
3.  List nodes and node pools through NVIDIA Run:ai:

    ```bash
    runai list nodes
    runai nodepool list
    ```

## Step 4: Running NCCL Tests on a Single Node

This step runs an `all_reduce` benchmark across four GPUs within a single node, validating intra-node NVLink and PCIe bandwidth.

### Submitting the Workload

Submit a training workload that requests four GPUs on one node and idles the container so you can run benchmarks interactively.

<Tabs>

<Tab title="UI">

1. To create the training workload, go to Workload manager → Workloads.
2. Click **+NEW WORKLOAD** and select **Training** from the dropdown menu.
3. Within the new training form, select the **cluster** and the `nccl-benchmarking` **project**.
4. Set the training workload **architecture** to **Standard**. A standard workload runs a single pod.
5. Select **Start from scratch**.
6. Enter `nccl-single-node` as the **name** for the workload, then click **CONTINUE**.
7. Under **Environment**, set the **Image URL** to `nvcr.io/nvidia/pytorch:26.01-py3`.
8. Under **Runtime settings**, click **+COMMAND & ARGUMENTS**, then set the **Command** to `bash` and the **Arguments** to `-c 'sleep 1d'`. The container idles for one day so you can launch the benchmark interactively.
9. Under **Compute resource**, set **GPU devices per pod** to `4`.
10. Click **CREATE TRAINING**.

</Tab>

<Tab title="API">

Copy the following example request and update the parameters as needed. For more details, see the [Trainings](https://run-ai-docs.nvidia.com/api/2.25/workloads/trainings) API:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 2](nccl-tests.md#step-2-creating-a-user-access-key).
* `<PROJECT-ID>` - The ID of the `nccl-benchmarking` project. Retrieve it via the [Get projects](https://run-ai-docs.nvidia.com/api/2.25/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the cluster. Retrieve it via the [Get clusters](https://run-ai-docs.nvidia.com/api/2.25/organizations/clusters#get-api-v1-clusters) API.

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/trainings' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "name": "nccl-single-node",
    "projectId": "<PROJECT-ID>",
    "clusterId": "<CLUSTER-UUID>",
    "spec": {
      "image": "nvcr.io/nvidia/pytorch:26.01-py3",
      "command": "bash",
      "args": "-c \"sleep 1d\"",
      "compute": {
        "gpuDevicesRequest": 4
      }
    }
  }'
```

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. For more details, see the [CLI reference](../../reference/cli/runai/runai_training_submit.md):

```bash
runai training standard submit nccl-single-node \
  -p nccl-benchmarking \
  -i nvcr.io/nvidia/pytorch:26.01-py3 \
  -g 4 \
  -- bash -c 'sleep 1d'
```

</Tab>

</Tabs>

### Running the Benchmark

The benchmark sweeps message sizes from 8 bytes to 1 GB doubling at each step (`-b 8 -e 1G -f 2`), assigns one GPU per rank (`-g 1`), performs two warmup iterations (`-w 2`), measures ten iterations (`--iters 10`), and runs ten validation checks (`-c 10`).

<Tabs>

<Tab title="UI">

1. Go to Workload manager → Workloads and find `nccl-single-node`.
2. Verify the **Status** column shows **Running**.
3. Click the value in the **Running/requested pods** column to open the pods view, then check the **Allocated GPUs** column to confirm GPU allocation per pod.
4. Back in the Workloads table, click the `nccl-single-node` row, then click **SHOW DETAILS** at the top right of the action bar.
5. Open the **Terminal** tab. Use the **Pod** and **Container** dropdowns to select the running pod and container.
6. In the terminal, run `nvidia-smi` to confirm four GPUs are visible.
7.  Run the four-GPU `all_reduce` benchmark with debug logging enabled:

    ```bash
    mpirun -np 4 \
      --allow-run-as-root \
      -x NCCL_DEBUG=INFO \
      /usr/local/bin/all_reduce_perf_mpi \
      -b 8 -e 1G -f 2 -g 1 -w 2 --iters 10 -c 10
    ```

For more details on the Workloads table, the pods view, and the Terminal tab, see [Workloads](../../workloads-in-nvidia-run-ai/workloads.md).

</Tab>

<Tab title="CLI v2">

The CLI commands below pass `-p nccl-benchmarking` explicitly. Alternatively, set the project as the CLI default once with `runai project set nccl-benchmarking` and omit the flag from subsequent commands.

1.  Confirm the workload is in **Running** state and review GPU allocation. The `exec` step in the next instruction fails if the pod is not yet running:

    ```bash
    runai training standard describe nccl-single-node -p nccl-benchmarking
    ```
2.  Once the workload is **Running**, open an interactive shell into the workload and verify GPU visibility:

    ```bash
    runai training standard exec nccl-single-node -p nccl-benchmarking -it -- bash
    nvidia-smi
    ```
3.  Run the four-GPU `all_reduce` benchmark with debug logging enabled:

    ```bash
    mpirun -np 4 \
      --allow-run-as-root \
      -x NCCL_DEBUG=INFO \
      /usr/local/bin/all_reduce_perf_mpi \
      -b 8 -e 1G -f 2 -g 1 -w 2 --iters 10 -c 10
    ```

</Tab>

</Tabs>

## Step 5: Running NCCL Tests Across Multiple Nodes in the Same Rack

This step runs an eight-GPU `all_reduce` benchmark spanning two nodes in the same rack, validating intra-rack inter-node fabric performance.

<Warning>

**Multi-node MPI requires an SSH server on the workers**

MPI launches ranks by having the master (launcher) pod SSH into the workers. The `nvcr.io/nvidia/pytorch` image does not ship an SSH server, so the worker command below installs `openssh-server` and starts `sshd` on port `2222` before idling. The `mpirun` commands then point MPI at that port with `-mca plm_rsh_args`. If you use a custom image that already bundles and starts `sshd`, you can skip the install step.

</Warning>

### Submitting the Workload

Submit a multi-worker training workload. Two workers each request four GPUs, for a total of eight GPUs across two nodes. Each worker installs and starts `sshd`, then idles so benchmarks can be launched interactively from the master.

<Tabs>

<Tab title="UI">

1. To create the training workload, go to Workload manager → Workloads.
2. Click **+NEW WORKLOAD** and select **Training** from the dropdown menu.
3. Within the new training form, select the **cluster** and the `nccl-benchmarking` **project**.
4. Set the training workload **architecture** to **Distributed**. A distributed workload runs multiple processes that can span across different nodes.
5. Select **Start from scratch**.
6. Set the framework for the distributed workload to **MPI**. If MPI isn't available, see [Distributed training prerequisites](../../installation-and-configuration/installation/install-using-helm/prerequisites-and-features.md#distributed-training) for details on enabling.
7. Enter `nccl-multi-node` as the **name** for the workload, then click **CONTINUE**.
8. Under **Environment**, set the **Image URL** to `nvcr.io/nvidia/pytorch:26.01-py3`.
9. Under **Runtime settings**, click **+COMMAND & ARGUMENTS**, then set the **Command** to `bash` and the **Arguments** to `-c 'apt-get update -qq && apt-get install -y --no-install-recommends openssh-server && mkdir -p /var/run/sshd && /usr/sbin/sshd -p 2222 -o StrictModes=no && sleep 1d'`. This installs and starts an SSH server on port `2222` so the MPI launcher can reach the workers, then idles the container.
10. Set the **number of workers** to `2`. Combined with the master pod, this produces three pods; only the workers run on GPU nodes.
11. Under **Compute resource**, set **GPU devices per pod** to `4`.
12. Click **CONTINUE** to proceed to the master form. The master pod acts as the MPI launcher and should not request GPUs. Leave **GPU devices per pod** unset or set to `0` for the master, then click **CONTINUE**.
13. Click **CREATE TRAINING**.

</Tab>

<Tab title="API">

Copy the following example request and update the parameters as needed. For more details, see the [Distributed](https://run-ai-docs.nvidia.com/api/2.25/workloads/distributed) API:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 2](nccl-tests.md#step-2-creating-a-user-access-key).
* `<PROJECT-ID>` - The ID of the `nccl-benchmarking` project. Retrieve it via the [Get projects](https://run-ai-docs.nvidia.com/api/2.25/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the cluster. Retrieve it via the [Get clusters](https://run-ai-docs.nvidia.com/api/2.25/organizations/clusters#get-api-v1-clusters) API.

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/distributed' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "name": "nccl-multi-node",
    "projectId": "<PROJECT-ID>",
    "clusterId": "<CLUSTER-UUID>",
    "masterSpecSameAsWorker": true,
    "spec": {
      "image": "nvcr.io/nvidia/pytorch:26.01-py3",
      "command": "bash",
      "args": "-c \"apt-get update -qq && apt-get install -y --no-install-recommends openssh-server && mkdir -p /var/run/sshd && /usr/sbin/sshd -p 2222 -o StrictModes=no && sleep 1d\"",
      "compute": {
        "gpuDevicesRequest": 4
      },
      "numWorkers": 2,
      "distributedFramework": "MPI"
    }
  }'
```

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. For more details, see the [CLI reference](../../reference/cli/runai/runai_training_submit.md):

```bash
runai training mpi submit nccl-multi-node \
  -p nccl-benchmarking \
  -i nvcr.io/nvidia/pytorch:26.01-py3 \
  -g 4 \
  --workers 2 \
  --slots-per-worker 4 \
  --large-shm \
  --master-command bash \
  --master-args "-c 'sleep 1d'" \
  -- bash -c 'apt-get update -qq && apt-get install -y --no-install-recommends openssh-server && mkdir -p /var/run/sshd && /usr/sbin/sshd -p 2222 -o StrictModes=no && sleep 1d'
```

The `--large-shm` flag increases the container's shared memory, which some GPUs need for NCCL buffers. It is optional but harmless to include.

</Tab>

</Tabs>

### Running the Benchmark

The benchmark distributes four ranks per node using the hostfile generated by the MPI operator at `/etc/mpi/hostfile`, and uses the same parameter sweep as the single-node test (`-b 8 -e 1G -f 2 -g 1 -w 2 --iters 10 -c 10`). The `-mca plm_rsh_args` flag tells MPI to reach the workers over the SSH server started on port `2222`. Before running the full benchmark, run a quick `hostname` check to confirm MPI connectivity across the workers.

<Tabs>

<Tab title="UI">

1. Go to Workload manager → Workloads and find `nccl-multi-node`.
2. Verify the **Status** column shows **Running**.
3. Click the value in the **Running/requested pods** column to open the pods view, then check the **Allocated GPUs** column to confirm GPU allocation per pod.
4. Back in the Workloads table, click the `nccl-multi-node` row, then click **SHOW DETAILS** at the top right of the action bar.
5. Open the **Terminal** tab. From the **Pod** dropdown, select a pod, then choose its container from the **Container** dropdown.
6. In the terminal, run `nvidia-smi`. A single `nvidia-smi` only reports the 4 GPUs on the pod's own node. To verify all 8 GPUs across the workload, switch the **Pod** dropdown to each of the other pods and rerun `nvidia-smi` in each.
7.  Verify MPI connectivity across the workers by running `hostname` on all eight ranks:

    ```bash
    mpirun -np 8 \
      --hostfile /etc/mpi/hostfile \
      --map-by ppr:4:node \
      --allow-run-as-root \
      -mca plm_rsh_args "-p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
      hostname
    ```
8.  Once connectivity is confirmed, run the eight-GPU `all_reduce` benchmark:

    ```bash
    mpirun -np 8 \
      --hostfile /etc/mpi/hostfile \
      --map-by ppr:4:node \
      --allow-run-as-root \
      -mca plm_rsh_args "-p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
      -x NCCL_DEBUG=INFO \
      /usr/local/bin/all_reduce_perf_mpi \
      -b 8 -e 1G -f 2 -g 1 -w 2 --iters 10 -c 10
    ```

For more details on the Workloads table, the pods view, and the Terminal tab, see [Workloads](../../workloads-in-nvidia-run-ai/workloads.md).

</Tab>

<Tab title="CLI v2">

The CLI commands below pass `-p nccl-benchmarking` explicitly. Alternatively, set the project as the CLI default once with `runai project set nccl-benchmarking` and omit the flag from subsequent commands.

1.  Confirm the workload is in **Running** state and review GPU allocation. The `exec` step in the next instruction fails if the pod is not yet running:

    ```bash
    runai training mpi describe nccl-multi-node -p nccl-benchmarking
    ```
2.  Once the workload is **Running**, open an interactive shell into the workload and verify GPU visibility. `nvidia-smi` only reports the 4 GPUs on the pod's own node. To verify all 8 GPUs across the workload, repeat the `exec` for each pod (use the pod names from the `describe` output above with `--pod`) and rerun `nvidia-smi` in each:

    ```bash
    runai training mpi exec nccl-multi-node -p nccl-benchmarking -it -- bash
    nvidia-smi
    ```
3.  Verify MPI connectivity across the workers by running `hostname` on all eight ranks:

    ```bash
    mpirun -np 8 \
      --hostfile /etc/mpi/hostfile \
      --map-by ppr:4:node \
      --allow-run-as-root \
      -mca plm_rsh_args "-p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
      hostname
    ```
4.  Once connectivity is confirmed, run the eight-GPU `all_reduce` benchmark:

    ```bash
    mpirun -np 8 \
      --hostfile /etc/mpi/hostfile \
      --map-by ppr:4:node \
      --allow-run-as-root \
      -mca plm_rsh_args "-p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
      -x NCCL_DEBUG=INFO \
      /usr/local/bin/all_reduce_perf_mpi \
      -b 8 -e 1G -f 2 -g 1 -w 2 --iters 10 -c 10
    ```

</Tab>

</Tabs>

## Step 6: Running NCCL Tests Across Multiple Racks

This step runs a twelve-GPU `all_reduce` benchmark spanning three nodes placed in different racks, validating inter-rack fabric performance.

<Warning>

**Multi-node MPI requires an SSH server on the workers**

MPI launches ranks by having the master (launcher) pod SSH into the workers. The `nvcr.io/nvidia/pytorch` image does not ship an SSH server, so the worker command below installs `openssh-server` and starts `sshd` on port `2222` before idling. The `mpirun` commands then point MPI at that port with `-mca plm_rsh_args`. If you use a custom image that already bundles and starts `sshd`, you can skip the install step.

</Warning>

### Submitting the Workload

Submit a workload that scales the multi-worker pattern to three workers across three nodes. Each worker installs and starts `sshd`, then idles.

<Tabs>

<Tab title="UI">

1. To create the training workload, go to Workload manager → Workloads.
2. Click **+NEW WORKLOAD** and select **Training** from the dropdown menu.
3. Within the new training form, select the **cluster** and the `nccl-benchmarking` **project**.
4. Set the training workload **architecture** to **Distributed**.
5. Select **Start from scratch**.
6. Set the framework for the distributed workload to **MPI**.
7. Enter `nccl-multi-rack` as the **name** for the workload, then click **CONTINUE**.
8. Under **Environment**, set the **Image URL** to `nvcr.io/nvidia/pytorch:26.01-py3`.
9. Under **Runtime settings**, click **+COMMAND & ARGUMENTS**, then set the **Command** to `bash` and the **Arguments** to `-c 'apt-get update -qq && apt-get install -y --no-install-recommends openssh-server && mkdir -p /var/run/sshd && /usr/sbin/sshd -p 2222 -o StrictModes=no && sleep 1d'`. This installs and starts an SSH server on port `2222` so the MPI launcher can reach the workers, then idles the container.
10. Set the **number of workers** to `3`. To force the workers onto nodes in different racks, set node affinities or labels through your cluster administrator.
11. Under **Compute resource**, set **GPU devices per pod** to `4`.
12. Click **CONTINUE** to proceed to the master form. The master pod acts as the MPI launcher and should not request GPUs. Leave **GPU devices per pod** unset or set to `0` for the master, then click **CONTINUE**.
13. Click **CREATE TRAINING**.

</Tab>

<Tab title="API">

Copy the following example request and update the parameters as needed. For more details, see the [Distributed](https://run-ai-docs.nvidia.com/api/2.25/workloads/distributed) API:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 2](nccl-tests.md#step-2-creating-a-user-access-key).
* `<PROJECT-ID>` - The ID of the `nccl-benchmarking` project. Retrieve it via the [Get projects](https://run-ai-docs.nvidia.com/api/2.25/organizations/projects#get-api-v1-org-unit-projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the cluster. Retrieve it via the [Get clusters](https://run-ai-docs.nvidia.com/api/2.25/organizations/clusters#get-api-v1-clusters) API.

```bash
curl -L 'https://<COMPANY-URL>/api/v1/workloads/distributed' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "name": "nccl-multi-rack",
    "projectId": "<PROJECT-ID>",
    "clusterId": "<CLUSTER-UUID>",
    "masterSpecSameAsWorker": true,
    "spec": {
      "image": "nvcr.io/nvidia/pytorch:26.01-py3",
      "command": "bash",
      "args": "-c \"apt-get update -qq && apt-get install -y --no-install-recommends openssh-server && mkdir -p /var/run/sshd && /usr/sbin/sshd -p 2222 -o StrictModes=no && sleep 1d\"",
      "compute": {
        "gpuDevicesRequest": 4
      },
      "numWorkers": 3,
      "distributedFramework": "MPI"
    }
  }'
```

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. For more details, see the [CLI reference](../../reference/cli/runai/runai_training_submit.md):

```bash
runai training mpi submit nccl-multi-rack \
  -p nccl-benchmarking \
  -i nvcr.io/nvidia/pytorch:26.01-py3 \
  -g 4 \
  --workers 3 \
  --slots-per-worker 4 \
  --large-shm \
  --master-command bash \
  --master-args "-c 'sleep 1d'" \
  -- bash -c 'apt-get update -qq && apt-get install -y --no-install-recommends openssh-server && mkdir -p /var/run/sshd && /usr/sbin/sshd -p 2222 -o StrictModes=no && sleep 1d'
```

</Tab>

</Tabs>

### Running the Benchmark

The benchmark distributes four ranks per node using the hostfile generated by the MPI operator at `/etc/mpi/hostfile`, and uses the same parameter sweep as the previous tests (`-b 8 -e 1G -f 2 -g 1 -w 2 --iters 10 -c 10`). As in Step 5, `-mca plm_rsh_args` reaches the workers over port `2222`, and a `hostname` check confirms MPI connectivity before the full benchmark.

<Tabs>

<Tab title="UI">

1. Go to Workload manager → Workloads and find `nccl-multi-rack`.
2. Verify the **Status** column shows **Running**.
3. Click the value in the **Running/requested pods** column to open the pods view, then check the **Allocated GPUs** column to confirm GPU allocation per pod.
4. Back in the Workloads table, click the `nccl-multi-rack` row, then click **SHOW DETAILS** at the top right of the action bar.
5. Open the **Terminal** tab. From the **Pod** dropdown, select a pod, then choose its container from the **Container** dropdown.
6. In the terminal, run `nvidia-smi`. A single `nvidia-smi` only reports the 4 GPUs on the pod's own node. To verify all 12 GPUs across the workload, switch the **Pod** dropdown to each of the other pods and rerun `nvidia-smi` in each.
7.  Verify MPI connectivity across the workers by running `hostname` on all twelve ranks:

    ```bash
    mpirun -np 12 \
      --hostfile /etc/mpi/hostfile \
      --map-by ppr:4:node \
      --allow-run-as-root \
      -mca plm_rsh_args "-p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
      hostname
    ```
8.  Once connectivity is confirmed, run the twelve-GPU `all_reduce` benchmark:

    ```bash
    mpirun -np 12 \
      --hostfile /etc/mpi/hostfile \
      --map-by ppr:4:node \
      --allow-run-as-root \
      -mca plm_rsh_args "-p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
      -x NCCL_DEBUG=INFO \
      /usr/local/bin/all_reduce_perf_mpi \
      -b 8 -e 1G -f 2 -g 1 -w 2 --iters 10 -c 10
    ```

For more details on the Workloads table, the pods view, and the Terminal tab, see [Workloads](../../workloads-in-nvidia-run-ai/workloads.md).

</Tab>

<Tab title="CLI v2">

The CLI commands below pass `-p nccl-benchmarking` explicitly. Alternatively, set the project as the CLI default once with `runai project set nccl-benchmarking` and omit the flag from subsequent commands.

1.  Confirm the workload is in **Running** state and review GPU allocation. The `exec` step in the next instruction fails if the pod is not yet running:

    ```bash
    runai training mpi describe nccl-multi-rack -p nccl-benchmarking
    ```
2.  Once the workload is **Running**, open an interactive shell into the workload and verify GPU visibility. `nvidia-smi` only reports the 4 GPUs on the pod's own node. To verify all 12 GPUs across the workload, repeat the `exec` for each pod (use the pod names from the `describe` output above with `--pod`) and rerun `nvidia-smi` in each:

    ```bash
    runai training mpi exec nccl-multi-rack -p nccl-benchmarking -it -- bash
    nvidia-smi
    ```
3.  Verify MPI connectivity across the workers by running `hostname` on all twelve ranks:

    ```bash
    mpirun -np 12 \
      --hostfile /etc/mpi/hostfile \
      --map-by ppr:4:node \
      --allow-run-as-root \
      -mca plm_rsh_args "-p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
      hostname
    ```
4.  Once connectivity is confirmed, run the twelve-GPU `all_reduce` benchmark:

    ```bash
    mpirun -np 12 \
      --hostfile /etc/mpi/hostfile \
      --map-by ppr:4:node \
      --allow-run-as-root \
      -mca plm_rsh_args "-p 2222 -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
      -x NCCL_DEBUG=INFO \
      /usr/local/bin/all_reduce_perf_mpi \
      -b 8 -e 1G -f 2 -g 1 -w 2 --iters 10 -c 10
    ```

</Tab>

</Tabs>

## Step 7: Running Other NCCL Collective Tests

The previous steps use `all_reduce_perf_mpi`, but the NCCL test suite includes additional binaries that exercise different collective operations. Substitute the binary in the `mpirun` command above to characterize the collective most relevant to your workload:

| Binary                                | Typical workload pattern                                    |
| ------------------------------------- | ----------------------------------------------------------- |
| `all_reduce_perf_mpi`                 | Default; bandwidth and latency for gradient synchronization |
| `all_gather_perf_mpi`                 | Tensor-parallel forward path or activations gather          |
| `reduce_scatter_perf_mpi`             | ZeRO or sharded-optimizer step                              |
| `broadcast_perf_mpi`                  | Weight broadcast from rank 0                                |
| `reduce_perf_mpi`                     | One-to-all reduction                                        |
| `alltoall_perf_mpi`                   | Mixture-of-Experts dispatch or sequence-parallel            |
| `sendrecv_perf_mpi`                   | Pipeline-parallel hop                                       |
| `gather_perf_mpi`, `scatter_perf_mpi` | Asymmetric collectives                                      |

## Step 8: Cleaning Up the Environment

The benchmark containers idle for one day and are reclaimed automatically when the sleep expires. To release GPUs sooner, delete the workloads manually.
