# WorkflowTaskStorage

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ConfigMapVolume** | Pointer to [**[]ConfigMapInstance**](ConfigMapInstance.md) | Set of config map volumes to use in the workload | [optional] 
**EmptyDirVolume** | Pointer to [**[]EmptyDirInstance**](EmptyDirInstance.md) | A list of emptyDir volumes to mount in the workload. | [optional] 
**HostPath** | Pointer to [**[]HostPathInstance**](HostPathInstance.md) | Set of host paths to use in the workload. | [optional] 
**SecretVolume** | Pointer to [**[]SecretInstance**](SecretInstance.md) | Set of secret volumes to use in the workload | [optional] 

## Methods

### NewWorkflowTaskStorage

`func NewWorkflowTaskStorage() *WorkflowTaskStorage`

NewWorkflowTaskStorage instantiates a new WorkflowTaskStorage object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkflowTaskStorageWithDefaults

`func NewWorkflowTaskStorageWithDefaults() *WorkflowTaskStorage`

NewWorkflowTaskStorageWithDefaults instantiates a new WorkflowTaskStorage object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetConfigMapVolume

`func (o *WorkflowTaskStorage) GetConfigMapVolume() []ConfigMapInstance`

GetConfigMapVolume returns the ConfigMapVolume field if non-nil, zero value otherwise.

### GetConfigMapVolumeOk

`func (o *WorkflowTaskStorage) GetConfigMapVolumeOk() (*[]ConfigMapInstance, bool)`

GetConfigMapVolumeOk returns a tuple with the ConfigMapVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetConfigMapVolume

`func (o *WorkflowTaskStorage) SetConfigMapVolume(v []ConfigMapInstance)`

SetConfigMapVolume sets ConfigMapVolume field to given value.

### HasConfigMapVolume

`func (o *WorkflowTaskStorage) HasConfigMapVolume() bool`

HasConfigMapVolume returns a boolean if a field has been set.

### SetConfigMapVolumeNil

`func (o *WorkflowTaskStorage) SetConfigMapVolumeNil(b bool)`

 SetConfigMapVolumeNil sets the value for ConfigMapVolume to be an explicit nil

### UnsetConfigMapVolume
`func (o *WorkflowTaskStorage) UnsetConfigMapVolume()`

UnsetConfigMapVolume ensures that no value is present for ConfigMapVolume, not even an explicit nil
### GetEmptyDirVolume

`func (o *WorkflowTaskStorage) GetEmptyDirVolume() []EmptyDirInstance`

GetEmptyDirVolume returns the EmptyDirVolume field if non-nil, zero value otherwise.

### GetEmptyDirVolumeOk

`func (o *WorkflowTaskStorage) GetEmptyDirVolumeOk() (*[]EmptyDirInstance, bool)`

GetEmptyDirVolumeOk returns a tuple with the EmptyDirVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEmptyDirVolume

`func (o *WorkflowTaskStorage) SetEmptyDirVolume(v []EmptyDirInstance)`

SetEmptyDirVolume sets EmptyDirVolume field to given value.

### HasEmptyDirVolume

`func (o *WorkflowTaskStorage) HasEmptyDirVolume() bool`

HasEmptyDirVolume returns a boolean if a field has been set.

### SetEmptyDirVolumeNil

`func (o *WorkflowTaskStorage) SetEmptyDirVolumeNil(b bool)`

 SetEmptyDirVolumeNil sets the value for EmptyDirVolume to be an explicit nil

### UnsetEmptyDirVolume
`func (o *WorkflowTaskStorage) UnsetEmptyDirVolume()`

UnsetEmptyDirVolume ensures that no value is present for EmptyDirVolume, not even an explicit nil
### GetHostPath

`func (o *WorkflowTaskStorage) GetHostPath() []HostPathInstance`

GetHostPath returns the HostPath field if non-nil, zero value otherwise.

### GetHostPathOk

`func (o *WorkflowTaskStorage) GetHostPathOk() (*[]HostPathInstance, bool)`

GetHostPathOk returns a tuple with the HostPath field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetHostPath

`func (o *WorkflowTaskStorage) SetHostPath(v []HostPathInstance)`

SetHostPath sets HostPath field to given value.

### HasHostPath

`func (o *WorkflowTaskStorage) HasHostPath() bool`

HasHostPath returns a boolean if a field has been set.

### SetHostPathNil

`func (o *WorkflowTaskStorage) SetHostPathNil(b bool)`

 SetHostPathNil sets the value for HostPath to be an explicit nil

### UnsetHostPath
`func (o *WorkflowTaskStorage) UnsetHostPath()`

UnsetHostPath ensures that no value is present for HostPath, not even an explicit nil
### GetSecretVolume

`func (o *WorkflowTaskStorage) GetSecretVolume() []SecretInstance`

GetSecretVolume returns the SecretVolume field if non-nil, zero value otherwise.

### GetSecretVolumeOk

`func (o *WorkflowTaskStorage) GetSecretVolumeOk() (*[]SecretInstance, bool)`

GetSecretVolumeOk returns a tuple with the SecretVolume field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSecretVolume

`func (o *WorkflowTaskStorage) SetSecretVolume(v []SecretInstance)`

SetSecretVolume sets SecretVolume field to given value.

### HasSecretVolume

`func (o *WorkflowTaskStorage) HasSecretVolume() bool`

HasSecretVolume returns a boolean if a field has been set.

### SetSecretVolumeNil

`func (o *WorkflowTaskStorage) SetSecretVolumeNil(b bool)`

 SetSecretVolumeNil sets the value for SecretVolume to be an explicit nil

### UnsetSecretVolume
`func (o *WorkflowTaskStorage) UnsetSecretVolume()`

UnsetSecretVolume ensures that no value is present for SecretVolume, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


