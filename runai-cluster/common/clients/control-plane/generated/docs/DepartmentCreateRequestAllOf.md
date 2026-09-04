# DepartmentCreateRequestAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesCreate**](NodePoolAssignedResourcesCreate.md) | Resources assigned to this Department per Node Pool | [optional] 

## Methods

### NewDepartmentCreateRequestAllOf

`func NewDepartmentCreateRequestAllOf() *DepartmentCreateRequestAllOf`

NewDepartmentCreateRequestAllOf instantiates a new DepartmentCreateRequestAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentCreateRequestAllOfWithDefaults

`func NewDepartmentCreateRequestAllOfWithDefaults() *DepartmentCreateRequestAllOf`

NewDepartmentCreateRequestAllOfWithDefaults instantiates a new DepartmentCreateRequestAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *DepartmentCreateRequestAllOf) GetNodePoolsResources() []NodePoolAssignedResourcesCreate`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentCreateRequestAllOf) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesCreate, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentCreateRequestAllOf) SetNodePoolsResources(v []NodePoolAssignedResourcesCreate)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentCreateRequestAllOf) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


