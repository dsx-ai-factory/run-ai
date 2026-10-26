# DistributedInferenceTemplate

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateMetaFields**](TemplateMetaFields.md) |  | 
**Spec** | Pointer to [**DistributedInferenceTemplateSpec**](DistributedInferenceTemplateSpec.md) |  | [optional] 
**Runnable** | Pointer to [**RunnableInfo**](RunnableInfo.md) |  | [optional] 
**CombinedSpec** | Pointer to [**DistributedInferenceTemplateSpec**](DistributedInferenceTemplateSpec.md) |  | [optional] 

## Methods

### NewDistributedInferenceTemplate

`func NewDistributedInferenceTemplate(meta TemplateMetaFields, ) *DistributedInferenceTemplate`

NewDistributedInferenceTemplate instantiates a new DistributedInferenceTemplate object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceTemplateWithDefaults

`func NewDistributedInferenceTemplateWithDefaults() *DistributedInferenceTemplate`

NewDistributedInferenceTemplateWithDefaults instantiates a new DistributedInferenceTemplate object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DistributedInferenceTemplate) GetMeta() TemplateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DistributedInferenceTemplate) GetMetaOk() (*TemplateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DistributedInferenceTemplate) SetMeta(v TemplateMetaFields)`

SetMeta sets Meta field to given value.


### GetSpec

`func (o *DistributedInferenceTemplate) GetSpec() DistributedInferenceTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *DistributedInferenceTemplate) GetSpecOk() (*DistributedInferenceTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *DistributedInferenceTemplate) SetSpec(v DistributedInferenceTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *DistributedInferenceTemplate) HasSpec() bool`

HasSpec returns a boolean if a field has been set.

### GetRunnable

`func (o *DistributedInferenceTemplate) GetRunnable() RunnableInfo`

GetRunnable returns the Runnable field if non-nil, zero value otherwise.

### GetRunnableOk

`func (o *DistributedInferenceTemplate) GetRunnableOk() (*RunnableInfo, bool)`

GetRunnableOk returns a tuple with the Runnable field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRunnable

`func (o *DistributedInferenceTemplate) SetRunnable(v RunnableInfo)`

SetRunnable sets Runnable field to given value.

### HasRunnable

`func (o *DistributedInferenceTemplate) HasRunnable() bool`

HasRunnable returns a boolean if a field has been set.

### GetCombinedSpec

`func (o *DistributedInferenceTemplate) GetCombinedSpec() DistributedInferenceTemplateSpec`

GetCombinedSpec returns the CombinedSpec field if non-nil, zero value otherwise.

### GetCombinedSpecOk

`func (o *DistributedInferenceTemplate) GetCombinedSpecOk() (*DistributedInferenceTemplateSpec, bool)`

GetCombinedSpecOk returns a tuple with the CombinedSpec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCombinedSpec

`func (o *DistributedInferenceTemplate) SetCombinedSpec(v DistributedInferenceTemplateSpec)`

SetCombinedSpec sets CombinedSpec field to given value.

### HasCombinedSpec

`func (o *DistributedInferenceTemplate) HasCombinedSpec() bool`

HasCombinedSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


