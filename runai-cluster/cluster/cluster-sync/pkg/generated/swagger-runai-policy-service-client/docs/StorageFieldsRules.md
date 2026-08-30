# StorageFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataVolume** | Pointer to [**NullableDataVolumesRules**](DataVolumesRules.md) |  | [optional] 
**Pvc** | Pointer to [**NullablePvcsRules**](PvcsRules.md) |  | [optional] 
**Git** | Pointer to [**NullableGitsRules**](GitsRules.md) |  | [optional] 
**ConfigMapVolume** | Pointer to [**NullableConfigMapsRules**](ConfigMapsRules.md) |  | [optional] 
**SecretVolume** | Pointer to [**NullableSecretsRules**](SecretsRules.md) |  | [optional] 
**HostPath** | Pointer to [**NullableHostPathsRules**](HostPathsRules.md) |  | [optional] 
**Nfs** | Pointer to [**NullableNfssRules**](NfssRules.md) |  | [optional] 
**S3** | Pointer to [**NullableS3sRules**](S3sRules.md) |  | [optional] 

## Methods

### NewStorageFieldsRules

`func NewStorageFieldsRules() *StorageFieldsRules`

NewStorageFieldsRules instantiates a new StorageFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageFieldsRulesWithDefaults

`func NewStorageFieldsRulesWithDefaults() *StorageFieldsRules`

NewStorageFieldsRulesWithDefaults instantiates a new StorageFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataVolume

`func (o *StorageFieldsRules) GetDataVolume() DataVolumesRules`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *StorageFieldsRules) GetDataVolumeOk() (*DataVolumesRules, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *StorageFieldsRules) SetDataVolume(v DataVolumesRules)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *StorageFieldsRules) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *StorageFieldsRules) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *StorageFieldsRules) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetPvc

`func (o *StorageFieldsRules) GetPvc() PvcsRules`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *StorageFieldsRules) GetPvcOk() (*PvcsRules, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *StorageFieldsRules) SetPvc(v PvcsRules)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *StorageFieldsRules) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *StorageFieldsRules) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *StorageFieldsRules) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetGit

`func (o *StorageFieldsRules) GetGit() GitsRules`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *StorageFieldsRules) GetGitOk() (*GitsRules, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *StorageFieldsRules) SetGit(v GitsRules)`

SetGit sets Git field to given value.

### HasGit

`func (o *StorageFieldsRules) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *StorageFieldsRules) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *StorageFieldsRules) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetConfigMapVolume

`func (o *StorageFieldsRules) GetConfigMapVolume() ConfigMapsRules`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *StorageFieldsRules) GetConfigMapVolumeOk() (*ConfigMapsRules, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *StorageFieldsRules) SetConfigMapVolume(v ConfigMapsRules)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *StorageFieldsRules) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *StorageFieldsRules) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *StorageFieldsRules) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetSecretVolume

`func (o *StorageFieldsRules) GetSecretVolume() SecretsRules`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *StorageFieldsRules) GetSecretVolumeOk() (*SecretsRules, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *StorageFieldsRules) SetSecretVolume(v SecretsRules)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *StorageFieldsRules) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *StorageFieldsRules) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *StorageFieldsRules) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil
### GetHostPath

`func (o *StorageFieldsRules) GetHostPath() HostPathsRules`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *StorageFieldsRules) GetHostPathOk() (*HostPathsRules, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *StorageFieldsRules) SetHostPath(v HostPathsRules)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *StorageFieldsRules) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *StorageFieldsRules) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *StorageFieldsRules) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *StorageFieldsRules) GetNfs() NfssRules`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *StorageFieldsRules) GetNfsOk() (*NfssRules, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *StorageFieldsRules) SetNfs(v NfssRules)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *StorageFieldsRules) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *StorageFieldsRules) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *StorageFieldsRules) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetS3

`func (o *StorageFieldsRules) GetS3() S3sRules`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *StorageFieldsRules) GetS3Ok() (*S3sRules, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *StorageFieldsRules) SetS3(v S3sRules)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *StorageFieldsRules) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *StorageFieldsRules) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *StorageFieldsRules) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


