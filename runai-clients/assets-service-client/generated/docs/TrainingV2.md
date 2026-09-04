# TrainingV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**WorkloadMeta**](WorkloadMeta.md) |  | 
**Assets** | [**AssetsRef**](AssetsRef.md) |  | 
**SpecificRunParams** | Pointer to [**NullableTrainingSpecificRunParams**](TrainingSpecificRunParams.md) |  | [optional] 
**PolicyInfo** | Pointer to [**NullablePolicyInfo**](PolicyInfo.md) |  | [optional] 
**Distributed** | Pointer to [**NullableDistributedData**](DistributedData.md) |  | [optional] 

## Methods

### NewTrainingV2

`func NewTrainingV2(meta WorkloadMeta, assets AssetsRef, ) *TrainingV2`

NewTrainingV2 instantiates a new TrainingV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingV2WithDefaults

`func NewTrainingV2WithDefaults() *TrainingV2`

NewTrainingV2WithDefaults instantiates a new TrainingV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *TrainingV2) GetMeta() WorkloadMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *TrainingV2) GetMetaOk() (*WorkloadMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *TrainingV2) SetMeta(v WorkloadMeta)`

SetMeta sets Meta field to given value.


### GetAssets

`func (o *TrainingV2) GetAssets() AssetsRef`

GetAssets returns the Assets field if non-nil, zero value otherwise.

### GetAssetsOk

`func (o *TrainingV2) GetAssetsOk() (*AssetsRef, bool)`

GetAssetsOk returns a tuple with the Assets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAssets

`func (o *TrainingV2) SetAssets(v AssetsRef)`

SetAssets sets Assets field to given value.


### GetSpecificRunParams

`func (o *TrainingV2) GetSpecificRunParams() TrainingSpecificRunParams`

GetSpecificRunParams returns the SpecificRunParams field if non-nil, zero value otherwise.

### GetSpecificRunParamsOk

`func (o *TrainingV2) GetSpecificRunParamsOk() (*TrainingSpecificRunParams, bool)`

GetSpecificRunParamsOk returns a tuple with the SpecificRunParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecificRunParams

`func (o *TrainingV2) SetSpecificRunParams(v TrainingSpecificRunParams)`

SetSpecificRunParams sets SpecificRunParams field to given value.

### HasSpecificRunParams

`func (o *TrainingV2) HasSpecificRunParams() bool`

HasSpecificRunParams returns a boolean if a field has been set.

### SetSpecificRunParamsNil

`func (o *TrainingV2) SetSpecificRunParamsNil(b bool)`

 SetSpecificRunParamsNil sets the value for SpecificRunParams to be an explicit nil

### UnsetSpecificRunParams
`func (o *TrainingV2) UnsetSpecificRunParams()`

UnsetSpecificRunParams ensures that no value is present for SpecificRunParams, not even an explicit nil
### GetPolicyInfo

`func (o *TrainingV2) GetPolicyInfo() PolicyInfo`

GetPolicyInfo returns the PolicyInfo field if non-nil, zero value otherwise.

### GetPolicyInfoOk

`func (o *TrainingV2) GetPolicyInfoOk() (*PolicyInfo, bool)`

GetPolicyInfoOk returns a tuple with the PolicyInfo field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicyInfo

`func (o *TrainingV2) SetPolicyInfo(v PolicyInfo)`

SetPolicyInfo sets PolicyInfo field to given value.

### HasPolicyInfo

`func (o *TrainingV2) HasPolicyInfo() bool`

HasPolicyInfo returns a boolean if a field has been set.

### SetPolicyInfoNil

`func (o *TrainingV2) SetPolicyInfoNil(b bool)`

 SetPolicyInfoNil sets the value for PolicyInfo to be an explicit nil

### UnsetPolicyInfo
`func (o *TrainingV2) UnsetPolicyInfo()`

UnsetPolicyInfo ensures that no value is present for PolicyInfo, not even an explicit nil
### GetDistributed

`func (o *TrainingV2) GetDistributed() DistributedData`

GetDistributed returns the Distributed field if non-nil, zero value otherwise.

### GetDistributedOk

`func (o *TrainingV2) GetDistributedOk() (*DistributedData, bool)`

GetDistributedOk returns a tuple with the Distributed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributed

`func (o *TrainingV2) SetDistributed(v DistributedData)`

SetDistributed sets Distributed field to given value.

### HasDistributed

`func (o *TrainingV2) HasDistributed() bool`

HasDistributed returns a boolean if a field has been set.

### SetDistributedNil

`func (o *TrainingV2) SetDistributedNil(b bool)`

 SetDistributedNil sets the value for Distributed to be an explicit nil

### UnsetDistributed
`func (o *TrainingV2) UnsetDistributed()`

UnsetDistributed ensures that no value is present for Distributed, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


