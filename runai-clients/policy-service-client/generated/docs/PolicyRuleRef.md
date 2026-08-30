# PolicyRuleRef

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Field** | **string** | The unique name of the field. | 
**SpecSelector** | **string** | The unique name of the spec selector | 
**ItemSelector** | Pointer to **NullableString** | The name of the item selector that picks specific items of an itemized field. Omit for non-itemized fields. | [optional] 
**SelectorParams** | Pointer to **map[string]interface{}** | The parameter values supplied to the spec selector and item selector at evaluation time, as a key-value map. | [optional] 

## Methods

### NewPolicyRuleRef

`func NewPolicyRuleRef(field string, specSelector string, ) *PolicyRuleRef`

NewPolicyRuleRef instantiates a new PolicyRuleRef object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicyRuleRefWithDefaults

`func NewPolicyRuleRefWithDefaults() *PolicyRuleRef`

NewPolicyRuleRefWithDefaults instantiates a new PolicyRuleRef object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetField

`func (o *PolicyRuleRef) GetField() string`

GetField returns the Field field if non-nil, zero value otherwise.

### GetFieldOk

`func (o *PolicyRuleRef) GetFieldOk() (*string, bool)`

GetFieldOk returns a tuple with the Field field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetField

`func (o *PolicyRuleRef) SetField(v string)`

SetField sets Field field to given value.


### GetSpecSelector

`func (o *PolicyRuleRef) GetSpecSelector() string`

GetSpecSelector returns the SpecSelector field if non-nil, zero value otherwise.

### GetSpecSelectorOk

`func (o *PolicyRuleRef) GetSpecSelectorOk() (*string, bool)`

GetSpecSelectorOk returns a tuple with the SpecSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSpecSelector

`func (o *PolicyRuleRef) SetSpecSelector(v string)`

SetSpecSelector sets SpecSelector field to given value.


### GetItemSelector

`func (o *PolicyRuleRef) GetItemSelector() string`

GetItemSelector returns the ItemSelector field if non-nil, zero value otherwise.

### GetItemSelectorOk

`func (o *PolicyRuleRef) GetItemSelectorOk() (*string, bool)`

GetItemSelectorOk returns a tuple with the ItemSelector field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetItemSelector

`func (o *PolicyRuleRef) SetItemSelector(v string)`

SetItemSelector sets ItemSelector field to given value.

### HasItemSelector

`func (o *PolicyRuleRef) HasItemSelector() bool`

HasItemSelector returns a boolean if a field has been set.

### SetItemSelectorNil

`func (o *PolicyRuleRef) SetItemSelectorNil(b bool)`

 SetItemSelectorNil sets the value for ItemSelector to be an explicit nil

### UnsetItemSelector
`func (o *PolicyRuleRef) UnsetItemSelector()`

UnsetItemSelector ensures that no value is present for ItemSelector, not even an explicit nil
### GetSelectorParams

`func (o *PolicyRuleRef) GetSelectorParams() map[string]interface{}`

GetSelectorParams returns the SelectorParams field if non-nil, zero value otherwise.

### GetSelectorParamsOk

`func (o *PolicyRuleRef) GetSelectorParamsOk() (*map[string]interface{}, bool)`

GetSelectorParamsOk returns a tuple with the SelectorParams field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSelectorParams

`func (o *PolicyRuleRef) SetSelectorParams(v map[string]interface{})`

SetSelectorParams sets SelectorParams field to given value.

### HasSelectorParams

`func (o *PolicyRuleRef) HasSelectorParams() bool`

HasSelectorParams returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


