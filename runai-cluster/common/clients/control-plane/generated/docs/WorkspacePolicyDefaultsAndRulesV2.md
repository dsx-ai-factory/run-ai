# WorkspacePolicyDefaultsAndRulesV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Defaults** | Pointer to [**NullableWorkspacePolicyDefaultsAndRulesV2Defaults**](WorkspacePolicyDefaultsAndRulesV2Defaults.md) |  | [optional] 
**Rules** | Pointer to [**NullableWorkspacePolicyDefaultsAndRulesV2Rules**](WorkspacePolicyDefaultsAndRulesV2Rules.md) |  | [optional] 
**ImposedAssets** | Pointer to **[]string** | a list of datasource assets to import into the policy. these datasources will be included in any workload that will be created in the scope of the policy. | [optional] 

## Methods

### NewWorkspacePolicyDefaultsAndRulesV2

`func NewWorkspacePolicyDefaultsAndRulesV2() *WorkspacePolicyDefaultsAndRulesV2`

NewWorkspacePolicyDefaultsAndRulesV2 instantiates a new WorkspacePolicyDefaultsAndRulesV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspacePolicyDefaultsAndRulesV2WithDefaults

`func NewWorkspacePolicyDefaultsAndRulesV2WithDefaults() *WorkspacePolicyDefaultsAndRulesV2`

NewWorkspacePolicyDefaultsAndRulesV2WithDefaults instantiates a new WorkspacePolicyDefaultsAndRulesV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetDefaults

`func (o *WorkspacePolicyDefaultsAndRulesV2) GetDefaults() WorkspacePolicyDefaultsAndRulesV2Defaults`

GetDefaults returns the Defaults field if non-nil, zero value otherwise.

### GetDefaultsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2) GetDefaultsOk() (*WorkspacePolicyDefaultsAndRulesV2Defaults, bool)`

GetDefaultsOk returns a tuple with the Defaults field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaults

`func (o *WorkspacePolicyDefaultsAndRulesV2) SetDefaults(v WorkspacePolicyDefaultsAndRulesV2Defaults)`

SetDefaults sets Defaults field to given value.

### HasDefaults

`func (o *WorkspacePolicyDefaultsAndRulesV2) HasDefaults() bool`

HasDefaults returns a boolean if a field has been set.

### SetDefaultsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2) SetDefaultsNil(b bool)`

 SetDefaultsNil sets the value for Defaults to be an explicit nil

### UnsetDefaults
`func (o *WorkspacePolicyDefaultsAndRulesV2) UnsetDefaults()`

UnsetDefaults ensures that no value is present for Defaults, not even an explicit nil
### GetRules

`func (o *WorkspacePolicyDefaultsAndRulesV2) GetRules() WorkspacePolicyDefaultsAndRulesV2Rules`

GetRules returns the Rules field if non-nil, zero value otherwise.

### GetRulesOk

`func (o *WorkspacePolicyDefaultsAndRulesV2) GetRulesOk() (*WorkspacePolicyDefaultsAndRulesV2Rules, bool)`

GetRulesOk returns a tuple with the Rules field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRules

`func (o *WorkspacePolicyDefaultsAndRulesV2) SetRules(v WorkspacePolicyDefaultsAndRulesV2Rules)`

SetRules sets Rules field to given value.

### HasRules

`func (o *WorkspacePolicyDefaultsAndRulesV2) HasRules() bool`

HasRules returns a boolean if a field has been set.

### SetRulesNil

`func (o *WorkspacePolicyDefaultsAndRulesV2) SetRulesNil(b bool)`

 SetRulesNil sets the value for Rules to be an explicit nil

### UnsetRules
`func (o *WorkspacePolicyDefaultsAndRulesV2) UnsetRules()`

UnsetRules ensures that no value is present for Rules, not even an explicit nil
### GetImposedAssets

`func (o *WorkspacePolicyDefaultsAndRulesV2) GetImposedAssets() []string`

GetImposedAssets returns the ImposedAssets field if non-nil, zero value otherwise.

### GetImposedAssetsOk

`func (o *WorkspacePolicyDefaultsAndRulesV2) GetImposedAssetsOk() (*[]string, bool)`

GetImposedAssetsOk returns a tuple with the ImposedAssets field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetImposedAssets

`func (o *WorkspacePolicyDefaultsAndRulesV2) SetImposedAssets(v []string)`

SetImposedAssets sets ImposedAssets field to given value.

### HasImposedAssets

`func (o *WorkspacePolicyDefaultsAndRulesV2) HasImposedAssets() bool`

HasImposedAssets returns a boolean if a field has been set.

### SetImposedAssetsNil

`func (o *WorkspacePolicyDefaultsAndRulesV2) SetImposedAssetsNil(b bool)`

 SetImposedAssetsNil sets the value for ImposedAssets to be an explicit nil

### UnsetImposedAssets
`func (o *WorkspacePolicyDefaultsAndRulesV2) UnsetImposedAssets()`

UnsetImposedAssets ensures that no value is present for ImposedAssets, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


