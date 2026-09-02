# NimServicePolicy

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyMeta**](PolicyMeta.md) |  | [optional] 
**Policy** | Pointer to [**NullableNimServicePolicyDefaultsAndRules**](NimServicePolicyDefaultsAndRules.md) |  | [optional] 
**Effective** | Pointer to [**NullableNimServicePolicyDefaultsAndRules**](NimServicePolicyDefaultsAndRules.md) |  | [optional] 
**EffectiveUpdatedAt** | Pointer to **time.Time** | The time at which an event occurred. e.g. the time an effective policy has been updated | [optional] 

## Methods

### NewNimServicePolicy

`func NewNimServicePolicy() *NimServicePolicy`

NewNimServicePolicy instantiates a new NimServicePolicy object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServicePolicyWithDefaults

`func NewNimServicePolicyWithDefaults() *NimServicePolicy`

NewNimServicePolicyWithDefaults instantiates a new NimServicePolicy object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *NimServicePolicy) GetMeta() PolicyMeta`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *NimServicePolicy) GetMetaOk() (*PolicyMeta, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *NimServicePolicy) SetMeta(v PolicyMeta)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *NimServicePolicy) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *NimServicePolicy) GetPolicy() NimServicePolicyDefaultsAndRules`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *NimServicePolicy) GetPolicyOk() (*NimServicePolicyDefaultsAndRules, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *NimServicePolicy) SetPolicy(v NimServicePolicyDefaultsAndRules)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *NimServicePolicy) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *NimServicePolicy) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *NimServicePolicy) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil
### GetEffective

`func (o *NimServicePolicy) GetEffective() NimServicePolicyDefaultsAndRules`

GetEffective returns the Effective field if non-nil, zero value otherwise.

### GetEffectiveOk

`func (o *NimServicePolicy) GetEffectiveOk() (*NimServicePolicyDefaultsAndRules, bool)`

GetEffectiveOk returns a tuple with the Effective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffective

`func (o *NimServicePolicy) SetEffective(v NimServicePolicyDefaultsAndRules)`

SetEffective sets Effective field to given value.

### HasEffective

`func (o *NimServicePolicy) HasEffective() bool`

HasEffective returns a boolean if a field has been set.

### SetEffectiveNil

`func (o *NimServicePolicy) SetEffectiveNil(b bool)`

 SetEffectiveNil sets the value for Effective to be an explicit nil

### UnsetEffective
`func (o *NimServicePolicy) UnsetEffective()`

UnsetEffective ensures that no value is present for Effective, not even an explicit nil
### GetEffectiveUpdatedAt

`func (o *NimServicePolicy) GetEffectiveUpdatedAt() time.Time`

GetEffectiveUpdatedAt returns the EffectiveUpdatedAt field if non-nil, zero value otherwise.

### GetEffectiveUpdatedAtOk

`func (o *NimServicePolicy) GetEffectiveUpdatedAtOk() (*time.Time, bool)`

GetEffectiveUpdatedAtOk returns a tuple with the EffectiveUpdatedAt field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffectiveUpdatedAt

`func (o *NimServicePolicy) SetEffectiveUpdatedAt(v time.Time)`

SetEffectiveUpdatedAt sets EffectiveUpdatedAt field to given value.

### HasEffectiveUpdatedAt

`func (o *NimServicePolicy) HasEffectiveUpdatedAt() bool`

HasEffectiveUpdatedAt returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


