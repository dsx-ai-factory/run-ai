# TrainingTemplate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**TrainingTemplateSpec**](TrainingTemplateSpec.md) |  | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**CombinedSpec** | Pointer to [**TrainingTemplateSpec**](TrainingTemplateSpec.md) |  | [optional] 

## Methods

### NewTrainingTemplate

`func NewTrainingTemplate(meta TemplateMetaFields, ) *TrainingTemplate`

NewTrainingTemplate instantiates a new TrainingTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingTemplateWithDefaults

`func NewTrainingTemplateWithDefaults() *TrainingTemplate`

NewTrainingTemplateWithDefaults instantiates a new TrainingTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *TrainingTemplate) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *TrainingTemplate) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *TrainingTemplate) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *TrainingTemplate) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *TrainingTemplate) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *TrainingTemplate) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *TrainingTemplate) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *TrainingTemplate) GetSpec() TrainingTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *TrainingTemplate) GetSpecOk() (*TrainingTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *TrainingTemplate) SetSpec(v TrainingTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *TrainingTemplate) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetRunnable

`func (o *TrainingTemplate) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *TrainingTemplate) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *TrainingTemplate) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *TrainingTemplate) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *TrainingTemplate) GetCombinedSpec() TrainingTemplateSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *TrainingTemplate) GetCombinedSpecOk() (*TrainingTemplateSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *TrainingTemplate) SetCombinedSpec(v TrainingTemplateSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *TrainingTemplate) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


