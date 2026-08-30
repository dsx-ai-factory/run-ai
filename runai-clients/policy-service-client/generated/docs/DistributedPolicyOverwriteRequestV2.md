# DistributedPolicyOverwriteRequestV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyCreationFields**](PolicyCreationFields.md) |  | [optional] 
**Policy** | Pointer to [**NullableDistributedPolicyDefaultsAndRulesV2**](DistributedPolicyDefaultsAndRulesV2.md) |  | [optional] 

## Methods

### NewDistributedPolicyOverwriteRequestV2

`func NewDistributedPolicyOverwriteRequestV2() *DistributedPolicyOverwriteRequestV2`

NewDistributedPolicyOverwriteRequestV2 instantiates a new DistributedPolicyOverwriteRequestV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedPolicyOverwriteRequestV2WithDefaults

`func NewDistributedPolicyOverwriteRequestV2WithDefaults() *DistributedPolicyOverwriteRequestV2`

NewDistributedPolicyOverwriteRequestV2WithDefaults instantiates a new DistributedPolicyOverwriteRequestV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DistributedPolicyOverwriteRequestV2) GetMeta() PolicyCreationFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DistributedPolicyOverwriteRequestV2) GetMetaOk() (*PolicyCreationFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DistributedPolicyOverwriteRequestV2) SetMeta(v PolicyCreationFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *DistributedPolicyOverwriteRequestV2) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *DistributedPolicyOverwriteRequestV2) GetPolicy() DistributedPolicyDefaultsAndRulesV2`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *DistributedPolicyOverwriteRequestV2) GetPolicyOk() (*DistributedPolicyDefaultsAndRulesV2, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *DistributedPolicyOverwriteRequestV2) SetPolicy(v DistributedPolicyDefaultsAndRulesV2)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *DistributedPolicyOverwriteRequestV2) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *DistributedPolicyOverwriteRequestV2) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *DistributedPolicyOverwriteRequestV2) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


