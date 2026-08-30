# NimServiceServingPortAuthorizationTypeRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory, default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable, default to true | [optional] 
**Blocked** | Pointer to **NullableBool** | Blocks the field to prevent workloads from specifying any value for it. | [optional] 
**Options** | Pointer to [**[]NimServiceServingPortAuthorizationTypeOptionsOptionsInner**](NimServiceServingPortAuthorizationTypeOptionsOptionsInner.md) |  | [optional] 

## Methods

### NewNimServiceServingPortAuthorizationTypeRules

`func NewNimServiceServingPortAuthorizationTypeRules() *NimServiceServingPortAuthorizationTypeRules`

NewNimServiceServingPortAuthorizationTypeRules instantiates a new NimServiceServingPortAuthorizationTypeRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServiceServingPortAuthorizationTypeRulesWithDefaults

`func NewNimServiceServingPortAuthorizationTypeRulesWithDefaults() *NimServiceServingPortAuthorizationTypeRules`

NewNimServiceServingPortAuthorizationTypeRulesWithDefaults instantiates a new NimServiceServingPortAuthorizationTypeRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *NimServiceServingPortAuthorizationTypeRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *NimServiceServingPortAuthorizationTypeRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *NimServiceServingPortAuthorizationTypeRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *NimServiceServingPortAuthorizationTypeRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *NimServiceServingPortAuthorizationTypeRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *NimServiceServingPortAuthorizationTypeRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *NimServiceServingPortAuthorizationTypeRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *NimServiceServingPortAuthorizationTypeRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *NimServiceServingPortAuthorizationTypeRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *NimServiceServingPortAuthorizationTypeRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *NimServiceServingPortAuthorizationTypeRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *NimServiceServingPortAuthorizationTypeRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *NimServiceServingPortAuthorizationTypeRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *NimServiceServingPortAuthorizationTypeRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *NimServiceServingPortAuthorizationTypeRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *NimServiceServingPortAuthorizationTypeRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *NimServiceServingPortAuthorizationTypeRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *NimServiceServingPortAuthorizationTypeRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetBlocked

`func (o *NimServiceServingPortAuthorizationTypeRules) GetBlocked() bool`

GetBlocked returns the Blocked field if non-nil, zero value otherwise.

### GetBlockedOk

`func (o *NimServiceServingPortAuthorizationTypeRules) GetBlockedOk() (*bool, bool)`

GetBlockedOk returns a tuple with the Blocked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlocked

`func (o *NimServiceServingPortAuthorizationTypeRules) SetBlocked(v bool)`

SetBlocked sets Blocked field to given value.

### HasBlocked

`func (o *NimServiceServingPortAuthorizationTypeRules) HasBlocked() bool`

HasBlocked returns a boolean if a field has been set.

### SetBlockedNil

`func (o *NimServiceServingPortAuthorizationTypeRules) SetBlockedNil(b bool)`

 SetBlockedNil sets the value for Blocked to be an explicit nil

### UnsetBlocked
`func (o *NimServiceServingPortAuthorizationTypeRules) UnsetBlocked()`

UnsetBlocked ensures that no value is present for Blocked, not even an explicit nil
### GetOptions

`func (o *NimServiceServingPortAuthorizationTypeRules) GetOptions() []NimServiceServingPortAuthorizationTypeOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *NimServiceServingPortAuthorizationTypeRules) GetOptionsOk() (*[]NimServiceServingPortAuthorizationTypeOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *NimServiceServingPortAuthorizationTypeRules) SetOptions(v []NimServiceServingPortAuthorizationTypeOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *NimServiceServingPortAuthorizationTypeRules) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *NimServiceServingPortAuthorizationTypeRules) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *NimServiceServingPortAuthorizationTypeRules) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


