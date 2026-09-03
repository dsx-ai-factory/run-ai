# DistributedPolicyDefaultsV2Worker

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to [**NullableAnnotationsDefaults**](AnnotationsDefaults.md) |  | [optional] 
**Args** | Pointer to **NullableString** | Arguments to the command that the container running the workload executes. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (completed or failed) will be automatically deleted. The default is 30 days. Log retention is managed separately. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed (not applicable to Inference workloads). The default value is 6. | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**Command** | Pointer to **NullableString** | A command to the server as the entry point of the container running the workload. | [optional] 
**Compute** | Pointer to [**NullableSupersetDefaultsAllOfCompute**](SupersetDefaultsAllOfCompute.md) |  | [optional] 
**CreateHomeDir** | Pointer to **NullableBool** | When set to &#x60;true&#x60;, creates a home directory for the container. | [optional] 
**EnvironmentVariables** | Pointer to [**NullableEnvironmentVariablesDefaults**](EnvironmentVariablesDefaults.md) |  | [optional] 
**ExposedUrls** | Pointer to [**NullableExposedUrlsDefaults**](ExposedUrlsDefaults.md) |  | [optional] 
**Image** | Pointer to **NullableString** | Docker image name. For more information, see [Images](https://kubernetes.io/docs/concepts/containers/images). The image name is mandatory for creating a workload. | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicy**](ImagePullPolicy.md) |  | [optional] 
**ImagePullSecrets** | Pointer to [**NullableImagePullSecretsDefaults**](ImagePullSecretsDefaults.md) |  | [optional] 
**Labels** | Pointer to [**NullableLabelsDefaults**](LabelsDefaults.md) |  | [optional] 
**NodeAffinityRequired** | Pointer to [**NullableNodeAffinityRequired**](NodeAffinityRequired.md) |  | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**NodeType** | Pointer to **NullableString** | Deprecated: use a workload policy node selector to target node types instead. Nodes (machines), or a group of nodes on which the workload will run. To use this feature, your Administrator will need to label nodes. For more information, see [Group Nodes](https://docs.run.ai/latest/admin/researcher-setup/limit-to-node-group). When using this flag with with Project-based affinity, it refines the list of allowable node groups set in the Project. For more information, see [Projects](https://docshub.run.ai/guides/platform-management/aiinitiatives/organization/projects). | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinity**](PodAffinity.md) |  | [optional] 
**Ports** | Pointer to [**NullablePortsDefaults**](PortsDefaults.md) |  | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**PriorityClass** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Probes** | Pointer to [**NullableProbes**](Probes.md) |  | [optional] 
**RelatedUrls** | Pointer to [**NullableRelatedUrlsDefaults**](RelatedUrlsDefaults.md) |  | [optional] 
**RestartPolicy** | Pointer to [**NullableRestartPolicy**](RestartPolicy.md) |  | [optional] 
**Security** | Pointer to [**NullableSupersetDefaultsAllOfSecurity**](SupersetDefaultsAllOfSecurity.md) |  | [optional] 
**Stdin** | Pointer to **NullableBool** | Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF | [optional] 
**Storage** | Pointer to [**NullableSupersetDefaultsAllOfStorage**](SupersetDefaultsAllOfStorage.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to **NullableBool** | Indicates if the job should be terminated by the system after it has been preempted. | [optional] 
**TerminationGracePeriodSeconds** | Pointer to **NullableInt64** | Duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). | [optional] 
**Tolerations** | Pointer to [**NullableTolerationsDefaults**](TolerationsDefaults.md) |  | [optional] 
**Tty** | Pointer to **NullableBool** | Whether this container should allocate a TTY for itself, also requires &#39;stdin&#39; to be true. | [optional] 
**WorkingDir** | Pointer to **NullableString** | Container&#39;s working directory. If not specified, the container runtime default will be used. This may be configured in the container image. | [optional] 
**CleanPodPolicy** | Pointer to [**NullableCleanPodPolicy**](CleanPodPolicy.md) |  | [optional] 
**DistributedFramework** | Pointer to [**NullableDistributedFramework**](DistributedFramework.md) |  | [optional] 
**MaxReplicas** | Pointer to **NullableInt32** | the upper limit for the number of worker pods that can be set by the autoscaler. Cannot be smaller than MinReplicas. (applicable only for PyTorch) | [optional] 
**MinReplicas** | Pointer to **NullableInt32** | the lower limit for the number of worker pods to which the training job can scale down. (applicable only for PyTorch) | [optional] 
**MpiLauncherCreationPolicy** | Pointer to [**NullableMpiLauncherCreationPolicy**](MpiLauncherCreationPolicy.md) |  | [optional] 
**NumWorkers** | Pointer to **NullableInt32** | the number of workers that will be allocated for running the workload. | [optional] 
**SlotsPerWorker** | Pointer to **NullableInt32** | Specifies the number of slots per worker used in hostfile. Defaults to 1. (applicable only for MPI) | [optional] [default to 1]
**SshAuthMountPath** | Pointer to **NullableString** | Specifies the directory where SSH keys are mounted. (applicable only for MPI) | [optional] [default to "/root/.ssh"]

## Methods

### NewDistributedPolicyDefaultsV2Worker

`func NewDistributedPolicyDefaultsV2Worker() *DistributedPolicyDefaultsV2Worker`

NewDistributedPolicyDefaultsV2Worker instantiates a new DistributedPolicyDefaultsV2Worker object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedPolicyDefaultsV2WorkerWithDefaults

`func NewDistributedPolicyDefaultsV2WorkerWithDefaults() *DistributedPolicyDefaultsV2Worker`

NewDistributedPolicyDefaultsV2WorkerWithDefaults instantiates a new DistributedPolicyDefaultsV2Worker object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *DistributedPolicyDefaultsV2Worker) GetAnnotations() AnnotationsDefaults`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetAnnotationsOk() (*AnnotationsDefaults, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *DistributedPolicyDefaultsV2Worker) SetAnnotations(v AnnotationsDefaults)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *DistributedPolicyDefaultsV2Worker) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *DistributedPolicyDefaultsV2Worker) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *DistributedPolicyDefaultsV2Worker) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *DistributedPolicyDefaultsV2Worker) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *DistributedPolicyDefaultsV2Worker) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *DistributedPolicyDefaultsV2Worker) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *DistributedPolicyDefaultsV2Worker) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *DistributedPolicyDefaultsV2Worker) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *DistributedPolicyDefaultsV2Worker) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *DistributedPolicyDefaultsV2Worker) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *DistributedPolicyDefaultsV2Worker) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *DistributedPolicyDefaultsV2Worker) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *DistributedPolicyDefaultsV2Worker) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *DistributedPolicyDefaultsV2Worker) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *DistributedPolicyDefaultsV2Worker) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *DistributedPolicyDefaultsV2Worker) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetCategory

`func (o *DistributedPolicyDefaultsV2Worker) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *DistributedPolicyDefaultsV2Worker) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *DistributedPolicyDefaultsV2Worker) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *DistributedPolicyDefaultsV2Worker) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *DistributedPolicyDefaultsV2Worker) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *DistributedPolicyDefaultsV2Worker) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCommand

