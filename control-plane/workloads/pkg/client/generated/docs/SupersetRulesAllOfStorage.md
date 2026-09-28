# SupersetRulesAllOfStorage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMapVolume** | Pointer to [**NullableConfigMapsRules**](ConfigMapsRules.md) |  | [optional] 
**DataVolume** | Pointer to [**NullableDataVolumesRules**](DataVolumesRules.md) |  | [optional] 
**EmptyDirVolume** | Pointer to [**NullableEmptyDirsRules**](EmptyDirsRules.md) |  | [optional] 
**Git** | Pointer to [**NullableGitsRules**](GitsRules.md) |  | [optional] 
**HostPath** | Pointer to [**NullableHostPathsRules**](HostPathsRules.md) |  | [optional] 
**Nfs** | Pointer to [**NullableNfssRules**](NfssRules.md) |  | [optional] 
**Pvc** | Pointer to [**NullablePvcsRules**](PvcsRules.md) |  | [optional] 
**S3** | Pointer to [**NullableS3sRules**](S3sRules.md) |  | [optional] 
**SecretVolume** | Pointer to [**NullableSecretsRules**](SecretsRules.md) |  | [optional] 

## Methods

### NewSupersetRulesAllOfStorage

`func NewSupersetRulesAllOfStorage() *SupersetRulesAllOfStorage`

NewSupersetRulesAllOfStorage instantiates a new SupersetRulesAllOfStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetRulesAllOfStorageWithDefaults

`func NewSupersetRulesAllOfStorageWithDefaults() *SupersetRulesAllOfStorage`

NewSupersetRulesAllOfStorageWithDefaults instantiates a new SupersetRulesAllOfStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *SupersetRulesAllOfStorage) GetConfigMapVolume() ConfigMapsRules`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *SupersetRulesAllOfStorage) GetConfigMapVolumeOk() (*ConfigMapsRules, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *SupersetRulesAllOfStorage) SetConfigMapVolume(v ConfigMapsRules)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *SupersetRulesAllOfStorage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *SupersetRulesAllOfStorage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *SupersetRulesAllOfStorage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetDataVolume

`func (o *SupersetRulesAllOfStorage) GetDataVolume() DataVolumesRules`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *SupersetRulesAllOfStorage) GetDataVolumeOk() (*DataVolumesRules, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *SupersetRulesAllOfStorage) SetDataVolume(v DataVolumesRules)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *SupersetRulesAllOfStorage) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *SupersetRulesAllOfStorage) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *SupersetRulesAllOfStorage) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *SupersetRulesAllOfStorage) GetEmptyDirVolume() EmptyDirsRules`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *SupersetRulesAllOfStorage) GetEmptyDirVolumeOk() (*EmptyDirsRules, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *SupersetRulesAllOfStorage) SetEmptyDirVolume(v EmptyDirsRules)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *SupersetRulesAllOfStorage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *SupersetRulesAllOfStorage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *SupersetRulesAllOfStorage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetGit

`func (o *SupersetRulesAllOfStorage) GetGit() GitsRules`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *SupersetRulesAllOfStorage) GetGitOk() (*GitsRules, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *SupersetRulesAllOfStorage) SetGit(v GitsRules)`

SetGit sets Git field to given value.

### HasGit

`func (o *SupersetRulesAllOfStorage) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *SupersetRulesAllOfStorage) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *SupersetRulesAllOfStorage) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetHostPath

`func (o *SupersetRulesAllOfStorage) GetHostPath() HostPathsRules`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *SupersetRulesAllOfStorage) GetHostPathOk() (*HostPathsRules, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *SupersetRulesAllOfStorage) SetHostPath(v HostPathsRules)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *SupersetRulesAllOfStorage) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *SupersetRulesAllOfStorage) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *SupersetRulesAllOfStorage) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *SupersetRulesAllOfStorage) GetNfs() NfssRules`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *SupersetRulesAllOfStorage) GetNfsOk() (*NfssRules, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *SupersetRulesAllOfStorage) SetNfs(v NfssRules)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *SupersetRulesAllOfStorage) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *SupersetRulesAllOfStorage) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *SupersetRulesAllOfStorage) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetPvc

`func (o *SupersetRulesAllOfStorage) GetPvc() PvcsRules`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *SupersetRulesAllOfStorage) GetPvcOk() (*PvcsRules, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *SupersetRulesAllOfStorage) SetPvc(v PvcsRules)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *SupersetRulesAllOfStorage) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *SupersetRulesAllOfStorage) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *SupersetRulesAllOfStorage) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetS3

`func (o *SupersetRulesAllOfStorage) GetS3() S3sRules`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *SupersetRulesAllOfStorage) GetS3Ok() (*S3sRules, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *SupersetRulesAllOfStorage) SetS3(v S3sRules)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *SupersetRulesAllOfStorage) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *SupersetRulesAllOfStorage) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *SupersetRulesAllOfStorage) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil
### GetSecretVolume

`func (o *SupersetRulesAllOfStorage) GetSecretVolume() SecretsRules`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *SupersetRulesAllOfStorage) GetSecretVolumeOk() (*SecretsRules, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *SupersetRulesAllOfStorage) SetSecretVolume(v SecretsRules)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *SupersetRulesAllOfStorage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *SupersetRulesAllOfStorage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *SupersetRulesAllOfStorage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


