# QuantityRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory, default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable, default to true | [optional] 
**Min** | Pointer to **NullableString** | The minimum value that the field can be assigned to. | [optional] 
**Max** | Pointer to **NullableString** | The maximum value that the field can be assigned to. | [optional] 

## Methods

### NewQuantityRules

`func NewQuantityRules() *QuantityRules`

NewQuantityRules instantiates a new QuantityRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewQuantityRulesWithDefaults

`func NewQuantityRulesWithDefaults() *QuantityRules`

NewQuantityRulesWithDefaults instantiates a new QuantityRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *QuantityRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *QuantityRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *QuantityRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *QuantityRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *QuantityRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *QuantityRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *QuantityRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *QuantityRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *QuantityRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *QuantityRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *QuantityRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *QuantityRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *QuantityRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *QuantityRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *QuantityRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *QuantityRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *QuantityRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *QuantityRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetMin

`func (o *QuantityRules) GetMin() string`

GetMin returns the Min field if non-nil, zero value otherwise.

### GetMinOk

`func (o *QuantityRules) GetMinOk() (*string, bool)`

GetMinOk returns a tuple with the Min field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMin

`func (o *QuantityRules) SetMin(v string)`

SetMin sets Min field to given value.

### HasMin

`func (o *QuantityRules) HasMin() bool`

HasMin returns a boolean if a field has been set.

### SetMinNil

`func (o *QuantityRules) SetMinNil(b bool)`

 SetMinNil sets the value for Min to be an explicit nil

### UnsetMin
`func (o *QuantityRules) UnsetMin()`

UnsetMin ensures that no value is present for Min, not even an explicit nil
### GetMax

`func (o *QuantityRules) GetMax() string`

GetMax returns the Max field if non-nil, zero value otherwise.

### GetMaxOk

`func (o *QuantityRules) GetMaxOk() (*string, bool)`

GetMaxOk returns a tuple with the Max field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMax

`func (o *QuantityRules) SetMax(v string)`

SetMax sets Max field to given value.

### HasMax

`func (o *QuantityRules) HasMax() bool`

HasMax returns a boolean if a field has been set.

### SetMaxNil

`func (o *QuantityRules) SetMaxNil(b bool)`

 SetMaxNil sets the value for Max to be an explicit nil

### UnsetMax
`func (o *QuantityRules) UnsetMax()`

UnsetMax ensures that no value is present for Max, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