`func (o *DistributedPolicyDefaultsV2Worker) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *DistributedPolicyDefaultsV2Worker) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *DistributedPolicyDefaultsV2Worker) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *DistributedPolicyDefaultsV2Worker) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *DistributedPolicyDefaultsV2Worker) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *DistributedPolicyDefaultsV2Worker) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetCompute

`func (o *DistributedPolicyDefaultsV2Worker) GetCompute() SupersetDefaultsAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *DistributedPolicyDefaultsV2Worker) GetComputeOk() (*SupersetDefaultsAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *DistributedPolicyDefaultsV2Worker) SetCompute(v SupersetDefaultsAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *DistributedPolicyDefaultsV2Worker) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *DistributedPolicyDefaultsV2Worker) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *DistributedPolicyDefaultsV2Worker) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetCreateHomeDir

`func (o *DistributedPolicyDefaultsV2Worker) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *DistributedPolicyDefaultsV2Worker) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *DistributedPolicyDefaultsV2Worker) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *DistributedPolicyDefaultsV2Worker) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *DistributedPolicyDefaultsV2Worker) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *DistributedPolicyDefaultsV2Worker) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetEnvironmentVariables

`func (o *DistributedPolicyDefaultsV2Worker) GetEnvironmentVariables() EnvironmentVariablesDefaults`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *DistributedPolicyDefaultsV2Worker) GetEnvironmentVariablesOk() (*EnvironmentVariablesDefaults, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *DistributedPolicyDefaultsV2Worker) SetEnvironmentVariables(v EnvironmentVariablesDefaults)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *DistributedPolicyDefaultsV2Worker) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *DistributedPolicyDefaultsV2Worker) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *DistributedPolicyDefaultsV2Worker) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetExposedUrls

`func (o *DistributedPolicyDefaultsV2Worker) GetExposedUrls() ExposedUrlsDefaults`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetExposedUrlsOk() (*ExposedUrlsDefaults, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *DistributedPolicyDefaultsV2Worker) SetExposedUrls(v ExposedUrlsDefaults)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *DistributedPolicyDefaultsV2Worker) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *DistributedPolicyDefaultsV2Worker) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetImage

`func (o *DistributedPolicyDefaultsV2Worker) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *DistributedPolicyDefaultsV2Worker) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *DistributedPolicyDefaultsV2Worker) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *DistributedPolicyDefaultsV2Worker) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *DistributedPolicyDefaultsV2Worker) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *DistributedPolicyDefaultsV2Worker) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *DistributedPolicyDefaultsV2Worker) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *DistributedPolicyDefaultsV2Worker) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *DistributedPolicyDefaultsV2Worker) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *DistributedPolicyDefaultsV2Worker) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *DistributedPolicyDefaultsV2Worker) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *DistributedPolicyDefaultsV2Worker) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *DistributedPolicyDefaultsV2Worker) GetImagePullSecrets() ImagePullSecretsDefaults`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetImagePullSecretsOk() (*ImagePullSecretsDefaults, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *DistributedPolicyDefaultsV2Worker) SetImagePullSecrets(v ImagePullSecretsDefaults)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *DistributedPolicyDefaultsV2Worker) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *DistributedPolicyDefaultsV2Worker) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *DistributedPolicyDefaultsV2Worker) GetLabels() LabelsDefaults`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetLabelsOk() (*LabelsDefaults, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *DistributedPolicyDefaultsV2Worker) SetLabels(v LabelsDefaults)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *DistributedPolicyDefaultsV2Worker) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *DistributedPolicyDefaultsV2Worker) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetNodeAffinityRequired

