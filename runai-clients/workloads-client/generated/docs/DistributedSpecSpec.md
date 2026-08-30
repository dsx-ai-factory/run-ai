# DistributedSpecSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) | Set of annotations to populate into the container running the workload. | [optional] 
**Args** | Pointer to **NullableString** | Arguments to the command that the container running the workload executes. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (completed or failed) will be automatically deleted. The default is 30 days. Log retention is managed separately. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed (not applicable to Inference workloads). The default value is 6. | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**CleanPodPolicy** | Pointer to [**NullableCleanPodPolicy**](CleanPodPolicy.md) |  | [optional] 
**Command** | Pointer to **NullableString** | A command to the server as the entry point of the container running the workload. | [optional] 
**Compute** | Pointer to [**NullableWorkspaceSpecSpecAllOfCompute**](WorkspaceSpecSpecAllOfCompute.md) |  | [optional] 
**CreateHomeDir** | Pointer to **NullableBool** | When set to &#x60;true&#x60;, creates a home directory for the container. | [optional] 
**DistributedFramework** | Pointer to [**NullableDistributedFramework**](DistributedFramework.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**[]EnvironmentVariable**](EnvironmentVariable.md) | Set of environment variables to populate into the container running the workload. | [optional] 
**ExposedUrls** | Pointer to [**[]ExposedUrl**](ExposedUrl.md) | Set of container ports that the workload exposes via URLs. | [optional] 
**Image** | Pointer to **NullableString** | Docker image name. For more information, see [Images](https://kubernetes.io/docs/concepts/containers/images). The image name is mandatory for creating a workload. | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicy**](ImagePullPolicy.md) |  | [optional] 
**ImagePullSecrets** | Pointer to [**[]ImagePullSecret**](ImagePullSecret.md) | A list of references to Kubernetes secrets in the same namespace used for pulling container images. | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) | Set of labels to populate into the container running the workload. | [optional] 
**MaxReplicas** | Pointer to **NullableInt32** | the upper limit for the number of worker pods that can be set by the autoscaler. Cannot be smaller than MinReplicas. (applicable only for PyTorch) | [optional] 
**MinReplicas** | Pointer to **NullableInt32** | the lower limit for the number of worker pods to which the training job can scale down. (applicable only for PyTorch) | [optional] 
**MpiLauncherCreationPolicy** | Pointer to [**NullableMpiLauncherCreationPolicy**](MpiLauncherCreationPolicy.md) |  | [optional] 
**NodeAffinityRequired** | Pointer to [**NullableNodeAffinityRequired**](NodeAffinityRequired.md) |  | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workload on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**NodeType** | Pointer to **NullableString** | Deprecated: use a workload policy node selector to target node types instead. Nodes (machines), or a group of nodes on which the workload will run. To use this feature, your Administrator will need to label nodes. For more information, see [Group Nodes](https://docs.run.ai/latest/admin/researcher-setup/limit-to-node-group). When using this flag with with Project-based affinity, it refines the list of allowable node groups set in the Project. For more information, see [Projects](https://docshub.run.ai/guides/platform-management/aiinitiatives/organization/projects). | [optional] 
**NumWorkers** | Pointer to **NullableInt32** | the number of workers that will be allocated for running the workload. | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinity**](PodAffinity.md) |  | [optional] 
**Ports** | Pointer to [**[]Port**](Port.md) | Set of container ports that the workload exposes. | [optional] 
**Preemptibility** | Pointer to [**NullablePreemptibility**](Preemptibility.md) |  | [optional] 
**PriorityClass** | Pointer to **NullableString** | Specifies the priority class for the workload, which determines its scheduling behavior. Valid values are: very-low, low, medium-low, medium, medium-high, high, and very-high. Each workload type has a default priority. To view the default priority for each workload type, use the GET /workload-types endpoint. Once you change the priority from the default value defined for that workload type, the preemptibility field is not automatically updated. Make sure to set the desired preemptibility value. | [optional] 
**Probes** | Pointer to [**NullableProbes**](Probes.md) |  | [optional] 
**RelatedUrls** | Pointer to [**[]RelatedUrl**](RelatedUrl.md) | Set of URLs that are related to the workload. | [optional] 
**RestartPolicy** | Pointer to [**NullableRestartPolicy**](RestartPolicy.md) |  | [optional] 
**Security** | Pointer to [**NullableWorkspaceSpecSpecAllOfSecurity**](WorkspaceSpecSpecAllOfSecurity.md) |  | [optional] 
**SegmentSize** | Pointer to **NullableInt32** | The number of workers per topology segment. Workers in the same segment will be placed on nodes that are as close as possible, reducing latency between them. Set this to match your application&#39;s parallel segment size. Omit this field to disable segmentation.  | [optional] 
**SlotsPerWorker** | Pointer to **NullableInt32** | Specifies the number of slots per worker used in hostfile. Defaults to 1. (applicable only for MPI) | [optional] [default to 1]
**SshAuthMountPath** | Pointer to **NullableString** | Specifies the directory where SSH keys are mounted. (applicable only for MPI) | [optional] [default to "/root/.ssh"]
**Stdin** | Pointer to **NullableBool** | Whether this container should allocate a buffer for stdin in the container runtime. If this is not set, reads from stdin in the container will always result in EOF | [optional] 
**Storage** | Pointer to [**NullableWorkspaceSpecSpecAllOfStorage**](WorkspaceSpecSpecAllOfStorage.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to **NullableBool** | Indicates if the job should be terminated by the system after it has been preempted. | [optional] 
**TerminationGracePeriodSeconds** | Pointer to **NullableInt64** | Duration in seconds the pod needs to terminate gracefully upon probe failure. The grace period is the duration in seconds after the processes running in the pod are sent a termination signal and the time when the processes are forcibly halted with a kill signal. Set this value longer than the expected cleanup time for your process. Value must be non-negative integer. The value zero indicates stop immediately via the kill signal (no opportunity to shut down). | [optional] 
**Tolerations** | Pointer to [**[]Toleration**](Toleration.md) | Set of tolerations to apply to the workload. | [optional] 
**Tty** | Pointer to **NullableBool** | Whether this container should allocate a TTY for itself, also requires &#39;stdin&#39; to be true. | [optional] 
**WorkingDir** | Pointer to **NullableString** | Container&#39;s working directory. If not specified, the container runtime default will be used. This may be configured in the container image. | [optional] 

## Methods

### NewDistributedSpecSpec

`func NewDistributedSpecSpec() *DistributedSpecSpec`

NewDistributedSpecSpec instantiates a new DistributedSpecSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedSpecSpecWithDefaults

`func NewDistributedSpecSpecWithDefaults() *DistributedSpecSpec`

NewDistributedSpecSpecWithDefaults instantiates a new DistributedSpecSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *DistributedSpecSpec) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *DistributedSpecSpec) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *DistributedSpecSpec) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *DistributedSpecSpec) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *DistributedSpecSpec) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *DistributedSpecSpec) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *DistributedSpecSpec) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *DistributedSpecSpec) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *DistributedSpecSpec) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *DistributedSpecSpec) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *DistributedSpecSpec) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *DistributedSpecSpec) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *DistributedSpecSpec) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *DistributedSpecSpec) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *DistributedSpecSpec) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *DistributedSpecSpec) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *DistributedSpecSpec) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *DistributedSpecSpec) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *DistributedSpecSpec) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *DistributedSpecSpec) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *DistributedSpecSpec) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *DistributedSpecSpec) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *DistributedSpecSpec) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *DistributedSpecSpec) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetCategory

