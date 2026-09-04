# TrainingPolicyDefaultsAndRulesV2Rules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Command** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Args** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**Image** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**ImagePullPolicy** | Pointer to [**NullableImagePullPolicyRules**](ImagePullPolicyRules.md) |  | [optional] 
**WorkingDir** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**CreateHomeDir** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**NodeType** | Pointer to [**NullableStringRules**](StringRules.md) |  | [optional] 
**NodePools** | Pointer to [**NullableArrayRules**](ArrayRules.md) |  | [optional] 
**TerminateAfterPreemption** | Pointer to [**NullableBooleanRules**](BooleanRules.md) |  | [optional] 
**AutoDeletionTimeAfterCompletionSeconds** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**BackoffLimit** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**EnvironmentVariables** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Annotations** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Labels** | Pointer to [**NullableInstancesRules**](InstancesRules.md) |  | [optional] 
**Ports** | Pointer to [**NullablePortsRules**](PortsRules.md) |  | [optional] 
**ExposedUrls** | Pointer to [**NullableExposedUrlsRules**](ExposedUrlsRules.md) |  | [optional] 
**RelatedUrls** | Pointer to [**NullableRelatedUrlsRules**](RelatedUrlsRules.md) |  | [optional] 
**PriorityClass** | Pointer to [**NullablePriorityClassRules**](PriorityClassRules.md) |  | [optional] 
**Completions** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Parallelism** | Pointer to [**NullableIntegerRules**](IntegerRules.md) |  | [optional] 
**Security** | Pointer to [**NullableSecurityFlatFieldsRules**](SecurityFlatFieldsRules.md) |  | [optional] 
**Compute** | Pointer to [**NullableComputeFieldsRules**](ComputeFieldsRules.md) |  | [optional] 
**Storage** | Pointer to [**NullableStorageFieldsRules**](StorageFieldsRules.md) |  | [optional] 

## Methods

### NewTrainingPolicyDefaultsAndRulesV2Rules

`func NewTrainingPolicyDefaultsAndRulesV2Rules() *TrainingPolicyDefaultsAndRulesV2Rules`

NewTrainingPolicyDefaultsAndRulesV2Rules instantiates a new TrainingPolicyDefaultsAndRulesV2Rules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingPolicyDefaultsAndRulesV2RulesWithDefaults

`func NewTrainingPolicyDefaultsAndRulesV2RulesWithDefaults() *TrainingPolicyDefaultsAndRulesV2Rules`

NewTrainingPolicyDefaultsAndRulesV2RulesWithDefaults instantiates a new TrainingPolicyDefaultsAndRulesV2Rules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetCommand

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetCommand() StringRules`

GetCommand returns the Command field if non-nil, zero value otherwise.

### GetCommandOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetCommandOk() (*StringRules, bool)`

GetCommandOk returns a tuple with the Command field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCommand

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetCommand(v StringRules)`

SetCommand sets Command field to given value.

### HasCommand

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasCommand() bool`

HasCommand returns a boolean if a field has been set.

### SetCommandNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetCommandNil(b bool)`

 SetCommandNil sets the value for Command to be an explicit nil

### UnsetCommand
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetCommand()`

UnsetCommand ensures that no value is present for Command, not even an explicit nil
### GetArgs

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetArgs() StringRules`

GetArgs returns the Args field if non-nil, zero value otherwise.

### GetArgsOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetArgsOk() (*StringRules, bool)`

GetArgsOk returns a tuple with the Args field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetArgs

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetArgs(v StringRules)`

SetArgs sets Args field to given value.

### HasArgs

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasArgs() bool`

HasArgs returns a boolean if a field has been set.

### SetArgsNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetArgsNil(b bool)`

 SetArgsNil sets the value for Args to be an explicit nil

### UnsetArgs
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetArgs()`

UnsetArgs ensures that no value is present for Args, not even an explicit nil
### GetImage

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetImage() StringRules`

GetImage returns the Image field if non-nil, zero value otherwise.

### GetImageOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetImageOk() (*StringRules, bool)`

GetImageOk returns a tuple with the Image field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImage

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetImage(v StringRules)`

SetImage sets Image field to given value.

### HasImage

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasImage() bool`

HasImage returns a boolean if a field has been set.

### SetImageNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetImageNil(b bool)`

 SetImageNil sets the value for Image to be an explicit nil

### UnsetImage
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetImage()`

UnsetImage ensures that no value is present for Image, not even an explicit nil
### GetImagePullPolicy

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetImagePullPolicy() ImagePullPolicyRules`

GetImagePullPolicy returns the ImagePullPolicy field if non-nil, zero value otherwise.

### GetImagePullPolicyOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetImagePullPolicyOk() (*ImagePullPolicyRules, bool)`

GetImagePullPolicyOk returns a tuple with the ImagePullPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImagePullPolicy

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetImagePullPolicy(v ImagePullPolicyRules)`

SetImagePullPolicy sets ImagePullPolicy field to given value.

### HasImagePullPolicy

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasImagePullPolicy() bool`

HasImagePullPolicy returns a boolean if a field has been set.

### SetImagePullPolicyNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetImagePullPolicyNil(b bool)`

 SetImagePullPolicyNil sets the value for ImagePullPolicy to be an explicit nil

### UnsetImagePullPolicy
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetImagePullPolicy()`

UnsetImagePullPolicy ensures that no value is present for ImagePullPolicy, not even an explicit nil
### GetWorkingDir

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetWorkingDir() StringRules`

GetWorkingDir returns the WorkingDir field if non-nil, zero value otherwise.

### GetWorkingDirOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetWorkingDirOk() (*StringRules, bool)`

GetWorkingDirOk returns a tuple with the WorkingDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkingDir

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetWorkingDir(v StringRules)`

SetWorkingDir sets WorkingDir field to given value.

### HasWorkingDir

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasWorkingDir() bool`

HasWorkingDir returns a boolean if a field has been set.

### SetWorkingDirNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetWorkingDirNil(b bool)`

 SetWorkingDirNil sets the value for WorkingDir to be an explicit nil

### UnsetWorkingDir
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetWorkingDir()`

UnsetWorkingDir ensures that no value is present for WorkingDir, not even an explicit nil
### GetCreateHomeDir

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetCreateHomeDir() BooleanRules`

GetCreateHomeDir returns the CreateHomeDir field if non-nil, zero value otherwise.

### GetCreateHomeDirOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetCreateHomeDirOk() (*BooleanRules, bool)`

GetCreateHomeDirOk returns a tuple with the CreateHomeDir field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCreateHomeDir

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetCreateHomeDir(v BooleanRules)`

SetCreateHomeDir sets CreateHomeDir field to given value.

### HasCreateHomeDir

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasCreateHomeDir() bool`

HasCreateHomeDir returns a boolean if a field has been set.

### SetCreateHomeDirNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetCreateHomeDirNil(b bool)`

 SetCreateHomeDirNil sets the value for CreateHomeDir to be an explicit nil

### UnsetCreateHomeDir
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetCreateHomeDir()`

UnsetCreateHomeDir ensures that no value is present for CreateHomeDir, not even an explicit nil
### GetNodeType

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetNodeType() StringRules`

GetNodeType returns the NodeType field if non-nil, zero value otherwise.

### GetNodeTypeOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetNodeTypeOk() (*StringRules, bool)`

GetNodeTypeOk returns a tuple with the NodeType field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodeType

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetNodeType(v StringRules)`

SetNodeType sets NodeType field to given value.

### HasNodeType

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasNodeType() bool`

HasNodeType returns a boolean if a field has been set.

### SetNodeTypeNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetNodeTypeNil(b bool)`

 SetNodeTypeNil sets the value for NodeType to be an explicit nil

### UnsetNodeType
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetNodeType()`

UnsetNodeType ensures that no value is present for NodeType, not even an explicit nil
### GetNodePools

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetNodePools() ArrayRules`

GetNodePools returns the NodePools field if non-nil, zero value otherwise.

### GetNodePoolsOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetNodePoolsOk() (*ArrayRules, bool)`

GetNodePoolsOk returns a tuple with the NodePools field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePools

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetNodePools(v ArrayRules)`

SetNodePools sets NodePools field to given value.

### HasNodePools

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasNodePools() bool`

HasNodePools returns a boolean if a field has been set.

### SetNodePoolsNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetNodePoolsNil(b bool)`

 SetNodePoolsNil sets the value for NodePools to be an explicit nil

