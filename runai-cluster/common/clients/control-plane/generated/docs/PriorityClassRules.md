# PriorityClassRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory, default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable, default to true | [optional] 
**Options** | Pointer to [**[]PriorityClassOptionsOptionsInner**](PriorityClassOptionsOptionsInner.md) |  | [optional] 

## Methods

### NewPriorityClassRules

`func NewPriorityClassRules() *PriorityClassRules`

NewPriorityClassRules instantiates a new PriorityClassRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPriorityClassRulesWithDefaults

`func NewPriorityClassRulesWithDefaults() *PriorityClassRules`

NewPriorityClassRulesWithDefaults instantiates a new PriorityClassRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *PriorityClassRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *PriorityClassRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *PriorityClassRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *PriorityClassRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *PriorityClassRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *PriorityClassRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *PriorityClassRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *PriorityClassRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *PriorityClassRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *PriorityClassRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *PriorityClassRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *PriorityClassRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *PriorityClassRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *PriorityClassRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *PriorityClassRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *PriorityClassRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *PriorityClassRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *PriorityClassRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetOptions

`func (o *PriorityClassRules) GetOptions() []PriorityClassOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *PriorityClassRules) GetOptionsOk() (*[]PriorityClassOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *PriorityClassRules) SetOptions(v []PriorityClassOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *PriorityClassRules) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *PriorityClassRules) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *PriorityClassRules) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


