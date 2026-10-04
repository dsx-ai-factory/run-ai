# InferenceTemplateCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateCreationMetaFields**](TemplateCreationMetaFields.md) |  | 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**InferenceTemplateSpec**](InferenceTemplateSpec.md) |  | [optional] 

## Methods

### NewInferenceTemplateCreationRequest

`func NewInferenceTemplateCreationRequest(meta TemplateCreationMetaFields, ) *InferenceTemplateCreationRequest`

NewInferenceTemplateCreationRequest instantiates a new InferenceTemplateCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceTemplateCreationRequestWithDefaults

`func NewInferenceTemplateCreationRequestWithDefaults() *InferenceTemplateCreationRequest`

NewInferenceTemplateCreationRequestWithDefaults instantiates a new InferenceTemplateCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *InferenceTemplateCreationRequest) GetMeta() TemplateCreationMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *InferenceTemplateCreationRequest) GetMetaOk() (*TemplateCreationMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *InferenceTemplateCreationRequest) SetMeta(v TemplateCreationMetaFields)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *InferenceTemplateCreationRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *InferenceTemplateCreationRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *InferenceTemplateCreationRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *InferenceTemplateCreationRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *InferenceTemplateCreationRequest) GetSpec() InferenceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InferenceTemplateCreationRequest) GetSpecOk() (*InferenceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InferenceTemplateCreationRequest) SetSpec(v InferenceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InferenceTemplateCreationRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


