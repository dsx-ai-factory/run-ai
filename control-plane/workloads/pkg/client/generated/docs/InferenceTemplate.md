# InferenceTemplate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**InferenceTemplateSpec**](InferenceTemplateSpec.md) |  | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**CombinedSpec** | Pointer to [**InferenceTemplateSpec**](InferenceTemplateSpec.md) |  | [optional] 

## Methods

### NewInferenceTemplate

`func NewInferenceTemplate(meta TemplateMetaFields, ) *InferenceTemplate`

NewInferenceTemplate instantiates a new InferenceTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferenceTemplateWithDefaults

`func NewInferenceTemplateWithDefaults() *InferenceTemplate`

NewInferenceTemplateWithDefaults instantiates a new InferenceTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *InferenceTemplate) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *InferenceTemplate) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *InferenceTemplate) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *InferenceTemplate) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *InferenceTemplate) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *InferenceTemplate) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *InferenceTemplate) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *InferenceTemplate) GetSpec() InferenceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *InferenceTemplate) GetSpecOk() (*InferenceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *InferenceTemplate) SetSpec(v InferenceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *InferenceTemplate) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetRunnable

`func (o *InferenceTemplate) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *InferenceTemplate) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *InferenceTemplate) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *InferenceTemplate) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *InferenceTemplate) GetCombinedSpec() InferenceTemplateSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *InferenceTemplate) GetCombinedSpecOk() (*InferenceTemplateSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *InferenceTemplate) SetCombinedSpec(v InferenceTemplateSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *InferenceTemplate) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


