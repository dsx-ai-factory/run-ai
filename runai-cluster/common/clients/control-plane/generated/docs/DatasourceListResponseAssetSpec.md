# DatasourceListResponseAssetSpec

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**HostPath** | Pointer to [**NullableDatasourceListResponseAssetSpecHostPath**](DatasourceListResponseAssetSpecHostPath.md) |  | [optional] 
**Nfs** | Pointer to [**NullableDatasourceListResponseAssetSpecNfs**](DatasourceListResponseAssetSpecNfs.md) |  | [optional] 
**Pvc** | Pointer to [**NullableDatasourceListResponseAssetSpecPvc**](DatasourceListResponseAssetSpecPvc.md) |  | [optional] 
**Git** | Pointer to [**NullableDatasourceListResponseAssetSpecGit**](DatasourceListResponseAssetSpecGit.md) |  | [optional] 
**S3** | Pointer to [**NullableDatasourceListResponseAssetSpecS3**](DatasourceListResponseAssetSpecS3.md) |  | [optional] 
**ConfigMap** | Pointer to [**NullableDatasourceListResponseAssetSpecConfigMap**](DatasourceListResponseAssetSpecConfigMap.md) |  | [optional] 
**Secret** | Pointer to [**NullableDatasourceListResponseAssetSpecSecret**](DatasourceListResponseAssetSpecSecret.md) |  | [optional] 

## Methods

### NewDatasourceListResponseAssetSpec

`func NewDatasourceListResponseAssetSpec() *DatasourceListResponseAssetSpec`

NewDatasourceListResponseAssetSpec instantiates a new DatasourceListResponseAssetSpec object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDatasourceListResponseAssetSpecWithDefaults

`func NewDatasourceListResponseAssetSpecWithDefaults() *DatasourceListResponseAssetSpec`

NewDatasourceListResponseAssetSpecWithDefaults instantiates a new DatasourceListResponseAssetSpec object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetHostPath

`func (o *DatasourceListResponseAssetSpec) GetHostPath() DatasourceListResponseAssetSpecHostPath`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *DatasourceListResponseAssetSpec) GetHostPathOk() (*DatasourceListResponseAssetSpecHostPath, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *DatasourceListResponseAssetSpec) SetHostPath(v DatasourceListResponseAssetSpecHostPath)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *DatasourceListResponseAssetSpec) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *DatasourceListResponseAssetSpec) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *DatasourceListResponseAssetSpec) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetNfs

`func (o *DatasourceListResponseAssetSpec) GetNfs() DatasourceListResponseAssetSpecNfs`

GetNfs returns the Nfs field if non-nil, zero value otherwise.

### GetNfsOk

`func (o *DatasourceListResponseAssetSpec) GetNfsOk() (*DatasourceListResponseAssetSpecNfs, bool)`

GetNfsOk returns a tuple with the Nfs field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNfs

`func (o *DatasourceListResponseAssetSpec) SetNfs(v DatasourceListResponseAssetSpecNfs)`

SetNfs sets Nfs field to given value.

### HasNfs

`func (o *DatasourceListResponseAssetSpec) HasNfs() bool`

HasNfs returns a boolean if a field has been set.

### SetNfsNil

`func (o *DatasourceListResponseAssetSpec) SetNfsNil(b bool)`

 SetNfsNil sets the value for Nfs to be an explicit nil

### UnsetNfs
`func (o *DatasourceListResponseAssetSpec) UnsetNfs()`

UnsetNfs ensures that no value is present for Nfs, not even an explicit nil
### GetPvc

`func (o *DatasourceListResponseAssetSpec) GetPvc() DatasourceListResponseAssetSpecPvc`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *DatasourceListResponseAssetSpec) GetPvcOk() (*DatasourceListResponseAssetSpecPvc, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *DatasourceListResponseAssetSpec) SetPvc(v DatasourceListResponseAssetSpecPvc)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *DatasourceListResponseAssetSpec) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *DatasourceListResponseAssetSpec) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *DatasourceListResponseAssetSpec) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetGit

`func (o *DatasourceListResponseAssetSpec) GetGit() DatasourceListResponseAssetSpecGit`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *DatasourceListResponseAssetSpec) GetGitOk() (*DatasourceListResponseAssetSpecGit, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *DatasourceListResponseAssetSpec) SetGit(v DatasourceListResponseAssetSpecGit)`

SetGit sets Git field to given value.

### HasGit

`func (o *DatasourceListResponseAssetSpec) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *DatasourceListResponseAssetSpec) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *DatasourceListResponseAssetSpec) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetS3

`func (o *DatasourceListResponseAssetSpec) GetS3() DatasourceListResponseAssetSpecS3`

GetS3 returns the S3 field if non-nil, zero value otherwise.

### GetS3Ok

`func (o *DatasourceListResponseAssetSpec) GetS3Ok() (*DatasourceListResponseAssetSpecS3, bool)`

GetS3Ok returns a tuple with the S3 field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetS3

`func (o *DatasourceListResponseAssetSpec) SetS3(v DatasourceListResponseAssetSpecS3)`

SetS3 sets S3 field to given value.

### HasS3

`func (o *DatasourceListResponseAssetSpec) HasS3() bool`

HasS3 returns a boolean if a field has been set.

### SetS3Nil

`func (o *DatasourceListResponseAssetSpec) SetS3Nil(b bool)`

 SetS3Nil sets the value for S3 to be an explicit nil

### UnsetS3
`func (o *DatasourceListResponseAssetSpec) UnsetS3()`

UnsetS3 ensures that no value is present for S3, not even an explicit nil
### GetConfigMap

`func (o *DatasourceListResponseAssetSpec) GetConfigMap() DatasourceListResponseAssetSpecConfigMap`

GetConfigMap returns the ConfigMap field if non-nil, zero value otherwise.

### GetConfigMapOk

`func (o *DatasourceListResponseAssetSpec) GetConfigMapOk() (*DatasourceListResponseAssetSpecConfigMap, bool)`

GetConfigMapOk returns a tuple with the ConfigMap field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMap

`func (o *DatasourceListResponseAssetSpec) SetConfigMap(v DatasourceListResponseAssetSpecConfigMap)`

SetConfigMap sets ConfigMap field to given value.

### HasConfigMap

`func (o *DatasourceListResponseAssetSpec) HasConfigMap() bool`

HasConfigMap returns a boolean if a field has been set.

### SetConfigMapNil

`func (o *DatasourceListResponseAssetSpec) SetConfigMapNil(b bool)`

 SetConfigMapNil sets the value for ConfigMap to be an explicit nil

### UnsetConfigMap
`func (o *DatasourceListResponseAssetSpec) UnsetConfigMap()`

UnsetConfigMap ensures that no value is present for ConfigMap, not even an explicit nil
### GetSecret

`func (o *DatasourceListResponseAssetSpec) GetSecret() DatasourceListResponseAssetSpecSecret`

GetSecret returns the Secret field if non-nil, zero value otherwise.

### GetSecretOk

`func (o *DatasourceListResponseAssetSpec) GetSecretOk() (*DatasourceListResponseAssetSpecSecret, bool)`

GetSecretOk returns a tuple with the Secret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecret

`func (o *DatasourceListResponseAssetSpec) SetSecret(v DatasourceListResponseAssetSpecSecret)`

SetSecret sets Secret field to given value.

### HasSecret

`func (o *DatasourceListResponseAssetSpec) HasSecret() bool`

HasSecret returns a boolean if a field has been set.

### SetSecretNil

`func (o *DatasourceListResponseAssetSpec) SetSecretNil(b bool)`

 SetSecretNil sets the value for Secret to be an explicit nil

### UnsetSecret
`func (o *DatasourceListResponseAssetSpec) UnsetSecret()`

UnsetSecret ensures that no value is present for Secret, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


