# IntegerRulesOptional

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable. default to true | [optional] 
**Min** | Pointer to **NullableInt32** | The minimum value that the field can be assigned to. | [optional] 
**Max** | Pointer to **NullableInt32** | The maximum value that the field can be assigned to. | [optional] 
**Step** | Pointer to **NullableInt32** | The minimal difference between two values the field can be assigned to. For example, min&#x3D;2, max&#x3D;10, step&#x3D;2 implies that the values the field can hold are 2, 4, 6, 8 and 10. | [optional] 

## Methods

### NewIntegerRulesOptional

`func NewIntegerRulesOptional() *IntegerRulesOptional`

NewIntegerRulesOptional instantiates a new IntegerRulesOptional object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewIntegerRulesOptionalWithDefaults

`func NewIntegerRulesOptionalWithDefaults() *IntegerRulesOptional`

NewIntegerRulesOptionalWithDefaults instantiates a new IntegerRulesOptional object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *IntegerRulesOptional) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *IntegerRulesOptional) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *IntegerRulesOptional) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *IntegerRulesOptional) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *IntegerRulesOptional) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *IntegerRulesOptional) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetCanEdit

`func (o *IntegerRulesOptional) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *IntegerRulesOptional) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *IntegerRulesOptional) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *IntegerRulesOptional) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *IntegerRulesOptional) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *IntegerRulesOptional) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetMin

`func (o *IntegerRulesOptional) GetMin() int32`

GetMin returns the Min field if non-nil, zero value otherwise.

### GetMinOk

`func (o *IntegerRulesOptional) GetMinOk() (*int32, bool)`

GetMinOk returns a tuple with the Min field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMin

`func (o *IntegerRulesOptional) SetMin(v int32)`

SetMin sets Min field to given value.

### HasMin

`func (o *IntegerRulesOptional) HasMin() bool`

HasMin returns a boolean if a field has been set.

### SetMinNil

`func (o *IntegerRulesOptional) SetMinNil(b bool)`

 SetMinNil sets the value for Min to be an explicit nil

### UnsetMin
`func (o *IntegerRulesOptional) UnsetMin()`

UnsetMin ensures that no value is present for Min, not even an explicit nil
### GetMax

`func (o *IntegerRulesOptional) GetMax() int32`

GetMax returns the Max field if non-nil, zero value otherwise.

### GetMaxOk

`func (o *IntegerRulesOptional) GetMaxOk() (*int32, bool)`

GetMaxOk returns a tuple with the Max field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMax

`func (o *IntegerRulesOptional) SetMax(v int32)`

SetMax sets Max field to given value.

### HasMax

`func (o *IntegerRulesOptional) HasMax() bool`

HasMax returns a boolean if a field has been set.

### SetMaxNil

`func (o *IntegerRulesOptional) SetMaxNil(b bool)`

 SetMaxNil sets the value for Max to be an explicit nil

### UnsetMax
`func (o *IntegerRulesOptional) UnsetMax()`

UnsetMax ensures that no value is present for Max, not even an explicit nil
### GetStep

`func (o *IntegerRulesOptional) GetStep() int32`

GetStep returns the Step field if non-nil, zero value otherwise.

### GetStepOk

`func (o *IntegerRulesOptional) GetStepOk() (*int32, bool)`

GetStepOk returns a tuple with the Step field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStep

`func (o *IntegerRulesOptional) SetStep(v int32)`

SetStep sets Step field to given value.

### HasStep

`func (o *IntegerRulesOptional) HasStep() bool`

HasStep returns a boolean if a field has been set.

### SetStepNil

`func (o *IntegerRulesOptional) SetStepNil(b bool)`

 SetStepNil sets the value for Step to be an explicit nil

### UnsetStep
`func (o *IntegerRulesOptional) UnsetStep()`

UnsetStep ensures that no value is present for Step, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


