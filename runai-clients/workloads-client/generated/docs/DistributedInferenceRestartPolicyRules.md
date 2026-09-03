# DistributedInferenceRestartPolicyRules

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory, default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable, default to true | [optional] 
**Blocked** | Pointer to **NullableBool** | Blocks the field to prevent workloads from specifying any value for it. | [optional] 
**Options** | Pointer to [**[]DistributedInferenceRestartPolicyOptionsOptionsInner**](DistributedInferenceRestartPolicyOptionsOptionsInner.md) | Limit the permitted values for restartPolicy to a given set of values. | [optional] 

## Methods

### NewDistributedInferenceRestartPolicyRules

`func NewDistributedInferenceRestartPolicyRules() *DistributedInferenceRestartPolicyRules`

NewDistributedInferenceRestartPolicyRules instantiates a new DistributedInferenceRestartPolicyRules object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceRestartPolicyRulesWithDefaults

`func NewDistributedInferenceRestartPolicyRulesWithDefaults() *DistributedInferenceRestartPolicyRules`

NewDistributedInferenceRestartPolicyRulesWithDefaults instantiates a new DistributedInferenceRestartPolicyRules object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *DistributedInferenceRestartPolicyRules) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *DistributedInferenceRestartPolicyRules) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *DistributedInferenceRestartPolicyRules) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *DistributedInferenceRestartPolicyRules) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *DistributedInferenceRestartPolicyRules) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *DistributedInferenceRestartPolicyRules) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *DistributedInferenceRestartPolicyRules) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *DistributedInferenceRestartPolicyRules) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *DistributedInferenceRestartPolicyRules) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *DistributedInferenceRestartPolicyRules) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *DistributedInferenceRestartPolicyRules) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *DistributedInferenceRestartPolicyRules) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *DistributedInferenceRestartPolicyRules) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *DistributedInferenceRestartPolicyRules) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *DistributedInferenceRestartPolicyRules) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *DistributedInferenceRestartPolicyRules) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *DistributedInferenceRestartPolicyRules) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *DistributedInferenceRestartPolicyRules) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetBlocked

`func (o *DistributedInferenceRestartPolicyRules) GetBlocked() bool`

GetBlocked returns the Blocked field if non-nil, zero value otherwise.

### GetBlockedOk

`func (o *DistributedInferenceRestartPolicyRules) GetBlockedOk() (*bool, bool)`

GetBlockedOk returns a tuple with the Blocked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlocked

`func (o *DistributedInferenceRestartPolicyRules) SetBlocked(v bool)`

SetBlocked sets Blocked field to given value.

### HasBlocked

`func (o *DistributedInferenceRestartPolicyRules) HasBlocked() bool`

HasBlocked returns a boolean if a field has been set.

### SetBlockedNil

`func (o *DistributedInferenceRestartPolicyRules) SetBlockedNil(b bool)`

 SetBlockedNil sets the value for Blocked to be an explicit nil

### UnsetBlocked
`func (o *DistributedInferenceRestartPolicyRules) UnsetBlocked()`

UnsetBlocked ensures that no value is present for Blocked, not even an explicit nil
### GetOptions

`func (o *DistributedInferenceRestartPolicyRules) GetOptions() []DistributedInferenceRestartPolicyOptionsOptionsInner`

GetOptions returns the Options field if non-nil, zero value otherwise.

### GetOptionsOk

`func (o *DistributedInferenceRestartPolicyRules) GetOptionsOk() (*[]DistributedInferenceRestartPolicyOptionsOptionsInner, bool)`

GetOptionsOk returns a tuple with the Options field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetOptions

`func (o *DistributedInferenceRestartPolicyRules) SetOptions(v []DistributedInferenceRestartPolicyOptionsOptionsInner)`

SetOptions sets Options field to given value.

### HasOptions

`func (o *DistributedInferenceRestartPolicyRules) HasOptions() bool`

HasOptions returns a boolean if a field has been set.

### SetOptionsNil

`func (o *DistributedInferenceRestartPolicyRules) SetOptionsNil(b bool)`

 SetOptionsNil sets the value for Options to be an explicit nil

### UnsetOptions
`func (o *DistributedInferenceRestartPolicyRules) UnsetOptions()`

UnsetOptions ensures that no value is present for Options, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


