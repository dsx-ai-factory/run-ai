# NimServicePolicyOverwriteRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**PolicyCreationFields**](PolicyCreationFields.md) |  | [optional] 
**Policy** | Pointer to [**NullableNimServicePolicyDefaultsAndRules**](NimServicePolicyDefaultsAndRules.md) |  | [optional] 

## Methods

### NewNimServicePolicyOverwriteRequest

`func NewNimServicePolicyOverwriteRequest() *NimServicePolicyOverwriteRequest`

NewNimServicePolicyOverwriteRequest instantiates a new NimServicePolicyOverwriteRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNimServicePolicyOverwriteRequestWithDefaults

`func NewNimServicePolicyOverwriteRequestWithDefaults() *NimServicePolicyOverwriteRequest`

NewNimServicePolicyOverwriteRequestWithDefaults instantiates a new NimServicePolicyOverwriteRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *NimServicePolicyOverwriteRequest) GetMeta() PolicyCreationFields`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *NimServicePolicyOverwriteRequest) GetMetaOk() (*PolicyCreationFields, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *NimServicePolicyOverwriteRequest) SetMeta(v PolicyCreationFields)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *NimServicePolicyOverwriteRequest) HasMeta() bool`

HasMeta returns a boolean if a field has been set.

### GetPolicy

`func (o *NimServicePolicyOverwriteRequest) GetPolicy() NimServicePolicyDefaultsAndRules`

GetPolicy returns the Policy field if non-nil, zero value otherwise.

### GetPolicyOk

`func (o *NimServicePolicyOverwriteRequest) GetPolicyOk() (*NimServicePolicyDefaultsAndRules, bool)`

GetPolicyOk returns a tuple with the Policy field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetPolicy

`func (o *NimServicePolicyOverwriteRequest) SetPolicy(v NimServicePolicyDefaultsAndRules)`

SetPolicy sets Policy field to given value.

### HasPolicy

`func (o *NimServicePolicyOverwriteRequest) HasPolicy() bool`

HasPolicy returns a boolean if a field has been set.

### SetPolicyNil

`func (o *NimServicePolicyOverwriteRequest) SetPolicyNil(b bool)`

 SetPolicyNil sets the value for Policy to be an explicit nil

### UnsetPolicy
`func (o *NimServicePolicyOverwriteRequest) UnsetPolicy()`

UnsetPolicy ensures that no value is present for Policy, not even an explicit nil

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


