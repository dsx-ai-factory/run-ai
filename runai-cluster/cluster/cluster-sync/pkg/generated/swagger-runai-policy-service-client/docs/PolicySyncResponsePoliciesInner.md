# PolicySyncResponsePoliciesInner

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | [**PolicyCreationFields**](PolicyCreationFields.md) |  | 
**Effective** | Pointer to [**NullableSupersetPolicy**](SupersetPolicy.md) |  | [optional] 
**Distributed** | Pointer to [**NullableDistributedPolicyDefaultsAndRulesV2**](DistributedPolicyDefaultsAndRulesV2.md) |  | [optional] 
**IsDeleted** | Pointer to **NullableBool** |  | [optional] 
**UpdateCounter** | Pointer to **int64** | number which uniquely identifies each update. | [optional] 

## Methods

### NewPolicySyncResponsePoliciesInner

`func NewPolicySyncResponsePoliciesInner(meta PolicyCreationFields, ) *PolicySyncResponsePoliciesInner`

NewPolicySyncResponsePoliciesInner instantiates a new PolicySyncResponsePoliciesInner object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPolicySyncResponsePoliciesInnerWithDefaults

`func NewPolicySyncResponsePoliciesInnerWithDefaults() *PolicySyncResponsePoliciesInner`

NewPolicySyncResponsePoliciesInnerWithDefaults instantiates a new PolicySyncResponsePoliciesInner object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *PolicySyncResponsePoliciesInner) GetMeta() PolicyCreationFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *PolicySyncResponsePoliciesInner) GetMetaOk() (*PolicyCreationFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *PolicySyncResponsePoliciesInner) SetMeta(v PolicyCreationFields)`

SetMeta sets Meta field to given value.


### GetEffective

`func (o *PolicySyncResponsePoliciesInner) GetEffective() SupersetPolicy`

GetEffective returns the Effective field if non-nil, zero value otherwise.

### GetEffectiveOk

`func (o *PolicySyncResponsePoliciesInner) GetEffectiveOk() (*SupersetPolicy, bool)`

GetEffectiveOk returns a tuple with the Effective field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetEffective

`func (o *PolicySyncResponsePoliciesInner) SetEffective(v SupersetPolicy)`

SetEffective sets Effective field to given value.

### HasEffective

`func (o *PolicySyncResponsePoliciesInner) HasEffective() bool`

HasEffective returns a boolean if a field has been set.

### SetEffectiveNil

`func (o *PolicySyncResponsePoliciesInner) SetEffectiveNil(b bool)`

 SetEffectiveNil sets the value for Effective to be an explicit nil

### UnsetEffective
`func (o *PolicySyncResponsePoliciesInner) UnsetEffective()`

UnsetEffective ensures that no value is present for Effective, not even an explicit nil
### GetDistributed

`func (o *PolicySyncResponsePoliciesInner) GetDistributed() DistributedPolicyDefaultsAndRulesV2`

GetDistributed returns the Distributed field if non-nil, zero value otherwise.

### GetDistributedOk

`func (o *PolicySyncResponsePoliciesInner) GetDistributedOk() (*DistributedPolicyDefaultsAndRulesV2, bool)`

GetDistributedOk returns a tuple with the Distributed field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetDistributed

`func (o *PolicySyncResponsePoliciesInner) SetDistributed(v DistributedPolicyDefaultsAndRulesV2)`

SetDistributed sets Distributed field to given value.

### HasDistributed

`func (o *PolicySyncResponsePoliciesInner) HasDistributed() bool`

HasDistributed returns a boolean if a field has been set.

### SetDistributedNil

`func (o *PolicySyncResponsePoliciesInner) SetDistributedNil(b bool)`

 SetDistributedNil sets the value for Distributed to be an explicit nil

### UnsetDistributed
`func (o *PolicySyncResponsePoliciesInner) UnsetDistributed()`

UnsetDistributed ensures that no value is present for Distributed, not even an explicit nil
### GetIsDeleted

`func (o *PolicySyncResponsePoliciesInner) GetIsDeleted() bool`

GetIsDeleted returns the IsDeleted field if non-nil, zero value otherwise.

### GetIsDeletedOk

`func (o *PolicySyncResponsePoliciesInner) GetIsDeletedOk() (*bool, bool)`

GetIsDeletedOk returns a tuple with the IsDeleted field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetIsDeleted

`func (o *PolicySyncResponsePoliciesInner) SetIsDeleted(v bool)`

SetIsDeleted sets IsDeleted field to given value.

### HasIsDeleted

`func (o *PolicySyncResponsePoliciesInner) HasIsDeleted() bool`

HasIsDeleted returns a boolean if a field has been set.

### SetIsDeletedNil

`func (o *PolicySyncResponsePoliciesInner) SetIsDeletedNil(b bool)`

 SetIsDeletedNil sets the value for IsDeleted to be an explicit nil

### UnsetIsDeleted
`func (o *PolicySyncResponsePoliciesInner) UnsetIsDeleted()`

UnsetIsDeleted ensures that no value is present for IsDeleted, not even an explicit nil
### GetUpdateCounter

`func (o *PolicySyncResponsePoliciesInner) GetUpdateCounter() int64`

GetUpdateCounter returns the UpdateCounter field if non-nil, zero value otherwise.

### GetUpdateCounterOk

`func (o *PolicySyncResponsePoliciesInner) GetUpdateCounterOk() (*int64, bool)`

GetUpdateCounterOk returns a tuple with the UpdateCounter field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetUpdateCounter

`func (o *PolicySyncResponsePoliciesInner) SetUpdateCounter(v int64)`

SetUpdateCounter sets UpdateCounter field to given value.

### HasUpdateCounter

`func (o *PolicySyncResponsePoliciesInner) HasUpdateCounter() bool`

HasUpdateCounter returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


