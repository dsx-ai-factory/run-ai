# RAG Blueprint Deployment from NGC Catalog

This tutorial demonstrates how to deploy the NVIDIA RAG Blueprint on the NVIDIA Run:ai platform using AI applications. You can use this workflow as a reference and adapt it for other Blueprint charts and hardware configurations.

In this tutorial, you will learn how to:

* Create a Docker registry credential and a generic secret to authenticate with NGC
* Create an AI application using the NGC catalog
* Deploy the RAG Blueprint using standard GPU allocation
* Redeploy using GPU fractions to optimize resource utilization
* Access the deployed application via the API

## Prerequisites

Before you start, make sure the following requirements are met:

* Your administrator has:
  * Created a [project](../../platform-management/aiinitiatives/organization/projects.md) for you.
* You have:
  * An NGC account with an active NGC API key. To obtain a key, go to [NGC](https://catalog.ngc.nvidia.com/) → Setup → API Keys, then generate or copy an existing key.

<Note>

**Note**

Workloads using GPU fractions are likely to achieve better performance when scheduled using a spread placement strategy. It is recommended that your administrator configures the node pool's scheduling placement strategy to spread mode before deploying this application.

</Note>

## Step 1: Logging In

<Tabs>

<Tab title="UI">

Browse to the provided NVIDIA Run:ai user interface and log in with your credentials.

</Tab>

<Tab title="API">

To use the API, you will need to obtain a token as shown in [Creating a user access key](rag-blueprint-deployment.md#step-2-creating-a-user-access-key).

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

## Step 3: Creating the Docker Registry Credential

The RAG Blueprint chart requires access to NVIDIA's private container registry (`nvcr.io`) to pull its container images. This credential is referenced in the application overrides as a Kubernetes secret, making the registry authentication available to the chart at deployment time.

1. Go to Workload manager → Credentials
2. Click **+NEW CREDENTIAL** and select **Docker registry**
3. Select a [scope](../../workloads-in-nvidia-run-ai/workload-assets/overview.md#asset-scope)
4. Enter a **name** for the credential (for example, `ngc-secret`)
5. Select **New secret**
6. Enter the following:
   * **Username**: `$oauthtoken`
   * **Password**: your `<NGC API key>`
   * **Docker registry URL**: `nvcr.io`
7. Click **CREATE CREDENTIAL**

Once created, note the **Kubernetes name** from the [Credentials](../../workloads-in-nvidia-run-ai/workload-assets/credentials.md#credentials-table) table - you will need it when deploying the RAG blueprint.

## Step 4: Creating the Generic Secret Credential

The RAG Blueprint chart requires an NGC API key to authenticate with NGC during deployment. This credential is referenced in the application overrides as a Kubernetes secret, making the key available to the chart at runtime.

1. Go to Workload manager → Credentials
2. Click **+NEW CREDENTIAL** and select **Generic secret**
3. Select a [scope](../../workloads-in-nvidia-run-ai/workload-assets/overview.md#asset-scope)
4. Enter a **name** for the credential (for example, `ngc-api`)
5. Optional: Provide a description of the credential
6. Add the following key-value pairs:
   * Key: `NGC_API_KEY` - Value: your `<NGC API key>`
   * Key: `NVIDIA_API_KEY` - Value: your `<NGC API key>`
7. Click **CREATE CREDENTIAL**

Once created, note the **Kubernetes name** from the [Credentials](../../workloads-in-nvidia-run-ai/workload-assets/credentials.md#credentials-table) table - you will need it when deploying the RAG blueprint.

## Step 5: Deploying the RAG Blueprint

The RAG Blueprint deploys a set of NVIDIA NIM microservices including an LLM, embedding model, reranker, and document processing pipeline. By default, the chart allocates 8 GPUs across these components using the chart's default resource configuration.

<Tabs>

<Tab title="UI">

1. To create an AI application, go to Workload manager → AI applications
2. Click **+ NEW AI APPLICATION**
3. Select the **cluster** and **project**
4. Enter a unique **name** for the AI application (for example, `my-rag-app`)
5. Under **Source**, select **NGC catalog**:
   * Set the **Repository** to **Nvidia Blueprints**
   * Set the **Chart name** to `nvidia-blueprint-rag`
   * Set the **Chart version** to `v2.3.0`
6.  Under **Set application overrides**, paste the following, replacing the placeholder values with the Kubernetes names of the credentials you created:

    * `ngc-secret` - The Kubernetes name of the Docker registry credential created in [Step 3](rag-blueprint-deployment.md#step-3-creating-the-docker-registry-credential).
    * `ngc-api` - The Kubernetes name of the Generic secret credential created in [Step 4](rag-blueprint-deployment.md#step-4-creating-the-generic-secret-credential).

    ```yaml
    imagePullSecret:
      create: false
      name: "ngc-secret"

    ngcApiSecret:
      create: false
      name: "ngc-api"
    ```

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the following parameters. For more details, see [AI applications](api:main-api:POST/api/v1/ai-applications) API:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 2](rag-blueprint-deployment.md#step-2-creating-a-user-access-key).
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](api:main-api:GET/api/v1/org-unit/projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](api:main-api:GET/api/v1/clusters) API.
* `ngc-secret` - The Kubernetes name of the Docker registry credential created in [Step 3](rag-blueprint-deployment.md#step-3-creating-the-docker-registry-credential).
* `ngc-api` - The Kubernetes name of the Generic secret credential created in [Step 4](rag-blueprint-deployment.md#step-4-creating-the-generic-secret-credential).

```bash
curl -L 'https://<COMPANY-URL>/api/v1/ai-applications' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{  
    "projectId": "<PROJECT-ID>",
    "clusterId": "<CLUSTER-UUID>",
    "name": "<AI-Application-Name>",
    "spec": {
        "chartRef": {
            "repo": "https://helm.ngc.nvidia.com/nvidia/blueprint",
            "chartName": "nvidia-blueprint-rag",
            "chartVersion": "v2.3.0",
        },
        "values": {
            "imagePullSecret": {
                "create": false,
                "name": "ngc-secret"
            },
            "ngcApiSecret": {
                "create": false,
                "name": "ngc-api"
            }
        }
    }
}'
```

</Tab>

</Tabs>

## Step 6: Verifying the Deployment

After creating the AI application, you can monitor its status and the workloads it creates.

* Go to Workload manager → AI applications and wait for the application to reach a **Running** status.
* Go to Workload manager → Workloads and filter by **AI application**, then select your application name to see all associated workloads (such as the LLM, embedding model, reranker, and document processing components).

All workloads must reach a **Running** status before the application is ready to use.

## Step 7: Deploying the RAG Blueprint with GPU Fractions

Instead of allocating full GPUs, you can use NVIDIA Run:ai GPU fractions to share GPU resources across the RAG Blueprint components, reducing total GPU consumption to approximately 3.3 GPUs. This is done by replacing the chart's default resource requests with NVIDIA Run:ai GPU memory annotations in the application overrides.

### How the Configuration Works

<Note>

**Note**

The GPU memory values in this configuration are provided as a reference. They have not been optimized for performance and may need to be adjusted for your specific environment.

</Note>

* `gpu-memory` - Specifies the amount of GPU memory (in MiB) to allocate to the component. NVIDIA Run:ai uses this annotation to calculate the fractional GPU slice required, rather than reserving an entire GPU device.
* `gpu-memory-num-devices` - The number of GPU devices the component can access. Set to `1` for all components in this configuration.
* `resources: null` - Removes the chart's default Kubernetes resource requests and limits, allowing NVIDIA Run:ai's Scheduler to manage GPU allocation using the annotations above instead.
* `nvidia.com/gpu: "0"` - Explicitly sets GPU resource requests to zero for `nv-ingest`, as it does not require direct GPU access but relies on sub-NIMs for GPU-accelerated processing.

The following table summarizes the GPU memory requested per component:

| Component                               | GPU Memory (MiB) |
| --------------------------------------- | ---------------- |
| `nvidia-nim-llama-32-nv-embedqa-1b-v2`  | 8192             |
| `nvidia-nim-llama-32-nv-rerankqa-1b-v2` | 8192             |
| `milvus`                                | 4096             |
| `paddleocr-nim`                         | 8192             |
| `nemoretriever-graphic-elements-v1`     | 32768            |
| `nemoretriever-page-elements-v2`        | 53248            |
| `nemoretriever-table-structure-v1`      | 49152            |

### Submitting the AI Application

<Tabs>

<Tab title="UI">

1. To create an AI application, go to Workload manager → AI applications
2. Click **+ NEW AI APPLICATION**
3. Select the **cluster** and **project**
4. Enter a unique **name** for the AI application (for example, `my-rag-app`)
5. Under **Source**, select **NGC catalog**:
   * Set the **Repository** to **Nvidia Blueprints**
   * Set the **Chart name** to `nvidia-blueprint-rag`
   * Set the **Chart version** to `v2.3.0`
6.  Under **Set application overrides**, paste the following, replacing the placeholder values with the Kubernetes names of the credentials you created:

    * `ngc-secret` - The Kubernetes name of the Docker registry credential created in [Step 3](rag-blueprint-deployment.md#step-3-creating-the-docker-registry-credential).
    * `ngc-api` - The Kubernetes name of the Generic secret credential created in [Step 4](rag-blueprint-deployment.md#step-4-creating-the-generic-secret-credential).

    ```yaml
    imagePullSecret:
      create: false
      name: "ngc-secret"

    ngcApiSecret:
      create: false
      name: "ngc-api"

    nvidia-nim-llama-32-nv-embedqa-1b-v2:
      podAnnotations:
        gpu-memory: "8192"
        gpu-memory-num-devices: "1"
      resources: null

    nvidia-nim-llama-32-nv-rerankqa-1b-v2:
      podAnnotations:
        gpu-memory: "8192"
        gpu-memory-num-devices: "1"
      resources: null

    nv-ingest:
      enabled: true
      resources:
        limits:
          cpu: "4"
          memory: 24Gi
          nvidia.com/gpu: "0"
        requests:
          cpu: "2"
          memory: 16Gi

      # Milvus Vector Database
      milvus:
        annotations:
          gpu-memory: "4096"
          gpu-memory-num-devices: "1"
        standalone:
          resources: null

      # -----------------------------------------------------------------------------
      # Sub-NIMs for document processing
      # -----------------------------------------------------------------------------

      # PaddleOCR NIM
      paddleocr-nim:
        podAnnotations:
          gpu-memory: "8192"
          gpu-memory-num-devices: "1"
        resources: null

      # Graphic Elements Extraction
      nemoretriever-graphic-elements-v1:
        podAnnotations:
          gpu-memory: "32768"
          gpu-memory-num-devices: "1"
        resources: null

      # Page Elements Extraction
      nemoretriever-page-elements-v2:
        podAnnotations:
          gpu-memory: "53248" # Adjusted to 52 GB
          gpu-memory-num-devices: "1"
        resources: null

      # Table Structure Extraction
      nemoretriever-table-structure-v1:
        podAnnotations:
          gpu-memory: "49152" # Adjusted to 48 GB
          gpu-memory-num-devices: "1"
        resources: null
    ```

</Tab>

<Tab title="API">

Copy the following command to your terminal. Make sure to update the following parameters. For more details, see [AI applications](api:main-api:POST/api/v1/ai-applications) API:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface.
* `<TOKEN>` - The API access token obtained in [Step 2](rag-blueprint-deployment.md#step-2-creating-a-user-access-key).
* `<PROJECT-ID>` - The ID of the Project the workload is running on. You can get the Project ID via the [Get Projects](api:main-api:GET/api/v1/org-unit/projects) API.
* `<CLUSTER-UUID>` - The unique identifier of the Cluster. You can get the Cluster UUID via the [Get Clusters](api:main-api:GET/api/v1/clusters) API.
* `ngc-secret` - The Kubernetes name of the Docker registry credential created in [Step 3](rag-blueprint-deployment.md#step-3-creating-the-docker-registry-credential).
* `ngc-api` - The Kubernetes name of the Generic secret credential created in [Step 4](rag-blueprint-deployment.md#step-4-creating-the-generic-secret-credential).

```shellscript
curl -L 'https://<COMPANY-URL>/api/v1/ai-applications' \
-H 'Content-Type: application/json' \
-H 'Authorization: Bearer <TOKEN>' \
-d '{
    "projectId": "<PROJECT-ID>",
    "clusterId": "<CLUSTER-UUID>",
    "name": "<AI-APPLICATION-NAME>",
    "spec": {
        "chartRef": {
            "repo": "https://helm.ngc.nvidia.com/nvidia/blueprint",
            "chartName": "nvidia-blueprint-rag",
            "chartVersion": "v2.3.0"
        },
        "values": {
            "imagePullSecret": {
                "create": false,
                "name": "ngc-secret"
            },
            "ngcApiSecret": {
                "create": false,
                "name": "ngc-api"
            },
            "nvidia-nim-llama-32-nv-embedqa-1b-v2": {
                "podAnnotations": {
                    "gpu-memory": "8192",
                    "gpu-memory-num-devices": "1"
                },
                "resources": null
            },
            "nvidia-nim-llama-32-nv-rerankqa-1b-v2": {
                "podAnnotations": {
                    "gpu-memory": "8192",
                    "gpu-memory-num-devices": "1"
                },
                "resources": null
            },
            "nv-ingest": {
                "enabled": true,
                "resources": {
                    "limits": {
                        "cpu": "4",
                        "memory": "24Gi",
                        "nvidia.com/gpu": "0"
                    },
                    "requests": {
                        "cpu": "2",
                        "memory": "16Gi"
                    }
                },
                "milvus": {
                    "annotations": {
                        "gpu-memory": "4096",
                        "gpu-memory-num-devices": "1"
                    },
                    "standalone": {
                        "resources": null
                    }
                },
                "paddleocr-nim": {
                    "podAnnotations": {
                        "gpu-memory": "8192",
                        "gpu-memory-num-devices": "1"
                    },
                    "resources": null
                },
                "nemoretriever-graphic-elements-v1": {
                    "podAnnotations": {
                        "gpu-memory": "32768",
                        "gpu-memory-num-devices": "1"
                    },
                    "resources": null
                },
                "nemoretriever-page-elements-v2": {
                    "podAnnotations": {
                        "gpu-memory": "53248",
                        "gpu-memory-num-devices": "1"
                    },
                    "resources": null
                },
                "nemoretriever-table-structure-v1": {
                    "podAnnotations": {
                        "gpu-memory": "49152",
                        "gpu-memory-num-devices": "1"
                    },
                    "resources": null
                }
            }
        }
    }
}'
```

</Tab>

</Tabs>

## Step 8: Verifying the Fractions Deployment

After creating the AI application, verify that it is running with fractional GPU allocation.

* To check the application status, go to Workload manager → AI applications and wait for the application to reach a **Running** status.
* To verify the GPU allocation, go to Workload manager → Workloads and filter by **AI application** to view the individual workloads. The total GPU compute allocation across all components should reflect approximately 3.3 GPUs, compared to 8 GPUs in the standard deployment.

All workloads must reach a **Running** status before the application is ready to use.

## Step 9: Accessing the RAG Application

Once the application is running, you can retrieve its access URLs using the UI or the API. Use these URLs to interact with the RAG application - for example, to send queries to the chat interface, call the LLM API, or integrate the RAG pipeline into your own applications.

<Tabs>

<Tab title="UI">

Go to Workload manager → AI applications. The **Connection(s)** column displays the network endpoints exposed by the application. If there is one endpoint, the URL is shown directly. If there are multiple, a count is shown. Click it to open the connections panel and view each endpoint's URL, port, and connection type.

</Tab>

<Tab title="API">

To get the endpoints for your AI application, use the [Get AI application](api:main-api:GET/api/v1/ai-applications/{aiApplicationId}) API, which returns the list of network endpoints exposed by the application and its associated workloads. You can also retrieve endpoints for a specific workload directly using the [Get workload endpoints](api:main-api:GET/api/v1/workloads) API.

</Tab>

</Tabs>