`func (o *DistributedPolicyDefaultsV2Worker) GetNodeAffinityRequired() NodeAffinityRequired`

GetNodeAffinityRequired returns the NodeAffinityRequired field if non-nil, zero value otherwise.

### GetNodeAffinityRequiredOk

`func (o *DistributedPolicyDefaultsV2Worker) GetNodeAffinityRequiredOk() (*NodeAffinityRequired, bool)`

GetNodeAffinityRequiredOk returns a tuple with the NodeAffinityRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinityRequired

`func (o *DistributedPolicyDefaultsV2Worker) SetNodeAffinityRequired(v NodeAffinityRequired)`

SetNodeAffinityRequired sets NodeAffinityRequired field to given value.

### HasNodeAffinityRequired

`func (o *DistributedPolicyDefaultsV2Worker) HasNodeAffinityRequired() bool`

HasNodeAffinityRequired returns a boolean if a field has been set.

### SetNodeAffinityRequiredNil

`func (o *DistributedPolicyDefaultsV2Worker) SetNodeAffinityRequiredNil(b bool)`

 SetNodeAffinityRequiredNil sets the value for NodeAffinityRequired to be an explicit nil

### UnsetNodeAffinityRequired
`func (o *DistributedPolicyDefaultsV2Worker) UnsetNodeAffinityRequired()`

UnsetNodeAffinityRequired ensures that no value is present for NodeAffinityRequired, not even an explicit nil
### GetNodePools

`func (o *DistributedPolicyDefaultsV2Worker) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *DistributedPolicyDefaultsV2Worker) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *DistributedPolicyDefaultsV2Worker) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *DistributedPolicyDefaultsV2Worker) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetNodeType

`func (o *DistributedPolicyDefaultsV2Worker) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *DistributedPolicyDefaultsV2Worker) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *DistributedPolicyDefaultsV2Worker) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *DistributedPolicyDefaultsV2Worker) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *DistributedPolicyDefaultsV2Worker) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *DistributedPolicyDefaultsV2Worker) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetPodAffinity

