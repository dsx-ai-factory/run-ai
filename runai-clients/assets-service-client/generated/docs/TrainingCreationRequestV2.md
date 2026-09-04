# TrainingCreationRequestV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**WorkloadCreationMeta**](WorkloadCreationMeta.md) |  | [optional] 
**Assets** | Pointer to [**AssetsIds**](AssetsIds.md) |  | [optional] 
**SpecificRunParams** | Pointer to [**NullableTrainingSpecificRunParams**](TrainingSpecificRunParams.md) |  | [optional] 
**Distributed** | Pointer to [**NullableDistributedParams**](DistributedParams.md) |  | [optional] 

## Methods

### NewTrainingCreationRequestV2

`func NewTrainingCreationRequestV2() *TrainingCreationRequestV2`

NewTrainingCreationRequestV2 instantiates a new TrainingCreationRequestV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingCreationRequestV2WithDefaults

`func NewTrainingCreationRequestV2WithDefaults() *TrainingCreationRequestV2`

NewTrainingCreationRequestV2WithDefaults instantiates a new TrainingCreationRequestV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *TrainingCreationRequestV2) GetMeta() WorkloadCreationMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *TrainingCreationRequestV2) GetMetaOk() (*WorkloadCreationMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *TrainingCreationRequestV2) SetMeta(v WorkloadCreationMeta)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *TrainingCreationRequestV2) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetAssets

`func (o *TrainingCreationRequestV2) GetAssets() AssetsIds`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *TrainingCreationRequestV2) GetAssetsOk() (*AssetsIds, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *TrainingCreationRequestV2) SetAssets(v AssetsIds)`

SetAssets sets Assets field to given value.

### HasAssets

`func (o *TrainingCreationRequestV2) HasAssets() bool`

HasAssets returns a boolean if a field has been set.

### GetSpecificRunParams

`func (o *TrainingCreationRequestV2) GetSpecificRunParams() TrainingSpecificRunParams`

GetSpecificRunParams returns the SpecificRunParams field if non-nil, zero value otherwise.

### GetSpecificRunParamsOk

`func (o *TrainingCreationRequestV2) GetSpecificRunParamsOk() (*TrainingSpecificRunParams, bool)`

GetSpecificRunParamsOk returns a tuple with the SpecificRunParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificRunParams

`func (o *TrainingCreationRequestV2) SetSpecificRunParams(v TrainingSpecificRunParams)`

SetSpecificRunParams sets SpecificRunParams field to given value.

### HasSpecificRunParams

`func (o *TrainingCreationRequestV2) HasSpecificRunParams() bool`

HasSpecificRunParams returns a boolean if a field has been set.

### SetSpecificRunParamsNil

`func (o *TrainingCreationRequestV2) SetSpecificRunParamsNil(b bool)`

 SetSpecificRunParamsNil sets the value for SpecificRunParams to be an explicit nil

### UnsetSpecificRunParams
`func (o *TrainingCreationRequestV2) UnsetSpecificRunParams()`

UnsetSpecificRunParams ensures that no value is present for SpecificRunParams, not even an explicit nil
### GetDistributed

`func (o *TrainingCreationRequestV2) GetDistributed() DistributedParams`

GetDistributed returns the Distributed field if non-nil, zero value otherwise.

### GetDistributedOk

`func (o *TrainingCreationRequestV2) GetDistributedOk() (*DistributedParams, bool)`

GetDistributedOk returns a tuple with the Distributed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributed

`func (o *TrainingCreationRequestV2) SetDistributed(v DistributedParams)`

SetDistributed sets Distributed field to given value.

### HasDistributed

`func (o *TrainingCreationRequestV2) HasDistributed() bool`

HasDistributed returns a boolean if a field has been set.

### SetDistributedNil

`func (o *TrainingCreationRequestV2) SetDistributedNil(b bool)`

 SetDistributedNil sets the value for Distributed to be an explicit nil

### UnsetDistributed
`func (o *TrainingCreationRequestV2) UnsetDistributed()`

UnsetDistributed ensures that no value is present for Distributed, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


