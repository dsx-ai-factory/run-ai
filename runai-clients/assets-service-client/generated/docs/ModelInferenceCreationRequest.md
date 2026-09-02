# ModelInferenceCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Name** | **string** | The name of the workload. | 
**ClusterId** | **string** | The id of the cluster. | 
**ProjectId** | **int32** | The id of the project. | 
**Namespace** | **string** | Namespace of the project in which the workload should be created. | 
**Assets** | Pointer to [**ModelInferenceCreationRequestAssets**](ModelInferenceCreationRequestAssets.md) |  | [optional] 
**ModelId** | **string** | The id of the model. | 
**SpecificEnv** | Pointer to [**NullableSpecificRunParams**](SpecificRunParams.md) |  | [optional] 

## Methods

### NewModelInferenceCreationRequest

`func NewModelInferenceCreationRequest(name string, clusterId string, projectId int32, namespace string, modelId string, ) *ModelInferenceCreationRequest`

NewModelInferenceCreationRequest instantiates a new ModelInferenceCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewModelInferenceCreationRequestWithDefaults

`func NewModelInferenceCreationRequestWithDefaults() *ModelInferenceCreationRequest`

NewModelInferenceCreationRequestWithDefaults instantiates a new ModelInferenceCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetName

`func (o *ModelInferenceCreationRequest) GetName() string`

GetName returns the Name field if non-nil, zero value otherwise.

### GetNameOk

`func (o *ModelInferenceCreationRequest) GetNameOk() (*string, bool)`

GetNameOk returns a tuple with the Name field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetName

`func (o *ModelInferenceCreationRequest) SetName(v string)`

SetName sets Name field to given value.


### GetClusterId

`func (o *ModelInferenceCreationRequest) GetClusterId() string`

GetClusterId returns the ClusterId field if non-nil, zero value otherwise.

### GetClusterIdOk

`func (o *ModelInferenceCreationRequest) GetClusterIdOk() (*string, bool)`

GetClusterIdOk returns a tuple with the ClusterId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClusterId

`func (o *ModelInferenceCreationRequest) SetClusterId(v string)`

SetClusterId sets ClusterId field to given value.


### GetProjectId

`func (o *ModelInferenceCreationRequest) GetProjectId() int32`

GetProjectId returns the ProjectId field if non-nil, zero value otherwise.

### GetProjectIdOk

`func (o *ModelInferenceCreationRequest) GetProjectIdOk() (*int32, bool)`

GetProjectIdOk returns a tuple with the ProjectId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetProjectId

`func (o *ModelInferenceCreationRequest) SetProjectId(v int32)`

SetProjectId sets ProjectId field to given value.


### GetNamespace

`func (o *ModelInferenceCreationRequest) GetNamespace() string`

GetNamespace returns the Namespace field if non-nil, zero value otherwise.

### GetNamespaceOk

`func (o *ModelInferenceCreationRequest) GetNamespaceOk() (*string, bool)`

GetNamespaceOk returns a tuple with the Namespace field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetNamespace

`func (o *ModelInferenceCreationRequest) SetNamespace(v string)`

SetNamespace sets Namespace field to given value.


### GetAssets

`func (o *ModelInferenceCreationRequest) GetAssets() ModelInferenceCreationRequestAssets`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *ModelInferenceCreationRequest) GetAssetsOk() (*ModelInferenceCreationRequestAssets, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *ModelInferenceCreationRequest) SetAssets(v ModelInferenceCreationRequestAssets)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *ModelInferenceCreationRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetModelId

`func (o *ModelInferenceCreationRequest) GetModelId() string`

GetModelId returns the ModelId field if non-nil, zero value otherwise.

### GetModelIdOk

`func (o *ModelInferenceCreationRequest) GetModelIdOk() (*string, bool)`

GetModelIdOk returns a tuple with the ModelId field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetModelId

`func (o *ModelInferenceCreationRequest) SetModelId(v string)`

SetModelId sets ModelId field to given value.


### GetSpecificEnv

`func (o *ModelInferenceCreationRequest) GetSpecificEnv() SpecificRunParams`

GetSpecificEnv returns the SpecificEnv field if non-nil, zero value otherwise.

### GetSpecificEnvOk

`func (o *ModelInferenceCreationRequest) GetSpecificEnvOk() (*SpecificRunParams, bool)`

GetSpecificEnvOk returns a tuple with the SpecificEnv field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificEnv

`func (o *ModelInferenceCreationRequest) SetSpecificEnv(v SpecificRunParams)`

SetSpecificEnv sets SpecificEnv field to given value.

### HasSpecificEnv

`func (o *ModelInferenceCreationRequest) HasSpecificEnv() bool`

HasSpecificEnv returns a boolean if a field has been set.

### SetSpecificEnvNil

`func (o *ModelInferenceCreationRequest) SetSpecificEnvNil(b bool)`

 SetSpecificEnvNil sets the value for SpecificEnv to be an explicit nil

### UnsetSpecificEnv
`func (o *ModelInferenceCreationRequest) UnsetSpecificEnv()`

UnsetSpecificEnv ensures that no value is present for SpecificEnv, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


