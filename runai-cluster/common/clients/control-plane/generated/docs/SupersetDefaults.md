# SupersetDefaults

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
**EnvironmentVariables** | Pointer to [**NullableEnvironmentVariablesDefaults**](EnvironmentVariablesDefaults.md) |  | [optional] 
**Annotations** | Pointer to [**NullableAnnotationsDefaults**](AnnotationsDefaults.md) |  | [optional] 
**Labels** | Pointer to [**NullableLabelsDefaults**](LabelsDefaults.md) |  | [optional] 
**Tolerations** | Pointer to [**NullableTolerationsDefaults**](TolerationsDefaults.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to **NullableBool** | Indicates if the job should be terminated by the system after it has been preempted. | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to **NullableInt64** | Specifies the duration after which a finished workload (completed or failed) will be automatically deleted. The default is 30 days. | [optional] 
**BackoffLimit** | Pointer to **NullableInt64** | Specifies the number of retries before marking a workload as failed (not applicable to Inference workloads). The default value is 6. | [optional] 
**Ports** | Pointer to [**NullablePortsDefaults**](PortsDefaults.md) |  | [optional] 
**ExposedUrls** | Pointer to [**NullableExposedUrlsDefaults**](ExposedUrlsDefaults.md) |  | [optional] 
**RelatedUrls** | Pointer to [**NullableRelatedUrlsDefaults**](RelatedUrlsDefaults.md) |  | [optional] 
**PriorityClass** | Pointer to [**NullablePriorityClass**](PriorityClass.md) |  | [optional] 
**Completions** | Pointer to **NullableInt32** | Used with Hyperparameter Optimization. Specifies the number of successful pods the job should reach to be completed. The Job will be marked as successful once the specified amount of pods has been reached. | [optional] 
**Parallelism** | Pointer to **NullableInt32** | Used with Hyperparameter Optimization. Specifies the maximum number of pods the workload should run at any given time. | [optional] 
**NumWorkers** | Pointer to **NullableInt32** | the number of workers that will be allocated for running the workload. | [optional] 
**DistributedFramework** | Pointer to [**NullableDistributedFramework**](DistributedFramework.md) |  | [optional] 
**ServingPort** | Pointer to [**NullableServingPort**](ServingPort.md) |  | [optional] 
**Autoscaling** | Pointer to [**NullableAutoScaling**](AutoScaling.md) |  | [optional] 
**Security** | Pointer to [**NullableSecurityFlatFields**](SecurityFlatFields.md) |  | [optional] 
**Compute** | Pointer to [**NullableComputeFieldsDefaults**](ComputeFieldsDefaults.md) |  | [optional] 
**Storage** | Pointer to [**NullableStorageFieldsDefaults**](StorageFieldsDefaults.md) |  | [optional] 

## Methods

### NewSupersetDefaults

`func NewSupersetDefaults() *SupersetDefaults`

NewSupersetDefaults instantiates a new SupersetDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetDefaultsWithDefaults

`func NewSupersetDefaultsWithDefaults() *SupersetDefaults`

NewSupersetDefaultsWithDefaults instantiates a new SupersetDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *SupersetDefaults) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *SupersetDefaults) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *SupersetDefaults) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *SupersetDefaults) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *SupersetDefaults) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *SupersetDefaults) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *SupersetDefaults) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *SupersetDefaults) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *SupersetDefaults) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *SupersetDefaults) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *SupersetDefaults) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *SupersetDefaults) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetImage

`func (o *SupersetDefaults) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *SupersetDefaults) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *SupersetDefaults) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *SupersetDefaults) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *SupersetDefaults) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *SupersetDefaults) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *SupersetDefaults) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *SupersetDefaults) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *SupersetDefaults) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *SupersetDefaults) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *SupersetDefaults) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *SupersetDefaults) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetWorkingDir

`func (o *SupersetDefaults) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *SupersetDefaults) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *SupersetDefaults) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *SupersetDefaults) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *SupersetDefaults) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *SupersetDefaults) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCreateHomeDir

`func (o *SupersetDefaults) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *SupersetDefaults) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *SupersetDefaults) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *SupersetDefaults) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *SupersetDefaults) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *SupersetDefaults) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetProbes

`func (o *SupersetDefaults) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *SupersetDefaults) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *SupersetDefaults) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *SupersetDefaults) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *SupersetDefaults) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *SupersetDefaults) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetNodeType

`func (o *SupersetDefaults) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *SupersetDefaults) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *SupersetDefaults) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *SupersetDefaults) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *SupersetDefaults) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *SupersetDefaults) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *SupersetDefaults) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *SupersetDefaults) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *SupersetDefaults) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *SupersetDefaults) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *SupersetDefaults) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *SupersetDefaults) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *SupersetDefaults) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *SupersetDefaults) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *SupersetDefaults) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *SupersetDefaults) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *SupersetDefaults) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *SupersetDefaults) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetEnvironmentVariables

`func (o *SupersetDefaults) GetEnvironmentVariables() EnvironmentVariablesDefaults`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *SupersetDefaults) GetEnvironmentVariablesOk() (*EnvironmentVariablesDefaults, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *SupersetDefaults) SetEnvironmentVariables(v EnvironmentVariablesDefaults)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *SupersetDefaults) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *SupersetDefaults) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *SupersetDefaults) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetAnnotations

`func (o *SupersetDefaults) GetAnnotations() AnnotationsDefaults`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *SupersetDefaults) GetAnnotationsOk() (*AnnotationsDefaults, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *SupersetDefaults) SetAnnotations(v AnnotationsDefaults)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *SupersetDefaults) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *SupersetDefaults) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *SupersetDefaults) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *SupersetDefaults) GetLabels() LabelsDefaults`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *SupersetDefaults) GetLabelsOk() (*LabelsDefaults, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *SupersetDefaults) SetLabels(v LabelsDefaults)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *SupersetDefaults) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *SupersetDefaults) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *SupersetDefaults) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetTolerations

`func (o *SupersetDefaults) GetTolerations() TolerationsDefaults`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *SupersetDefaults) GetTolerationsOk() (*TolerationsDefaults, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *SupersetDefaults) SetTolerations(v TolerationsDefaults)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *SupersetDefaults) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *SupersetDefaults) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *SupersetDefaults) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *SupersetDefaults) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *SupersetDefaults) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *SupersetDefaults) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *SupersetDefaults) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *SupersetDefaults) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *SupersetDefaults) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *SupersetDefaults) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *SupersetDefaults) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *SupersetDefaults) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *SupersetDefaults) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *SupersetDefaults) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *SupersetDefaults) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *SupersetDefaults) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *SupersetDefaults) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *SupersetDefaults) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *SupersetDefaults) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *SupersetDefaults) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *SupersetDefaults) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetPorts

`func (o *SupersetDefaults) GetPorts() PortsDefaults`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *SupersetDefaults) GetPortsOk() (*PortsDefaults, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *SupersetDefaults) SetPorts(v PortsDefaults)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *SupersetDefaults) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *SupersetDefaults) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *SupersetDefaults) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetExposedUrls

`func (o *SupersetDefaults) GetExposedUrls() ExposedUrlsDefaults`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *SupersetDefaults) GetExposedUrlsOk() (*ExposedUrlsDefaults, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *SupersetDefaults) SetExposedUrls(v ExposedUrlsDefaults)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *SupersetDefaults) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *SupersetDefaults) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *SupersetDefaults) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetRelatedUrls

`func (o *SupersetDefaults) GetRelatedUrls() RelatedUrlsDefaults`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *SupersetDefaults) GetRelatedUrlsOk() (*RelatedUrlsDefaults, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *SupersetDefaults) SetRelatedUrls(v RelatedUrlsDefaults)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *SupersetDefaults) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *SupersetDefaults) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *SupersetDefaults) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetPriorityClass

`func (o *SupersetDefaults) GetPriorityClass() PriorityClass`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *SupersetDefaults) GetPriorityClassOk() (*PriorityClass, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *SupersetDefaults) SetPriorityClass(v PriorityClass)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *SupersetDefaults) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *SupersetDefaults) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *SupersetDefaults) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetCompletions

`func (o *SupersetDefaults) GetCompletions() int32`

GetCompletions returns the Completions field if non-nil, zero value otherwise.

### GetCompletionsOk

`func (o *SupersetDefaults) GetCompletionsOk() (*int32, bool)`

GetCompletionsOk returns a tuple with the Completions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletions

`func (o *SupersetDefaults) SetCompletions(v int32)`

SetCompletions sets Completions field to given value.

### HasCompletions

`func (o *SupersetDefaults) HasCompletions() bool`

HasCompletions returns a boolean if a field has been set.

### SetCompletionsNil

`func (o *SupersetDefaults) SetCompletionsNil(b bool)`

 SetCompletionsNil sets the value for Completions to be an explicit nil

### UnsetCompletions
`func (o *SupersetDefaults) UnsetCompletions()`

UnsetCompletions ensures that no value is present for Completions, not even an explicit nil
### GetParallelism

`func (o *SupersetDefaults) GetParallelism() int32`

GetParallelism returns the Parallelism field if non-nil, zero value otherwise.

### GetParallelismOk

`func (o *SupersetDefaults) GetParallelismOk() (*int32, bool)`

GetParallelismOk returns a tuple with the Parallelism field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParallelism

`func (o *SupersetDefaults) SetParallelism(v int32)`

SetParallelism sets Parallelism field to given value.

### HasParallelism

