# DepartmentUserDefinedProperties

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | Pointer to **string** | The name of the department. | [optional] 
**DeservedGpus** | Pointer to **float32** | Deserved GPUs for the department. | [optional] 
**AllowOverQuota** | Pointer to **bool** | Is over quota allowed for the department. | [optional] 
**MaxAllowedGpus** | Pointer to **float32** | Max allowed GPUs for the department. | [optional] 
**Resources** | Pointer to [**AssignedResources**](AssignedResources.md) |  | [optional] 
**NodePoolsResources** | Pointer to [**[]NodePoolAssignedResources**](NodePoolAssignedResources.md) | Resources assigned to the Department per node pool | [optional] 

## Methods

### NewDepartmentUserDefinedProperties

`func NewDepartmentUserDefinedProperties() *DepartmentUserDefinedProperties`

NewDepartmentUserDefinedProperties instantiates a new DepartmentUserDefinedProperties object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDepartmentUserDefinedPropertiesWithDefaults

`func NewDepartmentUserDefinedPropertiesWithDefaults() *DepartmentUserDefinedProperties`

NewDepartmentUserDefinedPropertiesWithDefaults instantiates a new DepartmentUserDefinedProperties object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *DepartmentUserDefinedProperties) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *DepartmentUserDefinedProperties) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *DepartmentUserDefinedProperties) SetName(v string)`

SetName sets Name field to given value.

### HasName

`func (o *DepartmentUserDefinedProperties) HasName() bool`

HasName returns a boolean if a field has been set.

### GetDeservedGpus

`func (o *DepartmentUserDefinedProperties) GetDeservedGpus() float32`

GetDeservedGpus returns the DeservedGpus field if non-nil, zero value otherwise.

### GetDeservedGpusOk

`func (o *DepartmentUserDefinedProperties) GetDeservedGpusOk() (*float32, bool)`

GetDeservedGpusOk returns a tuple with the DeservedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDeservedGpus

`func (o *DepartmentUserDefinedProperties) SetDeservedGpus(v float32)`

SetDeservedGpus sets DeservedGpus field to given value.

### HasDeservedGpus

`func (o *DepartmentUserDefinedProperties) HasDeservedGpus() bool`

HasDeservedGpus returns a boolean if a field has been set.

### GetAllowOverQuota

`func (o *DepartmentUserDefinedProperties) GetAllowOverQuota() bool`

GetAllowOverQuota returns the AllowOverQuota field if non-nil, zero value otherwise.

### GetAllowOverQuotaOk

`func (o *DepartmentUserDefinedProperties) GetAllowOverQuotaOk() (*bool, bool)`

GetAllowOverQuotaOk returns a tuple with the AllowOverQuota field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverQuota

`func (o *DepartmentUserDefinedProperties) SetAllowOverQuota(v bool)`

SetAllowOverQuota sets AllowOverQuota field to given value.

### HasAllowOverQuota

`func (o *DepartmentUserDefinedProperties) HasAllowOverQuota() bool`

HasAllowOverQuota returns a boolean if a field has been set.

### GetMaxAllowedGpus

`func (o *DepartmentUserDefinedProperties) GetMaxAllowedGpus() float32`

GetMaxAllowedGpus returns the MaxAllowedGpus field if non-nil, zero value otherwise.

### GetMaxAllowedGpusOk

`func (o *DepartmentUserDefinedProperties) GetMaxAllowedGpusOk() (*float32, bool)`

GetMaxAllowedGpusOk returns a tuple with the MaxAllowedGpus field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMaxAllowedGpus

`func (o *DepartmentUserDefinedProperties) SetMaxAllowedGpus(v float32)`

SetMaxAllowedGpus sets MaxAllowedGpus field to given value.

### HasMaxAllowedGpus

`func (o *DepartmentUserDefinedProperties) HasMaxAllowedGpus() bool`

HasMaxAllowedGpus returns a boolean if a field has been set.

### GetResources

`func (o *DepartmentUserDefinedProperties) GetResources() AssignedResources`

GetResources returns the Resources field if non-nil, zero value otherwise.

### GetResourcesOk

`func (o *DepartmentUserDefinedProperties) GetResourcesOk() (*AssignedResources, bool)`

GetResourcesOk returns a tuple with the Resources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetResources

`func (o *DepartmentUserDefinedProperties) SetResources(v AssignedResources)`

SetResources sets Resources field to given value.

### HasResources

`func (o *DepartmentUserDefinedProperties) HasResources() bool`

HasResources returns a boolean if a field has been set.

### GetNodePoolsResources

`func (o *DepartmentUserDefinedProperties) GetNodePoolsResources() []NodePoolAssignedResources`

GetNodePoolsResources returns the NodePoolsResources field if non-nil, zero value otherwise.

### GetNodePoolsResourcesOk

`func (o *DepartmentUserDefinedProperties) GetNodePoolsResourcesOk() (*[]NodePoolAssignedResources, bool)`

GetNodePoolsResourcesOk returns a tuple with the NodePoolsResources field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNodePoolsResources

`func (o *DepartmentUserDefinedProperties) SetNodePoolsResources(v []NodePoolAssignedResources)`

SetNodePoolsResources sets NodePoolsResources field to given value.

### HasNodePoolsResources

`func (o *DepartmentUserDefinedProperties) HasNodePoolsResources() bool`

HasNodePoolsResources returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


