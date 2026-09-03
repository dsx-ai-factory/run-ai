# InferencePolicyV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyMeta**](PolicyMeta.md) |  | [optional] 
**Policy** | Pointer to [**NullableInferencePolicyDefaultsAndRulesV2**](InferencePolicyDefaultsAndRulesV2.md) |  | [optional] 
**Effective** | Pointer to [**NullableInferencePolicyDefaultsAndRulesV2**](InferencePolicyDefaultsAndRulesV2.md) |  | [optional] 
**EffectiveUpdatedAt** | Pointer to **time.Time** | The time at which an event occurred. e.g. the time an effective policy has been updated | [optional] 

## Methods

### NewInferencePolicyV2

`func NewInferencePolicyV2() *InferencePolicyV2`

NewInferencePolicyV2 instantiates a new InferencePolicyV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencePolicyV2WithDefaults

`func NewInferencePolicyV2WithDefaults() *InferencePolicyV2`

NewInferencePolicyV2WithDefaults instantiates a new InferencePolicyV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *InferencePolicyV2) GetMeta() PolicyMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *InferencePolicyV2) GetMetaOk() (*PolicyMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *InferencePolicyV2) SetMeta(v PolicyMeta)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *InferencePolicyV2) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *InferencePolicyV2) GetPolicy() InferencePolicyDefaultsAndRulesV2`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *InferencePolicyV2) GetPolicyOk() (*InferencePolicyDefaultsAndRulesV2, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *InferencePolicyV2) SetPolicy(v InferencePolicyDefaultsAndRulesV2)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *InferencePolicyV2) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *InferencePolicyV2) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *InferencePolicyV2) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil
### GetEffective

`func (o *InferencePolicyV2) GetEffective() InferencePolicyDefaultsAndRulesV2`

GetEffective returns the Effective field if non-nil, zero value otherwise.

### GetEffectiveOk

`func (o *InferencePolicyV2) GetEffectiveOk() (*InferencePolicyDefaultsAndRulesV2, bool)`

GetEffectiveOk returns a tuple with the Effective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffective

`func (o *InferencePolicyV2) SetEffective(v InferencePolicyDefaultsAndRulesV2)`

SetEffective sets Effective field to given value.

### HasEffective

`func (o *InferencePolicyV2) HasEffective() bool`

HasEffective returns a boolean if a field has been set.

### SetEffectiveNil

`func (o *InferencePolicyV2) SetEffectiveNil(b bool)`

 SetEffectiveNil sets the value for Effective to be an explicit nil

### UnsetEffective
`func (o *InferencePolicyV2) UnsetEffective()`

UnsetEffective ensures that no value is present for Effective, not even an explicit nil
### GetEffectiveUpdatedAt

`func (o *InferencePolicyV2) GetEffectiveUpdatedAt() time.Time`

GetEffectiveUpdatedAt returns the EffectiveUpdatedAt field if non-nil, zero value otherwise.

### GetEffectiveUpdatedAtOk

`func (o *InferencePolicyV2) GetEffectiveUpdatedAtOk() (*time.Time, bool)`

GetEffectiveUpdatedAtOk returns a tuple with the EffectiveUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveUpdatedAt

`func (o *InferencePolicyV2) SetEffectiveUpdatedAt(v time.Time)`

SetEffectiveUpdatedAt sets EffectiveUpdatedAt field to given value.

### HasEffectiveUpdatedAt

`func (o *InferencePolicyV2) HasEffectiveUpdatedAt() bool`

HasEffectiveUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


