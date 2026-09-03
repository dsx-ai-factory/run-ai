# PolicyRule

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Field** | **string** | The unique name of the field. | 
**SpecSelector** | **string** | The unique name of the spec selector | 
**ItemSelector** | Pointer to **NullableString** | The name of the item selector that picks specific items of an itemized field. Omit for non-itemized fields. | [optional] 
**SelectorParams** | Pointer to **map[string]interface{}** | The parameter values supplied to the spec selector and item selector at evaluation time, as a key-value map. | [optional] 
**AllowOverride** | Pointer to **NullableBool** | when false, a child scope cannot override this rule. default to false. | [optional] 
**Enforcements** | [**[]PolicyRuleEnforcement**](PolicyRuleEnforcement.md) | The set of enforcements that define the rule&#39;s behavior. Each enforcement pairs an enforcement type with its value, and multiple enforcements may combine on a single rule to constrain the targeted field. | 

## Methods

### NewPolicyRule

`func NewPolicyRule(field string, specSelector string, enforcements []PolicyRuleEnforcement, ) *PolicyRule`

NewPolicyRule instantiates a new PolicyRule object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyRuleWithDefaults

`func NewPolicyRuleWithDefaults() *PolicyRule`

NewPolicyRuleWithDefaults instantiates a new PolicyRule object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetField

`func (o *PolicyRule) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *PolicyRule) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *PolicyRule) SetField(v string)`

SetField sets Field field to given value.


### GetSpecSelector

`func (o *PolicyRule) GetSpecSelector() string`

GetSpecSelector returns the SpecSelector field if non-nil, zero value otherwise.

### GetSpecSelectorOk

`func (o *PolicyRule) GetSpecSelectorOk() (*string, bool)`

GetSpecSelectorOk returns a tuple with the SpecSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelector

`func (o *PolicyRule) SetSpecSelector(v string)`

SetSpecSelector sets SpecSelector field to given value.


### GetItemSelector

`func (o *PolicyRule) GetItemSelector() string`

GetItemSelector returns the ItemSelector field if non-nil, zero value otherwise.

### GetItemSelectorOk

`func (o *PolicyRule) GetItemSelectorOk() (*string, bool)`

GetItemSelectorOk returns a tuple with the ItemSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSelector

`func (o *PolicyRule) SetItemSelector(v string)`

SetItemSelector sets ItemSelector field to given value.

### HasItemSelector

`func (o *PolicyRule) HasItemSelector() bool`

HasItemSelector returns a boolean if a field has been set.

### SetItemSelectorNil

`func (o *PolicyRule) SetItemSelectorNil(b bool)`

 SetItemSelectorNil sets the value for ItemSelector to be an explicit nil

### UnsetItemSelector
`func (o *PolicyRule) UnsetItemSelector()`

UnsetItemSelector ensures that no value is present for ItemSelector, not even an explicit nil
### GetSelectorParams

`func (o *PolicyRule) GetSelectorParams() map[string]interface{}`

GetSelectorParams returns the SelectorParams field if non-nil, zero value otherwise.

### GetSelectorParamsOk

`func (o *PolicyRule) GetSelectorParamsOk() (*map[string]interface{}, bool)`

GetSelectorParamsOk returns a tuple with the SelectorParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectorParams

`func (o *PolicyRule) SetSelectorParams(v map[string]interface{})`

SetSelectorParams sets SelectorParams field to given value.

### HasSelectorParams

`func (o *PolicyRule) HasSelectorParams() bool`

HasSelectorParams returns a boolean if a field has been set.

### GetAllowOverride

`func (o *PolicyRule) GetAllowOverride() bool`

GetAllowOverride returns the AllowOverride field if non-nil, zero value otherwise.

### GetAllowOverrideOk

`func (o *PolicyRule) GetAllowOverrideOk() (*bool, bool)`

GetAllowOverrideOk returns a tuple with the AllowOverride field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetAllowOverride

`func (o *PolicyRule) SetAllowOverride(v bool)`

SetAllowOverride sets AllowOverride field to given value.

### HasAllowOverride

`func (o *PolicyRule) HasAllowOverride() bool`

HasAllowOverride returns a boolean if a field has been set.

### SetAllowOverrideNil

`func (o *PolicyRule) SetAllowOverrideNil(b bool)`

 SetAllowOverrideNil sets the value for AllowOverride to be an explicit nil

### UnsetAllowOverride
`func (o *PolicyRule) UnsetAllowOverride()`

UnsetAllowOverride ensures that no value is present for AllowOverride, not even an explicit nil
### GetEnforcements

`func (o *PolicyRule) GetEnforcements() []PolicyRuleEnforcement`

GetEnforcements returns the Enforcements field if non-nil, zero value otherwise.

### GetEnforcementsOk

`func (o *PolicyRule) GetEnforcementsOk() (*[]PolicyRuleEnforcement, bool)`

GetEnforcementsOk returns a tuple with the Enforcements field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEnforcements

`func (o *PolicyRule) SetEnforcements(v []PolicyRuleEnforcement)`

SetEnforcements sets Enforcements field to given value.



[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


