# DepartmentUpdateRequest1

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to this Department per Node Pool | [optional] 
**Name** | **string** | The name of the department. | 
**DeservedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;deserved&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Deserved GPUs for the department. | [optional] 
**AllowOverQuota** | Pointer to **bool** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Is over quota allowed for the department. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Deprecated. Instead, use &#x60;maxAllowed&#x60; for the relevant resource type under &#x60;NodePoolResources&#x60;. Max allowed GPUs for the department. | [optional] 
**Resources** | Pointer to [**DepartmentCommonFieldsResources**](DepartmentCommonFieldsResources.md) |  | [optional] 

## Methods

### NewDepartmentUpdateRequest1

`func NewDepartmentUpdateRequest1(name string, ) *DepartmentUpdateRequest1`

NewDepartmentUpdateRequest1 instantiates a new DepartmentUpdateRequest1 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentUpdateRequest1WithDefaults

`func NewDepartmentUpdateRequest1WithDefaults() *DepartmentUpdateRequest1`

NewDepartmentUpdateRequest1WithDefaults instantiates a new DepartmentUpdateRequest1 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetNodePoolsResources

`func (o *DepartmentUpdateRequest1) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentUpdateRequest1) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentUpdateRequest1) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentUpdateRequest1) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.

### GetName

`func (o *DepartmentUpdateRequest1) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentUpdateRequest1) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentUpdateRequest1) SetName(v string)`

SetName sets Name field to given value.


### GetDeservedGpus

`func (o *DepartmentUpdateRequest1) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *DepartmentUpdateRequest1) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *DepartmentUpdateRequest1) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *DepartmentUpdateRequest1) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetAllowOverQuota

`func (o *DepartmentUpdateRequest1) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *DepartmentUpdateRequest1) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *DepartmentUpdateRequest1) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *DepartmentUpdateRequest1) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *DepartmentUpdateRequest1) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *DepartmentUpdateRequest1) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *DepartmentUpdateRequest1) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *DepartmentUpdateRequest1) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetResources

`func (o *DepartmentUpdateRequest1) GetResources() DepartmentCommonFieldsResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentUpdateRequest1) GetResourcesOk() (*DepartmentCommonFieldsResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentUpdateRequest1) SetResources(v DepartmentCommonFieldsResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DepartmentUpdateRequest1) HasResources() bool`

HasResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


