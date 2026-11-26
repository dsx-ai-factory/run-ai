# WorkflowS3StorageRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **NullableString** | The name of the profile, which tasks use to reference it. It must be unique among the profiles of its kind. (mandatory) | [optional] 
**Type** | Pointer to [**NullableWorkflowStorageType**](WorkflowStorageType.md) |  | [optional] 
**Inputs** | Pointer to [**[]WorkflowStorageRefInput**](WorkflowStorageRefInput.md) | The outputs of other tasks that the task reads. | [optional] 
**Outputs** | Pointer to [**NullableWorkflowS3StorageRefOutputs**](WorkflowS3StorageRefOutputs.md) |  | [optional] 

## Methods

### NewWorkflowS3StorageRef

`func NewWorkflowS3StorageRef() *WorkflowS3StorageRef`

NewWorkflowS3StorageRef instantiates a new WorkflowS3StorageRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowS3StorageRefWithDefaults

`func NewWorkflowS3StorageRefWithDefaults() *WorkflowS3StorageRef`

NewWorkflowS3StorageRefWithDefaults instantiates a new WorkflowS3StorageRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkflowS3StorageRef) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkflowS3StorageRef) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkflowS3StorageRef) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *WorkflowS3StorageRef) HasName() bool`

HasName returns a boolean if a field has been set.

### SetNameNil

`func (o *WorkflowS3StorageRef) SetNameNil(b bool)`

 SetNameNil sets the value for Name to be an explicit nil

### UnsetName
`func (o *WorkflowS3StorageRef) UnsetName()`

UnsetName ensures that no value is present for Name, not even an explicit nil
### GetType

`func (o *WorkflowS3StorageRef) GetType() WorkflowStorageType`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *WorkflowS3StorageRef) GetTypeOk() (*WorkflowStorageType, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *WorkflowS3StorageRef) SetType(v WorkflowStorageType)`

SetType sets Type field to given value.

### HasType

`func (o *WorkflowS3StorageRef) HasType() bool`

HasType returns a boolean if a field has been set.

### SetTypeNil

`func (o *WorkflowS3StorageRef) SetTypeNil(b bool)`

 SetTypeNil sets the value for Type to be an explicit nil

### UnsetType
`func (o *WorkflowS3StorageRef) UnsetType()`

UnsetType ensures that no value is present for Type, not even an explicit nil
### GetInputs

`func (o *WorkflowS3StorageRef) GetInputs() []WorkflowStorageRefInput`

GetInputs returns the Inputs field if non-nil, zero value otherwise.

### GetInputsOk

`func (o *WorkflowS3StorageRef) GetInputsOk() (*[]WorkflowStorageRefInput, bool)`

GetInputsOk returns a tuple with the Inputs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetInputs

`func (o *WorkflowS3StorageRef) SetInputs(v []WorkflowStorageRefInput)`

SetInputs sets Inputs field to given value.

### HasInputs

`func (o *WorkflowS3StorageRef) HasInputs() bool`

HasInputs returns a boolean if a field has been set.

### SetInputsNil

`func (o *WorkflowS3StorageRef) SetInputsNil(b bool)`

 SetInputsNil sets the value for Inputs to be an explicit nil

### UnsetInputs
`func (o *WorkflowS3StorageRef) UnsetInputs()`

UnsetInputs ensures that no value is present for Inputs, not even an explicit nil
### GetOutputs

`func (o *WorkflowS3StorageRef) GetOutputs() WorkflowS3StorageRefOutputs`

GetOutputs returns the Outputs field if non-nil, zero value otherwise.

### GetOutputsOk

`func (o *WorkflowS3StorageRef) GetOutputsOk() (*WorkflowS3StorageRefOutputs, bool)`

GetOutputsOk returns a tuple with the Outputs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOutputs

`func (o *WorkflowS3StorageRef) SetOutputs(v WorkflowS3StorageRefOutputs)`

SetOutputs sets Outputs field to given value.

### HasOutputs

`func (o *WorkflowS3StorageRef) HasOutputs() bool`

HasOutputs returns a boolean if a field has been set.

### SetOutputsNil

`func (o *WorkflowS3StorageRef) SetOutputsNil(b bool)`

 SetOutputsNil sets the value for Outputs to be an explicit nil

### UnsetOutputs
`func (o *WorkflowS3StorageRef) UnsetOutputs()`

UnsetOutputs ensures that no value is present for Outputs, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


