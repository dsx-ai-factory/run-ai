# InferencePolicyDefaultsV2AllOfStorage

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
**SecretVolume** | Pointer to [**NullableSecretsDefaults**](SecretsDefaults.md) |  | [optional] 

## Methods

### NewInferencePolicyDefaultsV2AllOfStorage

`func NewInferencePolicyDefaultsV2AllOfStorage() *InferencePolicyDefaultsV2AllOfStorage`

NewInferencePolicyDefaultsV2AllOfStorage instantiates a new InferencePolicyDefaultsV2AllOfStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencePolicyDefaultsV2AllOfStorageWithDefaults

`func NewInferencePolicyDefaultsV2AllOfStorageWithDefaults() *InferencePolicyDefaultsV2AllOfStorage`

NewInferencePolicyDefaultsV2AllOfStorageWithDefaults instantiates a new InferencePolicyDefaultsV2AllOfStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetConfigMapVolume() ConfigMapsDefaults`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetConfigMapVolumeOk() (*ConfigMapsDefaults, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetConfigMapVolume(v ConfigMapsDefaults)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *InferencePolicyDefaultsV2AllOfStorage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetDataVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetDataVolume() DataVolumesDefaults`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetDataVolumeOk() (*DataVolumesDefaults, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetDataVolume(v DataVolumesDefaults)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *InferencePolicyDefaultsV2AllOfStorage) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetEmptyDirVolume() EmptyDirsDefaults`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetEmptyDirVolumeOk() (*EmptyDirsDefaults, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetEmptyDirVolume(v EmptyDirsDefaults)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *InferencePolicyDefaultsV2AllOfStorage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetGit

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetGit() GitsDefaults`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetGitOk() (*GitsDefaults, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetGit(v GitsDefaults)`

SetGit sets Git field to given value.

### HasGit

`func (o *InferencePolicyDefaultsV2AllOfStorage) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *InferencePolicyDefaultsV2AllOfStorage) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetHostPath

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetHostPath() HostPathsDefaults`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetHostPathOk() (*HostPathsDefaults, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetHostPath(v HostPathsDefaults)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *InferencePolicyDefaultsV2AllOfStorage) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *InferencePolicyDefaultsV2AllOfStorage) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetNfs() NfssDefaults`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetNfsOk() (*NfssDefaults, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetNfs(v NfssDefaults)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *InferencePolicyDefaultsV2AllOfStorage) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *InferencePolicyDefaultsV2AllOfStorage) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetPvc

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetPvc() PvcsDefaults`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetPvcOk() (*PvcsDefaults, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetPvc(v PvcsDefaults)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *InferencePolicyDefaultsV2AllOfStorage) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *InferencePolicyDefaultsV2AllOfStorage) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetSecretVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetSecretVolume() SecretsDefaults`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *InferencePolicyDefaultsV2AllOfStorage) GetSecretVolumeOk() (*SecretsDefaults, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetSecretVolume(v SecretsDefaults)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *InferencePolicyDefaultsV2AllOfStorage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *InferencePolicyDefaultsV2AllOfStorage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *InferencePolicyDefaultsV2AllOfStorage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


