# WorkspaceSpecSpecAllOfStorage

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

### NewWorkspaceSpecSpecAllOfStorage

`func NewWorkspaceSpecSpecAllOfStorage() *WorkspaceSpecSpecAllOfStorage`

NewWorkspaceSpecSpecAllOfStorage instantiates a new WorkspaceSpecSpecAllOfStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspaceSpecSpecAllOfStorageWithDefaults

`func NewWorkspaceSpecSpecAllOfStorageWithDefaults() *WorkspaceSpecSpecAllOfStorage`

NewWorkspaceSpecSpecAllOfStorageWithDefaults instantiates a new WorkspaceSpecSpecAllOfStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *WorkspaceSpecSpecAllOfStorage) GetConfigMapVolume() []ConfigMapInstance`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *WorkspaceSpecSpecAllOfStorage) GetConfigMapVolumeOk() (*[]ConfigMapInstance, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *WorkspaceSpecSpecAllOfStorage) SetConfigMapVolume(v []ConfigMapInstance)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *WorkspaceSpecSpecAllOfStorage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *WorkspaceSpecSpecAllOfStorage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *WorkspaceSpecSpecAllOfStorage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetDataVolume

`func (o *WorkspaceSpecSpecAllOfStorage) GetDataVolume() []DataVolumeInstance`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *WorkspaceSpecSpecAllOfStorage) GetDataVolumeOk() (*[]DataVolumeInstance, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *WorkspaceSpecSpecAllOfStorage) SetDataVolume(v []DataVolumeInstance)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *WorkspaceSpecSpecAllOfStorage) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *WorkspaceSpecSpecAllOfStorage) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *WorkspaceSpecSpecAllOfStorage) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *WorkspaceSpecSpecAllOfStorage) GetEmptyDirVolume() []EmptyDirInstance`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *WorkspaceSpecSpecAllOfStorage) GetEmptyDirVolumeOk() (*[]EmptyDirInstance, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *WorkspaceSpecSpecAllOfStorage) SetEmptyDirVolume(v []EmptyDirInstance)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *WorkspaceSpecSpecAllOfStorage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *WorkspaceSpecSpecAllOfStorage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *WorkspaceSpecSpecAllOfStorage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetGit

`func (o *WorkspaceSpecSpecAllOfStorage) GetGit() []GitInstance`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *WorkspaceSpecSpecAllOfStorage) GetGitOk() (*[]GitInstance, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *WorkspaceSpecSpecAllOfStorage) SetGit(v []GitInstance)`

SetGit sets Git field to given value.

### HasGit

`func (o *WorkspaceSpecSpecAllOfStorage) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *WorkspaceSpecSpecAllOfStorage) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *WorkspaceSpecSpecAllOfStorage) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetHostPath

`func (o *WorkspaceSpecSpecAllOfStorage) GetHostPath() []HostPathInstance`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *WorkspaceSpecSpecAllOfStorage) GetHostPathOk() (*[]HostPathInstance, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *WorkspaceSpecSpecAllOfStorage) SetHostPath(v []HostPathInstance)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *WorkspaceSpecSpecAllOfStorage) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *WorkspaceSpecSpecAllOfStorage) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *WorkspaceSpecSpecAllOfStorage) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *WorkspaceSpecSpecAllOfStorage) GetNfs() []NfsInstance`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *WorkspaceSpecSpecAllOfStorage) GetNfsOk() (*[]NfsInstance, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *WorkspaceSpecSpecAllOfStorage) SetNfs(v []NfsInstance)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *WorkspaceSpecSpecAllOfStorage) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *WorkspaceSpecSpecAllOfStorage) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *WorkspaceSpecSpecAllOfStorage) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetPvc

`func (o *WorkspaceSpecSpecAllOfStorage) GetPvc() []PvcInstance`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *WorkspaceSpecSpecAllOfStorage) GetPvcOk() (*[]PvcInstance, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *WorkspaceSpecSpecAllOfStorage) SetPvc(v []PvcInstance)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *WorkspaceSpecSpecAllOfStorage) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *WorkspaceSpecSpecAllOfStorage) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *WorkspaceSpecSpecAllOfStorage) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetS3

`func (o *WorkspaceSpecSpecAllOfStorage) GetS3() []S3Instance`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *WorkspaceSpecSpecAllOfStorage) GetS3Ok() (*[]S3Instance, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *WorkspaceSpecSpecAllOfStorage) SetS3(v []S3Instance)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *WorkspaceSpecSpecAllOfStorage) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *WorkspaceSpecSpecAllOfStorage) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *WorkspaceSpecSpecAllOfStorage) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil
### GetSecretVolume

`func (o *WorkspaceSpecSpecAllOfStorage) GetSecretVolume() []SecretInstance`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *WorkspaceSpecSpecAllOfStorage) GetSecretVolumeOk() (*[]SecretInstance, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *WorkspaceSpecSpecAllOfStorage) SetSecretVolume(v []SecretInstance)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *WorkspaceSpecSpecAllOfStorage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *WorkspaceSpecSpecAllOfStorage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *WorkspaceSpecSpecAllOfStorage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