### UnsetNodePools
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetNodePools()`

UnsetNodePools ensures that no value is present for NodePools, not even an explicit nil
### GetTerminateAfterPreemption

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetTerminateAfterPreemption() BooleanRules`

GetTerminateAfterPreemption returns the TerminateAfterPreemption field if non-nil, zero value otherwise.

### GetTerminateAfterPreemptionOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetTerminateAfterPreemptionOk() (*BooleanRules, bool)`

GetTerminateAfterPreemptionOk returns a tuple with the TerminateAfterPreemption field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTerminateAfterPreemption

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetTerminateAfterPreemption(v BooleanRules)`

SetTerminateAfterPreemption sets TerminateAfterPreemption field to given value.

### HasTerminateAfterPreemption

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasTerminateAfterPreemption() bool`

HasTerminateAfterPreemption returns a boolean if a field has been set.

### SetTerminateAfterPreemptionNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetTerminateAfterPreemptionNil(b bool)`

 SetTerminateAfterPreemptionNil sets the value for TerminateAfterPreemption to be an explicit nil

### UnsetTerminateAfterPreemption
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetTerminateAfterPreemption()`

UnsetTerminateAfterPreemption ensures that no value is present for TerminateAfterPreemption, not even an explicit nil
### GetAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetAutoDeletionTimeAfterCompletionSeconds() IntegerRules`

GetAutoDeletionTimeAfterCompletionSeconds returns the AutoDeletionTimeAfterCompletionSeconds field if non-nil, zero value otherwise.

### GetAutoDeletionTimeAfterCompletionSecondsOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetAutoDeletionTimeAfterCompletionSecondsOk() (*IntegerRules, bool)`

GetAutoDeletionTimeAfterCompletionSecondsOk returns a tuple with the AutoDeletionTimeAfterCompletionSeconds field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetAutoDeletionTimeAfterCompletionSeconds(v IntegerRules)`

SetAutoDeletionTimeAfterCompletionSeconds sets AutoDeletionTimeAfterCompletionSeconds field to given value.

### HasAutoDeletionTimeAfterCompletionSeconds

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasAutoDeletionTimeAfterCompletionSeconds() bool`

HasAutoDeletionTimeAfterCompletionSeconds returns a boolean if a field has been set.

### SetAutoDeletionTimeAfterCompletionSecondsNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetAutoDeletionTimeAfterCompletionSecondsNil(b bool)`

 SetAutoDeletionTimeAfterCompletionSecondsNil sets the value for AutoDeletionTimeAfterCompletionSeconds to be an explicit nil

### UnsetAutoDeletionTimeAfterCompletionSeconds
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetAutoDeletionTimeAfterCompletionSeconds()`

UnsetAutoDeletionTimeAfterCompletionSeconds ensures that no value is present for AutoDeletionTimeAfterCompletionSeconds, not even an explicit nil
### GetBackoffLimit

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetBackoffLimit() IntegerRules`

GetBackoffLimit returns the BackoffLimit field if non-nil, zero value otherwise.

### GetBackoffLimitOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetBackoffLimitOk() (*IntegerRules, bool)`

GetBackoffLimitOk returns a tuple with the BackoffLimit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBackoffLimit

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetBackoffLimit(v IntegerRules)`

SetBackoffLimit sets BackoffLimit field to given value.

### HasBackoffLimit

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasBackoffLimit() bool`

HasBackoffLimit returns a boolean if a field has been set.

### SetBackoffLimitNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetBackoffLimitNil(b bool)`

 SetBackoffLimitNil sets the value for BackoffLimit to be an explicit nil

### UnsetBackoffLimit
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetBackoffLimit()`

UnsetBackoffLimit ensures that no value is present for BackoffLimit, not even an explicit nil
### GetEnvironmentVariables

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetEnvironmentVariables() InstancesRules`

GetEnvironmentVariables returns the EnvironmentVariables field if non-nil, zero value otherwise.

### GetEnvironmentVariablesOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetEnvironmentVariablesOk() (*InstancesRules, bool)`

GetEnvironmentVariablesOk returns a tuple with the EnvironmentVariables field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnvironmentVariables

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetEnvironmentVariables(v InstancesRules)`

SetEnvironmentVariables sets EnvironmentVariables field to given value.

### HasEnvironmentVariables

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasEnvironmentVariables() bool`

