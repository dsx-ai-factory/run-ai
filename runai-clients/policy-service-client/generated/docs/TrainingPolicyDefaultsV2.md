# TrainingPolicyDefaultsV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Annotations** | Pointer to [**NullableAnnotationsDefaults**](AnnotationsDefaults.md) |  | [optional] 
**Args** | Pointer to **NullableString** | Arguments to the command that the container running the workload executes. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (completed or failed) will be automatically deleted. The default is 30 days. Log retention is managed separately. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed (not applicable to Inference workloads). The default value is 6. | [optional] 
**Category** | Pointer to **NullableString** | Specify the workload category assigned to the workload. Categories are used to classify and monitor different types of workloads within the NVIDIA Run:ai platform. | [optional] 
**Command** | Pointer to **NullableString** | A command to the server as the entry point of the container running the workload. | [optional] 
**Completions** | Pointer to **NullableInt32** | Used with Hyperparameter Optimization. Specifies the number of successful pods the job should reach to be completed. The Job will be marked as successful once the specified amount of pods has been reached (applicable to standard training only). | [optional] 
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
**Parallelism** | Pointer to **NullableInt32** | Used with Hyperparameter Optimization. Specifies the maximum number of pods the workload should run at any given time (applicable to standard training only). | [optional] 
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

## Methods

### NewTrainingPolicyDefaultsV2

`func NewTrainingPolicyDefaultsV2() *TrainingPolicyDefaultsV2`

NewTrainingPolicyDefaultsV2 instantiates a new TrainingPolicyDefaultsV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingPolicyDefaultsV2WithDefaults

`func NewTrainingPolicyDefaultsV2WithDefaults() *TrainingPolicyDefaultsV2`

NewTrainingPolicyDefaultsV2WithDefaults instantiates a new TrainingPolicyDefaultsV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAnnotations

`func (o *TrainingPolicyDefaultsV2) GetAnnotations() AnnotationsDefaults`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *TrainingPolicyDefaultsV2) GetAnnotationsOk() (*AnnotationsDefaults, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *TrainingPolicyDefaultsV2) SetAnnotations(v AnnotationsDefaults)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *TrainingPolicyDefaultsV2) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *TrainingPolicyDefaultsV2) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *TrainingPolicyDefaultsV2) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetArgs

`func (o *TrainingPolicyDefaultsV2) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *TrainingPolicyDefaultsV2) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *TrainingPolicyDefaultsV2) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *TrainingPolicyDefaultsV2) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *TrainingPolicyDefaultsV2) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *TrainingPolicyDefaultsV2) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingPolicyDefaultsV2) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *TrainingPolicyDefaultsV2) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingPolicyDefaultsV2) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingPolicyDefaultsV2) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *TrainingPolicyDefaultsV2) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *TrainingPolicyDefaultsV2) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *TrainingPolicyDefaultsV2) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *TrainingPolicyDefaultsV2) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *TrainingPolicyDefaultsV2) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *TrainingPolicyDefaultsV2) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *TrainingPolicyDefaultsV2) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *TrainingPolicyDefaultsV2) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetCategory

`func (o *TrainingPolicyDefaultsV2) GetCategory() string`

GetCategory returns the Category field if non-nil, zero value otherwise.

### GetCategoryOk

`func (o *TrainingPolicyDefaultsV2) GetCategoryOk() (*string, bool)`

GetCategoryOk returns a tuple with the Category field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCategory

`func (o *TrainingPolicyDefaultsV2) SetCategory(v string)`

SetCategory sets Category field to given value.

### HasCategory

`func (o *TrainingPolicyDefaultsV2) HasCategory() bool`

HasCategory returns a boolean if a field has been set.

### SetCategoryNil

`func (o *TrainingPolicyDefaultsV2) SetCategoryNil(b bool)`

 SetCategoryNil sets the value for Category to be an explicit nil

### UnsetCategory
`func (o *TrainingPolicyDefaultsV2) UnsetCategory()`

UnsetCategory ensures that no value is present for Category, not even an explicit nil
### GetCommand

`func (o *TrainingPolicyDefaultsV2) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *TrainingPolicyDefaultsV2) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *TrainingPolicyDefaultsV2) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *TrainingPolicyDefaultsV2) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *TrainingPolicyDefaultsV2) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *TrainingPolicyDefaultsV2) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetCompletions

`func (o *TrainingPolicyDefaultsV2) GetCompletions() int32`

GetCompletions returns the Completions field if non-nil, zero value otherwise.

