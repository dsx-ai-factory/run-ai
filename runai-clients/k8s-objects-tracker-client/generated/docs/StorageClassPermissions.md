# StorageClassPermissions

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**AllowedForAssets** | Pointer to **NullableBool** | Determines whether PVC assets can be created using this storage class. | [optional] 
**AllowedForWorkloads** | Pointer to **NullableBool** | Allows workloads to include PVCs that use this storage class. When disabled, users cannot (a) include non-existing PVC assets created with this class in asset-based submissions, or (b) attach PVCs with existing&#x3D;false to workloads using this class. | [optional] 
**AllowedForEphemeralVolumes** | Pointer to **NullableBool** | Allows workloads to create ephemeral volumes using this storage class. When disabled, users cannot (a) create PVC assets with autoDelete&#x3D;true and ephemeral&#x3D;true, or (b) include PVC instances with isEphemeral&#x3D;false that use this storage class. | [optional] 
**AllowedForPersistentVolumes** | Pointer to **NullableBool** | Allows workloads to create persistent volumes using this storage class. When disabled, users cannot (a) create PVC assets with autoDelete&#x3D;true and ephemeral&#x3D;true, or (b) include PVC instances with isEphemeral&#x3D;false that use this storage class. | [optional] 
**AllowDataSharing** | Pointer to **NullableBool** | Allows creation of PVCs using a shared volume. | [optional] 

## Methods

### NewStorageClassPermissions

`func NewStorageClassPermissions() *StorageClassPermissions`

NewStorageClassPermissions instantiates a new StorageClassPermissions object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageClassPermissionsWithDefaults

`func NewStorageClassPermissionsWithDefaults() *StorageClassPermissions`

NewStorageClassPermissionsWithDefaults instantiates a new StorageClassPermissions object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetAllowedForAssets

`func (o *StorageClassPermissions) GetAllowedForAssets() bool`

GetAllowedForAssets returns the AllowedForAssets field if non-nil, zero value otherwise.

### GetAllowedForAssetsOk

`func (o *StorageClassPermissions) GetAllowedForAssetsOk() (*bool, bool)`

GetAllowedForAssetsOk returns a tuple with the AllowedForAssets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowedForAssets

`func (o *StorageClassPermissions) SetAllowedForAssets(v bool)`

SetAllowedForAssets sets AllowedForAssets field to given value.

### HasAllowedForAssets

`func (o *StorageClassPermissions) HasAllowedForAssets() bool`

HasAllowedForAssets returns a boolean if a field has been set.

### SetAllowedForAssetsNil

`func (o *StorageClassPermissions) SetAllowedForAssetsNil(b bool)`

 SetAllowedForAssetsNil sets the value for AllowedForAssets to be an explicit nil

### UnsetAllowedForAssets
`func (o *StorageClassPermissions) UnsetAllowedForAssets()`

UnsetAllowedForAssets ensures that no value is present for AllowedForAssets, not even an explicit nil
### GetAllowedForWorkloads

`func (o *StorageClassPermissions) GetAllowedForWorkloads() bool`

GetAllowedForWorkloads returns the AllowedForWorkloads field if non-nil, zero value otherwise.

### GetAllowedForWorkloadsOk

`func (o *StorageClassPermissions) GetAllowedForWorkloadsOk() (*bool, bool)`

GetAllowedForWorkloadsOk returns a tuple with the AllowedForWorkloads field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowedForWorkloads

`func (o *StorageClassPermissions) SetAllowedForWorkloads(v bool)`

SetAllowedForWorkloads sets AllowedForWorkloads field to given value.

### HasAllowedForWorkloads

`func (o *StorageClassPermissions) HasAllowedForWorkloads() bool`

HasAllowedForWorkloads returns a boolean if a field has been set.

### SetAllowedForWorkloadsNil

`func (o *StorageClassPermissions) SetAllowedForWorkloadsNil(b bool)`

 SetAllowedForWorkloadsNil sets the value for AllowedForWorkloads to be an explicit nil

### UnsetAllowedForWorkloads
`func (o *StorageClassPermissions) UnsetAllowedForWorkloads()`

