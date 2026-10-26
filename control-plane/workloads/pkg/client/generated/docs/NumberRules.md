# NumberRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory. default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable. default to true | [optional] 
**Min** | Pointer to **NullableFloat32** | The minimum value that the field can be assigned to. | [optional] 
**Max** | Pointer to **NullableFloat32** | The maximum value that the field can be assigned to. | [optional] 
**Step** | Pointer to **NullableFloat32** | The minimal difference between two values the field can be assigned to. For example, min&#x3D;2, max&#x3D;3, step&#x3D;0.25 implies that the values the field can hold are 2, 2.25, 2.5 and 3. | [optional] 
**DefaultFrom** | Pointer to [**NullableDefaultFromRule**](DefaultFromRule.md) |  | [optional] 

## Methods

### NewNumberRules

`func NewNumberRules() *NumberRules`

NewNumberRules instantiates a new NumberRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNumberRulesWithDefaults

`func NewNumberRulesWithDefaults() *NumberRules`

NewNumberRulesWithDefaults instantiates a new NumberRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *NumberRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *NumberRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *NumberRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *NumberRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *NumberRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *NumberRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *NumberRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *NumberRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *NumberRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *NumberRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *NumberRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *NumberRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *NumberRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *NumberRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *NumberRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *NumberRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *NumberRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *NumberRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetMin

`func (o *NumberRules) GetMin() float32`

GetMin returns the Min field if non-nil, zero value otherwise.

### GetMinOk

`func (o *NumberRules) GetMinOk() (*float32, bool)`

GetMinOk returns a tuple with the Min field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMin

`func (o *NumberRules) SetMin(v float32)`

SetMin sets Min field to given value.

### HasMin

`func (o *NumberRules) HasMin() bool`

HasMin returns a boolean if a field has been set.

### SetMinNil

`func (o *NumberRules) SetMinNil(b bool)`

 SetMinNil sets the value for Min to be an explicit nil

### UnsetMin
`func (o *NumberRules) UnsetMin()`

UnsetMin ensures that no value is present for Min, not even an explicit nil
### GetMax

`func (o *NumberRules) GetMax() float32`

GetMax returns the Max field if non-nil, zero value otherwise.

### GetMaxOk

`func (o *NumberRules) GetMaxOk() (*float32, bool)`

GetMaxOk returns a tuple with the Max field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMax

`func (o *NumberRules) SetMax(v float32)`

SetMax sets Max field to given value.

### HasMax

`func (o *NumberRules) HasMax() bool`

HasMax returns a boolean if a field has been set.

### SetMaxNil

`func (o *NumberRules) SetMaxNil(b bool)`

 SetMaxNil sets the value for Max to be an explicit nil

### UnsetMax
`func (o *NumberRules) UnsetMax()`

UnsetMax ensures that no value is present for Max, not even an explicit nil
### GetStep

`func (o *NumberRules) GetStep() float32`

GetStep returns the Step field if non-nil, zero value otherwise.

### GetStepOk

`func (o *NumberRules) GetStepOk() (*float32, bool)`

GetStepOk returns a tuple with the Step field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStep

`func (o *NumberRules) SetStep(v float32)`

SetStep sets Step field to given value.

### HasStep

`func (o *NumberRules) HasStep() bool`

HasStep returns a boolean if a field has been set.

### SetStepNil

`func (o *NumberRules) SetStepNil(b bool)`

 SetStepNil sets the value for Step to be an explicit nil

### UnsetStep
`func (o *NumberRules) UnsetStep()`

UnsetStep ensures that no value is present for Step, not even an explicit nil
### GetDefaultFrom

`func (o *NumberRules) GetDefaultFrom() DefaultFromRule`

GetDefaultFrom returns the DefaultFrom field if non-nil, zero value otherwise.

### GetDefaultFromOk

`func (o *NumberRules) GetDefaultFromOk() (*DefaultFromRule, bool)`

GetDefaultFromOk returns a tuple with the DefaultFrom field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDefaultFrom

`func (o *NumberRules) SetDefaultFrom(v DefaultFromRule)`

SetDefaultFrom sets DefaultFrom field to given value.

### HasDefaultFrom

`func (o *NumberRules) HasDefaultFrom() bool`

HasDefaultFrom returns a boolean if a field has been set.

### SetDefaultFromNil

`func (o *NumberRules) SetDefaultFromNil(b bool)`

 SetDefaultFromNil sets the value for DefaultFrom to be an explicit nil

### UnsetDefaultFrom
`func (o *NumberRules) UnsetDefaultFrom()`

UnsetDefaultFrom ensures that no value is present for DefaultFrom, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


