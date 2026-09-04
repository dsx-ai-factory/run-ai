# InferenceV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**WorkloadMeta**](WorkloadMeta.md) |  | 
**Assets** | [**AssetsRef**](AssetsRef.md) |  | 
**SpecificRunParams** | Pointer to [**NullableInferenceSpecificRunParams**](InferenceSpecificRunParams.md) |  | [optional] 

## Methods

### NewInferenceV2

`func NewInferenceV2(meta WorkloadMeta, assets AssetsRef, ) *InferenceV2`

NewInferenceV2 instantiates a new InferenceV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceV2WithDefaults

`func NewInferenceV2WithDefaults() *InferenceV2`

NewInferenceV2WithDefaults instantiates a new InferenceV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *InferenceV2) GetMeta() WorkloadMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *InferenceV2) GetMetaOk() (*WorkloadMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *InferenceV2) SetMeta(v WorkloadMeta)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *InferenceV2) GetAssets() AssetsRef`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *InferenceV2) GetAssetsOk() (*AssetsRef, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *InferenceV2) SetAssets(v AssetsRef)`

SetAssets sets Assets field to given value.


### GetSpecificRunParams

`func (o *InferenceV2) GetSpecificRunParams() InferenceSpecificRunParams`

GetSpecificRunParams returns the SpecificRunParams field if non-nil, zero value otherwise.

### GetSpecificRunParamsOk

`func (o *InferenceV2) GetSpecificRunParamsOk() (*InferenceSpecificRunParams, bool)`

GetSpecificRunParamsOk returns a tuple with the SpecificRunParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificRunParams

`func (o *InferenceV2) SetSpecificRunParams(v InferenceSpecificRunParams)`

SetSpecificRunParams sets SpecificRunParams field to given value.

### HasSpecificRunParams

`func (o *InferenceV2) HasSpecificRunParams() bool`

HasSpecificRunParams returns a boolean if a field has been set.

### SetSpecificRunParamsNil

`func (o *InferenceV2) SetSpecificRunParamsNil(b bool)`

 SetSpecificRunParamsNil sets the value for SpecificRunParams to be an explicit nil

### UnsetSpecificRunParams
`func (o *InferenceV2) UnsetSpecificRunParams()`

UnsetSpecificRunParams ensures that no value is present for SpecificRunParams, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


