# CommonStorageFields

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataVolume** | Pointer to [**[]DataVolumeInstance**](DataVolumeInstance.md) | Set of data volumes to use in the workload. | [optional] 
**Pvc** | Pointer to [**[]PvcInstance**](PvcInstance.md) | Set of pvc persistent volume claims to use in the workload. | [optional] 
**Git** | Pointer to [**[]GitInstance**](GitInstance.md) | Set of git repositories to use in the workload. | [optional] 
**ConfigMapVolume** | Pointer to [**[]ConfigMapInstance**](ConfigMapInstance.md) | Set of config map volumes to use in the workload | [optional] 
**SecretVolume** | Pointer to [**[]SecretInstance1**](SecretInstance1.md) | Set of secret volumes to use in the workload | [optional] 

## Methods

### NewCommonStorageFields

`func NewCommonStorageFields() *CommonStorageFields`

NewCommonStorageFields instantiates a new CommonStorageFields object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommonStorageFieldsWithDefaults

`func NewCommonStorageFieldsWithDefaults() *CommonStorageFields`

NewCommonStorageFieldsWithDefaults instantiates a new CommonStorageFields object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataVolume

`func (o *CommonStorageFields) GetDataVolume() []DataVolumeInstance`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *CommonStorageFields) GetDataVolumeOk() (*[]DataVolumeInstance, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *CommonStorageFields) SetDataVolume(v []DataVolumeInstance)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *CommonStorageFields) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *CommonStorageFields) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *CommonStorageFields) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetPvc

`func (o *CommonStorageFields) GetPvc() []PvcInstance`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *CommonStorageFields) GetPvcOk() (*[]PvcInstance, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *CommonStorageFields) SetPvc(v []PvcInstance)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *CommonStorageFields) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *CommonStorageFields) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *CommonStorageFields) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetGit

`func (o *CommonStorageFields) GetGit() []GitInstance`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *CommonStorageFields) GetGitOk() (*[]GitInstance, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *CommonStorageFields) SetGit(v []GitInstance)`

SetGit sets Git field to given value.

### HasGit

`func (o *CommonStorageFields) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *CommonStorageFields) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *CommonStorageFields) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetConfigMapVolume

`func (o *CommonStorageFields) GetConfigMapVolume() []ConfigMapInstance`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *CommonStorageFields) GetConfigMapVolumeOk() (*[]ConfigMapInstance, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *CommonStorageFields) SetConfigMapVolume(v []ConfigMapInstance)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *CommonStorageFields) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *CommonStorageFields) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *CommonStorageFields) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetSecretVolume

`func (o *CommonStorageFields) GetSecretVolume() []SecretInstance1`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *CommonStorageFields) GetSecretVolumeOk() (*[]SecretInstance1, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *CommonStorageFields) SetSecretVolume(v []SecretInstance1)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *CommonStorageFields) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *CommonStorageFields) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *CommonStorageFields) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