HasEnvironmentVariables returns a boolean if a field has been set.

### SetEnvironmentVariablesNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetEnvironmentVariablesNil(b bool)`

 SetEnvironmentVariablesNil sets the value for EnvironmentVariables to be an explicit nil

### UnsetEnvironmentVariables
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetEnvironmentVariables()`

UnsetEnvironmentVariables ensures that no value is present for EnvironmentVariables, not even an explicit nil
### GetAnnotations

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetAnnotations() InstancesRules`

GetAnnotations returns the Annotations field if non-nil, zero value otherwise.

### GetAnnotationsOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetAnnotationsOk() (*InstancesRules, bool)`

GetAnnotationsOk returns a tuple with the Annotations field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAnnotations

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetAnnotations(v InstancesRules)`

SetAnnotations sets Annotations field to given value.

### HasAnnotations

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasAnnotations() bool`

HasAnnotations returns a boolean if a field has been set.

### SetAnnotationsNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetAnnotationsNil(b bool)`

 SetAnnotationsNil sets the value for Annotations to be an explicit nil

### UnsetAnnotations
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetAnnotations()`

UnsetAnnotations ensures that no value is present for Annotations, not even an explicit nil
### GetLabels

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetLabels() InstancesRules`

GetLabels returns the Labels field if non-nil, zero value otherwise.

### GetLabelsOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetLabelsOk() (*InstancesRules, bool)`

GetLabelsOk returns a tuple with the Labels field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetLabels

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetLabels(v InstancesRules)`

SetLabels sets Labels field to given value.

### HasLabels

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasLabels() bool`

HasLabels returns a boolean if a field has been set.

### SetLabelsNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetLabelsNil(b bool)`

 SetLabelsNil sets the value for Labels to be an explicit nil

### UnsetLabels
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetLabels()`

UnsetLabels ensures that no value is present for Labels, not even an explicit nil
### GetPorts

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetPorts() PortsRules`

GetPorts returns the Ports field if non-nil, zero value otherwise.

### GetPortsOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetPortsOk() (*PortsRules, bool)`

GetPortsOk returns a tuple with the Ports field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPorts

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetPorts(v PortsRules)`

SetPorts sets Ports field to given value.

### HasPorts

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasPorts() bool`

HasPorts returns a boolean if a field has been set.

### SetPortsNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetPortsNil(b bool)`

 SetPortsNil sets the value for Ports to be an explicit nil

### UnsetPorts
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetPorts()`

UnsetPorts ensures that no value is present for Ports, not even an explicit nil
### GetExposedUrls

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetExposedUrls() ExposedUrlsRules`

GetExposedUrls returns the ExposedUrls field if non-nil, zero value otherwise.

### GetExposedUrlsOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetExposedUrlsOk() (*ExposedUrlsRules, bool)`

GetExposedUrlsOk returns a tuple with the ExposedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExposedUrls

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetExposedUrls(v ExposedUrlsRules)`

SetExposedUrls sets ExposedUrls field to given value.

### HasExposedUrls

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasExposedUrls() bool`

HasExposedUrls returns a boolean if a field has been set.

### SetExposedUrlsNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetExposedUrlsNil(b bool)`

 SetExposedUrlsNil sets the value for ExposedUrls to be an explicit nil

### UnsetExposedUrls
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetExposedUrls()`

UnsetExposedUrls ensures that no value is present for ExposedUrls, not even an explicit nil
### GetRelatedUrls

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetRelatedUrls() RelatedUrlsRules`

GetRelatedUrls returns the RelatedUrls field if non-nil, zero value otherwise.

### GetRelatedUrlsOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetRelatedUrlsOk() (*RelatedUrlsRules, bool)`

GetRelatedUrlsOk returns a tuple with the RelatedUrls field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRelatedUrls

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetRelatedUrls(v RelatedUrlsRules)`

SetRelatedUrls sets RelatedUrls field to given value.

### HasRelatedUrls

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasRelatedUrls() bool`

HasRelatedUrls returns a boolean if a field has been set.

### SetRelatedUrlsNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetRelatedUrlsNil(b bool)`

 SetRelatedUrlsNil sets the value for RelatedUrls to be an explicit nil