### GetCompletionsOk

`func (o *TrainingPolicyDefaultsV2) GetCompletionsOk() (*int32, bool)`

GetCompletionsOk returns a tuple with the Completions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletions

`func (o *TrainingPolicyDefaultsV2) SetCompletions(v int32)`

SetCompletions sets Completions field to given value.

### HasCompletions

`func (o *TrainingPolicyDefaultsV2) HasCompletions() bool`

HasCompletions returns a boolean if a field has been set.

### SetCompletionsNil

`func (o *TrainingPolicyDefaultsV2) SetCompletionsNil(b bool)`

 SetCompletionsNil sets the value for Completions to be an explicit nil

### UnsetCompletions
`func (o *TrainingPolicyDefaultsV2) UnsetCompletions()`

UnsetCompletions ensures that no value is present for Completions, not even an explicit nil
### GetCompute

`func (o *TrainingPolicyDefaultsV2) GetCompute() SupersetDefaultsAllOfCompute`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *TrainingPolicyDefaultsV2) GetComputeOk() (*SupersetDefaultsAllOfCompute, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *TrainingPolicyDefaultsV2) SetCompute(v SupersetDefaultsAllOfCompute)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *TrainingPolicyDefaultsV2) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *TrainingPolicyDefaultsV2) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *TrainingPolicyDefaultsV2) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetCreateHomeDir

`func (o *TrainingPolicyDefaultsV2) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *TrainingPolicyDefaultsV2) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *TrainingPolicyDefaultsV2) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *TrainingPolicyDefaultsV2) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *TrainingPolicyDefaultsV2) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *TrainingPolicyDefaultsV2) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetEnvironmentVariables

`func (o *TrainingPolicyDefaultsV2) GetEnvironmentVariables() EnvironmentVariablesDefaults`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *TrainingPolicyDefaultsV2) GetEnvironmentVariablesOk() (*EnvironmentVariablesDefaults, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *TrainingPolicyDefaultsV2) SetEnvironmentVariables(v EnvironmentVariablesDefaults)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *TrainingPolicyDefaultsV2) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *TrainingPolicyDefaultsV2) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *TrainingPolicyDefaultsV2) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetExposedUrls

`func (o *TrainingPolicyDefaultsV2) GetExposedUrls() ExposedUrlsDefaults`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *TrainingPolicyDefaultsV2) GetExposedUrlsOk() (*ExposedUrlsDefaults, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *TrainingPolicyDefaultsV2) SetExposedUrls(v ExposedUrlsDefaults)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *TrainingPolicyDefaultsV2) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *TrainingPolicyDefaultsV2) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *TrainingPolicyDefaultsV2) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetImage

`func (o *TrainingPolicyDefaultsV2) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *TrainingPolicyDefaultsV2) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *TrainingPolicyDefaultsV2) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *TrainingPolicyDefaultsV2) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *TrainingPolicyDefaultsV2) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *TrainingPolicyDefaultsV2) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *TrainingPolicyDefaultsV2) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *TrainingPolicyDefaultsV2) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *TrainingPolicyDefaultsV2) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *TrainingPolicyDefaultsV2) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *TrainingPolicyDefaultsV2) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *TrainingPolicyDefaultsV2) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetImagePullSecrets

`func (o *TrainingPolicyDefaultsV2) GetImagePullSecrets() ImagePullSecretsDefaults`

GetImagePullSecrets returns the ImagePullSecrets field if non-nil, zero value otherwise.

### GetImagePullSecretsOk

`func (o *TrainingPolicyDefaultsV2) GetImagePullSecretsOk() (*ImagePullSecretsDefaults, bool)`

GetImagePullSecretsOk returns a tuple with the ImagePullSecrets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullSecrets

`func (o *TrainingPolicyDefaultsV2) SetImagePullSecrets(v ImagePullSecretsDefaults)`

SetImagePullSecrets sets ImagePullSecrets field to given value.

### HasImagePullSecrets

`func (o *TrainingPolicyDefaultsV2) HasImagePullSecrets() bool`

HasImagePullSecrets returns a boolean if a field has been set.

### SetImagePullSecretsNil

`func (o *TrainingPolicyDefaultsV2) SetImagePullSecretsNil(b bool)`

 SetImagePullSecretsNil sets the value for ImagePullSecrets to be an explicit nil

### UnsetImagePullSecrets
`func (o *TrainingPolicyDefaultsV2) UnsetImagePullSecrets()`

