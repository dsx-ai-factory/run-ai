# PVCSyncDetails

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Size** | **string** | Size is the size of the PVC. | 
**StorageClass** | **string** | StorageClass is the name of the StorageClass required by the claim. If not specified, the default StorageClass will be used. | 
**AccessModes** | Pointer to [**NullablePvcAccessModes**](PvcAccessModes.md) |  | [optional] 
**VolumeMode** | Pointer to **NullableString** | The volume mode required by the claim, either Filesystem (default) or Block. | [optional] 
**ExistingPvc** | Pointer to **NullableBool** | Whether to assume that the PVC exists. If set to true, PVC is assumed to exist. If set to false, the PVC will be create if it does not exist. | [optional] [default to false]
**ClaimName** | **string** | A given name for the PVC. Allowed referencing it across workspaces. ClaimName is mandatory for creating a pvc. | 
**ReadOnly** | Pointer to **NullableBool** | Whether the path to the PVC permits only read access. | [optional] [default to true]

## Methods

### NewPVCSyncDetails

`func NewPVCSyncDetails(size string, storageClass string, claimName string, ) *PVCSyncDetails`

NewPVCSyncDetails instantiates a new PVCSyncDetails object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPVCSyncDetailsWithDefaults

`func NewPVCSyncDetailsWithDefaults() *PVCSyncDetails`

NewPVCSyncDetailsWithDefaults instantiates a new PVCSyncDetails object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSize

`func (o *PVCSyncDetails) GetSize() string`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *PVCSyncDetails) GetSizeOk() (*string, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *PVCSyncDetails) SetSize(v string)`

SetSize sets Size field to given value.


### GetStorageClass

`func (o *PVCSyncDetails) GetStorageClass() string`

GetStorageClass returns the StorageClass field if non-nil, zero value otherwise.

### GetStorageClassOk

`func (o *PVCSyncDetails) GetStorageClassOk() (*string, bool)`

GetStorageClassOk returns a tuple with the StorageClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClass

`func (o *PVCSyncDetails) SetStorageClass(v string)`

SetStorageClass sets StorageClass field to given value.


### GetAccessModes

`func (o *PVCSyncDetails) GetAccessModes() PvcAccessModes`

GetAccessModes returns the AccessModes field if non-nil, zero value otherwise.

### GetAccessModesOk

`func (o *PVCSyncDetails) GetAccessModesOk() (*PvcAccessModes, bool)`

GetAccessModesOk returns a tuple with the AccessModes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessModes

`func (o *PVCSyncDetails) SetAccessModes(v PvcAccessModes)`

SetAccessModes sets AccessModes field to given value.

### HasAccessModes

`func (o *PVCSyncDetails) HasAccessModes() bool`

HasAccessModes returns a boolean if a field has been set.

### SetAccessModesNil

`func (o *PVCSyncDetails) SetAccessModesNil(b bool)`

 SetAccessModesNil sets the value for AccessModes to be an explicit nil

### UnsetAccessModes
`func (o *PVCSyncDetails) UnsetAccessModes()`

UnsetAccessModes ensures that no value is present for AccessModes, not even an explicit nil
### GetVolumeMode

`func (o *PVCSyncDetails) GetVolumeMode() string`

GetVolumeMode returns the VolumeMode field if non-nil, zero value otherwise.

### GetVolumeModeOk

`func (o *PVCSyncDetails) GetVolumeModeOk() (*string, bool)`

GetVolumeModeOk returns a tuple with the VolumeMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeMode

`func (o *PVCSyncDetails) SetVolumeMode(v string)`

SetVolumeMode sets VolumeMode field to given value.

### HasVolumeMode

`func (o *PVCSyncDetails) HasVolumeMode() bool`

HasVolumeMode returns a boolean if a field has been set.

### SetVolumeModeNil

`func (o *PVCSyncDetails) SetVolumeModeNil(b bool)`

 SetVolumeModeNil sets the value for VolumeMode to be an explicit nil

### UnsetVolumeMode
`func (o *PVCSyncDetails) UnsetVolumeMode()`

UnsetVolumeMode ensures that no value is present for VolumeMode, not even an explicit nil
### GetExistingPvc

`func (o *PVCSyncDetails) GetExistingPvc() bool`

GetExistingPvc returns the ExistingPvc field if non-nil, zero value otherwise.

### GetExistingPvcOk

`func (o *PVCSyncDetails) GetExistingPvcOk() (*bool, bool)`

GetExistingPvcOk returns a tuple with the ExistingPvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetExistingPvc

`func (o *PVCSyncDetails) SetExistingPvc(v bool)`

SetExistingPvc sets ExistingPvc field to given value.

### HasExistingPvc

`func (o *PVCSyncDetails) HasExistingPvc() bool`

HasExistingPvc returns a boolean if a field has been set.

### SetExistingPvcNil

`func (o *PVCSyncDetails) SetExistingPvcNil(b bool)`

 SetExistingPvcNil sets the value for ExistingPvc to be an explicit nil

### UnsetExistingPvc
`func (o *PVCSyncDetails) UnsetExistingPvc()`

UnsetExistingPvc ensures that no value is present for ExistingPvc, not even an explicit nil
### GetClaimName

`func (o *PVCSyncDetails) GetClaimName() string`

GetClaimName returns the ClaimName field if non-nil, zero value otherwise.

### GetClaimNameOk

`func (o *PVCSyncDetails) GetClaimNameOk() (*string, bool)`

GetClaimNameOk returns a tuple with the ClaimName field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClaimName

`func (o *PVCSyncDetails) SetClaimName(v string)`

SetClaimName sets ClaimName field to given value.


### GetReadOnly

`func (o *PVCSyncDetails) GetReadOnly() bool`

GetReadOnly returns the ReadOnly field if non-nil, zero value otherwise.

### GetReadOnlyOk

`func (o *PVCSyncDetails) GetReadOnlyOk() (*bool, bool)`

GetReadOnlyOk returns a tuple with the ReadOnly field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReadOnly

`func (o *PVCSyncDetails) SetReadOnly(v bool)`

SetReadOnly sets ReadOnly field to given value.

### HasReadOnly

`func (o *PVCSyncDetails) HasReadOnly() bool`

HasReadOnly returns a boolean if a field has been set.

### SetReadOnlyNil

`func (o *PVCSyncDetails) SetReadOnlyNil(b bool)`

 SetReadOnlyNil sets the value for ReadOnly to be an explicit nil

### UnsetReadOnly
`func (o *PVCSyncDetails) UnsetReadOnly()`

UnsetReadOnly ensures that no value is present for ReadOnly, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