UnsetAllowedForWorkloads ensures that no value is present for AllowedForWorkloads, not even an explicit nil
### GetAllowedForEphemeralVolumes

`func (o *StorageClassPermissions) GetAllowedForEphemeralVolumes() bool`

GetAllowedForEphemeralVolumes returns the AllowedForEphemeralVolumes field if non-nil, zero value otherwise.

### GetAllowedForEphemeralVolumesOk

`func (o *StorageClassPermissions) GetAllowedForEphemeralVolumesOk() (*bool, bool)`

GetAllowedForEphemeralVolumesOk returns a tuple with the AllowedForEphemeralVolumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowedForEphemeralVolumes

`func (o *StorageClassPermissions) SetAllowedForEphemeralVolumes(v bool)`

SetAllowedForEphemeralVolumes sets AllowedForEphemeralVolumes field to given value.

### HasAllowedForEphemeralVolumes

`func (o *StorageClassPermissions) HasAllowedForEphemeralVolumes() bool`

HasAllowedForEphemeralVolumes returns a boolean if a field has been set.

### SetAllowedForEphemeralVolumesNil

`func (o *StorageClassPermissions) SetAllowedForEphemeralVolumesNil(b bool)`

 SetAllowedForEphemeralVolumesNil sets the value for AllowedForEphemeralVolumes to be an explicit nil

### UnsetAllowedForEphemeralVolumes
`func (o *StorageClassPermissions) UnsetAllowedForEphemeralVolumes()`

UnsetAllowedForEphemeralVolumes ensures that no value is present for AllowedForEphemeralVolumes, not even an explicit nil
### GetAllowedForPersistentVolumes

`func (o *StorageClassPermissions) GetAllowedForPersistentVolumes() bool`

GetAllowedForPersistentVolumes returns the AllowedForPersistentVolumes field if non-nil, zero value otherwise.

### GetAllowedForPersistentVolumesOk

`func (o *StorageClassPermissions) GetAllowedForPersistentVolumesOk() (*bool, bool)`

GetAllowedForPersistentVolumesOk returns a tuple with the AllowedForPersistentVolumes field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowedForPersistentVolumes

`func (o *StorageClassPermissions) SetAllowedForPersistentVolumes(v bool)`

SetAllowedForPersistentVolumes sets AllowedForPersistentVolumes field to given value.

### HasAllowedForPersistentVolumes

`func (o *StorageClassPermissions) HasAllowedForPersistentVolumes() bool`

HasAllowedForPersistentVolumes returns a boolean if a field has been set.

### SetAllowedForPersistentVolumesNil

`func (o *StorageClassPermissions) SetAllowedForPersistentVolumesNil(b bool)`

 SetAllowedForPersistentVolumesNil sets the value for AllowedForPersistentVolumes to be an explicit nil

### UnsetAllowedForPersistentVolumes
`func (o *StorageClassPermissions) UnsetAllowedForPersistentVolumes()`

UnsetAllowedForPersistentVolumes ensures that no value is present for AllowedForPersistentVolumes, not even an explicit nil
### GetAllowDataSharing

`func (o *StorageClassPermissions) GetAllowDataSharing() bool`

GetAllowDataSharing returns the AllowDataSharing field if non-nil, zero value otherwise.

### GetAllowDataSharingOk

`func (o *StorageClassPermissions) GetAllowDataSharingOk() (*bool, bool)`

GetAllowDataSharingOk returns a tuple with the AllowDataSharing field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowDataSharing

`func (o *StorageClassPermissions) SetAllowDataSharing(v bool)`

SetAllowDataSharing sets AllowDataSharing field to given value.

### HasAllowDataSharing

`func (o *StorageClassPermissions) HasAllowDataSharing() bool`

HasAllowDataSharing returns a boolean if a field has been set.

### SetAllowDataSharingNil

`func (o *StorageClassPermissions) SetAllowDataSharingNil(b bool)`

 SetAllowDataSharingNil sets the value for AllowDataSharing to be an explicit nil

### UnsetAllowDataSharing
`func (o *StorageClassPermissions) UnsetAllowDataSharing()`

UnsetAllowDataSharing ensures that no value is present for AllowDataSharing, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


