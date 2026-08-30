# DepartmentCreateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the department. | 
**DeservedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;deserved&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Deserved GPUs for the department. | [optional] 
**AllowOverQuota** | Pointer to **bool** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Is over quota allowed for the department. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Max allowed GPUs for the department. | [optional] 
**Resources** | Pointer to [**AssignedResources**](AssignedResources.md) | Deprecated. Instead, use &#39;nodePoolsResources&#39;. Total resources assigned to the Department. Can only be used in PUT/POST when there is a single Node Pool in the system. The resources returned in GET are the sum of all Node Pool Resources. | [optional] 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesCreate**](NodePoolAssignedResourcesCreate.md) | Resources assigned to this Department per Node Pool | [optional] 

## Methods

### NewDepartmentCreateRequest

`func NewDepartmentCreateRequest(name string, ) *DepartmentCreateRequest`

NewDepartmentCreateRequest instantiates a new DepartmentCreateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentCreateRequestWithDefaults

`func NewDepartmentCreateRequestWithDefaults() *DepartmentCreateRequest`

NewDepartmentCreateRequestWithDefaults instantiates a new DepartmentCreateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DepartmentCreateRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentCreateRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentCreateRequest) SetName(v string)`

SetName sets Name field to given value.


### GetDeservedGpus

`func (o *DepartmentCreateRequest) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *DepartmentCreateRequest) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *DepartmentCreateRequest) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *DepartmentCreateRequest) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetAllowOverQuota

`func (o *DepartmentCreateRequest) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *DepartmentCreateRequest) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *DepartmentCreateRequest) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *DepartmentCreateRequest) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *DepartmentCreateRequest) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *DepartmentCreateRequest) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *DepartmentCreateRequest) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *DepartmentCreateRequest) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetResources

`func (o *DepartmentCreateRequest) GetResources() AssignedResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentCreateRequest) GetResourcesOk() (*AssignedResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentCreateRequest) SetResources(v AssignedResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DepartmentCreateRequest) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetNodePoolsResources

`func (o *DepartmentCreateRequest) GetNodePoolsResources() []NodePoolAssignedResourcesCreate`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentCreateRequest) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesCreate, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentCreateRequest) SetNodePoolsResources(v []NodePoolAssignedResourcesCreate)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentCreateRequest) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


