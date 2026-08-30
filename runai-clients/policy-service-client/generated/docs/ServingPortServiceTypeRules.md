# ServingPortServiceTypeRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory, default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable, default to true | [optional] 
**Blocked** | Pointer to **NullableBool** | Blocks the field to prevent workloads from specifying any value for it. | [optional] 
**Options** | Pointer to [**[]ServingPortServiceTypeOptionsOptionsInner**](ServingPortServiceTypeOptionsOptionsInner.md) |  | [optional] 

## Methods

### NewServingPortServiceTypeRules

`func NewServingPortServiceTypeRules() *ServingPortServiceTypeRules`

NewServingPortServiceTypeRules instantiates a new ServingPortServiceTypeRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServingPortServiceTypeRulesWithDefaults

`func NewServingPortServiceTypeRulesWithDefaults() *ServingPortServiceTypeRules`

NewServingPortServiceTypeRulesWithDefaults instantiates a new ServingPortServiceTypeRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *ServingPortServiceTypeRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *ServingPortServiceTypeRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *ServingPortServiceTypeRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *ServingPortServiceTypeRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *ServingPortServiceTypeRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *ServingPortServiceTypeRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *ServingPortServiceTypeRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *ServingPortServiceTypeRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *ServingPortServiceTypeRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *ServingPortServiceTypeRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *ServingPortServiceTypeRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *ServingPortServiceTypeRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *ServingPortServiceTypeRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *ServingPortServiceTypeRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *ServingPortServiceTypeRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *ServingPortServiceTypeRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *ServingPortServiceTypeRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *ServingPortServiceTypeRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetBlocked

`func (o *ServingPortServiceTypeRules) GetBlocked() bool`

GetBlocked returns the Blocked field if non-nil, zero value otherwise.

### GetBlockedOk

`func (o *ServingPortServiceTypeRules) GetBlockedOk() (*bool, bool)`

GetBlockedOk returns a tuple with the Blocked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlocked

`func (o *ServingPortServiceTypeRules) SetBlocked(v bool)`

SetBlocked sets Blocked field to given value.

### HasBlocked

`func (o *ServingPortServiceTypeRules) HasBlocked() bool`

HasBlocked returns a boolean if a field has been set.

### SetBlockedNil

`func (o *ServingPortServiceTypeRules) SetBlockedNil(b bool)`

 SetBlockedNil sets the value for Blocked to be an explicit nil

### UnsetBlocked
`func (o *ServingPortServiceTypeRules) UnsetBlocked()`

UnsetBlocked ensures that no value is present for Blocked, not even an explicit nil
### GetOptions

`func (o *ServingPortServiceTypeRules) GetOptions() []ServingPortServiceTypeOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *ServingPortServiceTypeRules) GetOptionsOk() (*[]ServingPortServiceTypeOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *ServingPortServiceTypeRules) SetOptions(v []ServingPortServiceTypeOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *ServingPortServiceTypeRules) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *ServingPortServiceTypeRules) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *ServingPortServiceTypeRules) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


