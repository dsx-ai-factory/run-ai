# WorkloadCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**ClusterId** | **string** | The id of the cluster. | 
**ProjectId** | **int32** | The id of the project. | 
**Namespace** | **string** | Namespace of the project in which the workload should be created. | 
**Assets** | [**AssetsIds**](AssetsIds.md) |  | 
**SpecificEnv** | Pointer to [**NullableSpecificRunParams**](SpecificRunParams.md) |  | [optional] 
**Distributed** | Pointer to [**NullableDistributed**](Distributed.md) |  | [optional] 

## Methods

### NewWorkloadCreationRequest

`func NewWorkloadCreationRequest(name string, clusterId string, projectId int32, namespace string, assets AssetsIds, ) *WorkloadCreationRequest`

NewWorkloadCreationRequest instantiates a new WorkloadCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkloadCreationRequestWithDefaults

`func NewWorkloadCreationRequestWithDefaults() *WorkloadCreationRequest`

NewWorkloadCreationRequestWithDefaults instantiates a new WorkloadCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *WorkloadCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *WorkloadCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *WorkloadCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *WorkloadCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *WorkloadCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *WorkloadCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *WorkloadCreationRequest) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *WorkloadCreationRequest) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *WorkloadCreationRequest) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.


### GetNamespace

`func (o *WorkloadCreationRequest) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *WorkloadCreationRequest) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *WorkloadCreationRequest) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetAssets

`func (o *WorkloadCreationRequest) GetAssets() AssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *WorkloadCreationRequest) GetAssetsOk() (*AssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *WorkloadCreationRequest) SetAssets(v AssetsIds)`

SetAssets sets Assets field to given value.


### GetSpecificEnv

`func (o *WorkloadCreationRequest) GetSpecificEnv() SpecificRunParams`

GetSpecificEnv returns the SpecificEnv field if non-nil, zero value otherwise.

### GetSpecificEnvOk

`func (o *WorkloadCreationRequest) GetSpecificEnvOk() (*SpecificRunParams, bool)`

GetSpecificEnvOk returns a tuple with the SpecificEnv field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificEnv

`func (o *WorkloadCreationRequest) SetSpecificEnv(v SpecificRunParams)`

SetSpecificEnv sets SpecificEnv field to given value.

### HasSpecificEnv

`func (o *WorkloadCreationRequest) HasSpecificEnv() bool`

HasSpecificEnv returns a boolean if a field has been set.

### SetSpecificEnvNil

`func (o *WorkloadCreationRequest) SetSpecificEnvNil(b bool)`

 SetSpecificEnvNil sets the value for SpecificEnv to be an explicit nil

### UnsetSpecificEnv
`func (o *WorkloadCreationRequest) UnsetSpecificEnv()`

UnsetSpecificEnv ensures that no value is present for SpecificEnv, not even an explicit nil
### GetDistributed

`func (o *WorkloadCreationRequest) GetDistributed() Distributed`

GetDistributed returns the Distributed field if non-nil, zero value otherwise.

### GetDistributedOk

`func (o *WorkloadCreationRequest) GetDistributedOk() (*Distributed, bool)`

GetDistributedOk returns a tuple with the Distributed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributed

`func (o *WorkloadCreationRequest) SetDistributed(v Distributed)`

SetDistributed sets Distributed field to given value.

### HasDistributed

`func (o *WorkloadCreationRequest) HasDistributed() bool`

HasDistributed returns a boolean if a field has been set.

### SetDistributedNil

`func (o *WorkloadCreationRequest) SetDistributedNil(b bool)`

 SetDistributedNil sets the value for Distributed to be an explicit nil

### UnsetDistributed
`func (o *WorkloadCreationRequest) UnsetDistributed()`

UnsetDistributed ensures that no value is present for Distributed, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