`func (o *DistributedSpecSpec) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *DistributedSpecSpec) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *DistributedSpecSpec) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *DistributedSpecSpec) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *DistributedSpecSpec) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *DistributedSpecSpec) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCleanPodPolicy

`func (o *DistributedSpecSpec) GetCleanPodPolicy() CleanPodPolicy`

GetCleanPodPolicy returns the CleanPodPolicy field if non-nil, zero value otherwise.

### GetCleanPodPolicyOk

`func (o *DistributedSpecSpec) GetCleanPodPolicyOk() (*CleanPodPolicy, bool)`

GetCleanPodPolicyOk returns a tuple with the CleanPodPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCleanPodPolicy

`func (o *DistributedSpecSpec) SetCleanPodPolicy(v CleanPodPolicy)`

SetCleanPodPolicy sets CleanPodPolicy field to given value.

### HasCleanPodPolicy

`func (o *DistributedSpecSpec) HasCleanPodPolicy() bool`

HasCleanPodPolicy returns a boolean if a field has been set.

### SetCleanPodPolicyNil

`func (o *DistributedSpecSpec) SetCleanPodPolicyNil(b bool)`

 SetCleanPodPolicyNil sets the value for CleanPodPolicy to be an explicit nil

### UnsetCleanPodPolicy
`func (o *DistributedSpecSpec) UnsetCleanPodPolicy()`

UnsetCleanPodPolicy ensures that no value is present for CleanPodPolicy, not even an explicit nil
### GetCommand

`func (o *DistributedSpecSpec) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *DistributedSpecSpec) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *DistributedSpecSpec) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *DistributedSpecSpec) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *DistributedSpecSpec) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *DistributedSpecSpec) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetCompute

`func (o *DistributedSpecSpec) GetCompute() WorkspaceSpecSpecAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *DistributedSpecSpec) GetComputeOk() (*WorkspaceSpecSpecAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *DistributedSpecSpec) SetCompute(v WorkspaceSpecSpecAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *DistributedSpecSpec) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *DistributedSpecSpec) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *DistributedSpecSpec) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetCreateHomeDir

`func (o *DistributedSpecSpec) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *DistributedSpecSpec) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *DistributedSpecSpec) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *DistributedSpecSpec) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *DistributedSpecSpec) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *DistributedSpecSpec) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetDistributedFramework

`func (o *DistributedSpecSpec) GetDistributedFramework() DistributedFramework`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *DistributedSpecSpec) GetDistributedFrameworkOk() (*DistributedFramework, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *DistributedSpecSpec) SetDistributedFramework(v DistributedFramework)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *DistributedSpecSpec) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### SetDistributedFrameworkNil

`func (o *DistributedSpecSpec) SetDistributedFrameworkNil(b bool)`

 SetDistributedFrameworkNil sets the value for DistributedFramework to be an explicit nil

### UnsetDistributedFramework
`func (o *DistributedSpecSpec) UnsetDistributedFramework()`

UnsetDistributedFramework ensures that no value is present for DistributedFramework, not even an explicit nil
### GetEnvironmentVariables

`func (o *DistributedSpecSpec) GetEnvironmentVariables() []EnvironmentVariable`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *DistributedSpecSpec) GetEnvironmentVariablesOk() (*[]EnvironmentVariable, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *DistributedSpecSpec) SetEnvironmentVariables(v []EnvironmentVariable)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *DistributedSpecSpec) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *DistributedSpecSpec) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *DistributedSpecSpec) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetExposedUrls

`func (o *DistributedSpecSpec) GetExposedUrls() []ExposedUrl`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *DistributedSpecSpec) GetExposedUrlsOk() (*[]ExposedUrl, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *DistributedSpecSpec) SetExposedUrls(v []ExposedUrl)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *DistributedSpecSpec) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *DistributedSpecSpec) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *DistributedSpecSpec) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetImage

`func (o *DistributedSpecSpec) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *DistributedSpecSpec) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *DistributedSpecSpec) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *DistributedSpecSpec) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *DistributedSpecSpec) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *DistributedSpecSpec) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *DistributedSpecSpec) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *DistributedSpecSpec) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *DistributedSpecSpec) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *DistributedSpecSpec) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *DistributedSpecSpec) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *DistributedSpecSpec) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *DistributedSpecSpec) GetImagePullSecrets() []ImagePullSecret`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *DistributedSpecSpec) GetImagePullSecretsOk() (*[]ImagePullSecret, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *DistributedSpecSpec) SetImagePullSecrets(v []ImagePullSecret)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *DistributedSpecSpec) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *DistributedSpecSpec) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *DistributedSpecSpec) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *DistributedSpecSpec) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *DistributedSpecSpec) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *DistributedSpecSpec) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *DistributedSpecSpec) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *DistributedSpecSpec) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *DistributedSpecSpec) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetMaxReplicas

`func (o *DistributedSpecSpec) GetMaxReplicas() int32`

GetMaxReplicas returns the MaxReplicas field if non-nil, zero value otherwise.

### GetMaxReplicasOk

`func (o *DistributedSpecSpec) GetMaxReplicasOk() (*int32, bool)`

GetMaxReplicasOk returns a tuple with the MaxReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxReplicas

`func (o *DistributedSpecSpec) SetMaxReplicas(v int32)`

SetMaxReplicas sets MaxReplicas field to given value.

### HasMaxReplicas

`func (o *DistributedSpecSpec) HasMaxReplicas() bool`

HasMaxReplicas returns a boolean if a field has been set.

### SetMaxReplicasNil

`func (o *DistributedSpecSpec) SetMaxReplicasNil(b bool)`

 SetMaxReplicasNil sets the value for MaxReplicas to be an explicit nil

### UnsetMaxReplicas
`func (o *DistributedSpecSpec) UnsetMaxReplicas()`

UnsetMaxReplicas ensures that no value is present for MaxReplicas, not even an explicit nil
### GetMinReplicas

`func (o *DistributedSpecSpec) GetMinReplicas() int32`

GetMinReplicas returns the MinReplicas field if non-nil, zero value otherwise.

### GetMinReplicasOk

`func (o *DistributedSpecSpec) GetMinReplicasOk() (*int32, bool)`

GetMinReplicasOk returns a tuple with the MinReplicas field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMinReplicas

`func (o *DistributedSpecSpec) SetMinReplicas(v int32)`

SetMinReplicas sets MinReplicas field to given value.

### HasMinReplicas

`func (o *DistributedSpecSpec) HasMinReplicas() bool`

HasMinReplicas returns a boolean if a field has been set.

### SetMinReplicasNil

`func (o *DistributedSpecSpec) SetMinReplicasNil(b bool)`

 SetMinReplicasNil sets the value for MinReplicas to be an explicit nil

### UnsetMinReplicas
`func (o *DistributedSpecSpec) UnsetMinReplicas()`

UnsetMinReplicas ensures that no value is present for MinReplicas, not even an explicit nil
### GetMpiLauncherCreationPolicy

`func (o *DistributedSpecSpec) GetMpiLauncherCreationPolicy() MpiLauncherCreationPolicy`

GetMpiLauncherCreationPolicy returns the MpiLauncherCreationPolicy field if non-nil, zero value otherwise.

### GetMpiLauncherCreationPolicyOk

`func (o *DistributedSpecSpec) GetMpiLauncherCreationPolicyOk() (*MpiLauncherCreationPolicy, bool)`

GetMpiLauncherCreationPolicyOk returns a tuple with the MpiLauncherCreationPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMpiLauncherCreationPolicy

`func (o *DistributedSpecSpec) SetMpiLauncherCreationPolicy(v MpiLauncherCreationPolicy)`

SetMpiLauncherCreationPolicy sets MpiLauncherCreationPolicy field to given value.

### HasMpiLauncherCreationPolicy

`func (o *DistributedSpecSpec) HasMpiLauncherCreationPolicy() bool`

HasMpiLauncherCreationPolicy returns a boolean if a field has been set.

### SetMpiLauncherCreationPolicyNil

`func (o *DistributedSpecSpec) SetMpiLauncherCreationPolicyNil(b bool)`

 SetMpiLauncherCreationPolicyNil sets the value for MpiLauncherCreationPolicy to be an explicit nil

### UnsetMpiLauncherCreationPolicy
`func (o *DistributedSpecSpec) UnsetMpiLauncherCreationPolicy()`

UnsetMpiLauncherCreationPolicy ensures that no value is present for MpiLauncherCreationPolicy, not even an explicit nil
### GetNodeAffinityRequired

`func (o *DistributedSpecSpec) GetNodeAffinityRequired() NodeAffinityRequired`

GetNodeAffinityRequired returns the NodeAffinityRequired field if non-nil, zero value otherwise.

### GetNodeAffinityRequiredOk