UnsetImagePullSecrets ensures that no value is present for ImagePullSecrets, not even an explicit nil
### GetLabels

`func (o *TrainingPolicyDefaultsV2) GetLabels() LabelsDefaults`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *TrainingPolicyDefaultsV2) GetLabelsOk() (*LabelsDefaults, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *TrainingPolicyDefaultsV2) SetLabels(v LabelsDefaults)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *TrainingPolicyDefaultsV2) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *TrainingPolicyDefaultsV2) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *TrainingPolicyDefaultsV2) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetNodeAffinityRequired

`func (o *TrainingPolicyDefaultsV2) GetNodeAffinityRequired() NodeAffinityRequired`

GetNodeAffinityRequired returns the NodeAffinityRequired field if non-nil, zero value otherwise.

### GetNodeAffinityRequiredOk

`func (o *TrainingPolicyDefaultsV2) GetNodeAffinityRequiredOk() (*NodeAffinityRequired, bool)`

GetNodeAffinityRequiredOk returns a tuple with the NodeAffinityRequired field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeAffinityRequired

`func (o *TrainingPolicyDefaultsV2) SetNodeAffinityRequired(v NodeAffinityRequired)`

SetNodeAffinityRequired sets NodeAffinityRequired field to given value.

### HasNodeAffinityRequired

`func (o *TrainingPolicyDefaultsV2) HasNodeAffinityRequired() bool`

HasNodeAffinityRequired returns a boolean if a field has been set.

### SetNodeAffinityRequiredNil

`func (o *TrainingPolicyDefaultsV2) SetNodeAffinityRequiredNil(b bool)`

 SetNodeAffinityRequiredNil sets the value for NodeAffinityRequired to be an explicit nil

### UnsetNodeAffinityRequired
`func (o *TrainingPolicyDefaultsV2) UnsetNodeAffinityRequired()`

UnsetNodeAffinityRequired ensures that no value is present for NodeAffinityRequired, not even an explicit nil
### GetNodePools

`func (o *TrainingPolicyDefaultsV2) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *TrainingPolicyDefaultsV2) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *TrainingPolicyDefaultsV2) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *TrainingPolicyDefaultsV2) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *TrainingPolicyDefaultsV2) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *TrainingPolicyDefaultsV2) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetNodeType

`func (o *TrainingPolicyDefaultsV2) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *TrainingPolicyDefaultsV2) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *TrainingPolicyDefaultsV2) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *TrainingPolicyDefaultsV2) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *TrainingPolicyDefaultsV2) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *TrainingPolicyDefaultsV2) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetParallelism

`func (o *TrainingPolicyDefaultsV2) GetParallelism() int32`

GetParallelism returns the Parallelism field if non-nil, zero value otherwise.

### GetParallelismOk

`func (o *TrainingPolicyDefaultsV2) GetParallelismOk() (*int32, bool)`

GetParallelismOk returns a tuple with the Parallelism field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParallelism

`func (o *TrainingPolicyDefaultsV2) SetParallelism(v int32)`

SetParallelism sets Parallelism field to given value.

### HasParallelism

`func (o *TrainingPolicyDefaultsV2) HasParallelism() bool`

HasParallelism returns a boolean if a field has been set.

### SetParallelismNil

`func (o *TrainingPolicyDefaultsV2) SetParallelismNil(b bool)`

 SetParallelismNil sets the value for Parallelism to be an explicit nil

### UnsetParallelism
`func (o *TrainingPolicyDefaultsV2) UnsetParallelism()`

UnsetParallelism ensures that no value is present for Parallelism, not even an explicit nil
### GetPodAffinity

