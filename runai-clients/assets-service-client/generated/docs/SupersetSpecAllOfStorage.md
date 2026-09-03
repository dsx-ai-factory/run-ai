# SupersetSpecAllOfStorage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMapVolume** | Pointer to [**[]ConfigMapInstance**](ConfigMapInstance.md) | Set of config map volumes to use in the workload | [optional] 
**DataVolume** | Pointer to [**[]DataVolumeInstance**](DataVolumeInstance.md) | Set of data volumes to use in the workload. | [optional] 
**EmptyDirVolume** | Pointer to [**[]EmptyDirInstance**](EmptyDirInstance.md) | A list of emptyDir volumes to mount in the workload. | [optional] 
**Git** | Pointer to [**[]GitInstance**](GitInstance.md) | Set of git repositories to use in the workload. | [optional] 
**HostPath** | Pointer to [**[]HostPathInstance**](HostPathInstance.md) | Set of host paths to use in the workload. | [optional] 
**Nfs** | Pointer to [**[]NfsInstance**](NfsInstance.md) | Set of nfs volumes to use in the workload. | [optional] 
**Pvc** | Pointer to [**[]PvcInstance**](PvcInstance.md) | Set of pvc persistent volume claims to use in the workload. | [optional] 
**S3** | Pointer to [**[]S3Instance**](S3Instance.md) | Set of s3 buckets to use in the workload. | [optional] 
**SecretVolume** | Pointer to [**[]SecretInstance**](SecretInstance.md) | Set of secret volumes to use in the workload | [optional] 

## Methods

### NewSupersetSpecAllOfStorage

`func NewSupersetSpecAllOfStorage() *SupersetSpecAllOfStorage`

NewSupersetSpecAllOfStorage instantiates a new SupersetSpecAllOfStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSupersetSpecAllOfStorageWithDefaults

`func NewSupersetSpecAllOfStorageWithDefaults() *SupersetSpecAllOfStorage`

NewSupersetSpecAllOfStorageWithDefaults instantiates a new SupersetSpecAllOfStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *SupersetSpecAllOfStorage) GetConfigMapVolume() []ConfigMapInstance`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *SupersetSpecAllOfStorage) GetConfigMapVolumeOk() (*[]ConfigMapInstance, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *SupersetSpecAllOfStorage) SetConfigMapVolume(v []ConfigMapInstance)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *SupersetSpecAllOfStorage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *SupersetSpecAllOfStorage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *SupersetSpecAllOfStorage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetDataVolume

`func (o *SupersetSpecAllOfStorage) GetDataVolume() []DataVolumeInstance`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *SupersetSpecAllOfStorage) GetDataVolumeOk() (*[]DataVolumeInstance, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *SupersetSpecAllOfStorage) SetDataVolume(v []DataVolumeInstance)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *SupersetSpecAllOfStorage) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *SupersetSpecAllOfStorage) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *SupersetSpecAllOfStorage) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *SupersetSpecAllOfStorage) GetEmptyDirVolume() []EmptyDirInstance`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *SupersetSpecAllOfStorage) GetEmptyDirVolumeOk() (*[]EmptyDirInstance, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *SupersetSpecAllOfStorage) SetEmptyDirVolume(v []EmptyDirInstance)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *SupersetSpecAllOfStorage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *SupersetSpecAllOfStorage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *SupersetSpecAllOfStorage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetGit

`func (o *SupersetSpecAllOfStorage) GetGit() []GitInstance`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *SupersetSpecAllOfStorage) GetGitOk() (*[]GitInstance, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *SupersetSpecAllOfStorage) SetGit(v []GitInstance)`

SetGit sets Git field to given value.

### HasGit

`func (o *SupersetSpecAllOfStorage) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *SupersetSpecAllOfStorage) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *SupersetSpecAllOfStorage) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetHostPath

`func (o *SupersetSpecAllOfStorage) GetHostPath() []HostPathInstance`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *SupersetSpecAllOfStorage) GetHostPathOk() (*[]HostPathInstance, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *SupersetSpecAllOfStorage) SetHostPath(v []HostPathInstance)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *SupersetSpecAllOfStorage) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *SupersetSpecAllOfStorage) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *SupersetSpecAllOfStorage) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *SupersetSpecAllOfStorage) GetNfs() []NfsInstance`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *SupersetSpecAllOfStorage) GetNfsOk() (*[]NfsInstance, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *SupersetSpecAllOfStorage) SetNfs(v []NfsInstance)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *SupersetSpecAllOfStorage) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *SupersetSpecAllOfStorage) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *SupersetSpecAllOfStorage) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetPvc

`func (o *SupersetSpecAllOfStorage) GetPvc() []PvcInstance`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *SupersetSpecAllOfStorage) GetPvcOk() (*[]PvcInstance, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *SupersetSpecAllOfStorage) SetPvc(v []PvcInstance)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *SupersetSpecAllOfStorage) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *SupersetSpecAllOfStorage) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *SupersetSpecAllOfStorage) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetS3

`func (o *SupersetSpecAllOfStorage) GetS3() []S3Instance`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *SupersetSpecAllOfStorage) GetS3Ok() (*[]S3Instance, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *SupersetSpecAllOfStorage) SetS3(v []S3Instance)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *SupersetSpecAllOfStorage) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *SupersetSpecAllOfStorage) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *SupersetSpecAllOfStorage) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil
### GetSecretVolume

`func (o *SupersetSpecAllOfStorage) GetSecretVolume() []SecretInstance`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *SupersetSpecAllOfStorage) GetSecretVolumeOk() (*[]SecretInstance, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *SupersetSpecAllOfStorage) SetSecretVolume(v []SecretInstance)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *SupersetSpecAllOfStorage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *SupersetSpecAllOfStorage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *SupersetSpecAllOfStorage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


