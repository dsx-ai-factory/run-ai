# DistributedInferenceStartupPolicyFieldRulesStartupPolicy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**SourceOfRule** | Pointer to [**NullableSourceOfRule**](SourceOfRule.md) |  | [optional] 
**Required** | Pointer to **NullableBool** | Whether the field is mandatory, default to false. | [optional] 
**CanEdit** | Pointer to **NullableBool** | Whether the value of the field is editable, default to true | [optional] 
**Blocked** | Pointer to **NullableBool** | Blocks the field to prevent workloads from specifying any value for it. | [optional] 
**StartupPolicy** | Pointer to [**NullableDistributedInferenceStartupPolicy**](DistributedInferenceStartupPolicy.md) |  | [optional] [default to DISTRIBUTEDINFERENCESTARTUPPOLICY_LEADER_CREATED]

## Methods

### NewDistributedInferenceStartupPolicyFieldRulesStartupPolicy

`func NewDistributedInferenceStartupPolicyFieldRulesStartupPolicy() *DistributedInferenceStartupPolicyFieldRulesStartupPolicy`

NewDistributedInferenceStartupPolicyFieldRulesStartupPolicy instantiates a new DistributedInferenceStartupPolicyFieldRulesStartupPolicy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedInferenceStartupPolicyFieldRulesStartupPolicyWithDefaults

`func NewDistributedInferenceStartupPolicyFieldRulesStartupPolicyWithDefaults() *DistributedInferenceStartupPolicyFieldRulesStartupPolicy`

NewDistributedInferenceStartupPolicyFieldRulesStartupPolicyWithDefaults instantiates a new DistributedInferenceStartupPolicyFieldRulesStartupPolicy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetSourceOfRule

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetSourceOfRule() SourceOfRule`

GetSourceOfRule returns the SourceOfRule field if non-nil, zero value otherwise.

### GetSourceOfRuleOk

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetSourceOfRuleOk() (*SourceOfRule, bool)`

GetSourceOfRuleOk returns a tuple with the SourceOfRule field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetSourceOfRule

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetSourceOfRule(v SourceOfRule)`

SetSourceOfRule sets SourceOfRule field to given value.

### HasSourceOfRule

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) HasSourceOfRule() bool`

HasSourceOfRule returns a boolean if a field has been set.

### SetSourceOfRuleNil

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetSourceOfRuleNil(b bool)`

 SetSourceOfRuleNil sets the value for SourceOfRule to be an explicit nil

### UnsetSourceOfRule
`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) UnsetSourceOfRule()`

UnsetSourceOfRule ensures that no value is present for SourceOfRule, not even an explicit nil
### GetRequired

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetRequired() bool`

GetRequired returns the Required field if non-nil, zero value otherwise.

### GetRequiredOk

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetRequiredOk() (*bool, bool)`

GetRequiredOk returns a tuple with the Required field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetRequired

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetRequired(v bool)`

SetRequired sets Required field to given value.

### HasRequired

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) HasRequired() bool`

HasRequired returns a boolean if a field has been set.

### SetRequiredNil

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetRequiredNil(b bool)`

 SetRequiredNil sets the value for Required to be an explicit nil

### UnsetRequired
`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) UnsetRequired()`

UnsetRequired ensures that no value is present for Required, not even an explicit nil
### GetCanEdit

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetCanEdit() bool`

GetCanEdit returns the CanEdit field if non-nil, zero value otherwise.

### GetCanEditOk

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetCanEditOk() (*bool, bool)`

GetCanEditOk returns a tuple with the CanEdit field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetCanEdit

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetCanEdit(v bool)`

SetCanEdit sets CanEdit field to given value.

### HasCanEdit

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) HasCanEdit() bool`

HasCanEdit returns a boolean if a field has been set.

### SetCanEditNil

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetCanEditNil(b bool)`

 SetCanEditNil sets the value for CanEdit to be an explicit nil

### UnsetCanEdit
`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) UnsetCanEdit()`

UnsetCanEdit ensures that no value is present for CanEdit, not even an explicit nil
### GetBlocked

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetBlocked() bool`

GetBlocked returns the Blocked field if non-nil, zero value otherwise.

### GetBlockedOk

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetBlockedOk() (*bool, bool)`

GetBlockedOk returns a tuple with the Blocked field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetBlocked

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetBlocked(v bool)`

SetBlocked sets Blocked field to given value.

### HasBlocked

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) HasBlocked() bool`

HasBlocked returns a boolean if a field has been set.

### SetBlockedNil

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetBlockedNil(b bool)`

 SetBlockedNil sets the value for Blocked to be an explicit nil

### UnsetBlocked
`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) UnsetBlocked()`

UnsetBlocked ensures that no value is present for Blocked, not even an explicit nil
### GetStartupPolicy

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetStartupPolicy() DistributedInferenceStartupPolicy`

GetStartupPolicy returns the StartupPolicy field if non-nil, zero value otherwise.

### GetStartupPolicyOk

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) GetStartupPolicyOk() (*DistributedInferenceStartupPolicy, bool)`

GetStartupPolicyOk returns a tuple with the StartupPolicy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetStartupPolicy

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetStartupPolicy(v DistributedInferenceStartupPolicy)`

SetStartupPolicy sets StartupPolicy field to given value.

### HasStartupPolicy

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) HasStartupPolicy() bool`

HasStartupPolicy returns a boolean if a field has been set.

### SetStartupPolicyNil

`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) SetStartupPolicyNil(b bool)`

 SetStartupPolicyNil sets the value for StartupPolicy to be an explicit nil

### UnsetStartupPolicy
`func (o *DistributedInferenceStartupPolicyFieldRulesStartupPolicy) UnsetStartupPolicy()`

UnsetStartupPolicy ensures that no value is present for StartupPolicy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


