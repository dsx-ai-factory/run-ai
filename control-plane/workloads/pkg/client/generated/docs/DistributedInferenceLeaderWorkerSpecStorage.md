# DistributedInferenceLeaderWorkerSpecStorage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMapVolume** | Pointer to [**[]ConfigMapInstance**](ConfigMapInstance.md) | Set of config map volumes to use in the workload | [optional] 
**EmptyDirVolume** | Pointer to [**[]EmptyDirInstance**](EmptyDirInstance.md) | A list of emptyDir volumes to mount in the workload. | [optional] 
**Pvc** | Pointer to [**[]PvcInstance**](PvcInstance.md) | Set of pvc persistent volume claims to use in the workload. | [optional] 
**SecretVolume** | Pointer to [**[]SecretInstance**](SecretInstance.md) | Set of secret volumes to use in the workload | [optional] 

## Methods

### NewDistributedInferenceLeaderWorkerSpecStorage

`func NewDistributedInferenceLeaderWorkerSpecStorage() *DistributedInferenceLeaderWorkerSpecStorage`

NewDistributedInferenceLeaderWorkerSpecStorage instantiates a new DistributedInferenceLeaderWorkerSpecStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceLeaderWorkerSpecStorageWithDefaults

`func NewDistributedInferenceLeaderWorkerSpecStorageWithDefaults() *DistributedInferenceLeaderWorkerSpecStorage`

NewDistributedInferenceLeaderWorkerSpecStorageWithDefaults instantiates a new DistributedInferenceLeaderWorkerSpecStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *DistributedInferenceLeaderWorkerSpecStorage) GetConfigMapVolume() []ConfigMapInstance`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *DistributedInferenceLeaderWorkerSpecStorage) GetConfigMapVolumeOk() (*[]ConfigMapInstance, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *DistributedInferenceLeaderWorkerSpecStorage) SetConfigMapVolume(v []ConfigMapInstance)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *DistributedInferenceLeaderWorkerSpecStorage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *DistributedInferenceLeaderWorkerSpecStorage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *DistributedInferenceLeaderWorkerSpecStorage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *DistributedInferenceLeaderWorkerSpecStorage) GetEmptyDirVolume() []EmptyDirInstance`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *DistributedInferenceLeaderWorkerSpecStorage) GetEmptyDirVolumeOk() (*[]EmptyDirInstance, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *DistributedInferenceLeaderWorkerSpecStorage) SetEmptyDirVolume(v []EmptyDirInstance)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *DistributedInferenceLeaderWorkerSpecStorage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *DistributedInferenceLeaderWorkerSpecStorage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *DistributedInferenceLeaderWorkerSpecStorage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetPvc

`func (o *DistributedInferenceLeaderWorkerSpecStorage) GetPvc() []PvcInstance`

GetPvc returns the Pvc field if non-nil, zero value otherwise.

### GetPvcOk

`func (o *DistributedInferenceLeaderWorkerSpecStorage) GetPvcOk() (*[]PvcInstance, bool)`

GetPvcOk returns a tuple with the Pvc field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPvc

`func (o *DistributedInferenceLeaderWorkerSpecStorage) SetPvc(v []PvcInstance)`

SetPvc sets Pvc field to given value.

### HasPvc

`func (o *DistributedInferenceLeaderWorkerSpecStorage) HasPvc() bool`

HasPvc returns a boolean if a field has been set.

### SetPvcNil

`func (o *DistributedInferenceLeaderWorkerSpecStorage) SetPvcNil(b bool)`

 SetPvcNil sets the value for Pvc to be an explicit nil

### UnsetPvc
`func (o *DistributedInferenceLeaderWorkerSpecStorage) UnsetPvc()`

UnsetPvc ensures that no value is present for Pvc, not even an explicit nil
### GetSecretVolume

`func (o *DistributedInferenceLeaderWorkerSpecStorage) GetSecretVolume() []SecretInstance`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *DistributedInferenceLeaderWorkerSpecStorage) GetSecretVolumeOk() (*[]SecretInstance, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *DistributedInferenceLeaderWorkerSpecStorage) SetSecretVolume(v []SecretInstance)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *DistributedInferenceLeaderWorkerSpecStorage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *DistributedInferenceLeaderWorkerSpecStorage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *DistributedInferenceLeaderWorkerSpecStorage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


