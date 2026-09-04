# DistributedPolicyDefaultsAndRulesV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Defaults** | Pointer to [**NullableDistributedPolicyDefaultsV2**](DistributedPolicyDefaultsV2.md) |  | [optional] 
**Rules** | Pointer to [**NullableDistributedPolicyRulesV2**](DistributedPolicyRulesV2.md) |  | [optional] 
**ImposedAssets** | Pointer to [**NullableDistributedImposedAssets**](DistributedImposedAssets.md) |  | [optional] 
**Status** | Pointer to [**NullablePolicyValidationStatus**](PolicyValidationStatus.md) |  | [optional] 
**DefaultFromOrder** | Pointer to [**NullableDefaultFromOrder**](DefaultFromOrder.md) |  | [optional] 

## Methods

### NewDistributedPolicyDefaultsAndRulesV2

`func NewDistributedPolicyDefaultsAndRulesV2() *DistributedPolicyDefaultsAndRulesV2`

NewDistributedPolicyDefaultsAndRulesV2 instantiates a new DistributedPolicyDefaultsAndRulesV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedPolicyDefaultsAndRulesV2WithDefaults

`func NewDistributedPolicyDefaultsAndRulesV2WithDefaults() *DistributedPolicyDefaultsAndRulesV2`

NewDistributedPolicyDefaultsAndRulesV2WithDefaults instantiates a new DistributedPolicyDefaultsAndRulesV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefaults

`func (o *DistributedPolicyDefaultsAndRulesV2) GetDefaults() DistributedPolicyDefaultsV2`

GetDefaults returns the Defaults field if non-nil, zero value otherwise.

### GetDefaultsOk

`func (o *DistributedPolicyDefaultsAndRulesV2) GetDefaultsOk() (*DistributedPolicyDefaultsV2, bool)`

GetDefaultsOk returns a tuple with the Defaults field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaults

`func (o *DistributedPolicyDefaultsAndRulesV2) SetDefaults(v DistributedPolicyDefaultsV2)`

SetDefaults sets Defaults field to given value.

### HasDefaults

`func (o *DistributedPolicyDefaultsAndRulesV2) HasDefaults() bool`

HasDefaults returns a boolean if a field has been set.

### SetDefaultsNil

`func (o *DistributedPolicyDefaultsAndRulesV2) SetDefaultsNil(b bool)`

 SetDefaultsNil sets the value for Defaults to be an explicit nil

### UnsetDefaults
`func (o *DistributedPolicyDefaultsAndRulesV2) UnsetDefaults()`

UnsetDefaults ensures that no value is present for Defaults, not even an explicit nil
### GetRules

`func (o *DistributedPolicyDefaultsAndRulesV2) GetRules() DistributedPolicyRulesV2`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *DistributedPolicyDefaultsAndRulesV2) GetRulesOk() (*DistributedPolicyRulesV2, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *DistributedPolicyDefaultsAndRulesV2) SetRules(v DistributedPolicyRulesV2)`

SetRules sets Rules field to given value.

### HasRules

`func (o *DistributedPolicyDefaultsAndRulesV2) HasRules() bool`

HasRules returns a boolean if a field has been set.

### SetRulesNil

`func (o *DistributedPolicyDefaultsAndRulesV2) SetRulesNil(b bool)`

 SetRulesNil sets the value for Rules to be an explicit nil

### UnsetRules
`func (o *DistributedPolicyDefaultsAndRulesV2) UnsetRules()`

UnsetRules ensures that no value is present for Rules, not even an explicit nil
### GetImposedAssets

`func (o *DistributedPolicyDefaultsAndRulesV2) GetImposedAssets() DistributedImposedAssets`

GetImposedAssets returns the ImposedAssets field if non-nil, zero value otherwise.

### GetImposedAssetsOk

`func (o *DistributedPolicyDefaultsAndRulesV2) GetImposedAssetsOk() (*DistributedImposedAssets, bool)`

GetImposedAssetsOk returns a tuple with the ImposedAssets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImposedAssets

`func (o *DistributedPolicyDefaultsAndRulesV2) SetImposedAssets(v DistributedImposedAssets)`

SetImposedAssets sets ImposedAssets field to given value.

### HasImposedAssets

`func (o *DistributedPolicyDefaultsAndRulesV2) HasImposedAssets() bool`

HasImposedAssets returns a boolean if a field has been set.

### SetImposedAssetsNil

`func (o *DistributedPolicyDefaultsAndRulesV2) SetImposedAssetsNil(b bool)`

 SetImposedAssetsNil sets the value for ImposedAssets to be an explicit nil

### UnsetImposedAssets
`func (o *DistributedPolicyDefaultsAndRulesV2) UnsetImposedAssets()`

UnsetImposedAssets ensures that no value is present for ImposedAssets, not even an explicit nil
### GetStatus

`func (o *DistributedPolicyDefaultsAndRulesV2) GetStatus() PolicyValidationStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *DistributedPolicyDefaultsAndRulesV2) GetStatusOk() (*PolicyValidationStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *DistributedPolicyDefaultsAndRulesV2) SetStatus(v PolicyValidationStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *DistributedPolicyDefaultsAndRulesV2) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *DistributedPolicyDefaultsAndRulesV2) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *DistributedPolicyDefaultsAndRulesV2) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetDefaultFromOrder

`func (o *DistributedPolicyDefaultsAndRulesV2) GetDefaultFromOrder() DefaultFromOrder`

GetDefaultFromOrder returns the DefaultFromOrder field if non-nil, zero value otherwise.

### GetDefaultFromOrderOk

`func (o *DistributedPolicyDefaultsAndRulesV2) GetDefaultFromOrderOk() (*DefaultFromOrder, bool)`

GetDefaultFromOrderOk returns a tuple with the DefaultFromOrder field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultFromOrder

`func (o *DistributedPolicyDefaultsAndRulesV2) SetDefaultFromOrder(v DefaultFromOrder)`

SetDefaultFromOrder sets DefaultFromOrder field to given value.

### HasDefaultFromOrder

`func (o *DistributedPolicyDefaultsAndRulesV2) HasDefaultFromOrder() bool`

HasDefaultFromOrder returns a boolean if a field has been set.

### SetDefaultFromOrderNil

`func (o *DistributedPolicyDefaultsAndRulesV2) SetDefaultFromOrderNil(b bool)`

 SetDefaultFromOrderNil sets the value for DefaultFromOrder to be an explicit nil

### UnsetDefaultFromOrder
`func (o *DistributedPolicyDefaultsAndRulesV2) UnsetDefaultFromOrder()`

UnsetDefaultFromOrder ensures that no value is present for DefaultFromOrder, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


