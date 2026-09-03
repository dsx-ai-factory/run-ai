# ClaimInfo

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Size** | Pointer to **NullableString** | Requested size for the PVC. Mandatory when existingPvc is false. Recommended sizes: TB/GB/MB/TIB/GIB/MIB | [optional] 
**StorageClass** | Pointer to **NullableString** | Storage class name to associate with the PVC. This parameter may be omitted if there is a single storage class in the system, or you are using the default storage class. For more information, see [Storage class](https://kubernetes.io/docs/concepts/storage/storage-classes). | [optional] 
**AccessModes** | Pointer to [**NullablePvcAccessModes**](PvcAccessModes.md) |  | [optional] 
**VolumeMode** | Pointer to [**NullablePvcVolumeMode**](PvcVolumeMode.md) |  | [optional] 
**AddedAttrValues** | Pointer to [**[]PvcAddedAttrValue**](PvcAddedAttrValue.md) | an optional array of key-values pairs that are written as annotations on the created PVC. the allowed attributes are determined according to the storage class configuration (see k8s-objects-tracker for further info). | [optional] 

## Methods

### NewClaimInfo

`func NewClaimInfo() *ClaimInfo`

NewClaimInfo instantiates a new ClaimInfo object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewClaimInfoWithDefaults

`func NewClaimInfoWithDefaults() *ClaimInfo`

NewClaimInfoWithDefaults instantiates a new ClaimInfo object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSize

`func (o *ClaimInfo) GetSize() string`

GetSize returns the Size field if non-nil, zero value otherwise.

### GetSizeOk

`func (o *ClaimInfo) GetSizeOk() (*string, bool)`

GetSizeOk returns a tuple with the Size field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSize

`func (o *ClaimInfo) SetSize(v string)`

SetSize sets Size field to given value.

### HasSize

`func (o *ClaimInfo) HasSize() bool`

HasSize returns a boolean if a field has been set.

### SetSizeNil

`func (o *ClaimInfo) SetSizeNil(b bool)`

 SetSizeNil sets the value for Size to be an explicit nil

### UnsetSize
`func (o *ClaimInfo) UnsetSize()`

UnsetSize ensures that no value is present for Size, not even an explicit nil
### GetStorageClass

`func (o *ClaimInfo) GetStorageClass() string`

GetStorageClass returns the StorageClass field if non-nil, zero value otherwise.

### GetStorageClassOk

`func (o *ClaimInfo) GetStorageClassOk() (*string, bool)`

GetStorageClassOk returns a tuple with the StorageClass field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStorageClass

`func (o *ClaimInfo) SetStorageClass(v string)`

SetStorageClass sets StorageClass field to given value.

### HasStorageClass

`func (o *ClaimInfo) HasStorageClass() bool`

HasStorageClass returns a boolean if a field has been set.

### SetStorageClassNil

`func (o *ClaimInfo) SetStorageClassNil(b bool)`

 SetStorageClassNil sets the value for StorageClass to be an explicit nil

### UnsetStorageClass
`func (o *ClaimInfo) UnsetStorageClass()`

UnsetStorageClass ensures that no value is present for StorageClass, not even an explicit nil
### GetAccessModes

`func (o *ClaimInfo) GetAccessModes() PvcAccessModes`

GetAccessModes returns the AccessModes field if non-nil, zero value otherwise.

### GetAccessModesOk

`func (o *ClaimInfo) GetAccessModesOk() (*PvcAccessModes, bool)`

GetAccessModesOk returns a tuple with the AccessModes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAccessModes

`func (o *ClaimInfo) SetAccessModes(v PvcAccessModes)`

SetAccessModes sets AccessModes field to given value.

### HasAccessModes

`func (o *ClaimInfo) HasAccessModes() bool`

HasAccessModes returns a boolean if a field has been set.

### SetAccessModesNil

`func (o *ClaimInfo) SetAccessModesNil(b bool)`

 SetAccessModesNil sets the value for AccessModes to be an explicit nil

### UnsetAccessModes
`func (o *ClaimInfo) UnsetAccessModes()`

UnsetAccessModes ensures that no value is present for AccessModes, not even an explicit nil
### GetVolumeMode

`func (o *ClaimInfo) GetVolumeMode() PvcVolumeMode`

GetVolumeMode returns the VolumeMode field if non-nil, zero value otherwise.

### GetVolumeModeOk

`func (o *ClaimInfo) GetVolumeModeOk() (*PvcVolumeMode, bool)`

GetVolumeModeOk returns a tuple with the VolumeMode field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetVolumeMode

`func (o *ClaimInfo) SetVolumeMode(v PvcVolumeMode)`

SetVolumeMode sets VolumeMode field to given value.

### HasVolumeMode

`func (o *ClaimInfo) HasVolumeMode() bool`

HasVolumeMode returns a boolean if a field has been set.

### SetVolumeModeNil

`func (o *ClaimInfo) SetVolumeModeNil(b bool)`

 SetVolumeModeNil sets the value for VolumeMode to be an explicit nil

### UnsetVolumeMode
`func (o *ClaimInfo) UnsetVolumeMode()`

UnsetVolumeMode ensures that no value is present for VolumeMode, not even an explicit nil
### GetAddedAttrValues

`func (o *ClaimInfo) GetAddedAttrValues() []PvcAddedAttrValue`

GetAddedAttrValues returns the AddedAttrValues field if non-nil, zero value otherwise.

### GetAddedAttrValuesOk

`func (o *ClaimInfo) GetAddedAttrValuesOk() (*[]PvcAddedAttrValue, bool)`

GetAddedAttrValuesOk returns a tuple with the AddedAttrValues field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAddedAttrValues

`func (o *ClaimInfo) SetAddedAttrValues(v []PvcAddedAttrValue)`

SetAddedAttrValues sets AddedAttrValues field to given value.

### HasAddedAttrValues

`func (o *ClaimInfo) HasAddedAttrValues() bool`

HasAddedAttrValues returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


