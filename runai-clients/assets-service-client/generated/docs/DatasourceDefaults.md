# DatasourceDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**HostPath** | Pointer to [**NullableHostPath**](HostPath.md) |  | [optional] 
**Nfs** | Pointer to [**NullableNfs**](Nfs.md) |  | [optional] 
**Pvc** | Pointer to [**Pvc**](Pvc.md) |  | [optional] 
**Git** | Pointer to [**NullableGitCommon**](GitCommon.md) |  | [optional] 
**S3** | Pointer to [**NullableS3Common**](S3Common.md) |  | [optional] 
**DataVolume** | Pointer to [**NullableDataVolume**](DataVolume.md) |  | [optional] 
**ConfigMapVolume** | Pointer to [**NullableConfigMap**](ConfigMap.md) |  | [optional] 
**SecretVolume** | Pointer to [**Secret**](Secret.md) |  | [optional] 

## Methods

### NewDatasourceDefaults

`func NewDatasourceDefaults() *DatasourceDefaults`

NewDatasourceDefaults instantiates a new DatasourceDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceDefaultsWithDefaults

`func NewDatasourceDefaultsWithDefaults() *DatasourceDefaults`

NewDatasourceDefaultsWithDefaults instantiates a new DatasourceDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHostPath

`func (o *DatasourceDefaults) GetHostPath() HostPath`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *DatasourceDefaults) GetHostPathOk() (*HostPath, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *DatasourceDefaults) SetHostPath(v HostPath)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *DatasourceDefaults) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *DatasourceDefaults) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *DatasourceDefaults) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *DatasourceDefaults) GetNfs() Nfs`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *DatasourceDefaults) GetNfsOk() (*Nfs, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *DatasourceDefaults) SetNfs(v Nfs)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *DatasourceDefaults) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *DatasourceDefaults) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *DatasourceDefaults) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetPvc

`func (o *DatasourceDefaults) GetPvc() Pvc`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *DatasourceDefaults) GetPvcOk() (*Pvc, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *DatasourceDefaults) SetPvc(v Pvc)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *DatasourceDefaults) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### GetGit

`func (o *DatasourceDefaults) GetGit() GitCommon`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *DatasourceDefaults) GetGitOk() (*GitCommon, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *DatasourceDefaults) SetGit(v GitCommon)`

SetGit sets Git field to given value.

### HasGit

`func (o *DatasourceDefaults) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *DatasourceDefaults) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *DatasourceDefaults) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetS3

`func (o *DatasourceDefaults) GetS3() S3Common`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *DatasourceDefaults) GetS3Ok() (*S3Common, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *DatasourceDefaults) SetS3(v S3Common)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *DatasourceDefaults) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *DatasourceDefaults) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *DatasourceDefaults) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil
### GetDataVolume

`func (o *DatasourceDefaults) GetDataVolume() DataVolume`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *DatasourceDefaults) GetDataVolumeOk() (*DataVolume, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *DatasourceDefaults) SetDataVolume(v DataVolume)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *DatasourceDefaults) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *DatasourceDefaults) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *DatasourceDefaults) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetConfigMapVolume

`func (o *DatasourceDefaults) GetConfigMapVolume() ConfigMap`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *DatasourceDefaults) GetConfigMapVolumeOk() (*ConfigMap, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *DatasourceDefaults) SetConfigMapVolume(v ConfigMap)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *DatasourceDefaults) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *DatasourceDefaults) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *DatasourceDefaults) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetSecretVolume

`func (o *DatasourceDefaults) GetSecretVolume() Secret`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *DatasourceDefaults) GetSecretVolumeOk() (*Secret, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *DatasourceDefaults) SetSecretVolume(v Secret)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *DatasourceDefaults) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


