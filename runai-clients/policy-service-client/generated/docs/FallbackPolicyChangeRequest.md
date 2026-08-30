# FallbackPolicyChangeRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyCreationFields**](PolicyCreationFields.md) |  | [optional] 
**Policy** | Pointer to [**NullableFallbackPolicyDefaultsAndRules**](FallbackPolicyDefaultsAndRules.md) |  | [optional] 
**Reset** | Pointer to **[]string** | set of fields in jsonpath format that is requested to clear their policy (default and rules) | [optional] 

## Methods

### NewFallbackPolicyChangeRequest

`func NewFallbackPolicyChangeRequest() *FallbackPolicyChangeRequest`

NewFallbackPolicyChangeRequest instantiates a new FallbackPolicyChangeRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewFallbackPolicyChangeRequestWithDefaults

`func NewFallbackPolicyChangeRequestWithDefaults() *FallbackPolicyChangeRequest`

NewFallbackPolicyChangeRequestWithDefaults instantiates a new FallbackPolicyChangeRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *FallbackPolicyChangeRequest) GetMeta() PolicyCreationFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *FallbackPolicyChangeRequest) GetMetaOk() (*PolicyCreationFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *FallbackPolicyChangeRequest) SetMeta(v PolicyCreationFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *FallbackPolicyChangeRequest) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *FallbackPolicyChangeRequest) GetPolicy() FallbackPolicyDefaultsAndRules`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *FallbackPolicyChangeRequest) GetPolicyOk() (*FallbackPolicyDefaultsAndRules, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *FallbackPolicyChangeRequest) SetPolicy(v FallbackPolicyDefaultsAndRules)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *FallbackPolicyChangeRequest) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *FallbackPolicyChangeRequest) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *FallbackPolicyChangeRequest) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil
### GetReset

`func (o *FallbackPolicyChangeRequest) GetReset() []string`

GetReset returns the Reset field if non-nil, zero value otherwise.

### GetResetOk

`func (o *FallbackPolicyChangeRequest) GetResetOk() (*[]string, bool)`

GetResetOk returns a tuple with the Reset field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReset

`func (o *FallbackPolicyChangeRequest) SetReset(v []string)`

SetReset sets Reset field to given value.

### HasReset

`func (o *FallbackPolicyChangeRequest) HasReset() bool`

HasReset returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