`func (o *DistributedSpecSpec) GetNodeAffinityRequiredOk() (*NodeAffinityRequired, bool)`

GetNodeAffinityRequiredOk returns a tuple with the NodeAffinityRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinityRequired

`func (o *DistributedSpecSpec) SetNodeAffinityRequired(v NodeAffinityRequired)`

SetNodeAffinityRequired sets NodeAffinityRequired field to given value.

### HasNodeAffinityRequired

`func (o *DistributedSpecSpec) HasNodeAffinityRequired() bool`

HasNodeAffinityRequired returns a boolean if a field has been set.

### SetNodeAffinityRequiredNil

`func (o *DistributedSpecSpec) SetNodeAffinityRequiredNil(b bool)`

 SetNodeAffinityRequiredNil sets the value for NodeAffinityRequired to be an explicit nil

### UnsetNodeAffinityRequired
`func (o *DistributedSpecSpec) UnsetNodeAffinityRequired()`

UnsetNodeAffinityRequired ensures that no value is present for NodeAffinityRequired, not even an explicit nil
### GetNodePools

`func (o *DistributedSpecSpec) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *DistributedSpecSpec) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *DistributedSpecSpec) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *DistributedSpecSpec) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *DistributedSpecSpec) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *DistributedSpecSpec) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetNodeType

`func (o *DistributedSpecSpec) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *DistributedSpecSpec) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *DistributedSpecSpec) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *DistributedSpecSpec) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *DistributedSpecSpec) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *DistributedSpecSpec) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNumWorkers

`func (o *DistributedSpecSpec) GetNumWorkers() int32`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *DistributedSpecSpec) GetNumWorkersOk() (*int32, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *DistributedSpecSpec) SetNumWorkers(v int32)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *DistributedSpecSpec) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### SetNumWorkersNil

`func (o *DistributedSpecSpec) SetNumWorkersNil(b bool)`

 SetNumWorkersNil sets the value for NumWorkers to be an explicit nil

### UnsetNumWorkers
`func (o *DistributedSpecSpec) UnsetNumWorkers()`

UnsetNumWorkers ensures that no value is present for NumWorkers, not even an explicit nil
### GetPodAffinity

`func (o *DistributedSpecSpec) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *DistributedSpecSpec) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *DistributedSpecSpec) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *DistributedSpecSpec) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *DistributedSpecSpec) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *DistributedSpecSpec) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetPorts

`func (o *DistributedSpecSpec) GetPorts() []Port`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *DistributedSpecSpec) GetPortsOk() (*[]Port, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *DistributedSpecSpec) SetPorts(v []Port)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *DistributedSpecSpec) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *DistributedSpecSpec) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *DistributedSpecSpec) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetPreemptibility

`func (o *DistributedSpecSpec) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *DistributedSpecSpec) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *DistributedSpecSpec) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *DistributedSpecSpec) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *DistributedSpecSpec) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *DistributedSpecSpec) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *DistributedSpecSpec) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *DistributedSpecSpec) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *DistributedSpecSpec) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *DistributedSpecSpec) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *DistributedSpecSpec) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *DistributedSpecSpec) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *DistributedSpecSpec) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *DistributedSpecSpec) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *DistributedSpecSpec) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *DistributedSpecSpec) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *DistributedSpecSpec) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *DistributedSpecSpec) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetRelatedUrls

`func (o *DistributedSpecSpec) GetRelatedUrls() []RelatedUrl`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *DistributedSpecSpec) GetRelatedUrlsOk() (*[]RelatedUrl, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *DistributedSpecSpec) SetRelatedUrls(v []RelatedUrl)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *DistributedSpecSpec) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *DistributedSpecSpec) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *DistributedSpecSpec) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetRestartPolicy

`func (o *DistributedSpecSpec) GetRestartPolicy() RestartPolicy`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *DistributedSpecSpec) GetRestartPolicyOk() (*RestartPolicy, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *DistributedSpecSpec) SetRestartPolicy(v RestartPolicy)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *DistributedSpecSpec) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *DistributedSpecSpec) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *DistributedSpecSpec) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetSecurity

