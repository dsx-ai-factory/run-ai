# DepartmentV1CommonFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResourcesV1**](NodePoolAssignedResourcesV1.md) | Resources assigned to the Department per node pool. | [optional] 
**Name** | Pointer to **string** | The name of the department. | [optional] 
**DeservedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;deserved&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Deserved GPUs for the department. | [optional] 
**AllowOverQuota** | Pointer to **bool** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Is over quota allowed for the department. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Max allowed GPUs for the department. | [optional] 
**Resources** | Pointer to [**AssignedResourcesV1**](AssignedResourcesV1.md) | Deprecated. Instead, use &#39;nodePoolsResources&#39;. Total resources assigned to the Department. Can only be used in PUT/POST when there is a single Node Pool in the system. The resources returned in GET are the sum of all Node Pool Resources. | [optional] 

## Methods

### NewDepartmentV1CommonFields

`func NewDepartmentV1CommonFields() *DepartmentV1CommonFields`

NewDepartmentV1CommonFields instantiates a new DepartmentV1CommonFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentV1CommonFieldsWithDefaults

`func NewDepartmentV1CommonFieldsWithDefaults() *DepartmentV1CommonFields`

NewDepartmentV1CommonFieldsWithDefaults instantiates a new DepartmentV1CommonFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *DepartmentV1CommonFields) GetNodePoolsResources() []NodePoolAssignedResourcesV1`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentV1CommonFields) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResourcesV1, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentV1CommonFields) SetNodePoolsResources(v []NodePoolAssignedResourcesV1)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentV1CommonFields) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetName

`func (o *DepartmentV1CommonFields) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentV1CommonFields) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentV1CommonFields) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DepartmentV1CommonFields) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *DepartmentV1CommonFields) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *DepartmentV1CommonFields) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *DepartmentV1CommonFields) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *DepartmentV1CommonFields) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetAllowOverQuota

`func (o *DepartmentV1CommonFields) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *DepartmentV1CommonFields) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *DepartmentV1CommonFields) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *DepartmentV1CommonFields) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *DepartmentV1CommonFields) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *DepartmentV1CommonFields) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *DepartmentV1CommonFields) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *DepartmentV1CommonFields) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetResources

`func (o *DepartmentV1CommonFields) GetResources() AssignedResourcesV1`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentV1CommonFields) GetResourcesOk() (*AssignedResourcesV1, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentV1CommonFields) SetResources(v AssignedResourcesV1)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DepartmentV1CommonFields) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