`func (o *TrainingPolicyDefaultsV2) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *TrainingPolicyDefaultsV2) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *TrainingPolicyDefaultsV2) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *TrainingPolicyDefaultsV2) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *TrainingPolicyDefaultsV2) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *TrainingPolicyDefaultsV2) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetPorts

`func (o *TrainingPolicyDefaultsV2) GetPorts() PortsDefaults`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *TrainingPolicyDefaultsV2) GetPortsOk() (*PortsDefaults, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *TrainingPolicyDefaultsV2) SetPorts(v PortsDefaults)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *TrainingPolicyDefaultsV2) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *TrainingPolicyDefaultsV2) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *TrainingPolicyDefaultsV2) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetPreemptibility

`func (o *TrainingPolicyDefaultsV2) GetPreemptibility() Preemptibility`

GetPreemptibility returns the Preemptibility field if non-nil, zero value otherwise.

### GetPreemptibilityOk

`func (o *TrainingPolicyDefaultsV2) GetPreemptibilityOk() (*Preemptibility, bool)`

GetPreemptibilityOk returns a tuple with the Preemptibility field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPreemptibility

`func (o *TrainingPolicyDefaultsV2) SetPreemptibility(v Preemptibility)`

SetPreemptibility sets Preemptibility field to given value.

### HasPreemptibility

`func (o *TrainingPolicyDefaultsV2) HasPreemptibility() bool`

HasPreemptibility returns a boolean if a field has been set.

### SetPreemptibilityNil

`func (o *TrainingPolicyDefaultsV2) SetPreemptibilityNil(b bool)`

 SetPreemptibilityNil sets the value for Preemptibility to be an explicit nil

### UnsetPreemptibility
`func (o *TrainingPolicyDefaultsV2) UnsetPreemptibility()`

UnsetPreemptibility ensures that no value is present for Preemptibility, not even an explicit nil
### GetPriorityClass

`func (o *TrainingPolicyDefaultsV2) GetPriorityClass() string`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *TrainingPolicyDefaultsV2) GetPriorityClassOk() (*string, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *TrainingPolicyDefaultsV2) SetPriorityClass(v string)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *TrainingPolicyDefaultsV2) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *TrainingPolicyDefaultsV2) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *TrainingPolicyDefaultsV2) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetProbes

`func (o *TrainingPolicyDefaultsV2) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *TrainingPolicyDefaultsV2) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *TrainingPolicyDefaultsV2) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *TrainingPolicyDefaultsV2) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *TrainingPolicyDefaultsV2) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *TrainingPolicyDefaultsV2) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetRelatedUrls

`func (o *TrainingPolicyDefaultsV2) GetRelatedUrls() RelatedUrlsDefaults`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *TrainingPolicyDefaultsV2) GetRelatedUrlsOk() (*RelatedUrlsDefaults, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *TrainingPolicyDefaultsV2) SetRelatedUrls(v RelatedUrlsDefaults)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *TrainingPolicyDefaultsV2) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *TrainingPolicyDefaultsV2) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *TrainingPolicyDefaultsV2) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetRestartPolicy

`func (o *TrainingPolicyDefaultsV2) GetRestartPolicy() RestartPolicy`

GetRestartPolicy returns the RestartPolicy field if non-nil, zero value otherwise.

### GetRestartPolicyOk

`func (o *TrainingPolicyDefaultsV2) GetRestartPolicyOk() (*RestartPolicy, bool)`

GetRestartPolicyOk returns a tuple with the RestartPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRestartPolicy

`func (o *TrainingPolicyDefaultsV2) SetRestartPolicy(v RestartPolicy)`

SetRestartPolicy sets RestartPolicy field to given value.

### HasRestartPolicy

`func (o *TrainingPolicyDefaultsV2) HasRestartPolicy() bool`

HasRestartPolicy returns a boolean if a field has been set.

### SetRestartPolicyNil

`func (o *TrainingPolicyDefaultsV2) SetRestartPolicyNil(b bool)`

 SetRestartPolicyNil sets the value for RestartPolicy to be an explicit nil

### UnsetRestartPolicy
`func (o *TrainingPolicyDefaultsV2) UnsetRestartPolicy()`

UnsetRestartPolicy ensures that no value is present for RestartPolicy, not even an explicit nil
### GetSecurity

`func (o *TrainingPolicyDefaultsV2) GetSecurity() SupersetDefaultsAllOfSecurity`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *TrainingPolicyDefaultsV2) GetSecurityOk() (*SupersetDefaultsAllOfSecurity, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *TrainingPolicyDefaultsV2) SetSecurity(v SupersetDefaultsAllOfSecurity)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *TrainingPolicyDefaultsV2) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *TrainingPolicyDefaultsV2) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *TrainingPolicyDefaultsV2) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetStdin

`func (o *TrainingPolicyDefaultsV2) GetStdin() bool`

GetStdin returns the Stdin field if non-nil, zero value otherwise.

### GetStdinOk

`func (o *TrainingPolicyDefaultsV2) GetStdinOk() (*bool, bool)`

GetStdinOk returns a tuple with the Stdin field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStdin

`func (o *TrainingPolicyDefaultsV2) SetStdin(v bool)`

SetStdin sets Stdin field to given value.

### HasStdin

