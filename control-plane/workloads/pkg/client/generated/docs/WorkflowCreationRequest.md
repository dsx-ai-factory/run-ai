# WorkflowCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**UseGivenNameAsPrefix** | Pointer to **bool** | When true, the requested name will be treated as a prefix. The final name of the workload will be composed of the name followed by a random set of characters. | [optional] [default to false]
**ProjectId** | **string** | The id of the project. | 
**ClusterId** | **string** | The id of the cluster. | 
**Todo** | Pointer to **bool** |  | [optional] 
**Groups** | Pointer to [**[]WorkflowGroupsInner**](WorkflowGroupsInner.md) | Definitions of the workflow groups. | [optional] 
**Compute** | Pointer to [**[]WorkflowComputeProfile**](WorkflowComputeProfile.md) | Defines the workflow compute profiles. | [optional] 
**Placement** | Pointer to [**[]WorkflowPlacementProfile**](WorkflowPlacementProfile.md) | Defines the workflow placement profiles. | [optional] 
**Storage** | Pointer to [**[]WorkflowStorageProfile**](WorkflowStorageProfile.md) | Defines the workflow storage profiles. | [optional] 

## Methods

### NewWorkflowCreationRequest

`func NewWorkflowCreationRequest(name string, projectId string, clusterId string, ) *WorkflowCreationRequest`

NewWorkflowCreationRequest instantiates a new WorkflowCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowCreationRequestWithDefaults

`func NewWorkflowCreationRequestWithDefaults() *WorkflowCreationRequest`

NewWorkflowCreationRequestWithDefaults instantiates a new WorkflowCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkflowCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkflowCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkflowCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetUseGivenNameAsPrefix

`func (o *WorkflowCreationRequest) GetUseGivenNameAsPrefix() bool`

GetUseGivenNameAsPrefix returns the UseGivenNameAsPrefix field if non-nil, zero value otherwise.

### GetUseGivenNameAsPrefixOk

`func (o *WorkflowCreationRequest) GetUseGivenNameAsPrefixOk() (*bool, bool)`

GetUseGivenNameAsPrefixOk returns a tuple with the UseGivenNameAsPrefix field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUseGivenNameAsPrefix

`func (o *WorkflowCreationRequest) SetUseGivenNameAsPrefix(v bool)`

SetUseGivenNameAsPrefix sets UseGivenNameAsPrefix field to given value.

### HasUseGivenNameAsPrefix

`func (o *WorkflowCreationRequest) HasUseGivenNameAsPrefix() bool`

HasUseGivenNameAsPrefix returns a boolean if a field has been set.

### GetProjectId

`func (o *WorkflowCreationRequest) GetProjectId() string`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkflowCreationRequest) GetProjectIdOk() (*string, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkflowCreationRequest) SetProjectId(v string)`

SetProjectId sets ProjectId field to given value.


### GetClusterId

`func (o *WorkflowCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkflowCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkflowCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetTodo

`func (o *WorkflowCreationRequest) GetTodo() bool`

GetTodo returns the Todo field if non-nil, zero value otherwise.

### GetTodoOk

`func (o *WorkflowCreationRequest) GetTodoOk() (*bool, bool)`

GetTodoOk returns a tuple with the Todo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTodo

`func (o *WorkflowCreationRequest) SetTodo(v bool)`

SetTodo sets Todo field to given value.

### HasTodo

`func (o *WorkflowCreationRequest) HasTodo() bool`

HasTodo returns a boolean if a field has been set.

### GetGroups

`func (o *WorkflowCreationRequest) GetGroups() []WorkflowGroupsInner`

GetGroups returns the Groups field if non-nil, zero value otherwise.

### GetGroupsOk

`func (o *WorkflowCreationRequest) GetGroupsOk() (*[]WorkflowGroupsInner, bool)`

GetGroupsOk returns a tuple with the Groups field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGroups

`func (o *WorkflowCreationRequest) SetGroups(v []WorkflowGroupsInner)`

SetGroups sets Groups field to given value.

### HasGroups

`func (o *WorkflowCreationRequest) HasGroups() bool`

HasGroups returns a boolean if a field has been set.

### GetCompute

`func (o *WorkflowCreationRequest) GetCompute() []WorkflowComputeProfile`

GetCompute returns the Compute field if non-nil, zero value otherwise.

### GetComputeOk

`func (o *WorkflowCreationRequest) GetComputeOk() (*[]WorkflowComputeProfile, bool)`

GetComputeOk returns a tuple with the Compute field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCompute

`func (o *WorkflowCreationRequest) SetCompute(v []WorkflowComputeProfile)`

SetCompute sets Compute field to given value.

### HasCompute

`func (o *WorkflowCreationRequest) HasCompute() bool`

HasCompute returns a boolean if a field has been set.

### GetPlacement

`func (o *WorkflowCreationRequest) GetPlacement() []WorkflowPlacementProfile`

GetPlacement returns the Placement field if non-nil, zero value otherwise.

### GetPlacementOk

`func (o *WorkflowCreationRequest) GetPlacementOk() (*[]WorkflowPlacementProfile, bool)`

GetPlacementOk returns a tuple with the Placement field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPlacement

`func (o *WorkflowCreationRequest) SetPlacement(v []WorkflowPlacementProfile)`

SetPlacement sets Placement field to given value.

### HasPlacement

`func (o *WorkflowCreationRequest) HasPlacement() bool`

HasPlacement returns a boolean if a field has been set.

### GetStorage

`func (o *WorkflowCreationRequest) GetStorage() []WorkflowStorageProfile`

GetStorage returns the Storage field if non-nil, zero value otherwise.

### GetStorageOk

`func (o *WorkflowCreationRequest) GetStorageOk() (*[]WorkflowStorageProfile, bool)`

GetStorageOk returns a tuple with the Storage field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorage

`func (o *WorkflowCreationRequest) SetStorage(v []WorkflowStorageProfile)`

SetStorage sets Storage field to given value.

### HasStorage

`func (o *WorkflowCreationRequest) HasStorage() bool`

HasStorage returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


