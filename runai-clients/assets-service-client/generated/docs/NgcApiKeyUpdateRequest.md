# NgcApiKeyUpdateRequest

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Meta** | Pointer to [**CredentialUpdateRequest**](CredentialUpdateRequest.md) |  | [optional] 

## Methods

### NewNgcApiKeyUpdateRequest

`func NewNgcApiKeyUpdateRequest() *NgcApiKeyUpdateRequest`

NewNgcApiKeyUpdateRequest instantiates a new NgcApiKeyUpdateRequest object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewNgcApiKeyUpdateRequestWithDefaults

`func NewNgcApiKeyUpdateRequestWithDefaults() *NgcApiKeyUpdateRequest`

NewNgcApiKeyUpdateRequestWithDefaults instantiates a new NgcApiKeyUpdateRequest object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetMeta

`func (o *NgcApiKeyUpdateRequest) GetMeta() CredentialUpdateRequest`

GetMeta returns the Meta field if non-nil, zero value otherwise.

### GetMetaOk

`func (o *NgcApiKeyUpdateRequest) GetMetaOk() (*CredentialUpdateRequest, bool)`

GetMetaOk returns a tuple with the Meta field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetMeta

`func (o *NgcApiKeyUpdateRequest) SetMeta(v CredentialUpdateRequest)`

SetMeta sets Meta field to given value.

### HasMeta

`func (o *NgcApiKeyUpdateRequest) HasMeta() bool`

HasMeta returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


