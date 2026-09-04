# CommonStorageFieldsDefaults

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataVolume** | Pointer to [**NullableDataVolumesDefaults**](DataVolumesDefaults.md) |  | [optional] 
**Pvc** | Pointer to [**NullablePvcsDefaults**](PvcsDefaults.md) |  | [optional] 
**Git** | Pointer to [**NullableGitsDefaults**](GitsDefaults.md) |  | [optional] 
**ConfigMapVolume** | Pointer to [**NullableConfigMapsDefaults**](ConfigMapsDefaults.md) |  | [optional] 
**SecretVolume** | Pointer to [**NullableSecretsDefaults**](SecretsDefaults.md) |  | [optional] 

## Methods

### NewCommonStorageFieldsDefaults

`func NewCommonStorageFieldsDefaults() *CommonStorageFieldsDefaults`

NewCommonStorageFieldsDefaults instantiates a new CommonStorageFieldsDefaults object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommonStorageFieldsDefaultsWithDefaults

`func NewCommonStorageFieldsDefaultsWithDefaults() *CommonStorageFieldsDefaults`

NewCommonStorageFieldsDefaultsWithDefaults instantiates a new CommonStorageFieldsDefaults object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataVolume

`func (o *CommonStorageFieldsDefaults) GetDataVolume() DataVolumesDefaults`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *CommonStorageFieldsDefaults) GetDataVolumeOk() (*DataVolumesDefaults, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *CommonStorageFieldsDefaults) SetDataVolume(v DataVolumesDefaults)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *CommonStorageFieldsDefaults) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *CommonStorageFieldsDefaults) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *CommonStorageFieldsDefaults) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetPvc

`func (o *CommonStorageFieldsDefaults) GetPvc() PvcsDefaults`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *CommonStorageFieldsDefaults) GetPvcOk() (*PvcsDefaults, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *CommonStorageFieldsDefaults) SetPvc(v PvcsDefaults)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *CommonStorageFieldsDefaults) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *CommonStorageFieldsDefaults) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *CommonStorageFieldsDefaults) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetGit

`func (o *CommonStorageFieldsDefaults) GetGit() GitsDefaults`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *CommonStorageFieldsDefaults) GetGitOk() (*GitsDefaults, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *CommonStorageFieldsDefaults) SetGit(v GitsDefaults)`

SetGit sets Git field to given value.

### HasGit

`func (o *CommonStorageFieldsDefaults) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *CommonStorageFieldsDefaults) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *CommonStorageFieldsDefaults) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetConfigMapVolume

`func (o *CommonStorageFieldsDefaults) GetConfigMapVolume() ConfigMapsDefaults`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *CommonStorageFieldsDefaults) GetConfigMapVolumeOk() (*ConfigMapsDefaults, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *CommonStorageFieldsDefaults) SetConfigMapVolume(v ConfigMapsDefaults)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *CommonStorageFieldsDefaults) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *CommonStorageFieldsDefaults) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *CommonStorageFieldsDefaults) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetSecretVolume

`func (o *CommonStorageFieldsDefaults) GetSecretVolume() SecretsDefaults`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *CommonStorageFieldsDefaults) GetSecretVolumeOk() (*SecretsDefaults, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *CommonStorageFieldsDefaults) SetSecretVolume(v SecretsDefaults)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *CommonStorageFieldsDefaults) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *CommonStorageFieldsDefaults) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *CommonStorageFieldsDefaults) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


