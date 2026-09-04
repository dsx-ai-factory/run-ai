# InferenceTemplateUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateUpdateMetaFields**](TemplateUpdateMetaFields.md) |  | 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**InferenceTemplateSpec**](InferenceTemplateSpec.md) |  | [optional] 

## Methods

### NewInferenceTemplateUpdateRequest

`func NewInferenceTemplateUpdateRequest(meta TemplateUpdateMetaFields, ) *InferenceTemplateUpdateRequest`

NewInferenceTemplateUpdateRequest instantiates a new InferenceTemplateUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceTemplateUpdateRequestWithDefaults

`func NewInferenceTemplateUpdateRequestWithDefaults() *InferenceTemplateUpdateRequest`

NewInferenceTemplateUpdateRequestWithDefaults instantiates a new InferenceTemplateUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *InferenceTemplateUpdateRequest) GetMeta() TemplateUpdateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *InferenceTemplateUpdateRequest) GetMetaOk() (*TemplateUpdateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *InferenceTemplateUpdateRequest) SetMeta(v TemplateUpdateMetaFields)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *InferenceTemplateUpdateRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *InferenceTemplateUpdateRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *InferenceTemplateUpdateRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *InferenceTemplateUpdateRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *InferenceTemplateUpdateRequest) GetSpec() InferenceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InferenceTemplateUpdateRequest) GetSpecOk() (*InferenceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InferenceTemplateUpdateRequest) SetSpec(v InferenceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InferenceTemplateUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


