# FallbackPolicy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyMeta**](PolicyMeta.md) |  | [optional] 
**Policy** | Pointer to [**NullableFallbackPolicyDefaultsAndRules**](FallbackPolicyDefaultsAndRules.md) |  | [optional] 
**Effective** | Pointer to [**NullableFallbackPolicyDefaultsAndRules**](FallbackPolicyDefaultsAndRules.md) |  | [optional] 
**EffectiveUpdatedAt** | Pointer to **time.Time** | The time at which an event occurred. e.g. the time an effective policy has been updated | [optional] 

## Methods

### NewFallbackPolicy

`func NewFallbackPolicy() *FallbackPolicy`

NewFallbackPolicy instantiates a new FallbackPolicy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFallbackPolicyWithDefaults

`func NewFallbackPolicyWithDefaults() *FallbackPolicy`

NewFallbackPolicyWithDefaults instantiates a new FallbackPolicy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *FallbackPolicy) GetMeta() PolicyMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *FallbackPolicy) GetMetaOk() (*PolicyMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *FallbackPolicy) SetMeta(v PolicyMeta)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *FallbackPolicy) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *FallbackPolicy) GetPolicy() FallbackPolicyDefaultsAndRules`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *FallbackPolicy) GetPolicyOk() (*FallbackPolicyDefaultsAndRules, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *FallbackPolicy) SetPolicy(v FallbackPolicyDefaultsAndRules)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *FallbackPolicy) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *FallbackPolicy) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *FallbackPolicy) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil
### GetEffective

`func (o *FallbackPolicy) GetEffective() FallbackPolicyDefaultsAndRules`

GetEffective returns the Effective field if non-nil, zero value otherwise.

### GetEffectiveOk

`func (o *FallbackPolicy) GetEffectiveOk() (*FallbackPolicyDefaultsAndRules, bool)`

GetEffectiveOk returns a tuple with the Effective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffective

`func (o *FallbackPolicy) SetEffective(v FallbackPolicyDefaultsAndRules)`

SetEffective sets Effective field to given value.

### HasEffective

`func (o *FallbackPolicy) HasEffective() bool`

HasEffective returns a boolean if a field has been set.

### SetEffectiveNil

`func (o *FallbackPolicy) SetEffectiveNil(b bool)`

 SetEffectiveNil sets the value for Effective to be an explicit nil

### UnsetEffective
`func (o *FallbackPolicy) UnsetEffective()`

UnsetEffective ensures that no value is present for Effective, not even an explicit nil
### GetEffectiveUpdatedAt

`func (o *FallbackPolicy) GetEffectiveUpdatedAt() time.Time`

GetEffectiveUpdatedAt returns the EffectiveUpdatedAt field if non-nil, zero value otherwise.

### GetEffectiveUpdatedAtOk

`func (o *FallbackPolicy) GetEffectiveUpdatedAtOk() (*time.Time, bool)`

GetEffectiveUpdatedAtOk returns a tuple with the EffectiveUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveUpdatedAt

`func (o *FallbackPolicy) SetEffectiveUpdatedAt(v time.Time)`

SetEffectiveUpdatedAt sets EffectiveUpdatedAt field to given value.

### HasEffectiveUpdatedAt

`func (o *FallbackPolicy) HasEffectiveUpdatedAt() bool`

HasEffectiveUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


