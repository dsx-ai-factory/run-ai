# InferencePolicyRulesV2AllOfStorage

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
**SecretVolume** | Pointer to [**NullableSecretsRules**](SecretsRules.md) |  | [optional] 

## Methods

### NewInferencePolicyRulesV2AllOfStorage

`func NewInferencePolicyRulesV2AllOfStorage() *InferencePolicyRulesV2AllOfStorage`

NewInferencePolicyRulesV2AllOfStorage instantiates a new InferencePolicyRulesV2AllOfStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencePolicyRulesV2AllOfStorageWithDefaults

`func NewInferencePolicyRulesV2AllOfStorageWithDefaults() *InferencePolicyRulesV2AllOfStorage`

NewInferencePolicyRulesV2AllOfStorageWithDefaults instantiates a new InferencePolicyRulesV2AllOfStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *InferencePolicyRulesV2AllOfStorage) GetConfigMapVolume() ConfigMapsRules`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *InferencePolicyRulesV2AllOfStorage) GetConfigMapVolumeOk() (*ConfigMapsRules, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *InferencePolicyRulesV2AllOfStorage) SetConfigMapVolume(v ConfigMapsRules)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *InferencePolicyRulesV2AllOfStorage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *InferencePolicyRulesV2AllOfStorage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *InferencePolicyRulesV2AllOfStorage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetDataVolume

`func (o *InferencePolicyRulesV2AllOfStorage) GetDataVolume() DataVolumesRules`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *InferencePolicyRulesV2AllOfStorage) GetDataVolumeOk() (*DataVolumesRules, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *InferencePolicyRulesV2AllOfStorage) SetDataVolume(v DataVolumesRules)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *InferencePolicyRulesV2AllOfStorage) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *InferencePolicyRulesV2AllOfStorage) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *InferencePolicyRulesV2AllOfStorage) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *InferencePolicyRulesV2AllOfStorage) GetEmptyDirVolume() EmptyDirsRules`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *InferencePolicyRulesV2AllOfStorage) GetEmptyDirVolumeOk() (*EmptyDirsRules, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *InferencePolicyRulesV2AllOfStorage) SetEmptyDirVolume(v EmptyDirsRules)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *InferencePolicyRulesV2AllOfStorage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *InferencePolicyRulesV2AllOfStorage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *InferencePolicyRulesV2AllOfStorage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetGit

`func (o *InferencePolicyRulesV2AllOfStorage) GetGit() GitsRules`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *InferencePolicyRulesV2AllOfStorage) GetGitOk() (*GitsRules, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *InferencePolicyRulesV2AllOfStorage) SetGit(v GitsRules)`

SetGit sets Git field to given value.

### HasGit

`func (o *InferencePolicyRulesV2AllOfStorage) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *InferencePolicyRulesV2AllOfStorage) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *InferencePolicyRulesV2AllOfStorage) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetHostPath

`func (o *InferencePolicyRulesV2AllOfStorage) GetHostPath() HostPathsRules`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *InferencePolicyRulesV2AllOfStorage) GetHostPathOk() (*HostPathsRules, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *InferencePolicyRulesV2AllOfStorage) SetHostPath(v HostPathsRules)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *InferencePolicyRulesV2AllOfStorage) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *InferencePolicyRulesV2AllOfStorage) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *InferencePolicyRulesV2AllOfStorage) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *InferencePolicyRulesV2AllOfStorage) GetNfs() NfssRules`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *InferencePolicyRulesV2AllOfStorage) GetNfsOk() (*NfssRules, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *InferencePolicyRulesV2AllOfStorage) SetNfs(v NfssRules)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *InferencePolicyRulesV2AllOfStorage) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *InferencePolicyRulesV2AllOfStorage) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *InferencePolicyRulesV2AllOfStorage) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetPvc

`func (o *InferencePolicyRulesV2AllOfStorage) GetPvc() PvcsRules`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *InferencePolicyRulesV2AllOfStorage) GetPvcOk() (*PvcsRules, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *InferencePolicyRulesV2AllOfStorage) SetPvc(v PvcsRules)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *InferencePolicyRulesV2AllOfStorage) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *InferencePolicyRulesV2AllOfStorage) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *InferencePolicyRulesV2AllOfStorage) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetSecretVolume

`func (o *InferencePolicyRulesV2AllOfStorage) GetSecretVolume() SecretsRules`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *InferencePolicyRulesV2AllOfStorage) GetSecretVolumeOk() (*SecretsRules, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *InferencePolicyRulesV2AllOfStorage) SetSecretVolume(v SecretsRules)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *InferencePolicyRulesV2AllOfStorage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *InferencePolicyRulesV2AllOfStorage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *InferencePolicyRulesV2AllOfStorage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


