# CommonStorageFieldsRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**DataVolume** | Pointer to [**NullableDataVolumesRules**](DataVolumesRules.md) |  | [optional] 
**Pvc** | Pointer to [**NullablePvcsRules**](PvcsRules.md) |  | [optional] 
**Git** | Pointer to [**NullableGitsRules**](GitsRules.md) |  | [optional] 
**ConfigMapVolume** | Pointer to [**NullableConfigMapsRules**](ConfigMapsRules.md) |  | [optional] 
**SecretVolume** | Pointer to [**NullableSecretsRules**](SecretsRules.md) |  | [optional] 

## Methods

### NewCommonStorageFieldsRules

`func NewCommonStorageFieldsRules() *CommonStorageFieldsRules`

NewCommonStorageFieldsRules instantiates a new CommonStorageFieldsRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewCommonStorageFieldsRulesWithDefaults

`func NewCommonStorageFieldsRulesWithDefaults() *CommonStorageFieldsRules`

NewCommonStorageFieldsRulesWithDefaults instantiates a new CommonStorageFieldsRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDataVolume

`func (o *CommonStorageFieldsRules) GetDataVolume() DataVolumesRules`

GetDataVolume returns the DataVolume field if non-nil, zero value otherwise.

### GetDataVolumeOk

`func (o *CommonStorageFieldsRules) GetDataVolumeOk() (*DataVolumesRules, bool)`

GetDataVolumeOk returns a tuple with the DataVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDataVolume

`func (o *CommonStorageFieldsRules) SetDataVolume(v DataVolumesRules)`

SetDataVolume sets DataVolume field to given value.

### HasDataVolume

`func (o *CommonStorageFieldsRules) HasDataVolume() bool`

HasDataVolume returns a boolean if a field has been set.

### SetDataVolumeNil

`func (o *CommonStorageFieldsRules) SetDataVolumeNil(b bool)`

 SetDataVolumeNil sets the value for DataVolume to be an explicit nil

### UnsetDataVolume
`func (o *CommonStorageFieldsRules) UnsetDataVolume()`

UnsetDataVolume ensures that no value is present for DataVolume, not even an explicit nil
### GetPvc

`func (o *CommonStorageFieldsRules) GetPvc() PvcsRules`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *CommonStorageFieldsRules) GetPvcOk() (*PvcsRules, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *CommonStorageFieldsRules) SetPvc(v PvcsRules)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *CommonStorageFieldsRules) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *CommonStorageFieldsRules) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *CommonStorageFieldsRules) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetGit

`func (o *CommonStorageFieldsRules) GetGit() GitsRules`

GetGit returns the Git field if non-nil, zero value otherwise.

### GetGitOk

`func (o *CommonStorageFieldsRules) GetGitOk() (*GitsRules, bool)`

GetGitOk returns a tuple with the Git field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetGit

`func (o *CommonStorageFieldsRules) SetGit(v GitsRules)`

SetGit sets Git field to given value.

### HasGit

`func (o *CommonStorageFieldsRules) HasGit() bool`

HasGit returns a boolean if a field has been set.

### SetGitNil

`func (o *CommonStorageFieldsRules) SetGitNil(b bool)`

 SetGitNil sets the value for Git to be an explicit nil

### UnsetGit
`func (o *CommonStorageFieldsRules) UnsetGit()`

UnsetGit ensures that no value is present for Git, not even an explicit nil
### GetConfigMapVolume

`func (o *CommonStorageFieldsRules) GetConfigMapVolume() ConfigMapsRules`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *CommonStorageFieldsRules) GetConfigMapVolumeOk() (*ConfigMapsRules, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *CommonStorageFieldsRules) SetConfigMapVolume(v ConfigMapsRules)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *CommonStorageFieldsRules) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *CommonStorageFieldsRules) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *CommonStorageFieldsRules) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetSecretVolume

`func (o *CommonStorageFieldsRules) GetSecretVolume() SecretsRules`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *CommonStorageFieldsRules) GetSecretVolumeOk() (*SecretsRules, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *CommonStorageFieldsRules) SetSecretVolume(v SecretsRules)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *CommonStorageFieldsRules) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *CommonStorageFieldsRules) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *CommonStorageFieldsRules) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


