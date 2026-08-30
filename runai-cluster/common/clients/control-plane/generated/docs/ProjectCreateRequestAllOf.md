# ProjectCreateRequestAllOf

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | Project Name | [optional] 
**DepartmentId** | Pointer to **int32** | ID of the department that owns the project. | [optional] 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesCreate**](NodePoolAssignedResourcesCreate.md) | Resources assigned to this Project per Node Pool. | [optional] 
**Namespace** | Pointer to **string** | The name of an existing namespace to use for the project in the cluster. Supported only for cluster versions 2.12 or higher. | [optional] 

## Methods

### NewProjectCreateRequestAllOf

`func NewProjectCreateRequestAllOf() *ProjectCreateRequestAllOf`

NewProjectCreateRequestAllOf instantiates a new ProjectCreateRequestAllOf object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewProjectCreateRequestAllOfWithDefaults

`func NewProjectCreateRequestAllOfWithDefaults() *ProjectCreateRequestAllOf`

NewProjectCreateRequestAllOfWithDefaults instantiates a new ProjectCreateRequestAllOf object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ProjectCreateRequestAllOf) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ProjectCreateRequestAllOf) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ProjectCreateRequestAllOf) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *ProjectCreateRequestAllOf) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDepartmentId

`func (o *ProjectCreateRequestAllOf) GetDepartmentId() int32`

GetDepartmentId returns the DepartmentId field if non-nil, zero value otherwise.

### GetDepartmentIdOk

`func (o *ProjectCreateRequestAllOf) GetDepartmentIdOk() (*int32, bool)`

GetDepartmentIdOk returns a tuple with the DepartmentId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDepartmentId

`func (o *ProjectCreateRequestAllOf) SetDepartmentId(v int32)`

SetDepartmentId sets DepartmentId field to given value.

### HasDepartmentId

`func (o *ProjectCreateRequestAllOf) HasDepartmentId() bool`

HasDepartmentId returns a boolean if a field has been set.

### GetNodePoolsResources

`func (o *ProjectCreateRequestAllOf) GetNodePoolsResources() []NodePoolAssignedResourcesCreate`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *ProjectCreateRequestAllOf) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesCreate, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *ProjectCreateRequestAllOf) SetNodePoolsResources(v []NodePoolAssignedResourcesCreate)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *ProjectCreateRequestAllOf) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetNamespace

`func (o *ProjectCreateRequestAllOf) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ProjectCreateRequestAllOf) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ProjectCreateRequestAllOf) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.

### HasNamespace

`func (o *ProjectCreateRequestAllOf) HasNamespace() bool`

HasNamespace returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


