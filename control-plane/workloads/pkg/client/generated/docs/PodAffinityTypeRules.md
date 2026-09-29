# PodAffinityTypeRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory, default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable, default to true | [optional] 
**Blocked** | Pointer to **NullableBool** | Blocks the field to prevent workloads from specifying any value for it. | [optional] 
**Options** | Pointer to [**[]PodAffinityTypeOptionsOptionsInner**](PodAffinityTypeOptionsOptionsInner.md) |  | [optional] 

## Methods

### NewPodAffinityTypeRules

`func NewPodAffinityTypeRules() *PodAffinityTypeRules`

NewPodAffinityTypeRules instantiates a new PodAffinityTypeRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPodAffinityTypeRulesWithDefaults

`func NewPodAffinityTypeRulesWithDefaults() *PodAffinityTypeRules`

NewPodAffinityTypeRulesWithDefaults instantiates a new PodAffinityTypeRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *PodAffinityTypeRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *PodAffinityTypeRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *PodAffinityTypeRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *PodAffinityTypeRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *PodAffinityTypeRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *PodAffinityTypeRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *PodAffinityTypeRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *PodAffinityTypeRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *PodAffinityTypeRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *PodAffinityTypeRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *PodAffinityTypeRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *PodAffinityTypeRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *PodAffinityTypeRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *PodAffinityTypeRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *PodAffinityTypeRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *PodAffinityTypeRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *PodAffinityTypeRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *PodAffinityTypeRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetBlocked

`func (o *PodAffinityTypeRules) GetBlocked() bool`

GetBlocked returns the Blocked field if non-nil, zero value otherwise.

### GetBlockedOk

`func (o *PodAffinityTypeRules) GetBlockedOk() (*bool, bool)`

GetBlockedOk returns a tuple with the Blocked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlocked

`func (o *PodAffinityTypeRules) SetBlocked(v bool)`

SetBlocked sets Blocked field to given value.

### HasBlocked

`func (o *PodAffinityTypeRules) HasBlocked() bool`

HasBlocked returns a boolean if a field has been set.

### SetBlockedNil

`func (o *PodAffinityTypeRules) SetBlockedNil(b bool)`

 SetBlockedNil sets the value for Blocked to be an explicit nil

### UnsetBlocked
`func (o *PodAffinityTypeRules) UnsetBlocked()`

UnsetBlocked ensures that no value is present for Blocked, not even an explicit nil
### GetOptions

`func (o *PodAffinityTypeRules) GetOptions() []PodAffinityTypeOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *PodAffinityTypeRules) GetOptionsOk() (*[]PodAffinityTypeOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *PodAffinityTypeRules) SetOptions(v []PodAffinityTypeOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *PodAffinityTypeRules) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *PodAffinityTypeRules) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *PodAffinityTypeRules) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


