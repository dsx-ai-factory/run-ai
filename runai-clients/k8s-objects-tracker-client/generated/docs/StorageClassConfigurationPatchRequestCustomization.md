# StorageClassConfigurationPatchRequestCustomization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessMode** | Pointer to [**NullableAccessModeCustomization**](AccessModeCustomization.md) |  | [optional] 
**VolumeMode** | Pointer to [**NullableVolumeModeCustomization**](VolumeModeCustomization.md) |  | [optional] 
**ClaimSize** | Pointer to [**NullableClaimSizeCustomization**](ClaimSizeCustomization.md) |  | [optional] 
**Attributes** | Pointer to [**[]PvcAddedAttribute**](PvcAddedAttribute.md) | Defines additional annotation attributes to be applied to PVCs created under this storage class. Each attribute describes an annotation key and its expected value format, display name, and purpose. | [optional] 
**AttributesToRemove** | Pointer to **[]string** |  | [optional] 

## Methods

### NewStorageClassConfigurationPatchRequestCustomization

`func NewStorageClassConfigurationPatchRequestCustomization() *StorageClassConfigurationPatchRequestCustomization`

NewStorageClassConfigurationPatchRequestCustomization instantiates a new StorageClassConfigurationPatchRequestCustomization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassConfigurationPatchRequestCustomizationWithDefaults

`func NewStorageClassConfigurationPatchRequestCustomizationWithDefaults() *StorageClassConfigurationPatchRequestCustomization`

NewStorageClassConfigurationPatchRequestCustomizationWithDefaults instantiates a new StorageClassConfigurationPatchRequestCustomization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessMode

`func (o *StorageClassConfigurationPatchRequestCustomization) GetAccessMode() AccessModeCustomization`

GetAccessMode returns the AccessMode field if non-nil, zero value otherwise.

### GetAccessModeOk

`func (o *StorageClassConfigurationPatchRequestCustomization) GetAccessModeOk() (*AccessModeCustomization, bool)`

GetAccessModeOk returns a tuple with the AccessMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessMode

`func (o *StorageClassConfigurationPatchRequestCustomization) SetAccessMode(v AccessModeCustomization)`

SetAccessMode sets AccessMode field to given value.

### HasAccessMode

`func (o *StorageClassConfigurationPatchRequestCustomization) HasAccessMode() bool`

HasAccessMode returns a boolean if a field has been set.

### SetAccessModeNil

`func (o *StorageClassConfigurationPatchRequestCustomization) SetAccessModeNil(b bool)`

 SetAccessModeNil sets the value for AccessMode to be an explicit nil

### UnsetAccessMode
`func (o *StorageClassConfigurationPatchRequestCustomization) UnsetAccessMode()`

UnsetAccessMode ensures that no value is present for AccessMode, not even an explicit nil
### GetVolumeMode

`func (o *StorageClassConfigurationPatchRequestCustomization) GetVolumeMode() VolumeModeCustomization`

GetVolumeMode returns the VolumeMode field if non-nil, zero value otherwise.

### GetVolumeModeOk

`func (o *StorageClassConfigurationPatchRequestCustomization) GetVolumeModeOk() (*VolumeModeCustomization, bool)`

GetVolumeModeOk returns a tuple with the VolumeMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeMode

`func (o *StorageClassConfigurationPatchRequestCustomization) SetVolumeMode(v VolumeModeCustomization)`

SetVolumeMode sets VolumeMode field to given value.

### HasVolumeMode

`func (o *StorageClassConfigurationPatchRequestCustomization) HasVolumeMode() bool`

HasVolumeMode returns a boolean if a field has been set.

### SetVolumeModeNil

`func (o *StorageClassConfigurationPatchRequestCustomization) SetVolumeModeNil(b bool)`

 SetVolumeModeNil sets the value for VolumeMode to be an explicit nil

### UnsetVolumeMode
`func (o *StorageClassConfigurationPatchRequestCustomization) UnsetVolumeMode()`

UnsetVolumeMode ensures that no value is present for VolumeMode, not even an explicit nil
### GetClaimSize

`func (o *StorageClassConfigurationPatchRequestCustomization) GetClaimSize() ClaimSizeCustomization`

GetClaimSize returns the ClaimSize field if non-nil, zero value otherwise.

### GetClaimSizeOk

`func (o *StorageClassConfigurationPatchRequestCustomization) GetClaimSizeOk() (*ClaimSizeCustomization, bool)`

GetClaimSizeOk returns a tuple with the ClaimSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimSize

`func (o *StorageClassConfigurationPatchRequestCustomization) SetClaimSize(v ClaimSizeCustomization)`

SetClaimSize sets ClaimSize field to given value.

### HasClaimSize

`func (o *StorageClassConfigurationPatchRequestCustomization) HasClaimSize() bool`

HasClaimSize returns a boolean if a field has been set.

### SetClaimSizeNil

`func (o *StorageClassConfigurationPatchRequestCustomization) SetClaimSizeNil(b bool)`

 SetClaimSizeNil sets the value for ClaimSize to be an explicit nil

### UnsetClaimSize
`func (o *StorageClassConfigurationPatchRequestCustomization) UnsetClaimSize()`

UnsetClaimSize ensures that no value is present for ClaimSize, not even an explicit nil
### GetAttributes

`func (o *StorageClassConfigurationPatchRequestCustomization) GetAttributes() []PvcAddedAttribute`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *StorageClassConfigurationPatchRequestCustomization) GetAttributesOk() (*[]PvcAddedAttribute, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *StorageClassConfigurationPatchRequestCustomization) SetAttributes(v []PvcAddedAttribute)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *StorageClassConfigurationPatchRequestCustomization) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.

### GetAttributesToRemove

`func (o *StorageClassConfigurationPatchRequestCustomization) GetAttributesToRemove() []string`

GetAttributesToRemove returns the AttributesToRemove field if non-nil, zero value otherwise.

### GetAttributesToRemoveOk

`func (o *StorageClassConfigurationPatchRequestCustomization) GetAttributesToRemoveOk() (*[]string, bool)`

GetAttributesToRemoveOk returns a tuple with the AttributesToRemove field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributesToRemove

`func (o *StorageClassConfigurationPatchRequestCustomization) SetAttributesToRemove(v []string)`

SetAttributesToRemove sets AttributesToRemove field to given value.

### HasAttributesToRemove

`func (o *StorageClassConfigurationPatchRequestCustomization) HasAttributesToRemove() bool`

HasAttributesToRemove returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


