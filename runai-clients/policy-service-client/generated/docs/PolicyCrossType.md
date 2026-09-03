# PolicyCrossType

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Workspace** | Pointer to [**NullableWorkspacePolicyV2**](WorkspacePolicyV2.md) |  | [optional] 
**Training** | Pointer to [**NullableTrainingPolicyV2**](TrainingPolicyV2.md) |  | [optional] 
**Distributed** | Pointer to [**NullableDistributedPolicyV2**](DistributedPolicyV2.md) |  | [optional] 
**Inference** | Pointer to [**NullableInferencePolicyV2**](InferencePolicyV2.md) |  | [optional] 
**DistributedInference** | Pointer to [**NullableDistributedInferencePolicyV2**](DistributedInferencePolicyV2.md) |  | [optional] 
**NimService** | Pointer to [**NullableNimServicePolicy**](NimServicePolicy.md) |  | [optional] 

## Methods

### NewPolicyCrossType

`func NewPolicyCrossType() *PolicyCrossType`

NewPolicyCrossType instantiates a new PolicyCrossType object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyCrossTypeWithDefaults

`func NewPolicyCrossTypeWithDefaults() *PolicyCrossType`

NewPolicyCrossTypeWithDefaults instantiates a new PolicyCrossType object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetWorkspace

`func (o *PolicyCrossType) GetWorkspace() WorkspacePolicyV2`

GetWorkspace returns the Workspace field if non-nil, zero value otherwise.

### GetWorkspaceOk

`func (o *PolicyCrossType) GetWorkspaceOk() (*WorkspacePolicyV2, bool)`

GetWorkspaceOk returns a tuple with the Workspace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspace

`func (o *PolicyCrossType) SetWorkspace(v WorkspacePolicyV2)`

SetWorkspace sets Workspace field to given value.

### HasWorkspace

`func (o *PolicyCrossType) HasWorkspace() bool`

HasWorkspace returns a boolean if a field has been set.

### SetWorkspaceNil

`func (o *PolicyCrossType) SetWorkspaceNil(b bool)`

 SetWorkspaceNil sets the value for Workspace to be an explicit nil

### UnsetWorkspace
`func (o *PolicyCrossType) UnsetWorkspace()`

UnsetWorkspace ensures that no value is present for Workspace, not even an explicit nil
### GetTraining

`func (o *PolicyCrossType) GetTraining() TrainingPolicyV2`

GetTraining returns the Training field if non-nil, zero value otherwise.

### GetTrainingOk

`func (o *PolicyCrossType) GetTrainingOk() (*TrainingPolicyV2, bool)`

GetTrainingOk returns a tuple with the Training field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraining

`func (o *PolicyCrossType) SetTraining(v TrainingPolicyV2)`

SetTraining sets Training field to given value.

### HasTraining

`func (o *PolicyCrossType) HasTraining() bool`

HasTraining returns a boolean if a field has been set.

### SetTrainingNil

`func (o *PolicyCrossType) SetTrainingNil(b bool)`

 SetTrainingNil sets the value for Training to be an explicit nil

### UnsetTraining
`func (o *PolicyCrossType) UnsetTraining()`

UnsetTraining ensures that no value is present for Training, not even an explicit nil
### GetDistributed

`func (o *PolicyCrossType) GetDistributed() DistributedPolicyV2`

GetDistributed returns the Distributed field if non-nil, zero value otherwise.

### GetDistributedOk

`func (o *PolicyCrossType) GetDistributedOk() (*DistributedPolicyV2, bool)`

GetDistributedOk returns a tuple with the Distributed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributed

`func (o *PolicyCrossType) SetDistributed(v DistributedPolicyV2)`

SetDistributed sets Distributed field to given value.

### HasDistributed

`func (o *PolicyCrossType) HasDistributed() bool`

HasDistributed returns a boolean if a field has been set.

### SetDistributedNil

`func (o *PolicyCrossType) SetDistributedNil(b bool)`

 SetDistributedNil sets the value for Distributed to be an explicit nil

### UnsetDistributed
`func (o *PolicyCrossType) UnsetDistributed()`

UnsetDistributed ensures that no value is present for Distributed, not even an explicit nil
### GetInference

`func (o *PolicyCrossType) GetInference() InferencePolicyV2`

GetInference returns the Inference field if non-nil, zero value otherwise.

### GetInferenceOk

`func (o *PolicyCrossType) GetInferenceOk() (*InferencePolicyV2, bool)`

GetInferenceOk returns a tuple with the Inference field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInference

`func (o *PolicyCrossType) SetInference(v InferencePolicyV2)`

SetInference sets Inference field to given value.

### HasInference

`func (o *PolicyCrossType) HasInference() bool`

HasInference returns a boolean if a field has been set.

### SetInferenceNil

`func (o *PolicyCrossType) SetInferenceNil(b bool)`

 SetInferenceNil sets the value for Inference to be an explicit nil

### UnsetInference
`func (o *PolicyCrossType) UnsetInference()`

UnsetInference ensures that no value is present for Inference, not even an explicit nil
### GetDistributedInference

`func (o *PolicyCrossType) GetDistributedInference() DistributedInferencePolicyV2`

GetDistributedInference returns the DistributedInference field if non-nil, zero value otherwise.

### GetDistributedInferenceOk

`func (o *PolicyCrossType) GetDistributedInferenceOk() (*DistributedInferencePolicyV2, bool)`

GetDistributedInferenceOk returns a tuple with the DistributedInference field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributedInference

`func (o *PolicyCrossType) SetDistributedInference(v DistributedInferencePolicyV2)`

SetDistributedInference sets DistributedInference field to given value.

### HasDistributedInference

`func (o *PolicyCrossType) HasDistributedInference() bool`

HasDistributedInference returns a boolean if a field has been set.

### SetDistributedInferenceNil

`func (o *PolicyCrossType) SetDistributedInferenceNil(b bool)`

 SetDistributedInferenceNil sets the value for DistributedInference to be an explicit nil

### UnsetDistributedInference
`func (o *PolicyCrossType) UnsetDistributedInference()`

UnsetDistributedInference ensures that no value is present for DistributedInference, not even an explicit nil
### GetNimService

`func (o *PolicyCrossType) GetNimService() NimServicePolicy`

GetNimService returns the NimService field if non-nil, zero value otherwise.

### GetNimServiceOk

`func (o *PolicyCrossType) GetNimServiceOk() (*NimServicePolicy, bool)`

GetNimServiceOk returns a tuple with the NimService field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNimService

`func (o *PolicyCrossType) SetNimService(v NimServicePolicy)`

SetNimService sets NimService field to given value.

### HasNimService

`func (o *PolicyCrossType) HasNimService() bool`

HasNimService returns a boolean if a field has been set.

### SetNimServiceNil

`func (o *PolicyCrossType) SetNimServiceNil(b bool)`

 SetNimServiceNil sets the value for NimService to be an explicit nil

### UnsetNimService
`func (o *PolicyCrossType) UnsetNimService()`

UnsetNimService ensures that no value is present for NimService, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


