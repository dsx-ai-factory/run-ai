# SupersetDefaultsAllOfStorage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMapVolume** | Pointer to [**NullableConfigMapsDefaults**](ConfigMapsDefaults.md) |  | [optional] 
**DataVolume** | Pointer to [**NullableDataVolumesDefaults**](DataVolumesDefaults.md) |  | [optional] 
**EmptyDirVolume** | Pointer to [**NullableEmptyDirsDefaults**](EmptyDirsDefaults.md) |  | [optional] 
**Git** | Pointer to [**NullableGitsDefaults**](GitsDefaults.md) |  | [optional] 
**HostPath** | Pointer to [**NullableHostPathsDefaults**](HostPathsDefaults.md) |  | [optional] 
**Nfs** | Pointer to [**NullableNfssDefaults**](NfssDefaults.md) |  | [optional] 
**Pvc** | Pointer to [**NullablePvcsDefaults**](PvcsDefaults.md) |  | [optional] 
**S3** | Pointer to [**NullableS3sDefaults**](S3sDefaults.md) |  | [optional] 
**SecretVolume** | Pointer to [**NullableSecretsDefaults**](SecretsDefaults.md) |  | [optional] 

## Methods

### NewSupersetDefaultsAllOfStorage

`func NewSupersetDefaultsAllOfStorage() *SupersetDefaultsAllOfStorage`

NewSupersetDefaultsAllOfStorage instantiates a new SupersetDefaultsAllOfStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetDefaultsAllOfStorageWithDefaults

`func NewSupersetDefaultsAllOfStorageWithDefaults() *SupersetDefaultsAllOfStorage`

NewSupersetDefaultsAllOfStorageWithDefaults instantiates a new SupersetDefaultsAllOfStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *SupersetDefaultsAllOfStorage) GetConfigMapVolume() ConfigMapsDefaults`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *SupersetDefaultsAllOfStorage) GetConfigMapVolumeOk() (*ConfigMapsDefaults, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *SupersetDefaultsAllOfStorage) SetConfigMapVolume(v ConfigMapsDefaults)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *SupersetDefaultsAllOfStorage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *SupersetDefaultsAllOfStorage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *SupersetDefaultsAllOfStorage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetDataVolume

`func (o *SupersetDefaultsAllOfStorage) GetDataVolume() DataVolumesDefaults`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *SupersetDefaultsAllOfStorage) GetDataVolumeOk() (*DataVolumesDefaults, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *SupersetDefaultsAllOfStorage) SetDataVolume(v DataVolumesDefaults)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *SupersetDefaultsAllOfStorage) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *SupersetDefaultsAllOfStorage) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *SupersetDefaultsAllOfStorage) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *SupersetDefaultsAllOfStorage) GetEmptyDirVolume() EmptyDirsDefaults`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *SupersetDefaultsAllOfStorage) GetEmptyDirVolumeOk() (*EmptyDirsDefaults, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *SupersetDefaultsAllOfStorage) SetEmptyDirVolume(v EmptyDirsDefaults)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *SupersetDefaultsAllOfStorage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *SupersetDefaultsAllOfStorage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *SupersetDefaultsAllOfStorage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetGit

`func (o *SupersetDefaultsAllOfStorage) GetGit() GitsDefaults`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *SupersetDefaultsAllOfStorage) GetGitOk() (*GitsDefaults, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *SupersetDefaultsAllOfStorage) SetGit(v GitsDefaults)`

SetGit sets Git field to given value.

### HasGit

`func (o *SupersetDefaultsAllOfStorage) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *SupersetDefaultsAllOfStorage) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *SupersetDefaultsAllOfStorage) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetHostPath

`func (o *SupersetDefaultsAllOfStorage) GetHostPath() HostPathsDefaults`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *SupersetDefaultsAllOfStorage) GetHostPathOk() (*HostPathsDefaults, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *SupersetDefaultsAllOfStorage) SetHostPath(v HostPathsDefaults)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *SupersetDefaultsAllOfStorage) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *SupersetDefaultsAllOfStorage) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *SupersetDefaultsAllOfStorage) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *SupersetDefaultsAllOfStorage) GetNfs() NfssDefaults`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *SupersetDefaultsAllOfStorage) GetNfsOk() (*NfssDefaults, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *SupersetDefaultsAllOfStorage) SetNfs(v NfssDefaults)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *SupersetDefaultsAllOfStorage) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *SupersetDefaultsAllOfStorage) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *SupersetDefaultsAllOfStorage) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetPvc

`func (o *SupersetDefaultsAllOfStorage) GetPvc() PvcsDefaults`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *SupersetDefaultsAllOfStorage) GetPvcOk() (*PvcsDefaults, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *SupersetDefaultsAllOfStorage) SetPvc(v PvcsDefaults)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *SupersetDefaultsAllOfStorage) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *SupersetDefaultsAllOfStorage) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *SupersetDefaultsAllOfStorage) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetS3

`func (o *SupersetDefaultsAllOfStorage) GetS3() S3sDefaults`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *SupersetDefaultsAllOfStorage) GetS3Ok() (*S3sDefaults, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *SupersetDefaultsAllOfStorage) SetS3(v S3sDefaults)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *SupersetDefaultsAllOfStorage) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *SupersetDefaultsAllOfStorage) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *SupersetDefaultsAllOfStorage) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil
### GetSecretVolume

`func (o *SupersetDefaultsAllOfStorage) GetSecretVolume() SecretsDefaults`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *SupersetDefaultsAllOfStorage) GetSecretVolumeOk() (*SecretsDefaults, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *SupersetDefaultsAllOfStorage) SetSecretVolume(v SecretsDefaults)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *SupersetDefaultsAllOfStorage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *SupersetDefaultsAllOfStorage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *SupersetDefaultsAllOfStorage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


