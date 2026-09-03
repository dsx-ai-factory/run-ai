# ServingPortProtocolRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory, default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable, default to true | [optional] 
**Blocked** | Pointer to **NullableBool** | Blocks the field to prevent workloads from specifying any value for it. | [optional] 
**Options** | Pointer to [**[]ServingPortProtocolOptionsOptionsInner**](ServingPortProtocolOptionsOptionsInner.md) |  | [optional] 

## Methods

### NewServingPortProtocolRules

`func NewServingPortProtocolRules() *ServingPortProtocolRules`

NewServingPortProtocolRules instantiates a new ServingPortProtocolRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewServingPortProtocolRulesWithDefaults

`func NewServingPortProtocolRulesWithDefaults() *ServingPortProtocolRules`

NewServingPortProtocolRulesWithDefaults instantiates a new ServingPortProtocolRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *ServingPortProtocolRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *ServingPortProtocolRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *ServingPortProtocolRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *ServingPortProtocolRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *ServingPortProtocolRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *ServingPortProtocolRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *ServingPortProtocolRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *ServingPortProtocolRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *ServingPortProtocolRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *ServingPortProtocolRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *ServingPortProtocolRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *ServingPortProtocolRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *ServingPortProtocolRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *ServingPortProtocolRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *ServingPortProtocolRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *ServingPortProtocolRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *ServingPortProtocolRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *ServingPortProtocolRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetBlocked

`func (o *ServingPortProtocolRules) GetBlocked() bool`

GetBlocked returns the Blocked field if non-nil, zero value otherwise.

### GetBlockedOk

`func (o *ServingPortProtocolRules) GetBlockedOk() (*bool, bool)`

GetBlockedOk returns a tuple with the Blocked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlocked

`func (o *ServingPortProtocolRules) SetBlocked(v bool)`

SetBlocked sets Blocked field to given value.

### HasBlocked

`func (o *ServingPortProtocolRules) HasBlocked() bool`

HasBlocked returns a boolean if a field has been set.

### SetBlockedNil

`func (o *ServingPortProtocolRules) SetBlockedNil(b bool)`

 SetBlockedNil sets the value for Blocked to be an explicit nil

### UnsetBlocked
`func (o *ServingPortProtocolRules) UnsetBlocked()`

UnsetBlocked ensures that no value is present for Blocked, not even an explicit nil
### GetOptions

`func (o *ServingPortProtocolRules) GetOptions() []ServingPortProtocolOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *ServingPortProtocolRules) GetOptionsOk() (*[]ServingPortProtocolOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *ServingPortProtocolRules) SetOptions(v []ServingPortProtocolOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *ServingPortProtocolRules) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *ServingPortProtocolRules) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *ServingPortProtocolRules) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