### UnsetRelatedUrls
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetRelatedUrls()`

UnsetRelatedUrls ensures that no value is present for RelatedUrls, not even an explicit nil
### GetPriorityClass

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetPriorityClass() PriorityClassRules`

GetPriorityClass returns the PriorityClass field if non-nil, zero value otherwise.

### GetPriorityClassOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetPriorityClassOk() (*PriorityClassRules, bool)`

GetPriorityClassOk returns a tuple with the PriorityClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPriorityClass

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetPriorityClass(v PriorityClassRules)`

SetPriorityClass sets PriorityClass field to given value.

### HasPriorityClass

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasPriorityClass() bool`

HasPriorityClass returns a boolean if a field has been set.

### SetPriorityClassNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetPriorityClassNil(b bool)`

 SetPriorityClassNil sets the value for PriorityClass to be an explicit nil

### UnsetPriorityClass
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetPriorityClass()`

UnsetPriorityClass ensures that no value is present for PriorityClass, not even an explicit nil
### GetCompletions

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetCompletions() IntegerRules`

GetCompletions returns the Completions field if non-nil, zero value otherwise.

### GetCompletionsOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetCompletionsOk() (*IntegerRules, bool)`

GetCompletionsOk returns a tuple with the Completions field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompletions

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetCompletions(v IntegerRules)`

SetCompletions sets Completions field to given value.

### HasCompletions

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasCompletions() bool`

HasCompletions returns a boolean if a field has been set.

### SetCompletionsNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetCompletionsNil(b bool)`

 SetCompletionsNil sets the value for Completions to be an explicit nil

### UnsetCompletions
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetCompletions()`

UnsetCompletions ensures that no value is present for Completions, not even an explicit nil
### GetParallelism

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetParallelism() IntegerRules`

GetParallelism returns the Parallelism field if non-nil, zero value otherwise.

### GetParallelismOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetParallelismOk() (*IntegerRules, bool)`

GetParallelismOk returns a tuple with the Parallelism field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetParallelism

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetParallelism(v IntegerRules)`

SetParallelism sets Parallelism field to given value.

### HasParallelism

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasParallelism() bool`

HasParallelism returns a boolean if a field has been set.

### SetParallelismNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetParallelismNil(b bool)`

 SetParallelismNil sets the value for Parallelism to be an explicit nil

### UnsetParallelism
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetParallelism()`

UnsetParallelism ensures that no value is present for Parallelism, not even an explicit nil
### GetSecurity

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetSecurity() SecurityFlatFieldsRules`

GetSecurity returns the Security field if non-nil, zero value otherwise.

### GetSecurityOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetSecurityOk() (*SecurityFlatFieldsRules, bool)`

GetSecurityOk returns a tuple with the Security field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecurity

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetSecurity(v SecurityFlatFieldsRules)`

SetSecurity sets Security field to given value.

### HasSecurity

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasSecurity() bool`

HasSecurity returns a boolean if a field has been set.

### SetSecurityNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetSecurityNil(b bool)`

 SetSecurityNil sets the value for Security to be an explicit nil

### UnsetSecurity
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetSecurity()`

UnsetSecurity ensures that no value is present for Security, not even an explicit nil
### GetCompute

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetCompute() ComputeFieldsRules`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetComputeOk() (*ComputeFieldsRules, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetCompute(v ComputeFieldsRules)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### SetComputeNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetComputeNil(b bool)`

 SetComputeNil sets the value for Compute to be an explicit nil

### UnsetCompute
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetCompute()`

UnsetCompute ensures that no value is present for Compute, not even an explicit nil
### GetStorage

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetStorage() StorageFieldsRules`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) GetStorageOk() (*StorageFieldsRules, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetStorage(v StorageFieldsRules)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) HasStorage() bool`

HasStorage returns a boolean if a field has been set.

### SetStorageNil

`func (o *TrainingPolicyDefaultsAndRulesV2Rules) SetStorageNil(b bool)`

 SetStorageNil sets the value for Storage to be an explicit nil

### UnsetStorage
`func (o *TrainingPolicyDefaultsAndRulesV2Rules) UnsetStorage()`

UnsetStorage ensures that no value is present for Storage, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


