# NodeTypesPerWorkload

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Training** | Pointer to **[]string** | ids of the node types for training workloads | [optional] 
**Workspace** | Pointer to **[]string** | ids of the node types for workspace workloads | [optional] 
**Names** | Pointer to **map[string]string** | mapping of ids to names | [optional] 

## Methods

### NewNodeTypesPerWorkload

`func NewNodeTypesPerWorkload() *NodeTypesPerWorkload`

NewNodeTypesPerWorkload instantiates a new NodeTypesPerWorkload object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNodeTypesPerWorkloadWithDefaults

`func NewNodeTypesPerWorkloadWithDefaults() *NodeTypesPerWorkload`

NewNodeTypesPerWorkloadWithDefaults instantiates a new NodeTypesPerWorkload object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetTraining

`func (o *NodeTypesPerWorkload) GetTraining() []string`

GetTraining returns the Training field if non-nil, zero value otherwise.

### GetTrainingOk

`func (o *NodeTypesPerWorkload) GetTrainingOk() (*[]string, bool)`

GetTrainingOk returns a tuple with the Training field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetTraining

`func (o *NodeTypesPerWorkload) SetTraining(v []string)`

SetTraining sets Training field to given value.

### HasTraining

`func (o *NodeTypesPerWorkload) HasTraining() bool`

HasTraining returns a boolean if a field has been set.

### GetWorkspace

`func (o *NodeTypesPerWorkload) GetWorkspace() []string`

GetWorkspace returns the Workspace field if non-nil, zero value otherwise.

### GetWorkspaceOk

`func (o *NodeTypesPerWorkload) GetWorkspaceOk() (*[]string, bool)`

GetWorkspaceOk returns a tuple with the Workspace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetWorkspace

`func (o *NodeTypesPerWorkload) SetWorkspace(v []string)`

SetWorkspace sets Workspace field to given value.

### HasWorkspace

`func (o *NodeTypesPerWorkload) HasWorkspace() bool`

HasWorkspace returns a boolean if a field has been set.

### GetNames

`func (o *NodeTypesPerWorkload) GetNames() map[string]string`

GetNames returns the Names field if non-nil, zero value otherwise.

### GetNamesOk

`func (o *NodeTypesPerWorkload) GetNamesOk() (*map[string]string, bool)`

GetNamesOk returns a tuple with the Names field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNames

`func (o *NodeTypesPerWorkload) SetNames(v map[string]string)`

SetNames sets Names field to given value.

### HasNames

`func (o *NodeTypesPerWorkload) HasNames() bool`

HasNames returns a boolean if a field has been set.

### SetNamesNil

`func (o *NodeTypesPerWorkload) SetNamesNil(b bool)`

 SetNamesNil sets the value for Names to be an explicit nil

### UnsetNames
`func (o *NodeTypesPerWorkload) UnsetNames()`

UnsetNames ensures that no value is present for Names, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


