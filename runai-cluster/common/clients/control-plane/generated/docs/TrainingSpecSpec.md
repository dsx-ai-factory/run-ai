# TrainingSpecSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to **NullableString** | A command to the server as the entry point of the container running the workspace. | [optional] 
**Args** | Pointer to **NullableString** | Arguments to the command that the container running the workspace executes. | [optional] 
**Image** | Pointer to **NullableString** | Docker image name. For more information, see [Images](https://kubernetes.io/docs/concepts/containers/images). The image name is mandatory for creating a workspace. | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicy**](ImagePullPolicy.md) |  | [optional] 
**WorkingDir** | Pointer to **NullableString** | Container&#39;s working directory. If not specified, the container runtime default will be used. This may be configured in the container image. | [optional] 
**CreateHomeDir** | Pointer to **NullableBool** | When set to &#x60;true&#x60;, creates a home directory for the container. | [optional] 
**Probes** | Pointer to [**NullableProbes**](Probes.md) |  | [optional] 
**NodeType** | Pointer to **NullableString** | Nodes (machines), or a group of nodes on which the workload will run. To use this feature, your Administrator will need to label nodes. For more information, see [Group Nodes](https://docs.run.ai/latest/admin/researcher-setup/limit-to-node-group). When using this flag with with Project-based affinity, it refines the list of allowable node groups set in the Project. For more information, see [Projects](https://docs.run.ai/admin/admin-ui-setup/project-setup). | [optional] 
**NodePools** | Pointer to **[]string** | A prioritized list of node pools for the scheduler to run the workspace on. The scheduler will always try to use the first node pool before moving to the next one if the first is not available. | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinity**](PodAffinity.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**[]EnvironmentVariable**](EnvironmentVariable.md) | Set of environment variables to populate into the container running the workspace. | [optional] 
**Annotations** | Pointer to [**[]Annotation**](Annotation.md) | Set of annotations to populate into the container running the workspace. | [optional] 
**Labels** | Pointer to [**[]Label**](Label.md) | Set of labels to populate into the container running the workspace. | [optional] 
**Tolerations** | Pointer to [**[]Toleration**](Toleration.md) | Set of tolerations to apply to the workload. | [optional] 
**TerminateAfterPreemption** | Pointer to **NullableBool** | Indicates if the job should be terminated by the system after it has been preempted. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (completed or failed) will be automatically deleted. The default is 30 days. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed (not applicable to Inference workloads). The default value is 6. | [optional] 
**Ports** | Pointer to [**[]Port**](Port.md) | Set of container ports that the workload exposes. | [optional] 
**ExposedUrls** | Pointer to [**[]ExposedUrl**](ExposedUrl.md) | Set of container ports that the workload exposes via URLs. | [optional] 
**RelatedUrls** | Pointer to [**[]RelatedUrl**](RelatedUrl.md) | Set of URLs that are related to the workload. | [optional] 
**PriorityClass** | Pointer to [**NullablePriorityClass**](PriorityClass.md) |  | [optional] 
**Completions** | Pointer to **NullableInt32** | Used with Hyperparameter Optimization. Specifies the number of successful pods the job should reach to be completed. The Job will be marked as successful once the specified amount of pods has been reached. | [optional] 
**Parallelism** | Pointer to **NullableInt32** | Used with Hyperparameter Optimization. Specifies the maximum number of pods the workload should run at any given time. | [optional] 
**Compute** | Pointer to [**NullableComputeFields**](ComputeFields.md) |  | [optional] 
**Storage** | Pointer to [**NullableStorageFields**](StorageFields.md) |  | [optional] 
**Security** | Pointer to [**NullableSecurityFlatFields**](SecurityFlatFields.md) |  | [optional] 

## Methods

### NewTrainingSpecSpec

`func NewTrainingSpecSpec() *TrainingSpecSpec`

NewTrainingSpecSpec instantiates a new TrainingSpecSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingSpecSpecWithDefaults

`func NewTrainingSpecSpecWithDefaults() *TrainingSpecSpec`

NewTrainingSpecSpecWithDefaults instantiates a new TrainingSpecSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *TrainingSpecSpec) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *TrainingSpecSpec) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *TrainingSpecSpec) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *TrainingSpecSpec) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *TrainingSpecSpec) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *TrainingSpecSpec) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *TrainingSpecSpec) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *TrainingSpecSpec) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *TrainingSpecSpec) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *TrainingSpecSpec) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *TrainingSpecSpec) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *TrainingSpecSpec) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetImage

`func (o *TrainingSpecSpec) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *TrainingSpecSpec) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *TrainingSpecSpec) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *TrainingSpecSpec) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *TrainingSpecSpec) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *TrainingSpecSpec) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *TrainingSpecSpec) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *TrainingSpecSpec) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *TrainingSpecSpec) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *TrainingSpecSpec) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *TrainingSpecSpec) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *TrainingSpecSpec) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetWorkingDir

`func (o *TrainingSpecSpec) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *TrainingSpecSpec) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *TrainingSpecSpec) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *TrainingSpecSpec) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *TrainingSpecSpec) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *TrainingSpecSpec) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCreateHomeDir

`func (o *TrainingSpecSpec) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *TrainingSpecSpec) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *TrainingSpecSpec) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *TrainingSpecSpec) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *TrainingSpecSpec) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *TrainingSpecSpec) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetProbes

`func (o *TrainingSpecSpec) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *TrainingSpecSpec) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *TrainingSpecSpec) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *TrainingSpecSpec) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *TrainingSpecSpec) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *TrainingSpecSpec) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetNodeType

`func (o *TrainingSpecSpec) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *TrainingSpecSpec) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *TrainingSpecSpec) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *TrainingSpecSpec) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *TrainingSpecSpec) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *TrainingSpecSpec) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *TrainingSpecSpec) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *TrainingSpecSpec) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *TrainingSpecSpec) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *TrainingSpecSpec) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *TrainingSpecSpec) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *TrainingSpecSpec) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *TrainingSpecSpec) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *TrainingSpecSpec) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *TrainingSpecSpec) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *TrainingSpecSpec) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *TrainingSpecSpec) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *TrainingSpecSpec) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetEnvironmentVariables

`func (o *TrainingSpecSpec) GetEnvironmentVariables() []EnvironmentVariable`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *TrainingSpecSpec) GetEnvironmentVariablesOk() (*[]EnvironmentVariable, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *TrainingSpecSpec) SetEnvironmentVariables(v []EnvironmentVariable)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *TrainingSpecSpec) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *TrainingSpecSpec) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *TrainingSpecSpec) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetAnnotations

`func (o *TrainingSpecSpec) GetAnnotations() []Annotation`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *TrainingSpecSpec) GetAnnotationsOk() (*[]Annotation, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *TrainingSpecSpec) SetAnnotations(v []Annotation)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *TrainingSpecSpec) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *TrainingSpecSpec) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *TrainingSpecSpec) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *TrainingSpecSpec) GetLabels() []Label`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *TrainingSpecSpec) GetLabelsOk() (*[]Label, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *TrainingSpecSpec) SetLabels(v []Label)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *TrainingSpecSpec) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *TrainingSpecSpec) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *TrainingSpecSpec) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetTolerations

`func (o *TrainingSpecSpec) GetTolerations() []Toleration`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *TrainingSpecSpec) GetTolerationsOk() (*[]Toleration, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *TrainingSpecSpec) SetTolerations(v []Toleration)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *TrainingSpecSpec) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *TrainingSpecSpec) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *TrainingSpecSpec) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *TrainingSpecSpec) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *TrainingSpecSpec) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *TrainingSpecSpec) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *TrainingSpecSpec) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *TrainingSpecSpec) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *TrainingSpecSpec) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingSpecSpec) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *TrainingSpecSpec) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingSpecSpec) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingSpecSpec) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *TrainingSpecSpec) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *TrainingSpecSpec) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *TrainingSpecSpec) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *TrainingSpecSpec) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *TrainingSpecSpec) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *TrainingSpecSpec) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *TrainingSpecSpec) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *TrainingSpecSpec) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetPorts

`func (o *TrainingSpecSpec) GetPorts() []Port`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *TrainingSpecSpec) GetPortsOk() (*[]Port, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *TrainingSpecSpec) SetPorts(v []Port)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *TrainingSpecSpec) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *TrainingSpecSpec) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *TrainingSpecSpec) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetExposedUrls

`func (o *TrainingSpecSpec) GetExposedUrls() []ExposedUrl`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *TrainingSpecSpec) GetExposedUrlsOk() (*[]ExposedUrl, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *TrainingSpecSpec) SetExposedUrls(v []ExposedUrl)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *TrainingSpecSpec) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *TrainingSpecSpec) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *TrainingSpecSpec) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetRelatedUrls

`func (o *TrainingSpecSpec) GetRelatedUrls() []RelatedUrl`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *TrainingSpecSpec) GetRelatedUrlsOk() (*[]RelatedUrl, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *TrainingSpecSpec) SetRelatedUrls(v []RelatedUrl)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *TrainingSpecSpec) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *TrainingSpecSpec) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *TrainingSpecSpec) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetPriorityClass

`func (o *TrainingSpecSpec) GetPriorityClass() PriorityClass`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *TrainingSpecSpec) GetPriorityClassOk() (*PriorityClass, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *TrainingSpecSpec) SetPriorityClass(v PriorityClass)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *TrainingSpecSpec) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *TrainingSpecSpec) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *TrainingSpecSpec) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetCompletions

`func (o *TrainingSpecSpec) GetCompletions() int32`

GetCompletions returns the Completions field if non-nil, zero value otherwise.

### GetCompletionsOk

`func (o *TrainingSpecSpec) GetCompletionsOk() (*int32, bool)`

GetCompletionsOk returns a tuple with the Completions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletions

`func (o *TrainingSpecSpec) SetCompletions(v int32)`

SetCompletions sets Completions field to given value.

### HasCompletions

`func (o *TrainingSpecSpec) HasCompletions() bool`

HasCompletions returns a boolean if a field has been set.

### SetCompletionsNil

`func (o *TrainingSpecSpec) SetCompletionsNil(b bool)`

 SetCompletionsNil sets the value for Completions to be an explicit nil

### UnsetCompletions
`func (o *TrainingSpecSpec) UnsetCompletions()`

UnsetCompletions ensures that no value is present for Completions, not even an explicit nil
### GetParallelism

`func (o *TrainingSpecSpec) GetParallelism() int32`

GetParallelism returns the Parallelism field if non-nil, zero value otherwise.

### GetParallelismOk

`func (o *TrainingSpecSpec) GetParallelismOk() (*int32, bool)`

GetParallelismOk returns a tuple with the Parallelism field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParallelism

`func (o *TrainingSpecSpec) SetParallelism(v int32)`

SetParallelism sets Parallelism field to given value.

### HasParallelism

`func (o *TrainingSpecSpec) HasParallelism() bool`

HasParallelism returns a boolean if a field has been set.

### SetParallelismNil

`func (o *TrainingSpecSpec) SetParallelismNil(b bool)`

 SetParallelismNil sets the value for Parallelism to be an explicit nil

### UnsetParallelism
`func (o *TrainingSpecSpec) UnsetParallelism()`

UnsetParallelism ensures that no value is present for Parallelism, not even an explicit nil
### GetCompute

`func (o *TrainingSpecSpec) GetCompute() ComputeFields`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *TrainingSpecSpec) GetComputeOk() (*ComputeFields, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *TrainingSpecSpec) SetCompute(v ComputeFields)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *TrainingSpecSpec) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *TrainingSpecSpec) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *TrainingSpecSpec) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetStorage

`func (o *TrainingSpecSpec) GetStorage() StorageFields`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *TrainingSpecSpec) GetStorageOk() (*StorageFields, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *TrainingSpecSpec) SetStorage(v StorageFields)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *TrainingSpecSpec) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *TrainingSpecSpec) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *TrainingSpecSpec) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetSecurity

`func (o *TrainingSpecSpec) GetSecurity() SecurityFlatFields`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *TrainingSpecSpec) GetSecurityOk() (*SecurityFlatFields, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *TrainingSpecSpec) SetSecurity(v SecurityFlatFields)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *TrainingSpecSpec) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *TrainingSpecSpec) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *TrainingSpecSpec) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


