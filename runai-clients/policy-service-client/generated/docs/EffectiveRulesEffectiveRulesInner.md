# EffectiveRulesEffectiveRulesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Field** | **string** | The unique name of the field. | 
**SpecSelector** | **string** | The unique name of the spec selector | 
**ItemSelector** | Pointer to **NullableString** | The name of the item selector that picks specific items of an itemized field. Omit for non-itemized fields. | [optional] 
**SelectorParams** | Pointer to **map[string]interface{}** | The parameter values supplied to the spec selector and item selector at evaluation time, as a key-value map. | [optional] 
**AllowOverride** | Pointer to **NullableBool** | when false, a child scope cannot override this rule. default to false. | [optional] 
**Enforcements** | [**[]PolicyRuleEnforcement**](PolicyRuleEnforcement.md) | The set of enforcements that define the rule&#39;s behavior. Each enforcement pairs an enforcement type with its value, and multiple enforcements may combine on a single rule to constrain the targeted field. | 
**SourceOfRule** | [**NullableSourceOfRule**](SourceOfRule.md) |  | 

## Methods

### NewEffectiveRulesEffectiveRulesInner

`func NewEffectiveRulesEffectiveRulesInner(field string, specSelector string, enforcements []PolicyRuleEnforcement, sourceOfRule NullableSourceOfRule, ) *EffectiveRulesEffectiveRulesInner`

NewEffectiveRulesEffectiveRulesInner instantiates a new EffectiveRulesEffectiveRulesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewEffectiveRulesEffectiveRulesInnerWithDefaults

`func NewEffectiveRulesEffectiveRulesInnerWithDefaults() *EffectiveRulesEffectiveRulesInner`

NewEffectiveRulesEffectiveRulesInnerWithDefaults instantiates a new EffectiveRulesEffectiveRulesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetField

`func (o *EffectiveRulesEffectiveRulesInner) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *EffectiveRulesEffectiveRulesInner) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *EffectiveRulesEffectiveRulesInner) SetField(v string)`

SetField sets Field field to given value.


### GetSpecSelector

`func (o *EffectiveRulesEffectiveRulesInner) GetSpecSelector() string`

GetSpecSelector returns the SpecSelector field if non-nil, zero value otherwise.

### GetSpecSelectorOk

`func (o *EffectiveRulesEffectiveRulesInner) GetSpecSelectorOk() (*string, bool)`

GetSpecSelectorOk returns a tuple with the SpecSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelector

`func (o *EffectiveRulesEffectiveRulesInner) SetSpecSelector(v string)`

SetSpecSelector sets SpecSelector field to given value.


### GetItemSelector

`func (o *EffectiveRulesEffectiveRulesInner) GetItemSelector() string`

GetItemSelector returns the ItemSelector field if non-nil, zero value otherwise.

### GetItemSelectorOk

`func (o *EffectiveRulesEffectiveRulesInner) GetItemSelectorOk() (*string, bool)`

GetItemSelectorOk returns a tuple with the ItemSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSelector

`func (o *EffectiveRulesEffectiveRulesInner) SetItemSelector(v string)`

SetItemSelector sets ItemSelector field to given value.

### HasItemSelector

`func (o *EffectiveRulesEffectiveRulesInner) HasItemSelector() bool`

HasItemSelector returns a boolean if a field has been set.

### SetItemSelectorNil

`func (o *EffectiveRulesEffectiveRulesInner) SetItemSelectorNil(b bool)`

 SetItemSelectorNil sets the value for ItemSelector to be an explicit nil

### UnsetItemSelector
`func (o *EffectiveRulesEffectiveRulesInner) UnsetItemSelector()`

UnsetItemSelector ensures that no value is present for ItemSelector, not even an explicit nil
### GetSelectorParams

`func (o *EffectiveRulesEffectiveRulesInner) GetSelectorParams() map[string]interface{}`

GetSelectorParams returns the SelectorParams field if non-nil, zero value otherwise.

### GetSelectorParamsOk

`func (o *EffectiveRulesEffectiveRulesInner) GetSelectorParamsOk() (*map[string]interface{}, bool)`

GetSelectorParamsOk returns a tuple with the SelectorParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectorParams

`func (o *EffectiveRulesEffectiveRulesInner) SetSelectorParams(v map[string]interface{})`

SetSelectorParams sets SelectorParams field to given value.

### HasSelectorParams

`func (o *EffectiveRulesEffectiveRulesInner) HasSelectorParams() bool`

HasSelectorParams returns a boolean if a field has been set.

### GetAllowOverride

`func (o *EffectiveRulesEffectiveRulesInner) GetAllowOverride() bool`

GetAllowOverride returns the AllowOverride field if non-nil, zero value otherwise.

### GetAllowOverrideOk

`func (o *EffectiveRulesEffectiveRulesInner) GetAllowOverrideOk() (*bool, bool)`

GetAllowOverrideOk returns a tuple with the AllowOverride field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverride

`func (o *EffectiveRulesEffectiveRulesInner) SetAllowOverride(v bool)`

SetAllowOverride sets AllowOverride field to given value.

### HasAllowOverride

`func (o *EffectiveRulesEffectiveRulesInner) HasAllowOverride() bool`

HasAllowOverride returns a boolean if a field has been set.

### SetAllowOverrideNil

`func (o *EffectiveRulesEffectiveRulesInner) SetAllowOverrideNil(b bool)`

 SetAllowOverrideNil sets the value for AllowOverride to be an explicit nil

### UnsetAllowOverride
`func (o *EffectiveRulesEffectiveRulesInner) UnsetAllowOverride()`

UnsetAllowOverride ensures that no value is present for AllowOverride, not even an explicit nil
### GetEnforcements

`func (o *EffectiveRulesEffectiveRulesInner) GetEnforcements() []PolicyRuleEnforcement`

GetEnforcements returns the Enforcements field if non-nil, zero value otherwise.

### GetEnforcementsOk

`func (o *EffectiveRulesEffectiveRulesInner) GetEnforcementsOk() (*[]PolicyRuleEnforcement, bool)`

GetEnforcementsOk returns a tuple with the Enforcements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnforcements

`func (o *EffectiveRulesEffectiveRulesInner) SetEnforcements(v []PolicyRuleEnforcement)`

SetEnforcements sets Enforcements field to given value.


### GetSourceOfRule

`func (o *EffectiveRulesEffectiveRulesInner) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *EffectiveRulesEffectiveRulesInner) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *EffectiveRulesEffectiveRulesInner) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.


### SetSourceOfRuleNil

`func (o *EffectiveRulesEffectiveRulesInner) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *EffectiveRulesEffectiveRulesInner) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


