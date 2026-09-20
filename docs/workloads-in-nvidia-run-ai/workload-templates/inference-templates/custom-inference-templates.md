# Custom Inference Templates

This guide explains how to create custom inference templates for reuse during workload submission. To manage templates, see [Workload Templates](../README.md).

<Info>

**Note**

* Flexible workload templates is enabled by default and applies only to flexible workload submission (enabled by default). If unavailable, contact your administrator to enable it under **General settings** → Workloads → Flexible workload templates.
* **Docker registry URL for inference workloads** - For Knative-based inference workloads, Docker Hub credentials must be configured using `https://index.docker.io/v1/` as the registry URL. Credentials configured with `docker.io` result in `401 Unauthorized` errors for Knative-based inference workloads due to differences in how image digests are resolved during image pull. See [Credentials](../../workload-assets/credentials.md#docker-registry) for more details.

</Info>

## Before You Start
If the image you are planning to use is hosted in NGC, make sure you have an NGC account with an active NGC API key. To obtain a key, go to [NGC](https://catalog.ngc.nvidia.com/) → Setup → API Keys, then generate or copy an existing key. In NVIDIA Run:ai, store the key either as a [shared secret](../../workload-assets/credentials.md#ngc-api-key) (created by an administrator) or a [user credential](../../../settings/user-settings/user-credentials.md) (created under User settings).

## Workload Priority and Preemption
By default, inference workloads in NVIDIA Run:ai are assigned a **Very high** priority and are **non-preemptible**. These defaults ensures that inference workloads, which often serve real-time or latency-sensitive traffic, are guaranteed the resources they need. The default very-high priority ensures the fastest possible scheduling, and the default non-preemptible value guarantees that the workload will not be disrupted by other workloads once it starts running.

You can override the defaults by configuring priority and preemptibility. For more details, see [Workload priority and preemption](../../../platform-management/runai-scheduler/scheduling/workload-priority-control.md).

## Linking Assets

When loading an existing asset, environment or compute resource, into a template, you can choose whether to link the asset or use it without linking. Linked assets remain connected to the template. Any updates made to the original environment or compute resource are automatically reflected in the template. While linked, the asset fields in the template cannot be modified.

<Info>

**Note**

Linking data source assets is currently not supported.

</Info>

## Adding a New Template

1. To add a new template, go to Workload manager → Templates.
2. Click **+NEW TEMPLATE** and select **Inference** from the dropdown menu**.**
3. Within the new inference template form, select the **scope**.
4. Set the inference server to **custom** (if applicable). This allows you to use an inference server not listed, such as SGLang. This requires a custom container image and manual server configuration.
5. Enter a unique **name** for the inference template. If the name already exists in the project, you will be requested to submit a different name.
6. Click **CONTINUE**

### Setting Up an Environment

<Info>

**Note**

* NGC catalog is disabled by default. If unavailable, your administrator must enable it under **General settings** → Workloads → NGC catalog.
* To select an image from the NGC private registry**,** your administrator must configure it under **General settings** → Workloads → NGC private registry.

</Info>

**Load from existing setup**

1. Click the **load** icon. A side pane appears, displaying a list of available environments. Select an environment from the list.
2. Alternatively, click the **➕** icon in the side pane to create a new environment. For step-by-step instructions, see [Environments](../../workload-assets/environments.md).
3. Choose whether to **link** the environment when applying it to the template. See [Linking assets](custom-inference-templates.md#linking-assets) for more details.

**Provide your own settings**

Manually configure the settings below as needed.

**Configure environment**

1. Set the **environment image**:
   * Select **Custom image** and add the **Image URL** or update the URL of the existing setup.
   * **Select from the NGC catalog** and then set how to **access the NGC catalog**:
     * **As a guest** - Choose the **image name** and **tag** from the dropdown.
     * **Authenticated** - Configure **access** to NGC and then choose the **image name** and **tag** from the dropdown:
       * Under **Source**, select [**Shared secret**](../../workload-assets/credentials.md#ngc-api-key) or [**My credentials**](../../../settings/user-settings/user-credentials.md).
       * The **Type** field is fixed to **NGC API key**.
       * Under **Credential name**, select an existing credential from the dropdown. If you select **My credentials**, you can also create a new credential directly from the dropdown. The credential is also saved under **User settings** → Credentials.
   * **Select from the NGC private registry** and then set how to access the registry:
     * Select a **registry** from the dropdown.
     * Under **Source**, select [**Shared secret**](../../workload-assets/credentials.md#ngc-api-key) or [**My credentials**](../../../settings/user-settings/user-credentials.md).
     * The **Type** field is fixed to **NGC API key**.
     * Under **Credential name**, select an existing credential from the dropdown. If you select **My credentials**, you can also create a new credential directly from the dropdown. The credential is also saved under **User settings** → Credentials.
2. Set the condition for pulling the image by selecting the **image pull policy**. It is recommended to pull the image only if it's not already present on the host.
3. Set an inference **serving endpoint**. The connection protocol and the container port are defined within the environment:
   * Select **HTTP** or **gRPC** and enter a corresponding **container port**
   * Modify who can access the endpoint. See [Accessing the inference workload](../../using-inference/custom-inference.md#accessing-the-inference-workload) for more details:
     * By default, **Public** is selected giving everyone within the network access to the endpoint with no authentication
     * If you select **All authenticated users and service accounts**, access is given to everyone within the organization’s account that can log in (to NVIDIA Run:ai or SSO).
     * For **Specific group(s)**, enter **group names** as they appear in your identity provider. You must be a member of one of the groups listed to have access.
     * For **Specific user(s) and service account(s)**, enter a valid user email or name. If you remove yourself, you will lose access.
4. Set whether inference serving endpoints can be accessed externally or only by internal cluster traffic:
   * **External (Public access)** - Available only if your administrator has configured **Knative** to support external access. See the [Inference requirements](../../../installation-and-configuration/installation/install-using-helm/prerequisites-and-features.md#standard-inference) section for more details.
   * **Internal only** - Endpoint is accessible only within cluster traffic.
5. Set the connection for your **tool(s)**. If you are loading from existing setup, the tools are configured as part of the environment.
   * Select the connection type:
     * **Auto generate** - A unique URL / port is automatically created for each workload using the environment.
     * **Custom URL** / **Custom port** - Manually define the URL or port. For custom port, make sure to enter a port between `30000` and `32767.` If the node port is already in use, the workload will fail and display an error message.
     * **Load Balancer** - Set the container port. Connection handling is managed by the load balancer. For more information, see [External access to containers](../../../installation-and-configuration/configuration/networking/external-access-to-containers.md).
   * Modify who can **access** the tool:
     * By default, **All authenticated users** **and service accounts** is selected giving access to everyone within the organization’s account.
     * For **Specific group(s)**, enter **group names** as they appear in your identity provider. You must be a member of one of the groups listed to have access to the tool.
     * For **Specific user(s) and service account(s)**, enter a valid user email or name. If you remove yourself, you will lose access to the tool.
6. Set the **command and arguments** for the container running the workload. If no command is added, the container will use the image’s default command (entry-point).
   * Modify the existing command or click **+COMMAND & ARGUMENTS** to add a new command.
   * Set multiple arguments separated by spaces, using the following format (e.g.: `--arg1=val1`).
7.  Set the **environment variable(s)**:

    * Modify the existing environment variable(s) or click **+ENVIRONMENT VARIABLE**. The existing environment variables may include instructions to guide you with entering the correct values.
    * You can either select **Custom** to define a value manually, or choose an existing value from [**Shared secret**](../../workload-assets/credentials.md) or [**ConfigMap**](../../workload-assets/environments.md).

    Some environment variables are automatically injected by NVIDIA Run:ai. See [Built-in workload environment variables](../../workloads.md#built-in-workload-environment-variables) for more details.
8. Enter a path pointing to the **container's working directory**
9. Set where the UID, GID, and supplementary groups for the container should be taken from. If you select **Custom**, you’ll need to manually enter the **UID,** **GID and** **Supplementary groups values**.
10. Select additional Linux capabilities for the container from the drop-down menu. This grants certain privileges to a container without granting all the root user's privileges.

### Setting Up Compute Resources

<Info>

**Note**

* GPU memory limit is disabled by default. If unavailable, your administrator must enable it under **General settings** → Resources → GPU resource optimization.
* Replica autoscaling and toleration(s) can still be modified even when the compute resource asset is linked.

</Info>

**Load from existing setup**

1. Click the **load** icon. A side pane appears, displaying a list of available compute resources. Select a compute resource from the list.
2. Alternatively, click the **➕** icon in the side pane to create a new compute resource. For step-by-step instructions, see [Compute resources](../../workload-assets/compute-resources.md).
3. Choose whether to **link** the compute resource when applying it to the template. See [Linking assets](custom-inference-templates.md#linking-assets) for more details.

**Provide your own settings**

Manually configure the settings below as needed.

**Configure compute resources**

1. Set the number of **GPU devices** per pod (physical GPUs).
2. Enable **GPU fractioning** to set the GPU memory per device using either a fraction of a GPU device’s memory **(% of device)** or a GPU memory unit **(MB/GB)**:
   * **Request** - The minimum GPU memory allocated per device. Each pod in the workload receives at least this amount per device it uses.
   * **Limit** - The maximum GPU memory allocated per device. Each pod in the workload receives **at most** this amount of GPU memory for each device(s) the pod utilizes. This is disabled by default, to enable see the above note.
3. Set the **CPU resources**
   * Set **CPU compute resources** per pod by choosing the unit (**cores** or **millicores**):
     * **Request** - The minimum amount of CPU compute provisioned per pod. Each running pod receives this amount of CPU compute.
     * **Limit** - The maximum amount of CPU compute a pod can use. Each pod receives **at most** this amount of CPU compute. By default, the limit is set to **Auto** which means that the pod may consume up to the node's maximum available CPU compute resources.
   * Set the **CPU memory per pod** by selecting the unit (**MB** or **GB**):
     * **Request** - The minimum amount of CPU memory provisioned per pod. Each running pod receives this amount of CPU memory.
     * **Limit** - The maximum amount of CPU memory a pod can use. Each pod receives at most this amount of CPU memory. By default, the limit is set to **Auto** which means that the pod may consume up to the node's maximum available CPU memory resources.
4. Set **extended resource(s)**
   * Enable **Increase shared memory size** to allow the shared memory size available to the pod to increase from the default 64MB to the node's total available memory or the CPU memory limit, if set above.
   * Click **+EXTENDED RESOURCES** to add resource/quantity pairs. For more information on how to set extended resources, see the [Extended resources](https://kubernetes.io/docs/tasks/configure-pod-container/extended-resource/) and [Quantity](https://kubernetes.io/docs/reference/kubernetes-api/common-definitions/quantity/) guides.
5. Set the **minimum and maximum** number of replicas to be scaled up and down to meet the changing demands of inference services:
   * If the number of minimum and maximum replicas are different, autoscaling will be triggered and you'll need to set **conditions for creating a new replica**. A replica will be created every time a condition is met. When a condition is no longer met after a replica was created, the replica will be automatically deleted to save resources.
   * Select one of the **variables to set** the conditions for creating a new replica. The variable's values will be monitored via the container's port. When you set a **value**, this value is the threshold at which autoscaling is triggered.
6. Set when the replicas should be automatically **scaled down to zero**. This allows compute resources to be freed up when the model is inactive (i.e., there are no requests being sent). Automatic scaling to zero is enabled only when the minimum number of replicas in the previous step is set to 0.
7. Click **+TOLERATION** to allow the workload to be scheduled on a node with a matching taint. Select the **operator** and the **effect**:
   * If you select **Exists**, the effect will be applied if the key exists on the node.
   * If you select **Equals,** the effect will be applied if the key and the value set match the value on the node.

### Setting Up Data & Storage

<Info>

**Note**

* Data volumes is enabled by default. If unavailable, contact your administrator to enable it under **General settings** → Workloads → Data volumes.
* If Data volumes is not enabled, **Data & storage** appears as **Data sources** only, and no data volumes will be available.
* S3 data sources are not supported for inference workloads.

</Info>

**Load from existing setup**

1. Click the **load** icon. A side pane appears, displaying a list of available data sources/volumes. Select a data source/volume from the list.
2. Optionally, customize any of the data source's predefined fields as shown below. The changes will apply to this template only and will not affect the selected data source:
   * **Container path** - Enter the **container path** to set the **data target location**.
   * **ConfigMap** **sub-path** - Specify a **sub-path** (file/key) inside the ConfigMap to mount (for example, `app.properties`). This lets you mount a single file from an existing ConfigMap.
3. Alternatively, click the **➕** icon in the side pane to create a new data source/volume. For step-by-step instructions, see [Data sources](../../workload-assets/overview.md) or [Data volumes](../../workload-assets/data-volumes.md).

**Configure data sources for a one-time configuration**

<Info>

**Note**

[PVCs](../../workload-assets/data-volumes.md), [Secrets](../../workload-assets/credentials.md), [ConfigMaps](../../workload-assets/environments.md) and [Data volumes](../../workload-assets/data-volumes.md) cannot be added as a one-time configuration.

</Info>

1. Click the **➕** icon and choose the data source from the dropdown menu. You can add multiple data sources.
2. Once selected, set the **data origin** according to the required fields and enter the **container path** to set the **data target location**.

The required fields vary by data source. For detailed configuration options and usage guidelines for each data source type, see [Data sources](../../workload-assets/overview.md).

**Configure EmptyDir for a one-time configuration**

Use **EmptyDir** to allocate temporary storage that exists only for the lifetime of the workload. The **EmptyDir** volume is ephemeral, the volume and its data is deleted every time the workload’s status changes to “Stopped

1. Set the **size** and **units** to define the maximum storage capacity.
2. Enter the **container path** to set the **data target location**.
3. Select the **storage medium**:
   * **Disk-backed storage** - The data is stored on the node's filesystem, which persists across container restarts but not pod rescheduling, and is slower than memory
   * **Memory-backed storage** - The data is stored in RAM, which provides faster access but is not persistent, and will be lost if the pod restarts

**Configure Volume for a one-time configuration**

Select **Volume** to allocate a storage space to your workspace that is persistent across restarts:

1. Set the **Storage class** to **None** or select an existing storage class from the list. To add new storage classes, and for additional information, see [Kubernetes storage classes](../../../installation-and-configuration/configuration/shared-storage.md). If the administrator defined the storage class configuration, the rest of the fields will appear accordingly.
2. Select one or more **access mode(s)** and define the **claim size** and its **units**.
3. Select the **volume mode**:
   1. **Filesystem** (default) - The volume will be mounted as a filesystem, enabling the usage of directories and files.
   2. **Block** - The volume is exposed as a block storage, which can be formatted or used directly by applications without a filesystem.
4. Enter the **container path** to set the **data target location**.

### Setting Up General Settings

<Info>

**Note**

The following general settings are optional.

</Info>

1. Set whether the workload may be interrupted by selecting **Preemptible**, **Semi-preemptible**, or **Non-preemptible**:
   * Preemptible workloads may be interrupted if resources are needed for higher-priority workloads.
   * Semi-preemptible workloads guarantee their minimum required replicas run without interruption, while any replicas beyond that minimum behave as preemptible replicas.
   * Non-preemptible workloads use the project's available GPU quota and will not be interrupted once they start running.
2. Set the **workload priority**. Choose the appropriate priority level for the workload. Higher-priority workloads are scheduled before lower-priority ones.
3. Set the **workload initialization timeout.** This is the maximum amount of time the system will wait for the workload to start and become ready. If the workload does not start within this time, it will automatically fail. Enter a value between 5 seconds and 720 minutes. If you do not set a value, the default is taken from [Knative’s max-revision-timeout-seconds](https://knative.dev/docs/serving/configuration/config-defaults/#max-revision-timeout-seconds).
4. Set the **request timeout.** This defines the maximum time allowed to process an end-user request. If the system does not receive a response within this time, the request will be ignored. Enter a value between 5 seconds and 10 minutes. If you do not set a value, the default is taken from [Knative’s revision-timeout-seconds](https://knative.dev/docs/serving/configuration/config-defaults/#revision-timeout-seconds).
5. Set **annotations(s).** Kubernetes annotations are key-value pairs attached to the workload. They are used for storing additional descriptive metadata to enable documentation, monitoring and automation.
6. Set **labels(s).** Kubernetes labels are key-value pairs attached to the workload. They are used for categorizing to enable querying.

### Completing the Template

1. Before finalizing your template, review your configurations and make any necessary adjustments.
2. Click **CREATE TEMPLATE**

## Using API

Go to the [Workload templates](https://run-ai-docs.nvidia.com/api/2.26/workloads/workload-templates) API reference to view the available actions.
