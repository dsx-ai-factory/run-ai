# DistributedPolicyChangeRequestV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyCreationFields**](PolicyCreationFields.md) |  | [optional] 
**Policy** | Pointer to [**NullableDistributedPolicyDefaultsAndRulesV2**](DistributedPolicyDefaultsAndRulesV2.md) |  | [optional] 
**Reset** | Pointer to **[]string** | set of fields in jsonpath format that is requested to clear their policy (default and rules) | [optional] 

## Methods

### NewDistributedPolicyChangeRequestV2

`func NewDistributedPolicyChangeRequestV2() *DistributedPolicyChangeRequestV2`

NewDistributedPolicyChangeRequestV2 instantiates a new DistributedPolicyChangeRequestV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewDistributedPolicyChangeRequestV2WithDefaults

`func NewDistributedPolicyChangeRequestV2WithDefaults() *DistributedPolicyChangeRequestV2`

NewDistributedPolicyChangeRequestV2WithDefaults instantiates a new DistributedPolicyChangeRequestV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *DistributedPolicyChangeRequestV2) GetMeta() PolicyCreationFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *DistributedPolicyChangeRequestV2) GetMetaOk() (*PolicyCreationFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *DistributedPolicyChangeRequestV2) SetMeta(v PolicyCreationFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *DistributedPolicyChangeRequestV2) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *DistributedPolicyChangeRequestV2) GetPolicy() DistributedPolicyDefaultsAndRulesV2`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *DistributedPolicyChangeRequestV2) GetPolicyOk() (*DistributedPolicyDefaultsAndRulesV2, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *DistributedPolicyChangeRequestV2) SetPolicy(v DistributedPolicyDefaultsAndRulesV2)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *DistributedPolicyChangeRequestV2) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *DistributedPolicyChangeRequestV2) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *DistributedPolicyChangeRequestV2) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil
### GetReset

`func (o *DistributedPolicyChangeRequestV2) GetReset() []string`

GetReset returns the Reset field if non-nil, zero value otherwise.

### GetResetOk

`func (o *DistributedPolicyChangeRequestV2) GetResetOk() (*[]string, bool)`

GetResetOk returns a tuple with the Reset field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReset

`func (o *DistributedPolicyChangeRequestV2) SetReset(v []string)`

SetReset sets Reset field to given value.

### HasReset

`func (o *DistributedPolicyChangeRequestV2) HasReset() bool`

HasReset returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


