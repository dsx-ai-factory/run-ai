# DistributedInferenceLeaderWorkerDefaultsV2Storage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMapVolume** | Pointer to [**NullableConfigMapsDefaults**](ConfigMapsDefaults.md) |  | [optional] 
**EmptyDirVolume** | Pointer to [**NullableEmptyDirsDefaults**](EmptyDirsDefaults.md) |  | [optional] 
**Pvc** | Pointer to [**NullablePvcsDefaults**](PvcsDefaults.md) |  | [optional] 
**SecretVolume** | Pointer to [**NullableSecretsDefaults**](SecretsDefaults.md) |  | [optional] 

## Methods

### NewDistributedInferenceLeaderWorkerDefaultsV2Storage

`func NewDistributedInferenceLeaderWorkerDefaultsV2Storage() *DistributedInferenceLeaderWorkerDefaultsV2Storage`

NewDistributedInferenceLeaderWorkerDefaultsV2Storage instantiates a new DistributedInferenceLeaderWorkerDefaultsV2Storage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceLeaderWorkerDefaultsV2StorageWithDefaults

`func NewDistributedInferenceLeaderWorkerDefaultsV2StorageWithDefaults() *DistributedInferenceLeaderWorkerDefaultsV2Storage`

NewDistributedInferenceLeaderWorkerDefaultsV2StorageWithDefaults instantiates a new DistributedInferenceLeaderWorkerDefaultsV2Storage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) GetConfigMapVolume() ConfigMapsDefaults`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) GetConfigMapVolumeOk() (*ConfigMapsDefaults, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) SetConfigMapVolume(v ConfigMapsDefaults)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) GetEmptyDirVolume() EmptyDirsDefaults`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) GetEmptyDirVolumeOk() (*EmptyDirsDefaults, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) SetEmptyDirVolume(v EmptyDirsDefaults)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetPvc

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) GetPvc() PvcsDefaults`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) GetPvcOk() (*PvcsDefaults, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) SetPvc(v PvcsDefaults)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetSecretVolume

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) GetSecretVolume() SecretsDefaults`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) GetSecretVolumeOk() (*SecretsDefaults, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) SetSecretVolume(v SecretsDefaults)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *DistributedInferenceLeaderWorkerDefaultsV2Storage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


