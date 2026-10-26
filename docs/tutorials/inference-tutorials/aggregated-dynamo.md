# NVIDIA Dynamo Aggregated Inference Deployment

This tutorial demonstrates how to run a single-node, aggregated inference workload using the RedHatAI/Llama-3.3-70B-Instruct-FP8-dynamic model with [NVIDIA Dynamo](https://docs.nvidia.com/dynamo/getting-started/quickstart) and vLLM on the NVIDIA Run:ai platform. You can use this workflow as a reference and adapt it for your own models, container images, and hardware configurations.

<Note>

**Note**

Verify the specific version of Dynamo is supported on your target hardware according to the [Dynamo support matrix](https://docs.nvidia.com/dynamo/dev/resources/support-matrix).

</Note>

In this tutorial, you will learn how to:

* Set up environment prerequisites for Dynamo-based inference workload
* Create a user access key for API integrations with NVIDIA Run:ai
* Submit and run a single-node inference workload through the NVIDIA Run:ai user interface, API or CLI
* Access the inference endpoint to send requests

## Prerequisites

### Cluster Prerequisites

Ensure that the NVIDIA Dynamo platform is installed in the cluster. This includes the required Custom Resource Definitions (CRDs) and Dynamo platform components.

The following steps must be performed by your administrator.

1.  Set environment variables:

    ```bash
    export NAMESPACE=dynamo-system
    export RELEASE_VERSION=1.5.0
    ```
2.  Install the Dynamo platform:

    ```bash
    helm fetch https://helm.ngc.nvidia.com/nvidia/ai-dynamo/charts/dynamo-platform-${RELEASE_VERSION}.tgz
    helm install dynamo-platform dynamo-platform-${RELEASE_VERSION}.tgz \
      --namespace ${NAMESPACE} \
      --create-namespace \
      --set "grove.enabled=true"
    ```

### NVIDIA Run:ai Prerequisites

Before you start, make sure the following requirements are met:

* Your administrator has:
  * Created a [project](../../platform-management/aiinitiatives/organization/projects.md) for you.
* You have:
  * Access to the model if you are using a gated Hugging Face model that requires an `HF_TOKEN`. Generate a token from your [Hugging Face account](https://huggingface.co/) and provide it to the workload as an environment variable. The model used in this tutorial is not gated and does not require authentication.

## Step 1: Logging In

<Tabs>

<Tab title="UI">

Browse to the provided NVIDIA Run:ai user interface and log in with your credentials.

</Tab>

<Tab title="API">

To use the API, you will need to obtain a token as shown in [Creating a user access key](aggregated-dynamo.md#step-2-creating-a-user-access-key).

</Tab>

<Tab title="CLI v2">

Run the below --help command to obtain the login options and log in according to your setup:

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

## Step 3: Creating a PVC Data Source

To reduce cold-start latency and cache model weights across runs, create a data source in the form of a Persistent Volume Claim (PVC). The PVC can be mounted to workloads and will persist after the workload completes, allowing any data it contains to be reused.

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
   * Enter `2 TB` for the **claim size** to ensure we have plenty of capacity for future workloads
   * Select **Filesystem** (default) as the volume mode. The volume will be mounted as a filesystem, enabling the usage of directories and files.
   * Set the **Container path** to `/opt/models` which is where the PVC will be mounted inside containers.
6. Click **CREATE DATA SOURCE**

After creating the data source, wait for the PVC to be provisioned. The PVC claim name (which is displayed in the UI as the Kubernetes name) will appear in the Data sources grid once it’s ready. This claim name is the exact value that will be used for the `<pvc-claim-name>` when submitting the workload.

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the following parameters:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 2](aggregated-dynamo.md#step-2-creating-a-user-access-key).

For all other parameters within the JSON body, refer to the [PVC](api:main-api:POST/api/v1/asset/datasource/pvc) API.

After creating the data source, wait for the PVC to be provisioned. Use the [List PVC asset](api:main-api:GET/api/v1/asset/datasource/pvc) API to retrieve the claim name. This claim name is the exact value that will be used for the `<pvc-claim-name>` when submitting the workload.

```bash
curl -L 'https://<COMPANY-URL>/api/v1/asset/datasource/pvc' \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <TOKEN>' \
  -d '{
    "meta": {
      "name": "<pvc-name>",
      "scope": "<scope>"
    },
    "spec": {
      "path": "/opt/models",
      "existingPvc": false,
      "claimInfo": {
        "size": "2TB",
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

## Step 4: Creating the Workload

The configuration below submits a DynamoGraphDeployment that defines a single-node, aggregated inference workload, where request handling, prefill, and decode are executed within a single Dynamo service on one node.

### How the Configuration Works

* Defines a Dynamo-based inference workload that is registered and managed by NVIDIA Run:ai.
* Uses vLLM as the backend framework for model execution.
* Deploys a frontend service that exposes an HTTP endpoint for inference requests.
* Deploys a single worker replica (`VllmPrefillWorker`) that executes model inference on GPUs.
* Configures the workload as a single-node, aggregated deployment, where all inference runs within one worker replica on one node.
* The worker requests 4 GPUs, demonstrating how to run an aggregated inference workload for a large model on a single node.

For more details, see the [NVIDIA Dynamo documentation](https://docs.nvidia.com/dynamo/kubernetes-deployment/deployment-guide).

### Submitting the Workload

<Tabs>

<Tab title="UI">

1. To create a workload, go to Workload manager → Workloads.
2. Click **+ NEW WORKLOAD** and select **Via YAML** from the dropdown.
3. In the YAML submission form, select the **cluster** where the workload will run.
4. Upload or paste your YAML manifest. To upload a file, click **UPLOAD YAML FILE** and choose your YAML. To paste the YAML, insert it directly into the editor. Before submitting, update the following fields in the manifest:
   * `<pvc-claim-name>` - The claim name associated with the PVC created in [Step 3](aggregated-dynamo.md#step-3-creating-a-pvc-data-source).
   * `<your-hf-token>` - Required only for gated Hugging Face models. Generate a token from your [Hugging Face account](https://huggingface.co/) and provide it to the workload as an environment variable.
   * `<INGRESS-CLASS>` - The ingress controller class name configured in your cluster (for example, `haproxy`). Contact your administrator to obtain the correct value.
   * `<DEPLOYMENT-PREFIX>` - A unique prefix for the deployment endpoint (for example, `llama3-agg`).
   * `<INFERENCE-DOMAIN>` - The suffix including the cluster [FQDN](../../installation-and-configuration/system-requirements/networking-setup.md#fqdn-and-dns). Combined with the prefix, this forms the full hostname for the workload endpoint (for example, `inference.mycorp.com` if your DNS wildcard covers `*.inference.mycorp.com`). You can retrieve the cluster domain via the [Get Clusters Minimal](api:main-api:GET/api/v1/clusters/minimal) API - use the `domain` field in the response.

```yaml
apiVersion: nvidia.com/v1alpha1
kind: DynamoGraphDeployment
metadata:
  name: llama3-70b-agg
spec:
  backendFramework: vllm
  pvcs:
    - name: <pvc-claim-name>
      create: false
  services:
    Frontend:
      componentType: frontend
      volumeMounts:
        - name: <pvc-claim-name>
          mountPoint: /opt/models
      extraPodSpec:
        mainContainer:
          image: nvcr.io/nvidia/ai-dynamo/vllm-runtime:1.5.0
          workingDir: /workspace/examples/backends/vllm
      envs:
        - name: HF_HOME
          value: /opt/models
        - name: HF_TOKEN
          value: <your-hf-token>
      replicas: 1
      ingress:
        enabled: true
        ingressControllerClassName: <INGRESS-CLASS>
        host: <DEPLOYMENT-PREFIX>
        hostSuffix: <INFERENCE-DOMAIN>
    VllmPrefillWorker:
      componentType: worker
      volumeMounts:
        - name: <pvc-claim-name>
          mountPoint: /opt/models
      sharedMemory:
        size: 20Gi
      extraPodSpec:
        mainContainer:
          env:
            - name: SERVED_MODEL_NAME
              value: "RedHatAI/Llama-3.3-70B-Instruct-FP8-dynamic"
            - name: MODEL_PATH
              value: "/opt/models/hub/models--RedHatAI--Llama-3.3-70B-Instruct-FP8-dynamic/snapshots/ddb4128556dfcff99e0c41aee159ea6c3e655dcd"
            - name: HF_HOME
              value: /opt/models
            - name: HF_TOKEN
              value: <your-hf-token>
          args:
          - "python3 -m dynamo.vllm --model $MODEL_PATH --served-model-name $SERVED_MODEL_NAME --tensor-parallel-size 4 --data-parallel-size 1 --disable-log-requests --gpu-memory-utilization 0.90 --no-enable-prefix-caching --block-size 128"
          command:
          - /bin/sh
          - -c
          image: nvcr.io/nvidia/ai-dynamo/vllm-runtime:1.5.0
          workingDir: /workspace/examples/backends/vllm
      replicas: 1
      resources:
        limits:
          gpu: "4"
        requests:
          gpu: "4"
```

5. Select a **project**. If the `namespace` is not defined in the YAML, select a **project** from the submission form.
6. Set **Multi-Node NVLink (MNNVL) acceleration** to **Not required**. The workload will not use MNNVL acceleration, even if MNNVL-capable nodes are available.
7. Click **CREATE WORKLOAD**

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the following parameters. For more details, see [Workload V2](api:main-api:POST/api/v2/workloads) API:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 2](aggregated-dynamo.md#step-2-creating-a-user-access-key).
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](api:main-api:GET/api/v1/org-unit/projects) API.
* `<pvc-claim-name>` - The claim name associated with the PVC created in [Step 3](aggregated-dynamo.md#step-3-creating-a-pvc-data-source).
* `<your-hf-token>` - Required only for gated Hugging Face models. Generate a token from your [Hugging Face account](https://huggingface.co/) and provide it to the workload as an environment variable.
* `<INGRESS-CLASS>` - The ingress controller class name configured in your cluster (for example, `haproxy`). Contact your administrator to obtain the correct value.
* `<DEPLOYMENT-PREFIX>` - A unique prefix for the deployment endpoint (for example, `llama3-agg`).
* `<INFERENCE-DOMAIN>` - The suffix including the cluster [FQDN](../../installation-and-configuration/system-requirements/networking-setup.md#fqdn-and-dns). Combined with the prefix, this forms the full hostname for the workload endpoint (for example, `inference.mycorp.com` if your DNS wildcard covers `*.inference.mycorp.com`). You can retrieve the cluster domain via the [Get Clusters Minimal](api:main-api:GET/api/v1/clusters/minimal) API - use the `domain` field in the response.

```bash
curl -L 'https://<COMPANY-URL>/api/v2/workloads' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{
    "metadata": {
        "name": "llama3-agg",
        "projectId": "<PROJECT-ID>"
    },
    "manifest": {
        "apiVersion": "nvidia.com/v1alpha1",
        "kind": "DynamoGraphDeployment",
        "metadata": {
            "name": "llama3-70b-agg"
        },
        "spec": {
            "backendFramework": "vllm",
            "pvcs": [
                {
                    "name": "<pvc-claim-name>",
                    "create": false
                }
            ],
            "services": {
                "Frontend": {
                    "componentType": "frontend",
                    "dynamoNamespace": "llama3-70b-agg",
                    "volumeMounts": [
                        {
                            "name": "<pvc-claim-name>",
                            "mountPoint": "/opt/models"
                        }
                    ],
                    "extraPodSpec": {
                        "mainContainer": {
                            "image": "nvcr.io/nvidia/ai-dynamo/vllm-runtime:1.5.0",
                            "workingDir": "/workspace/examples/backends/vllm"
                        }
                    },
                    "envs": [
                        {
                            "name": "HF_HOME",
                            "value": "/opt/models"
                        },
                        {
                            "name": "HF_TOKEN",
                            "value": "<your-hf-token>"
                        }
                    ],
                    "replicas": 1,
                    "ingress": {
                        "enabled": true,
                        "ingressControllerClassName": "<INGRESS-CLASS>",
                        "host": "<DEPLOYMENT-PREFIX>",
                        "hostSuffix": "<INFERENCE-DOMAIN>"
                    }
                },
                "VllmPrefillWorker": {
                    "componentType": "worker",
                    "dynamoNamespace": "llama3-70b-agg",
                    "volumeMounts": [
                        {
                            "name": "<pvc-claim-name>",
                            "mountPoint": "/opt/models"
                        }
                    ],
                    "sharedMemory": {
                        "size": "20Gi"
                    },
                    "extraPodSpec": {
                        "mainContainer": {
                            "env": [
                                {
                                    "name": "SERVED_MODEL_NAME",
                                    "value": "RedHatAI/Llama-3.3-70B-Instruct-FP8-dynamic"
                                },
                                {
                                    "name": "MODEL_PATH",
                                    "value": "/opt/models/hub/models--RedHatAI--Llama-3.3-70B-Instruct-FP8-dynamic/snapshots/ddb4128556dfcff99e0c41aee159ea6c3e655dcd"
                                },
                                {
                                    "name": "HF_HOME",
                                    "value": "/opt/models"
                                },
                                {
                                    "name": "HF_TOKEN",
                                    "value": "<your-hf-token>"
                                }
                            ],
                            "args": [
                                "python3 -m dynamo.vllm --model $MODEL_PATH --served-model-name $SERVED_MODEL_NAME --tensor-parallel-size 4 --data-parallel-size 1 --disable-log-requests --gpu-memory-utilization 0.90 --no-enable-prefix-caching --block-size 128"
                            ],
                            "command": [
                                "/bin/sh",
                                "-c"
                            ],
                            "image": "nvcr.io/nvidia/ai-dynamo/vllm-runtime:1.5.0",
                            "workingDir": "/workspace/examples/backends/vllm"
                        }
                    },
                    "replicas": 1,
                    "resources": {
                        "limits": {
                            "gpu": "4"
                        },
                        "requests": {
                            "gpu": "4"
                        }
                    }
                }
            }
        }
    }
}'

```

</Tab>

<Tab title="CLI v2">

Copy the following command to your terminal. For more details, see [CLI reference](../../reference/cli/runai/runai_workload_submit.md):

1.  Save the following YAML file. Before submitting, update the following fields in the manifest:

    * `<pvc-claim-name>` - The claim name associated with the PVC created in [Step 3](aggregated-dynamo.md#step-3-creating-a-pvc-data-source).
    * `<your-hf-token>` - Required only for gated Hugging Face models. Generate a token from your [Hugging Face account](https://huggingface.co/) and provide it to the workload as an environment variable.
    * `<INGRESS-CLASS>` - The ingress controller class name configured in your cluster (for example, `haproxy`). Contact your administrator to obtain the correct value.
    * `<DEPLOYMENT-PREFIX>` - A unique prefix for the deployment endpoint (for example, `llama3-agg`).
    * `<INFERENCE-DOMAIN>` - The suffix including the cluster [FQDN](../../installation-and-configuration/system-requirements/networking-setup.md#fqdn-and-dns). Combined with the prefix, this forms the full hostname for the workload endpoint (for example, `inference.mycorp.com` if your DNS wildcard covers `*.inference.mycorp.com`). You can retrieve the cluster domain via the [Get Clusters Minimal](api:main-api:GET/api/v1/clusters/minimal) API - use the `domain` field in the response.

    ```yaml
    apiVersion: nvidia.com/v1alpha1
    kind: DynamoGraphDeployment
    metadata:
      name: llama3-70b-agg
    spec:
      backendFramework: vllm
      pvcs:
        - name: <pvc-claim-name>
          create: false
      services:
        Frontend:
          componentType: frontend
          volumeMounts:
            - name: <pvc-claim-name>
              mountPoint: /opt/models
          extraPodSpec:
            mainContainer:
              image: nvcr.io/nvidia/ai-dynamo/vllm-runtime:1.5.0
              workingDir: /workspace/examples/backends/vllm
          envs:
            - name: HF_HOME
              value: /opt/models
            - name: HF_TOKEN
              value: <your-hf-token>
          replicas: 1
          ingress:
            enabled: true
            ingressControllerClassName: <INGRESS-CLASS>
            host: <DEPLOYMENT-PREFIX>
            hostSuffix: <INFERENCE-DOMAIN>
        VllmPrefillWorker:
          componentType: worker
          volumeMounts:
            - name: <pvc-claim-name>
              mountPoint: /opt/models
          sharedMemory:
            size: 20Gi
          extraPodSpec:
            mainContainer:
              env:
                - name: SERVED_MODEL_NAME
                  value: "RedHatAI/Llama-3.3-70B-Instruct-FP8-dynamic"
                - name: MODEL_PATH
                  value: "/opt/models/hub/models--RedHatAI--Llama-3.3-70B-Instruct-FP8-dynamic/snapshots/ddb4128556dfcff99e0c41aee159ea6c3e655dcd"
                - name: HF_HOME
                  value: /opt/models
                - name: HF_TOKEN
                  value: <your-hf-token>
              args:
              - "python3 -m dynamo.vllm --model $MODEL_PATH --served-model-name $SERVED_MODEL_NAME --tensor-parallel-size 4 --data-parallel-size 1 --disable-log-requests --gpu-memory-utilization 0.90 --no-enable-prefix-caching --block-size 128"
              command:
              - /bin/sh
              - -c
              image: nvcr.io/nvidia/ai-dynamo/vllm-runtime:1.5.0
              workingDir: /workspace/examples/backends/vllm
          replicas: 1
          resources:
            limits:
              gpu: "4"
            requests:
              gpu: "4"
    ```
2.  Run the following command. Make sure to update the below with the name of your project:

    ```bash
    runai workload submit \
      --file <name>.yaml \
      --project <project-name>
    ```

</Tab>

</Tabs>

## Step 5: Verifying the Workload Status

After submitting the workload, wait for it to reach the Running status in the [Workloads](../../workloads-in-nvidia-run-ai/workloads.md) table. A workload becomes Ready to accept inference requests only after all its pods have fully initialized, including model loading.

Large models may require several minutes to load their weights, especially when the model is stored on a PVC. During this time, the workload may remain in **Initializing** even though pods are already running.

To monitor progress, select the workload and click the SHOW DETAILS button at the upper-right side of the action bar. The details pane appears, presenting the Logs tab to track model-download and model-loading progress. Select the relevant pod from the dropdown and review the pod logs.

Once the workload reaches **Running** and shows an available **Connection**, you can proceed to access the inference endpoint.

## Step 6: Accessing the Inference Workload

You can programmatically consume an inference workload via API by making direct calls to the serving endpoint, typically from other workloads or external integrations. Once an inference workload is deployed, the serving endpoint URL appears in the **Connections** column of the [Workloads](../../workloads-in-nvidia-run-ai/workloads.md) table. To retrieve the service endpoint programmatically, use the [Get Workloads](api:main-api:GET/api/v1/workloads) API. The endpoint URL will be available in the response body under `endpoints`.

<Note>

**Note**

For clusters below version 2.25, the endpoint URL will be available in the response body under `urls`.

</Note>

```bash
#replace <serving-endpoint-url> and <model-name> (e.g. "meta-llama/Llama-3.1-8B-Instruct")
curl <serving-endpoint-url>/v1/chat/completions \
  -H "Content-Type: application/json" \
  -d '{
    "model": "<model-name>",
    "messages": [{
      "role": "user",
      "content": "Tell me about NVIDIA Dynamo"
    }]
  }'
```

<Note>

**Note**

Configuring access to the inference endpoint is not supported for Dynamo-based deployments. As a result, the inference serving endpoint created in this tutorial is not protected.

Access control for inference endpoints (such as restricting access to authenticated users, service accounts, or groups) is planned for a future release.

</Note>