`func (o *SupersetDefaults) HasParallelism() bool`

HasParallelism returns a boolean if a field has been set.

### SetParallelismNil

`func (o *SupersetDefaults) SetParallelismNil(b bool)`

 SetParallelismNil sets the value for Parallelism to be an explicit nil

### UnsetParallelism
`func (o *SupersetDefaults) UnsetParallelism()`

UnsetParallelism ensures that no value is present for Parallelism, not even an explicit nil
### GetNumWorkers

`func (o *SupersetDefaults) GetNumWorkers() int32`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *SupersetDefaults) GetNumWorkersOk() (*int32, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *SupersetDefaults) SetNumWorkers(v int32)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *SupersetDefaults) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### SetNumWorkersNil

`func (o *SupersetDefaults) SetNumWorkersNil(b bool)`

 SetNumWorkersNil sets the value for NumWorkers to be an explicit nil

### UnsetNumWorkers
`func (o *SupersetDefaults) UnsetNumWorkers()`

UnsetNumWorkers ensures that no value is present for NumWorkers, not even an explicit nil
### GetDistributedFramework

`func (o *SupersetDefaults) GetDistributedFramework() DistributedFramework`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *SupersetDefaults) GetDistributedFrameworkOk() (*DistributedFramework, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *SupersetDefaults) SetDistributedFramework(v DistributedFramework)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *SupersetDefaults) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### SetDistributedFrameworkNil

`func (o *SupersetDefaults) SetDistributedFrameworkNil(b bool)`

 SetDistributedFrameworkNil sets the value for DistributedFramework to be an explicit nil

### UnsetDistributedFramework
`func (o *SupersetDefaults) UnsetDistributedFramework()`

UnsetDistributedFramework ensures that no value is present for DistributedFramework, not even an explicit nil
### GetServingPort

`func (o *SupersetDefaults) GetServingPort() ServingPort`

GetServingPort returns the ServingPort field if non-nil, zero value otherwise.

### GetServingPortOk

`func (o *SupersetDefaults) GetServingPortOk() (*ServingPort, bool)`

GetServingPortOk returns a tuple with the ServingPort field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetServingPort

`func (o *SupersetDefaults) SetServingPort(v ServingPort)`

SetServingPort sets ServingPort field to given value.

### HasServingPort

`func (o *SupersetDefaults) HasServingPort() bool`

HasServingPort returns a boolean if a field has been set.

### SetServingPortNil

`func (o *SupersetDefaults) SetServingPortNil(b bool)`

 SetServingPortNil sets the value for ServingPort to be an explicit nil

### UnsetServingPort
`func (o *SupersetDefaults) UnsetServingPort()`

UnsetServingPort ensures that no value is present for ServingPort, not even an explicit nil
### GetAutoscaling

`func (o *SupersetDefaults) GetAutoscaling() AutoScaling`

GetAutoscaling returns the Autoscaling field if non-nil, zero value otherwise.

### GetAutoscalingOk

`func (o *SupersetDefaults) GetAutoscalingOk() (*AutoScaling, bool)`

GetAutoscalingOk returns a tuple with the Autoscaling field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoscaling

`func (o *SupersetDefaults) SetAutoscaling(v AutoScaling)`

SetAutoscaling sets Autoscaling field to given value.

### HasAutoscaling

`func (o *SupersetDefaults) HasAutoscaling() bool`

HasAutoscaling returns a boolean if a field has been set.

### SetAutoscalingNil

`func (o *SupersetDefaults) SetAutoscalingNil(b bool)`

 SetAutoscalingNil sets the value for Autoscaling to be an explicit nil

### UnsetAutoscaling
`func (o *SupersetDefaults) UnsetAutoscaling()`

UnsetAutoscaling ensures that no value is present for Autoscaling, not even an explicit nil
### GetSecurity

`func (o *SupersetDefaults) GetSecurity() SecurityFlatFields`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *SupersetDefaults) GetSecurityOk() (*SecurityFlatFields, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *SupersetDefaults) SetSecurity(v SecurityFlatFields)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *SupersetDefaults) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *SupersetDefaults) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *SupersetDefaults) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetCompute

`func (o *SupersetDefaults) GetCompute() ComputeFieldsDefaults`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *SupersetDefaults) GetComputeOk() (*ComputeFieldsDefaults, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *SupersetDefaults) SetCompute(v ComputeFieldsDefaults)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *SupersetDefaults) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *SupersetDefaults) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *SupersetDefaults) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetStorage

`func (o *SupersetDefaults) GetStorage() StorageFieldsDefaults`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *SupersetDefaults) GetStorageOk() (*StorageFieldsDefaults, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *SupersetDefaults) SetStorage(v StorageFieldsDefaults)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *SupersetDefaults) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *SupersetDefaults) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *SupersetDefaults) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


