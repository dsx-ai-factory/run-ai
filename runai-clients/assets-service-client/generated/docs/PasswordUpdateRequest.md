# PasswordUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**CredentialUpdateRequest**](CredentialUpdateRequest.md) |  | [optional] 

## Methods

### NewPasswordUpdateRequest

`func NewPasswordUpdateRequest() *PasswordUpdateRequest`

NewPasswordUpdateRequest instantiates a new PasswordUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewPasswordUpdateRequestWithDefaults

`func NewPasswordUpdateRequestWithDefaults() *PasswordUpdateRequest`

NewPasswordUpdateRequestWithDefaults instantiates a new PasswordUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *PasswordUpdateRequest) GetMeta() CredentialUpdateRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *PasswordUpdateRequest) GetMetaOk() (*CredentialUpdateRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *PasswordUpdateRequest) SetMeta(v CredentialUpdateRequest)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *PasswordUpdateRequest) HasMeta() bool`

HasMeta returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


