# RegenerateServiceAccountSecretResponse

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**ClientSecret** | Pointer to **string** | Secret credential paired with the client ID. Used to obtain authentication tokens. Displayed only once upon regeneration. | [optional] 

## Methods

### NewRegenerateServiceAccountSecretResponse

`func NewRegenerateServiceAccountSecretResponse() *RegenerateServiceAccountSecretResponse`

NewRegenerateServiceAccountSecretResponse instantiates a new RegenerateServiceAccountSecretResponse object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewRegenerateServiceAccountSecretResponseWithDefaults

`func NewRegenerateServiceAccountSecretResponseWithDefaults() *RegenerateServiceAccountSecretResponse`

NewRegenerateServiceAccountSecretResponseWithDefaults instantiates a new RegenerateServiceAccountSecretResponse object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetClientSecret

`func (o *RegenerateServiceAccountSecretResponse) GetClientSecret() string`

GetClientSecret returns the ClientSecret field if non-nil, zero value otherwise.

### GetClientSecretOk

`func (o *RegenerateServiceAccountSecretResponse) GetClientSecretOk() (*string, bool)`

GetClientSecretOk returns a tuple with the ClientSecret field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetClientSecret

`func (o *RegenerateServiceAccountSecretResponse) SetClientSecret(v string)`

SetClientSecret sets ClientSecret field to given value.

### HasClientSecret

`func (o *RegenerateServiceAccountSecretResponse) HasClientSecret() bool`

HasClientSecret returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


