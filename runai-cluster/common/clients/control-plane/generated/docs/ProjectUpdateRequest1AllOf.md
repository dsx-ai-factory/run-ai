# ProjectUpdateRequest1AllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to this Project per Node Pool. | [optional] 

## Methods

### NewProjectUpdateRequest1AllOf

`func NewProjectUpdateRequest1AllOf() *ProjectUpdateRequest1AllOf`

NewProjectUpdateRequest1AllOf instantiates a new ProjectUpdateRequest1AllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectUpdateRequest1AllOfWithDefaults

`func NewProjectUpdateRequest1AllOfWithDefaults() *ProjectUpdateRequest1AllOf`

NewProjectUpdateRequest1AllOfWithDefaults instantiates a new ProjectUpdateRequest1AllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *ProjectUpdateRequest1AllOf) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *ProjectUpdateRequest1AllOf) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *ProjectUpdateRequest1AllOf) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *ProjectUpdateRequest1AllOf) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


