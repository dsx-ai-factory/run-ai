# DepartmentV1CreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesV1**](NodePoolAssignedResourcesV1.md) | Resources assigned to the Department per node pool. | [optional] 
**Name** | **string** | The name of the department. | 
**DeservedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;deserved&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Deserved GPUs for the department. | [optional] 
**AllowOverQuota** | Pointer to **bool** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Is over quota allowed for the department. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Max allowed GPUs for the department. | [optional] 
**Resources** | Pointer to [**DepartmentV1CommonFieldsResources**](DepartmentV1CommonFieldsResources.md) |  | [optional] 

## Methods

### NewDepartmentV1CreationRequest

`func NewDepartmentV1CreationRequest(name string, ) *DepartmentV1CreationRequest`

NewDepartmentV1CreationRequest instantiates a new DepartmentV1CreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentV1CreationRequestWithDefaults

`func NewDepartmentV1CreationRequestWithDefaults() *DepartmentV1CreationRequest`

NewDepartmentV1CreationRequestWithDefaults instantiates a new DepartmentV1CreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *DepartmentV1CreationRequest) GetNodePoolsResources() []NodePoolAssignedResourcesV1`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentV1CreationRequest) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesV1, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentV1CreationRequest) SetNodePoolsResources(v []NodePoolAssignedResourcesV1)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentV1CreationRequest) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetName

`func (o *DepartmentV1CreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentV1CreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentV1CreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetDeservedGpus

`func (o *DepartmentV1CreationRequest) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *DepartmentV1CreationRequest) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *DepartmentV1CreationRequest) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *DepartmentV1CreationRequest) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetAllowOverQuota

`func (o *DepartmentV1CreationRequest) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *DepartmentV1CreationRequest) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *DepartmentV1CreationRequest) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *DepartmentV1CreationRequest) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *DepartmentV1CreationRequest) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *DepartmentV1CreationRequest) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *DepartmentV1CreationRequest) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *DepartmentV1CreationRequest) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetResources

`func (o *DepartmentV1CreationRequest) GetResources() DepartmentV1CommonFieldsResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentV1CreationRequest) GetResourcesOk() (*DepartmentV1CommonFieldsResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentV1CreationRequest) SetResources(v DepartmentV1CommonFieldsResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DepartmentV1CreationRequest) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


