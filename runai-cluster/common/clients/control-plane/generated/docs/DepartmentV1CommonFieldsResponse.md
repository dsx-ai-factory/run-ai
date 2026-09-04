# DepartmentV1CommonFieldsResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesV1Response**](NodePoolAssignedResourcesV1Response.md) | Resources assigned to the Department per node pool. | [optional] 
**Name** | Pointer to **string** | The name of the department. | [optional] 
**DeservedGpus** | **float32** | Deprecated. Instead, use &#x60;deserved&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Deserved GPUs for the department. | 
**AllowOverQuota** | Pointer to **bool** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Is over quota allowed for the department. | [optional] 
**MaxAllowedGpus** | **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Max allowed GPUs for the department. | 
**Resources** | Pointer to [**DepartmentV1CommonFieldsResponseResources**](DepartmentV1CommonFieldsResponseResources.md) |  | [optional] 

## Methods

### NewDepartmentV1CommonFieldsResponse

`func NewDepartmentV1CommonFieldsResponse(deservedGpus float32, maxAllowedGpus float32, ) *DepartmentV1CommonFieldsResponse`

NewDepartmentV1CommonFieldsResponse instantiates a new DepartmentV1CommonFieldsResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentV1CommonFieldsResponseWithDefaults

`func NewDepartmentV1CommonFieldsResponseWithDefaults() *DepartmentV1CommonFieldsResponse`

NewDepartmentV1CommonFieldsResponseWithDefaults instantiates a new DepartmentV1CommonFieldsResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *DepartmentV1CommonFieldsResponse) GetNodePoolsResources() []NodePoolAssignedResourcesV1Response`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentV1CommonFieldsResponse) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesV1Response, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentV1CommonFieldsResponse) SetNodePoolsResources(v []NodePoolAssignedResourcesV1Response)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentV1CommonFieldsResponse) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetName

`func (o *DepartmentV1CommonFieldsResponse) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentV1CommonFieldsResponse) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentV1CommonFieldsResponse) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DepartmentV1CommonFieldsResponse) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *DepartmentV1CommonFieldsResponse) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *DepartmentV1CommonFieldsResponse) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *DepartmentV1CommonFieldsResponse) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.


### GetAllowOverQuota

`func (o *DepartmentV1CommonFieldsResponse) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *DepartmentV1CommonFieldsResponse) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *DepartmentV1CommonFieldsResponse) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *DepartmentV1CommonFieldsResponse) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *DepartmentV1CommonFieldsResponse) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *DepartmentV1CommonFieldsResponse) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *DepartmentV1CommonFieldsResponse) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.


### GetResources

`func (o *DepartmentV1CommonFieldsResponse) GetResources() DepartmentV1CommonFieldsResponseResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentV1CommonFieldsResponse) GetResourcesOk() (*DepartmentV1CommonFieldsResponseResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentV1CommonFieldsResponse) SetResources(v DepartmentV1CommonFieldsResponseResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DepartmentV1CommonFieldsResponse) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