`func (o *TrainingPolicyDefaultsV2) HasStdin() bool`

HasStdin returns a boolean if a field has been set.

### SetStdinNil

`func (o *TrainingPolicyDefaultsV2) SetStdinNil(b bool)`

 SetStdinNil sets the value for Stdin to be an explicit nil

### UnsetStdin
`func (o *TrainingPolicyDefaultsV2) UnsetStdin()`

UnsetStdin ensures that no value is present for Stdin, not even an explicit nil
### GetStorage

`func (o *TrainingPolicyDefaultsV2) GetStorage() SupersetDefaultsAllOfStorage`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *TrainingPolicyDefaultsV2) GetStorageOk() (*SupersetDefaultsAllOfStorage, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *TrainingPolicyDefaultsV2) SetStorage(v SupersetDefaultsAllOfStorage)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *TrainingPolicyDefaultsV2) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *TrainingPolicyDefaultsV2) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *TrainingPolicyDefaultsV2) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *TrainingPolicyDefaultsV2) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *TrainingPolicyDefaultsV2) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *TrainingPolicyDefaultsV2) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *TrainingPolicyDefaultsV2) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *TrainingPolicyDefaultsV2) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *TrainingPolicyDefaultsV2) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetTerminationGracePeriodSeconds

`func (o *TrainingPolicyDefaultsV2) GetTerminationGracePeriodSeconds() int64`

GetTerminationGracePeriodSeconds returns the TerminationGracePeriodSeconds field if non-nil, zero value otherwise.

### GetTerminationGracePeriodSecondsOk

`func (o *TrainingPolicyDefaultsV2) GetTerminationGracePeriodSecondsOk() (*int64, bool)`

GetTerminationGracePeriodSecondsOk returns a tuple with the TerminationGracePeriodSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminationGracePeriodSeconds

`func (o *TrainingPolicyDefaultsV2) SetTerminationGracePeriodSeconds(v int64)`

SetTerminationGracePeriodSeconds sets TerminationGracePeriodSeconds field to given value.

### HasTerminationGracePeriodSeconds

`func (o *TrainingPolicyDefaultsV2) HasTerminationGracePeriodSeconds() bool`

HasTerminationGracePeriodSeconds returns a boolean if a field has been set.

### SetTerminationGracePeriodSecondsNil

`func (o *TrainingPolicyDefaultsV2) SetTerminationGracePeriodSecondsNil(b bool)`

 SetTerminationGracePeriodSecondsNil sets the value for TerminationGracePeriodSeconds to be an explicit nil

### UnsetTerminationGracePeriodSeconds
`func (o *TrainingPolicyDefaultsV2) UnsetTerminationGracePeriodSeconds()`

UnsetTerminationGracePeriodSeconds ensures that no value is present for TerminationGracePeriodSeconds, not even an explicit nil
### GetTolerations

`func (o *TrainingPolicyDefaultsV2) GetTolerations() TolerationsDefaults`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *TrainingPolicyDefaultsV2) GetTolerationsOk() (*TolerationsDefaults, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *TrainingPolicyDefaultsV2) SetTolerations(v TolerationsDefaults)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *TrainingPolicyDefaultsV2) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *TrainingPolicyDefaultsV2) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *TrainingPolicyDefaultsV2) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTty

`func (o *TrainingPolicyDefaultsV2) GetTty() bool`

GetTty returns the Tty field if non-nil, zero value otherwise.

### GetTtyOk

`func (o *TrainingPolicyDefaultsV2) GetTtyOk() (*bool, bool)`

GetTtyOk returns a tuple with the Tty field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTty

`func (o *TrainingPolicyDefaultsV2) SetTty(v bool)`

SetTty sets Tty field to given value.

### HasTty

`func (o *TrainingPolicyDefaultsV2) HasTty() bool`

HasTty returns a boolean if a field has been set.

### SetTtyNil

`func (o *TrainingPolicyDefaultsV2) SetTtyNil(b bool)`

 SetTtyNil sets the value for Tty to be an explicit nil

### UnsetTty
`func (o *TrainingPolicyDefaultsV2) UnsetTty()`

UnsetTty ensures that no value is present for Tty, not even an explicit nil
### GetWorkingDir

`func (o *TrainingPolicyDefaultsV2) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *TrainingPolicyDefaultsV2) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *TrainingPolicyDefaultsV2) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *TrainingPolicyDefaultsV2) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *TrainingPolicyDefaultsV2) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *TrainingPolicyDefaultsV2) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


