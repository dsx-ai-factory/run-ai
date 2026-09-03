# WorkspacePolicyV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyMeta**](PolicyMeta.md) |  | [optional] 
**Policy** | Pointer to [**NullableWorkspacePolicyDefaultsAndRulesV2**](WorkspacePolicyDefaultsAndRulesV2.md) |  | [optional] 
**Effective** | Pointer to [**NullableWorkspacePolicyDefaultsAndRulesV2**](WorkspacePolicyDefaultsAndRulesV2.md) |  | [optional] 
**EffectiveUpdatedAt** | Pointer to **time.Time** | The time at which an event occurred. e.g. the time an effective policy has been updated | [optional] 

## Methods

### NewWorkspacePolicyV2

`func NewWorkspacePolicyV2() *WorkspacePolicyV2`

NewWorkspacePolicyV2 instantiates a new WorkspacePolicyV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewWorkspacePolicyV2WithDefaults

`func NewWorkspacePolicyV2WithDefaults() *WorkspacePolicyV2`

NewWorkspacePolicyV2WithDefaults instantiates a new WorkspacePolicyV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *WorkspacePolicyV2) GetMeta() PolicyMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *WorkspacePolicyV2) GetMetaOk() (*PolicyMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *WorkspacePolicyV2) SetMeta(v PolicyMeta)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *WorkspacePolicyV2) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *WorkspacePolicyV2) GetPolicy() WorkspacePolicyDefaultsAndRulesV2`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *WorkspacePolicyV2) GetPolicyOk() (*WorkspacePolicyDefaultsAndRulesV2, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *WorkspacePolicyV2) SetPolicy(v WorkspacePolicyDefaultsAndRulesV2)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *WorkspacePolicyV2) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *WorkspacePolicyV2) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *WorkspacePolicyV2) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil
### GetEffective

`func (o *WorkspacePolicyV2) GetEffective() WorkspacePolicyDefaultsAndRulesV2`

GetEffective returns the Effective field if non-nil, zero value otherwise.

### GetEffectiveOk

`func (o *WorkspacePolicyV2) GetEffectiveOk() (*WorkspacePolicyDefaultsAndRulesV2, bool)`

GetEffectiveOk returns a tuple with the Effective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffective

`func (o *WorkspacePolicyV2) SetEffective(v WorkspacePolicyDefaultsAndRulesV2)`

SetEffective sets Effective field to given value.

### HasEffective

`func (o *WorkspacePolicyV2) HasEffective() bool`

HasEffective returns a boolean if a field has been set.

### SetEffectiveNil

`func (o *WorkspacePolicyV2) SetEffectiveNil(b bool)`

 SetEffectiveNil sets the value for Effective to be an explicit nil

### UnsetEffective
`func (o *WorkspacePolicyV2) UnsetEffective()`

UnsetEffective ensures that no value is present for Effective, not even an explicit nil
### GetEffectiveUpdatedAt

`func (o *WorkspacePolicyV2) GetEffectiveUpdatedAt() time.Time`

GetEffectiveUpdatedAt returns the EffectiveUpdatedAt field if non-nil, zero value otherwise.

### GetEffectiveUpdatedAtOk

`func (o *WorkspacePolicyV2) GetEffectiveUpdatedAtOk() (*time.Time, bool)`

GetEffectiveUpdatedAtOk returns a tuple with the EffectiveUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveUpdatedAt

`func (o *WorkspacePolicyV2) SetEffectiveUpdatedAt(v time.Time)`

SetEffectiveUpdatedAt sets EffectiveUpdatedAt field to given value.

### HasEffectiveUpdatedAt

`func (o *WorkspacePolicyV2) HasEffectiveUpdatedAt() bool`

HasEffectiveUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


