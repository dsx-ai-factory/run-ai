# TrainingTemplateCreationRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**TemplateCreationMetaFields**](TemplateCreationMetaFields.md) |  | 
**Assets** | Pointer to [**TemplateAssetsIds**](TemplateAssetsIds.md) |  | [optional] 
**Spec** | Pointer to [**TrainingTemplateSpec**](TrainingTemplateSpec.md) |  | [optional] 

## Methods

### NewTrainingTemplateCreationRequest

`func NewTrainingTemplateCreationRequest(meta TemplateCreationMetaFields, ) *TrainingTemplateCreationRequest`

NewTrainingTemplateCreationRequest instantiates a new TrainingTemplateCreationRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingTemplateCreationRequestWithDefaults

`func NewTrainingTemplateCreationRequestWithDefaults() *TrainingTemplateCreationRequest`

NewTrainingTemplateCreationRequestWithDefaults instantiates a new TrainingTemplateCreationRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *TrainingTemplateCreationRequest) GetMeta() TemplateCreationMetaFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *TrainingTemplateCreationRequest) GetMetaOk() (*TemplateCreationMetaFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *TrainingTemplateCreationRequest) SetMeta(v TemplateCreationMetaFields)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *TrainingTemplateCreationRequest) GetAssets() TemplateAssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *TrainingTemplateCreationRequest) GetAssetsOk() (*TemplateAssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *TrainingTemplateCreationRequest) SetAssets(v TemplateAssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *TrainingTemplateCreationRequest) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpec

`func (o *TrainingTemplateCreationRequest) GetSpec() TrainingTemplateSpec`

GetSpec returns the Spec field if non-nil, zero value otherwise.

### GetSpecOk

`func (o *TrainingTemplateCreationRequest) GetSpecOk() (*TrainingTemplateSpec, bool)`

GetSpecOk returns a tuple with the Spec field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpec

`func (o *TrainingTemplateCreationRequest) SetSpec(v TrainingTemplateSpec)`

SetSpec sets Spec field to given value.

### HasSpec

`func (o *TrainingTemplateCreationRequest) HasSpec() bool`

HasSpec returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


