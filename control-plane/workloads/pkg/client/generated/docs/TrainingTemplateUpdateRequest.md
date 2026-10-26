# TrainingTemplateUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateUpdateMetaFields**](TemplateUpdateMetaFields.md) |  | 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**TrainingTemplateSpec**](TrainingTemplateSpec.md) |  | [optional] 

## Methods

### NewTrainingTemplateUpdateRequest

`func NewTrainingTemplateUpdateRequest(meta TemplateUpdateMetaFields, ) *TrainingTemplateUpdateRequest`

NewTrainingTemplateUpdateRequest instantiates a new TrainingTemplateUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingTemplateUpdateRequestWithDefaults

`func NewTrainingTemplateUpdateRequestWithDefaults() *TrainingTemplateUpdateRequest`

NewTrainingTemplateUpdateRequestWithDefaults instantiates a new TrainingTemplateUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *TrainingTemplateUpdateRequest) GetMeta() TemplateUpdateMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *TrainingTemplateUpdateRequest) GetMetaOk() (*TemplateUpdateMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *TrainingTemplateUpdateRequest) SetMeta(v TemplateUpdateMetaFields)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *TrainingTemplateUpdateRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *TrainingTemplateUpdateRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *TrainingTemplateUpdateRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *TrainingTemplateUpdateRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *TrainingTemplateUpdateRequest) GetSpec() TrainingTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *TrainingTemplateUpdateRequest) GetSpecOk() (*TrainingTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *TrainingTemplateUpdateRequest) SetSpec(v TrainingTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *TrainingTemplateUpdateRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


