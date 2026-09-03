# IntegerRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory. default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable. default to true | [optional] 
**Min** | Pointer to **NullableInt32** | The minimum value that the field can be assigned to. | [optional] 
**Max** | Pointer to **NullableInt32** | The maximum value that the field can be assigned to. | [optional] 
**Step** | Pointer to **NullableInt32** | The minimal difference between two values the field can be assigned to. For example, min&#x3D;2, max&#x3D;10, step&#x3D;2 implies that the values the field can hold are 2, 4, 6, 8 and 10. | [optional] 

## Methods

### NewIntegerRules

`func NewIntegerRules() *IntegerRules`

NewIntegerRules instantiates a new IntegerRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIntegerRulesWithDefaults

`func NewIntegerRulesWithDefaults() *IntegerRules`

NewIntegerRulesWithDefaults instantiates a new IntegerRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *IntegerRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *IntegerRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *IntegerRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *IntegerRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *IntegerRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *IntegerRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *IntegerRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *IntegerRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *IntegerRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *IntegerRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *IntegerRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *IntegerRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *IntegerRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *IntegerRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *IntegerRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *IntegerRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *IntegerRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *IntegerRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetMin

`func (o *IntegerRules) GetMin() int32`

GetMin returns the Min field if non-nil, zero value otherwise.

### GetMinOk

`func (o *IntegerRules) GetMinOk() (*int32, bool)`

GetMinOk returns a tuple with the Min field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMin

`func (o *IntegerRules) SetMin(v int32)`

SetMin sets Min field to given value.

### HasMin

`func (o *IntegerRules) HasMin() bool`

HasMin returns a boolean if a field has been set.

### SetMinNil

`func (o *IntegerRules) SetMinNil(b bool)`

 SetMinNil sets the value for Min to be an explicit nil

### UnsetMin
`func (o *IntegerRules) UnsetMin()`

UnsetMin ensures that no value is present for Min, not even an explicit nil
### GetMax

`func (o *IntegerRules) GetMax() int32`

GetMax returns the Max field if non-nil, zero value otherwise.

### GetMaxOk

`func (o *IntegerRules) GetMaxOk() (*int32, bool)`

GetMaxOk returns a tuple with the Max field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMax

`func (o *IntegerRules) SetMax(v int32)`

SetMax sets Max field to given value.

### HasMax

`func (o *IntegerRules) HasMax() bool`

HasMax returns a boolean if a field has been set.

### SetMaxNil

`func (o *IntegerRules) SetMaxNil(b bool)`

 SetMaxNil sets the value for Max to be an explicit nil

### UnsetMax
`func (o *IntegerRules) UnsetMax()`

UnsetMax ensures that no value is present for Max, not even an explicit nil
### GetStep

`func (o *IntegerRules) GetStep() int32`

GetStep returns the Step field if non-nil, zero value otherwise.

### GetStepOk

`func (o *IntegerRules) GetStepOk() (*int32, bool)`

GetStepOk returns a tuple with the Step field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStep

`func (o *IntegerRules) SetStep(v int32)`

SetStep sets Step field to given value.

### HasStep

`func (o *IntegerRules) HasStep() bool`

HasStep returns a boolean if a field has been set.

### SetStepNil

`func (o *IntegerRules) SetStepNil(b bool)`

 SetStepNil sets the value for Step to be an explicit nil

### UnsetStep
`func (o *IntegerRules) UnsetStep()`

UnsetStep ensures that no value is present for Step, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


