# StorageClassSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Provisioner** | Pointer to **string** | The provisioner type (unique name) | [optional] 
**AllowVolumeExpansion** | Pointer to **bool** | Indicates whether the storage class allows volume expansion | [optional] 
**IsDefault** | **bool** | Indicates whether the storage class is the default one | 

## Methods

### NewStorageClassSpec

`func NewStorageClassSpec(isDefault bool, ) *StorageClassSpec`

NewStorageClassSpec instantiates a new StorageClassSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassSpecWithDefaults

`func NewStorageClassSpecWithDefaults() *StorageClassSpec`

NewStorageClassSpecWithDefaults instantiates a new StorageClassSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetProvisioner

`func (o *StorageClassSpec) GetProvisioner() string`

GetProvisioner returns the Provisioner field if non-nil, zero value otherwise.

### GetProvisionerOk

`func (o *StorageClassSpec) GetProvisionerOk() (*string, bool)`

GetProvisionerOk returns a tuple with the Provisioner field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProvisioner

`func (o *StorageClassSpec) SetProvisioner(v string)`

SetProvisioner sets Provisioner field to given value.

### HasProvisioner

`func (o *StorageClassSpec) HasProvisioner() bool`

HasProvisioner returns a boolean if a field has been set.

### GetAllowVolumeExpansion

`func (o *StorageClassSpec) GetAllowVolumeExpansion() bool`

GetAllowVolumeExpansion returns the AllowVolumeExpansion field if non-nil, zero value otherwise.

### GetAllowVolumeExpansionOk

`func (o *StorageClassSpec) GetAllowVolumeExpansionOk() (*bool, bool)`

GetAllowVolumeExpansionOk returns a tuple with the AllowVolumeExpansion field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowVolumeExpansion

`func (o *StorageClassSpec) SetAllowVolumeExpansion(v bool)`

SetAllowVolumeExpansion sets AllowVolumeExpansion field to given value.

### HasAllowVolumeExpansion

`func (o *StorageClassSpec) HasAllowVolumeExpansion() bool`

HasAllowVolumeExpansion returns a boolean if a field has been set.

### GetIsDefault

`func (o *StorageClassSpec) GetIsDefault() bool`

GetIsDefault returns the IsDefault field if non-nil, zero value otherwise.

### GetIsDefaultOk

`func (o *StorageClassSpec) GetIsDefaultOk() (*bool, bool)`

GetIsDefaultOk returns a tuple with the IsDefault field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDefault

`func (o *StorageClassSpec) SetIsDefault(v bool)`

SetIsDefault sets IsDefault field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


