# WorkloadSupportedTypes

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Inference** | Pointer to **NullableBool** | Is inference a supported workload type. | [optional] 
**Workspace** | Pointer to **NullableBool** | Is workspace a supported workload type. | [optional] 
**Training** | Pointer to **NullableBool** | Is training a supported workload type. | [optional] 
**Distributed** | Pointer to **NullableBool** | Is distributed a supported workload type. | [optional] 
**DistFramework** | Pointer to **NullableString** | The distributed training framework used in the workload. Note: The TF, XGBoost and JAX frameworks are deprecated and will be removed in a future release.  | [optional] 

## Methods

### NewWorkloadSupportedTypes

`func NewWorkloadSupportedTypes() *WorkloadSupportedTypes`

NewWorkloadSupportedTypes instantiates a new WorkloadSupportedTypes object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadSupportedTypesWithDefaults

`func NewWorkloadSupportedTypesWithDefaults() *WorkloadSupportedTypes`

NewWorkloadSupportedTypesWithDefaults instantiates a new WorkloadSupportedTypes object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetInference

`func (o *WorkloadSupportedTypes) GetInference() bool`

GetInference returns the Inference field if non-nil, zero value otherwise.

### GetInferenceOk

`func (o *WorkloadSupportedTypes) GetInferenceOk() (*bool, bool)`

GetInferenceOk returns a tuple with the Inference field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInference

`func (o *WorkloadSupportedTypes) SetInference(v bool)`

SetInference sets Inference field to given value.

### HasInference

`func (o *WorkloadSupportedTypes) HasInference() bool`

HasInference returns a boolean if a field has been set.

### SetInferenceNil

`func (o *WorkloadSupportedTypes) SetInferenceNil(b bool)`

 SetInferenceNil sets the value for Inference to be an explicit nil

### UnsetInference
`func (o *WorkloadSupportedTypes) UnsetInference()`

UnsetInference ensures that no value is present for Inference, not even an explicit nil
### GetWorkspace

`func (o *WorkloadSupportedTypes) GetWorkspace() bool`

GetWorkspace returns the Workspace field if non-nil, zero value otherwise.

### GetWorkspaceOk

`func (o *WorkloadSupportedTypes) GetWorkspaceOk() (*bool, bool)`

GetWorkspaceOk returns a tuple with the Workspace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspace

`func (o *WorkloadSupportedTypes) SetWorkspace(v bool)`

SetWorkspace sets Workspace field to given value.

### HasWorkspace

`func (o *WorkloadSupportedTypes) HasWorkspace() bool`

HasWorkspace returns a boolean if a field has been set.

### SetWorkspaceNil

`func (o *WorkloadSupportedTypes) SetWorkspaceNil(b bool)`

 SetWorkspaceNil sets the value for Workspace to be an explicit nil

### UnsetWorkspace
`func (o *WorkloadSupportedTypes) UnsetWorkspace()`

UnsetWorkspace ensures that no value is present for Workspace, not even an explicit nil
### GetTraining

`func (o *WorkloadSupportedTypes) GetTraining() bool`

GetTraining returns the Training field if non-nil, zero value otherwise.

### GetTrainingOk

`func (o *WorkloadSupportedTypes) GetTrainingOk() (*bool, bool)`

GetTrainingOk returns a tuple with the Training field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraining

`func (o *WorkloadSupportedTypes) SetTraining(v bool)`

SetTraining sets Training field to given value.

### HasTraining

`func (o *WorkloadSupportedTypes) HasTraining() bool`

HasTraining returns a boolean if a field has been set.

### SetTrainingNil

`func (o *WorkloadSupportedTypes) SetTrainingNil(b bool)`

 SetTrainingNil sets the value for Training to be an explicit nil

### UnsetTraining
`func (o *WorkloadSupportedTypes) UnsetTraining()`

UnsetTraining ensures that no value is present for Training, not even an explicit nil
### GetDistributed

`func (o *WorkloadSupportedTypes) GetDistributed() bool`

GetDistributed returns the Distributed field if non-nil, zero value otherwise.

### GetDistributedOk

`func (o *WorkloadSupportedTypes) GetDistributedOk() (*bool, bool)`

GetDistributedOk returns a tuple with the Distributed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributed

`func (o *WorkloadSupportedTypes) SetDistributed(v bool)`

SetDistributed sets Distributed field to given value.

### HasDistributed

`func (o *WorkloadSupportedTypes) HasDistributed() bool`

HasDistributed returns a boolean if a field has been set.

### SetDistributedNil

`func (o *WorkloadSupportedTypes) SetDistributedNil(b bool)`

 SetDistributedNil sets the value for Distributed to be an explicit nil

### UnsetDistributed
`func (o *WorkloadSupportedTypes) UnsetDistributed()`

UnsetDistributed ensures that no value is present for Distributed, not even an explicit nil
### GetDistFramework

`func (o *WorkloadSupportedTypes) GetDistFramework() string`

GetDistFramework returns the DistFramework field if non-nil, zero value otherwise.

### GetDistFrameworkOk

`func (o *WorkloadSupportedTypes) GetDistFrameworkOk() (*string, bool)`

GetDistFrameworkOk returns a tuple with the DistFramework field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistFramework

`func (o *WorkloadSupportedTypes) SetDistFramework(v string)`

SetDistFramework sets DistFramework field to given value.

### HasDistFramework

`func (o *WorkloadSupportedTypes) HasDistFramework() bool`

HasDistFramework returns a boolean if a field has been set.

### SetDistFrameworkNil

`func (o *WorkloadSupportedTypes) SetDistFrameworkNil(b bool)`

 SetDistFrameworkNil sets the value for DistFramework to be an explicit nil

### UnsetDistFramework
`func (o *WorkloadSupportedTypes) UnsetDistFramework()`

UnsetDistFramework ensures that no value is present for DistFramework, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


