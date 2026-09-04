# DistributedPolicyRulesV2Worker

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Args** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Image** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicyRules**](ImagePullPolicyRules.md) |  | [optional] 
**WorkingDir** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**CreateHomeDir** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**Probes** | Pointer to [**NullableProbesRules**](ProbesRules.md) |  | [optional] 
**NodeType** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**NodePools** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**PodAffinity** | Pointer to [**NullablePodAffinityRules**](PodAffinityRules.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Annotations** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Labels** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Tolerations** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**BackoffLimit** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Ports** | Pointer to [**NullablePortsRules**](PortsRules.md) |  | [optional] 
**ExposedUrls** | Pointer to [**NullableExposedUrlsRules**](ExposedUrlsRules.md) |  | [optional] 
**RelatedUrls** | Pointer to [**NullableRelatedUrlsRules**](RelatedUrlsRules.md) |  | [optional] 
**Security** | Pointer to [**NullableSecurityFlatFieldsRules**](SecurityFlatFieldsRules.md) |  | [optional] 
**Compute** | Pointer to [**NullableComputeFieldsRules**](ComputeFieldsRules.md) |  | [optional] 
**Storage** | Pointer to [**NullableStorageFieldsRules**](StorageFieldsRules.md) |  | [optional] 
**NumWorkers** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**DistributedFramework** | Pointer to [**NullableDistributedFrameworkRules**](DistributedFrameworkRules.md) |  | [optional] 

## Methods

### NewDistributedPolicyRulesV2Worker

`func NewDistributedPolicyRulesV2Worker() *DistributedPolicyRulesV2Worker`

NewDistributedPolicyRulesV2Worker instantiates a new DistributedPolicyRulesV2Worker object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedPolicyRulesV2WorkerWithDefaults

`func NewDistributedPolicyRulesV2WorkerWithDefaults() *DistributedPolicyRulesV2Worker`

NewDistributedPolicyRulesV2WorkerWithDefaults instantiates a new DistributedPolicyRulesV2Worker object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *DistributedPolicyRulesV2Worker) GetCommand() StringRules`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *DistributedPolicyRulesV2Worker) GetCommandOk() (*StringRules, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *DistributedPolicyRulesV2Worker) SetCommand(v StringRules)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *DistributedPolicyRulesV2Worker) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *DistributedPolicyRulesV2Worker) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *DistributedPolicyRulesV2Worker) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *DistributedPolicyRulesV2Worker) GetArgs() StringRules`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *DistributedPolicyRulesV2Worker) GetArgsOk() (*StringRules, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *DistributedPolicyRulesV2Worker) SetArgs(v StringRules)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *DistributedPolicyRulesV2Worker) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *DistributedPolicyRulesV2Worker) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *DistributedPolicyRulesV2Worker) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetImage

`func (o *DistributedPolicyRulesV2Worker) GetImage() StringRules`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *DistributedPolicyRulesV2Worker) GetImageOk() (*StringRules, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *DistributedPolicyRulesV2Worker) SetImage(v StringRules)`

SetImage sets Image field to given value.

### HasImage

`func (o *DistributedPolicyRulesV2Worker) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *DistributedPolicyRulesV2Worker) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *DistributedPolicyRulesV2Worker) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *DistributedPolicyRulesV2Worker) GetImagePullPolicy() ImagePullPolicyRules`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *DistributedPolicyRulesV2Worker) GetImagePullPolicyOk() (*ImagePullPolicyRules, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *DistributedPolicyRulesV2Worker) SetImagePullPolicy(v ImagePullPolicyRules)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *DistributedPolicyRulesV2Worker) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *DistributedPolicyRulesV2Worker) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *DistributedPolicyRulesV2Worker) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetWorkingDir

`func (o *DistributedPolicyRulesV2Worker) GetWorkingDir() StringRules`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *DistributedPolicyRulesV2Worker) GetWorkingDirOk() (*StringRules, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *DistributedPolicyRulesV2Worker) SetWorkingDir(v StringRules)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *DistributedPolicyRulesV2Worker) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *DistributedPolicyRulesV2Worker) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *DistributedPolicyRulesV2Worker) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCreateHomeDir

`func (o *DistributedPolicyRulesV2Worker) GetCreateHomeDir() BooleanRules`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *DistributedPolicyRulesV2Worker) GetCreateHomeDirOk() (*BooleanRules, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *DistributedPolicyRulesV2Worker) SetCreateHomeDir(v BooleanRules)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *DistributedPolicyRulesV2Worker) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *DistributedPolicyRulesV2Worker) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *DistributedPolicyRulesV2Worker) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetProbes

`func (o *DistributedPolicyRulesV2Worker) GetProbes() ProbesRules`

GetProbes returns the Probes field if non-nil, zero value otherwise.

### GetProbesOk

`func (o *DistributedPolicyRulesV2Worker) GetProbesOk() (*ProbesRules, bool)`

GetProbesOk returns a tuple with the Probes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProbes

`func (o *DistributedPolicyRulesV2Worker) SetProbes(v ProbesRules)`

SetProbes sets Probes field to given value.

### HasProbes

`func (o *DistributedPolicyRulesV2Worker) HasProbes() bool`

HasProbes returns a boolean if a field has been set.

### SetProbesNil

`func (o *DistributedPolicyRulesV2Worker) SetProbesNil(b bool)`

 SetProbesNil sets the value for Probes to be an explicit nil

### UnsetProbes
`func (o *DistributedPolicyRulesV2Worker) UnsetProbes()`

UnsetProbes ensures that no value is present for Probes, not even an explicit nil
### GetNodeType

`func (o *DistributedPolicyRulesV2Worker) GetNodeType() StringRules`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *DistributedPolicyRulesV2Worker) GetNodeTypeOk() (*StringRules, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *DistributedPolicyRulesV2Worker) SetNodeType(v StringRules)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *DistributedPolicyRulesV2Worker) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *DistributedPolicyRulesV2Worker) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *DistributedPolicyRulesV2Worker) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *DistributedPolicyRulesV2Worker) GetNodePools() ArrayRules`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *DistributedPolicyRulesV2Worker) GetNodePoolsOk() (*ArrayRules, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *DistributedPolicyRulesV2Worker) SetNodePools(v ArrayRules)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *DistributedPolicyRulesV2Worker) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *DistributedPolicyRulesV2Worker) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *DistributedPolicyRulesV2Worker) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetPodAffinity

`func (o *DistributedPolicyRulesV2Worker) GetPodAffinity() PodAffinityRules`

GetPodAffinity returns the PodAffinity field if non-nil, zero value otherwise.

### GetPodAffinityOk

`func (o *DistributedPolicyRulesV2Worker) GetPodAffinityOk() (*PodAffinityRules, bool)`

GetPodAffinityOk returns a tuple with the PodAffinity field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPodAffinity

`func (o *DistributedPolicyRulesV2Worker) SetPodAffinity(v PodAffinityRules)`

SetPodAffinity sets PodAffinity field to given value.

### HasPodAffinity

`func (o *DistributedPolicyRulesV2Worker) HasPodAffinity() bool`

HasPodAffinity returns a boolean if a field has been set.

### SetPodAffinityNil

`func (o *DistributedPolicyRulesV2Worker) SetPodAffinityNil(b bool)`

 SetPodAffinityNil sets the value for PodAffinity to be an explicit nil

### UnsetPodAffinity
`func (o *DistributedPolicyRulesV2Worker) UnsetPodAffinity()`

UnsetPodAffinity ensures that no value is present for PodAffinity, not even an explicit nil
### GetEnvironmentVariables

`func (o *DistributedPolicyRulesV2Worker) GetEnvironmentVariables() InstancesRules`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *DistributedPolicyRulesV2Worker) GetEnvironmentVariablesOk() (*InstancesRules, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *DistributedPolicyRulesV2Worker) SetEnvironmentVariables(v InstancesRules)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *DistributedPolicyRulesV2Worker) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *DistributedPolicyRulesV2Worker) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *DistributedPolicyRulesV2Worker) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetAnnotations

`func (o *DistributedPolicyRulesV2Worker) GetAnnotations() InstancesRules`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *DistributedPolicyRulesV2Worker) GetAnnotationsOk() (*InstancesRules, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *DistributedPolicyRulesV2Worker) SetAnnotations(v InstancesRules)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *DistributedPolicyRulesV2Worker) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *DistributedPolicyRulesV2Worker) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *DistributedPolicyRulesV2Worker) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *DistributedPolicyRulesV2Worker) GetLabels() InstancesRules`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *DistributedPolicyRulesV2Worker) GetLabelsOk() (*InstancesRules, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *DistributedPolicyRulesV2Worker) SetLabels(v InstancesRules)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *DistributedPolicyRulesV2Worker) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *DistributedPolicyRulesV2Worker) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *DistributedPolicyRulesV2Worker) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetTolerations

`func (o *DistributedPolicyRulesV2Worker) GetTolerations() InstancesRules`

GetTolerations returns the Tolerations field if non-nil, zero value otherwise.

### GetTolerationsOk

`func (o *DistributedPolicyRulesV2Worker) GetTolerationsOk() (*InstancesRules, bool)`

GetTolerationsOk returns a tuple with the Tolerations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTolerations

`func (o *DistributedPolicyRulesV2Worker) SetTolerations(v InstancesRules)`

SetTolerations sets Tolerations field to given value.

### HasTolerations

`func (o *DistributedPolicyRulesV2Worker) HasTolerations() bool`

HasTolerations returns a boolean if a field has been set.

### SetTolerationsNil

`func (o *DistributedPolicyRulesV2Worker) SetTolerationsNil(b bool)`

 SetTolerationsNil sets the value for Tolerations to be an explicit nil

### UnsetTolerations
`func (o *DistributedPolicyRulesV2Worker) UnsetTolerations()`

UnsetTolerations ensures that no value is present for Tolerations, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *DistributedPolicyRulesV2Worker) GetTerminateAfterPreemption() BooleanRules`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *DistributedPolicyRulesV2Worker) GetTerminateAfterPreemptionOk() (*BooleanRules, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *DistributedPolicyRulesV2Worker) SetTerminateAfterPreemption(v BooleanRules)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *DistributedPolicyRulesV2Worker) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *DistributedPolicyRulesV2Worker) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *DistributedPolicyRulesV2Worker) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *DistributedPolicyRulesV2Worker) GetAutoDeletionTimeAfterCompletionSeconds() IntegerRules`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *DistributedPolicyRulesV2Worker) GetAutoDeletionTimeAfterCompletionSecondsOk() (*IntegerRules, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *DistributedPolicyRulesV2Worker) SetAutoDeletionTimeAfterCompletionSeconds(v IntegerRules)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *DistributedPolicyRulesV2Worker) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *DistributedPolicyRulesV2Worker) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *DistributedPolicyRulesV2Worker) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *DistributedPolicyRulesV2Worker) GetBackoffLimit() IntegerRules`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *DistributedPolicyRulesV2Worker) GetBackoffLimitOk() (*IntegerRules, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *DistributedPolicyRulesV2Worker) SetBackoffLimit(v IntegerRules)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *DistributedPolicyRulesV2Worker) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *DistributedPolicyRulesV2Worker) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *DistributedPolicyRulesV2Worker) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetPorts

`func (o *DistributedPolicyRulesV2Worker) GetPorts() PortsRules`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *DistributedPolicyRulesV2Worker) GetPortsOk() (*PortsRules, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *DistributedPolicyRulesV2Worker) SetPorts(v PortsRules)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *DistributedPolicyRulesV2Worker) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *DistributedPolicyRulesV2Worker) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *DistributedPolicyRulesV2Worker) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetExposedUrls

`func (o *DistributedPolicyRulesV2Worker) GetExposedUrls() ExposedUrlsRules`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *DistributedPolicyRulesV2Worker) GetExposedUrlsOk() (*ExposedUrlsRules, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *DistributedPolicyRulesV2Worker) SetExposedUrls(v ExposedUrlsRules)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *DistributedPolicyRulesV2Worker) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *DistributedPolicyRulesV2Worker) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *DistributedPolicyRulesV2Worker) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetRelatedUrls

`func (o *DistributedPolicyRulesV2Worker) GetRelatedUrls() RelatedUrlsRules`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *DistributedPolicyRulesV2Worker) GetRelatedUrlsOk() (*RelatedUrlsRules, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *DistributedPolicyRulesV2Worker) SetRelatedUrls(v RelatedUrlsRules)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *DistributedPolicyRulesV2Worker) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *DistributedPolicyRulesV2Worker) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *DistributedPolicyRulesV2Worker) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetSecurity

`func (o *DistributedPolicyRulesV2Worker) GetSecurity() SecurityFlatFieldsRules`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *DistributedPolicyRulesV2Worker) GetSecurityOk() (*SecurityFlatFieldsRules, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *DistributedPolicyRulesV2Worker) SetSecurity(v SecurityFlatFieldsRules)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *DistributedPolicyRulesV2Worker) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *DistributedPolicyRulesV2Worker) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *DistributedPolicyRulesV2Worker) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetCompute

`func (o *DistributedPolicyRulesV2Worker) GetCompute() ComputeFieldsRules`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *DistributedPolicyRulesV2Worker) GetComputeOk() (*ComputeFieldsRules, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *DistributedPolicyRulesV2Worker) SetCompute(v ComputeFieldsRules)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *DistributedPolicyRulesV2Worker) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *DistributedPolicyRulesV2Worker) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *DistributedPolicyRulesV2Worker) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetStorage

`func (o *DistributedPolicyRulesV2Worker) GetStorage() StorageFieldsRules`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *DistributedPolicyRulesV2Worker) GetStorageOk() (*StorageFieldsRules, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *DistributedPolicyRulesV2Worker) SetStorage(v StorageFieldsRules)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *DistributedPolicyRulesV2Worker) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *DistributedPolicyRulesV2Worker) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *DistributedPolicyRulesV2Worker) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil
### GetNumWorkers

`func (o *DistributedPolicyRulesV2Worker) GetNumWorkers() IntegerRules`

GetNumWorkers returns the NumWorkers field if non-nil, zero value otherwise.

### GetNumWorkersOk

`func (o *DistributedPolicyRulesV2Worker) GetNumWorkersOk() (*IntegerRules, bool)`

GetNumWorkersOk returns a tuple with the NumWorkers field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNumWorkers

`func (o *DistributedPolicyRulesV2Worker) SetNumWorkers(v IntegerRules)`

SetNumWorkers sets NumWorkers field to given value.

### HasNumWorkers

`func (o *DistributedPolicyRulesV2Worker) HasNumWorkers() bool`

HasNumWorkers returns a boolean if a field has been set.

### SetNumWorkersNil

`func (o *DistributedPolicyRulesV2Worker) SetNumWorkersNil(b bool)`

 SetNumWorkersNil sets the value for NumWorkers to be an explicit nil

### UnsetNumWorkers
`func (o *DistributedPolicyRulesV2Worker) UnsetNumWorkers()`

UnsetNumWorkers ensures that no value is present for NumWorkers, not even an explicit nil
### GetDistributedFramework

`func (o *DistributedPolicyRulesV2Worker) GetDistributedFramework() DistributedFrameworkRules`

GetDistributedFramework returns the DistributedFramework field if non-nil, zero value otherwise.

### GetDistributedFrameworkOk

`func (o *DistributedPolicyRulesV2Worker) GetDistributedFrameworkOk() (*DistributedFrameworkRules, bool)`

GetDistributedFrameworkOk returns a tuple with the DistributedFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedFramework

`func (o *DistributedPolicyRulesV2Worker) SetDistributedFramework(v DistributedFrameworkRules)`

SetDistributedFramework sets DistributedFramework field to given value.

### HasDistributedFramework

`func (o *DistributedPolicyRulesV2Worker) HasDistributedFramework() bool`

HasDistributedFramework returns a boolean if a field has been set.

### SetDistributedFrameworkNil

`func (o *DistributedPolicyRulesV2Worker) SetDistributedFrameworkNil(b bool)`

 SetDistributedFrameworkNil sets the value for DistributedFramework to be an explicit nil

### UnsetDistributedFramework
`func (o *DistributedPolicyRulesV2Worker) UnsetDistributedFramework()`

UnsetDistributedFramework ensures that no value is present for DistributedFramework, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