`func (o *DistributedPolicyDefaultsV2Worker) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *DistributedPolicyDefaultsV2Worker) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *DistributedPolicyDefaultsV2Worker) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *DistributedPolicyDefaultsV2Worker) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *DistributedPolicyDefaultsV2Worker) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *DistributedPolicyDefaultsV2Worker) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetPorts

`func (o *DistributedPolicyDefaultsV2Worker) GetPorts() PortsDefaults`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetPortsOk() (*PortsDefaults, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *DistributedPolicyDefaultsV2Worker) SetPorts(v PortsDefaults)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *DistributedPolicyDefaultsV2Worker) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *DistributedPolicyDefaultsV2Worker) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetPreemptibility

`func (o *DistributedPolicyDefaultsV2Worker) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *DistributedPolicyDefaultsV2Worker) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *DistributedPolicyDefaultsV2Worker) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *DistributedPolicyDefaultsV2Worker) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *DistributedPolicyDefaultsV2Worker) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *DistributedPolicyDefaultsV2Worker) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *DistributedPolicyDefaultsV2Worker) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *DistributedPolicyDefaultsV2Worker) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *DistributedPolicyDefaultsV2Worker) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *DistributedPolicyDefaultsV2Worker) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *DistributedPolicyDefaultsV2Worker) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *DistributedPolicyDefaultsV2Worker) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *DistributedPolicyDefaultsV2Worker) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *DistributedPolicyDefaultsV2Worker) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *DistributedPolicyDefaultsV2Worker) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *DistributedPolicyDefaultsV2Worker) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *DistributedPolicyDefaultsV2Worker) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *DistributedPolicyDefaultsV2Worker) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetRelatedUrls

`func (o *DistributedPolicyDefaultsV2Worker) GetRelatedUrls() RelatedUrlsDefaults`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetRelatedUrlsOk() (*RelatedUrlsDefaults, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *DistributedPolicyDefaultsV2Worker) SetRelatedUrls(v RelatedUrlsDefaults)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *DistributedPolicyDefaultsV2Worker) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *DistributedPolicyDefaultsV2Worker) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetRestartPolicy

`func (o *DistributedPolicyDefaultsV2Worker) GetRestartPolicy() RestartPolicy`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *DistributedPolicyDefaultsV2Worker) GetRestartPolicyOk() (*RestartPolicy, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *DistributedPolicyDefaultsV2Worker) SetRestartPolicy(v RestartPolicy)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *DistributedPolicyDefaultsV2Worker) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *DistributedPolicyDefaultsV2Worker) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *DistributedPolicyDefaultsV2Worker) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetSecurity

`func (o *DistributedPolicyDefaultsV2Worker) GetSecurity() SupersetDefaultsAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *DistributedPolicyDefaultsV2Worker) GetSecurityOk() (*SupersetDefaultsAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *DistributedPolicyDefaultsV2Worker) SetSecurity(v SupersetDefaultsAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *DistributedPolicyDefaultsV2Worker) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *DistributedPolicyDefaultsV2Worker) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *DistributedPolicyDefaultsV2Worker) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetStdin

`func (o *DistributedPolicyDefaultsV2Worker) GetStdin() bool`

GetStdin returns the Stdin field if non-nil, zero value otherwise.

### GetStdinOk

`func (o *DistributedPolicyDefaultsV2Worker) GetStdinOk() (*bool, bool)`

GetStdinOk returns a tuple with the Stdin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdin

`func (o *DistributedPolicyDefaultsV2Worker) SetStdin(v bool)`

SetStdin sets Stdin field to given value.

### HasStdin

`func (o *DistributedPolicyDefaultsV2Worker) HasStdin() bool`

HasStdin returns a boolean if a field has been set.

### SetStdinNil

`func (o *DistributedPolicyDefaultsV2Worker) SetStdinNil(b bool)`

 SetStdinNil sets the value for Stdin to be an explicit nil

### UnsetStdin
`func (o *DistributedPolicyDefaultsV2Worker) UnsetStdin()`

UnsetStdin ensures that no value is present for Stdin, not even an explicit nil
### GetStorage

`func (o *DistributedPolicyDefaultsV2Worker) GetStorage() SupersetDefaultsAllOfStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *DistributedPolicyDefaultsV2Worker) GetStorageOk() (*SupersetDefaultsAllOfStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *DistributedPolicyDefaultsV2Worker) SetStorage(v SupersetDefaultsAllOfStorage)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *DistributedPolicyDefaultsV2Worker) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *DistributedPolicyDefaultsV2Worker) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *DistributedPolicyDefaultsV2Worker) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *DistributedPolicyDefaultsV2Worker) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *DistributedPolicyDefaultsV2Worker) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *DistributedPolicyDefaultsV2Worker) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *DistributedPolicyDefaultsV2Worker) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *DistributedPolicyDefaultsV2Worker) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *DistributedPolicyDefaultsV2Worker) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetTerminationGracePeriodSeconds

`func (o *DistributedPolicyDefaultsV2Worker) GetTerminationGracePeriodSeconds() int64`

GetTerminationGracePeriodSeconds returns the TerminationGracePeriodSeconds field if non-nil, zero value otherwise.

### GetTerminationGracePeriodSecondsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetTerminationGracePeriodSecondsOk() (*int64, bool)`

GetTerminationGracePeriodSecondsOk returns a tuple with the TerminationGracePeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminationGracePeriodSeconds

`func (o *DistributedPolicyDefaultsV2Worker) SetTerminationGracePeriodSeconds(v int64)`

SetTerminationGracePeriodSeconds sets TerminationGracePeriodSeconds field to given value.

### HasTerminationGracePeriodSeconds

`func (o *DistributedPolicyDefaultsV2Worker) HasTerminationGracePeriodSeconds() bool`

HasTerminationGracePeriodSeconds returns a boolean if a field has been set.

### SetTerminationGracePeriodSecondsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetTerminationGracePeriodSecondsNil(b bool)`

 SetTerminationGracePeriodSecondsNil sets the value for TerminationGracePeriodSeconds to be an explicit nil

### UnsetTerminationGracePeriodSeconds
`func (o *DistributedPolicyDefaultsV2Worker) UnsetTerminationGracePeriodSeconds()`

UnsetTerminationGracePeriodSeconds ensures that no value is present for TerminationGracePeriodSeconds, not even an explicit nil
### GetTolerations

`func (o *DistributedPolicyDefaultsV2Worker) GetTolerations() TolerationsDefaults`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *DistributedPolicyDefaultsV2Worker) GetTolerationsOk() (*TolerationsDefaults, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *DistributedPolicyDefaultsV2Worker) SetTolerations(v TolerationsDefaults)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *DistributedPolicyDefaultsV2Worker) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *DistributedPolicyDefaultsV2Worker) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *DistributedPolicyDefaultsV2Worker) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTty

`func (o *DistributedPolicyDefaultsV2Worker) GetTty() bool`

GetTty returns the Tty field if non-nil, zero value otherwise.

### GetTtyOk

`func (o *DistributedPolicyDefaultsV2Worker) GetTtyOk() (*bool, bool)`

GetTtyOk returns a tuple with the Tty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTty

`func (o *DistributedPolicyDefaultsV2Worker) SetTty(v bool)`

SetTty sets Tty field to given value.

### HasTty

`func (o *DistributedPolicyDefaultsV2Worker) HasTty() bool`

HasTty returns a boolean if a field has been set.

### SetTtyNil

`func (o *DistributedPolicyDefaultsV2Worker) SetTtyNil(b bool)`

 SetTtyNil sets the value for Tty to be an explicit nil

### UnsetTty
`func (o *DistributedPolicyDefaultsV2Worker) UnsetTty()`

UnsetTty ensures that no value is present for Tty, not even an explicit nil
### GetWorkingDir

