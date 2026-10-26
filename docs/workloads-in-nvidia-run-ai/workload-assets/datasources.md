# Data Sources

Data sources are a type of [workload assets](overview.md) and represent a location where data is actually stored. They may represent a remote data location, such as NFS, Git, or S3, or a Kubernetes local resource, such as PVC, ConfigMap, HostPath, or Secret.

This configuration simplifies the mapping of the data into the workload's file system and handles the mounting process during workload creation for reading and writing. These data sources are reusable and can be easily integrated and used by AI practitioners while submitting workloads across various scopes.

## Data Sources Table

The data sources table can be found under **Workload manager** in the NVIDIA Run:ai platform.

The data sources table provides a list of all the data sources defined in the platform and allows you to manage them.

<Note>

**Note**

**Data & storage** - with **Data sources** and **Data volumes -** is visible only if your administrator has enabled [Data volumes](data-volumes.md)**.**

</Note>

<figure><img src="../../assets/data-sources-new.png" alt=""/><figcaption></figcaption></figure>

The data sources table comprises the following columns:

| Column          | Description                                                                                                                                             |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Data source     | The name of the data source                                                                                                                             |
| Description     | A description of the data source                                                                                                                        |
| Type            | The type of data source connected – e.g., S3 bucket, PVC, or others                                                                                     |
| Status          | The different lifecycle [phases](../workloads.md#workload-status) and representation of the data source condition                                    |
| Scope           | The [scope](overview.md#asset-scope) of the data source within the organizational tree. Click the scope name to view the organizational tree diagram |
| Kubernetes name | The unique name of the data sources Kubernetes name as it appears in the cluster                                                                        |
| Created by      | The user who created the data source                                                                                                                    |
| Creation time   | The timestamp for when the data source was created                                                                                                      |
| Cluster         | The cluster that the data source is associated with                                                                                                     |

### Data Sources Status

The following table describes the data sources' condition and whether they were created successfully for the selected [scope](overview.md#asset-scope).

| Status          | Description                                                                                                                                                                   |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| No issues found | No issues were found while creating the data source                                                                                                                           |
| Issues found    | Issues were found while propagating the data source credentials                                                                                                               |
| Issues found    | The data source couldn't be created at the cluster                                                                                                                            |
| Creating…       | The data source is being created                                                                                                                                              |
| No status / "-" | When the data source's scope is an account, the current version of the cluster is not up to date, or the asset is not a cluster-syncing entity, the status can't be displayed |

### Customizing the Table View

* Filter - Click ADD FILTER, select the column to filter by, and enter the filter values
* Search - Click SEARCH and type the value to search by
* Sort - Click each column header to sort by
* Column selection - Click COLUMNS and select the columns to display in the table
* Download table - Click MORE and then click 'Download as CSV'. Export to CSV is limited to 20,000 rows.
* Refresh - Click REFRESH to update the table with the latest data

### Show/Hide Details

Click a PVC data source in the Data sources table and then click the SHOW DETAILS button at the upper-right side of the action bar. The details pane appears, presenting the following tabs:

#### Event History

Displays events describing notable actions and changes in the cluster. Use the filter to search through the history for specific events.

#### Details

Displays static PVC information, including:

* General information - Scope, creation time, owner
* Request settings - Access modes, storage request, volume mode, restrictions
* Storage class details - Name, provisioner type, volume expansion support, default status

## Adding a New Data Source

To create a new data source:

1. Click **+NEW DATA SOURCE**
2. Select the data source type from the list. Follow the step-by-step guide for each data source type:

<Accordion title="NFS">

**NFS**

A Network File System ([NFS](https://kubernetes.io/docs/concepts/storage/volumes/#nfs)) is a Kubernetes concept used for sharing storage in the cluster among different pods. Like a PVC, the NFS volume's content remains preserved, even outside the lifecycle of a single pod. However, unlike PVCs, which abstract storage management, NFS provides a method for network-based file sharing. The NFS volume can be pre-populated with data and can be mounted by multiple pod writers simultaneously. At NVIDIA Run:ai, an NFS-type data source is an abstraction that is mapped directly to a Kubernetes NFS volume. This integration allows multiple workloads under various scopes to mount and present the NFS data source.

1. Select the **cluster** under which to create this data source
2. Select a [**scope**](overview.md#asset-scope)
3. Enter a **name** for the data source. The name must be unique.
4. Optional: Provide a **description** of the data source
5. Set the data origin
   * Enter the **NFS server** (host name or host IP)
   * Enter the **NFS path**
6. Set the data target location
   * **Container path**
7. Optional: Restrictions
   * **Prevent data modification** - When enabled, the data will be mounted with read-only permissions
8. Click **CREATE DATA SOURCE**

</Accordion>

<Accordion title="PVC">

**PVC**

A Persistent Volume Claim ([PVC](https://kubernetes.io/docs/concepts/storage/persistent-volumes/)) is a Kubernetes concept used for managing storage in the cluster, which can be provisioned by an administrator or dynamically by Kubernetes using a StorageClass. PVCs allow users to request specific sizes and access modes (read/write once, read-only many).\
NVIDIA Run:ai ensures that data remains consistent and accessible across various scopes and workloads, beyond the lifecycle of individual pods, which is efficient while working with large datasets typically associated with AI projects.

1. Select the **cluster** under which to create this data source
2. Select a [**scope**](overview.md#asset-scope)
3. If the scope you selected is at the **cluster** or **department** level, you can choose how volumes are shared across projects within that scope:
   * **Separate volume for each project** (default) - A dedicated volume is created per project. Each project can access only its own allocated storage space.
   * **Shared volume between all projects** - A dedicated volume is created for all projects within the selected scope. This option always creates a new PVC - the **Existing PVC** option is not available. Recommended for collaboration and shared storage access across projects.\
     Note: A storage class that supports shared volumes must be configured in your cluster. Contact your administrator, who must enable data sharing for the storage class as described in [Enabling Shared Volumes for a Storage Class](#enabling-shared-volumes-for-a-storage-class).
4. Enter a **name** for the data source. The name must be unique.
5. Optional: Provide a **description** of the data source
6. Select PVC:
   *   **Existing PVC**

       This option is relevant when the purpose is to create a PVC-type data source based on an existing PVC in the cluster

       * Select a PVC from the list - (The list is empty if no existing PVCs were [created in advance](#creating-pvcs-in-advance))
   *   **New PVC** - creates a new PVC in the cluster. New PVCs are not added to the Existing PVCs list.

       When creating a PVC-type data source and selecting the 'New PVC' option, the PVC is immediately created in the cluster (even if no workload has requested this PVC).
7. Select the **storage class**
   * **None -** Proceed without defining a storage class
   * **Custom storage class** - Select a storage class from the list of available options. If the selected storage class includes custom configuration fields, those fields are displayed in the form. To add new storage classes and for additional information, see [Kubernetes storage classes](../../installation-and-configuration/configuration/shared-storage.md#kubernetes-storage-classes).
8. Select the **access mode(s)** (multiple modes can be selected)
   * **Read-write by one node** - The volume can be mounted as read-write by a single node.
   * **Read-only by many nodes** - The volume can be mounted as read-only by many nodes.
   * **Read-write by many nodes** - The volume can be mounted as read-write by many nodes.
9. Set the **claim size** and its **units**
10. Select the **volume mode**
    1. **File system** (default) - allows the volume to be mounted as a filesystem, enabling the usage of directories and files.
    2. **Block** - exposes the volume as a block storage, which can be formatted or used by applications directly without a filesystem.
11. Set the data target location
    * **container path**
12. Optional: **Prevent data modification** - When enabled, the data will be mounted with read-only permission.
13. Click **CREATE DATA SOURCE**

After the data source is created, check its status to monitor its proper creation across the selected scope.

</Accordion>

<Accordion title="S3 Bucket">

**S3 Bucket**

The [S3 bucket](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-resource-s3-bucket.html) data source enables the mapping of a remote S3 bucket into the workload's file system. Similar to a PVC, this mapping remains accessible across different workload executions, extending beyond the lifecycle of individual pods. However, unlike PVCs, data stored in an S3 bucket resides remotely, which may lead to decreased performance during the execution of heavy machine learning workloads. As part of the NVIDIA Run:ai connection to the S3 bucket, you can create [credentials](credentials.md) in order to access and map private buckets.

<Note>

**Note**

S3 data sources are not supported for custom inference workloads.

</Note>

1. Select the **cluster** under which to create this data source
2. Select a [**scope**](overview.md#asset-scope)
3. Enter a **name** for the data source. The name must be unique.
4. Optional: Provide a **description** of the data source
5. Set the data origin
   * Set the **S3 service URL**
   * Select the **credential**
     * **None** - for public buckets
     * **Credential names** - This option is relevant for private buckets based on existing credentials that were created for the scope. To add new credentials to the credentials list, and for additional information, check the [Credentials](credentials.md) article.
   * Enter the **bucket name**
6. Set the data target location
   * **container path**
7. Click **CREATE DATA SOURCE**

After a private data source is created, check its status to monitor its proper creation across the selected scope.

</Accordion>

<Accordion title="Git">

**Git**

A Git-type data source is an NVIDIA Run:ai integration that enables code to be copied from a Git branch into a dedicated folder in the container. It is mainly used to provide the workload with the latest code repository. As part of the integration with Git, in order to access private repositories, you can add predefined credentials to the data source mapping.

1. Select the **cluster** under which to create this data source
2. Select a [**scope**](overview.md#asset-scope)
3. Enter a **name** for the data source. The name must be unique.
4. Optional: Provide a **description** of the data source
5. Set the data origin
   * Set the **repository type** - **Public** or **Private**
   * If you select **Private**, choose the **credential**. The credential must be an existing credential that was created for the scope. Supported credential types include **Username & password** and **SSH keys**. To add new credentials to the credentials list, and for additional information, check the [Credentials](credentials.md) article.
   * Set the **Repository URL**
     * For username/password credentials, use HTTPS format
     * For SSH key credentials, use SSH format
   * Set the **Revision** (branch, tag, or hash). If left empty, the 'HEAD' (latest) will be used.
6. Set the data target location
   * **container path**
7. Click **CREATE DATA SOURCE**

After a private data source is created, check its status to monitor its proper creation across the selected scope.

</Accordion>

<Accordion title="Host path">

**Host path**

A [Host path](https://kubernetes.io/docs/concepts/storage/volumes/#hostpath) volume is a Kubernetes concept that enables mounting a host path file or a directory on the workload's file system. Like a PVC, the host path volume's data persists across workloads under various scopes. It also enables data serving from the hosting node.

Administrators can restrict the use of host path data sources with a workload policy. See [Policy YAML examples](../../platform-management/policies/policy-yaml-examples.md) and [Security best practices](../../installation-and-configuration/configuration/security/security-best-practices.md).

1. Select the **cluster** under which to create this data source
2. Select a [**scope**](overview.md#asset-scope)
3. Enter a **name** for the data source. The name must be unique.
4. Optional: Provide a **description** of the data source
5. Set the data origin
   * **host path**
6. Set the data target location
   * **container path**
7. Optional: **Prevent data modification** - When enabled, the data will be mounted with read-only permissions.
8. Click **CREATE DATA SOURCE**

</Accordion>

<Accordion title="ConfigMap">

**ConfigMap**

A [ConfigMap](https://kubernetes.io/docs/concepts/configuration/configmap/) data source is an NVIDIA Run:ai abstraction for the Kubernetes ConfigMap concept. The ConfigMap is used mainly for storage that can be mounted on the workload container for non-confidential data. It is usually represented in key-value pairs (e.g., environment variables, command-line arguments etc.). It allows you to decouple environment-specific system configurations from your container images, so that your applications are easily portable. ConfigMaps must be created on the cluster prior to being used within the NVIDIA Run:ai system.\\

1. Select the **cluster** under which to create this data source
2. Select a [**scope**](overview.md#asset-scope)
3. Enter a **name** for the data source. The name must be unique.
4. Optional: Provide a **description** of the data source
5. Set the data origin
   * Select the **ConfigMap name** (The list is empty if no existing ConfigMaps were [created in advance](#creating-configmaps-in-advance)).
6. Set the data target location
   * **container path**
7. Click **CREATE DATA SOURCE**

</Accordion>

<Accordion title="Secret">

**Secret**

A secret-type data source enables the mapping of a credential into the workload's file system. Credentials are a workload asset that simplify the complexities of Kubernetes [Secrets](https://kubernetes.io/docs/concepts/configuration/secret/). The credentials mask sensitive access information, such as passwords, tokens, and access keys, which are necessary for gaining access to various resources.

1. Select the **cluster** under which to create this data source
2. Select a [**scope**](overview.md#asset-scope)
3. Enter a **name** for the data source. The name must be unique.
4. Optional: Provide a **description** of the data source
5. Set the data origin
   *   Select the **credential**

       To add new credentials, and for additional information, check the [Credentials](credentials.md) article.
6. Set the data target location
   * **container path**
7. Click **CREATE DATA SOURCE**

After the data source is created, check its status to monitor its proper creation across the selected scope.

</Accordion>

<Note>

**Note**

It is also possible to add data sources directly when creating a specific [workspace](../using-workspaces/running-workspace.md), [training](../using-training/train-models.md) or [inference](../using-inference/nvidia-run-ai-inference-overview.md) workload.

</Note>

## Copying a Data Source

To copy an existing data source:

1. Select the data source you want to copy
2. Click **MAKE A COPY**
3. Enter a **name** for the data source. The name must be unique.
4. Update the data source and click **CREATE DATA SOURCE**

## Renaming a Data Source

To rename an existing data source:

1. Select the data source you want to rename
2. Click **Rename** and edit the name/description

## Deleting a Data Source

To delete a data source:

1. Select the data source you want to delete
2. Click **DELETE**
3. Confirm you want to delete the data source

<Note>

**Note**

It is not possible to delete a data source being used by an existing workload or template.

</Note>

## Creating PVCs in Advance

Add PVCs in advance to be used when creating a PVC-type data source via the NVIDIA Run:ai UI.

The actions taken by the admin are based on the scope (cluster, department or project) that the admin wants for data source of type PVC. Follow the steps below for each required scope:

<Note>

**Note**

These steps require administrator permissions.

</Note>

### **Cluster Scope**

1. Locate the PVC in the NVIDIA Run:ai namespace (`runai`)
2. Provide NVIDIA Run:ai with visibility and authorization to share the PVC to your selected scope by implementing the following label: `run.ai/cluster-wide: "true"`

The PVC is now displayed for that scope in the list of existing PVCs.

<Note>

**Note**

This step is also relevant for creating the data source of type PVC via API

</Note>

### **Department Scope**

1. Locate the PVC in the NVIDIA Run:ai namespace (`runai`)
2. To authorize NVIDIA Run:ai to use the PVC, label it: `run.ai/department: "id"`

The PVC is now displayed for that scope in the list of existing PVCs.

### **Project Scope**

Locate the PVC in the project's namespace.

The PVC is now displayed for that scope in the list of existing PVCs.

## Enabling Shared Volumes for a Storage Class

The **Shared volume between all projects** option is available only for storage classes that have data sharing enabled. Data sharing is not enabled automatically. An administrator must use the [Storage class configuration](api:main-api:GET/api/v1/storage-class-configuration) API to enable it explicitly for each storage class that should support shared volumes, by setting `permissions.allowDataSharing` to `true`.

<Note>

**Note**

This procedure requires administrator permissions.

</Note>

Replace the following placeholders in the commands below:

* `<COMPANY-URL>` - The link to the NVIDIA Run:ai user interface
* `<TOKEN>` - A valid [API access token](../../rest-api/how-to-authenticate-to-the-api.md)
* `<ID>` - The `id` of the existing storage class configuration
* `<CLUSTER-UUID>` - The UUID of the cluster
* `<STORAGE-CLASS-NAME>` - The name of the Kubernetes storage class

1. Check whether a storage class configuration exists for the storage class, and whether `allowDataSharing` is enabled:

   ```bash
   curl -X GET 'https://<COMPANY-URL>/api/v1/storage-class-configuration' \
     -H 'Authorization: Bearer <TOKEN>' \
     -H 'Content-Type: application/json'
   ```

2. If a configuration exists for the storage class but `allowDataSharing` is `false`, copy its `id` and update it:

   ```bash
   curl -X PATCH 'https://<COMPANY-URL>/api/v1/storage-class-configuration/<ID>' \
     -H 'Authorization: Bearer <TOKEN>' \
     -H 'Content-Type: application/json' \
     -d '{"permissions": {"allowDataSharing": true}}'
   ```

3. If no configuration exists for the storage class, create one:

   ```bash
   curl -X POST 'https://<COMPANY-URL>/api/v1/storage-class-configuration' \
     -H 'Authorization: Bearer <TOKEN>' \
     -H 'Content-Type: application/json' \
     -d '{
       "scopeType": "cluster",
       "scopeId": "<CLUSTER-UUID>",
       "name": "<STORAGE-CLASS-NAME>",
       "permissions": {
         "allowDataSharing": true
       }
     }'
   ```

Repeat these steps for each storage class that should support shared volumes.

## Creating ConfigMaps in Advance

Add ConfigMaps in advance to be used when creating a ConfigMap-type data source via the NVIDIA Run:ai UI.

### **Cluster Scope**

1. Locate the ConfigMap in the NVIDIA Run:ai namespace (`runai`)
2. To authorize NVIDIA Run:ai to use the ConfigMap, label it: `run.ai/cluster-wide: "true"`
3. The ConfigMap must have a label of `run.ai/resource: <resource-name>`

The ConfigMap is now displayed for that scope in the list of existing ConfigMaps.

### **Department Scope**

1. Locate the ConfigMap in the NVIDIA Run:ai namespace (`runai`)
2. To authorize NVIDIA Run:ai to use the ConfigMap, label it: `run.ai/department: "<department-id>"`
3. The ConfigMap must have a label of `run.ai/resource: <resource-name>`

The ConfigMap is now displayed for that scope in the list of existing ConfigMaps.

### **Project Scope**

1. Locate the ConfigMap in the project's namespace
2. The ConfigMap must have a label of `run.ai/resource: <resource-name>`

The ConfigMap is now displayed for that scope in the list of existing ConfigMaps.

## Using CLI

To view the available actions on data sources, see the data source [CLI v2 reference](../../reference/cli/runai/runai_datasource.md).

## Using API

To view the available actions, go to the [Data sources](api:main-api:GET/api/v1/asset/datasource) API reference.
