# DepartmentUpdateRequest1AllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to this Department per Node Pool | [optional] 

## Methods

### NewDepartmentUpdateRequest1AllOf

`func NewDepartmentUpdateRequest1AllOf() *DepartmentUpdateRequest1AllOf`

NewDepartmentUpdateRequest1AllOf instantiates a new DepartmentUpdateRequest1AllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentUpdateRequest1AllOfWithDefaults

`func NewDepartmentUpdateRequest1AllOfWithDefaults() *DepartmentUpdateRequest1AllOf`

NewDepartmentUpdateRequest1AllOfWithDefaults instantiates a new DepartmentUpdateRequest1AllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *DepartmentUpdateRequest1AllOf) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentUpdateRequest1AllOf) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentUpdateRequest1AllOf) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentUpdateRequest1AllOf) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