`func (o *DistributedPolicyDefaultsV2Worker) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *DistributedPolicyDefaultsV2Worker) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *DistributedPolicyDefaultsV2Worker) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *DistributedPolicyDefaultsV2Worker) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *DistributedPolicyDefaultsV2Worker) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *DistributedPolicyDefaultsV2Worker) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCleanPodPolicy

`func (o *DistributedPolicyDefaultsV2Worker) GetCleanPodPolicy() CleanPodPolicy`

GetCleanPodPolicy returns the CleanPodPolicy field if non-nil, zero value otherwise.

### GetCleanPodPolicyOk

`func (o *DistributedPolicyDefaultsV2Worker) GetCleanPodPolicyOk() (*CleanPodPolicy, bool)`

GetCleanPodPolicyOk returns a tuple with the CleanPodPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCleanPodPolicy

`func (o *DistributedPolicyDefaultsV2Worker) SetCleanPodPolicy(v CleanPodPolicy)`

SetCleanPodPolicy sets CleanPodPolicy field to given value.

### HasCleanPodPolicy

`func (o *DistributedPolicyDefaultsV2Worker) HasCleanPodPolicy() bool`

HasCleanPodPolicy returns a boolean if a field has been set.

### SetCleanPodPolicyNil

`func (o *DistributedPolicyDefaultsV2Worker) SetCleanPodPolicyNil(b bool)`

 SetCleanPodPolicyNil sets the value for CleanPodPolicy to be an explicit nil

### UnsetCleanPodPolicy
`func (o *DistributedPolicyDefaultsV2Worker) UnsetCleanPodPolicy()`

UnsetCleanPodPolicy ensures that no value is present for CleanPodPolicy, not even an explicit nil
### GetDistributedFramework

`func (o *DistributedPolicyDefaultsV2Worker) GetDistributedFramework() DistributedFramework`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *DistributedPolicyDefaultsV2Worker) GetDistributedFrameworkOk() (*DistributedFramework, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *DistributedPolicyDefaultsV2Worker) SetDistributedFramework(v DistributedFramework)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *DistributedPolicyDefaultsV2Worker) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### SetDistributedFrameworkNil

`func (o *DistributedPolicyDefaultsV2Worker) SetDistributedFrameworkNil(b bool)`

 SetDistributedFrameworkNil sets the value for DistributedFramework to be an explicit nil

### UnsetDistributedFramework
`func (o *DistributedPolicyDefaultsV2Worker) UnsetDistributedFramework()`

UnsetDistributedFramework ensures that no value is present for DistributedFramework, not even an explicit nil
### GetMaxReplicas

`func (o *DistributedPolicyDefaultsV2Worker) GetMaxReplicas() int32`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *DistributedPolicyDefaultsV2Worker) GetMaxReplicasOk() (*int32, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *DistributedPolicyDefaultsV2Worker) SetMaxReplicas(v int32)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *DistributedPolicyDefaultsV2Worker) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *DistributedPolicyDefaultsV2Worker) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *DistributedPolicyDefaultsV2Worker) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil
### GetMinReplicas

`func (o *DistributedPolicyDefaultsV2Worker) GetMinReplicas() int32`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *DistributedPolicyDefaultsV2Worker) GetMinReplicasOk() (*int32, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *DistributedPolicyDefaultsV2Worker) SetMinReplicas(v int32)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *DistributedPolicyDefaultsV2Worker) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *DistributedPolicyDefaultsV2Worker) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *DistributedPolicyDefaultsV2Worker) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetMpiLauncherCreationPolicy

`func (o *DistributedPolicyDefaultsV2Worker) GetMpiLauncherCreationPolicy() MpiLauncherCreationPolicy`

GetMpiLauncherCreationPolicy returns the MpiLauncherCreationPolicy field if non-nil, zero value otherwise.

### GetMpiLauncherCreationPolicyOk

`func (o *DistributedPolicyDefaultsV2Worker) GetMpiLauncherCreationPolicyOk() (*MpiLauncherCreationPolicy, bool)`

GetMpiLauncherCreationPolicyOk returns a tuple with the MpiLauncherCreationPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMpiLauncherCreationPolicy

`func (o *DistributedPolicyDefaultsV2Worker) SetMpiLauncherCreationPolicy(v MpiLauncherCreationPolicy)`

SetMpiLauncherCreationPolicy sets MpiLauncherCreationPolicy field to given value.

### HasMpiLauncherCreationPolicy

`func (o *DistributedPolicyDefaultsV2Worker) HasMpiLauncherCreationPolicy() bool`

HasMpiLauncherCreationPolicy returns a boolean if a field has been set.

### SetMpiLauncherCreationPolicyNil

`func (o *DistributedPolicyDefaultsV2Worker) SetMpiLauncherCreationPolicyNil(b bool)`

 SetMpiLauncherCreationPolicyNil sets the value for MpiLauncherCreationPolicy to be an explicit nil

### UnsetMpiLauncherCreationPolicy
`func (o *DistributedPolicyDefaultsV2Worker) UnsetMpiLauncherCreationPolicy()`

UnsetMpiLauncherCreationPolicy ensures that no value is present for MpiLauncherCreationPolicy, not even an explicit nil
### GetNumWorkers

`func (o *DistributedPolicyDefaultsV2Worker) GetNumWorkers() int32`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *DistributedPolicyDefaultsV2Worker) GetNumWorkersOk() (*int32, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *DistributedPolicyDefaultsV2Worker) SetNumWorkers(v int32)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *DistributedPolicyDefaultsV2Worker) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### SetNumWorkersNil

`func (o *DistributedPolicyDefaultsV2Worker) SetNumWorkersNil(b bool)`

 SetNumWorkersNil sets the value for NumWorkers to be an explicit nil

### UnsetNumWorkers
`func (o *DistributedPolicyDefaultsV2Worker) UnsetNumWorkers()`

UnsetNumWorkers ensures that no value is present for NumWorkers, not even an explicit nil
### GetSlotsPerWorker

`func (o *DistributedPolicyDefaultsV2Worker) GetSlotsPerWorker() int32`

GetSlotsPerWorker returns the SlotsPerWorker field if non-nil, zero value otherwise.

### GetSlotsPerWorkerOk

`func (o *DistributedPolicyDefaultsV2Worker) GetSlotsPerWorkerOk() (*int32, bool)`

GetSlotsPerWorkerOk returns a tuple with the SlotsPerWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSlotsPerWorker

`func (o *DistributedPolicyDefaultsV2Worker) SetSlotsPerWorker(v int32)`

SetSlotsPerWorker sets SlotsPerWorker field to given value.

### HasSlotsPerWorker

`func (o *DistributedPolicyDefaultsV2Worker) HasSlotsPerWorker() bool`

HasSlotsPerWorker returns a boolean if a field has been set.

### SetSlotsPerWorkerNil

`func (o *DistributedPolicyDefaultsV2Worker) SetSlotsPerWorkerNil(b bool)`

 SetSlotsPerWorkerNil sets the value for SlotsPerWorker to be an explicit nil

### UnsetSlotsPerWorker
`func (o *DistributedPolicyDefaultsV2Worker) UnsetSlotsPerWorker()`

UnsetSlotsPerWorker ensures that no value is present for SlotsPerWorker, not even an explicit nil
### GetSshAuthMountPath

`func (o *DistributedPolicyDefaultsV2Worker) GetSshAuthMountPath() string`

GetSshAuthMountPath returns the SshAuthMountPath field if non-nil, zero value otherwise.

### GetSshAuthMountPathOk

`func (o *DistributedPolicyDefaultsV2Worker) GetSshAuthMountPathOk() (*string, bool)`

GetSshAuthMountPathOk returns a tuple with the SshAuthMountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshAuthMountPath

`func (o *DistributedPolicyDefaultsV2Worker) SetSshAuthMountPath(v string)`

SetSshAuthMountPath sets SshAuthMountPath field to given value.

### HasSshAuthMountPath

`func (o *DistributedPolicyDefaultsV2Worker) HasSshAuthMountPath() bool`

HasSshAuthMountPath returns a boolean if a field has been set.

### SetSshAuthMountPathNil

`func (o *DistributedPolicyDefaultsV2Worker) SetSshAuthMountPathNil(b bool)`

 SetSshAuthMountPathNil sets the value for SshAuthMountPath to be an explicit nil

### UnsetSshAuthMountPath
`func (o *DistributedPolicyDefaultsV2Worker) UnsetSshAuthMountPath()`

UnsetSshAuthMountPath ensures that no value is present for SshAuthMountPath, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


