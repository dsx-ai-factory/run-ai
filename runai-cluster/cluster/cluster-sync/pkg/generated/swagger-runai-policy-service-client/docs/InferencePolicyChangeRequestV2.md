# InferencePolicyChangeRequestV2

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyCreationFields**](PolicyCreationFields.md) |  | [optional] 
**Policy** | Pointer to [**NullableInferencePolicyDefaultsAndRulesV2**](InferencePolicyDefaultsAndRulesV2.md) |  | [optional] 
**Reset** | Pointer to **[]string** | set of fields in jsonpath format that is requested to clear their policy (default and rules) | [optional] 

## Methods

### NewInferencePolicyChangeRequestV2

`func NewInferencePolicyChangeRequestV2() *InferencePolicyChangeRequestV2`

NewInferencePolicyChangeRequestV2 instantiates a new InferencePolicyChangeRequestV2 object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewInferencePolicyChangeRequestV2WithDefaults

`func NewInferencePolicyChangeRequestV2WithDefaults() *InferencePolicyChangeRequestV2`

NewInferencePolicyChangeRequestV2WithDefaults instantiates a new InferencePolicyChangeRequestV2 object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *InferencePolicyChangeRequestV2) GetMeta() PolicyCreationFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *InferencePolicyChangeRequestV2) GetMetaOk() (*PolicyCreationFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *InferencePolicyChangeRequestV2) SetMeta(v PolicyCreationFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *InferencePolicyChangeRequestV2) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *InferencePolicyChangeRequestV2) GetPolicy() InferencePolicyDefaultsAndRulesV2`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *InferencePolicyChangeRequestV2) GetPolicyOk() (*InferencePolicyDefaultsAndRulesV2, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *InferencePolicyChangeRequestV2) SetPolicy(v InferencePolicyDefaultsAndRulesV2)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *InferencePolicyChangeRequestV2) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *InferencePolicyChangeRequestV2) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *InferencePolicyChangeRequestV2) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil
### GetReset

`func (o *InferencePolicyChangeRequestV2) GetReset() []string`

GetReset returns the Reset field if non-nil, zero value otherwise.

### GetResetOk

`func (o *InferencePolicyChangeRequestV2) GetResetOk() (*[]string, bool)`

GetResetOk returns a tuple with the Reset field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReset

`func (o *InferencePolicyChangeRequestV2) SetReset(v []string)`

SetReset sets Reset field to given value.

### HasReset

`func (o *InferencePolicyChangeRequestV2) HasReset() bool`

HasReset returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