`func (o *DistributedSpecSpec) GetSecurity() WorkspaceSpecSpecAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *DistributedSpecSpec) GetSecurityOk() (*WorkspaceSpecSpecAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *DistributedSpecSpec) SetSecurity(v WorkspaceSpecSpecAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *DistributedSpecSpec) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *DistributedSpecSpec) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *DistributedSpecSpec) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetSegmentSize

`func (o *DistributedSpecSpec) GetSegmentSize() int32`

GetSegmentSize returns the SegmentSize field if non-nil, zero value otherwise.

### GetSegmentSizeOk

`func (o *DistributedSpecSpec) GetSegmentSizeOk() (*int32, bool)`

GetSegmentSizeOk returns a tuple with the SegmentSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSegmentSize

`func (o *DistributedSpecSpec) SetSegmentSize(v int32)`

SetSegmentSize sets SegmentSize field to given value.

### HasSegmentSize

`func (o *DistributedSpecSpec) HasSegmentSize() bool`

HasSegmentSize returns a boolean if a field has been set.

### SetSegmentSizeNil

`func (o *DistributedSpecSpec) SetSegmentSizeNil(b bool)`

 SetSegmentSizeNil sets the value for SegmentSize to be an explicit nil

### UnsetSegmentSize
`func (o *DistributedSpecSpec) UnsetSegmentSize()`

UnsetSegmentSize ensures that no value is present for SegmentSize, not even an explicit nil
### GetSlotsPerWorker

`func (o *DistributedSpecSpec) GetSlotsPerWorker() int32`

GetSlotsPerWorker returns the SlotsPerWorker field if non-nil, zero value otherwise.

### GetSlotsPerWorkerOk

`func (o *DistributedSpecSpec) GetSlotsPerWorkerOk() (*int32, bool)`

GetSlotsPerWorkerOk returns a tuple with the SlotsPerWorker field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSlotsPerWorker

`func (o *DistributedSpecSpec) SetSlotsPerWorker(v int32)`

SetSlotsPerWorker sets SlotsPerWorker field to given value.

### HasSlotsPerWorker

`func (o *DistributedSpecSpec) HasSlotsPerWorker() bool`

HasSlotsPerWorker returns a boolean if a field has been set.

### SetSlotsPerWorkerNil

`func (o *DistributedSpecSpec) SetSlotsPerWorkerNil(b bool)`

 SetSlotsPerWorkerNil sets the value for SlotsPerWorker to be an explicit nil

### UnsetSlotsPerWorker
`func (o *DistributedSpecSpec) UnsetSlotsPerWorker()`

UnsetSlotsPerWorker ensures that no value is present for SlotsPerWorker, not even an explicit nil
### GetSshAuthMountPath

`func (o *DistributedSpecSpec) GetSshAuthMountPath() string`

GetSshAuthMountPath returns the SshAuthMountPath field if non-nil, zero value otherwise.

### GetSshAuthMountPathOk

`func (o *DistributedSpecSpec) GetSshAuthMountPathOk() (*string, bool)`

GetSshAuthMountPathOk returns a tuple with the SshAuthMountPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSshAuthMountPath

`func (o *DistributedSpecSpec) SetSshAuthMountPath(v string)`

SetSshAuthMountPath sets SshAuthMountPath field to given value.

### HasSshAuthMountPath

`func (o *DistributedSpecSpec) HasSshAuthMountPath() bool`

HasSshAuthMountPath returns a boolean if a field has been set.

### SetSshAuthMountPathNil

`func (o *DistributedSpecSpec) SetSshAuthMountPathNil(b bool)`

 SetSshAuthMountPathNil sets the value for SshAuthMountPath to be an explicit nil

### UnsetSshAuthMountPath
`func (o *DistributedSpecSpec) UnsetSshAuthMountPath()`

UnsetSshAuthMountPath ensures that no value is present for SshAuthMountPath, not even an explicit nil
### GetStdin

`func (o *DistributedSpecSpec) GetStdin() bool`

GetStdin returns the Stdin field if non-nil, zero value otherwise.

### GetStdinOk

`func (o *DistributedSpecSpec) GetStdinOk() (*bool, bool)`

GetStdinOk returns a tuple with the Stdin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdin

`func (o *DistributedSpecSpec) SetStdin(v bool)`

SetStdin sets Stdin field to given value.

### HasStdin

`func (o *DistributedSpecSpec) HasStdin() bool`

HasStdin returns a boolean if a field has been set.

### SetStdinNil

`func (o *DistributedSpecSpec) SetStdinNil(b bool)`

 SetStdinNil sets the value for Stdin to be an explicit nil

### UnsetStdin
`func (o *DistributedSpecSpec) UnsetStdin()`

UnsetStdin ensures that no value is present for Stdin, not even an explicit nil
### GetStorage

`func (o *DistributedSpecSpec) GetStorage() WorkspaceSpecSpecAllOfStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *DistributedSpecSpec) GetStorageOk() (*WorkspaceSpecSpecAllOfStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *DistributedSpecSpec) SetStorage(v WorkspaceSpecSpecAllOfStorage)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *DistributedSpecSpec) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *DistributedSpecSpec) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *DistributedSpecSpec) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *DistributedSpecSpec) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *DistributedSpecSpec) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *DistributedSpecSpec) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *DistributedSpecSpec) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *DistributedSpecSpec) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *DistributedSpecSpec) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetTerminationGracePeriodSeconds

`func (o *DistributedSpecSpec) GetTerminationGracePeriodSeconds() int64`

GetTerminationGracePeriodSeconds returns the TerminationGracePeriodSeconds field if non-nil, zero value otherwise.

### GetTerminationGracePeriodSecondsOk

`func (o *DistributedSpecSpec) GetTerminationGracePeriodSecondsOk() (*int64, bool)`

GetTerminationGracePeriodSecondsOk returns a tuple with the TerminationGracePeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminationGracePeriodSeconds

`func (o *DistributedSpecSpec) SetTerminationGracePeriodSeconds(v int64)`

SetTerminationGracePeriodSeconds sets TerminationGracePeriodSeconds field to given value.

### HasTerminationGracePeriodSeconds

`func (o *DistributedSpecSpec) HasTerminationGracePeriodSeconds() bool`

HasTerminationGracePeriodSeconds returns a boolean if a field has been set.

### SetTerminationGracePeriodSecondsNil

`func (o *DistributedSpecSpec) SetTerminationGracePeriodSecondsNil(b bool)`

 SetTerminationGracePeriodSecondsNil sets the value for TerminationGracePeriodSeconds to be an explicit nil

### UnsetTerminationGracePeriodSeconds
`func (o *DistributedSpecSpec) UnsetTerminationGracePeriodSeconds()`

UnsetTerminationGracePeriodSeconds ensures that no value is present for TerminationGracePeriodSeconds, not even an explicit nil
### GetTolerations

`func (o *DistributedSpecSpec) GetTolerations() []Toleration`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *DistributedSpecSpec) GetTolerationsOk() (*[]Toleration, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *DistributedSpecSpec) SetTolerations(v []Toleration)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *DistributedSpecSpec) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *DistributedSpecSpec) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *DistributedSpecSpec) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTty

`func (o *DistributedSpecSpec) GetTty() bool`

GetTty returns the Tty field if non-nil, zero value otherwise.

### GetTtyOk

`func (o *DistributedSpecSpec) GetTtyOk() (*bool, bool)`

GetTtyOk returns a tuple with the Tty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTty

`func (o *DistributedSpecSpec) SetTty(v bool)`

SetTty sets Tty field to given value.

### HasTty

`func (o *DistributedSpecSpec) HasTty() bool`

HasTty returns a boolean if a field has been set.

### SetTtyNil

`func (o *DistributedSpecSpec) SetTtyNil(b bool)`

 SetTtyNil sets the value for Tty to be an explicit nil

### UnsetTty
`func (o *DistributedSpecSpec) UnsetTty()`

UnsetTty ensures that no value is present for Tty, not even an explicit nil
### GetWorkingDir

`func (o *DistributedSpecSpec) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *DistributedSpecSpec) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *DistributedSpecSpec) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *DistributedSpecSpec) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *DistributedSpecSpec) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *DistributedSpecSpec) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


