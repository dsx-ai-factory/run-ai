# SecretSecret

## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**Type** | Pointer to **string** | The type of the secret. | [optional] 
**Value** | Pointer to **string** | The value of the secret. | [optional] 

## Methods

### NewSecretSecret

`func NewSecretSecret() *SecretSecret`

NewSecretSecret instantiates a new SecretSecret object
This constructor will assign default values to properties that have it defined,
and makes sure properties required by API are set, but the set of arguments
will change when the set of required properties is changed

### NewSecretSecretWithDefaults

`func NewSecretSecretWithDefaults() *SecretSecret`

NewSecretSecretWithDefaults instantiates a new SecretSecret object
This constructor will only assign default values to properties that have it defined,
but it doesn't guarantee that properties required by API are set

### GetType

`func (o *SecretSecret) GetType() string`

GetType returns the Type field if non-nil, zero value otherwise.

### GetTypeOk

`func (o *SecretSecret) GetTypeOk() (*string, bool)`

GetTypeOk returns a tuple with the Type field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetType

`func (o *SecretSecret) SetType(v string)`

SetType sets Type field to given value.

### HasType

`func (o *SecretSecret) HasType() bool`

HasType returns a boolean if a field has been set.

### GetValue

`func (o *SecretSecret) GetValue() string`

GetValue returns the Value field if non-nil, zero value otherwise.

### GetValueOk

`func (o *SecretSecret) GetValueOk() (*string, bool)`

GetValueOk returns a tuple with the Value field if it's non-nil, zero value otherwise
and a boolean to check if the value has been set.

### SetValue

`func (o *SecretSecret) SetValue(v string)`

SetValue sets Value field to given value.

### HasValue

`func (o *SecretSecret) HasValue() bool`

HasValue returns a boolean if a field has been set.


[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


