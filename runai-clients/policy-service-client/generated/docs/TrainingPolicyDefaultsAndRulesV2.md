# TrainingPolicyDefaultsAndRulesV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Defaults** | Pointer to [**NullableTrainingPolicyDefaultsV2**](TrainingPolicyDefaultsV2.md) |  | [optional] 
**Rules** | Pointer to [**NullableTrainingPolicyRulesV2**](TrainingPolicyRulesV2.md) |  | [optional] 
**ImposedAssets** | Pointer to **[]string** | a list of datasource assets to import into the policy. these datasources will be included in any workload that will be created in the scope of the policy. | [optional] 
**Status** | Pointer to [**NullablePolicyValidationStatus**](PolicyValidationStatus.md) |  | [optional] 
**DefaultFromOrder** | Pointer to [**NullableDefaultFromOrder**](DefaultFromOrder.md) |  | [optional] 

## Methods

### NewTrainingPolicyDefaultsAndRulesV2

`func NewTrainingPolicyDefaultsAndRulesV2() *TrainingPolicyDefaultsAndRulesV2`

NewTrainingPolicyDefaultsAndRulesV2 instantiates a new TrainingPolicyDefaultsAndRulesV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewTrainingPolicyDefaultsAndRulesV2WithDefaults

`func NewTrainingPolicyDefaultsAndRulesV2WithDefaults() *TrainingPolicyDefaultsAndRulesV2`

NewTrainingPolicyDefaultsAndRulesV2WithDefaults instantiates a new TrainingPolicyDefaultsAndRulesV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefaults

`func (o *TrainingPolicyDefaultsAndRulesV2) GetDefaults() TrainingPolicyDefaultsV2`

GetDefaults returns the Defaults field if non-nil, zero value otherwise.

### GetDefaultsOk

`func (o *TrainingPolicyDefaultsAndRulesV2) GetDefaultsOk() (*TrainingPolicyDefaultsV2, bool)`

GetDefaultsOk returns a tuple with the Defaults field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaults

`func (o *TrainingPolicyDefaultsAndRulesV2) SetDefaults(v TrainingPolicyDefaultsV2)`

SetDefaults sets Defaults field to given value.

### HasDefaults

`func (o *TrainingPolicyDefaultsAndRulesV2) HasDefaults() bool`

HasDefaults returns a boolean if a field has been set.

### SetDefaultsNil

`func (o *TrainingPolicyDefaultsAndRulesV2) SetDefaultsNil(b bool)`

 SetDefaultsNil sets the value for Defaults to be an explicit nil

### UnsetDefaults
`func (o *TrainingPolicyDefaultsAndRulesV2) UnsetDefaults()`

UnsetDefaults ensures that no value is present for Defaults, not even an explicit nil
### GetRules

`func (o *TrainingPolicyDefaultsAndRulesV2) GetRules() TrainingPolicyRulesV2`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *TrainingPolicyDefaultsAndRulesV2) GetRulesOk() (*TrainingPolicyRulesV2, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *TrainingPolicyDefaultsAndRulesV2) SetRules(v TrainingPolicyRulesV2)`

SetRules sets Rules field to given value.

### HasRules

`func (o *TrainingPolicyDefaultsAndRulesV2) HasRules() bool`

HasRules returns a boolean if a field has been set.

### SetRulesNil

`func (o *TrainingPolicyDefaultsAndRulesV2) SetRulesNil(b bool)`

 SetRulesNil sets the value for Rules to be an explicit nil

### UnsetRules
`func (o *TrainingPolicyDefaultsAndRulesV2) UnsetRules()`

UnsetRules ensures that no value is present for Rules, not even an explicit nil
### GetImposedAssets

`func (o *TrainingPolicyDefaultsAndRulesV2) GetImposedAssets() []string`

GetImposedAssets returns the ImposedAssets field if non-nil, zero value otherwise.

### GetImposedAssetsOk

`func (o *TrainingPolicyDefaultsAndRulesV2) GetImposedAssetsOk() (*[]string, bool)`

GetImposedAssetsOk returns a tuple with the ImposedAssets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImposedAssets

`func (o *TrainingPolicyDefaultsAndRulesV2) SetImposedAssets(v []string)`

SetImposedAssets sets ImposedAssets field to given value.

### HasImposedAssets

`func (o *TrainingPolicyDefaultsAndRulesV2) HasImposedAssets() bool`

HasImposedAssets returns a boolean if a field has been set.

### SetImposedAssetsNil

`func (o *TrainingPolicyDefaultsAndRulesV2) SetImposedAssetsNil(b bool)`

 SetImposedAssetsNil sets the value for ImposedAssets to be an explicit nil

### UnsetImposedAssets
`func (o *TrainingPolicyDefaultsAndRulesV2) UnsetImposedAssets()`

UnsetImposedAssets ensures that no value is present for ImposedAssets, not even an explicit nil
### GetStatus

`func (o *TrainingPolicyDefaultsAndRulesV2) GetStatus() PolicyValidationStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *TrainingPolicyDefaultsAndRulesV2) GetStatusOk() (*PolicyValidationStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *TrainingPolicyDefaultsAndRulesV2) SetStatus(v PolicyValidationStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *TrainingPolicyDefaultsAndRulesV2) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *TrainingPolicyDefaultsAndRulesV2) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *TrainingPolicyDefaultsAndRulesV2) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetDefaultFromOrder

`func (o *TrainingPolicyDefaultsAndRulesV2) GetDefaultFromOrder() DefaultFromOrder`

GetDefaultFromOrder returns the DefaultFromOrder field if non-nil, zero value otherwise.

### GetDefaultFromOrderOk

`func (o *TrainingPolicyDefaultsAndRulesV2) GetDefaultFromOrderOk() (*DefaultFromOrder, bool)`

GetDefaultFromOrderOk returns a tuple with the DefaultFromOrder field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultFromOrder

`func (o *TrainingPolicyDefaultsAndRulesV2) SetDefaultFromOrder(v DefaultFromOrder)`

SetDefaultFromOrder sets DefaultFromOrder field to given value.

### HasDefaultFromOrder

`func (o *TrainingPolicyDefaultsAndRulesV2) HasDefaultFromOrder() bool`

HasDefaultFromOrder returns a boolean if a field has been set.

### SetDefaultFromOrderNil

`func (o *TrainingPolicyDefaultsAndRulesV2) SetDefaultFromOrderNil(b bool)`

 SetDefaultFromOrderNil sets the value for DefaultFromOrder to be an explicit nil

### UnsetDefaultFromOrder
`func (o *TrainingPolicyDefaultsAndRulesV2) UnsetDefaultFromOrder()`

UnsetDefaultFromOrder ensures that no value is present for DefaultFromOrder, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


