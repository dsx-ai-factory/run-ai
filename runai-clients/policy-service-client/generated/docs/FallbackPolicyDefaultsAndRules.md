# FallbackPolicyDefaultsAndRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Defaults** | Pointer to [**NullableSupersetDefaults**](SupersetDefaults.md) |  | [optional] 
**Rules** | Pointer to [**NullableSupersetRules**](SupersetRules.md) |  | [optional] 
**ImposedAssets** | Pointer to **[]string** | a list of datasource assets to import into the policy. these datasources will be included in any workload that will be created in the scope of the policy. | [optional] 
**Status** | Pointer to [**NullablePolicyValidationStatus**](PolicyValidationStatus.md) |  | [optional] 
**DefaultFromOrder** | Pointer to [**NullableDefaultFromOrder**](DefaultFromOrder.md) |  | [optional] 

## Methods

### NewFallbackPolicyDefaultsAndRules

`func NewFallbackPolicyDefaultsAndRules() *FallbackPolicyDefaultsAndRules`

NewFallbackPolicyDefaultsAndRules instantiates a new FallbackPolicyDefaultsAndRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFallbackPolicyDefaultsAndRulesWithDefaults

`func NewFallbackPolicyDefaultsAndRulesWithDefaults() *FallbackPolicyDefaultsAndRules`

NewFallbackPolicyDefaultsAndRulesWithDefaults instantiates a new FallbackPolicyDefaultsAndRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefaults

`func (o *FallbackPolicyDefaultsAndRules) GetDefaults() SupersetDefaults`

GetDefaults returns the Defaults field if non-nil, zero value otherwise.

### GetDefaultsOk

`func (o *FallbackPolicyDefaultsAndRules) GetDefaultsOk() (*SupersetDefaults, bool)`

GetDefaultsOk returns a tuple with the Defaults field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaults

`func (o *FallbackPolicyDefaultsAndRules) SetDefaults(v SupersetDefaults)`

SetDefaults sets Defaults field to given value.

### HasDefaults

`func (o *FallbackPolicyDefaultsAndRules) HasDefaults() bool`

HasDefaults returns a boolean if a field has been set.

### SetDefaultsNil

`func (o *FallbackPolicyDefaultsAndRules) SetDefaultsNil(b bool)`

 SetDefaultsNil sets the value for Defaults to be an explicit nil

### UnsetDefaults
`func (o *FallbackPolicyDefaultsAndRules) UnsetDefaults()`

UnsetDefaults ensures that no value is present for Defaults, not even an explicit nil
### GetRules

`func (o *FallbackPolicyDefaultsAndRules) GetRules() SupersetRules`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *FallbackPolicyDefaultsAndRules) GetRulesOk() (*SupersetRules, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *FallbackPolicyDefaultsAndRules) SetRules(v SupersetRules)`

SetRules sets Rules field to given value.

### HasRules

`func (o *FallbackPolicyDefaultsAndRules) HasRules() bool`

HasRules returns a boolean if a field has been set.

### SetRulesNil

`func (o *FallbackPolicyDefaultsAndRules) SetRulesNil(b bool)`

 SetRulesNil sets the value for Rules to be an explicit nil

### UnsetRules
`func (o *FallbackPolicyDefaultsAndRules) UnsetRules()`

UnsetRules ensures that no value is present for Rules, not even an explicit nil
### GetImposedAssets

`func (o *FallbackPolicyDefaultsAndRules) GetImposedAssets() []string`

GetImposedAssets returns the ImposedAssets field if non-nil, zero value otherwise.

### GetImposedAssetsOk

`func (o *FallbackPolicyDefaultsAndRules) GetImposedAssetsOk() (*[]string, bool)`

GetImposedAssetsOk returns a tuple with the ImposedAssets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImposedAssets

`func (o *FallbackPolicyDefaultsAndRules) SetImposedAssets(v []string)`

SetImposedAssets sets ImposedAssets field to given value.

### HasImposedAssets

`func (o *FallbackPolicyDefaultsAndRules) HasImposedAssets() bool`

HasImposedAssets returns a boolean if a field has been set.

### SetImposedAssetsNil

`func (o *FallbackPolicyDefaultsAndRules) SetImposedAssetsNil(b bool)`

 SetImposedAssetsNil sets the value for ImposedAssets to be an explicit nil

### UnsetImposedAssets
`func (o *FallbackPolicyDefaultsAndRules) UnsetImposedAssets()`

UnsetImposedAssets ensures that no value is present for ImposedAssets, not even an explicit nil
### GetStatus

`func (o *FallbackPolicyDefaultsAndRules) GetStatus() PolicyValidationStatus`

GetStatus returns the Status field if non-nil, zero value otherwise.

### GetStatusOk

`func (o *FallbackPolicyDefaultsAndRules) GetStatusOk() (*PolicyValidationStatus, bool)`

GetStatusOk returns a tuple with the Status field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStatus

`func (o *FallbackPolicyDefaultsAndRules) SetStatus(v PolicyValidationStatus)`

SetStatus sets Status field to given value.

### HasStatus

`func (o *FallbackPolicyDefaultsAndRules) HasStatus() bool`

HasStatus returns a boolean if a field has been set.

### SetStatusNil

`func (o *FallbackPolicyDefaultsAndRules) SetStatusNil(b bool)`

 SetStatusNil sets the value for Status to be an explicit nil

### UnsetStatus
`func (o *FallbackPolicyDefaultsAndRules) UnsetStatus()`

UnsetStatus ensures that no value is present for Status, not even an explicit nil
### GetDefaultFromOrder

`func (o *FallbackPolicyDefaultsAndRules) GetDefaultFromOrder() DefaultFromOrder`

GetDefaultFromOrder returns the DefaultFromOrder field if non-nil, zero value otherwise.

### GetDefaultFromOrderOk

`func (o *FallbackPolicyDefaultsAndRules) GetDefaultFromOrderOk() (*DefaultFromOrder, bool)`

GetDefaultFromOrderOk returns a tuple with the DefaultFromOrder field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultFromOrder

`func (o *FallbackPolicyDefaultsAndRules) SetDefaultFromOrder(v DefaultFromOrder)`

SetDefaultFromOrder sets DefaultFromOrder field to given value.

### HasDefaultFromOrder

`func (o *FallbackPolicyDefaultsAndRules) HasDefaultFromOrder() bool`

HasDefaultFromOrder returns a boolean if a field has been set.

### SetDefaultFromOrderNil

`func (o *FallbackPolicyDefaultsAndRules) SetDefaultFromOrderNil(b bool)`

 SetDefaultFromOrderNil sets the value for DefaultFromOrder to be an explicit nil

### UnsetDefaultFromOrder
`func (o *FallbackPolicyDefaultsAndRules) UnsetDefaultFromOrder()`

UnsetDefaultFromOrder ensures that no value is present for DefaultFromOrder, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


