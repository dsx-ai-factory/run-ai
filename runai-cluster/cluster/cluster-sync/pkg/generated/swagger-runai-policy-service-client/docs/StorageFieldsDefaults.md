# StorageFieldsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataVolume** | Pointer to [**NullableDataVolumesDefaults**](DataVolumesDefaults.md) |  | [optional] 
**Pvc** | Pointer to [**NullablePvcsDefaults**](PvcsDefaults.md) |  | [optional] 
**Git** | Pointer to [**NullableGitsDefaults**](GitsDefaults.md) |  | [optional] 
**ConfigMapVolume** | Pointer to [**NullableConfigMapsDefaults**](ConfigMapsDefaults.md) |  | [optional] 
**SecretVolume** | Pointer to [**NullableSecretsDefaults**](SecretsDefaults.md) |  | [optional] 
**HostPath** | Pointer to [**NullableHostPathsDefaults**](HostPathsDefaults.md) |  | [optional] 
**Nfs** | Pointer to [**NullableNfssDefaults**](NfssDefaults.md) |  | [optional] 
**S3** | Pointer to [**NullableS3sDefaults**](S3sDefaults.md) |  | [optional] 

## Methods

### NewStorageFieldsDefaults

`func NewStorageFieldsDefaults() *StorageFieldsDefaults`

NewStorageFieldsDefaults instantiates a new StorageFieldsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageFieldsDefaultsWithDefaults

`func NewStorageFieldsDefaultsWithDefaults() *StorageFieldsDefaults`

NewStorageFieldsDefaultsWithDefaults instantiates a new StorageFieldsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataVolume

`func (o *StorageFieldsDefaults) GetDataVolume() DataVolumesDefaults`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *StorageFieldsDefaults) GetDataVolumeOk() (*DataVolumesDefaults, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *StorageFieldsDefaults) SetDataVolume(v DataVolumesDefaults)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *StorageFieldsDefaults) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *StorageFieldsDefaults) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *StorageFieldsDefaults) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetPvc

`func (o *StorageFieldsDefaults) GetPvc() PvcsDefaults`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *StorageFieldsDefaults) GetPvcOk() (*PvcsDefaults, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *StorageFieldsDefaults) SetPvc(v PvcsDefaults)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *StorageFieldsDefaults) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *StorageFieldsDefaults) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *StorageFieldsDefaults) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetGit

`func (o *StorageFieldsDefaults) GetGit() GitsDefaults`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *StorageFieldsDefaults) GetGitOk() (*GitsDefaults, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *StorageFieldsDefaults) SetGit(v GitsDefaults)`

SetGit sets Git field to given value.

### HasGit

`func (o *StorageFieldsDefaults) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *StorageFieldsDefaults) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *StorageFieldsDefaults) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetConfigMapVolume

`func (o *StorageFieldsDefaults) GetConfigMapVolume() ConfigMapsDefaults`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *StorageFieldsDefaults) GetConfigMapVolumeOk() (*ConfigMapsDefaults, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *StorageFieldsDefaults) SetConfigMapVolume(v ConfigMapsDefaults)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *StorageFieldsDefaults) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *StorageFieldsDefaults) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *StorageFieldsDefaults) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetSecretVolume

`func (o *StorageFieldsDefaults) GetSecretVolume() SecretsDefaults`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *StorageFieldsDefaults) GetSecretVolumeOk() (*SecretsDefaults, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *StorageFieldsDefaults) SetSecretVolume(v SecretsDefaults)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *StorageFieldsDefaults) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *StorageFieldsDefaults) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *StorageFieldsDefaults) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil
### GetHostPath

`func (o *StorageFieldsDefaults) GetHostPath() HostPathsDefaults`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *StorageFieldsDefaults) GetHostPathOk() (*HostPathsDefaults, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *StorageFieldsDefaults) SetHostPath(v HostPathsDefaults)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *StorageFieldsDefaults) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *StorageFieldsDefaults) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *StorageFieldsDefaults) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *StorageFieldsDefaults) GetNfs() NfssDefaults`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *StorageFieldsDefaults) GetNfsOk() (*NfssDefaults, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *StorageFieldsDefaults) SetNfs(v NfssDefaults)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *StorageFieldsDefaults) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *StorageFieldsDefaults) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *StorageFieldsDefaults) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetS3

`func (o *StorageFieldsDefaults) GetS3() S3sDefaults`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *StorageFieldsDefaults) GetS3Ok() (*S3sDefaults, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *StorageFieldsDefaults) SetS3(v S3sDefaults)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *StorageFieldsDefaults) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *StorageFieldsDefaults) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *StorageFieldsDefaults) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


