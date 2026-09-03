# WorkspacePolicyDefaultsAndRulesV2Defaults

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
**Security** | Pointer to [**NullableSecurityFlatFields**](SecurityFlatFields.md) |  | [optional] 
**Compute** | Pointer to [**NullableComputeFieldsDefaults**](ComputeFieldsDefaults.md) |  | [optional] 
**Storage** | Pointer to [**NullableStorageFieldsDefaults**](StorageFieldsDefaults.md) |  | [optional] 

## Methods

### NewWorkspacePolicyDefaultsAndRulesV2Defaults

`func NewWorkspacePolicyDefaultsAndRulesV2Defaults() *WorkspacePolicyDefaultsAndRulesV2Defaults`

NewWorkspacePolicyDefaultsAndRulesV2Defaults instantiates a new WorkspacePolicyDefaultsAndRulesV2Defaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspacePolicyDefaultsAndRulesV2DefaultsWithDefaults

`func NewWorkspacePolicyDefaultsAndRulesV2DefaultsWithDefaults() *WorkspacePolicyDefaultsAndRulesV2Defaults`

NewWorkspacePolicyDefaultsAndRulesV2DefaultsWithDefaults instantiates a new WorkspacePolicyDefaultsAndRulesV2Defaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetCommand() string`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetCommandOk() (*string, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetCommand(v string)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetArgs() string`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetArgsOk() (*string, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetArgs(v string)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetImage

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetImage() string`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetImageOk() (*string, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetImage(v string)`

SetImage sets Image field to given value.

### HasImage

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetImagePullPolicy() ImagePullPolicy`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetImagePullPolicyOk() (*ImagePullPolicy, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetImagePullPolicy(v ImagePullPolicy)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetWorkingDir

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetWorkingDir() string`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetWorkingDirOk() (*string, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetWorkingDir(v string)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCreateHomeDir

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetCreateHomeDir() bool`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetCreateHomeDirOk() (*bool, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetCreateHomeDir(v bool)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetProbes

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetProbes() Probes`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetProbesOk() (*Probes, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetProbes(v Probes)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetNodeType

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetNodeType() string`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetNodeTypeOk() (*string, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetNodeType(v string)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetNodePools() []string`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetNodePoolsOk() (*[]string, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetNodePools(v []string)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetPodAffinity() PodAffinity`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetPodAffinityOk() (*PodAffinity, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetPodAffinity(v PodAffinity)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetEnvironmentVariables

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetEnvironmentVariables() EnvironmentVariablesDefaults`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetEnvironmentVariablesOk() (*EnvironmentVariablesDefaults, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetEnvironmentVariables(v EnvironmentVariablesDefaults)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetAnnotations

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetAnnotations() AnnotationsDefaults`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetAnnotationsOk() (*AnnotationsDefaults, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetAnnotations(v AnnotationsDefaults)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetLabels() LabelsDefaults`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetLabelsOk() (*LabelsDefaults, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetLabels(v LabelsDefaults)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetTolerations

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetTolerations() TolerationsDefaults`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetTolerationsOk() (*TolerationsDefaults, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetTolerations(v TolerationsDefaults)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetTerminateAfterPreemption() bool`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetTerminateAfterPreemptionOk() (*bool, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetTerminateAfterPreemption(v bool)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetAutoDeletionTimeAfterCompletionSeconds() int64`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetAutoDeletionTimeAfterCompletionSecondsOk() (*int64, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetAutoDeletionTimeAfterCompletionSeconds(v int64)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetBackoffLimit() int64`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetBackoffLimitOk() (*int64, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetBackoffLimit(v int64)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetPorts

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetPorts() PortsDefaults`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetPortsOk() (*PortsDefaults, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetPorts(v PortsDefaults)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetExposedUrls

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetExposedUrls() ExposedUrlsDefaults`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetExposedUrlsOk() (*ExposedUrlsDefaults, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetExposedUrls(v ExposedUrlsDefaults)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetRelatedUrls

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetRelatedUrls() RelatedUrlsDefaults`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetRelatedUrlsOk() (*RelatedUrlsDefaults, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetRelatedUrls(v RelatedUrlsDefaults)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetPriorityClass

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetPriorityClass() PriorityClass`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetPriorityClassOk() (*PriorityClass, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetPriorityClass(v PriorityClass)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetSecurity

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetSecurity() SecurityFlatFields`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetSecurityOk() (*SecurityFlatFields, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetSecurity(v SecurityFlatFields)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetCompute

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetCompute() ComputeFieldsDefaults`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetComputeOk() (*ComputeFieldsDefaults, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetCompute(v ComputeFieldsDefaults)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetStorage

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetStorage() StorageFieldsDefaults`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) GetStorageOk() (*StorageFieldsDefaults, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetStorage(v StorageFieldsDefaults)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *WorkspacePolicyDefaultsAndRulesV2Defaults) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


