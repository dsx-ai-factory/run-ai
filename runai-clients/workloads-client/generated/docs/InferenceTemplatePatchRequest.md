# InferenceTemplatePatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**NullableTemplatePatchMetaFields**](TemplatePatchMetaFields.md) |  | [optional] 
**FieldsToClear** | Pointer to **[]string** |  | [optional] 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**InferenceTemplateSpec**](InferenceTemplateSpec.md) |  | [optional] 

## Methods

### NewInferenceTemplatePatchRequest

`func NewInferenceTemplatePatchRequest() *InferenceTemplatePatchRequest`

NewInferenceTemplatePatchRequest instantiates a new InferenceTemplatePatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceTemplatePatchRequestWithDefaults

`func NewInferenceTemplatePatchRequestWithDefaults() *InferenceTemplatePatchRequest`

NewInferenceTemplatePatchRequestWithDefaults instantiates a new InferenceTemplatePatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *InferenceTemplatePatchRequest) GetMeta() TemplatePatchMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *InferenceTemplatePatchRequest) GetMetaOk() (*TemplatePatchMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *InferenceTemplatePatchRequest) SetMeta(v TemplatePatchMetaFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *InferenceTemplatePatchRequest) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### SetMetaNil

`func (o *InferenceTemplatePatchRequest) SetMetaNil(b bool)`

 SetMetaNil sets the value for Meta to be an explicit nil

### UnsetMeta
`func (o *InferenceTemplatePatchRequest) UnsetMeta()`

UnsetMeta ensures that no value is present for Meta, not even an explicit nil
### GetFieldsToClear

`func (o *InferenceTemplatePatchRequest) GetFieldsToClear() []string`

GetFieldsToClear returns the FieldsToClear field if non-nil, zero value otherwise.

### GetFieldsToClearOk

`func (o *InferenceTemplatePatchRequest) GetFieldsToClearOk() (*[]string, bool)`

GetFieldsToClearOk returns a tuple with the FieldsToClear field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldsToClear

`func (o *InferenceTemplatePatchRequest) SetFieldsToClear(v []string)`

SetFieldsToClear sets FieldsToClear field to given value.

### HasFieldsToClear

`func (o *InferenceTemplatePatchRequest) HasFieldsToClear() bool`

HasFieldsToClear returns a boolean if a field has been set.

### GetAssets

`func (o *InferenceTemplatePatchRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *InferenceTemplatePatchRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *InferenceTemplatePatchRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *InferenceTemplatePatchRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *InferenceTemplatePatchRequest) GetSpec() InferenceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InferenceTemplatePatchRequest) GetSpecOk() (*InferenceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InferenceTemplatePatchRequest) SetSpec(v InferenceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InferenceTemplatePatchRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


