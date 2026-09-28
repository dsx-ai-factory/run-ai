# DistributedInferenceLeaderWorkerRulesV2Storage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMapVolume** | Pointer to [**NullableConfigMapsRules**](ConfigMapsRules.md) |  | [optional] 
**EmptyDirVolume** | Pointer to [**NullableEmptyDirsRules**](EmptyDirsRules.md) |  | [optional] 
**Pvc** | Pointer to [**NullablePvcsRules**](PvcsRules.md) |  | [optional] 
**SecretVolume** | Pointer to [**NullableSecretsRules**](SecretsRules.md) |  | [optional] 

## Methods

### NewDistributedInferenceLeaderWorkerRulesV2Storage

`func NewDistributedInferenceLeaderWorkerRulesV2Storage() *DistributedInferenceLeaderWorkerRulesV2Storage`

NewDistributedInferenceLeaderWorkerRulesV2Storage instantiates a new DistributedInferenceLeaderWorkerRulesV2Storage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceLeaderWorkerRulesV2StorageWithDefaults

`func NewDistributedInferenceLeaderWorkerRulesV2StorageWithDefaults() *DistributedInferenceLeaderWorkerRulesV2Storage`

NewDistributedInferenceLeaderWorkerRulesV2StorageWithDefaults instantiates a new DistributedInferenceLeaderWorkerRulesV2Storage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) GetConfigMapVolume() ConfigMapsRules`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) GetConfigMapVolumeOk() (*ConfigMapsRules, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) SetConfigMapVolume(v ConfigMapsRules)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) GetEmptyDirVolume() EmptyDirsRules`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) GetEmptyDirVolumeOk() (*EmptyDirsRules, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) SetEmptyDirVolume(v EmptyDirsRules)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetPvc

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) GetPvc() PvcsRules`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) GetPvcOk() (*PvcsRules, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) SetPvc(v PvcsRules)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetSecretVolume

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) GetSecretVolume() SecretsRules`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) GetSecretVolumeOk() (*SecretsRules, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) SetSecretVolume(v SecretsRules)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *DistributedInferenceLeaderWorkerRulesV2Storage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


