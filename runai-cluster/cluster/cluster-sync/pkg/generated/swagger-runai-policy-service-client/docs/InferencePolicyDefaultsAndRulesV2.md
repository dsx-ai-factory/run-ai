# InferencePolicyDefaultsAndRulesV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Defaults** | Pointer to [**NullableInferencePolicyDefaultsAndRulesV2Defaults**](InferencePolicyDefaultsAndRulesV2Defaults.md) |  | [optional] 
**Rules** | Pointer to [**NullableInferencePolicyRulesV2**](InferencePolicyRulesV2.md) |  | [optional] 
**ImposedAssets** | Pointer to **[]string** | a list of datasource assets to import into the policy. these datasources will be included in any workload that will be created in the scope of the policy. | [optional] 

## Methods

### NewInferencePolicyDefaultsAndRulesV2

`func NewInferencePolicyDefaultsAndRulesV2() *InferencePolicyDefaultsAndRulesV2`

NewInferencePolicyDefaultsAndRulesV2 instantiates a new InferencePolicyDefaultsAndRulesV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencePolicyDefaultsAndRulesV2WithDefaults

`func NewInferencePolicyDefaultsAndRulesV2WithDefaults() *InferencePolicyDefaultsAndRulesV2`

NewInferencePolicyDefaultsAndRulesV2WithDefaults instantiates a new InferencePolicyDefaultsAndRulesV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefaults

`func (o *InferencePolicyDefaultsAndRulesV2) GetDefaults() InferencePolicyDefaultsAndRulesV2Defaults`

GetDefaults returns the Defaults field if non-nil, zero value otherwise.

### GetDefaultsOk

`func (o *InferencePolicyDefaultsAndRulesV2) GetDefaultsOk() (*InferencePolicyDefaultsAndRulesV2Defaults, bool)`

GetDefaultsOk returns a tuple with the Defaults field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaults

`func (o *InferencePolicyDefaultsAndRulesV2) SetDefaults(v InferencePolicyDefaultsAndRulesV2Defaults)`

SetDefaults sets Defaults field to given value.

### HasDefaults

`func (o *InferencePolicyDefaultsAndRulesV2) HasDefaults() bool`

HasDefaults returns a boolean if a field has been set.

### SetDefaultsNil

`func (o *InferencePolicyDefaultsAndRulesV2) SetDefaultsNil(b bool)`

 SetDefaultsNil sets the value for Defaults to be an explicit nil

### UnsetDefaults
`func (o *InferencePolicyDefaultsAndRulesV2) UnsetDefaults()`

UnsetDefaults ensures that no value is present for Defaults, not even an explicit nil
### GetRules

`func (o *InferencePolicyDefaultsAndRulesV2) GetRules() InferencePolicyRulesV2`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *InferencePolicyDefaultsAndRulesV2) GetRulesOk() (*InferencePolicyRulesV2, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *InferencePolicyDefaultsAndRulesV2) SetRules(v InferencePolicyRulesV2)`

SetRules sets Rules field to given value.

### HasRules

`func (o *InferencePolicyDefaultsAndRulesV2) HasRules() bool`

HasRules returns a boolean if a field has been set.

### SetRulesNil

`func (o *InferencePolicyDefaultsAndRulesV2) SetRulesNil(b bool)`

 SetRulesNil sets the value for Rules to be an explicit nil

### UnsetRules
`func (o *InferencePolicyDefaultsAndRulesV2) UnsetRules()`

UnsetRules ensures that no value is present for Rules, not even an explicit nil
### GetImposedAssets

`func (o *InferencePolicyDefaultsAndRulesV2) GetImposedAssets() []string`

GetImposedAssets returns the ImposedAssets field if non-nil, zero value otherwise.

### GetImposedAssetsOk

`func (o *InferencePolicyDefaultsAndRulesV2) GetImposedAssetsOk() (*[]string, bool)`

GetImposedAssetsOk returns a tuple with the ImposedAssets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImposedAssets

`func (o *InferencePolicyDefaultsAndRulesV2) SetImposedAssets(v []string)`

SetImposedAssets sets ImposedAssets field to given value.

### HasImposedAssets

`func (o *InferencePolicyDefaultsAndRulesV2) HasImposedAssets() bool`

HasImposedAssets returns a boolean if a field has been set.

### SetImposedAssetsNil

`func (o *InferencePolicyDefaultsAndRulesV2) SetImposedAssetsNil(b bool)`

 SetImposedAssetsNil sets the value for ImposedAssets to be an explicit nil

### UnsetImposedAssets
`func (o *InferencePolicyDefaultsAndRulesV2) UnsetImposedAssets()`

UnsetImposedAssets ensures that no value is present for ImposedAssets, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


