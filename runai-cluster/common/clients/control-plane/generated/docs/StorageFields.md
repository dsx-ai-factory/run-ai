# StorageFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataVolume** | Pointer to [**[]DataVolumeInstance**](DataVolumeInstance.md) | Set of data volumes to use in the workload. | [optional] 
**Pvc** | Pointer to [**[]PvcInstance**](PvcInstance.md) | Set of pvc persistent volume claims to use in the workload. | [optional] 
**Git** | Pointer to [**[]GitInstance**](GitInstance.md) | Set of git repositories to use in the workload. | [optional] 
**ConfigMapVolume** | Pointer to [**[]ConfigMapInstance**](ConfigMapInstance.md) | Set of config map volumes to use in the workload | [optional] 
**SecretVolume** | Pointer to [**[]SecretInstance1**](SecretInstance1.md) | Set of secret volumes to use in the workload | [optional] 
**HostPath** | Pointer to [**[]HostPathInstance**](HostPathInstance.md) | Set of host paths to use in the workload. | [optional] 
**Nfs** | Pointer to [**[]NfsInstance**](NfsInstance.md) | Set of nfs volumes to use in the workload. | [optional] 
**S3** | Pointer to [**[]S3Instance**](S3Instance.md) | Set of s3 buckets to use in the workload. | [optional] 

## Methods

### NewStorageFields

`func NewStorageFields() *StorageFields`

NewStorageFields instantiates a new StorageFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewStorageFieldsWithDefaults

`func NewStorageFieldsWithDefaults() *StorageFields`

NewStorageFieldsWithDefaults instantiates a new StorageFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataVolume

`func (o *StorageFields) GetDataVolume() []DataVolumeInstance`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *StorageFields) GetDataVolumeOk() (*[]DataVolumeInstance, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *StorageFields) SetDataVolume(v []DataVolumeInstance)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *StorageFields) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *StorageFields) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *StorageFields) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetPvc

`func (o *StorageFields) GetPvc() []PvcInstance`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *StorageFields) GetPvcOk() (*[]PvcInstance, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *StorageFields) SetPvc(v []PvcInstance)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *StorageFields) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *StorageFields) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *StorageFields) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetGit

`func (o *StorageFields) GetGit() []GitInstance`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *StorageFields) GetGitOk() (*[]GitInstance, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *StorageFields) SetGit(v []GitInstance)`

SetGit sets Git field to given value.

### HasGit

`func (o *StorageFields) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *StorageFields) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *StorageFields) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetConfigMapVolume

`func (o *StorageFields) GetConfigMapVolume() []ConfigMapInstance`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *StorageFields) GetConfigMapVolumeOk() (*[]ConfigMapInstance, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *StorageFields) SetConfigMapVolume(v []ConfigMapInstance)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *StorageFields) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *StorageFields) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *StorageFields) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetSecretVolume

`func (o *StorageFields) GetSecretVolume() []SecretInstance1`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *StorageFields) GetSecretVolumeOk() (*[]SecretInstance1, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *StorageFields) SetSecretVolume(v []SecretInstance1)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *StorageFields) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *StorageFields) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *StorageFields) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil
### GetHostPath

`func (o *StorageFields) GetHostPath() []HostPathInstance`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *StorageFields) GetHostPathOk() (*[]HostPathInstance, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *StorageFields) SetHostPath(v []HostPathInstance)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *StorageFields) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *StorageFields) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *StorageFields) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *StorageFields) GetNfs() []NfsInstance`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *StorageFields) GetNfsOk() (*[]NfsInstance, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *StorageFields) SetNfs(v []NfsInstance)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *StorageFields) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *StorageFields) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *StorageFields) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetS3

`func (o *StorageFields) GetS3() []S3Instance`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *StorageFields) GetS3Ok() (*[]S3Instance, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *StorageFields) SetS3(v []S3Instance)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *StorageFields) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *StorageFields) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *StorageFields) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


