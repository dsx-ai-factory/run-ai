# TrainingTemplatePatchRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**NullableTemplatePatchMetaFields**](TemplatePatchMetaFields.md) |  | [optional] 
**FieldsToClear** | Pointer to **[]string** |  | [optional] 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**TrainingTemplateSpec**](TrainingTemplateSpec.md) |  | [optional] 

## Methods

### NewTrainingTemplatePatchRequest

`func NewTrainingTemplatePatchRequest() *TrainingTemplatePatchRequest`

NewTrainingTemplatePatchRequest instantiates a new TrainingTemplatePatchRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingTemplatePatchRequestWithDefaults

`func NewTrainingTemplatePatchRequestWithDefaults() *TrainingTemplatePatchRequest`

NewTrainingTemplatePatchRequestWithDefaults instantiates a new TrainingTemplatePatchRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *TrainingTemplatePatchRequest) GetMeta() TemplatePatchMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *TrainingTemplatePatchRequest) GetMetaOk() (*TemplatePatchMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *TrainingTemplatePatchRequest) SetMeta(v TemplatePatchMetaFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *TrainingTemplatePatchRequest) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### SetMetaNil

`func (o *TrainingTemplatePatchRequest) SetMetaNil(b bool)`

 SetMetaNil sets the value for Meta to be an explicit nil

### UnsetMeta
`func (o *TrainingTemplatePatchRequest) UnsetMeta()`

UnsetMeta ensures that no value is present for Meta, not even an explicit nil
### GetFieldsToClear

`func (o *TrainingTemplatePatchRequest) GetFieldsToClear() []string`

GetFieldsToClear returns the FieldsToClear field if non-nil, zero value otherwise.

### GetFieldsToClearOk

`func (o *TrainingTemplatePatchRequest) GetFieldsToClearOk() (*[]string, bool)`

GetFieldsToClearOk returns a tuple with the FieldsToClear field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetFieldsToClear

`func (o *TrainingTemplatePatchRequest) SetFieldsToClear(v []string)`

SetFieldsToClear sets FieldsToClear field to given value.

### HasFieldsToClear

`func (o *TrainingTemplatePatchRequest) HasFieldsToClear() bool`

HasFieldsToClear returns a boolean if a field has been set.

### GetAssets

`func (o *TrainingTemplatePatchRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *TrainingTemplatePatchRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *TrainingTemplatePatchRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *TrainingTemplatePatchRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *TrainingTemplatePatchRequest) GetSpec() TrainingTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *TrainingTemplatePatchRequest) GetSpecOk() (*TrainingTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *TrainingTemplatePatchRequest) SetSpec(v TrainingTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *TrainingTemplatePatchRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


