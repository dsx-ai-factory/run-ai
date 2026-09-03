# AssetsRulesDatasources

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataVolume** | Pointer to [**NullableDataVolumesRules**](DataVolumesRules.md) |  | [optional] 
**Pvc** | Pointer to [**NullablePvcsRules**](PvcsRules.md) |  | [optional] 
**HostPath** | Pointer to [**NullableHostPathsRules**](HostPathsRules.md) |  | [optional] 
**Nfs** | Pointer to [**NullableNfssRules**](NfssRules.md) |  | [optional] 
**Git** | Pointer to [**NullableGitsRules**](GitsRules.md) |  | [optional] 
**ConfigMapVolume** | Pointer to [**NullableConfigMapsRules**](ConfigMapsRules.md) |  | [optional] 
**SecretVolume** | Pointer to [**NullableSecretsRules**](SecretsRules.md) |  | [optional] 
**EmptyDirVolume** | Pointer to [**NullableEmptyDirsRules**](EmptyDirsRules.md) |  | [optional] 
**S3** | Pointer to [**NullableS3sRules**](S3sRules.md) |  | [optional] 

## Methods

### NewAssetsRulesDatasources

`func NewAssetsRulesDatasources() *AssetsRulesDatasources`

NewAssetsRulesDatasources instantiates a new AssetsRulesDatasources object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewAssetsRulesDatasourcesWithDefaults

`func NewAssetsRulesDatasourcesWithDefaults() *AssetsRulesDatasources`

NewAssetsRulesDatasourcesWithDefaults instantiates a new AssetsRulesDatasources object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataVolume

`func (o *AssetsRulesDatasources) GetDataVolume() DataVolumesRules`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *AssetsRulesDatasources) GetDataVolumeOk() (*DataVolumesRules, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *AssetsRulesDatasources) SetDataVolume(v DataVolumesRules)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *AssetsRulesDatasources) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *AssetsRulesDatasources) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *AssetsRulesDatasources) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetPvc

`func (o *AssetsRulesDatasources) GetPvc() PvcsRules`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *AssetsRulesDatasources) GetPvcOk() (*PvcsRules, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *AssetsRulesDatasources) SetPvc(v PvcsRules)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *AssetsRulesDatasources) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *AssetsRulesDatasources) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *AssetsRulesDatasources) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetHostPath

`func (o *AssetsRulesDatasources) GetHostPath() HostPathsRules`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *AssetsRulesDatasources) GetHostPathOk() (*HostPathsRules, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *AssetsRulesDatasources) SetHostPath(v HostPathsRules)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *AssetsRulesDatasources) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *AssetsRulesDatasources) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *AssetsRulesDatasources) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *AssetsRulesDatasources) GetNfs() NfssRules`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *AssetsRulesDatasources) GetNfsOk() (*NfssRules, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *AssetsRulesDatasources) SetNfs(v NfssRules)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *AssetsRulesDatasources) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *AssetsRulesDatasources) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *AssetsRulesDatasources) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetGit

`func (o *AssetsRulesDatasources) GetGit() GitsRules`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *AssetsRulesDatasources) GetGitOk() (*GitsRules, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *AssetsRulesDatasources) SetGit(v GitsRules)`

SetGit sets Git field to given value.

### HasGit

`func (o *AssetsRulesDatasources) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *AssetsRulesDatasources) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *AssetsRulesDatasources) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetConfigMapVolume

`func (o *AssetsRulesDatasources) GetConfigMapVolume() ConfigMapsRules`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *AssetsRulesDatasources) GetConfigMapVolumeOk() (*ConfigMapsRules, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *AssetsRulesDatasources) SetConfigMapVolume(v ConfigMapsRules)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *AssetsRulesDatasources) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *AssetsRulesDatasources) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *AssetsRulesDatasources) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetSecretVolume

`func (o *AssetsRulesDatasources) GetSecretVolume() SecretsRules`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *AssetsRulesDatasources) GetSecretVolumeOk() (*SecretsRules, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *AssetsRulesDatasources) SetSecretVolume(v SecretsRules)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *AssetsRulesDatasources) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *AssetsRulesDatasources) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *AssetsRulesDatasources) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *AssetsRulesDatasources) GetEmptyDirVolume() EmptyDirsRules`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *AssetsRulesDatasources) GetEmptyDirVolumeOk() (*EmptyDirsRules, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *AssetsRulesDatasources) SetEmptyDirVolume(v EmptyDirsRules)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *AssetsRulesDatasources) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *AssetsRulesDatasources) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *AssetsRulesDatasources) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetS3

`func (o *AssetsRulesDatasources) GetS3() S3sRules`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *AssetsRulesDatasources) GetS3Ok() (*S3sRules, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *AssetsRulesDatasources) SetS3(v S3sRules)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *AssetsRulesDatasources) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *AssetsRulesDatasources) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *AssetsRulesDatasources) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


