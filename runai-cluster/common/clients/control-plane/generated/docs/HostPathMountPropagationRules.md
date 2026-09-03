# HostPathMountPropagationRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory, default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable, default to true | [optional] 
**Options** | Pointer to [**[]HostPathMountPropagationOptionsOptionsInner**](HostPathMountPropagationOptionsOptionsInner.md) |  | [optional] 

## Methods

### NewHostPathMountPropagationRules

`func NewHostPathMountPropagationRules() *HostPathMountPropagationRules`

NewHostPathMountPropagationRules instantiates a new HostPathMountPropagationRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewHostPathMountPropagationRulesWithDefaults

`func NewHostPathMountPropagationRulesWithDefaults() *HostPathMountPropagationRules`

NewHostPathMountPropagationRulesWithDefaults instantiates a new HostPathMountPropagationRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *HostPathMountPropagationRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *HostPathMountPropagationRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *HostPathMountPropagationRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *HostPathMountPropagationRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *HostPathMountPropagationRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *HostPathMountPropagationRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *HostPathMountPropagationRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *HostPathMountPropagationRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *HostPathMountPropagationRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *HostPathMountPropagationRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *HostPathMountPropagationRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *HostPathMountPropagationRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *HostPathMountPropagationRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *HostPathMountPropagationRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *HostPathMountPropagationRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *HostPathMountPropagationRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *HostPathMountPropagationRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *HostPathMountPropagationRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetOptions

`func (o *HostPathMountPropagationRules) GetOptions() []HostPathMountPropagationOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *HostPathMountPropagationRules) GetOptionsOk() (*[]HostPathMountPropagationOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *HostPathMountPropagationRules) SetOptions(v []HostPathMountPropagationOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *HostPathMountPropagationRules) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *HostPathMountPropagationRules) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *HostPathMountPropagationRules) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


