# InferenceCreationRequestV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**WorkloadCreationMeta**](WorkloadCreationMeta.md) |  | 
**Assets** | [**AssetsIds**](AssetsIds.md) |  | 
**SpecificRunParams** | Pointer to [**NullableInferenceSpecificRunParams**](InferenceSpecificRunParams.md) |  | [optional] 

## Methods

### NewInferenceCreationRequestV2

`func NewInferenceCreationRequestV2(meta WorkloadCreationMeta, assets AssetsIds, ) *InferenceCreationRequestV2`

NewInferenceCreationRequestV2 instantiates a new InferenceCreationRequestV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceCreationRequestV2WithDefaults

`func NewInferenceCreationRequestV2WithDefaults() *InferenceCreationRequestV2`

NewInferenceCreationRequestV2WithDefaults instantiates a new InferenceCreationRequestV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *InferenceCreationRequestV2) GetMeta() WorkloadCreationMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *InferenceCreationRequestV2) GetMetaOk() (*WorkloadCreationMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *InferenceCreationRequestV2) SetMeta(v WorkloadCreationMeta)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *InferenceCreationRequestV2) GetAssets() AssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *InferenceCreationRequestV2) GetAssetsOk() (*AssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *InferenceCreationRequestV2) SetAssets(v AssetsIds)`

SetAssets sets Assets field to given value.


### GetSpecificRunParams

`func (o *InferenceCreationRequestV2) GetSpecificRunParams() InferenceSpecificRunParams`

GetSpecificRunParams returns the SpecificRunParams field if non-nil, zero value otherwise.

### GetSpecificRunParamsOk

`func (o *InferenceCreationRequestV2) GetSpecificRunParamsOk() (*InferenceSpecificRunParams, bool)`

GetSpecificRunParamsOk returns a tuple with the SpecificRunParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificRunParams

`func (o *InferenceCreationRequestV2) SetSpecificRunParams(v InferenceSpecificRunParams)`

SetSpecificRunParams sets SpecificRunParams field to given value.

### HasSpecificRunParams

`func (o *InferenceCreationRequestV2) HasSpecificRunParams() bool`

HasSpecificRunParams returns a boolean if a field has been set.

### SetSpecificRunParamsNil

`func (o *InferenceCreationRequestV2) SetSpecificRunParamsNil(b bool)`

 SetSpecificRunParamsNil sets the value for SpecificRunParams to be an explicit nil

### UnsetSpecificRunParams
`func (o *InferenceCreationRequestV2) UnsetSpecificRunParams()`

UnsetSpecificRunParams ensures that no value is present for SpecificRunParams, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


