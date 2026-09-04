# NimServicePolicyChangeRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyCreationFields**](PolicyCreationFields.md) |  | [optional] 
**Policy** | Pointer to [**NullableNimServicePolicyDefaultsAndRules**](NimServicePolicyDefaultsAndRules.md) |  | [optional] 
**Reset** | Pointer to **[]string** | set of fields in jsonpath format that is requested to clear their policy (default and rules) | [optional] 

## Methods

### NewNimServicePolicyChangeRequest

`func NewNimServicePolicyChangeRequest() *NimServicePolicyChangeRequest`

NewNimServicePolicyChangeRequest instantiates a new NimServicePolicyChangeRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServicePolicyChangeRequestWithDefaults

`func NewNimServicePolicyChangeRequestWithDefaults() *NimServicePolicyChangeRequest`

NewNimServicePolicyChangeRequestWithDefaults instantiates a new NimServicePolicyChangeRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *NimServicePolicyChangeRequest) GetMeta() PolicyCreationFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *NimServicePolicyChangeRequest) GetMetaOk() (*PolicyCreationFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *NimServicePolicyChangeRequest) SetMeta(v PolicyCreationFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *NimServicePolicyChangeRequest) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *NimServicePolicyChangeRequest) GetPolicy() NimServicePolicyDefaultsAndRules`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *NimServicePolicyChangeRequest) GetPolicyOk() (*NimServicePolicyDefaultsAndRules, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *NimServicePolicyChangeRequest) SetPolicy(v NimServicePolicyDefaultsAndRules)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *NimServicePolicyChangeRequest) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *NimServicePolicyChangeRequest) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *NimServicePolicyChangeRequest) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil
### GetReset

`func (o *NimServicePolicyChangeRequest) GetReset() []string`

GetReset returns the Reset field if non-nil, zero value otherwise.

### GetResetOk

`func (o *NimServicePolicyChangeRequest) GetResetOk() (*[]string, bool)`

GetResetOk returns a tuple with the Reset field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetReset

`func (o *NimServicePolicyChangeRequest) SetReset(v []string)`

SetReset sets Reset field to given value.

### HasReset

`func (o *NimServicePolicyChangeRequest) HasReset() bool`

HasReset returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


