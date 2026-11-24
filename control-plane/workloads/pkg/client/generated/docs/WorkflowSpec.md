# WorkflowSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Todo** | Pointer to **bool** |  | [optional] 
**Groups** | Pointer to [**[]WorkflowGroupsInner**](WorkflowGroupsInner.md) | Definitions of the workflow groups. | [optional] 
**Compute** | Pointer to [**[]WorkflowComputeProfile**](WorkflowComputeProfile.md) | Defines the workflow compute profiles. | [optional] 
**Placement** | Pointer to [**[]WorkflowPlacementProfile**](WorkflowPlacementProfile.md) | Defines the workflow placement profiles. | [optional] 
**Storage** | Pointer to [**[]WorkflowStorageProfile**](WorkflowStorageProfile.md) | Defines the workflow storage profiles. | [optional] 

## Methods

### NewWorkflowSpec

`func NewWorkflowSpec() *WorkflowSpec`

NewWorkflowSpec instantiates a new WorkflowSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowSpecWithDefaults

`func NewWorkflowSpecWithDefaults() *WorkflowSpec`

NewWorkflowSpecWithDefaults instantiates a new WorkflowSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTodo

`func (o *WorkflowSpec) GetTodo() bool`

GetTodo returns the Todo field if non-nil, zero value otherwise.

### GetTodoOk

`func (o *WorkflowSpec) GetTodoOk() (*bool, bool)`

GetTodoOk returns a tuple with the Todo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTodo

`func (o *WorkflowSpec) SetTodo(v bool)`

SetTodo sets Todo field to given value.

### HasTodo

`func (o *WorkflowSpec) HasTodo() bool`

HasTodo returns a boolean if a field has been set.

### GetGroups

`func (o *WorkflowSpec) GetGroups() []WorkflowGroupsInner`

GetGroups returns the Groups field if non-nil, zero value otherwise.

### GetGroupsOk

`func (o *WorkflowSpec) GetGroupsOk() (*[]WorkflowGroupsInner, bool)`

GetGroupsOk returns a tuple with the Groups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroups

`func (o *WorkflowSpec) SetGroups(v []WorkflowGroupsInner)`

SetGroups sets Groups field to given value.

### HasGroups

`func (o *WorkflowSpec) HasGroups() bool`

HasGroups returns a boolean if a field has been set.

### GetCompute

`func (o *WorkflowSpec) GetCompute() []WorkflowComputeProfile`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *WorkflowSpec) GetComputeOk() (*[]WorkflowComputeProfile, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *WorkflowSpec) SetCompute(v []WorkflowComputeProfile)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *WorkflowSpec) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### GetPlacement

`func (o *WorkflowSpec) GetPlacement() []WorkflowPlacementProfile`

GetPlacement returns the Placement field if non-nil, zero value otherwise.

### GetPlacementOk

`func (o *WorkflowSpec) GetPlacementOk() (*[]WorkflowPlacementProfile, bool)`

GetPlacementOk returns a tuple with the Placement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacement

`func (o *WorkflowSpec) SetPlacement(v []WorkflowPlacementProfile)`

SetPlacement sets Placement field to given value.

### HasPlacement

`func (o *WorkflowSpec) HasPlacement() bool`

HasPlacement returns a boolean if a field has been set.

### GetStorage

`func (o *WorkflowSpec) GetStorage() []WorkflowStorageProfile`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *WorkflowSpec) GetStorageOk() (*[]WorkflowStorageProfile, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *WorkflowSpec) SetStorage(v []WorkflowStorageProfile)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *WorkflowSpec) HasStorage() bool`

HasStorage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


