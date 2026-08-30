# StorageClassCustomization

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AccessMode** | Pointer to [**NullableAccessModeCustomization**](AccessModeCustomization.md) |  | [optional] 
**VolumeMode** | Pointer to [**NullableVolumeModeCustomization**](VolumeModeCustomization.md) |  | [optional] 
**ClaimSize** | Pointer to [**NullableClaimSizeCustomization**](ClaimSizeCustomization.md) |  | [optional] 
**Attributes** | Pointer to [**[]PvcAddedAttribute**](PvcAddedAttribute.md) | Defines additional annotation attributes to be applied to PVCs created under this storage class. Each attribute describes an annotation key and its expected value format, display name, and purpose. | [optional] 

## Methods

### NewStorageClassCustomization

`func NewStorageClassCustomization() *StorageClassCustomization`

NewStorageClassCustomization instantiates a new StorageClassCustomization object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassCustomizationWithDefaults

`func NewStorageClassCustomizationWithDefaults() *StorageClassCustomization`

NewStorageClassCustomizationWithDefaults instantiates a new StorageClassCustomization object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAccessMode

`func (o *StorageClassCustomization) GetAccessMode() AccessModeCustomization`

GetAccessMode returns the AccessMode field if non-nil, zero value otherwise.

### GetAccessModeOk

`func (o *StorageClassCustomization) GetAccessModeOk() (*AccessModeCustomization, bool)`

GetAccessModeOk returns a tuple with the AccessMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessMode

`func (o *StorageClassCustomization) SetAccessMode(v AccessModeCustomization)`

SetAccessMode sets AccessMode field to given value.

### HasAccessMode

`func (o *StorageClassCustomization) HasAccessMode() bool`

HasAccessMode returns a boolean if a field has been set.

### SetAccessModeNil

`func (o *StorageClassCustomization) SetAccessModeNil(b bool)`

 SetAccessModeNil sets the value for AccessMode to be an explicit nil

### UnsetAccessMode
`func (o *StorageClassCustomization) UnsetAccessMode()`

UnsetAccessMode ensures that no value is present for AccessMode, not even an explicit nil
### GetVolumeMode

`func (o *StorageClassCustomization) GetVolumeMode() VolumeModeCustomization`

GetVolumeMode returns the VolumeMode field if non-nil, zero value otherwise.

### GetVolumeModeOk

`func (o *StorageClassCustomization) GetVolumeModeOk() (*VolumeModeCustomization, bool)`

GetVolumeModeOk returns a tuple with the VolumeMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeMode

`func (o *StorageClassCustomization) SetVolumeMode(v VolumeModeCustomization)`

SetVolumeMode sets VolumeMode field to given value.

### HasVolumeMode

`func (o *StorageClassCustomization) HasVolumeMode() bool`

HasVolumeMode returns a boolean if a field has been set.

### SetVolumeModeNil

`func (o *StorageClassCustomization) SetVolumeModeNil(b bool)`

 SetVolumeModeNil sets the value for VolumeMode to be an explicit nil

### UnsetVolumeMode
`func (o *StorageClassCustomization) UnsetVolumeMode()`

UnsetVolumeMode ensures that no value is present for VolumeMode, not even an explicit nil
### GetClaimSize

`func (o *StorageClassCustomization) GetClaimSize() ClaimSizeCustomization`

GetClaimSize returns the ClaimSize field if non-nil, zero value otherwise.

### GetClaimSizeOk

`func (o *StorageClassCustomization) GetClaimSizeOk() (*ClaimSizeCustomization, bool)`

GetClaimSizeOk returns a tuple with the ClaimSize field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimSize

`func (o *StorageClassCustomization) SetClaimSize(v ClaimSizeCustomization)`

SetClaimSize sets ClaimSize field to given value.

### HasClaimSize

`func (o *StorageClassCustomization) HasClaimSize() bool`

HasClaimSize returns a boolean if a field has been set.

### SetClaimSizeNil

`func (o *StorageClassCustomization) SetClaimSizeNil(b bool)`

 SetClaimSizeNil sets the value for ClaimSize to be an explicit nil

### UnsetClaimSize
`func (o *StorageClassCustomization) UnsetClaimSize()`

UnsetClaimSize ensures that no value is present for ClaimSize, not even an explicit nil
### GetAttributes

`func (o *StorageClassCustomization) GetAttributes() []PvcAddedAttribute`

GetAttributes returns the Attributes field if non-nil, zero value otherwise.

### GetAttributesOk

`func (o *StorageClassCustomization) GetAttributesOk() (*[]PvcAddedAttribute, bool)`

GetAttributesOk returns a tuple with the Attributes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAttributes

`func (o *StorageClassCustomization) SetAttributes(v []PvcAddedAttribute)`

SetAttributes sets Attributes field to given value.

### HasAttributes

`func (o *StorageClassCustomization) HasAttributes() bool`

HasAttributes returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


