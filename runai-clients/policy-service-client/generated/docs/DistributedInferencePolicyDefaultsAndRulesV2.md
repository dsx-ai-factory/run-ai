# DistributedInferencePolicyDefaultsAndRulesV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Defaults** | Pointer to [**NullableDistributedInferenceDefaultsV2**](DistributedInferenceDefaultsV2.md) |  | [optional] 
**Rules** | Pointer to [**NullableDistributedInferenceRulesV2**](DistributedInferenceRulesV2.md) |  | [optional] 
**ImposedAssets** | Pointer to [**NullableDistributedInferenceImposedAssets**](DistributedInferenceImposedAssets.md) |  | [optional] 
**Status** | Pointer to [**NullablePolicyValidationStatus**](PolicyValidationStatus.md) |  | [optional] 
**DefaultFromOrder** | Pointer to [**NullableDefaultFromOrder**](DefaultFromOrder.md) |  | [optional] 

## Methods

### NewDistributedInferencePolicyDefaultsAndRulesV2

`func NewDistributedInferencePolicyDefaultsAndRulesV2() *DistributedInferencePolicyDefaultsAndRulesV2`

NewDistributedInferencePolicyDefaultsAndRulesV2 instantiates a new DistributedInferencePolicyDefaultsAndRulesV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferencePolicyDefaultsAndRulesV2WithDefaults

`func NewDistributedInferencePolicyDefaultsAndRulesV2WithDefaults() *DistributedInferencePolicyDefaultsAndRulesV2`

NewDistributedInferencePolicyDefaultsAndRulesV2WithDefaults instantiates a new DistributedInferencePolicyDefaultsAndRulesV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefaults

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetDefaults() DistributedInferenceDefaultsV2`

GetDefaults returns the Defaults field if non-nil, zero value otherwise.

### GetDefaultsOk

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetDefaultsOk() (*DistributedInferenceDefaultsV2, bool)`

GetDefaultsOk returns a tuple with the Defaults field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaults

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetDefaults(v DistributedInferenceDefaultsV2)`

SetDefaults sets Defaults field to given value.

### HasDefaults

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) HasDefaults() bool`

HasDefaults returns a boolean if a field has been set.

### SetDefaultsNil

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetDefaultsNil(b bool)`

 SetDefaultsNil sets the value for Defaults to be an explicit nil

### UnsetDefaults
`func (o *DistributedInferencePolicyDefaultsAndRulesV2) UnsetDefaults()`

UnsetDefaults ensures that no value is present for Defaults, not even an explicit nil
### GetRules

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetRules() DistributedInferenceRulesV2`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetRulesOk() (*DistributedInferenceRulesV2, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetRules(v DistributedInferenceRulesV2)`

SetRules sets Rules field to given value.

### HasRules

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) HasRules() bool`

HasRules returns a boolean if a field has been set.

### SetRulesNil

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetRulesNil(b bool)`

 SetRulesNil sets the value for Rules to be an explicit nil

### UnsetRules
`func (o *DistributedInferencePolicyDefaultsAndRulesV2) UnsetRules()`

UnsetRules ensures that no value is present for Rules, not even an explicit nil
### GetImposedAssets

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetImposedAssets() DistributedInferenceImposedAssets`

GetImposedAssets returns the ImposedAssets field if non-nil, zero value otherwise.

### GetImposedAssetsOk

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetImposedAssetsOk() (*DistributedInferenceImposedAssets, bool)`

GetImposedAssetsOk returns a tuple with the ImposedAssets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImposedAssets

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetImposedAssets(v DistributedInferenceImposedAssets)`

SetImposedAssets sets ImposedAssets field to given value.

### HasImposedAssets

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) HasImposedAssets() bool`

HasImposedAssets returns a boolean if a field has been set.

### SetImposedAssetsNil

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetImposedAssetsNil(b bool)`

 SetImposedAssetsNil sets the value for ImposedAssets to be an explicit nil

### UnsetImposedAssets
`func (o *DistributedInferencePolicyDefaultsAndRulesV2) UnsetImposedAssets()`

UnsetImposedAssets ensures that no value is present for ImposedAssets, not even an explicit nil
### GetStatus

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetStatus() PolicyValidationStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetStatusOk() (*PolicyValidationStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetStatus(v PolicyValidationStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *DistributedInferencePolicyDefaultsAndRulesV2) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetDefaultFromOrder

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetDefaultFromOrder() DefaultFromOrder`

GetDefaultFromOrder returns the DefaultFromOrder field if non-nil, zero value otherwise.

### GetDefaultFromOrderOk

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) GetDefaultFromOrderOk() (*DefaultFromOrder, bool)`

GetDefaultFromOrderOk returns a tuple with the DefaultFromOrder field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultFromOrder

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetDefaultFromOrder(v DefaultFromOrder)`

SetDefaultFromOrder sets DefaultFromOrder field to given value.

### HasDefaultFromOrder

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) HasDefaultFromOrder() bool`

HasDefaultFromOrder returns a boolean if a field has been set.

### SetDefaultFromOrderNil

`func (o *DistributedInferencePolicyDefaultsAndRulesV2) SetDefaultFromOrderNil(b bool)`

 SetDefaultFromOrderNil sets the value for DefaultFromOrder to be an explicit nil

### UnsetDefaultFromOrder
`func (o *DistributedInferencePolicyDefaultsAndRulesV2) UnsetDefaultFromOrder()`

UnsetDefaultFromOrder ensures that no value is present for DefaultFromOrder, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


