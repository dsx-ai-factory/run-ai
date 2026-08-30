# DepartmentUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the department. | 
**DeservedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;deserved&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Deserved GPUs for the department. | [optional] 
**AllowOverQuota** | Pointer to **bool** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Is over quota allowed for the department. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Max allowed GPUs for the department. | [optional] 
**Resources** | Pointer to [**AssignedResources**](AssignedResources.md) | Deprecated. Instead, use &#39;nodePoolsResources&#39;. Total resources assigned to the Department. Can only be used in PUT/POST when there is a single Node Pool in the system. The resources returned in GET are the sum of all Node Pool Resources. | [optional] 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to this Department per Node Pool | [optional] 

## Methods

### NewDepartmentUpdateRequest

`func NewDepartmentUpdateRequest(name string, ) *DepartmentUpdateRequest`

NewDepartmentUpdateRequest instantiates a new DepartmentUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentUpdateRequestWithDefaults

`func NewDepartmentUpdateRequestWithDefaults() *DepartmentUpdateRequest`

NewDepartmentUpdateRequestWithDefaults instantiates a new DepartmentUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DepartmentUpdateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentUpdateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentUpdateRequest) SetName(v string)`

SetName sets Name field to given value.


### GetDeservedGpus

`func (o *DepartmentUpdateRequest) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *DepartmentUpdateRequest) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *DepartmentUpdateRequest) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *DepartmentUpdateRequest) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetAllowOverQuota

`func (o *DepartmentUpdateRequest) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *DepartmentUpdateRequest) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *DepartmentUpdateRequest) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *DepartmentUpdateRequest) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *DepartmentUpdateRequest) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *DepartmentUpdateRequest) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *DepartmentUpdateRequest) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *DepartmentUpdateRequest) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetResources

`func (o *DepartmentUpdateRequest) GetResources() AssignedResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentUpdateRequest) GetResourcesOk() (*AssignedResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentUpdateRequest) SetResources(v AssignedResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DepartmentUpdateRequest) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetNodePoolsResources

`func (o *DepartmentUpdateRequest) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentUpdateRequest) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentUpdateRequest) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentUpdateRequest) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


